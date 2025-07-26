package config

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// LoadConfig loads configuration from file and overrides with environment variables if present
func LoadConfig(path string) (*Config, error) {
	// 首先从配置文件加载默认配置
	config, err := loadFromFile(path)
	if err != nil {
		return nil, err
	}

	// 然后从环境变量覆盖配置
	overrideFromEnv(config)

	return config, nil
}

// loadFromFile loads configuration from YAML file
func loadFromFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// overrideFromEnv overrides configuration with environment variables
func overrideFromEnv(config *Config) {
	// Database configuration
	if val := os.Getenv("DB_HOST"); val != "" {
		config.Database.Host = val
	}
	if val := os.Getenv("DB_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.Database.Port = port
		}
	}
	if val := os.Getenv("DB_USER"); val != "" {
		config.Database.User = val
	}
	if val := os.Getenv("DB_PASSWORD"); val != "" {
		config.Database.Password = val
	}
	if val := os.Getenv("DB_NAME"); val != "" {
		config.Database.DBName = val
	}
	if val := os.Getenv("DB_SSLMODE"); val != "" {
		config.Database.SSLMode = val
	}

	// MinIO configuration
	if val := os.Getenv("MINIO_ENDPOINT"); val != "" {
		config.Minio.Endpoint = val
	}
	if val := os.Getenv("MINIO_ACCESS_KEY_ID"); val != "" {
		config.Minio.AccessKeyID = val
	}
	if val := os.Getenv("MINIO_SECRET_ACCESS_KEY"); val != "" {
		config.Minio.SecretAccessKey = val
	}
	if val := os.Getenv("MINIO_USE_SSL"); val != "" {
		config.Minio.UseSSL = strings.ToLower(val) == "true"
	}
	if val := os.Getenv("MINIO_BUCKET_NAME"); val != "" {
		config.Minio.BucketName = val
	}

	// Server configuration
	if val := os.Getenv("SERVER_PORT"); val != "" {
		// 确保端口格式正确（以冒号开头）
		if !strings.HasPrefix(val, ":") {
			val = ":" + val
		}
		config.Server.Port = val
	}
}
