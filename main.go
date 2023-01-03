package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/WalletService/cache"
	config "github.com/WalletService/config"
	"github.com/WalletService/controller"
	"github.com/WalletService/db"
	_ "github.com/WalletService/docs" // This line is necessary for go-swagger to find docs!
	"github.com/WalletService/middlewares"
	"github.com/WalletService/repository"
	"github.com/WalletService/scheduler"
	"github.com/WalletService/service"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var (
	c config.Config
	// httpRouter router.IRouter
	gormDb db.IDatabaseEngine
	gDb    *gorm.DB
)

// Cron
var (
	cronCache  cache.ICronCache
	reportCron scheduler.IReportCron
)

// User
var (
	userRepository repository.IUserRepository
	userService    service.IUserService
	userController controller.IUserController
)

// Wallet
var (
	walletCache      cache.IWalletCache
	walletRepository repository.IWalletRepository
	walletService    service.IWalletService
	walletController controller.IWalletController
)

// Transaction
var (
	transactionIdempotentCache cache.ITransactionIdempotentCache
	transactionRepository      repository.ITransactionRepository
	transactionService         service.ITransactionService
	transactionController      controller.ITransactionController
)

func main() {
	initConfig()
	// httpRouter = router.NewMuxRouter()
	// httpRouter.ADDVERSION("/api/v1")

	gormDb = db.NewGormDatabase()
	gDb = gormDb.GetDatabase(c.Database)
	gormDb.RunMigration()
	initCachingLayer()
	// initUserServiceContainer()
	// initWalletServiceContainer()
	// initTransactionServiceContainer()
	initCron()

	userRepository = repository.NewUserRepository(gDb)
	userService = service.NewUserService(userRepository)
	// userController = controller.NewUserController(userService)

	walletRepository = repository.NewWalletRepository(gDb)
	walletService = service.NewWalletService(walletRepository, userService, walletCache)
	walletController = controller.NewWalletController(walletService)

	transactionRepository = repository.NewTransactionRepository(gDb)
	transactionService = service.NewTransactionService(transactionRepository, walletService, userService, gDb)
	transactionController = controller.NewTransactionController(transactionService, transactionIdempotentCache)

	httpRouter := gin.Default()
	// httpRouter.Use(gin.Recovery())

	apiRoute := httpRouter.Group("/api/v1")
	{
		apiRoute.POST("/user/:id/wallet", walletController.CreateWallet)
		apiRoute.GET("/user/:id/wallet", walletController.GetWalletByUserId)
		apiRoute.GET("/wallet/:id", walletController.GetWallet)
		apiRoute.POST("/wallet/:id/block", walletController.BlockWallet)
		apiRoute.POST("/wallet/:id/unblock", walletController.UnBlockWallet)

		apiRoute.GET("/transaction", transactionController.GetTransactions)
		apiRoute.GET("/transaction/active", transactionController.GetActiveTransactions)
		apiRoute.GET("/transaction/:id", transactionController.GetTransaction)

		apiRoute.POST("/transfer", middlewares.DBTransactionMiddleware(gDb), transactionController.CreateTransfer)

		apiRoute.GET("/wallet/:id/transaction", transactionController.GetTransactionsByWalletId)
		apiRoute.POST("/wallet/:id/transaction", transactionController.PostTransaction)
		apiRoute.PUT("/transaction/active", transactionController.UpdateActiveTransactions)
	}

	httpRouter.Run(":" + c.App.Port)
}

func initConfig() {
	file, err := os.Open("./config.json")
	if err != nil {
		log.Printf("No ./config.json file found!! Terminating the server, error: %s\n", err.Error())
		panic("No config file found! Error : " + err.Error())
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&c)
	if err != nil {
		log.Printf("Error occurred while decoding json to config model, error: %s\n", err.Error())
		panic(err.Error())
	}
}

func initCachingLayer() {
	cacheEngine := cache.NewCache(c.Cache, c.Cache.Wallet.Db)
	cacheEngine.GetCacheClient()
	if err := cacheEngine.CheckConnection(); err != nil {
		panic(err)
	}
	walletCache = cache.NewWalletCache(cacheEngine, c.Cache.Wallet) // 0 means no expiry date

	cacheEngine2 := cache.NewCache(c.Cache, c.Cache.Idempotent.Db)
	cacheEngine2.GetCacheClient()
	if err := cacheEngine2.CheckConnection(); err != nil {
		panic(err)
	}
	transactionIdempotentCache = cache.NewTransactionIdempotentCache(cacheEngine2, c.Cache.Idempotent)

	cacheEngine3 := cache.NewCache(c.Cache, c.Cache.CronLock.Db)
	cacheEngine3.GetCacheClient()
	if err := cacheEngine3.CheckConnection(); err != nil {
		panic(err)
	}
	cronCache = cache.NewCronCache(cacheEngine3, c.Cache.CronLock)
}

// func initUserServiceContainer() {
// 	userRepository = repository.NewUserRepository(gDb)
// 	userService = service.NewUserService(userRepository)
// 	userController = controller.NewUserController(userService)

// httpRouter.GET("/user/{id}", userController.GetUser)
// httpRouter.GET("/user", userController.GetUsers)
// httpRouter.POST("/user", userController.PostUser)
// httpRouter.PUT("/user/{id}", userController.PutUser)
// httpRouter.DELETE("/user/{id}", userController.DeleteUser)
// }

// func initTransactionServiceContainer() {
// 	transactionRepository = repository.NewTransactionRepository(gDb)
// 	transactionService = service.NewTransactionService(transactionRepository, walletService)
// 	transactionController = controller.NewTransactionController(transactionService, walletService, transactionIdempotentCache)

// httpRouter.GET("/transaction", transactionController.GetTransactions)
// httpRouter.GET("/transaction/active", transactionController.GetActiveTransactions)
// httpRouter.GET("/transaction/{id}", transactionController.GetTransaction)
// httpRouter.GET("/wallet/{id}/transaction", transactionController.GetTransactionsByWalletId)
// httpRouter.POST("/wallet/{id}/transaction", transactionController.PostTransaction)
// httpRouter.PUT("/transaction/active", transactionController.UpdateActiveTransactions)
// }

func initCron() {
	reportCron = scheduler.NewReportCron(transactionService, cronCache)
	reportCron.StartReportCron()
}
