package main

import (
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
	dsn := os.Getenv("DSN")

	if err := db2.InitDB(dsn); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	port := os.Getenv("PORT")

	srv := router.NewServer()
	if err := srv.Run(port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
