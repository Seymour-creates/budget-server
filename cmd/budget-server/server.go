package main

import (
	"fmt"
	db2 "github.com/Seymour-creates/budget-server/internal/db"
	"github.com/Seymour-creates/budget-server/internal/router"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv"
	"log"
	"os"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Print("No dev.env file found")
	}

	dsn := os.Getenv("DSN")
	fmt.Println("Data source name? :", dsn)
	if err := db2.InitDB(dsn); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	port := os.Getenv("PORT")

	srv := router.NewServer()
	if err := srv.Run(port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
