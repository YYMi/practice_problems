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
  },
  {
    path: '/db-admin',
    name: 'DbAdmin',
    // 数据库管理页面（仅管理员可访问）
    component: () => import('../views/DbAdmin/index.vue'),
    meta: { requiresAdmin: true }
  },
  {
    path: '/collection',
    name: 'Collection',
    // 集合页面
    component: () => import('../views/Collection/index.vue')
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
  const userInfo = localStorage.getItem('user_info')
  let isAdmin = false

  // 解析用户信息，获取管理员状态
  if (userInfo) {
    try {
      const user = JSON.parse(userInfo)
      isAdmin = user.is_admin === 1
    } catch (e) {
      console.error('解析用户信息失败', e)
    }
  }

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
      // 3. 检查是否需要管理员权限
      if (to.meta.requiresAdmin && !isAdmin) {
        // 需要管理员权限但用户不是管理员
        alert('此页面需要管理员权限')
        next('/')
      } else {
        // 有 Token，放行
        next()
      }
    }
  }
})

export default router
