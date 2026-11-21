# Changelog

## 0.1.0（2026-11-21）

- 项目骨架：Go 1.26 + Gin + GORM（MySQL 8），分层 `config / model / store / service / token / apperr / handler / middleware / router`（高内聚低耦合：业务规则内聚 service，传输层经窄接口依赖，apperr 哨兵解耦错误语义）
- RESTful 认证 API v1：注册、登录、登出、当前用户（`/api/v1` 前缀）
- JWT 与主仓库 modern-chat-new 同口径（HS256，claims `uid`+`tv`，7 天，`mc_auth` cookie），共享 `JWT_SECRET` 可互认
- 前端门户 `web/`：Vue 3 + Vite 8 + TS + Pinia + UnoCSS 原子类 + 自研 `Ui*` 控件层（UiInput/UiButton/UiCheckbox/UiTag，与主仓库逐字对齐），tdesign 仅认证表单；完整 `--td-*`/`--mc-*` 令牌（动效/字号/圆角/阴影/暗色）；dev :5175
- CI：GitHub Actions 双端工作流（server gofmt/vet/build/test -race + 挂 MySQL 服务；web oxlint/eslint/build），push 与 PR 触发
- 邮件发送层：可插拔 `Mailer` 接口，**多厂商故障转移链**——`MAIL_PROVIDER` 逗号分隔多通道（如 `tencent_ses,smtp`），按序降级、错误聚合；内置 console / SMTP / 腾讯云 SES 三通道，其他厂商按同接口扩展
- 本地开发环境：docker-compose MySQL 8（3307）、`.env.example`、仓库强制 LF
