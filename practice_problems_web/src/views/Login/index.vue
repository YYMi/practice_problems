<template>
  <div class="login-container">
    <el-card class="login-card">
      <div class="login-layout">
        <!-- ================= 左侧：品牌与全诗 ================= -->
        <div class="left-side">
          <!-- 1. 顶部 Logo (保持横排，固定在顶部) -->
          <div class="brand-box">
            <el-icon :size="32"><Collection /></el-icon>
            <span class="brand-text">知识汇</span>
          </div>

          <!-- 2. 竖排诗词容器 -->
          <div class="poem-container">
            
            <!-- 第一组：标题与作者 (显示在最右侧) -->
            <div class="p-col p-meta">
              <span>晋</span>
              <span class="dot">·</span>
              <span>陶渊明</span>
              <span class="gap"></span>
              <span class="title">《移居》</span>
            </div>

            <!-- 第二组：正文 (从右向左依次排列) -->
            <div class="p-col">昔欲居南村，非为卜其宅。</div>
            <div class="p-col">闻多素心人，乐与数晨夕。</div>
            <div class="p-col">怀此颇有年，今日从兹役。</div>
            <div class="p-col">敝庐何必广，取足蔽床席。</div>
            <div class="p-col">邻曲时时来，抗言谈在昔。</div>

            <!-- 第三组：分隔线 (可选) -->
            <div class="p-line-break"></div>

            <!-- 第四组：核心金句 (显示在最左侧，高亮) -->
            <div class="p-col p-highlight">
              奇文共欣赏，疑义相与析。
            </div>
            
          </div>
        </div>

        <!-- ================= 右侧：表单区 ================= -->
        <div class="right-side">
          <h2 class="form-title">欢迎登录</h2>
          
          <el-tabs v-model="activeTab" class="custom-tabs" stretch>
            
            <!-- 登录面板 -->
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
                <el-button type="primary" class="w-100 gradient-btn" :loading="loading" @click="handleLogin" round>
                  立即登录
                </el-button>
              </el-form>
            </el-tab-pane>

            <!-- 注册面板 -->
            <el-tab-pane label="注册" name="register">
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
                <el-button type="success" class="w-100 gradient-btn-success" :loading="regLoading" @click="handleRegister" round>
                  确认注册并登录
                </el-button>
              </el-form>
            </el-tab-pane>

          </el-tabs>
        </div>
      </div>
    </el-card>

    <!-- ================= 强制修改密码弹窗 ================= -->
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
import { User, Lock, MagicStick, Check, Message, Collection } from '@element-plus/icons-vue'
import request from '../../utils/request'
import md5 from 'js-md5'

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
  
  let finalPassword = ''

  // 1. 如果输入了密码，校验长度并加密
  if (loginForm.password) {
    if (loginForm.password.length < 8) {
      return ElMessage.warning('密码长度错误：不能少于 8 位')
    }
    finalPassword = md5(loginForm.password)
  } else {
    // 2. 如果没输密码，传空字符串，让后端判断是否允许空密码登录
    finalPassword = ""
  }
  
  loading.value = true
  try {
    const loginPayload = {
      username: loginForm.username,
      password: finalPassword
    }

    const res: any = await request.post('/auth/login', loginPayload)
    
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
  if (pwdForm.newPassword.length < 8) {
    return ElMessage.warning('新密码长度不能少于 8 位')
  }

  try {
    const payload = { new_password: md5(pwdForm.newPassword) }
    const res: any = await request.put('/user/profile', payload)
    
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
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, message: '为了您的账号安全，密码长度不能少于 8 位', trigger: 'blur' }
  ],
  confirmPassword: [{ validator: validatePass2, trigger: 'blur' }]
}

const handleRegister = async () => {
  if (!registerFormRef.value) return
  
  await registerFormRef.value.validate(async (valid: boolean) => {
    if (valid) {
      regLoading.value = true
      try {
        const { confirmPassword, ...tempData } = registerForm
        
        // 注册必须 MD5
        const registerPayload = {
          ...tempData,
          password: md5(tempData.password)
        }

        const res: any = await request.post('/auth/register', registerPayload)
        
        if (res.data.code === 200) {
          ElMessage.success('注册成功，正在为您自动登录...')
          
          // 自动登录也要 MD5
          const autoLoginPayload = {
            username: registerForm.username,
            password: md5(registerForm.password)
          }
          
          const loginRes: any = await request.post('/auth/login', autoLoginPayload)
          
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
      } finally {
        regLoading.value = false
      }
    }
  })
}
</script>

<style scoped>
/* 背景容器 */
.login-container {
  height: 100vh;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  background-size: cover;
}

/* 卡片主体 */
.login-card {
  width: 750px;
  padding: 0;
  border-radius: 20px;
  overflow: hidden;
  border: none;
  box-shadow: 0 20px 50px rgba(0,0,0,0.2);
}

.login-layout {
  display: flex;
  height: 500px;
}

/* === 左侧容器 === */
.left-side {
  width: 40%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 30px 20px;
  display: flex;
  flex-direction: column; /* 整体还是上下结构：Logo在上，诗在下 */
  position: relative;
  overflow: hidden;
}

/* Logo 区域 (横排) */
.brand-box {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 24px;
  font-weight: bold;
  color: #fff;
  margin-bottom: 10px; /* 与诗词拉开距离 */
  flex-shrink: 0;
}

/* === 诗词竖排核心容器 === */
.poem-container {
  flex: 1;
  width: 100%;
  
  /* ★★★ 核心修复：布局策略改变 ★★★ */
  /* 1. 外层容器保持【横排 Flex】 */
  display: flex;
  
  /* 2. 【反向排列】：让写在 HTML 最前面的“标题”显示在最右边 */
  flex-direction: row-reverse; 
  
  /* 3. 居中对齐 */
  justify-content: center; 
  align-items: center;
  
  gap: 12px; /* 列间距 */
  user-select: none;
}

/* 每一列诗句 (单独竖排) */
.p-col {
  /* ★★★ 核心修复：让每一行单独竖起来 ★★★ */
  writing-mode: vertical-rl; 
  text-orientation: upright; /* 汉字直立 */
  
  font-family: "Kaiti SC", "STKaiti", "KaiTi", serif; /* 楷体 */
  color: #fff;
  font-size: 16px;
  letter-spacing: 6px;
  line-height: 1.2;
  opacity: 0.7;
  transition: all 0.3s;
  
  /* 强制不换行，确保一句话就是一列 */
  white-space: nowrap; 
  height: auto; /* 自适应高度 */
}

.p-col:hover {
  opacity: 1;
  text-shadow: 0 0 10px rgba(255,255,255,0.5);
}

/* 标题和作者 (最右侧) */
.p-meta {
  font-size: 12px;
  opacity: 0.5;
  /* 这里的 margin-left 其实视觉上是给右边的 Logo 留空隙，但在 row-reverse 下是左边距 */
  margin-left: 20px; 
  
  /* 内部元素排列 */
  display: flex;
  align-items: center; 
  justify-content: start; /* 顶对齐 */
}
.p-meta .gap {
  height: 15px; /* 竖排模式下 height 变成了横向宽度，这里用 height 撑开间距 */
}
.p-meta .title {
  font-weight: bold;
  margin-top: 10px;
}

/* 分隔空隙 */
.p-line-break {
  width: 10px; /* 物理占位 */
}

/* 核心金句 (最左侧，高亮) */
.p-highlight {
  font-size:  15px;
  font-weight: bold;
  opacity: 1;
  letter-spacing: 8px;
  color: #fff;
  text-shadow: 2px 2px 8px rgba(0,0,0,0.3);
  margin-right: 10px; /* 视觉上的左边距 */
}
/* === 右侧：白色表单区 === */
.right-side {
  width: 60%;
  background: #fff;
  padding: 40px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.form-title {
  font-size: 22px;
  color: #333;
  margin-bottom: 20px;
  font-weight: 700;
}

/* 通用样式 */
.w-100 {
  width: 100%;
  font-weight: bold;
  margin-top: 10px;
  height: 44px;
  font-size: 16px;
}

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

.gradient-btn-success {
  background: linear-gradient(90deg, #36d1dc, #5b86e5);
  border: none;
}
.gradient-btn-success:hover {
  opacity: 0.9;
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(91, 134, 229, 0.4);
}

.auth-form .el-input__wrapper {
  border-radius: 20px;
}

.mb-20 { margin-bottom: 20px; }

/* Tab 样式微调 */
:deep(.el-tabs__nav-wrap::after) { height: 1px; background-color: #ebeef5; }
:deep(.el-tabs__item) { font-size: 16px; color: #606266; }
:deep(.el-tabs__item.is-active) { color: #764ba2; font-weight: bold; }
:deep(.el-tabs__active-bar) { background-color: #764ba2; }
</style>
