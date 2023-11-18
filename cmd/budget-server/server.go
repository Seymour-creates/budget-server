package main

import (
	"github.com/Seymour-creates/budget-server/internal/router"
	_ "github.com/joho/godotenv"
	"log"
	"os"
)

func main() {

	port := os.Getenv("PORT")
	srv := router.NewServer()

	if err := srv.Run(port); err != nil {
		log.Fatal("Server failed to start: ", err)
	}

	log.Printf("Server running on port %v", port)
}
