// package middleware 包含HTTP请求处理的中间件组件
package middleware

import (
	"net/http" // 提供HTTP客户端和服务器实现

	"github.com/gorilla/sessions" // 提供cookie和文件系统会话存储
)

var (
	key = []byte("super-secret-key") // 用于加密和验证会话cookie的密钥
	// Store 是全局会话存储，用于管理用户会话
	Store *sessions.CookieStore
)

// init 初始化会话存储并配置cookie选项
// 在包被导入时自动执行
func init() {
	// 使用密钥创建新的cookie存储
	Store = sessions.NewCookieStore(key)
	// 配置cookie选项
	Store.Options.HttpOnly = true                     // 防止JavaScript访问cookie，增强安全性
	Store.Options.Secure = false                      // 是否仅通过HTTPS发送cookie，生产环境中应设为true
	Store.Options.SameSite = http.SameSiteDefaultMode // 控制第三方网站请求时cookie的发送策略，可根据生产需求调整
}

// AuthRequired 是一个中间件函数，用于验证用户是否已认证
// 如果用户未认证，将重定向到登录页面
//   - next: 下一个要执行的HTTP处理器
//   - http.Handler: 包装后的HTTP处理器
func AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从请求中获取会话
		session, _ := Store.Get(r, "session-name")

		// 检查用户是否已认证
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			// 未认证，重定向到登录页面
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		// 用户已认证，继续处理请求
		next.ServeHTTP(w, r)
	})
}
