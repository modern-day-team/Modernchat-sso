# Modernchat-sso

Modernchat 项目组的 SSO（单点登录）认证服务。

## 技术栈

Go 1.26 + Gin + GORM（MySQL 8）+ golang-jwt/v5（HS256）。RESTful API（`/api/v1` 前缀），JWT 口径与主仓库 [modern-chat-new](https://github.com/modern-day-team/modern-chat-new) 完全一致：claims `uid` + `tv`（token_version），TTL 7 天，`Authorization` 头与 `mc_auth` cookie 双通道——两个服务共享 `JWT_SECRET` 即可互相校验 token。

## 快速开始

```bash
# 1. 起本地 MySQL（映射 3307，避免与主仓库 3306 冲突）
docker compose up -d

# 2. 配置环境变量
cd server && cp .env.example .env   # 填入 JWT_SECRET

# 3. 运行（启动期 AutoMigrate 自动建表）
go run ./cmd/server

# 4. 前端门户（可选，另开终端）
cd ../web && npm install && npm run dev   # :5175，/api 已代理到 :8090
```

服务默认监听 `:8090`，健康检查 `GET /healthz`。

## API（v1，RESTful）

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/v1/users` | 注册（首个注册者自动成为 admin） |
| POST | `/api/v1/sessions` | 登录，返回 token 并种 `mc_auth` cookie |
| GET | `/api/v1/users/me` | 当前用户（需认证） |
| DELETE | `/api/v1/sessions/current` | 登出（token_version + 1，全部旧 token 失效） |

错误统一 `{"error": "<中文消息>"}`；JSON 字段 camelCase，时间 ISO8601 UTC。

## 约定

- 开发约定见仓库内 `AGENTS.md`（本地文件，不入库）
- 提交信息：Conventional Commits + 中文描述（`feat: xxx` / `fix: xxx`）
- 仓库强制 LF（`.gitattributes`）
- **严禁把凭据写入仓库**（项目组历史教训）
