package main

import (
	"fmt"
	"log"
	"net/http"

	"go-pro/0602/internal/config"
	"go-pro/0602/internal/database"
	"go-pro/0602/internal/handlers"
	"go-pro/0602/internal/middleware"
	minio_client "go-pro/0602/internal/minio"

	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("Starting knowledge base system...")

	// Load configuration from ./config.yaml
	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize PostgreSQL for store user messages
	db, err := database.InitPostgres(cfg)
	if err != nil {
		log.Fatalf("Error initializing PostgreSQL: %v", err)
	}
	defer db.Close()

	// Create tables in PostgreSQL if they don't exist
	err = database.CreateTables(db)
	if err != nil {
		log.Fatalf("Error creating tables: %v", err)
	}

	// Initialize UserStore and ensure default admin
	userStore := database.NewUserStore(db)
	userStore.EnsureDefaultAdmin()

	// Initialize MinIO with bucket from ./config.yaml.bucketName
	minioClient, err := minio_client.InitMinio(cfg)
	if err != nil {
		log.Fatalf("Error initializing MinIO: %v", err)
	}
	_ = minioClient // Use minioClient to avoid unused variable error for now

	r := mux.NewRouter()

	// Serve static files from the 'public' directory
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	// Authentication routes
	authHandler := handlers.NewAuthHandler(userStore)
	r.HandleFunc("/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")

	// Server register page
	r.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/register.html")
	}).Methods("GET")

	// Server register.html directly
	r.HandleFunc("/register.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/register.html")
	}).Methods("GET")

	// File management routes config
	fileHandler := handlers.NewFileHandler(minioClient, cfg)
	protectedRouter := r.PathPrefix("/api").Subrouter()
	protectedRouter.Use(middleware.AuthRequired)

	protectedRouter.HandleFunc("/upload", fileHandler.UploadFile).Methods("POST")
	protectedRouter.HandleFunc("/files", fileHandler.ListFiles).Methods("GET")
	protectedRouter.HandleFunc("/download", fileHandler.DownloadFile).Methods("GET")
	protectedRouter.HandleFunc("/delete", fileHandler.DeleteFile).Methods("DELETE")

	// Protected dashboard route config
	r.Handle("/dashboard", middleware.AuthRequired(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/dashboard.html")
	}))).Methods("GET")

	// Logout route config
	r.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		session, _ := middleware.Store.Get(r, "session-name")
		session.Values["authenticated"] = false
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusFound)
	}).Methods("POST")

	// Server login page
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/login.html")
	}).Methods("GET")

	// Server index page
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusFound)
	}).Methods("GET")

	fmt.Printf("Server listening on %s\n", cfg.Server.Port)
	log.Fatal(http.ListenAndServe(cfg.Server.Port, r))
}
