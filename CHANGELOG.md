# Changelog

## 0.1.0（2026-11-21）

- 项目骨架：Go 1.26 + Gin + GORM（MySQL 8），分层 `config / model / store / service / token / apperr / handler / middleware / router`（高内聚低耦合：业务规则内聚 service，传输层经窄接口依赖，apperr 哨兵解耦错误语义）
- RESTful 认证 API v1：注册、登录、登出、当前用户（`/api/v1` 前缀）
- JWT 与主仓库 modern-chat-new 同口径（HS256，claims `uid`+`tv`，7 天，`mc_auth` cookie），共享 `JWT_SECRET` 可互认
- 前端门户 `web/`：Vue 3 + Vite 8 + TS + Pinia + UnoCSS 原子类 + 自研 `Ui*` 控件层（UiInput/UiButton/UiCheckbox/UiTag，与主仓库逐字对齐），tdesign 仅认证表单；完整 `--td-*`/`--mc-*` 令牌（动效/字号/圆角/阴影/暗色）；dev :5175
- CI：GitHub Actions 双端工作流（server gofmt/vet/build/test -race + 挂 MySQL 服务；web oxlint/eslint/build），push 与 PR 触发
- 邮件发送层：可插拔 `Mailer` 接口，**多厂商故障转移链**——`MAIL_PROVIDER` 逗号分隔多通道（如 `tencent_ses,smtp`），按序降级、错误聚合；内置 console / SMTP / 腾讯云 SES 三通道，其他厂商按同接口扩展
- 在线更新：站长侧 `GET/POST /api/v1/admin/update*`（移植主仓库发行版模式：GitHub Release → sha256 fail-closed → 原子替换 → systemd 重启）；公开 `GET /api/v1/version`；`deploy/modernchat-sso@.service` systemd 模板单元
- 发版工作流：推送 `v*` tag 自动构建 linux amd64/arm64 发行包（含 web/dist）并发布 GitHub Release，版本号 ldflags 注入
- 邮箱验证码流程：`POST /api/v1/email-codes` + 注册强制邮箱验证（requireEmailVerify 默认开、首账号豁免、邮件未配置 503 降级守卫）；6 位码/10 分钟/60s 冷却/错 5 次作废，HMAC-SHA256 落库（key 派生自 JWT_SECRET）
- 登录安全加固：认证端点 IP 限流 10/min（内存令牌桶）+ 账号级失败锁定（按用户名 5 次锁 15 分钟，429 带剩余分钟）
- 忘记/重置密码：`POST /api/v1/password-resets(/:token)`，32 字节令牌 15 分钟单次（事务内原子消费防 TOCTOU），tv bump 全端失效；自身改密 `PATCH /api/v1/users/me/password`（SSO 新增）
- 站长账号管理：`GET /api/v1/admin/users` + 封禁/解封/重置密码/改角色（封 admin 403、最后站长保护、重置返回新密码）；AdminView（用户表格+操作）
- 站点开关体系：system_settings 表 + `GET/PUT /api/v1/(admin/)settings`（AllowRegister/RequireEmailVerify/维护模式/登录锁定参数）+ 维护模式 503 门禁（登录豁免）
- 存储层单测：emailcode 冷却/一次性/作废、login_attempt 锁定/清零、password_reset 单次消费（CI MySQL 实连）
- 本地开发环境：docker-compose MySQL 8（3307）、`.env.example`、仓库强制 LF
