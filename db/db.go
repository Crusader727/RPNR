package db

import (
	"fmt"
	"log"

	"../confreader"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

type dbConfig struct {
	Adapter  string
	Database string
}

func getDbConfig() (*dbConfig, error) {
	var conf dbConfig

	if err := confreader.ReadConfig("db", &conf); err != nil {
		return nil, fmt.Errorf("can't get db config: %s", err)
	}

	return &conf, nil
}

func init() {
	if err := initDB(); err != nil {
		log.Fatalf("can't init DB: %s", err)
	}
}

func initDB() error {
	cfg, err := getDbConfig()
	if err != nil {
		return fmt.Errorf("can't get db env config: %s", err)
	}

	db, err = gorm.Open(cfg.Adapter, cfg.Database)
	if err != nil {
		return fmt.Errorf("can't open gorm connection for cfg %+v: %s", cfg, err)
	}

	return nil
}

func Get() *gorm.DB {
	return db
}
