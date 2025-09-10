// package database 包含数据库连接和操作的相关功能
package database

import (
	"database/sql" // SQL数据库接口
	"fmt"          // 格式化输出
	"log"          // 日志记录

	"github.com/sulibao/knowledge/internal/config"

	_ "github.com/lib/pq" // PostgreSQL驱动程序，使用下划线导入表示仅初始化驱动
)

// InitPostgres 初始化PostgreSQL数据库连接并确保目标数据库存在
//   - 首先连接到PostgreSQL默认数据库
//   - 检查并创建应用程序所需的数据库（如果不存在）
//   - 建立与应用程序数据库的连接
//   - cfg: 包含数据库配置信息的配置对象
//   - *sql.DB: 数据库连接对象
//   - error: 如果连接过程中发生错误，返回相应的错误信息
func InitPostgres(cfg *config.Config) (*sql.DB, error) {
	// 先连接到PG数据库默认创建的postgres数据库，检查knowledge_base是否已经存在
	connStrDefault := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.SSLMode)

	// 打开与默认数据库的连接
	dbDefault, err := sql.Open("postgres", connStrDefault)
	if err != nil {
		return nil, fmt.Errorf("打开默认数据库postgres时出错: %w", err)
	}
	defer dbDefault.Close() // 确保函数结束时关闭默认数据库连接

	// 检查knowledge_base数据库是否已经存在
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)" // 查询系统表检查数据库是否存在
	err = dbDefault.QueryRow(query, cfg.Database.DBName).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("检查数据库是否存在时出错: %w", err)
	}

	// 数据库不存在时进行创建
	if !exists {
		log.Printf("Database '%s' does not exist, creating it...", cfg.Database.DBName)
		// 执行CREATE DATABASE语句创建数据库
		_, err = dbDefault.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.Database.DBName))
		if err != nil {
			return nil, fmt.Errorf("创建数据库时出错: %w", err)
		}
		log.Printf("Database '%s' created successfully.", cfg.Database.DBName)
	} else {
		log.Printf("Database '%s' already exists.", cfg.Database.DBName)
	}

	// 连接到knowledge_base数据库
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	// 打开与应用程序数据库的连接
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("在打开数据库时产生错误: %w", err)
	}

	// 测试数据库连接是否成功
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("连接数据库时出错: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL!")
	return db, nil // 返回数据库连接对象
}

// CreateTables 在数据库中创建必要的表结构（如果不存在）
//   - 创建用户表，用于存储用户信息
//   - db: 数据库连接对象
//   - error: 如果创建表过程中发生错误，返回相应的错误信息
func CreateTables(db *sql.DB) error {
	createUsersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,                    -- 用户ID，自增主键
		username VARCHAR(50) UNIQUE NOT NULL,     -- 用户名，唯一且非空
		password VARCHAR(255) NOT NULL            -- 密码哈希，非空
	);
	`

	// 执行创建表的SQL语句
	_, err := db.Exec(createUsersTableSQL)
	if err != nil {
		return fmt.Errorf("创建用户表时发生错误: %w", err)
	}

	log.Println("Users table checked/created successfully.")
	return nil
}
