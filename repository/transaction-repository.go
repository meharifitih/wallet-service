package repository

import (
	"log"

	"github.com/WalletService/model"
	. "github.com/WalletService/model"
	"github.com/jinzhu/gorm"
)

type ITransactionRepository interface {
	GetWallet(id string) (*Wallet, error)
	GetTransactionById(id int) (*Transaction, error)
	GetTransactionsByWalletId(id int) (*[]Transaction, error)
	GetAllTransactions() (*[]Transaction, error)
	GetAllActiveTransactions() (*[]Transaction, error)
	CreateTransaction(transaction *Transaction) (*Transaction, error)
	UpdateTransaction(transaction *Transaction) (*Transaction, error)
	UpdateAllActiveTransactions() error
	WithTrx(trxHandle *gorm.DB) *transactionRepository
	IncrementMoney(receiver uint, amount float64) (*Wallet, error)
	DecrementMoney(giver uint, amount float64) (*Wallet, error)
	//DeleteTransaction(transaction *Transaction) error
}

type transactionRepository struct {
	DB *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) ITransactionRepository {
	return &transactionRepository{db}
}

func (transactionRepository *transactionRepository) WithTrx(trxHandle *gorm.DB) *transactionRepository {
	if trxHandle == nil {
		log.Print("Transaction Database not found")
		return transactionRepository
	}
	transactionRepository.DB = trxHandle
	return transactionRepository
}
func (transactionRepository *transactionRepository) IncrementMoney(receiver uint, amount float64) (*Wallet, error) {
	wallet := model.Wallet{}
	err := transactionRepository.DB.Model(&wallet).Where("id=?", receiver).Update("balance", gorm.Expr("balance + ?", amount)).First(&wallet, receiver).Error
	return &wallet, err
}

func (transactionRepository *transactionRepository) DecrementMoney(giver uint, amount float64) (*Wallet, error) {
	wallet := model.Wallet{}
	err := transactionRepository.DB.Model(&wallet).Where("id=?", giver).Update("balance", gorm.Expr("balance - ?", amount)).First(&wallet, giver).Error
	return &wallet, err
}

func (transactionRepository *transactionRepository) GetTransactionById(id int) (*Transaction, error) {
	var transaction Transaction
	result := transactionRepository.DB.Preload("Wallet").Preload("Wallet.User").First(&transaction, id)
	return &transaction, result.Error
}

func (transactionRepository *transactionRepository) GetWallet(id string) (*Wallet, error) {
	var wallet Wallet
	result := transactionRepository.DB.Table("wallets").Where("id = ?", id).First(&wallet)
	return &wallet, result.Error
}

func (transactionRepository *transactionRepository) GetTransactionsByWalletId(id int) (*[]Transaction, error) {
	var transaction []Transaction
	result := transactionRepository.DB.Where("wallet_id = ?", id).Preload("Wallet").Preload("Wallet.User").Find(&transaction)
	return &transaction, result.Error
}

func (transactionRepository *transactionRepository) GetAllTransactions() (*[]Transaction, error) {
	var transaction []Transaction
	result := transactionRepository.DB.Preload("Wallet").Preload("Wallet.User").Find(&transaction)
	return &transaction, result.Error
}

func (transactionRepository *transactionRepository) GetAllActiveTransactions() (*[]Transaction, error) {
	var transaction []Transaction
	result := transactionRepository.DB.Where("active = ?", true).Preload("Wallet").Preload("Wallet.User").Find(&transaction)
	return &transaction, result.Error
}

func (transactionRepository *transactionRepository) CreateTransaction(transaction *Transaction) (*Transaction, error) {
	result := transactionRepository.DB.Create(transaction)
	return transaction, result.Error
}

func (transactionRepository *transactionRepository) UpdateTransaction(transaction *Transaction) (*Transaction, error) {
	result := transactionRepository.DB.Save(transaction)
	return transaction, result.Error
}

func (transactionRepository *transactionRepository) UpdateAllActiveTransactions() error {
	result := transactionRepository.DB.Model(Transaction{}).Where("active = ?", true).Updates(map[string]interface{}{"active": false})
	return result.Error
}

//func (transactionRepository *TransactionRepository) DeleteTransaction(transaction *Transaction) error {
//	result := transactionRepository.DB.Delete(transaction)
//	return result.Error
//}
