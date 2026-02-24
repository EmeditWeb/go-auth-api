package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Emeditweb/go-auth-api/internal/data"
	"github.com/joho/godotenv"
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
	WG     sync.WaitGroup // Tracks background tasks for graceful shutdown
}

func main() {
	// 1. Load Environment Variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	var cfg Config
	cfg.Port = 8080
	cfg.Env = os.Getenv("ENV")
	cfg.DB.DSN = os.Getenv("DB_DSN")

	// 2. Initialize Logger
	logger := log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime)

	// 3. Open Database Connection
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	logger.Println("database connection pool established")

	// 4. Initialize Application Dependencies
	app := &Application{
		Config: cfg,
		Logger: logger,
		Models: data.NewModels(db),
	}

	// 5. Start the Graceful Server
	err = app.serve()
	if err != nil {
		logger.Fatal(err)
	}
}

// serve manages the HTTP server lifecycle and Graceful Shutdown
func (app *Application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.Port),
		Handler:      app.routes(), // Routes defined in routes.go
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := make(chan error)

	// Background routine to intercept shutdown signals
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.Logger.Printf("shutting down server | signal: %s", s.String())

		// Give active requests 20 seconds to finish
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		// Wait for background WaitGroup tasks to complete
		app.Logger.Printf("completing background tasks | addr: %s", srv.Addr)
		app.WG.Wait()
		shutdownError <- nil
	}()

	app.Logger.Printf("starting %s server on port %d", app.Config.Env, app.Config.Port)

	// Start server. ListenAndServe returns ErrServerClosed during shutdown.
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Wait for the return value from the shutdown goroutine
	err = <-shutdownError
	if err != nil {
		return err
	}

	app.Logger.Printf("stopped server | addr: %s", srv.Addr)

	return nil
}

// openDB creates a connection pool and verifies it with a ping
func openDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DB.DSN)
	if err != nil {
		return nil, err
	}

	// Recommended connection pool limits
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