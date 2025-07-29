package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/sulibao/knowledge/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) UpdateUserPassword(username, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing new password: %w", err)
	}

	_, err = s.db.Exec("UPDATE users SET password = $1 WHERE username = $2", string(hashedPassword), username)
	if err != nil {
		return fmt.Errorf("error updating user password: %w", err)
	}
	return nil
}

func (s *UserStore) CreateUser(user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	_, err = s.db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}
	return nil
}

func (s *UserStore) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow("SELECT id, username, password FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Password)
	if err == sql.ErrNoRows {
		return nil, nil // User not found
	} else if err != nil {
		return nil, fmt.Errorf("error getting user by username: %w", err)
	}
	return &user, nil
}

func (s *UserStore) EnsureDefaultAdmin() {
	adminUser, err := s.GetUserByUsername("admin")
	if err != nil {
		log.Fatalf("Error checking for default admin: %v", err)
	}

	if adminUser == nil {
		log.Println("Default admin user not found, creating...")
		defaultAdmin := &models.User{
			Username: "admin",
			Password: "admin123", // This will be hashed by CreateUser
		}
		err = s.CreateUser(defaultAdmin)
		if err != nil {
			log.Fatalf("Error creating default admin user: %v", err)
		}
		log.Println("Default admin user 'admin' created successfully.")
	} else {
		log.Println("Default admin user 'admin' already exists. Ensuring password is 'admin123'...")
		err = s.UpdateUserPassword("admin", "admin123")
		if err != nil {
			log.Fatalf("Error updating default admin password: %v", err)
		}
		log.Println("Default admin user 'admin' password ensured to be 'admin123'.")
	}
}
