package main

import (
	"context"
	"crypto-portfolio-tracker/internal/db"
	"crypto-portfolio-tracker/internal/handlers"
	"crypto-portfolio-tracker/internal/services"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	log.Println("🚀 Starting Crypto Portfolio Tracker API...")

	// Initialize ScyllaDB
	scyllaDB, err := db.NewScyllaDB([]string{"localhost:9042"})
	if err != nil {
		log.Fatalf("Failed to connect to ScyllaDB: %v", err)
	}
	defer scyllaDB.Close()

	// Initialize schema
	if err := scyllaDB.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize ScyllaDB schema: %v", err)
	}

	// Initialize ElasticSearch
	elasticSearch, err := db.NewElasticSearch([]string{"http://localhost:9200"})
	if err != nil {
		log.Fatalf("Failed to connect to ElasticSearch: %v", err)
	}

	// Initialize ElasticSearch index
	if err := elasticSearch.InitIndex(); err != nil {
		log.Fatalf("Failed to initialize ElasticSearch index: %v", err)
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Crypto Portfolio Tracker API v1.0",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Initialize handlers
	h := handlers.NewHandler(scyllaDB, elasticSearch)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker := services.NewPriceWorker(scyllaDB, elasticSearch, 1*time.Minute)
	go worker.Start(ctx)

	// Routes
	api := app.Group("/api/v1")

	api.Get("/health", h.HealthCheck)
	api.Post("/tokens", h.AddToken)
	api.Get("/tokens/:id", h.GetToken)
	api.Get("/search", h.SearchTokens)
	api.Post("/sync", h.SyncTokens)
	api.Get("/history/:id", h.GetPriceHistory)
	api.Get("/tokens", h.GetAllTokens)
	api.Get("/analytics", h.GetAnalytics)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("\n🛑 Shutting down gracefully...")
		app.Shutdown()
	}()

	// Start server
	port := ":8080"
	log.Printf("✅ Server running on http://localhost%s", port)
	log.Println("📚 API Endpoints:")
	log.Println("   GET  /api/v1/health")
	log.Println("   POST /api/v1/tokens")
	log.Println("   GET  /api/v1/tokens/:id")
	log.Println("   GET  /api/v1/search?q=bitcoin")
	log.Println("   POST /api/v1/sync?limit=10")
	log.Println("   GET  /api/v1/history/:id?limit=100")
	log.Println("   GET  /api/v1/analytics")
	log.Println("   GET  /api/v1/tokens")

	if err := app.Listen(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
