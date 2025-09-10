// package config 包含应用程序配置的加载和管理功能
package config

// Config 结构体中定义应用程序的配置信息

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"` // SSL连接模式，如disable、require
	}
	// 这里时关于对象存储minio的相关信息配置
	Minio struct {
		Endpoint        string `yaml:"endpoint"` // MinIO服务端点地址，时需要带端口的，如 192.168.2.190:9000
		AccessKeyID     string `yaml:"accessKeyID"`
		SecretAccessKey string `yaml:"secretAccessKey"`
		UseSSL          bool   `yaml:"useSSL"` // 是否使用SSL连接
		BucketName      string `yaml:"bucketName"`
	}
	// 整个系统的web服务信息
	Server struct {
		Port string `yaml:"port"` // 服务器监听端口，格式参考如":8080"
	}
}
