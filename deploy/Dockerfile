# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 设置代理，加速依赖下载
ENV GOPROXY=https://goproxy.cn,direct

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -o nurture-api internal/main.go

# Run stage
FROM alpine:latest

WORKDIR /app

# 安装基础工具（可选，如需调试可保留）
RUN apk --no-cache add ca-certificates tzdata

# 从构建阶段复制二进制文件
COPY --from=builder /app/nurture-api .

# 复制配置文件（默认配置，实际运行时建议挂载）
COPY internal/etc/local.yaml internal/etc/local.yaml

# 暴露端口
EXPOSE 9000

CMD ["./nurture-api"]
