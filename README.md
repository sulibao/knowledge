# Knowledge Base 应用部署指南

## 环境变量配置

支持手动修改./config.yaml和声明变量的形式进行获取变量值

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

## golang环境运行

```golang
// 修改./config.yaml变量文件
go run main.go
```

## Docker 部署

### 构建 Docker 镜像

```bash
docker buildx build --provenance=false --push --tag registry_address/knowledge_base:tag --platform linux/amd64,linux/arm64 .
```

### 运行 Docker 容器

```bash
# 变量由命令行传参声明
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
  -p 30080:8080 \
  registry_address/knowledge_base:tag
```

## K8S/Chart部署

### 修改变量文件

```yaml
# 修改./helm-chart/knowledge-values.yaml变量文件，确认好镜像版本等信息
```

### 通过helm进行安装knowledge

```bash
helm -n ns install knowledge -f knowledge-values.yaml knowledge-base-0.1.4.tgz
```