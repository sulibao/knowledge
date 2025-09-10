// package config 包含应用程序配置的加载和管理功能
package config

import (
	"io/ioutil" // I/O实用工具函数
	"os"        // 操作系统功能接口
	"strconv"   // 字符串和基本数据类型之间的转换
	"strings"   // 字符串操作函数

	"gopkg.in/yaml.v2" // YAML格式的编码和解码，这个项目中主要用来处理yaml格式配置的中间件服务配置信息
)

// 此处LoadConfig的逻辑流程如下：
//   - 首先从指定的YAML配置文件加载默认配置
//   - 然后检查环境变量，如果存在相应的环境变量，则用其值覆盖配置文件中的值
//   - *Config: 加载并可能被环境变量覆盖后的配置对象
//   - error: 如果加载过程中发生错误，返回相应的错误信息

func LoadConfig(path string) (*Config, error) {
	// 首先从配置文件加载默认配置
	config, err := loadFromFile(path)
	if err != nil {
		return nil, err
	}

	// 然后从环境变量读取变量覆盖配置
	overrideFromEnv(config)

	return config, nil
}

// 此处loadFromFile的逻辑流程如下：
//   - 读取指定路径的YAML配置文件
//   - 将YAML内容解析为Config结构体
//   - *Config: 从文件加载的配置对象
//   - error: 如果读取或解析过程中发生错误，返回相应的错误信息
func loadFromFile(path string) (*Config, error) {
	// 读取配置文件内容
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 将YAML内容解析为Config结构体
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// 此处overrideFromEnv的逻辑流程如下：
//   - 检查各种环境变量是否存在
//   - 如果环境变量存在且非空，则用其值覆盖配置对象中的相应字段
//   - config: 要被环境变量覆盖的配置对象
func overrideFromEnv(config *Config) {
	// 数据库的配置信息
	if val := os.Getenv("DB_HOST"); val != "" {
		// 覆盖数据库主机地址
		config.Database.Host = val
	}
	if val := os.Getenv("DB_PORT"); val != "" {
		// 覆盖数据库端口（需要将字符串转换为整数）
		if port, err := strconv.Atoi(val); err == nil {
			config.Database.Port = port
		}
	}
	if val := os.Getenv("DB_USER"); val != "" {
		// 覆盖数据库用户名
		config.Database.User = val
	}
	if val := os.Getenv("DB_PASSWORD"); val != "" {
		// 覆盖数据库密码
		config.Database.Password = val
	}
	if val := os.Getenv("DB_NAME"); val != "" {
		// 覆盖数据库名称
		config.Database.DBName = val
	}
	if val := os.Getenv("DB_SSLMODE"); val != "" {
		// 覆盖数据库SSL模式
		config.Database.SSLMode = val
	}

	// MinIO的配置信息
	if val := os.Getenv("MINIO_ENDPOINT"); val != "" {
		// 覆盖MinIO服务端点
		config.Minio.Endpoint = val
	}
	if val := os.Getenv("MINIO_ACCESS_KEY_ID"); val != "" {
		// 覆盖MinIO访问密钥ID
		config.Minio.AccessKeyID = val
	}
	if val := os.Getenv("MINIO_SECRET_ACCESS_KEY"); val != "" {
		// 覆盖MinIO秘密访问密钥
		config.Minio.SecretAccessKey = val
	}
	if val := os.Getenv("MINIO_USE_SSL"); val != "" {
		// 覆盖MinIO是否使用SSL（将字符串"true"转换为布尔值true）
		config.Minio.UseSSL = strings.ToLower(val) == "true"
	}
	if val := os.Getenv("MINIO_BUCKET_NAME"); val != "" {
		// 覆盖MinIO存储桶名称
		config.Minio.BucketName = val
	}

	// 整个系统服务端的配置信息
	if val := os.Getenv("SERVER_PORT"); val != "" {
		// 确保端口格式，需要以"":port"的格式
		if !strings.HasPrefix(val, ":") {
			val = ":" + val
		}
		// 覆盖服务器端口
		config.Server.Port = val
	}
}
