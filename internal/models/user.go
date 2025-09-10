// package models 包含应用程序使用的数据模型定义
package models

// User 结构体定义了系统中用户的数据模型
// 用于用户注册、登录和身份验证
type User struct {
	ID       int    `json:"id"`       // 用户唯一标识符，数据库主键
	Username string `json:"username"` // 用户名，用于登录和显示
	Password string `json:"password"` // 密码，存储为bcrypt哈希值，用于身份验证
}
