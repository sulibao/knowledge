package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/sulibao/knowledge/internal/config"

	_ "github.com/lib/pq"
)

func InitPostgres(cfg *config.Config) (*sql.DB, error) {
	// 首先连接到默认的postgres数据库，用于检查knowledge_base是否存在
	connStrDefault := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.SSLMode)

	dbDefault, err := sql.Open("postgres", connStrDefault)
	if err != nil {
		return nil, fmt.Errorf("error opening default database: %w", err)
	}
	defer dbDefault.Close()

	// 检查knowledge_base数据库是否存在
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = dbDefault.QueryRow(query, cfg.Database.DBName).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("error checking if database exists: %w", err)
	}

	// 如果数据库不存在，则创建它
	if !exists {
		log.Printf("Database '%s' does not exist, creating it...", cfg.Database.DBName)
		_, err = dbDefault.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.Database.DBName))
		if err != nil {
			return nil, fmt.Errorf("error creating database: %w", err)
		}
		log.Printf("Database '%s' created successfully.", cfg.Database.DBName)
	} else {
		log.Printf("Database '%s' already exists.", cfg.Database.DBName)
	}

	// 现在连接到knowledge_base数据库
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL!")
	return db, nil
}

func CreateTables(db *sql.DB) error {
	createUsersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL
	);
	`

	_, err := db.Exec(createUsersTableSQL)
	if err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}

	log.Println("Users table checked/created successfully.")
	return nil
}
