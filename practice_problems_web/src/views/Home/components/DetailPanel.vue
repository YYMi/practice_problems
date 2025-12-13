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

  </main>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { EditPen, Delete, VideoPlay, Link, Close, Plus, Edit, Back } from "@element-plus/icons-vue";
import PointEditor from "../../../components/PointEditor.vue";
import ImageManager from "../../../components/ImageManager.vue";
import QuestionDrawer from "../../../components/QuestionDrawer.vue";
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
  flex-shrink: 0;
}

/* 难度标签颜色 */
.diff-easy {
  --el-tag-bg-color: #e1f3d8;
  --el-tag-border-color: #b3e19d;
  --el-tag-text-color: #67c23a;
}
.diff-medium {
  --el-tag-bg-color: #faecd8;
  --el-tag-border-color: #f3d19e;
  --el-tag-text-color: #e6a23c;
}
.diff-hard {
  --el-tag-bg-color: #fde2e2;
  --el-tag-border-color: #fab6b6;
  --el-tag-text-color: #f56c6c;
}
.diff-important {
  --el-tag-bg-color: #e9e4f0;
  --el-tag-border-color: #d4b9e9;
  --el-tag-text-color: #9b59b6;
}

.header-actions {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
}

/* 下半部分：左右分栏信息 */
.header-info-row {
  display: flex;
  align-items: center; /* 垂直居中对齐 */
  justify-content: space-between;
  gap: 20px;
}

/* 左侧：视频列表 */
.info-left-video {
  flex-shrink: 0;
}

/* 右侧：链接列表 */
.info-right-links {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.links-section {
  display: flex;
  align-items: flex-start;
  flex-wrap: wrap; /* 允许换行 */
  gap: 8px;
  font-size: 13px;
  color: #666;
}

.link-icon {
  flex-shrink: 0;
  margin-top: 2px;
}

.link-list {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.link-item-wrapper {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  background: rgba(64, 158, 255, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
  transition: all 0.2s;
}
.link-item-wrapper:hover {
  background: rgba(64, 158, 255, 0.2);
}

.link-item {
  color: #409eff;
  text-decoration: none;
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.link-item:hover {
  text-decoration: underline;
}

.remove-link-icon {
  font-size: 14px;
  color: #f56c6c;
  cursor: pointer;
  padding: 2px;
  border-radius: 2px;
  transition: all 0.2s;
  flex-shrink: 0;
}
.remove-link-icon:hover {
  background: rgba(245, 108, 108, 0.2);
  transform: scale(1.1);
}

/* 右侧：视频列表 (紧凑型) */
.video-compact-section {
  display: flex;
  align-items: center;
  gap: 10px;
}

.video-label {
  font-size: 12px;
  color: #909399;
}

.video-mini-list {
  display: flex;
  gap: 6px;
  align-items: center;
}

/* 修改 .mini-video-wrapper 样式，移除 hover 放大过多的效果，保持整洁 */
.mini-video-wrapper {
  width: 50px;
  height: 28px;
  border-radius: 6px; /* 稍微圆润一点 */
  overflow: hidden;
  position: relative;
  background: #2b2b2b; /* 深灰偏黑背景，质感更好 */
  border: 1px solid #dcdfe6;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}
/* 鼠标悬停时，三角形变亮或变色，增加交互感 */
.mini-video-wrapper:hover .bili-text {
  border-color: transparent transparent transparent #409eff; /* 悬停变蓝 */
  transform: translateX(1px) scale(1.1);
}
.mini-video-wrapper:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(0,0,0,0.15);
}
/* 占位符样式 */
.video-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}
.placeholder-icon {
  color: #fff;
  font-size: 16px;
}

/* 移除之前的粉色背景，改为透明或深色渐变 */
.bilibili-icon-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #333 0%, #444 100%); /* 增加一点微弱的渐变质感 */
}

/* 绘制中间的“播放三角” */
.bili-text {
  /* 清除之前的文字样式 */
  font-size: 0; 
  color: transparent;
  
  /* 用 CSS 绘制三角形 */
  width: 0;
  height: 0;
  border-style: solid;
  border-width: 5px 0 5px 8px; /* 控制三角形大小 */
  border-color: transparent transparent transparent #ffffff; /* 白色三角形 */
  opacity: 0.9;
  transform: translateX(1px); /* 视觉上居中修正 */
}

.mini-content {
  width: 200%;
  height: 200%;
  transform: scale(0.5);
  transform-origin: 0 0;
  pointer-events: none;
  object-fit: cover;
  display: block;
}

.click-mask {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  z-index: 10;
  background: transparent;
}

.add-video-btn {
  width: 28px;
  height: 28px;
  border: 1px dashed #c0c4cc;
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #909399;
  font-size: 12px;
  transition: all 0.2s;
}
.add-video-btn:hover {
  border-color: #409eff;
  color: #409eff;
  background: rgba(64,158,255,0.05);
}

.no-video-text {
  font-size: 12px;
  color: #c0c4cc;
}

/* ================= 3. 内容主体区域 ================= */
.detail-body {  flex: 1;
  overflow: hidden;
  position: relative;
  display: flex;
  flex-direction: column;
}

/* 主体布局 */
.detail-body-layout {
  flex: 1;
  display: flex;
  gap: 15px;
  overflow: hidden;
}

/* 左右两栏 */
.panel-column {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: rgba(255,255,255,0.5);
  border-radius: 8px;
  backdrop-filter: blur(10px);
}

.editor-column {
  flex: 2;
}

.image-column {
  flex: 1;
  min-width: 300px;
}

/* 标题栏 - 固定在顶部 */
.column-header {
  flex-shrink: 0;
  padding: 14px 20px;
  background: linear-gradient(to bottom, #fafbfc 0%, #f5f7fa 100%);
  border-bottom: 2px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  display: flex;
  align-items: center;
  gap: 10px;
  position: sticky;
  top: 0;
  z-index: 10;
}

.col-title {
  font-weight: 600;
  font-size: 16px;
  color: #303133;
  letter-spacing: 0.5px;
}

/* 内容区 - 可滚动 */
.column-content {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

/* 左右分栏布局 (左:内容 右:图片) */
.editor-layout {
  display: flex;
  flex: 1;
  height: 100%;
  overflow: hidden;
}

/* --- 左侧内容区 --- */
.view-area {
  flex: 1;
  padding: 20px 40px;
  overflow-y: auto;
  position: relative;
}

.markdown-body {
  line-height: 1.8;
  font-size: 15px;
  color: #333;
}
.markdown-body :deep(h1), .markdown-body :deep(h2) {
  border-bottom: none;
  color: #303133;
}
.markdown-body :deep(p) {
  margin-bottom: 16px;
}

/* 水印 */
.watermark {
  position: absolute;
  top: 20px;
  right: 20px;
  font-size: 12px;
  color: rgba(0,0,0,0.05);
  pointer-events: none;
  user-select: none;
  font-weight: bold;
}

/* --- 右侧图片管理区 --- */
.image-manager {
  width: 300px;
  border-left: 1px solid rgba(0,0,0,0.05);
  background: rgba(250, 250, 250, 0.5);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.img-mgr-header {
  padding: 10px 15px;
  border-bottom: 1px solid rgba(0,0,0,0.05);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(255,255,255,0.6);
}
.img-mgr-title {
  font-weight: 600;
  font-size: 14px;
  color: #606266;
}

.image-grid {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.image-item {
  margin-bottom: 15px;
  border: 1px solid #eee;
  border-radius: 8px;
  background: #fff;
  padding: 8px;
  transition: all 0.3s;
}
.image-item:hover {
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
  transform: translateY(-2px);
}
.img-preview {
  width: 100%;
  height: 120px;
  object-fit: contain;
  background: #f5f7fa;
  border-radius: 4px;
  cursor: zoom-in;
}
.img-name {
  margin-top: 8px;
  font-size: 12px;
  color: #606266;
  word-break: break-all;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.img-actions {
  margin-top: 8px;
  display: flex;
  justify-content: flex-end;
}

/* ================= 4. 浮动播放器内部样式 ================= */
.resizable-wrapper {
  width: 800px;
  min-width: 400px;
  min-height: 225px;
  /* 强制 16:9，高度自动算 */
  aspect-ratio: 16 / 9;
  height: auto !important; 
  background: #000;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  /* ★★★ 启用拖拽调整大小 ★★★ */
  resize: both;
  overflow: auto;
}

/* 调整大小时的透明遮罩 (防止 iframe 吞鼠标) */
.resize-mask {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  z-index: 998;
  background: transparent;
}

/* 右下角拖拽手柄 (可选优化) */
.resizable-wrapper::after {
  content: '';
  position: absolute;
  bottom: 0;
  right: 0;
  width: 15px;
  height: 15px;
  cursor: se-resize;
  z-index: 999;
  /* 一个小三角暗示可以拖拽 */
  background: linear-gradient(135deg, transparent 50%, rgba(255,255,255,0.5) 50%);
  pointer-events: auto;
}

/* 视频管理弹窗样式 */
.video-manage-tip {
  font-size: 13px;
  color: #909399;
  margin-bottom: 15px;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.video-list-edit {
  max-height: 400px;
  overflow-y: auto;
}

.video-edit-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.row-index {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
  min-width: 24px;
  flex-shrink: 0;
}

.video-edit-row .el-input {
  flex: 1;
}

.video-edit-row .el-button {
  flex-shrink: 0;
}

.add-row-btn {
  margin-top: 10px;
}

/* 滚动条美化 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
::-webkit-scrollbar-thumb {
  background: rgba(0,0,0,0.1);
  border-radius: 3px;
}
::-webkit-scrollbar-track {
  background: transparent;
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


