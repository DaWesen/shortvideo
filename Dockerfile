# 基础镜像
FROM golang:1.25.1-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
ARG SERVICE=gateway
RUN go build -o main ./cmd/${SERVICE}

# 使用轻量级镜像
FROM alpine:latest

# 安装必要的依赖
RUN apk add --no-cache ca-certificates

# 设置工作目录
WORKDIR /app

# 复制构建产物
COPY --from=builder /app/main .

# 复制配置文件
COPY configs/ ./configs/

# 暴露端口
ARG PORT=8080
EXPOSE ${PORT}

# 启动命令
CMD ["./main"]
