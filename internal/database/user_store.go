// package database 包含数据库连接和操作的相关功能
package database

import (
	"database/sql" // 提供SQL数据库接口
	"fmt"          // 格式化输出
	"log"          // 日志记录

	"github.com/sulibao/knowledge/internal/models" // 导入数据模型包

	"golang.org/x/crypto/bcrypt" // 用于密码哈希和验证
)

type UserStore struct {
	db *sql.DB
}

// NewUserStore 创建并返回一个新的UserStore实例
//   - db: 数据库连接对象
//   - *UserStore: 新创建的UserStore实例
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// UpdateUserPassword 更新指定用户的密码
//   - 对新密码进行bcrypt哈希处理
//   - 更新数据库中用户的密码字段
//   - username: 要更新密码的用户名
//   - newPassword: 新的明文密码
//   - error: 如果更新过程中发生错误，返回相应的错误信息
func (s *UserStore) UpdateUserPassword(username, newPassword string) error {
	// 对新密码进行bcrypt哈希处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("在对新密码进行哈希处理时出错: %w", err)
	}

	// 执行SQL更新语句
	_, err = s.db.Exec("UPDATE users SET password = $1 WHERE username = $2", string(hashedPassword), username)
	if err != nil {
		return fmt.Errorf("更新用户密码时出错: %w", err)
	}
	return nil
}

// CreateUser 在数据库中创建新用户
//   - 对用户密码进行bcrypt哈希处理
//   - 将用户信息插入数据库
//   - user: 包含用户信息的User对象
//   - error: 如果创建过程中发生错误，返回相应的错误信息
func (s *UserStore) CreateUser(user *models.User) error {
	// 对用户密码进行bcrypt哈希处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("在对密码进行哈希处理时出错: %w", err)
	}

	// 执行SQL插入语句
	_, err = s.db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("创建用户时出错: %w", err)
	}
	return nil
}

// GetUserByUsername 根据用户名查询用户信息
//   - 从数据库中查询指定用户名的用户信息
//   - username: 要查询的用户名
//   - *models.User: 如果用户存在，返回用户信息；如果用户不存在，返回nil
//   - error: 如果查询过程中发生错误，返回相应的错误信息
func (s *UserStore) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	// 执行SQL查询语句
	err := s.db.QueryRow("SELECT id, username, password FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Password)
	if err == sql.ErrNoRows {
		return nil, nil // 用户不存在
	} else if err != nil {
		return nil, fmt.Errorf("在根据用户名查询用户时出错: %w", err)
	}
	return &user, nil // 返回找到的用户信息
}

// EnsureDefaultAdmin 确保系统中存在默认的管理员用户
//   - 检查是否存在用户名为"admin"的用户
//   - 如果不存在，创建默认管理员用户
//   - 如果已存在，确保密码为默认值
func (s *UserStore) EnsureDefaultAdmin() {
	// 查询是否存在admin用户
	adminUser, err := s.GetUserByUsername("admin")
	if err != nil {
		log.Fatalf("Error checking for default admin: %v", err)
	}

	if adminUser == nil {
		// admin用户不存在，创建默认管理员
		log.Println("Default admin user not found, creating...")
		defaultAdmin := &models.User{
			Username: "admin",
			Password: "admin123", // 这个密码将在CreateUser函数中被哈希处理
		}
		err = s.CreateUser(defaultAdmin)
		if err != nil {
			log.Fatalf("Error creating default admin user: %v", err)
		}
		log.Println("Default admin user 'admin' created successfully.")
	} else {
		// admin用户已存在，确保密码为默认值
		log.Println("Default admin user 'admin' already exists. Ensuring password is 'admin123'...")
		err = s.UpdateUserPassword("admin", "admin123")
		if err != nil {
			log.Fatalf("Error updating default admin password: %v", err)
		}
		log.Println("Default admin user 'admin' password ensured to be 'admin123'.")
	}
}
