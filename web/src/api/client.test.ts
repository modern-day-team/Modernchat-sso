import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { ApiError, getToken, http, onUnauthorized, setToken } from './client'

const TOKEN_KEY = 'sso_token'
let store: Map<string, string>
let unauthorizedCalls = 0

function jsonResponse(status: number, body: unknown) {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

beforeEach(() => {
  store = new Map()
  unauthorizedCalls = 0
  vi.stubGlobal('localStorage', {
    getItem: (k: string) => store.get(k) ?? null,
    setItem: (k: string, v: string) => void store.set(k, v),
    removeItem: (k: string) => void store.delete(k),
  })
  vi.stubGlobal('location', { pathname: '/', search: '', href: '' })
  onUnauthorized(() => {
    unauthorizedCalls += 1
  })
})

afterEach(() => {
  vi.unstubAllGlobals()
  vi.restoreAllMocks()
})

describe('api client', () => {
  it('读写 token', () => {
    setToken('abc')
    expect(getToken()).toBe('abc')
    expect(store.get(TOKEN_KEY)).toBe('abc')
    setToken('')
    expect(getToken()).toBe('')
    expect(store.has(TOKEN_KEY)).toBe(false)
  })

  it('Bearer 认证与成功响应', async () => {
    store.set(TOKEN_KEY, 't1')
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(200, { ok: true }))
    vi.stubGlobal('fetch', fetchMock)
    await expect(http.get('/settings')).resolves.toEqual({ ok: true })
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/settings',
      expect.objectContaining({ headers: expect.objectContaining({ Authorization: 'Bearer t1' }) }),
    )
  })

  it('401 清 token 并触发统一钩子', async () => {
    store.set(TOKEN_KEY, 't1')
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(401, { error: '未登录' })))
    await expect(http.get('/users/me')).rejects.toBeInstanceOf(ApiError)
    expect(getToken()).toBe('')
    expect(unauthorizedCalls).toBe(1)
  })

  it('skipAuth 的 401 不清登录态', async () => {
    store.set(TOKEN_KEY, 't1')
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse(401, { error: '密码错误' }))
    vi.stubGlobal('fetch', fetchMock)
    await expect(http.post('/sessions', {}, { skipAuth: true })).rejects.toMatchObject({ status: 401 })
    expect(getToken()).toBe('t1')
    expect(unauthorizedCalls).toBe(0)
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/sessions',
      expect.objectContaining({
        headers: expect.not.objectContaining({ Authorization: expect.anything() }),
      }),
    )
  })

  it('后端 error 字段归一为中文 ApiError', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(400, { error: '用户名已存在' })))
    await expect(http.post('/users', {})).rejects.toMatchObject({
      status: 400,
      message: '用户名已存在',
    })
  })

  it('网络失败归一为可读中文', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))
    await expect(http.get('/settings')).rejects.toMatchObject({
      status: 0,
      message: '网络连接失败，请检查网络后重试',
    })
  })
})
