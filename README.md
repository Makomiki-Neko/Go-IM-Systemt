# Go-IM-Systemt
基于 GoZero 微服务架构构建的分布式即时通讯(IM)系统，集成 LLM 接入能力，同时支持基础即时通讯业务与智能化对话场景。

![登录](./login.jpeg)
![主页](./main.jpeg)
![会话](./chat.jpeg)
![智能体](./ai.jpeg)

## 项目简介
本项目采用 **Gateway / API / RPC** 三层微服务架构搭建分布式 IM 系统，完整实现即时通讯核心业务，同时接入大模型能力，提供 AI 智能体对话、群聊消息总结、群机器人等扩展功能。
适用于私有聊天服务、企业内部通讯、AI 对话机器人集成等场景。

## 技术栈
### 后端
- **开发语言**：Golang
- **微服务框架**：go-zero
- **通信协议**：gRPC、HTTP、WebSocket
- **ORM**：GORM
- **数据库**：MySQL
- **缓存**：Redis
- **消息队列**：RabbitMQ
- **分布式文件存储**：SeaweedFS-S3
- **LLM API**：OpenAI Chat Completions
- **服务发现**：ETCD
- **容器化部署**：Docker
### Web
- **VUE**
- **JS**

```mermaid
flowchart LR
    Client[Web / 其他客户端] -->|HTTP + JWT| API[API :8888]
    Client <-->|WebSocket + JWT| GW[Gateway :8889]
    API --> User[User RPC :9001]
    API --> Relation[Relation RPC :9002]
    API --> Chat[Chat RPC :9003]
    GW --> Chat
    GW --> AI[AI RPC :9004]
    User --> DB[(MySQL)]
    Relation --> DB
    Chat --> DB
    AI --> DB
    User --> Redis[(Redis)]
    Relation --> Redis
    Chat --> Redis
    AI --> Redis
    GW --> Redis
    Relation -->|群 Outbox / 好友事件| MQ[RabbitMQ]
    Chat -->|群 Outbox| MQ
    GW -->|旧私聊链路| MQ
    AI <-->|任务及结果 Outbox| MQ
    MQ --> GW
    AI --> LLM[模型服务]
    API -->|头像上传 Filer| FS[SeaweedFS]
    GW -->|生成 S3 PUT 签名| FS
    Client -->|文件直传| FS
```

ETCD 提供 RPC 注册与发现，未在图中展开。各进程默认共享 MySQL `gozero`；服务边界目前不等于数据库隔离边界。

## 架构分层
- 客户端
-    ↓
- Gateway 网关层 （连接管理、消息路由、长连接维护、心跳检测）
-    ↓
- API 服务层 （http 接口、参数校验、鉴权、请求转发）
-    ↓
- RPC 微服务层 （用户、关系、聊天、AI ）

| Service            | Entry                       | Type              | Port  | etcd key      |
|--------------------|-----------------------------|-------------------|-------|---------------|
| `rpc/user`         | `rpc/user/user.go`          | gRPC              | 9001  | `user.rpc`    |
| `rpc/relation`     | `rpc/relation/relation.go`  | gRPC + MQ consumer| 9002  | `relation.rpc`|
| `rpc/chat`         | `rpc/chat/chat.go`          | gRPC              | 9003  | `chat.rpc`    |
| `rpc/ai`           | `rpc/ai/ai.go`              | gRPC + MQ consumer| 9004  | `ai.rpc`      |
| `api`              | `api/api.go`                | REST (HTTP)       | 8888  | —             |
| `gateway`          | `gateway/gateway.go`        | REST + WebSocket  | 8889  | —             |
