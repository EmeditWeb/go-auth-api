package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"github.com/joho/godotenv"

	"github.com/Emeditweb/go-auth-api/internal/data"
	_ "github.com/lib/pq"
)

// Config holds all the settings for our application
type Config struct {
	Port int
	Env  string
	DB   struct {
		DSN string
	}
}

// Application acts as the central hub for dependency injection
type Application struct {
	Config Config
	Logger *log.Logger
	Models data.Models
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var cfg Config
	cfg.Port = 8080
	cfg.Env = os.Getenv("ENV")

	// Using the credentials that worked in your previous tests
	cfg.DB.DSN = os.Getenv("DB_DSN")

	logger := log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime)

	// 1. Open the Database connection
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	logger.Println("database connection pool established")

	// 2. Initialize the application with the NewModels constructor
	app := &Application{
		Config: cfg,
		Logger: logger,
		Models: data.NewModels(db),
	}

	// 3. Setup the HTTP Server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      app.routes(), // Routes defined in routes.go
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 4. Starting the server
	logger.Printf("starting %s server on port %d", cfg.Env, cfg.Port)
	err = server.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}
}

// openDB creates a connection pool and verifies it with a ping
func openDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DB.DSN)
	if err != nil {
		return nil, err
	}

	// Set connection pool limits (Optional but recommended)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(15 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}