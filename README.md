# silent-sign-backend

无声之韵智能手语双向翻译系统后端，基于 [go-zero](https://github.com/zeromicro/go-zero) 框架构建。

## 技术栈

| 组件       | 技术选型                          |
| ---------- | --------------------------------- |
| 框架       | go-zero v1.10 + GORM v1.31       |
| 数据库     | MySQL 8.x                        |
| 缓存       | Redis 7.x（go-redis v9）         |
| 认证       | JWT（golang-jwt v5）             |
| 邮件       | SMTP（go-gomail）                |
| IP 定位    | ip2region v4/v6                  |
| 密码加密   | golang.org/x/crypto（bcrypt）    |
| API 定义   | goctl API DSL                    |

## 项目结构

```
silent-sign-backend/
├── silent_sign.api          # API 定义文件（DSL）
├── silent_sign.go           # 入口文件
├── cmd/
│   └── cron/                # 定时任务（点赞同步等）
├── common/
│   └── errorx/              # 统一错误处理
├── data/                    # 静态数据（ip2region 库）
├── docs/                    # Swagger 文档
├── etc/                     # 配置文件
│   ├── silent_sign.yaml         # 生产配置（不提交 Git）
│   ├── silent_sign.dev.yaml     # 开发配置（不提交 Git）
│   └── silent_sign.example.yaml # 示例配置
├── internal/
│   ├── config/              # 配置结构体
│   ├── constvar/            # 常量定义
│   ├── handler/             # 路由处理器
│   ├── logic/               # 业务逻辑层
│   ├── middleware/          # JWT / IP 中间件
│   ├── pkg/                 # 工具包（respx 等）
│   ├── repo/                # 数据访问层（GORM + Redis）
│   ├── sse/                 # SSE 推送
│   ├── svc/                 # ServiceContext（DI 容器）
│   ├── types/               # 请求/响应类型（goctl 生成）
│   └── ws/                  # WebSocket
├── Dockerfile
├── docker-compose.yaml
└── docker-compose.example.yaml
```

## 快速开始

### 1. 准备配置

```bash
cp etc/silent_sign.example.yaml etc/silent_sign.dev.yaml
# 编辑 etc/silent_sign.dev.yaml，填入本地 MySQL / Redis 连接信息
```

### 2. 启动依赖服务

```bash
docker compose -f docker-compose.example.yaml up -d mysql redis
```

### 3. 运行服务

```bash
go run silent_sign.go -f etc/silent_sign.dev.yaml
```

服务默认监听 `0.0.0.0:8888`。

## API 模块

| 模块    | 前缀              | 中间件            | 说明                     |
| ------- | ----------------- | ----------------- | ------------------------ |
| Auth    | `/api/v1/auth`    | IPMiddleware      | 注册/登录/刷新令牌       |
| User    | `/api/v1/user`    | JWT + IP          | 用户信息/修改密码        |
| Moment  | `/api/v1/moment`  | JWT + IP          | 动态创建/列表/详情       |
| Comment | `/api/v1/comment` | JWT + IP          | 评论创建/列表            |
| Like    | `/api/v1/like`    | JWT + IP          | 点赞/取消（Redis-First） |
| Message | `/api/v1/message` | JWT + IP          | 私信发送/列表/已读       |
| Friend  | `/api/v1/friend`  | JWT + IP          | 好友申请/列表/删除       |
| SSE     | `/api/v1/sse`     | JWT               | 实时事件推送             |

## 代码生成

```bash
# 从 API 定义生成 handler / logic / types
goctl api go -api silent_sign.api -dir . --style go_zero

# 生成 Swagger 文档
goctl api swagger -api silent_sign.api -dir ./docs -filename swagger
```

## 构建

```bash
# 构建 API 服务
go build -o silent_sign .

# 构建定时任务
go build -o cron ./cmd/cron/
```

## 定时任务

```bash
# K8s 一次性执行模式
go run ./cmd/cron/ sync_like

# docker-compose 守护模式
go run ./cmd/cron/ sync_like --daemon --interval=60s
```

## Docker 部署

### 构建并推送镜像

```bash
docker buildx build --platform linux/amd64 \
  -t crpi-u5azhs6neq326bz0.cn-hangzhou.personal.cr.aliyuncs.com/yub_lu/silent_sign:0.0.1 \
  --push .
```

> Mac M 系列必须指定 `--platform linux/amd64`。

### 服务器部署

```bash
# 1. 创建目录并放入配置文件和 IP 定位库
mkdir -p ~/silent-sign/service/silent-sign/{etc,data}
#   - etc/silent_sign.yaml
#   - data/ip2region_v4.xdb
#   - data/ip2region_v6.xdb

# 2. 启动服务
docker compose up -d

# 3. 更新服务
docker compose pull && docker compose up -d
```

## 架构分层

```
API 请求 → Handler（解析参数）→ Logic（业务逻辑）→ Repo（DB/Redis）
```

- **Handler**：解析请求，调用 Logic，写入响应。
- **Logic**：核心业务逻辑，事务管理，权限校验。
- **Repo**：数据访问抽象，GORM 模型 + Redis 操作。

详细架构说明见 [Copilot Instructions](./.github/copilot-instructions.md)。