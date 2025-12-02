<template>
  <div class="login-container">
    <el-card class="login-card">
      <!-- ★★★ 修改1：更具品牌感的标题区域 ★★★ -->
      <div class="brand-section">
        <div class="title-header">知识汇 · Knowledge Hub</div>
        <div class="sub-slogan">知识来源于分析，成长始于分享</div>
      </div>
      
      <!-- 标签页切换：登录 / 注册 -->
      <el-tabs v-model="activeTab" class="custom-tabs" stretch>
        
        <!-- ================= 登录面板 ================= -->
        <el-tab-pane label="登录" name="login">
          <el-form :model="loginForm" ref="loginFormRef" size="large" @submit.prevent class="auth-form">
            <el-form-item prop="username">
              <el-input v-model="loginForm.username" placeholder="请输入用户名" :prefix-icon="User" />
            </el-form-item>
            <el-form-item prop="password">
              <el-input 
                v-model="loginForm.password" 
                type="password" 
                placeholder="请输入密码" 
                :prefix-icon="Lock" 
                show-password 
                @keyup.enter="handleLogin" 
              />
            </el-form-item>
            <!-- 登录按钮改为渐变紫 -->
            <el-button type="primary" class="w-100 gradient-btn" :loading="loading" @click="handleLogin" round>
              立即登录
            </el-button>
          </el-form>
        </el-tab-pane>

        <!-- ================= 注册面板 ================= -->
        <el-tab-pane label="注册新账号" name="register">
          <el-form :model="registerForm" ref="registerFormRef" size="large" :rules="registerRules" status-icon class="auth-form">
            
            <el-form-item prop="username">
              <el-input v-model="registerForm.username" placeholder="设置用户名" :prefix-icon="User" />
            </el-form-item>
            
            <el-form-item prop="password">
              <el-input v-model="registerForm.password" type="password" placeholder="设置密码" :prefix-icon="Lock" show-password />
            </el-form-item>

            <el-form-item prop="confirmPassword">
              <el-input v-model="registerForm.confirmPassword" type="password" placeholder="再次输入密码" :prefix-icon="Check" show-password />
            </el-form-item>

            <el-form-item prop="nickname">
              <el-input v-model="registerForm.nickname" placeholder="昵称 (选填)" :prefix-icon="MagicStick" />
            </el-form-item>

            <el-form-item prop="email">
              <el-input v-model="registerForm.email" placeholder="邮箱 (选填)" :prefix-icon="Message" />
            </el-form-item>

            <!-- 注册按钮也改为渐变紫 -->
            <el-button type="success" class="w-100 gradient-btn-success" :loading="regLoading" @click="handleRegister" round>
              确认注册并登录
            </el-button>
          </el-form>
        </el-tab-pane>

      </el-tabs>
    </el-card>

    <!-- 强制修改密码弹窗 (保持不变) -->
    <el-dialog
      v-model="pwdDialogVisible"
      title="首次登录 / 密码为空"
      width="400px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :show-close="false"
      center
    >
      <el-alert title="为了安全，请设置您的新密码" type="warning" :closable="false" class="mb-20" center show-icon />
      <el-form :model="pwdForm">
        <el-form-item label="新密码">
          <el-input v-model="pwdForm.newPassword" type="password" show-password placeholder="请输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button type="primary" @click="handleSubmitNewPwd" class="w-100 gradient-btn">确认修改并进入系统</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, MagicStick, Check, Message } from '@element-plus/icons-vue'
import request from '../../utils/request'

const router = useRouter()
const loading = ref(false)
const regLoading = ref(false)
const activeTab = ref('login') 

// =========== 登录逻辑 ===========
const loginForm = reactive({
  username: '',
  password: ''
})

const pwdDialogVisible = ref(false)
const pwdForm = reactive({ newPassword: '' })

const handleLogin = async () => {
  if (!loginForm.username) return ElMessage.warning('请输入用户名')
  
  loading.value = true
  try {
    const res: any = await request.post('/auth/login', loginForm)
    if (res.data.code === 200) {
      const { token, user_code, username, nickname, email, need_change_pwd } = res.data.data
      
      localStorage.setItem('auth_token', token)
      localStorage.setItem('user_info', JSON.stringify({ user_code, username, nickname, email }))

      if (need_change_pwd) {
        ElMessage.warning('检测到您的密码为空，请强制设置新密码！')
        pwdDialogVisible.value = true 
      } else {
        ElMessage.success('登录成功')
        router.push('/') 
      }
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const handleSubmitNewPwd = async () => {
  if (!pwdForm.newPassword) return ElMessage.warning('新密码不能为空')
  try {
    const res: any = await request.put('/user/profile', { new_password: pwdForm.newPassword })
    if (res.data.code === 200) {
      ElMessage.success('密码设置成功，欢迎进入系统')
      pwdDialogVisible.value = false
      router.push('/') 
    }
  } catch (e) { console.error(e) }
}

// =========== 注册逻辑 ===========
const registerFormRef = ref()
const registerForm = reactive({
  username: '',
  password: '',
  confirmPassword: '', 
  nickname: '',
  email: ''
})

const validatePass2 = (rule: any, value: any, callback: any) => {
  if (value === '') {
    callback(new Error('请再次输入密码'))
  } else if (value !== registerForm.password) {
    callback(new Error('两次输入密码不一致!'))
  } else {
    callback()
  }
}

const registerRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  confirmPassword: [{ validator: validatePass2, trigger: 'blur' }]
}

const handleRegister = async () => {
  if (!registerFormRef.value) return
  
  await registerFormRef.value.validate(async (valid: boolean) => {
    if (valid) {
      regLoading.value = true
      try {
        const { confirmPassword, ...postData } = registerForm
        const res: any = await request.post('/auth/register', postData)
        
        if (res.data.code === 200) {
          ElMessage.success('注册成功，正在为您自动登录...')
          
          loginForm.username = registerForm.username
          loginForm.password = registerForm.password
          
          const loginRes: any = await request.post('/auth/login', loginForm)
          
          if (loginRes.data.code === 200) {
            const { token, user_code, username, nickname, email, need_change_pwd } = loginRes.data.data
            localStorage.setItem('auth_token', token)
            localStorage.setItem('user_info', JSON.stringify({ user_code, username, nickname, email }))
            
            if (need_change_pwd) {
              pwdDialogVisible.value = true 
            } else {
              ElMessage.success(`欢迎加入，${nickname || username}`)
              router.push('/') 
            }
          }
        }
      } catch (e) {
        console.error(e)
        activeTab.value = 'login'
      } finally {
        regLoading.value = false
      }
    }
  })
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  /* 保持蓝紫渐变背景 */
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  background-size: cover;
}

.login-card {
  width: 440px; /* 稍微加宽一点 */
  padding: 20px 30px 40px;
  border-radius: 16px; /* 圆角加大 */
  border: none; /* 去掉默认边框 */
  /* 毛玻璃效果增强 */
  background-color: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(20px);
  box-shadow: 0 15px 35px rgba(0, 0, 0, 0.2);
}

.brand-section {
  text-align: center;
  margin-bottom: 30px;
}

.title-header {
  font-size: 28px;
  font-weight: 800;
  /* 标题使用渐变色 */
  background: linear-gradient(135deg, #667eea, #764ba2);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin-bottom: 8px;
  letter-spacing: 1px;
}

.sub-slogan {
  font-size: 14px;
  color: #909399;
  font-weight: 500;
  letter-spacing: 2px;
}

.w-100 {
  width: 100%;
  font-weight: bold;
  margin-top: 10px;
  height: 44px; /* 按钮加高 */
  font-size: 16px;
}

/* 登录按钮渐变 */
.gradient-btn {
  background: linear-gradient(90deg, #667eea, #764ba2);
  border: none;
  transition: all 0.3s;
}
.gradient-btn:hover {
  opacity: 0.9;
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(118, 75, 162, 0.4);
}

/* 注册按钮渐变 (稍微不同，用绿色系或者保持紫色系均可，这里保持紫色系一致性) */
.gradient-btn-success {
  background: linear-gradient(90deg, #36d1dc, #5b86e5); /* 蓝绿渐变区分一下注册 */
  border: none;
}
.gradient-btn-success:hover {
  opacity: 0.9;
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(91, 134, 229, 0.4);
}

.auth-form .el-input__wrapper {
  border-radius: 20px; /* 输入框更圆润 */
}

.mb-20 {
  margin-bottom: 20px;
}

:deep(.el-tabs__nav-wrap::after) {
  height: 1px;
  background-color: #ebeef5;
}
:deep(.el-tabs__item) {
  font-size: 16px;
  color: #606266;
}
:deep(.el-tabs__item.is-active) {
  color: #764ba2; /* 选中 Tab 变紫 */
  font-weight: bold;
}
:deep(.el-tabs__active-bar) {
  background-color: #764ba2; /* Tab 下划线变紫 */
}
</style>
