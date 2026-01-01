# Stage 1: Build stage
FROM golang:1.24-alpine AS builder

# 安装必要的包和工具
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

ENV GOPROXY=https://goproxy.cn,direct

# 下载依赖包（利用Docker层缓存）
RUN go mod download



# 复制源代码
COPY . .

# 更新模块
RUN go mod tidy

# 构建应用程序
# 禁用CGO以获得静态二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o nurture internal/main.go

# Stage 2: Production stage
FROM alpine:latest

# 安装必要的包（包括 curl 用于健康检查）
RUN apk --no-cache add ca-certificates tzdata curl

# 创建非root用户
RUN adduser -D -g '' appuser

# 设置工作目录
WORKDIR /app

# 复制构建的二进制文件
COPY --from=builder /app/nurture .

# 复制配置文件模板（以备不时之需，或者作为默认配置）
COPY --from=builder /app/internal/etc ./internal/etc

# 如果没有 local.yaml，则从 template.yaml 复制一个
RUN if [ ! -f internal/etc/local.yaml ]; then cp internal/etc/template.yaml internal/etc/local.yaml; fi

# 确保二进制文件可执行
RUN chmod +x /app/nurture

# 更改文件所有者
RUN chown -R appuser:appuser /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查 (使用 /ping 接口)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/ping || exit 1

# 运行应用程序
CMD ["./nurture"]
