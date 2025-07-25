package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/routes"
	"aayushsiwa/expense-tracker/secure"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var errRest = make(chan error)
var ctx context.Context

func init() {
	db.Init("records.db")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	key := os.Getenv("ENCRYPTION_KEY")
	if len(key) != 32 {
		log.Fatalf("ENCRYPTION_KEY must be 32 bytes, got %d", len(key))
	}

	secure.SetKey([]byte(key))
}

func main() {
	log.Println("Starting Expense Tracker Server...")

	server := gin.Default()
	prefix := "/api/v1"
	apiGroup := server.Group(prefix)

	apiRoutes := routes.NewRoutes()
	routes.AttachRoutes(apiGroup, apiRoutes)

	ctx := context.Background()
	log.Println("Server running at http://localhost:8080")
	if err := server.Run(":8080"); err != nil {
		slog.ErrorContext(ctx, "ListenAndServe error", "error", err)
	}
}
