package main

import (
	"fmt"
	db2 "github.com/Seymour-creates/budget-server/db"
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
	dbPort := os.Getenv("DB_PORT")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)
	if err := os.Setenv("DSN", dsn); err != nil {
		log.Printf("Error setting env variable DSN: --> %v", err)
	}
	if err := db2.InitDB(dsn); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	port := os.Getenv("PORT")

	srv := router.NewServer()
	if err := srv.Run(port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
