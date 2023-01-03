package controller

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/WalletService/cache"
	. "github.com/WalletService/model"
	"github.com/WalletService/service"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type ITransactionController interface {
	GetTransaction(c *gin.Context)
	GetTransactionsByWalletId(c *gin.Context)
	GetTransactions(c *gin.Context)
	GetActiveTransactions(c *gin.Context)
	PostTransaction(c *gin.Context)
	CreateTransfer(ctx *gin.Context)
	PutTransaction(c *gin.Context)
	UpdateActiveTransactions(c *gin.Context)
}

type transactionController struct{}

var (
	transactionService         service.ITransactionService
	transactionIdempotentCache cache.ITransactionIdempotentCache
)

func NewTransactionController(service service.ITransactionService, idempotent cache.ITransactionIdempotentCache) ITransactionController {
	transactionService = service
	transactionIdempotentCache = idempotent
	return &transactionController{}
}

func (transactionController *transactionController) GetTransaction(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid transaction ID")
		return
	}
	transaction, err := transactionService.GetTransactionService(id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, "Transaction not found")
		default:
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, transaction)
}

func (transactionController *transactionController) GetTransactionsByWalletId(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid wallet ID")
		return
	}
	transaction, err := transactionService.GetTransactionsByWalletIdService(id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, "Transactions not found")
		default:
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, transaction)
}

func (transactionController *transactionController) GetTransactions(c *gin.Context) {
	transactions, err := transactionService.GetTransactionsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, transactions)
}

func (transactionController *transactionController) GetActiveTransactions(c *gin.Context) {
	transactions, err := transactionService.GetActiveTransactionsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, transactions)
}

func (transactionController *transactionController) PostTransaction(c *gin.Context) {
	// check for x-idempotency-key
	key, err := transactionIdempotentCache.GetIdempotencyKey(c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	// if this request is not unique, return with idempotent transaction result
	if idempotentTransaction := transactionIdempotentCache.Get(key); idempotentTransaction != nil {
		c.JSON(http.StatusCreated, idempotentTransaction)
		return
	}
	var transaction Transaction
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}
	// vars := mux.Vars(r)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid transaction ID")
		return
	}
	defer c.Request.Body.Close()
	res, err := transactionService.PostTransactionService(&transaction, id)
	if err != nil {
		log.Printf("Not able to post Transaction : %s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	// Record this request with x-idempotency-key for certain expiry
	transactionIdempotentCache.Set(key, res)
	c.JSON(http.StatusCreated, res)
}

func (transactionController *transactionController) CreateTransfer(c *gin.Context) {
	txHandle := c.MustGet("db_trx").(*gorm.DB)

	var req TransferRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	wallets, err := transactionService.WithTrx(txHandle).HandleMoney(c, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, wallets)
}

func (transactionController *transactionController) PutTransaction(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid transaction ID")
		return
	}
	var b Transaction
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&b); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer c.Request.Body.Close()
	res, err := transactionService.UpdateTransactionService(id, &b)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, "Transaction not found")
		default:
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, res)
}

func (transactionController *transactionController) UpdateActiveTransactions(c *gin.Context) {
	err := transactionService.UpdateActiveTransactionsService()
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, "No Transactions updated!")
		default:
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, map[string]string{"message": "Transactions marked as inactive."})
}

//func (transactionController *TransactionController) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	id, err := strconv.Atoi(vars["id"])
//	if err != nil {
//		transactionController.RespondWithError(w, http.StatusBadRequest, "Invalid transaction ID")
//		return
//	}
//	err = transactionController.DeleteTransactionService(id)
//	if err != nil {
//		switch err {
//		case sql.ErrNoRows:
//			transactionController.RespondWithError(w, http.StatusNotFound, "Transaction not found")
//		default:
//			transactionController.RespondWithError(w, http.StatusInternalServerError, err.Error())
//		}
//		return
//	}
//	transactionController.RespondWithJSON(w, http.StatusOK, map[string]string{"result" : "success"})
//}
