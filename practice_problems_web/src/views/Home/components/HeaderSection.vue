<template>
  <header class="app-header">
    <!-- 1. 左侧品牌 Logo -->
    <div class="brand">
      <div class="logo-box"><el-icon><Collection /></el-icon></div>
      <div class="brand-text">
        <span class="main-name">题库</span>
        <span class="sub-name">Yu Song Song Ya!</span>
      </div>
      
      <!-- ★★★ 新增：模式切换 (紧跟 Logo) ★★★ -->
      <div class="mode-switch-area">
        <el-dropdown trigger="click" @command="handleModeChange">
          <span class="mode-badge" :class="viewMode">
            {{ getModeLabel(viewMode) }} <el-icon><CaretBottom /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="read" :disabled="viewMode === 'read'">📖 阅读模式 (纯净)</el-dropdown-item>
              <el-dropdown-item command="edit" :disabled="viewMode === 'edit'">📝 编辑模式 (默认)</el-dropdown-item>
              <!-- 只有配置了环境变量 VITE_SHOW_DEV_MODE=true 才显示 -->
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
        
        <!-- ★★★ 交互修改区域：三个点操作 ★★★ -->
        <!-- 阅读模式下完全隐藏操作按钮 -->
        <div class="pill-right-actions" @click.stop v-if="viewMode !== 'read'">
          
          <!-- 情况 A: 自己的资源 -> 下拉菜单 (修改、管理用户、删除) -->
          <el-dropdown 
            v-if="item.creatorCode === userInfo.user_code" 
            trigger="click" 
            @command="(cmd) => handleCommand(cmd, item)"
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

          <!-- 情况 B: 别人的资源 -> 作者信息弹窗 (带复制功能) -->
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
      
      <!-- ★★★ 添加科目按钮：仅在非阅读模式下显示 ★★★ -->
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
            
            <!-- 管理我的分享码入口 -->
            <el-button 
              type="warning" 
              plain 
              size="small" 
              class="w-100" 
              style="margin-bottom: 10px;"
              @click="manageDialogVisible = true"
            >
              管理我的分享码
            </el-button>

            <el-divider style="margin: 0 0 12px 0;" />
            
            <div class="upc-actions">
              <el-button type="primary" plain size="small" class="w-100" @click="$emit('open-profile')">修改信息</el-button>
              <el-button type="danger" plain size="small" class="w-100" @click="$emit('logout')">退出登录</el-button>
            </div>
          </div>
        </el-popover>
      </div>
    </div>

    <!-- ============ 弹窗区域 ============ -->
    <el-dialog v-model="subjectDialog.visible" :title="subjectDialog.isEdit ? '修改科目' : '添加科目'" width="400px">
      <el-form :model="subjectForm" @submit.prevent><el-form-item label="名称"><el-input v-model="subjectForm.name" @keydown.enter.prevent="$emit('submit-subject')" /></el-form-item></el-form>
      <template #footer><el-button type="primary" v-reclick="() => $emit('submit-subject')">确定</el-button></template>
    </el-dialog>

    <el-dialog v-model="profileDialog.visible" title="个人信息设置" width="420px">
      <el-form :model="profileForm" label-width="70px">
        <el-form-item label="昵称"><el-input v-model="profileForm.nickname" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="profileForm.email" /></el-form-item>
        <el-divider content-position="center">修改密码 (可选)</el-divider>
        <el-form-item label="旧密码"><el-input v-model="profileForm.oldPassword" type="password" show-password /></el-form-item>
        <el-form-item label="新密码"><el-input v-model="profileForm.newPassword" type="password" show-password /></el-form-item>
      </el-form>
      <template #footer><el-button @click="profileDialog.visible = false">取消</el-button><el-button type="primary" v-reclick="() => $emit('submit-profile')">保存修改</el-button></template>
    </el-dialog>

    <ShareDialog v-model:visible="shareDialogVisible" :subjects="subjects" :userInfo="userInfo" @refresh="$emit('refresh-subjects')" />
    <ShareManageDialog v-model:visible="manageDialogVisible" />
    <SubjectUserManager v-model:visible="userManagerVisible" :subjectId="currentManageSubject?.id" :subjectName="currentManageSubject?.name" />

  </header>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import { Collection, Edit, Delete, Plus, Share, MoreFilled, User, CopyDocument, CaretBottom } from "@element-plus/icons-vue";
import ShareDialog from "./ShareDialog.vue"; 
import ShareManageDialog from "./ShareManageDialog.vue"; 
import SubjectUserManager from "./SubjectUserManager.vue"; 

const props = defineProps([
  'subjects', 'currentSubject', 'userInfo', 
  'subjectDialog', 'subjectForm', 'profileDialog', 'profileForm',
  'viewMode' // <--- 接收 viewMode
]);
const emit = defineEmits([
  'select', 'open-dialog', 'delete', 'submit-subject', 'open-profile', 'submit-profile', 'logout', 'refresh-subjects',
  'update:viewMode' // <--- 发送模式更新
]);

// 读取环境变量
const showDevOption = import.meta.env.VITE_SHOW_DEV_MODE === 'true';

const shareDialogVisible = ref(false);
const manageDialogVisible = ref(false);
const userManagerVisible = ref(false);
const currentManageSubject = ref<any>(null);

// 模式切换处理
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
/* 基础样式 */
.app-header { height: 64px; background: #fff; border-bottom: 1px solid #e4e7ed; display: flex; align-items: center; padding: 0 24px; box-shadow: 0 2px 8px rgba(0,0,0,0.03); z-index: 10; flex-shrink: 0; }
.brand { display: flex; align-items: center; margin-right: 40px; }
.logo-box { width: 36px; height: 36px; background: linear-gradient(135deg, #409eff, #36cfc9); color: #fff; border-radius: 8px; display: flex; align-items: center; justify-content: center; font-size: 20px; margin-right: 10px; box-shadow: 0 2px 6px rgba(64, 158, 255, 0.3); }
.brand-text { display: flex; flex-direction: column; line-height: 1.1; }
.main-name { font-weight: 800; font-size: 16px; color: #2c3e50; }
.sub-name { font-size: 10px; color: #909399; text-transform: uppercase; letter-spacing: 1px; }

/* 模式切换 */
.mode-switch-area { margin-left: 15px; padding-left: 15px; border-left: 1px solid #e4e7ed; height: 24px; display: flex; align-items: center; }
.mode-badge { font-size: 12px; padding: 2px 8px; border-radius: 10px; cursor: pointer; display: flex; align-items: center; gap: 2px; user-select: none; transition: all 0.2s; }
.mode-badge:hover { opacity: 0.8; }
.mode-badge.read { background: #f0f9eb; color: #67c23a; border: 1px solid #c2e7b0; }
.mode-badge.edit { background: #ecf5ff; color: #409eff; border: 1px solid #d9ecff; }
.mode-badge.dev { background: #fdf6ec; color: #e6a23c; border: 1px solid #fbeaa8; }

/* 滚动区 */
.subject-scroll-area { display: flex; align-items: center; gap: 8px; flex: 1; overflow-x: auto; padding-bottom: 2px; }
.subject-scroll-area::-webkit-scrollbar { display: none; }

/* 胶囊 */
.subject-pill { padding: 6px 36px 6px 16px; border-radius: 6px; cursor: pointer; font-size: 14px; color: #606266; transition: all 0.3s; display: flex; align-items: center; position: relative; white-space: nowrap; overflow: hidden; border: 1px solid transparent; user-select: none; }
.subject-pill.is-mine { background-color: #ffffff; border-color: #e4e7ed; }
.subject-pill.is-mine:hover { background-color: #f2f6fc; border-color: #dcdfe6; }
.subject-pill.is-mine.active { background-color: #ecf5ff; color: #409eff; border-color: #b3d8ff; box-shadow: 0 2px 4px rgba(64, 158, 255, 0.1); }
.subject-pill.is-other { background-color: #fdf6ec; border-color: #faecd8; border-style: dashed; color: #e6a23c; }
.subject-pill.is-other:hover { background-color: #faecd8; }
.subject-pill.is-other.active { background-color: #fff7eb; color: #d48806; border-color: #e6a23c; border-style: solid; }
.subject-pill .dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; margin-right: 6px; }
.subject-name { font-weight: 500; position: relative; z-index: 2; text-shadow: 0 1px 0 rgba(255,255,255,0.8); }

/* 胶囊右侧操作 */
.pill-right-actions { position: absolute; right: 4px; top: 50%; transform: translateY(-50%); z-index: 10; opacity: 0; transition: opacity 0.2s; }
.subject-pill:hover .pill-right-actions { opacity: 1; }
.action-trigger { padding: 4px; border-radius: 4px; cursor: pointer; font-size: 14px; color: #909399; display: flex; align-items: center; }
.action-trigger:hover { background: rgba(0,0,0,0.05); color: #409eff; }

/* 右侧操作区 */
.header-right-actions { display: flex; align-items: center; gap: 15px; }
.share-btn { border-radius: 20px; padding: 8px 18px; }

/* 用户卡片 */
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

/* 作者信息卡片 */
.author-mini-card { padding: 5px; }
.am-header { display: flex; align-items: center; margin-bottom: 10px; gap: 10px; }
.am-title { font-weight: bold; font-size: 14px; color: #303133; }
.am-body { font-size: 12px; color: #606266; margin-bottom: 8px; }
.am-row { margin-bottom: 4px; display: flex; align-items: center; }
.am-copy { cursor: pointer; margin-left: 6px; color: #909399; vertical-align: middle; }
.am-copy:hover { color: #409eff; }
.am-tips { font-size: 10px; color: #909399; text-align: right; font-style: italic; }
</style>