package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"sync"
)

var (
	db   *sql.DB
	once sync.Once
)

func InitDB(dataSourceName string) error {
	once.Do(func() {
		var err error
		db, err = sql.Open("mysql", dataSourceName)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}

		log.Printf("successfully connected to db...")
		if err = db.Ping(); err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}
	})
	return nil
}

func GetDB() *sql.DB {
	if db == nil {
		dsn := os.Getenv("DSN")
		log.Printf("DB Connection not established... establishing now using DSN: %v", os.Getenv("DSN"))
		if err := InitDB(dsn); err != nil {
			log.Fatal("error second attempt to connect to db: ", err)
		}
	}
	return db
}
