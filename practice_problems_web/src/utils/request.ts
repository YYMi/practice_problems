import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import router from '../router' 

const service = axios.create({
  baseURL: '/api/v1', 
  timeout: 5000
})

// 1. 请求拦截器 (保持不变)
service.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 处理过期的公共函数
const handleLoginExpired = () => {
  // 避免重复弹窗
  if (document.querySelector('.el-message-box__wrapper')) return;

  ElMessageBox.alert('登录状态已过期，请重新登录', '系统提示', {
    confirmButtonText: '重新登录',
    type: 'warning',
    callback: () => {
      localStorage.removeItem('auth_token')
      localStorage.removeItem('user_info')
      // 强制刷新跳转，确保状态清空
      window.location.href = '/login' 
    }
  })
}

// 2. 响应拦截器 (修复版)
service.interceptors.response.use(
  (response) => {
    // ★★★ 修复点：只读取 data 用于判断，但最后返回完整的 response ★★★
    const res = response.data
    
    // 检查业务逻辑上的 401 (即使 HTTP 状态码是 200)
    if (res.code === 401) {
      handleLoginExpired()
      return Promise.reject(new Error(res.msg || '登录已过期'))
    }

    // ★★★ 重点：这里必须返回 response，而不是 res！ ★★★
    // 这样你之前的代码 res.data.data 才能正常工作
    return response
  },
  (error) => {
    // 检查 HTTP 协议上的 401
    if (error.response && error.response.status === 401) {
      handleLoginExpired()
    } else {
      // 其他错误提示
      ElMessage.error(error.response?.data?.msg || '网络请求失败')
    }
    return Promise.reject(error)
  }
)

export default service
