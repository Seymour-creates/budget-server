package main

import (
	"database/sql"
	"fmt"
	"github.com/Seymour-creates/budget-server/db"
	"github.com/Seymour-creates/budget-server/router"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv"
	"log"
	"os"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbHost := os.Getenv("DB_HOST")
	port := os.Getenv("PORT")
	println("env variables: ", dbHost)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser,
		dbPass,
		dbHost,
		dbPort,
		dbName,
	)

	database := db.InitDB(dsn)
	defer func(database *sql.DB) {
		fmt.Printf("Closing db connection...")
		err := database.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}(database)

	srv := router.NewServer()
	if err := srv.Run(port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
