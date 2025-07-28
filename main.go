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

	// 从./config.yaml中加载变量
	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// 初始化PostgreSQL，用于存储用户的数据
	db, err := database.InitPostgres(cfg)
	if err != nil {
		log.Fatalf("Error initializing PostgreSQL: %v", err)
	}
	defer db.Close()

	// 在PostgreSQL中创建用户表
	err = database.CreateTables(db)
	if err != nil {
		log.Fatalf("Error creating tables: %v", err)
	}

	// 在PostgreSQL中新建admin用户
	userStore := database.NewUserStore(db)
	userStore.EnsureDefaultAdmin()

	// 初始化Minio中的存储bucket
	minioClient, err := minio_client.InitMinio(cfg)
	if err != nil {
		log.Fatalf("Error initializing MinIO: %v", err)
	}
	_ = minioClient // Use minioClient to avoid unused variable error for now

	r := mux.NewRouter()

	// 从public目录加载静态资源
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	// 用户认证路由
	authHandler := handlers.NewAuthHandler(userStore)
	r.HandleFunc("/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")

	// 用户注册页面
	r.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/register.html")
	}).Methods("GET")

	r.HandleFunc("/register.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/register.html")
	}).Methods("GET")

	// 文件管理的路径
	fileHandler := handlers.NewFileHandler(minioClient, cfg)
	protectedRouter := r.PathPrefix("/api").Subrouter()
	protectedRouter.Use(middleware.AuthRequired)

	protectedRouter.HandleFunc("/upload", fileHandler.UploadFile).Methods("POST")
	protectedRouter.HandleFunc("/files", fileHandler.ListFiles).Methods("GET")
	protectedRouter.HandleFunc("/download", fileHandler.DownloadFile).Methods("GET")
	protectedRouter.HandleFunc("/delete", fileHandler.DeleteFile).Methods("DELETE")

	// 登录后的页面
	r.Handle("/dashboard", middleware.AuthRequired(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/dashboard.html")
	}))).Methods("GET")

	// 退出登录的路由
	r.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		session, _ := middleware.Store.Get(r, "session-name")
		session.Values["authenticated"] = false
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusFound)
	}).Methods("POST")

	// 登录页面
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/login.html")
	}).Methods("GET")

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusFound)
	}).Methods("GET")

	fmt.Printf("Server listening on %s\n", cfg.Server.Port)
	log.Fatal(http.ListenAndServe(cfg.Server.Port, r))
}
