# AGENTS.md — IMM (Instant Messaging)

## Project layout

```
Backend/    Go + GoZero v1.10.2 + Gorm v1.31.1 + gRPC microservices
Client/     Vue 3.4 + Vite 5 + Pinia 2.1 + TypeScript 5.5 (unfinished)
```

Group chat is **not implemented** — ignore group-related protos, handlers, and API routes.

## Backend — services & ports

| Service            | Entry                       | Type              | Port  | etcd key      |
|--------------------|-----------------------------|-------------------|-------|---------------|
| `rpc/user`         | `rpc/user/user.go`          | gRPC              | 9001  | `user.rpc`    |
| `rpc/relation`     | `rpc/relation/relation.go`  | gRPC + MQ consumer| 9002  | `relation.rpc`|
| `rpc/chat`         | `rpc/chat/chat.go`          | gRPC              | 9003  | `chat.rpc`    |
| `api`              | `api/api.go`                | REST (HTTP)       | 8888  | —             |
| `gateway`          | `gateway/gateway.go`        | REST + WebSocket  | 8889  | —             |
| `rpc/ai`           | *(empty directory)*         | —                 | —     | —             |

## Infrastructure

Everything must be running locally before starting any service:

| Component   | Address             |
|-------------|---------------------|
| etcd        | `127.0.0.1:12379`   |
| MySQL       | `127.0.0.1:3306`, DB `gozero`, user `root` / `123456` |
| Redis       | `127.0.0.1:6379`, no password |
| RabbitMQ    | `127.0.0.1:5672` (`amqp://guest:guest@127.0.0.1:5672/`) |
| SeaweedFS   | Master `localhost:9333`, Filer `localhost:8890`, S3 GW `localhost:8333` |

## Running services

Run RPC services first, then API/Gateway. All use `-f` to point at their YAML config:

```bash
# 1. RPCs
go run rpc/user/user.go    -f etc/user.yaml
go run rpc/relation/relation.go -f etc/relation.yaml
go run rpc/chat/chat.go    -f etc/chat.yaml

# 2. API & Gateway
go run api/api.go          -f etc/user-api.yaml
go run gateway/gateway.go  -f etc/gateway-api.yaml
```

- No Makefile, Dockerfile, or docker-compose exists.
- No `*_test.go` files anywhere in Backend.
- Two standalone manual test scripts at Backend root: `rabbitMQTest.go`, `SeaweedFSS3Test.go`.

## GoZero conventions

Services follow the GoZero scaffold pattern:
```
<svc>/
  <svc>.proto          protobuf definition (RPCs only)
  <svc>.go             main entry
  etc/<svc>.yaml       config
  <pb_pkg>/            generated protobuf Go code
  client/              generated gRPC client wrapper
  internal/
    config/config.go
    svc/servicecontext.go
    server/<name>server.go
    logic/<name>logic.go
```
For REST services (`api/`, `gateway/`), handlers live in `internal/handler/` instead of `server/`.

## Key backend details

- **JWT**: secret `"your-secret-key-2026-never-leak"`, expiry 3600s, refresh threshold 90s. Login returns a single `refresh_token`; heartbeat (`/api/user/heart`) refreshes the `access_token`.
- **Snowflake IDs**: machine IDs — user=1, relation=2, chat=3.
- **Duplicate models**: `common/models/` has the canonical GORM models. Some RPCs also have local copies under `rpc/<svc>/model/` (legacy, prefer `common/models/`).
- **RabbitMQ**: relation RPC declares exchange `im.events`; chat RPC declares exchange `im.chat`. Gateway has push queues `im.gateway.push.chat` and `im.gateway.push.event`. `Project.md` documents the MQ flow.
- **Redis keys**: `im:chat:unread:<userId>` (unread count hash), `offline:<type>:<userId>` (offline message queue), `user:status:<uid>` (online heartbeat sensor with 60s TTL; expiry triggers last-online MySQL update).
- **Auth failure**: returns Chinese error `{"code": 401, "msg": "认证失败"}`.
- **Common helpers**: `common/pkg/snowflake.go`, `common/pkg/rabbitMQ.go`, `common/pkg/redisStorage.go`, `common/pkg/user_unit.go` (bcrypt).

## Client — key facts

```bash
cd Client
npm run dev      # Vite dev server on port 3000, proxies /api → localhost:8888
npm run build    # vue-tsc typecheck then vite build
```

- **Friend APIs** use `application/x-www-form-urlencoded` (not JSON) — see `api/friend.ts` `toForm()` helper. The Go backend expects `form` tags.
- **Avatars**: relative paths resolve to `http://localhost:8890/<photo>` via `utils/avatar.ts`.
- **Optimistic messages**: sent messages get negative temp IDs (`-(Date.now() + random)`), marked `_sending: true`. After sending, the client auto-pulls `GetNewPrivateMsg` to get the server-confirmed message and replaces it. Failed messages get `_failed: true`.
- **ACK**: chat store fires `CustomEvent('im:ack')` every 5s; HomeView listens and sends `ack.PrivateMsgRead` via WebSocket.
- **PulledSet**: `friendStore.pulledSet` tracks which users' new messages have been fetched this WS connection. Reset on reconnect.
- **WebSocket**: connects to `ws://localhost:8889/ws?token=<access_token>`, heartbeat every 25s, auto-reconnect 3s delay.
- **HTTP heartbeat**: separate timer refreshes `access_token` via `/api/user/heart` every 180s (3 min).
- **No auth route guards** on Vue Router — auth check happens in `HomeView.vue` `onMounted`.
- **File uploads**: request S3 presigned URL via WS (`updateFile.<type>`), PUT file directly to SeaweedFS S3 gateway, then send a regular message with the `fileId` URL as content.
- **Client is unfinished** — treat code here as work-in-progress.

## Project docs

`Project.md` has the authoritative protocol spec (WebSocket message types, chat flow, pull strategy, Redis key patterns). Trust it over code when they conflict.
