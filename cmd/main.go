package main

import (
	"database/sql"
	"fmt"
	"os"

	//"fmt"
	"log"
	"net/http"

	"CyberEase/scanner"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
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

	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}
	if os.Getenv("OPENROUTER_API_KEY") == "" {
		log.Fatal("OPENROUTER_API_KEY is required")
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("themes"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/scan", scanner.ScanHandler)
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/", pageHandler)

	mux.HandleFunc("/sub", handleHome)
	mux.HandleFunc("/ws", handleWebSocket)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
