// package handlers 包含处理HTTP请求的处理器函数
package handlers

import (
	"encoding/json" // 用于JSON编码和解码
	"log"           // 用于日志记录
	"net/http"      // 提供HTTP客户端和服务器实现

	"github.com/sulibao/knowledge/internal/database"   // 导入数据库操作相关包
	"github.com/sulibao/knowledge/internal/middleware" // 导入中间件相关包
	"github.com/sulibao/knowledge/internal/models"     // 导入数据模型相关包

	"golang.org/x/crypto/bcrypt" // 用于密码哈希和验证
)

// AuthHandler 结构体处理与用户认证相关的HTTP请求
type AuthHandler struct {
	UserStore *database.UserStore // 用户存储接口，用于用户数据的CRUD操作
}

// NewAuthHandler 创建并返回一个新的AuthHandler实例
//   - userStore: 用户存储接口，用于访问用户数据
//   - *AuthHandler: 新创建的AuthHandler实例
func NewAuthHandler(userStore *database.UserStore) *AuthHandler {
	return &AuthHandler{UserStore: userStore}
}

// Register 处理用户注册请求
//   - 解析请求体中的用户信息
//   - 验证用户名和密码是否为空
//   - 检查用户名是否已存在
//   - 创建新用户并返回结果
//   - w: HTTP响应写入器
//   - r: HTTP请求
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// 解析请求体中的用户信息
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		// 请求体解析失败，返回400错误
		http.Error(w, "请求体解析失败", http.StatusBadRequest)
		return
	}

	// 验证用户名和密码是否为空
	if user.Username == "" || user.Password == "" {
		// 用户名或密码为空，返回400错误
		http.Error(w, "用户名和密码不能为空", http.StatusBadRequest)
		return
	}

	// 检查用户名是否已存在
	existingUser, err := h.UserStore.GetUserByUsername(user.Username)
	if err != nil {
		// 数据库查询错误，返回500错误
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	// 如果用户名已存在，返回409冲突错误
	if existingUser != nil {
		http.Error(w, "用户已存在", http.StatusConflict)
		return
	}

	// 创建新用户
	err = h.UserStore.CreateUser(&user)
	if err != nil {
		// 创建用户失败，返回500错误
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	// 注册成功，返回201状态码
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// Login 处理用户登录请求
//   - 解析请求体中的用户登录信息
//   - 验证用户是否存在
//   - 验证密码是否正确
//   - 创建会话并设置认证状态
//   - 返回登录结果
//   - w: HTTP响应写入器
//   - r: HTTP请求
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// 解析请求体中的用户登录信息
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		// 请求体解析失败，返回400错误
		http.Error(w, "请求体解析失败", http.StatusBadRequest)
		return
	}

	// 根据用户名查询用户
	existingUser, err := h.UserStore.GetUserByUsername(user.Username)
	if err != nil {
		// 数据库查询错误，返回500错误
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Internal server error"})
		return
	}

	// 如果用户不存在，返回401未授权错误
	if existingUser == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "登录失败，请检查用户名和密码！"})
		return
	}

	// 调试日志：记录登录尝试的用户名、数据库中的哈希密码和用户提供的明文密码
	log.Printf("Attempting login for user: %s\n", user.Username)
	log.Printf("Hashed password from DB: %s\n", existingUser.Password)
	log.Printf("Plain password from user: %s\n", user.Password)

	// 使用bcrypt比较哈希密码和明文密码
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password))
	if err != nil {
		// 密码不匹配，返回401未授权错误
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "登录失败，请检查用户名和密码！"})
		return
	}

	// 获取会话并设置认证状态
	session, _ := middleware.Store.Get(r, "session-name")
	session.Values["authenticated"] = true             // 设置认证状态为true
	session.Values["username"] = existingUser.Username // 保存用户名到会话
	// 保存会话到响应
	err = session.Save(r, w)
	if err != nil {
		// 记录会话保存错误，但继续处理
		log.Printf("Error saving session: %v\n", err)
	}
	// 记录登录后的会话认证状态
	log.Printf("Session authenticated status after login: %v\n", session.Values["authenticated"])

	// 登录成功，返回200状态码
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "登录成功"})
}
