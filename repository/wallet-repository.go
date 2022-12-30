package repository

import (
	. "github.com/WalletService/model"
	"github.com/jinzhu/gorm"
)

type TransferTxResult struct {
	Transfer    string `json:"transfer"`
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	FromEntry   string `json:"from_entry"`
	ToEntry     string `json:"to_entry"`
}

// func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
// 	var result TransferTxResult

// 	err := store.execTx(ctx, func(q *Queries) error {
// 		var err error

// 		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
// 			FromAccountID: arg.FromAccountID,
// 			ToAccountID:   arg.ToAccountID,
// 			Amount:        arg.Amount,
// 		})
// 		if err != nil {
// 			return err
// 		}

// 		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
// 			AccountID: arg.FromAccountID,
// 			Amount:    -arg.Amount,
// 		})
// 		if err != nil {
// 			return err
// 		}

// 		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
// 			AccountID: arg.ToAccountID,
// 			Amount:    arg.Amount,
// 		})
// 		if err != nil {
// 			return err
// 		}

// 		if arg.FromAccountID < arg.ToAccountID {
// 			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
// 		} else {
// 			result.ToAccount, result.FromAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
// 		}

// 		return nil
// 	})

// 	return result, err
// }

type IWalletRepository interface {
	GetWalletById(id int) (*Wallet, error)
	GetWalletByUserId(userID int) (*Wallet, error)
	CreateWallet(wallet *Wallet) (*Wallet, error)
	UpdateWallet(wallet *Wallet) (*Wallet, error)
	//DeleteWallet(wallet *Wallet) error
	//GetAllWallets() (*[]Wallet, error)
}

type walletRepository struct {
	DB *gorm.DB
}

func NewWalletRepository(db *gorm.DB) IWalletRepository {
	return &walletRepository{db}
}

func (walletRepository *walletRepository) GetWalletById(id int) (*Wallet, error) {
	var wallet Wallet
	result := walletRepository.DB.Preload("User").First(&wallet, id)
	return &wallet, result.Error
}

func (walletRepository *walletRepository) GetWalletByUserId(userID int) (*Wallet, error) {
	var wallet Wallet
	// use below association approach to avoid preload
	// result := walletRepository.DB.Where("user_id = ?", userID).Find(&wallet)
	// if result.Error != nil {
	// 	return nil, result.Error
	// }
	// err := walletRepository.DB.Model(&wallet).Association("user").Find(&wallet.User).Error
	// return &wallet, err
	result := walletRepository.DB.Where("user_id = ?", userID).Preload("User").First(&wallet)
	return &wallet, result.Error
}

func (walletRepository *walletRepository) CreateWallet(wallet *Wallet) (*Wallet, error) {
	result := walletRepository.DB.Create(wallet)
	return wallet, result.Error
}

func (walletRepository *walletRepository) UpdateWallet(wallet *Wallet) (*Wallet, error) {
	result := walletRepository.DB.Save(wallet)
	return wallet, result.Error
}

//
//func (walletRepository *WalletRepository) DeleteWallet(wallet *Wallet) error {
//	result := walletRepository.DB.Delete(wallet)
//	return result.Error
//}
//
//func (walletRepository *WalletRepository) GetAllWallets() (*[]Wallet, error) {
//	var wallet []Wallet
//	result := walletRepository.DB.Find(&wallet)
//	return &wallet, result.Error
//}

// func (walletRepository *walletRepository) Transfer(txn *Transaction,) (*Transaction, error) {
// 	walletRepository
// 	result := walletRepository.DB.Save(txn)
// 	return txn, result.Error
// }
