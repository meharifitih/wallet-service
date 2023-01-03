package controller

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	. "github.com/WalletService/model"
	"github.com/WalletService/service"
	"github.com/gin-gonic/gin"
)

type IWalletController interface {
	GetWallet(c *gin.Context)
	GetWalletByUserId(c *gin.Context)
	CreateWallet(c *gin.Context)
	BlockWallet(c *gin.Context)
	UnBlockWallet(c *gin.Context)
}

type walletController struct{}

var (
	walletService service.IWalletService
)

func NewWalletController(service service.IWalletService) IWalletController {
	walletService = service
	return &walletController{}
}

func (walletController *walletController) GetWallet(c *gin.Context) {
	// vars := mux.Vars(r)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid ID")
		return
	}
	wallet, err := walletService.GetWalletService(id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, "Wallet not found")
		default:
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, wallet)
}

func (walletController *walletController) GetWalletByUserId(c *gin.Context) {
	id := c.Param("id")
	wallet, err := walletService.GetWalletByUserIdService(id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			c.JSON(http.StatusNotFound, "Wallet not found")
		default:
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, wallet)
}

func (walletController *walletController) CreateWallet(c *gin.Context) {
	var wallet Wallet
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request payload")
		return
	}
	id := c.Param("id")
	defer c.Request.Body.Close()
	res, err := walletService.PostWalletService(&wallet, id)
	if err != nil {
		log.Printf("error 1 post Wallet : %s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (walletController *walletController) BlockWallet(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid wallet ID")
		return
	}
	err = walletService.BlockWalletService(id)
	if err != nil {
		log.Printf("Not able to block Wallet : %s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusLocked, map[string]string{"message": "Wallet is blocked successfully!"})
}

func (walletController *walletController) UnBlockWallet(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid wallet ID")
		return
	}
	err = walletService.UnBlockWalletService(id)
	if err != nil {
		log.Printf("Not able to unblock Wallet : %s", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusLocked, map[string]string{"message": "Wallet is unblocked successfully!"})
}

//func (walletController *walletController) PutWallet(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	id, err := strconv.Atoi(vars["id"])
//	if err != nil {
//		respondWithError(w, http.StatusBadRequest, "Invalid wallet ID")
//		return
//	}
//	var b Wallet
//	decoder := json.NewDecoder(r.Body)
//	if err := decoder.Decode(&b); err != nil {
//		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
//		return
//	}
//	defer r.Body.Close()
//	res, err := walletService.UpdateWalletService(id, &b)
//	if err != nil {
//		switch err {
//		case sql.ErrNoRows:
//			respondWithError(w, http.StatusNotFound, "Wallet not found")
//		default:
//			respondWithError(w, http.StatusInternalServerError, err.Error())
//		}
//		return
//	}
//	respondWithJSON(w, http.StatusOK, res)
//}
//
//func (walletController *walletController) DeleteWallet(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	id, err := strconv.Atoi(vars["id"])
//	if err != nil {
//		respondWithError(w, http.StatusBadRequest, "Invalid wallet ID")
//		return
//	}
//	err = walletService.DeleteWalletService(id)
//	if err != nil {
//		switch err {
//		case sql.ErrNoRows:
//			respondWithError(w, http.StatusNotFound, "Wallet not found")
//		default:
//			respondWithError(w, http.StatusInternalServerError, err.Error())
//		}
//		return
//	}
//	respondWithJSON(w, http.StatusOK, map[string]string{"result" : "success"})
//}
//
//func (walletController *walletController) GetWallets(w http.ResponseWriter, r *http.Request) {
//	wallets, err := walletService.GetWalletsService()
//	if err != nil {
//		respondWithError(w, http.StatusInternalServerError, err.Error())
//		return
//	}
//	respondWithJSON(w, http.StatusOK, wallets)
//}
