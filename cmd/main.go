package main

import (
	"database/sql"
	"fmt"

	//"fmt"
	"log"
	"net/http"

	"CyberEase/scanner"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

var (
	store = sessions.NewCookieStore([]byte("super-secret-key"))
	db    *sql.DB
)

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./users.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create users table if not exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL
    )`)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	initDB()
	defer db.Close()

	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("themes"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/scan", scanner.ScanHandler)
	// Routes without trailing slashes
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/", pageHandler)

	// http.HandleFunc("/monitoring", monitoringHandler)
	// http.HandleFunc("/audit", auditHandler)
	// http.HandleFunc("/incident", incidentHandler)
	// http.HandleFunc("/penetration", penetrationHandler)

	// // // Public routes
	// http.HandleFunc("/contactus", contactusHandler)
	// http.HandleFunc("/consultation", consultationHandler)
	// http.HandleFunc("/education", educationHandler)
	// http.HandleFunc("/training", trainingHandler)
	// http.HandleFunc("/sub", subHandler)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
