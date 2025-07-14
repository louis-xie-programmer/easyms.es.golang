# easyms.es.golang 微服务系统(代码扳手微信公众号文章案例)

[![Go Version](https://img.shields.io/badge/Go-1.23.2-blue.svg)](https://golang.org/dl/) [![License](https://img.shields.io/github/license/user/repo)](LICENSE)

## 项目简介
高性能微服务系统，集成 Elasticsearch、gRPC 和 HTTP/3 技术，适用于高并发场景下的数据检索与处理需求。支持 RESTful API 与 gRPC 混合服务架构。
以下是当前相关的博文，结合博文，可快速了解本系统的核心功能和实现原理：
- [Go+gRPC项目Swagger集成：3步搞定API文档自动化](https://mp.weixin.qq.com/s/UeoWib45B3eNwHNE9hmMiA)
- [Go + gRPC + HTTP/3：解锁下一代高性能通信](https://mp.weixin.qq.com/s/yEekCcjI3Jp4B7gumk8W5w)
- [Golang 的多任务调度系统：从 BaseJob 到 ProductJob 的"泛型"与"继承"实现](https://mp.weixin.qq.com/s/L-u_Ad9NZFO131NmV54jQQ)
- [Golang异步日志实战：通道+中间件的完美组合](https://mp.weixin.qq.com/s/BrWwJnEnyYE4wc4xvENTTQ)
- [从186ms到79ms！用fasthttp给Elasticsearch客户端打“鸡血”的实战笔记](https://mp.weixin.qq.com/s/vYYVuAAiwR6UH62Vo208vw)

后续，我们将围绕 Go、.NET、JavaScript/Node.js 构建的多语言微服务，以及基于 Elasticsearch 的搜索场景，系统拆解微服务架构与大中型搜索方案的设计与落地。干货持续更新，敬请关注「代码扳手」微信公众号。：

![wx.jpg](conf/wx.jpg)

## 核心特性
- **高性能网络协议**：支持 HTTP/2、HTTP/3 协议，采用 fasthttp 优化 Elasticsearch 请求
- **多协议服务**：同时支持 gRPC 服务通信和 RESTful API 接口
- **定时任务调度**：基于 cron/v3 实现的分布式任务调度系统
- **可视化管理**：包含 Vue.js 开发的任务管理界面
- **多数据库支持**：集成 Redis、SQLite、SQL Server 等多种数据存储方案

## 技术架构
### 设计模式
- 工厂模式：用于创建数据库连接池（如 Redis、SQL Server）
- 中间件模式：在 Gin 中使用 logger、认证等中间件
- 策略模式：不同 Job 类型通过接口统一调度
- 代理模式：gRPC-Gateway 自动生成 REST 到 gRPC 的代理层

### 主要组件
```
├── api/              # 主服务入口（HTTP/gRPC）
├── crob_job/         # 定时任务模块（含 Vue 管理界面）
├── protos/           # gRPC 接口定义及 Swagger 输出
├── service/          # 业务逻辑实现（价格、产品搜索）
├── fasthttp/         # 自定义 fasthttp 客户端封装
├── client/           # HTTP/3 客户端示例及 Swagger 工具
├── conf/             # 配置文件（YAML）
├── certs/            # 服务端和客户端证书
├── config/           # 配置加载模块
├── model/            # 数据模型定义
├── utility/          # 工具函数（字符串、类型判断等）
├── db/               # 数据库连接模块（Redis、SQLite、SQL Server）
└── easyes/           # Elasticsearch 操作封装

```

## 技术选型
| 组件 | 技术 |
|------|------|
| 后端框架 | Go 1.23.2 + Gin v1.10.0 |
| 远程调用 | gRPC v1.65.0 |
| HTTP 客户端 | fasthttp v1.57.0 |
| 网络协议 | quic-go v0.46.0 (HTTP/3) |
| 配置管理 | viper |
| 数据库 ORM | GORM v1.25.1 |
| Elasticsearch | go-elasticsearch/v8 |

## 快速开始
### 环境准备
```bash
# 初始化模块下载
go mod download

# 设置环境变量
go env -w GO111MODULE=on
export PATH=$PATH:$(go env GOPATH)/bin  # Linux/Mac
set PATH=%PATH%;%GOPATH%\bin          # Windows
```

### 构建项目
```bash
# 构建 Windows 可执行文件
GOOS=windows GOARCH=amd64 go build -o bin/easyjob.exe crob_job/job.go

# 构建 Linux 可执行文件
GOOS=linux GOARCH=amd64 go build -o bin/easyapi api/main.go
GOOS=linux GOARCH=amd64 go build -o bin/swagger client/swagger/main.go
```

### 运行服务
```bash
# 启动 API 服务
./bin/easyapi

# 启动定时任务
./bin/easyjob

# 启动 Swagger 文档服务
./bin/swagger
```

## 核心模块详解
### 1. gRPC 服务端 (api/) 
- 基于 Gin 框架和 gRPC 实现混合服务
- 支持 HTTP/2 和 HTTP/3 协议
- 使用 fasthttp 优化 Elasticsearch 请求
- 包含完整的 TLS 安全配置
- 实现优雅关闭机制
- 提供详细的日志记录和性能监控

### 2. 定时任务系统 (crob_job/)
- 基于 cron/v3 实现的分布式任务调度
- 支持多种任务类型：
  - 产品数据同步任务
  - 价格数据同步任务
  - Redis 缓存维护任务
  - 监控任务
- 提供 RESTful API 接口进行任务管理
- 包含 Vue.js 开发的可视化管理界面
- 支持动态调整任务执行周期
- 实现任务暂停/恢复功能

### 3. gRPC 客户端案例 (client/)
包含三个完整的客户端实现案例：

#### a. 标准 gRPC 客户端
- 实现基本的 gRPC 通信
- 支持双向 TLS 认证
- 包含元数据传递（Client-ID 和 Client-Secret）
- 提供完整的错误处理机制

#### b. HTTP/3 客户端
- 使用 quic-go 实现的 HTTP/3 客户端
- 支持 QUIC 协议
- 包含完整的 TLS 配置
- 实现自定义 dialer
- 提供连接复用机制

#### c. Swagger 文档服务
- 基于 gRPC-Gateway 实现 RESTful API 转换
- 集成 Swaggo 生成交互式文档
- 支持 TLS 加密
- 包含自定义拦截器实现
- 提供元数据传递功能

`go mod download` # 下载所有依赖
```
go env -w GO111MODULE=on # 环境变量
export PATH=$PATH:$(go env GOPATH)/bin   # 环境变量设置, 注意这里各系统会有差异,建议windows直接在系统中添加golang的path
```


# protc 命令, 注意proto对应的目录, 在根目录执行
```
protoc -I . \
--go_out . --go_opt paths=source_relative \
--go-grpc_out . --go-grpc_opt paths=source_relative \
--grpc-gateway_out . \
--grpc-gateway_opt paths=source_relative \
--grpc-gateway_opt generate_unbound_methods=true \
--openapiv2_out . --openapiv2_opt logtostderr=true \
protos/messages/*.proto protos/services/search.proto
```

# buf grpc 命令
buf lint
buf generate


# 注意事项:
1. 本案例中实现了自定义的http3实现及对应的客户端案例
2. 本案例中优化了es请求的http, 改用fasthttp, 通过自定义Transport实现了tls的支持及对相关参数进行优化
3. ubuntu 发布后权限赋值 : chmod 777 api/esgrpc/easyapi

## 证书生成指南

### 生成CA证书和域名证书
```bash
# 创建证书目录
cd easyms-es && mkdir -p certs

# 生成CA私钥和证书
cd certs
openssl genrsa -out ca.key 4096
openssl req -new -x509 -days 365 -key ca.key -out ca.crt -subj "/C=CN/ST=Beijing/L=Beijing/O=EasyMS/OU=CA/CN=easy.dev"

# 生成服务器私钥和CSR
openssl genrsa -out server.key 4096
openssl req -new -key server.key -out server.csr -subj "/C=CN/ST=Beijing/L=Beijing/O=EasyMS/OU=Server/CN=easy.dev"

# 创建openssl.cnf配置文件
cat > openssl.cnf << EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name

[req_distinguished_name]

[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = easy.dev
IP.1 = 127.0.0.1
EOF

# 生成服务器证书
openssl x509 -req -days 365 -in server.csr \
  -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -extfile openssl.cnf -extensions v3_req

# 生成客户端证书
openssl genrsa -out client.key 4096
openssl req -new -key client.key -out client.csr -subj "/C=CN/ST=Beijing/L=Beijing/O=EasyMS/OU=Client/CN=easy.dev"
openssl x509 -req -days 365 -in client.csr \
  -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out client.crt

# 验证证书
openssl x509 -in server.crt -text -noout
```

### 证书使用说明
1. 服务端配置：
   - 证书路径: `certs/server.crt`
   - 私钥路径: `certs/server.key`
2. 客户端配置：
   - CA证书: `certs/ca.crt`
   - 客户端证书: `certs/client.crt`
   - 客户端私钥: `certs/client.key`

### 证书验证方法
```bash
# 检查证书有效期
openssl x509 -in certs/server.crt -dates -noout

# 检查证书域名
openssl x509 -in certs/server.crt -text -noout | grep DNS
```