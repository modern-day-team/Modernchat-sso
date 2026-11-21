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

/** 站点开关（公开）：前端渲染依据 */
export const settingsApi = {
  get: () =>
    http.get<{ allowRegister: boolean; requireEmailVerify: boolean }>('/settings'),
}
