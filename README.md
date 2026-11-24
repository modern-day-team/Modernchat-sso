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
| GET | `/healthz` | 健康检查 |
| POST | `/api/v1/users` | 注册（首个注册者自动成为 admin） |
| POST | `/api/v1/sessions` | 登录，返回 `{token, user}` 并种 `mc_auth` cookie |
| GET | `/api/v1/users/me` | 当前用户（需认证） |
| DELETE | `/api/v1/sessions/current` | 登出（token_version + 1，全部旧 token 失效） |
| GET | `/api/v1/settings` | 站点开关（公开） |
| POST | `/api/v1/email-codes` | 发送邮箱验证码 |
| POST | `/api/v1/password-resets` | 忘记密码（恒 200，不暴露邮箱存在性） |
| POST | `/api/v1/password-resets/:token` | 重置密码（单次令牌，tv bump） |
| PATCH | `/api/v1/users/me/password` | 自身改密（tv bump） |
| GET | `/api/v1/version` | 构建版本（公开） |
| GET | `/api/v1/admin/settings` | 读取完整站点开关（站长；表单回填用，含维护模式与登录锁定参数） |
| PUT | `/api/v1/admin/settings` | 站点开关修改（站长；7 项字段，nil 保持原值，响应返回完整视图） |
| GET | `/api/v1/admin/users` | 用户分页列表（站长） |
| POST | `/api/v1/admin/users/:id/{ban,unban,reset-password,role}` | 用户管理（站长） |
| GET | `/api/v1/admin/update` | 更新状态与最近日志（站长） |
| POST | `/api/v1/admin/update/checks` | 检查新版本 |
| POST | `/api/v1/admin/update` | 发起更新任务 |

上表即为公开 API 契约摘要；完整实现以 `server/internal/router/router.go` 与 `server/internal/handler/` 为准。开发协作约定见本地 `AGENTS.md`（按项目组约定不入库，因此本页不设链接）。错误统一 `{"error": "<中文消息>"}`，JSON camelCase，时间 ISO8601 UTC。

门户的站长后台（`/admin`）提供账号管理与**站点设置**（`/admin/settings`）：登录锁定阈值 0（关闭）或 2-20、锁定时长 1-1440 分钟、维护提示 ≤200 字，前后端校验口径一致。封禁操作使用自研 `UiDialog`，不依赖浏览器原生提示框。

**与主仓库互认**：JWT 口径一致（HS256 / `uid`+`tv` / TTL 7 天）；REST 只认 `Authorization: Bearer`，cookie 仅供 `<img>`。主仓库已于 2026-11-23 按 `FRONTEND_URL` 白名单开启 CORS，并提供扫码登录、会话管理、客户端标识与版本协商——跨端接入详见主仓库 `docs/客户端接入指南.md`。

## 生产部署

`server/migrations/` 提供版本化初始迁移（与 5 个 GORM 模型一致）；`scripts/install.sh` 负责解包 Release、可选执行迁移、安装 `deploy/modernchat-sso@.service` systemd 模板；`deploy/nginx.conf.example` 提供静态门户 + `/api` 反代模板。迁移自动执行需要 `MYSQL_HOST` / `MYSQL_PORT` / `MYSQL_DATABASE` / `MYSQL_USER` / `MYSQL_PASSWORD` 环境变量；凭据不写入仓库。

## 约定

- 开发约定见仓库内 `AGENTS.md`（本地文件，不入库）
- 提交信息：Conventional Commits + 中文描述（`feat: xxx` / `fix: xxx`）
- 仓库强制 LF（`.gitattributes`）
- **严禁把凭据写入仓库**（项目组历史教训）
