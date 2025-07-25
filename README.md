# xledger

一个基于 go-micro 框架的微服务记账系统。

## 系统架构

### 核心服务
- `api-service`: API 网关，处理外部请求

### 目录结构
```
.
├── cmd/                        # 各个服务的入口点
│   └── api-service/            # API服务入口
│       └── main.go
|
├── database/                   # 数据库相关
│   │ 
│   ├── migrations/                 # 数据库迁移
│   │   └── 000000_user.down.sql    # 数据库表删除迁移
│   │   └── 000000_user.up.sql      # 数据库表创建迁移
│   │
│   ├── model/                  # 数据模型
│   │   └── user.go             # 用户模型
│   │
│   └── repo/                   # 数据仓储
│       ├── user.go             # 用户仓储接口
│       └── user_impl.go        # 用户仓储实现
│
├── deployments/                # 部署相关配置
│   ├── docker/
│   └── kubernetes/
│
├── document/                   # API 文档
│   └── swagger/
│
├── internal/                   # 内部实现代码
│   │
│   ├── http/                   # HTTP服务层
│   │   ├── handler/            # HTTP处理器
│   │   │   ├── dto/            # 数据传输对象
│   │   │   │   ├── public.go   # 公共数据传输对象
│   │   │   │   └── user.go     # 用户相关数据传输对象
│   │   │   └── user.go         # 用户相关处理器
│   │   │
│   │   ├── service/            # 业务服务层
│   │   │   ├── dto/            # 数据传输对象
│   │   │   └── user.go         # 用户服务接口
│   │   │
│   │   └── server.go           # HTTP服务器配置
│   │
│   └── micro/                  # 微服务相关
│       ├── proto/              # grpc 协议定义
│       └── service/            # 服务实现
│
├── pkg/                        # 可重用的公共包
│   └── common/                 # 通用功能
│
├── config/                     # 配置文件
│   ├── config.yml
│   └── config-uat.yml
│
├── scripts/                    # 部署和维护脚本
│   ├── deploy.sh
│   └── migrate.sh
│
├── bin/              # 编译后的二进制文件
├── Dockerfile       # Docker构建文件
├── docker-compose.yml  # Docker编排文件
├── go.mod           # Go模块定义
└── README.md        # 项目说明文档
```

### 基础设施组件
- MySQL (端口 3314): 主数据库
- Redis (端口 6387): 缓存和会话管理
- Swagger UI (端口 19901): API 文档

## API 文档

项目使用 OpenAPI (Swagger) 规范来描述 API。你可以通过以下方式查看 API 文档：

1. 启动 Swagger UI：
```bash
make swagger-ui
```

2. 访问文档：
打开浏览器访问 http://localhost:9000

Swagger UI 提供了：
- 完整的 API 端点列表
- 请求/响应模型
- 在线 API 测试功能
- API 认证支持

## 技术栈
- Go-Micro v4: 微服务框架
- gRPC: 服务间通信
- MySQL: 数据存储
- Redis: 缓存和会话
- RabbitMQ: 消息队列
- JWT & OAuth2: 认证授权
- Prometheus & Grafana: 监控系统
- Jaeger: 分布式追踪

## 运行要求
- Go 1.16+
- Docker & Docker Compose
- Make

## 快速开始

### 1. 启动基础设施
```bash
# 拉取镜像
make docker-pull

# 启动容器
make docker-up

# 初始化数据库
make init-db
```

### 2. 构建服务
```bash
make build
```

### 3. 运行服务
```bash
# 运行所有服务
make run

# 或者分别运行
make run-api    # API 网关
```

### 4. 监控和管理
```bash
make status     # 检查服务状态
make logs       # 查看服务日志
make mysql-cli  # 访问 MySQL
make redis-cli  # 访问 Redis
```

## 开发工具

### 代码生成和测试
```bash
make proto      # 生成 protobuf 代码
make test       # 运行测试
make clean      # 清理构建
make rebuild    # 重新构建
```

## 配置说明

主要配置文件位于 `config/config.yml`，包含：

### 系统配置
- 服务名称、版本、环境
- HTTP 服务器配置
- 日志配置

### 数据库配置
- MySQL 连接信息
- Redis 连接信息
- 连接池设置

### 消息队列配置
- RabbitMQ 连接信息
- 虚拟主机设置

### 监控配置
- Prometheus 设置
- Grafana 配置
- Jaeger 追踪配置

### OAuth2 配置
支持多个社交平台登录：
- GitHub
- Google
- WeChat
- QQ
- Weibo

### 安全配置
- IP 限制
- 账号绑定限制
- 社交账号同步设置

## 服务通信
- gRPC: 服务间同步通信
- RabbitMQ: 异步消息和事件处理
- Redis: 缓存和会话共享

## 监控和追踪
- Prometheus: 收集性能指标
- Grafana: 指标可视化和告警
- Jaeger: 分布式追踪
- Exporters: 基础设施指标收集

## 帮助命令
```bash
make help       # 显示所有可用命令
```

## 贡献指南
欢迎提交 Issue 和 Pull Request。在提交 PR 前，请确保：
1. 代码通过所有测试
2. 更新相关文档
3. 遵循项目的代码规范

## 许可证
[MIT License](LICENSE)


做一个记账ios软件app，包含登录，注册，创建类目，子类目，账单功能。页面都需要使用appledesign组件优化美化出更现代化的设计风格，软件的名称为xledger。参照这张图片来设计登录页面，但是软件名称为xledger。