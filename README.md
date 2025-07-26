# Knowledge Base 应用 Docker 和 Kubernetes 部署指南

## 环境变量配置

本应用支持通过环境变量配置所有参数，以下是可用的环境变量列表：

### 数据库配置

- `DB_HOST`: PostgreSQL 主机地址
- `DB_PORT`: PostgreSQL 端口
- `DB_USER`: PostgreSQL 用户名
- `DB_PASSWORD`: PostgreSQL 密码
- `DB_NAME`: PostgreSQL 数据库名
- `DB_SSLMODE`: PostgreSQL SSL 模式

### MinIO 配置

- `MINIO_ENDPOINT`: MinIO 服务端点
- `MINIO_ACCESS_KEY_ID`: MinIO 访问密钥 ID
- `MINIO_SECRET_ACCESS_KEY`: MinIO 秘密访问密钥
- `MINIO_USE_SSL`: 是否使用 SSL 连接 MinIO
- `MINIO_BUCKET_NAME`: MinIO 存储桶名称

### 服务器配置

- `SERVER_PORT`: 服务器监听端口

## Docker 部署

### 构建 Docker 镜像

```bash
docker build -t knowledge-base:latest .
```

### 运行 Docker 容器

```bash
docker run -d \
  --name knowledge-base \
  -p 8080:8080 \
  -e DB_HOST=192.168.2.190 \
  -e DB_PORT=35432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=SLBpg2025 \
  -e DB_NAME=knowledge_base \
  -e MINIO_ENDPOINT=192.168.2.190:39000 \
  -e MINIO_ACCESS_KEY_ID=admin \
  -e MINIO_SECRET_ACCESS_KEY=admin@2025 \
  -e MINIO_BUCKET_NAME=knowledge-bucket \
  registry.cn-chengdu.aliyuncs.com/su03/knowledge_base:2572602
```