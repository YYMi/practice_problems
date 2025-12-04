<template>
  <header class="app-header">
    <!-- 1. 左侧品牌 Logo -->
    <div class="brand">
      <div class="logo-box"><el-icon><Collection /></el-icon></div>
      <div class="brand-text">
        <span class="main-name">题库</span>
        <span class="sub-name">Yu Song Song Ya!</span>
      </div>
      
      <!-- 模式切换 -->
      <div class="mode-switch-area">
        <el-dropdown trigger="click" @command="handleModeChange">
          <span class="mode-badge" :class="viewMode">
            {{ getModeLabel(viewMode) }} <el-icon><CaretBottom /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="read" :disabled="viewMode === 'read'">📖 阅读模式 (纯净)</el-dropdown-item>
              <el-dropdown-item command="edit" :disabled="viewMode === 'edit'">📝 编辑模式 (默认)</el-dropdown-item>
              <el-dropdown-item v-if="showDevOption" command="dev" :disabled="viewMode === 'dev'" divided style="color: #e6a23c">
                🛠️ 开发模式 (强制显示)
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>
    
    <!-- 2. 中间科目滚动区 -->
    <div class="subject-scroll-area">
      <div
        v-for="item in subjects"
        :key="item.id"
        class="subject-pill"
        :class="{ 
          'active': currentSubject?.id === item.id,
          'is-mine': item.creatorCode === userInfo.user_code,
          'is-other': item.creatorCode !== userInfo.user_code 
        }"
        :style="getWatermarkStyle(item.creatorCode)"
        @click="$emit('select', item)"
        :title="item.creatorCode === userInfo.user_code ? '我的科目' : '创建者: ' + item.creatorCode"
      >
        <span class="dot" v-if="currentSubject?.id === item.id"></span>
        <span class="subject-name">{{ item.name }}</span>
        
        <!-- 交互修改区域：三个点操作 -->
        <div class="pill-right-actions" @click.stop v-if="viewMode !== 'read'">
          <!-- 情况 A: 自己的资源 -->
          <el-dropdown 
            v-if="item.creatorCode === userInfo.user_code" 
            trigger="click" 
            @command="(cmd:any) => handleCommand(cmd, item)"
          >
            <span class="action-trigger"><el-icon><MoreFilled /></el-icon></span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="edit" icon="Edit">修改名称</el-dropdown-item>
                <el-dropdown-item command="users" icon="User">管理授权用户</el-dropdown-item>
                <el-dropdown-item command="delete" icon="Delete" divided style="color: #f56c6c">删除科目</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <!-- 情况 B: 别人的资源 -->
          <el-popover
            v-else
            placement="bottom"
            :width="240"
            trigger="click"
          >
            <template #reference>
              <span class="action-trigger"><el-icon><MoreFilled /></el-icon></span>
            </template>
            <div class="author-mini-card">
              <div class="am-header">
                <el-avatar :size="30" style="background: #e6a23c">{{ item.creatorCode.charAt(0).toUpperCase() }}</el-avatar>
                <span class="am-title">资源来源</span>
              </div>
              <div class="am-body">
                <div class="am-row">
                  <strong>ID:</strong> {{ item.creatorCode }}
                  <el-icon class="am-copy" title="复制ID" @click="copyText(item.creatorCode)"><CopyDocument /></el-icon>
                </div>
                <div class="am-row" v-if="item.creatorName"><strong>昵称:</strong> {{ item.creatorName }}</div>
                <div class="am-row" v-if="item.creatorEmail">
                  <strong>邮箱:</strong> {{ item.creatorEmail }}
                  <el-icon class="am-copy" title="复制邮箱" @click="copyText(item.creatorEmail)"><CopyDocument /></el-icon>
                </div>
              </div>
              <div class="am-tips">您仅拥有查看权限</div>
            </div>
          </el-popover>
        </div>
      </div>
      
      <!-- 添加科目按钮 -->
      <el-button 
        v-if="viewMode !== 'read'"
        class="add-subject-btn" 
        type="primary" 
        icon="Plus" 
        circle 
        plain 
        @click="$emit('open-dialog')" 
      />
    </div>

    <!-- 3. 右侧操作区 -->
    <div class="header-right-actions">
      <!-- 公告按钮 -->
      <el-button 
        class="share-btn" 
        type="warning" 
        plain 
        icon="Bell" 
        @click="announcementVisible = true"
      > 
        公告
      </el-button>
      
      <!-- 分享按钮 -->
      <el-button 
        class="share-btn" 
        type="primary" 
        plain 
        icon="Share" 
        @click="shareDialogVisible = true"
      > 
        分享 & 绑定
      </el-button>

      <!-- 源码仓库按钮 (圆形版) -->
      <el-popover placement="bottom" :width="180" trigger="click" popper-class="repo-popover">
        <template #reference>
          <el-button 
            class="share-btn repo-btn-circle" 
            type="primary" 
            plain 
            circle
          > 
            <el-icon :size="18">
              <svg viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg">
                <path d="M511.6 76.3C264.3 76.2 64 276.4 64 523.5 64 718.9 189.3 885 363.8 946c23.5 5.9 19.9-10.8 19.9-22.2v-77.5c-135.7 15.9-141.2-73.9-150.3-88.9C215 726 171.5 718 184.5 703c6.9-15.9 29.1-4 48.3 14.3 16.7 23.1 51.7 26.3 73.6 20.5 11.7-19.2 29.8-41 53.3-51.5-109.7-16.2-198.3-44.2-198.3-192.1 0-43.6 17.9-82.9 48.2-113.7-14.4-34-20-96.3 5-158.8 0 0 47.8-14.6 156.5 58.8 45.1-12.3 93.5-18.5 141.8-18.5 48.3 0 96.7 6.2 141.9 18.5 108.6-73.4 156.3-58.8 156.3-58.8 25 62.5 19.4 124.8 5 158.8 30.4 30.8 48.1 70.1 48.1 113.7 0 148.3-88.7 175.8-198.5 191.9 30.9 21 54.9 60.6 54.9 122.2v150.2c0 11.5-3.5 28.2 20.1 22.2C834.7 884.9 960 718.8 960 523.5c0-247.1-200.3-447.3-448.4-447.2z" fill="currentColor"></path>
              </svg>
            </el-icon>
          </el-button>
        </template>
        <div class="repo-list">
          <a href="https://gitee.com/yuaizifeng/practice_problems" target="_blank" class="repo-item gitee">
            <svg class="icon" viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg" width="20" height="20"><path d="M512 1024C229.222 1024 0 794.778 0 512S229.222 0 512 0s512 229.222 512 512-229.222 512-512 512z m259.149-568.883h-290.74a25.293 25.293 0 0 0-25.292 25.293l-0.026 63.206c0 13.952 11.315 25.293 25.267 25.293h177.024c13.978 0 25.293 11.315 25.293 25.267v12.646a75.853 75.853 0 0 1-75.853 75.853h-240.23a25.293 25.293 0 0 1-25.267-25.293V417.382a75.853 75.853 0 0 1 75.853-75.853h353.946a25.293 25.293 0 0 0 25.267-25.292l0.077-63.207a25.293 25.293 0 0 0-25.268-25.293H417.152a189.62 189.62 0 0 0-189.62 189.645V771.15c0 13.977 11.316 25.293 25.294 25.293h372.94a170.65 170.65 0 0 0 170.65-170.65V480.384a25.293 25.293 0 0 0-25.293-25.267z" fill="#C71D23"></path></svg>
            <span>Gitee (码云)</span>
          </a>
          <a href="https://github.com/YYMi/practice_problems" target="_blank" class="repo-item github">
            <svg class="icon" viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg" width="20" height="20"><path d="M511.6 76.3C264.3 76.2 64 276.4 64 523.5 64 718.9 189.3 885 363.8 946c23.5 5.9 19.9-10.8 19.9-22.2v-77.5c-135.7 15.9-141.2-73.9-150.3-88.9C215 726 171.5 718 184.5 703c6.9-15.9 29.1-4 48.3 14.3 16.7 23.1 51.7 26.3 73.6 20.5 11.7-19.2 29.8-41 53.3-51.5-109.7-16.2-198.3-44.2-198.3-192.1 0-43.6 17.9-82.9 48.2-113.7-14.4-34-20-96.3 5-158.8 0 0 47.8-14.6 156.5 58.8 45.1-12.3 93.5-18.5 141.8-18.5 48.3 0 96.7 6.2 141.9 18.5 108.6-73.4 156.3-58.8 156.3-58.8 25 62.5 19.4 124.8 5 158.8 30.4 30.8 48.1 70.1 48.1 113.7 0 148.3-88.7 175.8-198.5 191.9 30.9 21 54.9 60.6 54.9 122.2v150.2c0 11.5-3.5 28.2 20.1 22.2C834.7 884.9 960 718.8 960 523.5c0-247.1-200.3-447.3-448.4-447.2z" fill="#333333"></path></svg>
            <span>GitHub</span>
          </a>
        </div>
      </el-popover>

      <!-- 用户头像 & 个人中心 -->
      <div class="header-user">
        <el-popover placement="bottom-end" :width="240" trigger="click">
          <template #reference>
            <div class="user-avatar-wrapper">
              <el-avatar :size="32" style="background-color: #409eff; cursor: pointer;">
                {{ userInfo.nickname ? userInfo.nickname.charAt(0).toUpperCase() : (userInfo.username ? userInfo.username.charAt(0).toUpperCase() : 'U') }}
              </el-avatar>
            </div>
          </template>
          <div class="user-profile-card">
            <div class="upc-header">
              <div class="upc-avatar">{{ userInfo.nickname ? userInfo.nickname.charAt(0).toUpperCase() : 'U' }}</div>
              <div class="upc-names">
                <div class="upc-nick">{{ userInfo.nickname || '未设置昵称' }}</div>
                <div class="upc-user">@{{ userInfo.username }}</div>
              </div>
            </div>
            <div class="upc-body">
              <div class="upc-item"><label>ID:</label> <span>{{ userInfo.user_code }}</span></div>
              <div class="upc-item"><label>邮箱:</label> <span>{{ userInfo.email || '未绑定' }}</span></div>
            </div>
            
            <el-button type="warning" plain size="small" class="w-100" style="margin-bottom: 10px;" @click="manageDialogVisible = true">
              管理我的分享码
            </el-button>

            <el-divider style="margin: 0 0 12px 0;" />
            
            <div class="upc-actions">
              <!-- 修改点击事件：先重置确认密码，再打开弹窗 -->
              <el-button type="primary" plain size="small" class="w-100" @click="openProfileDialog">修改信息</el-button>
              <el-button type="danger" plain size="small" class="w-100" @click="$emit('logout')">退出登录</el-button>
            </div>
          </div>
        </el-popover>
      </div>
    </div>

    <!-- ============ 弹窗区域 ============ -->
    
    <!-- 1. 科目弹窗 -->
    <el-dialog v-model="subjectDialog.visible" :title="subjectDialog.isEdit ? '修改科目' : '添加科目'" width="400px">
      <el-form :model="subjectForm" @submit.prevent><el-form-item label="名称"><el-input v-model="subjectForm.name" @keydown.enter.prevent="$emit('submit-subject')" /></el-form-item></el-form>
      <template #footer><el-button type="primary" v-reclick="() => $emit('submit-subject')">确定</el-button></template>
    </el-dialog>

       <!-- 2. 个人信息设置弹窗 (修复版) -->
    <el-dialog v-model="profileDialog.visible" title="个人信息设置" width="450px" @open="initProfileForm">
      <el-form 
        :model="localForm" 
        ref="profileFormRef" 
        :rules="profileRules" 
        label-width="80px"
        status-icon
      >
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="localForm.nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="localForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        
        <el-divider content-position="center">修改密码 (可选)</el-divider>
        
        <!-- 旧密码：不再强制必填，根据你的需求 -->
        <el-form-item label="旧密码" prop="oldPassword">
          <el-input 
            v-model="localForm.oldPassword" 
            type="password" 
            show-password 
            placeholder="若修改密码，请输入旧密码" 
          />
        </el-form-item>
        
        <el-form-item label="新密码" prop="newPassword">
          <el-input 
            v-model="localForm.newPassword" 
            type="password" 
            show-password 
            placeholder="8位以上新密码" 
          />
        </el-form-item>

        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input 
            v-model="localForm.confirmPassword" 
            type="password" 
            show-password 
            placeholder="请再次输入新密码" 
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="profileDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveProfile">保存修改</el-button>
      </template>
    </el-dialog>

    <!-- 3. 公告弹窗 -->
    <el-dialog 
      v-model="announcementVisible" 
      width="600px" 
      append-to-body
      class="clean-dialog"  
      :show-close="false"
    >
      <ShareAnnouncement 
        v-if="announcementVisible" 
        :userInfo="userInfo" 
        @close="announcementVisible = false" 
      />
    </el-dialog>

    <!-- 4. 其他业务弹窗 -->
    <ShareDialog v-model:visible="shareDialogVisible" :subjects="subjects" :userInfo="userInfo" @refresh="$emit('refresh-subjects')" />
    <ShareManageDialog v-model:visible="manageDialogVisible" />
    <SubjectUserManager v-model:visible="userManagerVisible" :subjectId="currentManageSubject?.id" :subjectName="currentManageSubject?.name" />

  </header>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { Bell } from "@element-plus/icons-vue";
import { ElMessage } from 'element-plus';
import { Collection, Edit, Delete, Plus, Share, MoreFilled, User, CopyDocument, CaretBottom } from "@element-plus/icons-vue";
import ShareDialog from "./ShareDialog.vue"; 
import ShareManageDialog from "./ShareManageDialog.vue"; 
import SubjectUserManager from "./SubjectUserManager.vue"; 
import ShareAnnouncement from '../../../components/ShareAnnouncement.vue';
import md5 from 'js-md5'; // ★★★ 引入 MD5 ★★★

const props = defineProps([
  'subjects', 'currentSubject', 'userInfo', 
  'subjectDialog', 'subjectForm', 'profileDialog', 'profileForm',
  'viewMode'
]);
const emit = defineEmits([
  'select', 'open-dialog', 'delete', 'submit-subject', 
  'open-profile', 'submit-profile', // submit-profile 现在会携带参数
  'logout', 'refresh-subjects', 'update:viewMode'
]);

// ★★★ 新增：本地表单数据 (解决父子组件数据同步延迟问题) ★★★
const localForm = reactive({
  nickname: '',
  email: '',
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
});

// 状态定义
const announcementVisible = ref(false);
const showDevOption = import.meta.env.VITE_SHOW_DEV_MODE === 'true';
const shareDialogVisible = ref(false);
const manageDialogVisible = ref(false);
const userManagerVisible = ref(false);
const currentManageSubject = ref<any>(null);

// ★★★ 个人中心相关逻辑 ★★★
const profileFormRef = ref();
const confirmNewPassword = ref(''); // 确认密码

// ★★★ 解决问题2：校验规则现在引用本地变量，反应极快 ★★★
const validateConfirmPwd = (rule: any, value: any, callback: any) => {
  if (localForm.newPassword && value === '') {
    callback(new Error('请再次输入新密码'));
  } else if (localForm.newPassword && value !== localForm.newPassword) {
    callback(new Error('两次输入的新密码不一致!'));
  } else {
    callback();
  }
};

// ★★★ 解决问题1：旧密码校验逻辑 ★★★
const validateOldPwd = (rule: any, value: any, callback: any) => {
    callback();
};

const profileRules = reactive({
  nickname: [
    { max: 20, message: '昵称过长', trigger: 'blur' }
  ],
  oldPassword: [
    { validator: validateOldPwd, trigger: 'blur' }
  ],
  newPassword: [
    { min: 8, message: '新密码长度不能少于 8 位', trigger: 'blur' }
  ],
  confirmPassword: [
    { validator: validateConfirmPwd, trigger: 'blur' } // 这里的 trigger 用 blur 或 change 都可以
  ]
});


// ★★★ 初始化表单：每次打开弹窗时执行 ★★★
// 在 <el-dialog @open="initProfileForm"> 中调用
const initProfileForm = () => {
  // 1. 从 props.userInfo 复制当前信息
  localForm.nickname = props.userInfo.nickname || '';
  localForm.email = props.userInfo.email || '';
  
  // 2. 清空密码框
  localForm.oldPassword = '';
  localForm.newPassword = '';
  localForm.confirmPassword = '';
  
  // 3. 清除之前的红色报错
  if (profileFormRef.value) {
    profileFormRef.value.clearValidate();
  }
};




// 打开弹窗时重置确认密码
const openProfileDialog = () => {
  confirmNewPassword.value = '';
  emit('open-profile');
};

// 关闭时重置表单
const resetProfileForm = () => {
  if(profileFormRef.value) profileFormRef.value.resetFields();
  confirmNewPassword.value = '';
};

// 保存修改
const handleSaveProfile = async () => {
  if (!profileFormRef.value) return;

  await profileFormRef.value.validate((valid: boolean) => {
    if (valid) {
      const payload: any = {
        nickname: localForm.nickname,
        email: localForm.email,
      };

      // 密码处理
      if (localForm.newPassword) {
        // 如果旧密码填了，就加密；没填(如果你允许空的话)传空字符串
        payload.old_password = localForm.oldPassword ? md5(localForm.oldPassword) : '';
        payload.new_password = md5(localForm.newPassword);
      }

      emit('submit-profile', payload);
    }
  });
};

// 辅助函数
const handleModeChange = (mode: string) => {
  emit('update:viewMode', mode);
};
const getModeLabel = (mode: string) => {
  switch(mode) {
    case 'read': return '阅读';
    case 'edit': return '编辑';
    case 'dev': return '开发';
    default: return '编辑';
  }
};
const handleCommand = (cmd: string, item: any) => {
  if (cmd === 'edit') emit('open-dialog', item);
  else if (cmd === 'delete') emit('delete', item);
  else if (cmd === 'users') {
    currentManageSubject.value = item;
    userManagerVisible.value = true;
  }
};
const copyText = (text: string) => {
  if(!text) return;
  navigator.clipboard.writeText(text);
  ElMessage.success('已复制');
};
const getWatermarkStyle = (code: string) => {
  const text = code || 'Unknown';
  const svgContent = `<svg xmlns='http://www.w3.org/2000/svg' width='90' height='40'><text x='50%' y='50%' font-size='11' font-weight='bold' fill='rgba(0,0,0,0.2)' font-family='Arial' text-anchor='middle' dominant-baseline='middle' transform='rotate(-15, 45, 20)'>${text}</text></svg>`;
  return { backgroundImage: `url("data:image/svg+xml;charset=utf-8,${encodeURIComponent(svgContent)}")`, backgroundRepeat: 'repeat', backgroundPosition: 'center' };
};
</script>

<style scoped>
/* ============================================================
   1. 头部容器：紫色渐变背景
   ============================================================ */
.app-header { 
  height: 64px; 
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-bottom: none; 
  display: flex; 
  align-items: center; 
  padding: 0 24px; 
  box-shadow: 0 4px 12px rgba(0,0,0,0.15); 
  z-index: 10; 
  flex-shrink: 0; 
  color: #fff; 
}

/* ============================================================
   2. Logo 区域
   ============================================================ */
.brand { display: flex; align-items: center; margin-right: 40px; }
.logo-box { 
  width: 36px; height: 36px; 
  background: #fff; 
  color: #764ba2; 
  border-radius: 8px; 
  display: flex; align-items: center; justify-content: center; 
  font-size: 20px; margin-right: 10px; 
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2); 
}
.brand-text { display: flex; flex-direction: column; line-height: 1.1; }
.main-name { font-weight: 800; font-size: 16px; color: #fff; } 
.sub-name { font-size: 10px; color: rgba(255,255,255,0.8); text-transform: uppercase; letter-spacing: 1px; }

/* 模式切换 */
.mode-switch-area { margin-left: 15px; padding-left: 15px; border-left: 1px solid rgba(255,255,255,0.3); height: 24px; display: flex; align-items: center; }
.mode-badge { font-size: 12px; padding: 2px 8px; border-radius: 10px; cursor: pointer; display: flex; align-items: center; gap: 2px; user-select: none; transition: all 0.2s; background: rgba(255,255,255,0.2); color: #fff; border: 1px solid transparent; }
.mode-badge:hover { background: rgba(255,255,255,0.3); }
.mode-badge.read { color: #e1f3d8; }
.mode-badge.edit { color: #fff; font-weight: bold; }
.mode-badge.dev { color: #ffd700; }

/* ============================================================
   3. 科目滚动区
   ============================================================ */
.subject-scroll-area { display: flex; align-items: center; gap: 8px; flex: 1; overflow-x: auto; padding-bottom: 2px; }
.subject-scroll-area::-webkit-scrollbar { display: none; }

.subject-pill { 
  padding: 6px 36px 6px 16px; 
  border-radius: 6px; cursor: pointer; font-size: 14px; 
  transition: all 0.3s; display: flex; align-items: center; 
  position: relative; white-space: nowrap; overflow: hidden; 
  border: 1px solid transparent; user-select: none; 
  background-color: rgba(255, 255, 255, 0.15);
  color: rgba(255, 255, 255, 0.9);
  border-color: transparent;
}
.subject-pill:hover { background-color: rgba(255, 255, 255, 0.25); color: #fff; }
.subject-pill.active { background-color: #fff !important; color: #764ba2 !important; border-color: #fff !important; box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2); }
.subject-pill.is-other { border: 1px dashed rgba(255, 255, 255, 0.5); background-color: rgba(255, 247, 235, 0.1); color: #ffeebb; }
.subject-pill.is-other.active { background-color: #fff7eb !important; color: #d48806 !important; border-style: solid; }
.subject-pill .dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; margin-right: 6px; }
.subject-name { font-weight: 500; position: relative; z-index: 2; }

.pill-right-actions { position: absolute; right: 4px; top: 50%; transform: translateY(-50%); z-index: 10; opacity: 0; transition: opacity 0.2s; }
.subject-pill:hover .pill-right-actions { opacity: 1; }
.action-trigger { padding: 4px; border-radius: 4px; cursor: pointer; font-size: 14px; color: rgba(255,255,255,0.7); display: flex; align-items: center; }
.subject-pill.active .action-trigger { color: #909399; } 
.subject-pill.active .action-trigger:hover { color: #764ba2; background: rgba(0,0,0,0.05); }
.subject-pill:not(.active) .action-trigger:hover { color: #fff; background: rgba(255,255,255,0.2); }

.add-subject-btn { color: #fff !important; border-color: rgba(255,255,255,0.5) !important; background: transparent !important; }
.add-subject-btn:hover { background: rgba(255,255,255,0.2) !important; border-color: #fff !important; }

/* ============================================================
   4. 右侧操作区
   ============================================================ */
.header-right-actions { display: flex; align-items: center; gap: 15px; }
.share-btn { border-radius: 20px; padding: 8px 18px; background: rgba(255,255,255,0.15) !important; border: 1px solid rgba(255,255,255,0.3) !important; color: #fff !important; }
.share-btn:hover { background: rgba(255,255,255,0.25) !important; }
.user-avatar-wrapper .el-avatar { border: 2px solid rgba(255,255,255,0.6); background-color: #fff !important; color: #764ba2 !important; font-weight: bold; }

/* 弹窗样式 */
.user-profile-card { padding: 5px; }
.upc-header { display: flex; align-items: center; margin-bottom: 15px; }
.upc-avatar { width: 48px; height: 48px; border-radius: 50%; background: linear-gradient(135deg, #667eea, #764ba2); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 20px; font-weight: bold; margin-right: 12px; box-shadow: 0 2px 8px rgba(0,0,0,0.15); }
.upc-names { display: flex; flex-direction: column; }
.upc-nick { font-size: 16px; font-weight: 600; color: #303133; line-height: 1.2; }
.upc-user { font-size: 12px; color: #909399; margin-top: 2px; }
.upc-body { font-size: 13px; color: #606266; margin-bottom: 10px; }
.upc-item { display: flex; margin-bottom: 6px; }
.upc-item label { color: #909399; width: 40px; margin-right: 5px; }
.w-100 { width: 100%; }
.upc-actions { display: flex; gap: 10px; justify-content: space-between; }
.upc-actions .el-button { flex: 1; }

.author-mini-card { padding: 5px; }
.am-header { display: flex; align-items: center; margin-bottom: 10px; gap: 10px; }
.am-title { font-weight: bold; font-size: 14px; color: #303133; }
.am-body { font-size: 12px; color: #606266; margin-bottom: 8px; }
.am-row { margin-bottom: 4px; display: flex; align-items: center; }
.am-copy { cursor: pointer; margin-left: 6px; color: #909399; vertical-align: middle; }
.am-copy:hover { color: #409eff; }
.am-tips { font-size: 10px; color: #909399; text-align: right; font-style: italic; }
</style>

<style>
/* 全局弹窗样式修正 */
.clean-dialog .el-dialog__header { display: none !important; }
.clean-dialog .el-dialog__body { padding: 0 !important; height: 100%; overflow: hidden; }
.clean-dialog { border-radius: 12px !important; overflow: hidden !important; box-shadow: 0 15px 40px rgba(0,0,0,0.3) !important; }

/* 源码仓库弹窗 */
.repo-list { display: flex; flex-direction: column; gap: 8px; }
.repo-item { display: flex; align-items: center; padding: 10px 12px; border-radius: 8px; text-decoration: none; color: #606266; transition: all 0.2s; font-size: 14px; font-weight: 500; background-color: #f9fafe; }
.repo-item svg { margin-right: 10px; }
.repo-item:hover { background-color: #f0f2f5; transform: translateX(4px); }
.repo-item.gitee:hover { color: #c71d23; background-color: rgba(199, 29, 35, 0.05); }
.repo-item.github:hover { color: #333; background-color: rgba(0, 0, 0, 0.05); }
.repo-btn-circle { width: 32px !important; height: 32px !important; padding: 0 !important; border-radius: 50% !important; display: flex; align-items: center; justify-content: center; }
</style>