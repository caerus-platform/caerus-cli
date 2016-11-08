# Caerus CLI

包含常用操作的命令行工具，将且不限于包括 marathon / docker / caerus suite 及其他工具。

## Getting Started

在 Release 中下载编译好的命令行工具或者拉取源码后自行编译。

## Usage

caerus-cli 是命令行工具，直接运行后会有相关的命令说明。

- marathon
- docker

### Config 相关命令

    # init 新建配置文件
    ./caerus-cli config init
    
    # info/i 打印配置信息
    ./caerus-cli config info

### Marathon 相关命令

> need to set marathon_host in ~/.caerus/config.yml

    # apps 查看所有应用
    ./caerus-cli m apps
    
    # app info/i 查看相关应用详情
    ./caerus-cli m app info /app/id
    
    # app logs 查看相关应用日志
    ./caerus-cli m app logs /app/id
    
    # app update/n 更新应用镜像
    ./caerus-cli m app u /app/id --image /image/repo
    
    # app update/n 强制更新应用镜像
    ./caerus-cli m app u /app/id --image /image/repo -f
    
    # app restart/r 重启应用
    ./caerus-cli m app r /app/id
    
    # app restart/r 强制重启应用
    ./caerus-cli m app r /app/id -f
    
    # app scale/s 启动应用
    ./caerus-cli m app s /app/id -n 1
    
    # app scale/s 停止应用
    ./caerus-cli m app s /app/id -n 0
    
    # app scale/s 强制停止应用
    ./caerus-cli m app s /app/id -n 0 -f
    
    # 在容器中运行 shell 命令
    ./caerus-cli m ssh /app/id --key /private/ssh/key/path -c "command"

### Docker 相关命令

    # docker containers/c 查看所有容器
    ./caerus-cli d c docker-host
    
    # docker logs 查看容器日志
    ./caerus-cli d logs docker-host container-id-or-name
    
    # docker ssh 在容器中执行命令
    ./caerus-cli d ssh docker-host container-id-or-name -c "bash" --key ~/.ssh/id_rsa -u root

### RabbitMQ 相关命令

> need to set marathon_host in ~/.caerus/config.yml

    # subscribe/s 订阅, -d 表示 durable
    ./caerus-cli mq s -x caerus.events -d -t topic "#"
    ./caerus-cli mq s -x direct_logs -t direct "info"

## Develop

    go fmt ./... && gofmt -s -w . && \
    go vet ./... && go get ./... && \
    go test ./... && \
    golint ./... && \
    gocyclo -avg -over 15 . && \
    errcheck ./...
