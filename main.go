package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	log.Printf("Loading env(s)...")
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading env | err: %v\n", err)
	}

	// Creating Router
	router := Arise()
	router.Run("0.0.0.0:" + os.Getenv("PORT"))
}
