/** 与后端 model.User 的 JSON 形状一致（camelCase；PasswordHash/TokenVersion 不下发） */
export interface User {
  id: number
  username: string
  nickname: string
  avatarColor: string
  email?: string
  role: 'user' | 'admin'
  bannedUntil?: string
  banReason?: string
  createdAt: string
}
