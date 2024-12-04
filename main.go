package main

import (
	"log"
	"net/http"
	"os"

	"jsonserver/internal/handlers"
	"jsonserver/internal/storage"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	password := os.Getenv("PASSWORD")
	if password == "" {
		log.Fatal("PASSWORD must be set in .env")
	}

	// Initialize SQLite storage
	store, err := storage.NewStorage("data.db")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize handlers
	apiHandler := handlers.NewAPIHandler(store, password)
	uiHandler, err := handlers.NewUIHandler(store, "internal/templates")
	if err != nil {
		log.Fatal(err)
	}

	// Setup routes
	http.HandleFunc("/", apiHandler.HandleRequest)
	http.HandleFunc("/ui", uiHandler.ServeHTTP)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
