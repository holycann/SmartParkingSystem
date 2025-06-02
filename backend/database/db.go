package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	// DB is the global database connection pool
	DB *sql.DB
	mu sync.Mutex
)

// Initialize sets up the database connection pool
func Initialize(connectionString ...string) error {
	mu.Lock()
	defer mu.Unlock()

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	var connStr string
	if len(connectionString) > 0 && connectionString[0] != "" {
		// Use provided connection string if available
		connStr = connectionString[0]
	} else {
		// Get database connection parameters from environment variables with fallbacks
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "holycan")
		password := getEnv("DB_PASSWORD", "ramaa212!")
		dbname := getEnv("DB_NAME", "smart_parking_db")
		sslmode := getEnv("DB_SSL_MODE", "disable")

		// Construct connection string
		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
	}

	// If DB is already initialized, close it first
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing existing database connection: %v", err)
		}
	}

	// Open a new connection
	log.Printf("Connecting to database: %s", connStr)
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Error opening database connection: %v", err)
		return err
	}

	// Configure the connection pool
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err = DB.Ping(); err != nil {
		log.Printf("Error pinging database: %v", err)
		return err
	}

	log.Println("Database connection established successfully")

	// Initialize database schema
	err = initSchema(DB)
	if err != nil {
		log.Printf("Error initializing database schema: %v", err)
		return err
	}

	return nil
}

// GetDB returns the global database connection
func GetDB() *sql.DB {
	return DB
}

// Ping checks if the database connection is alive
func Ping() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return DB.Ping()
}

// Close closes the database connection
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if DB != nil {
		err := DB.Close()
		DB = nil
		return err
	}
	return nil
}

// initSchema creates the necessary tables if they don't exist
func initSchema(db *sql.DB) error {
	log.Println("Initializing database schema...")

	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			phone VARCHAR(20) NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'user',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create parking_lots table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS parking_lots (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			address TEXT NOT NULL,
			city VARCHAR(100) NOT NULL,
			state VARCHAR(100) NOT NULL,
			zip_code VARCHAR(20) NOT NULL,
			latitude DECIMAL(10, 6),
			longitude DECIMAL(10, 6),
			total_spaces INT NOT NULL,
			hourly_rate DECIMAL(10, 2) NOT NULL,
			open_time VARCHAR(10),
			close_time VARCHAR(10),
			is_open_24h BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create parking_spaces table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS parking_spaces (
			id UUID PRIMARY KEY,
			parking_lot_id UUID NOT NULL REFERENCES parking_lots(id),
			space_number VARCHAR(20) NOT NULL,
			floor INT,
			type VARCHAR(50) NOT NULL,
			is_occupied BOOLEAN DEFAULT FALSE,
			last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(parking_lot_id, space_number)
		)
	`)
	if err != nil {
		return err
	}

	// Create vehicles table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS vehicles (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id),
			type VARCHAR(50) NOT NULL,
			license_plate VARCHAR(20) NOT NULL,
			brand VARCHAR(50) NOT NULL,
			model VARCHAR(50) NOT NULL,
			year INT NOT NULL,
			color VARCHAR(30) NOT NULL,
			is_default BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(license_plate)
		)
	`)
	if err != nil {
		return err
	}

	// Create reservations table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS reservations (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id),
			parking_lot_id UUID NOT NULL REFERENCES parking_lots(id),
			parking_space_id UUID NOT NULL REFERENCES parking_spaces(id),
			vehicle_id UUID NOT NULL REFERENCES vehicles(id),
			start_time TIMESTAMP WITH TIME ZONE NOT NULL,
			end_time TIMESTAMP WITH TIME ZONE NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			total_cost DECIMAL(10, 2) NOT NULL,
			payment_status VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create payments table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id UUID PRIMARY KEY,
			reservation_id UUID NOT NULL REFERENCES reservations(id),
			amount DECIMAL(10, 2) NOT NULL,
			currency VARCHAR(10) NOT NULL DEFAULT 'USD',
			method VARCHAR(50) NOT NULL,
			status VARCHAR(20) NOT NULL,
			transaction_id VARCHAR(100),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create notifications table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notifications (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id),
			type VARCHAR(50) NOT NULL,
			message TEXT NOT NULL,
			is_read BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Insert sample data for testing if the database is empty
	err = insertSampleData(db)
	if err != nil {
		return err
	}

	log.Println("Database schema initialized successfully")
	return nil
}

// insertSampleData inserts sample data for testing
func insertSampleData(db *sql.DB) error {
	// Check if we already have parking lots
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM parking_lots").Scan(&count)
	if err != nil {
		return err
	}

	// If we already have data, don't insert sample data
	if count > 0 {
		log.Println("Sample data already exists, skipping insertion")
		return nil
	}

	log.Println("Inserting sample data...")

	// Insert sample parking lots
	_, err = db.Exec(`
		INSERT INTO parking_lots (id, name, address, city, state, zip_code, latitude, longitude, total_spaces, hourly_rate, open_time, close_time, is_open_24h, created_at, updated_at)
		VALUES 
		('11111111-1111-1111-1111-111111111111', 'Downtown Parking', '123 Main St', 'Downtown', 'Selangor', '47500', 3.0319924, 101.373358, 50, 2.50, '00:00', '23:59', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
		('22222222-2222-2222-2222-222222222222', 'Shopping Mall Parking', '456 Market Ave', 'Westside', 'Selangor', '47500', 3.0319924, 101.373358, 100, 1.50, '06:00', '22:00', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
		('33333333-3333-3333-3333-333333333333', 'Airport Parking', '789 Airport Rd', 'Eastside', 'Selangor', '47500', 3.0319924, 101.373358, 200, 5.00, '00:00', '23:59', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return err
	}

	// Insert sample parking spaces for each lot
	for i := 1; i <= 20; i++ {
		_, err = db.Exec(`
			INSERT INTO parking_spaces (id, parking_lot_id, space_number, floor, type, is_occupied, last_updated, created_at, updated_at)
			VALUES 
			($1, '11111111-1111-1111-1111-111111111111', $2, 1, 'standard', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, fmt.Sprintf("1111%04d-1111-1111-1111-111111111111", i), fmt.Sprintf("A%d", i))
		if err != nil {
			return err
		}
	}

	for i := 1; i <= 30; i++ {
		_, err = db.Exec(`
			INSERT INTO parking_spaces (id, parking_lot_id, space_number, floor, type, is_occupied, last_updated, created_at, updated_at)
			VALUES 
			($1, '22222222-2222-2222-2222-222222222222', $2, 1, 'standard', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, fmt.Sprintf("2222%04d-2222-2222-2222-222222222222", i), fmt.Sprintf("B%d", i))
		if err != nil {
			return err
		}
	}

	for i := 1; i <= 50; i++ {
		_, err = db.Exec(`
			INSERT INTO parking_spaces (id, parking_lot_id, space_number, floor, type, is_occupied, last_updated, created_at, updated_at)
			VALUES 
			($1, '33333333-3333-3333-3333-333333333333', $2, 1, 'standard', false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, fmt.Sprintf("3333%04d-3333-3333-3333-333333333333", i), fmt.Sprintf("C%d", i))
		if err != nil {
			return err
		}
	}

	log.Println("Sample data inserted successfully")
	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
