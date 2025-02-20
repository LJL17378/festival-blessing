# 我们使用这个镜像是为了减少最终镜像的大小
FROM golang AS builder

# 设置 Go 构建环境变量
# GO111MODULE=on 启用 Go 模块支持，确保 Go 项目使用 go.mod 管理依赖
# CGO_ENABLED=0 禁用 Cgo，这样可以确保生成的应用是静态链接的，避免依赖系统的 C 库
# GOOS=linux 设置目标操作系统为 Linux
# GOARCH=amd64 设置目标架构为 amd64
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct


# 设置工作目录为 /build，所有后续命令将在该目录下执行
WORKDIR /build

# 将 Go 模块的依赖文件 go.mod 和 go.sum 复制到容器的工作目录中
COPY go.mod .
COPY go.sum .

# 运行 go mod download 下载 Go 项目的依赖
RUN go mod download

# 复制整个项目到容器的工作目录中
COPY . .

# 编译 Go 应用，并指定输出文件为 app
RUN go build -o app .

FROM ubuntu

# 将构建阶段中生成的 app 可执行文件复制到最终镜像中
COPY --from=builder /build/app .

# 修改 app 可执行文件的权限，确保容器启动时能执行它
RUN chmod 755 app

# 设置环境变量，表示构建 ID 或标识符
ENV BUILD_ID=dontKillMe

# 设置容器启动时的默认命令，nohup 用于让应用在后台运行
ENTRYPOINT ["./app"]