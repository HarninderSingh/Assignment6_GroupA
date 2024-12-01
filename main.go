package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Response struct for JSON output
type Response struct {
	CurrentTime string `json:"current_time"`
}

func main() {
	// MySQL DSN: username:password@tcp(host:port)/database
	dsn := "root:@tcp(127.0.0.1:3306)/time_logging" // Update 'root' and '' with your username and password if set

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

	// HTTP handler for /current-time
	http.HandleFunc("/current-time", func(w http.ResponseWriter, r *http.Request) {
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

	// Start the server
	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
