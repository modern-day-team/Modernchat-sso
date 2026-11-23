import { createRouter, createWebHistory } from 'vue-router'

import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'home', component: () => import('@/views/HomeView.vue') },
    { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue') },
    { path: '/register', name: 'register', component: () => import('@/views/RegisterView.vue') },
    // 忘记/重置不设登录守卫：登录态过期时也能走完流程（主仓库 /forgot 同款）
    { path: '/forgot', name: 'forgot', component: () => import('@/views/ForgotView.vue') },
    { path: '/reset', name: 'reset', component: () => import('@/views/ResetView.vue') },
    { path: '/admin', name: 'admin', component: () => import('@/views/AdminView.vue') },
    // 站点设置：与账号管理同级，复用同一套站长守卫
    { path: '/admin/settings', name: 'admin-settings', component: () => import('@/views/AdminSettingsView.vue') },
  ],
})

// 登录态守卫：未登录访问门户 → 登录页；已登录访问登录/注册 → 门户；/admin* 需站长角色
router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (to.name === 'admin' || to.name === 'admin-settings') {
    if (!auth.isLoggedIn) return { name: 'login', query: { redirect: to.fullPath } }
    // user 是内存态：刷新后为空，懒加载一次再判角色
    if (!auth.user) await auth.fetchMe()
    if (auth.user?.role !== 'admin') return { name: 'home' }
    return true
  }
  if (to.name === 'home' && !auth.isLoggedIn) return { name: 'login' }
  if ((to.name === 'login' || to.name === 'register') && auth.isLoggedIn) return { name: 'home' }
  return true
})

export default router
