package db

import (
	"log"
	"sync"

	"github.com/WalletService/config"
	. "github.com/WalletService/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type gormDatabase struct {
	client *gorm.DB
	once   sync.Once
}

func NewGormDatabase() IDatabaseEngine {
	return &gormDatabase{}
}

func InitDatabase(g *gormDatabase, config *config.Database) {
	url := config.User + ":" + config.Password + "@tcp(" + config.Server + ":" +
		config.Port + ")/" + config.Name + "?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(config.Engine, url)
	if err != nil {
		log.Println("Database connection failed : ", err)
	} else {
		log.Println("Database connection established!")
	}
	log.Println("MySql connection running on port 13306")
	g.client = db
}

// Making sure gormClient only initialise once as singleton
func (g *gormDatabase) GetDatabase(config config.Database) *gorm.DB {
	if g.client == nil {
		g.once.Do(func() {
			InitDatabase(g, &config)
		})
	}
	return g.client
}

func (g *gormDatabase) RunMigration() {
	if g.client == nil {
		panic("Initialize gorm db before running migrations")
	}
	// g.client.AutoMigrate(&User{}, &Wallet{}, &Transaction{})
	g.client.AutoMigrate(&Wallet{}, &Transaction{})

	//We need to add foreign keys manually.
	// g.client.Model(&Wallet{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	g.client.Model(&Wallet{}).AddForeignKey("user_id", "users(user_id)", "CASCADE", "CASCADE")
	g.client.Model(&Transaction{}).AddForeignKey("wallet_id", "wallets(id)", "CASCADE", "CASCADE")
}
