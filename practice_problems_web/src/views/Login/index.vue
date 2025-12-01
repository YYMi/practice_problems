<template>
  <div class="login-container">
    <el-card class="login-card">
      <div class="title-header">题库管理系统</div>
      
      <!-- 标签页切换：登录 / 注册 -->
      <el-tabs v-model="activeTab" class="custom-tabs" stretch>
        
        <!-- ================= 登录面板 ================= -->
        <el-tab-pane label="登录" name="login">
          <el-form :model="loginForm" ref="loginFormRef" size="large" @submit.prevent>
            <el-form-item prop="username">
              <el-input v-model="loginForm.username" placeholder="请输入用户名" :prefix-icon="User" />
            </el-form-item>
            <el-form-item prop="password">
              <el-input 
                v-model="loginForm.password" 
                type="password" 
                placeholder="请输入密码 (空密码可直接登录)" 
                :prefix-icon="Lock" 
                show-password 
                @keyup.enter="handleLogin" 
              />
            </el-form-item>
            <el-button type="primary" class="w-100" :loading="loading" @click="handleLogin" round>
              立即登录
            </el-button>
          </el-form>
        </el-tab-pane>

        <!-- ================= 注册面板 ================= -->
        <el-tab-pane label="注册新账号" name="register">
          <el-form :model="registerForm" ref="registerFormRef" size="large" :rules="registerRules" status-icon>
            
            <!-- 1. 用户名 -->
            <el-form-item prop="username">
              <el-input v-model="registerForm.username" placeholder="设置用户名" :prefix-icon="User" />
            </el-form-item>
            
            <!-- 2. 密码 -->
            <el-form-item prop="password">
              <el-input v-model="registerForm.password" type="password" placeholder="设置密码" :prefix-icon="Lock" show-password />
            </el-form-item>

            <!-- 3. 确认密码 -->
            <el-form-item prop="confirmPassword">
              <el-input v-model="registerForm.confirmPassword" type="password" placeholder="再次输入密码" :prefix-icon="Check" show-password />
            </el-form-item>

            <!-- 4. 昵称 (选填) -->
            <el-form-item prop="nickname">
              <el-input v-model="registerForm.nickname" placeholder="昵称 (选填)" :prefix-icon="MagicStick" />
            </el-form-item>

            <!-- 5. 邮箱 (选填) -->
            <el-form-item prop="email">
              <el-input v-model="registerForm.email" placeholder="邮箱 (选填)" :prefix-icon="Message" />
            </el-form-item>

            <el-button type="success" class="w-100" :loading="regLoading" @click="handleRegister" round>
              确认注册并登录
            </el-button>
          </el-form>
        </el-tab-pane>

      </el-tabs>
    </el-card>

    <!-- 强制修改密码弹窗 -->
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
        <el-button type="primary" @click="handleSubmitNewPwd" class="w-100">确认修改并进入系统</el-button>
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
      // 获取后端返回的所有信息
      const { token, user_code, username, nickname, email, need_change_pwd } = res.data.data
      
      // 存储 Token
      localStorage.setItem('auth_token', token)
      
      // 存储完整的用户信息，方便首页展示
      localStorage.setItem('user_info', JSON.stringify({ 
        user_code, 
        username, 
        nickname, 
        email 
      }))

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

// 校验两次密码是否一致
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
        // 1. 准备注册数据 (排除 confirmPassword)
        const { confirmPassword, ...postData } = registerForm
        
        // 2. 发送注册请求
        const res: any = await request.post('/auth/register', postData)
        
        if (res.data.code === 200) {
          ElMessage.success('注册成功，正在为您自动登录...')
          
          // =================================================
          // 🔥 核心逻辑：注册成功后，自动调用登录
          // =================================================
          
          // 准备登录参数
          loginForm.username = registerForm.username
          loginForm.password = registerForm.password
          
          // 调用登录接口
          const loginRes: any = await request.post('/auth/login', loginForm)
          
          if (loginRes.data.code === 200) {
            const { token, user_code, username, nickname, email, need_change_pwd } = loginRes.data.data
            
            // 保存数据
            localStorage.setItem('auth_token', token)
            localStorage.setItem('user_info', JSON.stringify({ 
              user_code, 
              username, 
              nickname, 
              email 
            }))
            
            // 跳转
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
        // 如果自动登录失败，至少切回登录 tab 让用户手动点一下
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  background-size: cover;
}

.login-card {
  width: 420px;
  padding: 10px 20px 30px;
  border-radius: 12px;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
  background-color: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
}

.title-header {
  text-align: center;
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 20px;
  letter-spacing: 2px;
}

.w-100 {
  width: 100%;
  font-weight: bold;
  margin-top: 10px;
}

.mb-20 {
  margin-bottom: 20px;
}

:deep(.el-tabs__nav-wrap::after) {
  height: 1px;
  background-color: #eee;
}
:deep(.el-tabs__item) {
  font-size: 16px;
}
</style>