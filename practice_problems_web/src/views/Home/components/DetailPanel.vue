<template>
  <main class="content-viewport">
    <!-- 空状态 -->
    <div v-if="!currentPoint" class="empty-state">
      <img src="https://gw.alipayobjects.com/zos/antfincdn/ZHrcdLPrvN/empty.svg" width="200">
      <p>请选择左侧知识点开始编辑</p>
    </div>
    
    <!-- 详情内容面板 -->
    <div v-else class="detail-panel custom-scrollbar">
          <!-- 1. 顶部 Header 区域 (紧凑版) -->
    <div class="detail-header">
      
      <!-- 上半部分：标题与操作按钮 -->
      <div class="header-top-row">
        <!-- 返回按钮 + 标题 -->
        <div class="point-title">
          <el-tooltip v-if="canGoBack" content="返回上一页" placement="bottom">
            <span class="back-link" @click="$emit('navigate-back')">
              <el-icon><Back /></el-icon> 返回
            </span>
          </el-tooltip>
          <span class="title-text">
            {{ currentPoint?.title }}
          </span>
          <el-icon v-if="hasPermission" class="title-edit-icon" @click="openEditTitle"><Edit /></el-icon>
          <el-tag v-if="currentPoint?.difficulty" :class="getDifficultyClass(currentPoint?.difficulty)" size="small" effect="plain" class="diff-tag">
            {{ getDifficultyLabel(currentPoint?.difficulty) }}
          </el-tag>
        </div>

        <!-- 右上角操作按钮 -->
        <div class="header-actions">
           <el-button v-if="hasPermission" type="danger" link :icon="Delete" @click="emit('delete', currentPoint)">删除</el-button>
           <el-button type="primary" size="small" @click="emit('open-practice')">
             <el-icon><collection /></el-icon> 练习 & 刷题
           </el-button>
        </div>
      </div>

      <!-- 下半部分：左右布局 (信息栏) -->
      <div class="header-info-row">
        
        <!-- 左侧：视频列表 -->
        <div class="info-left-video">
          <div class="video-compact-section">
            <span class="section-label video-label">视频讲解 ({{ parsedVideos.length }})：</span>
            
         
          <!-- 微型视频列表 -->
            <div class="video-mini-list">
              <div 
                v-for="(url, index) in parsedVideos" 
                :key="index" 
                class="mini-video-wrapper"
                title="点击播放"
                @click="openFloatingPlayer(url)"
              >
                <!-- ★★★ 修改核心：不再直接渲染 video 或 iframe，而是用纯 CSS/图标占位 ★★★ -->
                <!-- 这样可以彻底杜绝页面加载时的自动播放问题 -->
                
                <div class="video-placeholder">
                    <!-- 如果是 MP4，显示一个简化的图标 -->
                    <el-icon v-if="url.toLowerCase().includes('.mp4')" class="placeholder-icon"><VideoPlay /></el-icon>
                    
                    <!-- 如果是 B站/iframe，显示 B站 图标或通用播放图标 -->
                    <div v-else class="bilibili-icon-placeholder">
                        <span class="bili-text">TV</span>
                    </div>
                </div>

                <!-- 添加视频按钮保持不变 -->
              </div>
                
              <div v-if="hasPermission" class="add-video-btn" @click="openVideoDialog">
                <el-icon><Plus /></el-icon>
              </div>
              <div v-else-if="parsedVideos.length === 0" class="no-video-text">
                暂无视频
              </div>
            </div>
          </div>
        </div>

        <!-- 右侧：参考资料链接 -->
        <div class="info-right-links">
          <div class="links-section">
            <el-icon class="link-icon"><Link /></el-icon>
            <span class="section-label">参考资料：</span>
            
            <div class="link-list">
              <span 
                v-for="(link, index) in parsedLinks" 
                :key="index" 
                class="link-item-wrapper"
              >
                <a :href="formatUrl(link)" target="_blank" class="link-item">{{ link }}</a>
                <el-icon 
                  v-if="hasPermission" 
                  class="remove-link-icon" 
                  title="删除此链接"
                  @click="emit('remove-link', index)"
                ><Close /></el-icon>
              </span>
              
              <el-button v-if="hasPermission" type="primary" link size="small" @click="emit('add-link')">
                <el-icon><Plus /></el-icon> 添加链接
              </el-button>
            </div>
          </div>
        </div>

      </div>
    </div>

      
      <!-- 主体内容布局 (左编辑器，右图片) -->
      <div class="detail-body-layout">
        <div 
          class="panel-column editor-column"
          :class="{ 'is-mine': isPointOwner, 'is-others': !isPointOwner }"
        >
          <div class="column-content">
            <PointEditor 
              :pointId="currentPoint.id" 
              :pointTitle="currentPoint.title"
              :subjectId="currentSubject?.id || 0"
              :content="currentPoint.content" 
              :canEdit="hasPermission"
              :bindings="currentPointBindings"
              :pointsInfoMap="pointsInfoMap"
              @update="(val) => { if(currentPoint) currentPoint.content = val }" 
              @refresh-bindings="$emit('refresh-bindings')"
              @cache-point="(data) => $emit('cache-point', data)"
              @navigate-to-point="(data) => $emit('navigate-to-point', data)"
            />
          </div>
        </div>
        
        <div class="panel-column image-column">
          <div class="column-header">
            <span class="col-title">关联图片</span>
            <el-tag size="small" type="success" effect="plain">Assets</el-tag>
          </div>
          <div class="column-content">
            <ImageManager 
              :pointId="currentPoint.id" 
              :imagesJson="currentPoint.localImageNames" 
              :canEdit="hasPermission"
              @update="(val) => { if(currentPoint) currentPoint.localImageNames = val }" 
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 题目练习抽屉 -->
    <QuestionDrawer 
      v-if="currentPoint" 
      :visible="drawerVisible" 
      @update:visible="(val) => $emit('update:drawerVisible', val)" 
      :pointId="currentPoint.id" 
      :title="currentPoint.title"
      :viewMode="viewMode"       
      :userInfo="userInfo"       
      :isOwner="hasPermission"   
    />
    
    <!-- 修改标题弹窗 -->
    <el-dialog v-if="editTitleDialog" v-model="editTitleDialog.visible" title="修改知识点" width="400px">
      <el-form @submit.prevent label-width="50px">
        <el-form-item label="标题"><el-input v-model="editTitleDialog.title" @keydown.enter.prevent="$emit('submit-edit-title')" /></el-form-item>
        <el-form-item label="难度">
          <el-radio-group v-model="editTitleDialog.difficulty">
            <el-radio-button :label="0">简单</el-radio-button>
            <el-radio-button :label="1">中等</el-radio-button>
            <el-radio-button :label="2">困难</el-radio-button>
            <el-radio-button :label="3">重点</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editTitleDialog.visible = false">取消</el-button>
        <el-button type="primary" v-reclick="() => $emit('submit-edit-title')">保存</el-button>
      </template>
    </el-dialog>

    <!-- ★★★★★ 视频管理弹窗 ★★★★★ -->
    <el-dialog v-model="videoDialogVisible" title="管理讲解视频" width="600px">
      <div class="video-manage-tip">
        支持粘贴 B站 BV号 (如 BV1xxxx)、完整 URL 或 &lt;iframe&gt; 代码。
      </div>
      
      <div class="video-list-edit">
        <div v-for="(item, index) in tempVideoList" :key="index" class="video-edit-row">
          <span class="row-index">{{ index + 1 }}.</span>
          <el-input 
            v-model="tempVideoList[index]" 
            placeholder="粘贴 B站链接 / BV号 / iframe代码" 
            clearable
          />
          <el-button type="danger" icon="Delete" circle plain @click="removeVideoRow(index)" />
        </div>
        
        <el-button 
          v-if="tempVideoList.length < 10" 
          class="add-row-btn" 
          type="primary" 
          plain 
          icon="Plus" 
          @click="addVideoRow"
        >
          添加视频 ({{ tempVideoList.length }}/10)
        </el-button>
      </div>
      
      <template #footer>
        <el-button @click="videoDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitVideo">保存全部</el-button>
      </template>
    </el-dialog>

    <!-- ★★★★★ 悬浮播放器 (核心修复版) ★★★★★ -->
    <el-dialog
      v-model="playDialogVisible"
      title="视频播放 (右下角可拖拽大小)"
      width="auto" 
      class="resizable-video-dialog"
      append-to-body
      draggable
      align-center
      destroy-on-close
      show-close
      
      :modal="false"
      :lock-scroll="false"
      :close-on-click-modal="false"
      
      modal-class="video-overlay-transparent"
    >
      <!-- 
        @mousedown: 按下时标记正在拖拽
        @mouseup: 松开时取消标记
      -->
      <div class="resizable-wrapper" 
       @mousedown="isResizing = true" 
       @mouseup="isResizing = false"
       @mouseleave="isResizing = false">
    
    <!-- 遮罩层 (调整大小时防吞事件) -->
    <div v-show="isResizing" class="resize-mask"></div>

    <!-- ★★★ 核心修改：分情况渲染 ★★★ -->
    
    <!-- 情况 A: 如果是 MP4 直链，使用原生 video 标签 -->
    <video 
      v-if="currentPlayUrl.toLowerCase().includes('.mp4')"
      :src="currentPlayUrl"
      controls
      referrerpolicy="no-referrer" 
      style="width: 100%; height: 100%; object-fit: contain; background: #000;"
    ></video>

    <!-- 情况 B: 否则认为是 B 站或其他 iframe，使用 iframe 标签 -->
    <iframe 
      v-else
      :src="currentPlayUrl" 
      scrolling="no" 
      border="0" 
      frameborder="no" 
      framespacing="0" 
      allowfullscreen="true"
      style="width: 100%; height: 100%;"
    ></iframe>

  </div>
    </el-dialog>

    <!-- ★★★ AI 面试官弹窗 (全屏宽度，1/3 高度) ★★★ -->
    <el-dialog
      v-if="aiInterviewerVisible"
      v-model="aiInterviewerVisible"
      title="AI 模拟面试"
      width="100%"
      custom-class="full-width-one-third-height-dialog ai-interviewer-dialog"
      @open="() => {}"
      @close="() => {}"
      :modal="false"
      :show-close="false"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
    >
      <div class="ai-interviewer-container">
        <!-- 头部：知识点信息 + 剩余时长 -->
        <div class="interviewer-header">
          <div class="point-info">
            <el-icon class="info-icon"><Service /></el-icon>
            <span class="point-title">{{ currentPoint?.title }}</span>
          </div>
          
          <div class="header-right">
            <!-- 重新连接按钮（仅在断开时显示）-->
            <el-button 
              v-if="!isAIConnected" 
              size="small" 
              type="primary" 
              @click="reconnectAIInterviewer"
              :loading="isAIConnecting"
            >
              🔄 重新连接
            </el-button>
            
            <el-tag type="success" effect="dark" class="quota-tag">
              剩余时长: {{ formatTime(aiRemainingQuota) }}
            </el-tag>
          </div>
        </div>

        <!-- 聊天区域 -->
        <div ref="aiChatContainerRef" class="chat-container">
          <div v-if="aiMessages.length === 0" class="empty-chat">
            <el-icon :size="40" color="#909399"><ChatDotRound /></el-icon>
            <p>等待 AI 面试官连接...</p>
          </div>
          
          <div 
            v-for="(msg, index) in aiMessages" 
            :key="index" 
            :class="['message-item', msg.role]"
          >
            <div class="message-avatar" :class="`${msg.role}-avatar`">
              <el-icon :size="20" color="#fff">
                <User v-if="msg.role === 'user'" />
                <Service v-else />
              </el-icon>
            </div>
            <div class="message-content">
              <div class="message-bubble" v-html="msg.content"></div>
            </div>
          </div>
        </div>

        <!-- 输入区域 -->
        <div class="input-container">
          <el-input
            v-model="aiUserInput"
            type="textarea"
            :rows="2"
            placeholder="请输入你的回答..."
            :disabled="!isAIConnected || isAILoading"
          />
          <div class="input-actions">
            <el-button 
              @click="sendAIMessage" 
              :loading="isAILoading" 
              :disabled="!isAIConnected || !aiUserInput.trim()"
              class="gradient-btn"
              size="small"
            >
              <el-icon class="mr-1"><Promotion /></el-icon>
              发送
            </el-button>
            <el-button @click="resetAIInterview" size="small">
              <el-icon class="mr-1"><RefreshRight /></el-icon>
              重新开始
            </el-button>
          </div>
        </div>
      </div>
    </el-dialog>

  </main>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { EditPen, Delete, VideoPlay, Link, Close, Plus, Edit, Back, Service } from "@element-plus/icons-vue";
import PointEditor from "../../../components/PointEditor.vue";
import ImageManager from "../../../components/ImageManager.vue";
import QuestionDrawer from "../../../components/QuestionDrawer.vue";
import AIInterviewer from "../../../components/AIInterviewer.vue"; 
import { ElMessage } from 'element-plus';

const props = defineProps([
  'currentPoint', 'currentSubject', 'currentPointBindings', 'pointsInfoMap', 'isSubjectOwner', 'isPointOwner', 
  'subjectWatermarkText', 'parsedLinks', 'drawerVisible', 'editTitleDialog', 'canGoBack',
  'userInfo', 'viewMode' 
]);

const emit = defineEmits([
  'update:drawerVisible', 'update:currentPoint', 
  'open-edit-title', 'submit-edit-title', 'delete', 
  'add-link', 'remove-link', 
  'save-video',
  'open-practice', // 练习 & 刷题按钮
  'refresh-bindings', // 刷新绑定列表
  'cache-point', // 缓存知识点信息
  'navigate-to-point', // 跳转到知识点
  'navigate-back' // 返回上一个知识点
]);

// 权限判断
const hasPermission = computed(() => {
  if (props.viewMode === 'read') return false;
  if (props.viewMode === 'dev') return true;
  return !!props.isPointOwner || !!props.isSubjectOwner;
});

// 打开编辑标题弹窗
const openEditTitle = () => {
  if (!hasPermission.value) return;
  emit('open-edit-title');
};

// 难度标签样式
const getDifficultyClass = (difficulty: number | undefined) => {
  const map: Record<number, string> = {
    0: 'diff-easy',
    1: 'diff-medium',
    2: 'diff-hard',
    3: 'diff-important'
  };
  return map[difficulty ?? 0] || '';
};

// 难度标签文字
const getDifficultyLabel = (difficulty: number | undefined) => {
  const map: Record<number, string> = {
    0: '简单',
    1: '中等',
    2: '困难',
    3: '重点'
  };
  return map[difficulty ?? 0] || '简单';
};

// 链接格式化
const formatLinkText = (link: string) => {
  if (!link) return '';
  if (link.length <= 30) return link;
  const start = link.substring(0, 15);
  const end = link.substring(link.length - 15);
  return `${start}...${end}`;
};

const formatUrl = (url: string) => {
  if (!url) return '#';
  url = url.trim();
  if (!/^https?:\/\//i.test(url)) {
    return 'http://' + url;
  }
  return url;
};

// ==========================================
// ★★★★★ 视频相关逻辑 ★★★★★
// ==========================================

// 1. 解析数据库存的 JSON 字符串 -> 数组
const parsedVideos = computed(() => {
  // 兼容后端可能返回大写 VideoUrl 的情况
  const jsonStr = props.currentPoint?.videoUrl || props.currentPoint?.VideoUrl;
  if (!jsonStr) return [];
  try {
    const arr = JSON.parse(jsonStr);
    if (typeof arr === 'string') return [arr];
    return Array.isArray(arr) ? arr : [];
  } catch (e) {
    return jsonStr ? [jsonStr] : [];
  }
});

// 2. 将 URL 转换为 B站 Embed 地址
const getBilibiliEmbed = (url: string) => {
  if (!url) return '';
  const bvRegex = /(BV[a-zA-Z0-9]{10})/;
  const match = url.match(bvRegex);
  
  if (match) {
    const bvid = match[1];
    // page=1: 第一P
    // high_quality=1: 高清优先
    // danmaku=0: 关弹幕
    // autoplay=0: 默认不自动播，避免静音问题
    return `//player.bilibili.com/player.html?bvid=${bvid}&page=1&high_quality=1&danmaku=0&autoplay=0`;
  }
  return ''; 
};

// 3. 弹窗与表单状态
const videoDialogVisible = ref(false);
const tempVideoList = ref<string[]>([]);

const openVideoDialog = () => {
  tempVideoList.value = [...parsedVideos.value];
  if (tempVideoList.value.length === 0) {
    tempVideoList.value.push('');
  }
  videoDialogVisible.value = true;
};

const addVideoRow = () => {
  if (tempVideoList.value.length >= 10) {
    ElMessage.warning('最多添加 10 个视频');
    return;
  }
  tempVideoList.value.push('');
};

const removeVideoRow = (index: number) => {
  tempVideoList.value.splice(index, 1);
};

// 从列表中删除视频并保存
const removeVideoByIndex = (index: number) => {
  const newList = [...parsedVideos.value];
  newList.splice(index, 1);
  const jsonStr = JSON.stringify(newList);
  emit('save-video', jsonStr);
};

// 在 DetailPanel.vue 的 <script setup> 中

const submitVideo = () => {
  const validList = tempVideoList.value
    .map(v => v.trim())
    .filter(v => v !== '')
    .map(rawInput => {
      // 1. 如果是 B 站 iframe 代码，提取 src
      if (rawInput.includes('<iframe')) {
        const srcMatch = rawInput.match(/src=["'](.*?)["']/);
        if (srcMatch) return srcMatch[1];
      }

      // 2. 如果包含 .mp4 (直链)，直接保存，不进行 B 站正则处理
      // ★★★ 新增逻辑 ★★★
      if (rawInput.toLowerCase().includes('.mp4')) {
        return rawInput;
      }

      // 3. 尝试 B 站正则提取 (BV号)
      const bvRegex = /(BV[a-zA-Z0-9]{10})/;
      const match = rawInput.match(bvRegex);
      if (match) {
        const bvid = match[1];
        return `//player.bilibili.com/player.html?bvid=${bvid}&page=1&high_quality=1&danmaku=0&autoplay=0`;
      }

      // 4. 其他情况，原样保存
      return rawInput;
    });

  const jsonStr = JSON.stringify(validList);
  emit('save-video', jsonStr);
  videoDialogVisible.value = false;
};


// 4. 悬浮播放器控制
const playDialogVisible = ref(false);
const currentPlayUrl = ref('');
const isResizing = ref(false); // 拖拽状态

// 在 DetailPanel.vue 的 <script setup> 中

const openFloatingPlayer = (url: string) => {
  if (!url) return;
  
  if (url.toLowerCase().includes('.mp4')) {
    currentPlayUrl.value = url;
  } else {
    // 处理 B站 iframe 链接
    let playUrl = url;

    // 1. 强制 autoplay=0 (如果已有 autoplay=1 则替换，没有则追加)
    if (playUrl.includes('autoplay=')) {
        playUrl = playUrl.replace(/autoplay=1/g, 'autoplay=0');
    } else {
        playUrl += (playUrl.includes('?') ? '&' : '?') + 'autoplay=0';
    }

    // 2. 强制 danmaku=0
    if (!playUrl.includes('danmaku=')) {
        playUrl += (playUrl.includes('?') ? '&' : '?') + 'danmaku=0';
    }

    currentPlayUrl.value = playUrl;
  }

  playDialogVisible.value = true;
};

// ★★★ AI 面试官状态 ★★★
const aiInterviewerVisible = ref(false);
const isAIConnected = ref(false);
const isAIConnecting = ref(false);
const aiRemainingQuota = ref(0);
const aiMessages = ref<{ role: string; content: string }[]>([]);
const aiUserInput = ref('');
const isAILoading = ref(false);
const aiChatContainerRef = ref<HTMLElement | null>(null);

// 打开 AI 面试官弹窗
const openAIInterviewer = () => {
  aiInterviewerVisible.value = true;
};

// 监听 PointEditor 的 openAIInterviewer 事件
const handleOpenAIInterviewer = () => {
  openAIInterviewer();
};

// 格式化时间
const formatTime = (seconds: number) => {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
};

// 重新连接（占位）
const reconnectAIInterviewer = () => {
  console.log('重新连接 AI 面试官');
};

// 发送消息（占位）
const sendAIMessage = () => {
  console.log('发送 AI 消息');
};

// 重置面试（占位）
const resetAIInterview = () => {
  console.log('重置 AI 面试');
};

// 在 mounted 时监听事件
import { onMounted, onBeforeUnmount } from 'vue';
onMounted(() => {
  // 通过事件总线监听 PointEditor 发出的事件
  window.addEventListener('open-ai-interviewer', handleOpenAIInterviewer);
});

onBeforeUnmount(() => {
  window.removeEventListener('open-ai-interviewer', handleOpenAIInterviewer);
});
</script>

<style scoped>
/* ================= 1. 整体容器 ================= */
.detail-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.8); /* 整体微透背景 */
  backdrop-filter: blur(20px);
  position: relative;
  overflow: hidden;
}

.empty-state {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
}

/* ================= 2. 头部区域 (新版：左右紧凑布局) ================= */
.detail-header {
  padding: 15px 25px;
  border-bottom: 2px solid #e4e7ed;
  background: linear-gradient(to bottom, #fafbfc 0%, #f5f7fa 100%);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  flex-shrink: 0; /* 防止被挤压 */
  border-radius: 8px 8px 0 0; /* 只有顶部两个角是圆角 */
  margin-bottom: 10px; /* 增加与下方内容的间距 */
}

/* 上半部分：标题与按钮 */
.header-top-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px; 
}

.point-title {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0; /* 防止标题过长擑开 */
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: #909399;
  font-size: 13px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s;
  flex-shrink: 0;
}
.back-link:hover {
  color: #409eff;
  background: rgba(64, 158, 255, 0.1);
}
.back-link .el-icon {
  font-size: 14px;
}

.title-text {
  font-size: 20px;
  font-weight: 700;
  color: #1a1a1a;
  line-height: 1.4;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.title-edit-icon {
  font-size: 16px;
  color: #909399;
  cursor: pointer;
  margin-left: 6px;
  transition: all 0.2s;
}
.title-edit-icon:hover {
  color: #409eff;
}

.diff-tag {
  font-weight: normal;
}

/* ================= 3. 信息栏 (视频 + 链接) ================= */
.header-info-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 20px;
  flex-wrap: wrap;
}

/* 左侧视频区域 */
.info-left-video {
  flex: 1;
  min-width: 300px;
}

.video-compact-section {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.section-label {
  font-size: 13px;
  color: #606266;
  font-weight: 500;
  flex-shrink: 0;
}

.video-label {
  margin-right: 8px;
}

.video-mini-list {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.mini-video-wrapper {
  width: 40px;
  height: 30px;
  border-radius: 4px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}

.mini-video-wrapper:hover {
  transform: scale(1.05);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.placeholder-icon {
  color: #fff;
  font-size: 16px;
}

.bilibili-icon-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #fb7299 0%, #ff4d7d 100%);
  border-radius: 4px;
}

.bili-text {
  color: #fff;
  font-size: 12px;
  font-weight: bold;
}

.add-video-btn {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: #ecf5ff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}

.add-video-btn:hover {
  background: #409eff;
  color: #fff;
}

.no-video-text {
  font-size: 12px;
  color: #909399;
  font-style: italic;
}

/* 右侧链接区域 */
.info-right-links {
  flex: 1;
  min-width: 300px;
}

.links-section {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  flex-wrap: wrap;
}

.link-icon {
  font-size: 14px;
  color: #409eff;
  margin-top: 2px;
  flex-shrink: 0;
}

.link-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.link-item-wrapper {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: #f0f2f5;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.link-item {
  color: #409eff;
  text-decoration: none;
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.link-item:hover {
  text-decoration: underline;
}

.remove-link-icon {
  font-size: 12px;
  color: #909399;
  cursor: pointer;
  transition: color 0.2s;
}

.remove-link-icon:hover {
  color: #f56c6c;
}

/* ================= 4. 主体内容布局 ================= */
.detail-body-layout {
  display: flex;
  gap: 15px;
  flex: 1;
  overflow: hidden;
  padding: 0;
}

.panel-column {
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  flex: 1;
}

.column-header {
  padding: 12px 15px;
  border-bottom: 1px solid #ebeef5;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

.col-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.editor-column {
  flex: 2;
  min-width: 0;
}

.image-column {
  flex: 1;
  min-width: 250px;
  max-width: 300px;
}

.column-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* ================= 5. 视频管理弹窗 ================= */
.video-manage-tip {
  font-size: 12px;
  color: #909399;
  margin-bottom: 15px;
  padding: 8px 12px;
  background: #f8f9fa;
  border-radius: 4px;
  border-left: 3px solid #409eff;
}

.video-list-edit {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.video-edit-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.row-index {
  font-size: 13px;
  color: #909399;
  width: 20px;
  flex-shrink: 0;
}

.add-row-btn {
  align-self: flex-start;
  margin-top: 5px;
}

/* ================= 6. 悬浮播放器 ================= */
.resizable-video-dialog {
  position: fixed !important;
  right: 20px !important;
  bottom: 20px !important;
  width: auto !important;
  height: auto !important;
  margin: 0 !important;
}

.resizable-video-dialog .el-dialog__body {
  padding: 0 !important;
  overflow: hidden !important;
}

.resizable-wrapper {
  position: relative;
  width: 400px;
  height: 225px;
  resize: both;
  overflow: hidden;
  border: 1px solid #ddd;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.resize-mask {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 9999;
  cursor: se-resize;
}

.video-overlay-transparent {
  background-color: transparent !important;
}

/* ================= 7. AI 面试官弹窗样式 ================= */
.full-width-one-third-height-dialog {
  position: fixed !important;
  top: 0 !important;
  left: 0 !important;
  width: 100vw !important;
  height: 33vh !important; /* 高度为屏幕的 1/3 */
  margin: 0 !important;
  padding: 0 !important;
  border-radius: 0 !important;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15) !important;
}

.full-width-one-third-height-dialog .el-dialog__header {
  display: none !important;
}

.full-width-one-third-height-dialog .el-dialog__body {
  height: 100% !important;
  padding: 16px 20px !important;
  overflow: hidden !important;
}

.ai-interviewer-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.interviewer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 12px;
  border-bottom: 1px solid #ebeef5;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.point-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-icon {
  font-size: 18px;
  color: #764ba2;
}

.point-title {
  font-size: 15px;
  font-weight: 500;
  color: #303133;
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.quota-tag {
  font-weight: 500;
}

.chat-container {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
  margin-bottom: 12px;
}

.empty-chat {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #909399;
}

.empty-chat p {
  margin-top: 12px;
  font-size: 14px;
}

.message-item {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.message-item.user {
  flex-direction: row-reverse;
}

.message-avatar {
  flex-shrink: 0;
}

.ai-avatar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.user-avatar {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
}

.message-content {
  max-width: 70%;
}

.message-bubble {
  padding: 12px 16px;
  border-radius: 12px;
  font-size: 14px;
  line-height: 1.6;
  word-break: break-word;
}

.message-item.assistant .message-bubble {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 12px 12px 12px 4px;
}

.message-item.user .message-bubble {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  border-radius: 12px 12px 4px 12px;
}

.message-bubble.loading {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 16px 20px;
}

.loading-dot {
  width: 8px;
  height: 8px;
  background: #909399;
  border-radius: 50%;
  animation: bounce 1.4s infinite ease-in-out both;
}

.loading-dot:nth-child(1) { animation-delay: -0.32s; }
.loading-dot:nth-child(2) { animation-delay: -0.16s; }

@keyframes bounce {
  0%, 80%, 100% { transform: scale(0); }
  40% { transform: scale(1); }
}

.input-container {
  border-top: 1px solid #ebeef5;
  padding-top: 12px;
  flex-shrink: 0;
}

.input-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 10px;
}

.gradient-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  color: #fff;
}

.gradient-btn:hover {
  background: linear-gradient(135deg, #5a6fd6 0%, #6a4190 100%);
}

.mr-1 {
  margin-right: 4px;
}

.connection-error {
  margin-bottom: 12px;
}

.message-item.system .message-bubble {
  background: #fef0f0;
  border: 1px solid #fbc4c4;
  color: #f56c6c;
  border-radius: 8px;
}

.is-loading {
  animation: rotate 1.5s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>


<!-- ★★★ 全局穿透与样式修正 (无 scoped) ★★★ -->
<style>
/* 1. 穿透遮罩层 */
.video-overlay-transparent {
  pointer-events: none !important;
  background-color: transparent !important;
  overflow: hidden !important;
}

/* 2. 针对弹窗本体 (恢复白色背景) */
.video-overlay-transparent .el-dialog {
  pointer-events: auto !important;
  margin: 0 !important;
  
  /* ★★★ 改回白色背景 ★★★ */
  background: #fff !important; 
  border-radius: 6px !important;
  box-shadow: 0 10px 40px rgba(0,0,0,0.5) !important;
  
  display: flex !important;
  flex-direction: column !important;
  width: auto !important;
}

/* 3. 恢复标题栏样式 (白色背景) */
.video-overlay-transparent .el-dialog__header {
  padding: 15px 20px !important; /* 增加一点内边距让它更好看 */
  background: #fff !important;   /* ★★★ 白色背景 ★★★ */
  border-bottom: 1px solid #eee !important; /* 加个浅灰分割线 */
  margin: 0 !important;
  flex-shrink: 0;
  cursor: move !important; /* 鼠标变成移动图标 */
  user-select: none;
}

/* 标题文字颜色改回深色 */
.video-overlay-transparent .el-dialog__title {
  color: #303133 !important; /* 深灰色字体 */
  font-size: 16px !important;
  font-weight: 600 !important;
}

/* 关闭按钮颜色改回深色 */
.video-overlay-transparent .el-dialog__headerbtn {
  top: 18px !important;
}
.video-overlay-transparent .el-dialog__headerbtn .el-dialog__close {
  color: #909399 !important;
  font-size: 16px !important;
}
.video-overlay-transparent .el-dialog__headerbtn:hover .el-dialog__close {
  color: #409eff !important; /* hover 变蓝 */
}

/* 4. 内容区域 (Body) */
.video-overlay-transparent .el-dialog__body {
  /* ★★★ 这里加上 padding，就有了你想要的白边！★★★ */
  padding: 10px !important; 
  margin: 0 !important;
  background: #fff !important; /* 背景也是白的 */
  
  flex: 1;
  display: flex; 
  font-size: 0;
  height: auto !important;
}

/* 5. 针对 flex 布局容器 */
.video-overlay-transparent .el-overlay-dialog {
  pointer-events: none !important;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>


