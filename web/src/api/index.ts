import { http } from './client'
import type { User } from '@/types'

/** SSO 后端 RESTful v1（见仓库根 AGENTS.md 的 API 契约表） */
export const authApi = {
  /** 注册：后端 201 返回用户本体（不含 token），前端随后自动登录 */
  register: (body: {
    username: string
    password: string
    nickname: string
    email?: string
    /** 站点开启邮箱验证时必填（requireEmailVerify） */
    emailCode?: string
  }) =>
    // 公开接口：与登录态无关，不带 Authorization、401 也不清会话
    http.post<User>('/users', body, { skipAuth: true }),
  login: (body: { username: string; password: string }) =>
    http.post<{ token: string; user: User }>('/sessions', body, { skipAuth: true }),
  me: () => http.get<User>('/users/me'),
  /** 登出：后端 token_version+1，全部旧 token 失效 */
  logout: () => http.delete<void>('/sessions/current'),
  /** 发送邮箱验证码：冷却中响应剩余秒数（防枚举，恒 200） */
  sendEmailCode: (email: string) =>
    http.post<{ ok: true; cooldownSeconds: number }>('/email-codes', { email }, { skipAuth: true }),
  /** 忘记密码：恒 200（不暴露邮箱存在性）；重置链接走邮件 */
  forgot: (email: string) =>
    http.post<{ ok: true }>('/password-resets', { email }, { skipAuth: true }),
  /** 重置密码：令牌一次性（15 分钟有效） */
  reset: (token: string, newPassword: string) =>
    http.post<{ ok: true }>(`/password-resets/${encodeURIComponent(token)}`, { newPassword }, { skipAuth: true }),
  /** 自身改密：成功后全端失效（需重新登录） */
  changePassword: (oldPassword: string, newPassword: string) =>
    http.patch<{ ok: true }>('/users/me/password', { oldPassword, newPassword }),
}

/** 站长账号管理（仅 admin 角色；403 由统一错误处理） */
export const adminApi = {
  listUsers: (q: { page?: number; pageSize?: number; kw?: string; role?: string; status?: string }) => {
    const params = new URLSearchParams()
    if (q.page) params.set('page', String(q.page))
    if (q.pageSize) params.set('pageSize', String(q.pageSize))
    if (q.kw) params.set('kw', q.kw)
    if (q.role) params.set('role', q.role)
    if (q.status) params.set('status', q.status)
    const qs = params.toString()
    return http.get<{ users: User[]; total: number; page: number; pageSize: number }>(
      `/admin/users${qs ? `?${qs}` : ''}`,
    )
  },
  ban: (id: number, days: number, reason: string) =>
    http.post<{ ok: true }>(`/admin/users/${id}/ban`, { days, reason }),
  unban: (id: number) => http.post<{ ok: true }>(`/admin/users/${id}/unban`),
  resetPassword: (id: number) => http.post<{ newPassword: string }>(`/admin/users/${id}/reset-password`),
  setRole: (id: number, role: 'user' | 'admin') =>
    http.post<{ ok: true }>(`/admin/users/${id}/role`, { role }),
}

/** 站点设置（完整结构，与后端 handler/settings.go 的 adminView 一一对应） */
export interface SiteSettings {
  allowRegister: boolean
  requireEmailVerify: boolean
  requireInviteCode: boolean
  maintenanceMode: boolean
  maintenanceMessage: string
  /** 0 = 关闭账号级失败锁定 */
  loginMaxAttempts: number
  loginLockMinutes: number
}

/** 站点开关：公开读取用于前端渲染，站长读写用于后台设置页 */
export const settingsApi = {
  get: () =>
    http.get<Pick<SiteSettings, 'allowRegister' | 'requireEmailVerify'>>('/settings'),
  /** 站长读取完整设置（后台表单回填） */
  getAdmin: () => http.get<SiteSettings>('/admin/settings'),
  /** 站长保存设置（字段可选，后端 nil 字段保持原值；此处提交完整表单） */
  update: (body: SiteSettings) => http.put<SiteSettings>('/admin/settings', body),
}
