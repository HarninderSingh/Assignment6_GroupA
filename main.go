package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Response struct for JSON output
type Response struct {
	CurrentTime string `json:"current_time"`
}

func main() {
	// Get the DSN from the environment variable
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable is not set")
	}
	fmt.Printf("Using DSN: %s\n", dsn)

	// Open a connection to MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
	fmt.Println("Connected to MySQL Database.")

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/current-time", func(w http.ResponseWriter, r *http.Request) {
		loc, _ := time.LoadLocation("America/Toronto")
		currentTime := time.Now().In(loc)

		query := `INSERT INTO time_log (timestamp) VALUES (?)`
		if _, err := db.Exec(query, currentTime); err != nil {
			http.Error(w, "Database insert failed", http.StatusInternalServerError)
			log.Printf("Error inserting time: %v", err)
			return
		}

		response := Response{
			CurrentTime: currentTime.Format(time.RFC1123),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Channel to listen for termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Println("Server running on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// Wait for termination signal
	<-stop
	log.Println("Shutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %v", err)
	}

	log.Println("Server exited gracefully.")
}
