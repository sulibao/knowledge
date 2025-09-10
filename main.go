// ------------------------------------------------------------
// @Author: sulibao(cn-chengdu)
// @Email: sulibao2003@163.com
// @Last Modified by: sulibao(cn-chengdu)
// @Description: If you have any questions, please contact me at the above email address.
// ------------------------------------------------------------

// package main 定义了应用程序的入口点
package main

// 导入所需的包
import (
	"fmt"      // 用于格式化输出
	"log"      // 用于日志记录
	"net/http" // 提供HTTP客户端和服务器实现

	// 导入内部包
	"github.com/sulibao/knowledge/internal/config"             // 配置管理
	"github.com/sulibao/knowledge/internal/database"           // 数据库操作
	"github.com/sulibao/knowledge/internal/handlers"           // HTTP请求处理器
	"github.com/sulibao/knowledge/internal/middleware"         // 中间件
	minio_client "github.com/sulibao/knowledge/internal/minio" // MinIO客户端

	"github.com/gorilla/mux" // 用于HTTP路由管理的第三方包
)

// main 函数是程序的入口点，负责初始化和启动整个知识库系统
func main() {
	// 打印启动信息
	fmt.Println("Starting knowledge base system...")

	// 从./config.yaml中加载配置变量
	// LoadConfig函数读取配置文件并解析为Config结构体
	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		// 如果配置加载失败，记录错误并终止程序
		log.Fatalf("Error loading configuration: %v", err)
	}

	// 初始化PostgreSQL数据库连接，用于存储用户数据
	// InitPostgres函数根据配置建立数据库连接
	db, err := database.InitPostgres(cfg)
	if err != nil {
		// 如果数据库初始化失败，记录错误并终止程序
		log.Fatalf("Error initializing PostgreSQL: %v", err)
	}
	// 确保在程序结束时关闭数据库连接
	defer db.Close()

	// 在PostgreSQL中创建必要的数据表（如用户表）
	// CreateTables函数检查表是否存在，不存在则创建
	err = database.CreateTables(db)
	if err != nil {
		// 如果表创建失败，记录错误并终止程序
		log.Fatalf("Error creating tables: %v", err)
	}

	// 创建用户存储服务并确保默认管理员用户存在
	// NewUserStore函数创建一个用于操作用户数据的服务
	userStore := database.NewUserStore(db)
	// EnsureDefaultAdmin函数确保系统中存在默认的admin用户
	// 如果不存在则创建，如果存在则确保密码正确
	userStore.EnsureDefaultAdmin()

	// 初始化MinIO对象存储服务及其存储桶
	// InitMinio函数连接到MinIO服务并确保所需的存储桶存在
	minioClient, err := minio_client.InitMinio(cfg)
	if err != nil {
		// 如果MinIO初始化失败，记录错误并终止程序
		log.Fatalf("Error initializing MinIO: %v", err)
	}
	// 临时使用变量以避免未使用变量的编译错误
	_ = minioClient // 注：实际上下面的代码会使用这个变量

	// 创建HTTP路由器，用于处理所有HTTP请求
	r := mux.NewRouter()

	// 配置静态资源服务
	// 将/public/路径下的请求映射到./public目录中的文件
	// StripPrefix移除URL前缀，FileServer提供静态文件服务
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	// 配置用户认证相关的路由
	// 创建认证处理器，负责处理用户注册和登录
	authHandler := handlers.NewAuthHandler(userStore)
	// 注册POST请求处理函数，处理用户注册
	r.HandleFunc("/register", authHandler.Register).Methods("POST")
	// 注册POST请求处理函数，处理用户登录
	r.HandleFunc("/login", authHandler.Login).Methods("POST")

	// 配置用户注册页面的路由
	// 处理对/register路径的GET请求，返回注册页面
	r.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		// ServeFile函数将指定文件的内容发送到客户端
		http.ServeFile(w, r, "./public/register.html")
	}).Methods("GET")

	// 处理对/register.html路径的GET请求，同样返回注册页面
	// 这是为了支持直接访问register.html的URL
	r.HandleFunc("/register.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/register.html")
	}).Methods("GET")

	// 配置文件管理相关的路由
	// 创建文件处理器，负责处理文件上传、下载等操作
	fileHandler := handlers.NewFileHandler(minioClient, cfg)
	// 创建需要认证的子路由器，所有/api前缀的请求都需要认证
	protectedRouter := r.PathPrefix("/api").Subrouter()
	// 使用认证中间件，确保只有已登录用户才能访问这些API
	protectedRouter.Use(middleware.AuthRequired)

	// 注册文件上传API，处理POST请求
	protectedRouter.HandleFunc("/upload", fileHandler.UploadFile).Methods("POST")
	// 注册文件列表API，处理GET请求
	protectedRouter.HandleFunc("/files", fileHandler.ListFiles).Methods("GET")
	// 注册文件下载API，处理GET请求
	protectedRouter.HandleFunc("/download", fileHandler.DownloadFile).Methods("GET")
	// 注册文件删除API，处理DELETE请求
	protectedRouter.HandleFunc("/delete", fileHandler.DeleteFile).Methods("DELETE")

	// 配置登录后的仪表盘页面路由
	// 使用认证中间件包装处理函数，确保只有已登录用户才能访问仪表盘
	r.Handle("/dashboard", middleware.AuthRequired(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 返回仪表盘页面
		http.ServeFile(w, r, "./public/dashboard.html")
	}))).Methods("GET")

	// 配置退出登录的路由
	r.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		// 获取用户会话
		session, _ := middleware.Store.Get(r, "session-name")
		// 将认证状态设置为false，表示用户已退出登录
		session.Values["authenticated"] = false
		// 保存会话状态
		session.Save(r, w)
		// 重定向到登录页面
		http.Redirect(w, r, "/login", http.StatusFound)
	}).Methods("POST")

	// 配置登录页面的路由
	// 处理对/login路径的GET请求，返回登录页面
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/login.html")
	}).Methods("GET")

	// 配置根路径的路由
	// 当用户访问网站根目录时，重定向到登录页面
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusFound)
	}).Methods("GET")

	// 打印服务器启动信息，显示监听的端口
	fmt.Printf("Server listening on %s\n", cfg.Server.Port)
	// 启动HTTP服务器
	// ListenAndServe函数会阻塞当前goroutine
	// 如果服务器因错误而停止，log.Fatal会记录错误并终止程序
	log.Fatal(http.ListenAndServe(cfg.Server.Port, r))
}
