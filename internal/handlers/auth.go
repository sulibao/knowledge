package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/sulibao/knowledge/internal/database"
	"github.com/sulibao/knowledge/internal/middleware"
	"github.com/sulibao/knowledge/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	UserStore *database.UserStore
}

func NewAuthHandler(userStore *database.UserStore) *AuthHandler {
	return &AuthHandler{UserStore: userStore}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Password == "" {
		http.Error(w, "Username and password cannot be empty", http.StatusBadRequest)
		return
	}

	existingUser, err := h.UserStore.GetUserByUsername(user.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	if existingUser != nil {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	err = h.UserStore.CreateUser(&user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	existingUser, err := h.UserStore.GetUserByUsername(user.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	if existingUser == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "登录失败，请检查用户名和密码！"})
		return
	}

	// Debugging: Log the hashed password from DB and plain password from user
	log.Printf("Attempting login for user: %s\n", user.Username)
	log.Printf("Hashed password from DB: %s\n", existingUser.Password)
	log.Printf("Plain password from user: %s\n", user.Password)

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "登录失败，请检查用户名和密码！"})
		return
	}

	session, _ := middleware.Store.Get(r, "session-name")
	session.Values["authenticated"] = true
	session.Values["username"] = existingUser.Username
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Error saving session: %v\n", err)
	}
	log.Printf("Session authenticated status after login: %v\n", session.Values["authenticated"])

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "登录成功"})
}
