import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { authApi } from '@/api'
import { getToken, onUnauthorized, setToken } from '@/api/client'
import type { User } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(getToken())
  const user = ref<User | null>(null)

  const isLoggedIn = computed(() => !!token.value)

  function applyAuth(payload: { token: string; user: User }) {
    token.value = payload.token
    setToken(payload.token)
    user.value = payload.user
  }

  async function login(username: string, password: string) {
    applyAuth(await authApi.login({ username, password }))
  }

  /** 注册成功后自动登录（后端注册接口只返回用户本体，不含 token） */
  async function register(
    username: string,
    password: string,
    nickname: string,
    email?: string,
    emailCode?: string,
  ) {
    await authApi.register({ username, password, nickname, email, emailCode })
    await login(username, password)
  }

  async function fetchMe() {
    if (!token.value) return
    try {
      user.value = await authApi.me()
    } catch {
      /* 401 已由 client 统一清理并回登录页；其余错误就地吞掉，刷新页面可重试 */
    }
  }

  async function logout() {
    try {
      await authApi.logout()
    } finally {
      // 无论接口成败（含 token 已失效），本地会话一律清干净
      token.value = ''
      setToken('')
      user.value = null
    }
  }

  // 401 统一处理：同步内存态，避免 store 仍是「已登录」的假象
  onUnauthorized(() => {
    token.value = ''
    user.value = null
  })

  return { token, user, isLoggedIn, login, register, fetchMe, logout }
})
