/** REST 客户端：统一前缀 /api/v1、Bearer 鉴权、错误归一（沿用主仓库 client 模式）。 */

const BASE = '/api/v1'
const TOKEN_KEY = 'sso_token'

/**
 * 请求超时：没有超时的话，代理/后端任一环节挂住时 fetch 会永久 pending，
 * 页面上的 loading 也就永远不会停。
 */
const REQUEST_TIMEOUT = 15_000

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) ?? ''
}

export function setToken(token: string) {
  if (token) localStorage.setItem(TOKEN_KEY, token)
  else localStorage.removeItem(TOKEN_KEY)
}

/**
 * 401 处理钩子：由 auth store 注册，用来同步清掉 Pinia 里的 token/user。
 * 只清 localStorage 会让内存中的 store 仍是「已登录」，路由守卫与页面状态会失真。
 */
let unauthorizedHandler: (() => void) | null = null

export function onUnauthorized(fn: () => void) {
  unauthorizedHandler = fn
}

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE'
  body?: unknown
  /**
   * 公开接口（登录/注册）：不携带 Authorization 头，401 不清理登录态、
   * 不跳登录页——失败由调用方就地展示错误（票据失效 ≠ 会话失效）。
   */
  skipAuth?: boolean
}

/** 401 统一回登录页，并带上来源：会话中途过期后还能回到原页面。 */
function redirectToLogin() {
  if (location.pathname === '/login') return
  const from = location.pathname + location.search
  location.href = from === '/' ? '/login' : `/login?redirect=${encodeURIComponent(from)}`
}

function buildUrl(path: string): string {
  return BASE + path
}

/** 强制超时包装：Promise.race 保证调用方一定在超时时间内拿到结果。 */
function withTimeout<T>(task: Promise<T>, ms: number, message: string, ctrl: AbortController): Promise<T> {
  let timer: ReturnType<typeof setTimeout> | undefined
  const timeout = new Promise<never>((_, reject) => {
    timer = setTimeout(() => {
      ctrl.abort()
      reject(new ApiError(0, message))
    }, ms)
  })
  return Promise.race([task, timeout]).finally(() => {
    if (timer) clearTimeout(timer)
  })
}

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const token = opts.skipAuth ? '' : getToken()
  const ctrl = new AbortController()
  const run = async (): Promise<T> => {
    let res: Response
    try {
      res = await fetch(buildUrl(path), {
        method: opts.method ?? 'GET',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: opts.body === undefined ? undefined : JSON.stringify(opts.body),
        signal: ctrl.signal,
      })
    } catch (e) {
      // fetch 只在网络层失败时抛 TypeError：包装成可读中文，避免把英文原文直接弹给用户
      const aborted = (e as Error)?.name === 'AbortError'
      throw new ApiError(0, aborted ? '请求超时，请稍后重试' : '网络连接失败，请检查网络后重试')
    }

    // token 失效：清登录态回登录页（skipAuth 的公开接口除外）
    if (res.status === 401 && !opts.skipAuth) {
      setToken('')
      unauthorizedHandler?.()
      redirectToLogin()
      throw new ApiError(401, '登录已失效，请重新登录')
    }

    const data = (await res.json().catch(() => ({}))) as Record<string, unknown>
    if (!res.ok) {
      // 后端错误统一 {"error": "<中文消息>"}；204 无 body（登出）
      throw new ApiError(res.status, typeof data.error === 'string' ? data.error : `请求失败（${res.status}）`)
    }
    return data as T
  }
  return withTimeout(run(), REQUEST_TIMEOUT, '请求超时，请稍后重试', ctrl)
}

export const http = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown, opts?: { skipAuth?: boolean }) =>
    request<T>(path, { method: 'POST', body, skipAuth: opts?.skipAuth }),
  patch: <T>(path: string, body?: unknown) => request<T>(path, { method: 'PATCH', body }),
  delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
}
