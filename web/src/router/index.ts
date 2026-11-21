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
  ],
})

// 登录态守卫：未登录访问门户 → 登录页；已登录访问登录/注册 → 门户
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.name === 'home' && !auth.isLoggedIn) return { name: 'login' }
  if ((to.name === 'login' || to.name === 'register') && auth.isLoggedIn) return { name: 'home' }
  return true
})

export default router
