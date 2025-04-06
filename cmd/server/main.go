// cmd/server/main.go
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"concert-ticket-api/api/grpc"
	"concert-ticket-api/api/rest"
	"concert-ticket-api/config"
	"concert-ticket-api/internal/repository/postgres"
	"concert-ticket-api/internal/service"
	"concert-ticket-api/pkg/db"
	"concert-ticket-api/pkg/logger"

	_ "github.com/jmoiron/sqlx"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	migrateOnly := flag.Bool("migrate", false, "run migrations and exit")
	waitForDB := flag.Bool("wait-for-db", false, "wait for database to be available")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		panic(err)
	}

	// Initialize logger
	log := logger.NewLogger(cfg.LogLevel)
	log.Info("Starting Concert Ticket Reservation API")

	// Wait for database if requested (useful in Docker/Kubernetes environments)
	if *waitForDB {
		log.Info("Waiting for database to be available...")
		if err := db.WaitForDatabase(cfg.Database, 60*time.Second); err != nil {
			log.Error("Failed to connect to database after waiting: %v", err)
			os.Exit(1)
		}
	}

	// Connect to database
	database, err := db.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Error("Failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer database.Close()

	// Run database migrations
	log.Info("Running database migrations...")
	if err := db.RunMigrations(cfg.Database, "scripts/migrations"); err != nil {
		log.Error("Failed to run migrations: %v", err)
		os.Exit(1)
	}
	log.Info("Migrations completed successfully")

	// Exit if only running migrations
	if *migrateOnly {
		log.Info("Migration only mode, exiting")
		os.Exit(0)
	}

	// Initialize repositories
	concertRepo := postgres.NewConcertRepository(database)
	bookingRepo := postgres.NewBookingRepository(database)

	// Initialize services
	concertService := service.NewConcertService(concertRepo)
	bookingService := service.NewBookingService(bookingRepo, concertRepo, cfg.MaxRetries)

	// Start REST API server
	restServer := rest.NewServer(concertService, bookingService, log, cfg.RESTPort)
	go func() {
		log.Info("Starting REST API server on port %d", cfg.RESTPort)
		if err := restServer.Start(); err != nil {
			log.Error("REST server error: %v", err)
			os.Exit(1)
		}
	}()

	// Start gRPC server
	grpcServer := grpc.NewServer(concertService, bookingService, log, cfg.GRPCPort)
	go func() {
		log.Info("Starting gRPC server on port %d", cfg.GRPCPort)
		if err := grpcServer.Start(); err != nil {
			log.Error("gRPC server error: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Info("Shutting down servers...")

	// Create a timeout context for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown servers gracefully
	if err := restServer.Shutdown(ctx); err != nil {
		log.Error("REST server shutdown error: %v", err)
	}

	grpcServer.Shutdown()

	log.Info("Servers stopped")
}
