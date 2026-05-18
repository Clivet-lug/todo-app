package databases

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// GLOBAL CONNECTIONS
var db *sql.DB

// CONNECT TO POSTGRESQL
func ConnectDB() *sql.DB {
	godotenv.Load()
	connStr := os.Getenv("DATABASE_URL")

	if connStr == "" {
		connStr = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
		)
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Cannot reach database:", err)
	}

	fmt.Println("Connected to PostgreSQL successfully")
	return db
}
