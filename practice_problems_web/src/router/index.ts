import { createRouter, createWebHistory } from 'vue-router'
// 引入类型定义，解决 TS 报错
import type { RouteLocationNormalized, NavigationGuardNext } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Home',
    // 使用动态导入，确保文件路径正确：src/views/Home/index.vue
    component: () => import('../views/Home/index.vue')
  },
  {
    path: '/login',
    name: 'Login',
    // 使用动态导入，确保文件路径正确：src/views/Login/index.vue
    component: () => import('../views/Login/index.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 全局前置守卫
router.beforeEach((to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
  // 获取本地存储的 Token
  const token = localStorage.getItem('auth_token')

  // 1. 如果要去的是登录页
  if (to.path === '/login') {
    if (token) {
      // 如果已经有 Token 了，直接踢回首页，避免重复登录
      next('/')
    } else {
      // 没有 Token，允许进入登录页
      next()
    }
  } else {
    // 2. 如果去的不是登录页 (比如首页)
    if (!token) {
      // 没有 Token，强制跳转到登录页
      next('/login')
    } else {
      // 有 Token，放行
      next()
    }
  }
})

export default router
