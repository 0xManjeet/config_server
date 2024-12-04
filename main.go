package main

import (
	"log"
	"os"

	"jsonserver/internal/handlers"
	"jsonserver/internal/storage"

	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp"
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

	corsAllowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsAllowedOrigins == "" {
		corsAllowedOrigins = "noxchat.in"
	}

	// Initialize SQLite storage
	store, err := storage.NewStorage("data.db")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize handlers
	apiHandler := handlers.NewAPIHandler(store, password, corsAllowedOrigins)
	uiHandler, err := handlers.NewUIHandler(store, "internal/templates")
	if err != nil {
		log.Fatal(err)
	}

	// Create a request handler that routes between API and UI
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		if path == "/ui" {
			uiHandler.HandleFastHTTP(ctx)
		} else {
			apiHandler.HandleFastHTTP(ctx)
		}
	}

	// Configure server
	server := &fasthttp.Server{
		Handler:            requestHandler,
		Name:               "JsonServer",
		MaxRequestBodySize: 1024 * 1024, // 1MB
		Concurrency:        100000,      // Default is 256*1024
	}

	log.Printf("Server starting on port %s", port)
	if err := server.ListenAndServe(":" + port); err != nil {
		log.Fatal(err)
	}
}
