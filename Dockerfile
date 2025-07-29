FROM registry.cn-chengdu.aliyuncs.com/su03/golang:1.23.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM registry.cn-chengdu.aliyuncs.com/su03/alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .
# 配置文件作为默认配置，可以被环境变量覆盖
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/public/ ./public/

# 设置环境变量默认值
ENV DB_HOST="192.168.2.190" \
    DB_PORT="35432" \
    DB_USER="postgres" \
    DB_PASSWORD="SLBpg2025" \
    DB_NAME="knowledge_base" \
    DB_SSLMODE="disable" \
    MINIO_ENDPOINT="192.168.2.190:39000" \
    MINIO_ACCESS_KEY_ID="admin" \
    MINIO_SECRET_ACCESS_KEY="admin@2025" \
    MINIO_USE_SSL="false" \
    MINIO_BUCKET_NAME="knowledge-bucket" \
    SERVER_PORT="8080"

EXPOSE 8080

CMD ["./main"]
