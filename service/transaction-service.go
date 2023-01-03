package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	. "github.com/WalletService/model"
	"github.com/WalletService/repository"
	"github.com/jinzhu/gorm"
)

type ITransactionService interface {
	GetTransactionService(id int) (*Transaction, error)
	GetTransactionsByWalletIdService(id int) (*[]Transaction, error)
	GetTransactionsService() (*[]Transaction, error)
	GetActiveTransactionsService() (*[]Transaction, error)
	PostTransactionService(transaction *Transaction, walletID int) (*Transaction, error)
	UpdateTransactionService(id int, transaction *Transaction) (*Transaction, error)
	UpdateActiveTransactionsService() error
	validAccount(ctx context.Context, accountID string, currency string) (*Wallet, bool)
	WithTrx(trxHandle *gorm.DB) *transactionService
	HandleMoney(ctx context.Context, transfer *TransferRequest) ([]Wallet, error)
	//DeleteTransactionService(id int) error
}

type transactionService struct{}

var (
	transactionRepository repository.ITransactionRepository
	iWalletService        IWalletService
	uService              IUserService
	DB                    *gorm.DB
)

func NewTransactionService(repository repository.ITransactionRepository, iService IWalletService, uSer IUserService, db *gorm.DB) ITransactionService {
	transactionRepository = repository
	iWalletService = iService
	uService = uSer
	return &transactionService{}
}

func (transactionService *transactionService) GetTransactionService(id int) (*Transaction, error) {
	return transactionRepository.GetTransactionById(id)
}

func (transactionService *transactionService) GetTransactionsByWalletIdService(id int) (*[]Transaction, error) {
	return transactionRepository.GetTransactionsByWalletId(id)
}

func (transactionService *transactionService) GetTransactionsService() (*[]Transaction, error) {
	return transactionRepository.GetAllTransactions()
}

func (transactionService *transactionService) GetActiveTransactionsService() (*[]Transaction, error) {
	return transactionRepository.GetAllActiveTransactions()
}

func (transactionService *transactionService) PostTransactionService(transaction *Transaction, walletID int) (*Transaction, error) {
	txnType := strings.ToLower(transaction.TxnType)
	if txnType != "credit" && txnType != "debit" {
		return nil, errors.New("txn type can only be credit or debit")
	}
	wallet, err := iWalletService.GetWalletService(walletID)
	if err != nil {
		return nil, err
	}
	if wallet.IsBlock {
		return nil, errors.New("this wallet is blocked. can't perform any transactions")
	}
	if txnType == "credit" {
		wallet.Balance += transaction.Amount
	} else {
		if wallet.Balance < transaction.Amount {
			return nil, errors.New("wallet balance is insufficient to deduct given amount")
		}
		wallet.Balance -= transaction.Amount
	}
	if _, err = iWalletService.UpdateWalletService(wallet); err != nil {
		return nil, err
	}
	transaction.WalletID = uint(walletID)
	transaction.Wallet = *wallet
	return transactionRepository.CreateTransaction(transaction)
}

func (transactionService *transactionService) validAccount(ctx context.Context, accountID string, currency string) (*Wallet, bool) {
	user, _ := uService.GetUserByPhone(accountID)
	if user.UserID != "" {
		accountID = user.UserID
	}

	account, err := iWalletService.GetWalletByUserIdService(accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return account, false
		}
		return account, false
	}

	if account.Currency != currency {
		// err = fmt.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		return account, false
	}

	return account, true
}

func (transactionService *transactionService) UpdateActiveTransactionsService() error {
	return transactionRepository.UpdateAllActiveTransactions()
}

func (transactionService *transactionService) UpdateTransactionService(id int, transaction *Transaction) (*Transaction, error) {
	res, err := transactionRepository.GetTransactionById(id)
	if err != nil {
		return nil, err
	}
	updateTransactionsGivenFields(res, transaction)
	return transactionRepository.UpdateTransaction(res)
}

//func (transactionService *TransactionService) DeleteTransactionService(id int) error {
//	res, err := transactionService.GetTransactionById(id)
//	if err != nil {
//		return err
//	}
//	return transactionService.DeleteTransaction(res)
//}

func updateTransactionsGivenFields(u1 *Transaction, u2 *Transaction) {
	if len(u2.TxnType) != 0 {
		u1.TxnType = u2.TxnType
	}
	if u2.Amount != 0.0 {
		u1.Amount = u2.Amount
	}
	// if provided json value is true, then only update it
	if u2.Active {
		u1.Active = u2.Active
	}
}

func (transactionService *transactionService) WithTrx(trxHandle *gorm.DB) *transactionService {

	transactionRepository = transactionRepository.WithTrx(trxHandle)
	return transactionService
}

func (transactionService *transactionService) HandleMoney(ctx context.Context, transfer *TransferRequest) ([]Wallet, error) {
	wallets := []Wallet{}
	fromAccount, valid := transactionService.validAccount(ctx, transfer.FromAccountID, transfer.Currency)
	if !valid {
		return wallets, errors.New("account not valid")
	}

	toAccount, valid := transactionService.validAccount(ctx, transfer.ToAccountID, transfer.Currency)
	if !valid {
		return wallets, errors.New("account not valid")
	}

	if fromAccount.IsBlock {
		return wallets, errors.New("this wallet is blocked. can't perform any transactions")
	}

	if toAccount.IsBlock {
		return wallets, errors.New("this wallet is blocked. can't perform any transactions")
	}

	if transfer.Amount > fromAccount.Balance {
		return wallets, errors.New("insufficient amount in account")
	}

	sender, err := transactionRepository.DecrementMoney(fromAccount.ID, transfer.Amount)
	if err != nil {
		return wallets, err
	}
	txn1 := Transaction{
		TxnType:  "debit",
		Amount:   transfer.Amount,
		WalletID: fromAccount.ID,
	}
	_, err = transactionRepository.CreateTransaction(&txn1)
	if err != nil {
		return wallets, err
	}

	err = walletCache.Set(transfer.FromAccountID, sender)
	if err != nil {
		return wallets, err
	}

	wallets = append(wallets, *sender)

	receiver, err := transactionRepository.IncrementMoney(toAccount.ID, transfer.Amount)
	if err != nil {
		return wallets, err
	}

	txn2 := Transaction{
		TxnType:  "credit",
		Amount:   transfer.Amount,
		WalletID: toAccount.ID,
	}
	_, err = transactionRepository.CreateTransaction(&txn2)
	if err != nil {
		return wallets, err
	}

	err = walletCache.Set(transfer.ToAccountID, receiver)
	if err != nil {
		return wallets, err
	}

	wallets = append(wallets, *receiver)

	return wallets, nil
}
