package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func InitDB(dataSourceName string) *sql.DB {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Printf("successfully connected to db...")
	return db
}
