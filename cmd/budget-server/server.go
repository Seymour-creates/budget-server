package main

import (
	"log"
	"os"

	"github.com/Seymour-creates/budget-server/internal/router"
	_ "github.com/joho/godotenv"
)

func main() {
	port := os.Getenv("PORT")
	log.Printf("port: %v", port)
	srv := router.ConfigServer()

	if err := srv.Run(port); err != nil {
		log.Fatal("Server failed to start: ", err)
	}

	log.Printf("Server running on port %v", port)
}
