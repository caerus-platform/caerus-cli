# Caerus CLI

包含常用操作的命令行工具，将且不限于包括 marathon / docker / caerus suite 及其他工具。

## Getting Started

在 Release 中下载编译好的命令行工具或者拉取源码后自行编译。

## Usage

caerus-cli 是命令行工具，直接运行后会有相关的命令说明。

- marathon
- docker

### Marathon 相关命令

    # 查看所有应用
    ./caerus-cli m apps
    # 查看相关应用详情
    ./caerus-cli m app /app/id
    # 查看相关应用日志
    ./caerus-cli m app logs /app/id
    # 在容器中运行 shell 命令
    ./caerus-cli m ssh /app/id --key /private/ssh/key/path -c "command"
