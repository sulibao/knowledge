FROM registry.cn-chengdu.aliyuncs.com/su03/golang:1.23.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM registry.cn-chengdu.aliyuncs.com/su03/alpine:latest
RUN apk --no-cache update && \
    apk --no-cache add \
        ca-certificates curl vim bash-completion busybox-extras && \
        rm -rf /var/cache/apk/* 
WORKDIR /root/
COPY --from=builder /app/main .
# 配置文件作为默认配置，可以被环境变量覆盖
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/public/ ./public/

# 设置环境变量默认值
ENV DB_HOST="xxx" \
    DB_PORT="xxx" \
    DB_USER="xxx" \
    DB_PASSWORD="xxx" \
    DB_NAME="xxx" \
    DB_SSLMODE="xxx" \
    MINIO_ENDPOINT="xxx" \
    MINIO_ACCESS_KEY_ID="xxx" \
    MINIO_SECRET_ACCESS_KEY="xxx" \
    MINIO_USE_SSL="false" \
    MINIO_BUCKET_NAME="xxx" \
    SERVER_PORT="xxx"

EXPOSE 8080

CMD ["./main"]