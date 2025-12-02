<template>
  <div class="announcement-board">
    
    <!-- ★★★ 1. 头部：整合了 标题 + 刷新 + 关闭 ★★★ -->
    <div class="board-header-purple">
      <div class="bh-left">
        <div class="bh-icon-box"><el-icon><BellFilled /></el-icon></div>
        <span class="bh-title">分享公告板</span>
      </div>
      <div class="bh-right">
        <!-- 刷新按钮 -->
        <el-button 
          class="action-btn" 
          circle 
          link
          @click="fetchData" 
          :loading="loading"
        >
          <el-icon size="18"><Refresh /></el-icon>
        </el-button>
        
        <div class="divider-v"></div>

        <!-- 关闭按钮 -->
        <el-icon class="close-btn" @click="emit('close')"><Close /></el-icon>
      </div>
    </div>

    <!-- 内容区域 -->
    <div class="board-content custom-scrollbar" v-loading="loading">
      <el-empty v-if="!loading && list.length === 0" description="暂无分享公告" :image-size="100" />
      
      <!-- 卡片列表 -->
      <div v-else class="card-list">
        <div v-for="item in list" :key="item.id" class="share-card">
          
          <!-- A. 顶部：用户 + 分享码 -->
          <div class="card-top-row">
            <div class="user-profile">
              <el-avatar :size="36" :style="{ backgroundColor: stringToColor(item.creatorCode || 'U') }">
                {{ (item.creatorCode || 'U').charAt(0).toUpperCase() }}
              </el-avatar>
              <div class="meta-info">
                <div class="username">{{ item.creatorCode }}</div>
                <div class="publish-time">{{ formatTime(item.createTime) }} 发布</div>
              </div>
            </div>

            <div class="code-badge-wrapper" @click="handleCopy(item.shareCode)">
              <el-icon><Link /></el-icon>
              <span class="code-font">{{ item.shareCode }}</span>
              <el-icon class="copy-icon"><CopyDocument /></el-icon>
            </div>
          </div>

              <!-- B. 卡片中部：大气公告内容 (去掉了Label) -->
          <div class="card-main-content">
            <!-- 之前的 note-highlight-box 删掉，换成这个 -->
            <div class="quote-box">
              <!-- 左上角的装饰图标 -->
              <el-icon class="quote-icon left"><ChatLineSquare /></el-icon>
              
              <!-- 直接显示内容，没有“公告内容”四个字了 -->
              <p class="note-text" :class="{ 'empty-text': !item.note }">
                {{ item.note || '作者没有话要说...' }}
              </p>
            </div>
          </div>

          <!-- C. 底部：时间 + 操作 -->
          <div class="card-footer-row">
            <div class="footer-info" :class="{ 'near-expire': isNearExpire(item.expireTime) }">
              <el-icon><Timer /></el-icon>
              <span>{{ formatExpire(item.expireTime) }}</span>
            </div>

            <div class="footer-actions" v-if="isOwner(item)">
              <el-button link type="primary" size="small" @click="handleEdit(item)">编辑</el-button>
              <span class="divider">|</span>
              <el-popconfirm title="确定删除？" @confirm="handleDelete(item)">
                <template #reference>
                  <el-button link type="danger" size="small">删除</el-button>
                </template>
              </el-popconfirm>
            </div>
          </div>

        </div>
      </div>
    </div>

    <!-- 修改弹窗 (保持不变) -->
    <el-dialog
      v-model="dialogVisible"
      width="460px"
      append-to-body
      class="custom-dialog"
      :show-close="false"
    >
      <template #header>
        <div class="dialog-header-purple sub-header">
          <div class="dh-icon-box"><el-icon><Edit /></el-icon></div>
          <span class="dh-title-center">修改公告</span>
          <el-icon class="dh-close" @click="dialogVisible = false"><Close /></el-icon>
        </div>
      </template>

      <div class="dialog-body-content">
        <el-form :model="form" label-position="top" class="stylish-form">
          <el-form-item label="关联分享码">
            <div class="read-only-box"><el-icon><Link /></el-icon> {{ form.shareCode }}</div>
          </el-form-item>
          <el-form-item label="过期时间">
            <el-date-picker v-model="form.expireTime" type="datetime" placeholder="选择失效时间" value-format="YYYY-MM-DD HH:mm:ss" class="full-width-picker" />
          </el-form-item>
          <el-form-item label="备注说明">
            <el-input v-model="form.note" type="textarea" :rows="4" placeholder="请输入公告内容..." maxlength="200" show-word-limit class="custom-textarea" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer-simple">
          <el-button class="btn-cancel" @click="dialogVisible = false">取消</el-button>
          <el-button class="btn-submit purple-btn" type="primary" @click="handleSubmit" :loading="submitting">保存修改</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
// 记得引入 Quote 图标
import { BellFilled, CopyDocument, Delete, Edit, Refresh, Timer, Close, Link, ChatLineSquare } from '@element-plus/icons-vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { 
  getShareAnnouncements, 
  deleteShareAnnouncement, 
  updateShareAnnouncement 
} from '../api/share';

const props = defineProps<{ userInfo: any }>();
// ★★★ 新增 emit 定义 ★★★
const emit = defineEmits(['close']);

const loading = ref(false);
const list = ref<any[]>([]);
const dialogVisible = ref(false);
const submitting = ref(false);
const currentEditId = ref<number | null>(null);

const form = ref({ shareCode: '', note: '', expireTime: '' });

const fetchData = async () => {
  loading.value = true;
  try {
    const res = await getShareAnnouncements();
    if (res.data.code === 200) {
      list.value = res.data.data || [];
    }
  } finally {
    loading.value = false;
  }
};

const handleDelete = (row: any) => {
  deleteShareAnnouncement(row.id).then(() => {
    ElMessage.success('删除成功');
    fetchData();
  }).catch(() => ElMessage.error('删除失败'));
};

const handleEdit = (row: any) => {
  currentEditId.value = row.id;
  form.value = {
    shareCode: row.shareCode,
    note: row.note,
    expireTime: row.expireTime
  };
  dialogVisible.value = true;
};

const handleSubmit = async () => {
  if (!currentEditId.value) return;
  submitting.value = true;
  try {
    await updateShareAnnouncement(currentEditId.value, {
      note: form.value.note,
      expireTime: form.value.expireTime
    });
    ElMessage.success('修改成功');
    dialogVisible.value = false;
    fetchData();
  } catch (e) {
    ElMessage.error('修改失败');
  } finally {
    submitting.value = false;
  }
};

const isOwner = (row: any) => {
  return props.userInfo && row.creatorCode === props.userInfo.user_code;
};

const handleCopy = (text: string) => {
  navigator.clipboard.writeText(text);
  ElMessage.success('分享码已复制');
};

const stringToColor = (str: string) => {
  let hash = 0;
  for (let i = 0; i < str.length; i++) hash = str.charCodeAt(i) + ((hash << 5) - hash);
  const c = (hash & 0x00ffffff).toString(16).toUpperCase();
  return '#' + '00000'.substring(0, 6 - c.length) + c;
};

const isNearExpire = (dateStr: string) => {
  if (!dateStr) return false;
  const diff = new Date(dateStr).getTime() - Date.now();
  return diff > 0 && diff < 3 * 24 * 3600 * 1000;
};

const formatTime = (str: string) => str ? str.split(' ')[0] : '刚刚';
const formatExpire = (str: string) => str ? str.split(' ')[0] + ' 截止' : '永久有效';

onMounted(() => fetchData());
</script>

<style scoped>
.announcement-board {
  background: #f5f7fa;
  height: 70vh; /* 给一个固定高度，确保能滚动 */
  display: flex;
  flex-direction: column;
}

/* ================= 1. 头部 (整合版) ================= */
.board-header-purple {
  height: 54px;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  color: #fff;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(118, 75, 162, 0.2);
}
/* 子弹窗的头部微调 */
.sub-header { box-shadow: none; }

.bh-left { display: flex; align-items: center; gap: 10px; }
.bh-icon-box {
  width: 28px; height: 28px;
  background: rgba(255,255,255,0.2);
  border-radius: 6px;
  display: flex; align-items: center; justify-content: center;
  font-size: 16px;
}
.bh-title { font-size: 16px; font-weight: 600; letter-spacing: 0.5px; }

.bh-right { display: flex; align-items: center; gap: 5px; }
.action-btn { color: #fff; font-size: 18px; }
.action-btn:hover { color: rgba(255,255,255,0.8); background: transparent; }
.divider-v { width: 1px; height: 16px; background: rgba(255,255,255,0.3); margin: 0 8px; }
.close-btn { font-size: 20px; cursor: pointer; opacity: 0.8; padding: 4px; border-radius: 4px; }
.close-btn:hover { opacity: 1; background: rgba(255,255,255,0.2); }

/* 内容区 */
.board-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

/* 列表布局 */
.card-list { display: flex; flex-direction: column; gap: 20px; }

/* ================= 2. 卡片样式 ================= */
.share-card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.03);
  border: 1px solid #f0f2f5;
  overflow: hidden;
  transition: all 0.3s ease;
}
.share-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(118, 75, 162, 0.1);
  border-color: #d3adf7;
}

/* A. 顶部 */
.card-top-row {
  padding: 12px 20px;
  display: flex; justify-content: space-between; align-items: center;
  border-bottom: 1px solid #f9f9f9;
  background: #fff;
}
.user-profile { display: flex; gap: 10px; align-items: center; }
.user-profile .el-avatar { border: 2px solid #fff; box-shadow: 0 2px 6px rgba(0,0,0,0.1); font-weight: bold; color: #fff; }
.meta-info { display: flex; flex-direction: column; }
.username { font-size: 14px; font-weight: 700; color: #303133; }
.publish-time { font-size: 12px; color: #909399; }

.code-badge-wrapper {
  display: inline-flex; align-items: center; gap: 6px;
  background: #f7f8fa; padding: 4px 10px; border-radius: 6px;
  border: 1px solid #e4e7ed; cursor: pointer; transition: all 0.2s;
  font-size: 12px; color: #606266;
}
.code-badge-wrapper:hover {
  background: #fff; border-color: #764ba2; color: #764ba2;
}
.code-font { font-family: 'Monaco', monospace; font-weight: 600; }

/* 引用块样式 */
.card-main-content { padding: 20px 25px; }

.quote-box {
  position: relative;
  background: #f8f9fb; /* 极淡的灰背景 */
  border-radius: 8px;
  padding: 15px 20px 15px 40px; /* 左边留宽给引号图标 */
}

.quote-icon {
  position: absolute; left: 12px; top: 12px;
  font-size: 24px; color: #e0e3e9; /* 很淡的灰色图标 */
}

.note-text {
  font-size: 15px;
  color: #303133;
  line-height: 1.6;
  margin: 0;
  font-weight: 500;
  white-space: pre-wrap; /* 保留换行 */
}

.empty-text { color: #a8abb2; font-style: italic; }

/* C. 底部 */
.card-footer-row {
  padding: 10px 20px;
  background: #fcfcfc;
  border-top: 1px solid #f0f2f5;
  display: flex; justify-content: space-between; align-items: center;
  font-size: 12px; color: #909399;
}
.footer-info { display: flex; align-items: center; gap: 5px; }
.footer-info.near-expire { color: #f56c6c; font-weight: bold; }

.footer-actions { display: flex; align-items: center; }
.divider { color: #e4e7ed; margin: 0 8px; font-size: 10px; }

/* 滚动条 */
.custom-scrollbar::-webkit-scrollbar { width: 6px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #dcdfe6; border-radius: 4px; }
</style>
