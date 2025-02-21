# 构建阶段使用golang镜像
FROM golang AS builder

# 设置Go环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct

# 设置工作目录
WORKDIR /build

# 复制依赖文件并下载依赖
COPY go.mod .
COPY go.sum .
RUN go mod download

# 复制源代码并构建
COPY . .
RUN go build -o app .

# 使用轻量级alpine作为运行时镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /build/app .

# 设置启动命令
CMD ["./app"]