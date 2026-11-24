# Changelog

## [未发布]

- 站长后台补齐站点设置页（`/admin/settings`）：注册开关（允许注册 / 邀请码 / 邮箱验证）、维护模式与提示文案、登录失败锁定阈值与时长，共 7 项；保存即时生效
- 新增 `GET /api/v1/admin/settings` 返回完整开关（此前公开 GET 与 PUT 仅回 2 个字段，站长无法读取其余 5 项）；PUT 响应同步改为完整视图
- 补齐站点设置的服务端校验：维护提示 ≤200 字、失败阈值 0（关闭）或 2-20、锁定时长 1-1440 分钟（对齐主仓库口径）
- 新增自研 `UiDialog` 控件（v-model:visible / Esc / 遮罩关闭 / 焦点与滚动管理），封禁操作改用对话框一次填写天数与原因，移除 `window.prompt`
- 修复站点设置读不到的隐藏 bug：`SettingStore.Get` 裸写 `Where("key = ?")`，`key` 是 MySQL 8.0 保留字导致 SQL 语法错误，所有设置读取静默回退代码默认值（表现为保存后值不生效）；改用结构体条件让 GORM 转义列名，并补往返单测防回归
- 门户视觉重做：`App.vue` 加品牌头 / 背景光晕与网格 / 页脚，登录页加品牌标识、副标题与辅助入口；所有页面统一 elevation 层次
- 修复输入框在白卡片上隐形的问题：`UiInput` 改 `bg-page` + 新 `line-strong` 边框档（UnoCSS theme 同步新增），不再与卡片同色
- 补齐 `UiSwitch` / `UiInputNumber` 两个自研控件（对齐主仓库手感：spring 回弹 / 令牌动效）；站点设置页改用开关 + 数字步进，按三组卡片组织，加未保存提示、维护模式警示、加载骨架
- 账号管理页重做：概览统计卡、带图标搜索栏、用户头像列、表头底色与行 hover、空态、封禁对话框警示条
- 中英文文案同步（对话框通用按钮、封禁标题、站点设置全字段、后台统计与提示）
- 补齐认证层回归测试：JWT 签发/解析往返、错误密钥与篡改 token 拒绝、Auth 中间件 Bearer/cookie 回退/401/403/500 分支、Admin 角色校验 403
- 补齐配置与路由测试：JWT 密钥门禁（空/短拒绝）、开发默认值、环境变量覆盖（多厂商邮件/模板 ID/布尔开关）、路由装配（公开端点免鉴权、受保护端点 401、未知路径 404）

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
