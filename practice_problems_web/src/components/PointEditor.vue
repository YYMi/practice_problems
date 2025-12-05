<template>
  <div class="content-column">
    <!-- 顶部标题与操作栏 -->
    <div class="section-header">
      <div class="left-group">
        <div class="section-title">
          <el-icon class="mr-1"><Reading /></el-icon> 知识点详解
        </div>

        <!-- 【模块1：播放控制条】 (固定在标题右侧) -->
        <!-- 无论是否播放都显示，停止状态下按钮禁用，防止布局跳动 -->
        <div class="player-module" v-if="!isEditing && content">
          
          <!-- 1. 语速滑块 -->
          <div class="speed-box">
            <span class="speed-label">{{ speechRate.toFixed(1) }}x</span>
            <el-slider 
              v-model="speechRate" 
              :min="0.1" 
              :max="3.0" 
              :step="0.1" 
              size="small" 
              class="custom-slider"
              @change="handleRateChange" 
            />
          </div>

          <div class="divider-small"></div>

          <!-- 2. 暂停/继续 (是一个按钮，切换图标) -->
          <el-tooltip :content="speechStatus === 'paused' ? '继续' : '暂停'" placement="top">
            <el-button 
              link 
              class="control-btn"
              :class="speechStatus === 'playing' ? 'warning' : 'success'"
              :disabled="speechStatus === 'stopped'"
              @click="togglePauseResume"
            >
              <el-icon size="16">
                <VideoPlay v-if="speechStatus === 'paused'" />
                <VideoPause v-else />
              </el-icon>
            </el-button>
          </el-tooltip>

          <!-- 3. 停止 -->
          <el-tooltip content="停止" placement="top">
            <el-button 
              link 
              class="control-btn danger"
              :disabled="speechStatus === 'stopped'"
              @click="handleStop"
            >
              <el-icon size="16"><SwitchButton /></el-icon>
            </el-button>
          </el-tooltip>
        </div>
      </div>
      
      <!-- 右侧操作区 -->
      <div class="action-area">
        
        <!-- 【模块2：朗读触发按钮】 (在编辑按钮左边) -->
        <div class="trigger-group" v-if="!isEditing && content">
          
          <!-- A. 朗读选中 -->
          <el-tooltip 
            :content="selectedText ? '朗读选中的文字' : '请先在下方选择文字'" 
            placement="top"
          >
            <el-button 
              link 
              class="trigger-btn"
              :class="(readingMode === 'full' || !selectedText) ? 'disabled-text' : 'primary-text'"
              :disabled="readingMode === 'full' || !selectedText"
              @click="startSelectedReading"
            >
              <el-icon class="mr-1"><ChatLineSquare /></el-icon> 朗读
            </el-button>
          </el-tooltip>

          <!-- B. 全文朗读 -->
          <el-tooltip content="从头朗读全文" placement="top">
            <el-button 
              link 
              class="trigger-btn"
              :class="readingMode === 'full' ? 'highlight-text' : 'primary-text'"
              @click="startFullReading"
            >
              <el-icon class="mr-1"><Microphone /></el-icon> 全文朗读
            </el-button>
          </el-tooltip>
        </div>

        <!-- 分割线 -->
        <div class="divider-vertical" v-if="(!isEditing && content) && canEdit"></div>

        <!-- 【模块3：编辑按钮】 -->
        <div class="edit-controls" v-if="canEdit">
          <el-button 
            v-if="!isEditing" 
            type="primary" 
            size="small" 
            icon="Edit" 
            class="gradient-btn" 
            @click="startEdit"
          >
            编辑内容
          </el-button>
          <div v-else class="edit-actions">
            <el-button size="small" @click="cancelEdit" class="cancel-btn">取消</el-button>
            <el-button 
              type="primary" 
              size="small" 
              icon="Check" 
              class="gradient-btn" 
              @click="saveEdit"
            >
              保存
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <div class="content-box custom-scrollbar">
      <div v-if="isEditing" class="editor-wrapper">
        <Toolbar
          style="border-bottom: 1px solid rgba(0,0,0,0.05)"
          :editor="editorRef"
          :defaultConfig="toolbarConfig"
          :mode="mode"
        />
        <Editor
          style="flex: 1; overflow-y: hidden;"
          v-model="innerContent"
          :defaultConfig="editorConfig"
          :mode="mode"
          @onCreated="handleCreated"
        />
      </div>

      <!-- 预览模式 -->
      <div 
        v-else 
        class="html-preview" 
        ref="previewRef" 
        @mouseup="captureSelection" 
        @touchend="captureSelection"
      >
        <div v-if="content" v-html="content" class="markdown-body"></div>
        <div v-else class="empty-tip">
          <el-icon :size="40"><Edit /></el-icon>
          <p>暂无详细内容，请点击右上角编辑开始录入</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, onBeforeUnmount, watch, onMounted } from "vue";
import { ElMessage } from "element-plus";
import { Editor, Toolbar } from "@wangeditor/editor-for-vue";
import { Microphone, VideoPause, VideoPlay, SwitchButton, Reading, Edit, Check, ChatLineSquare } from '@element-plus/icons-vue';
import { uploadImage, updatePoint } from "../api/point";
import '@wangeditor/editor/dist/css/style.css'; 

const props = defineProps({
  pointId: { type: Number, required: true },
  content: { type: String, default: '' },
  canEdit: { type: Boolean, default: false }
});

const emit = defineEmits(["update"]);

// 编辑器相关
const editorRef = shallowRef();
const mode = "default";
const isEditing = ref(false);
const innerContent = ref("");
const previewRef = ref<HTMLElement | null>(null); 

// 语音相关
type SpeechStatus = 'stopped' | 'playing' | 'paused';
type ReadingMode = 'full' | 'selected' | 'none'; 

const speechStatus = ref<SpeechStatus>('stopped');
const readingMode = ref<ReadingMode>('none'); 
const speechRate = ref(1.0);
const selectedText = ref(""); 
const synth = window.speechSynthesis;
let utterance: SpeechSynthesisUtterance | null = null;

// 续播记录
const currentCharIndex = ref(0); 
const currentFullText = ref(""); 

onMounted(() => {
  const savedRate = localStorage.getItem('user-speech-rate');
  if (savedRate) speechRate.value = parseFloat(savedRate);
  document.addEventListener('click', handleGlobalClick);
});

onBeforeUnmount(() => {
  handleStop();
  document.removeEventListener('click', handleGlobalClick);
  const editor = editorRef.value;
  if (editor == null) return;
  editor.destroy();
});

const handleGlobalClick = (e: MouseEvent) => {};

const stripHtml = (html: string) => {
  const tmp = document.createElement("DIV");
  tmp.innerHTML = html;
  return tmp.textContent || tmp.innerText || "";
};

/**
 * 暂停/继续 切换逻辑
 */
const togglePauseResume = () => {
  if (speechStatus.value === 'playing') {
    synth.pause();
    speechStatus.value = 'paused';
  } else if (speechStatus.value === 'paused') {
    synth.resume();
    speechStatus.value = 'playing';
  }
};

const handleRateChange = (val: number) => {
  localStorage.setItem('user-speech-rate', val.toString()); 
  if (speechStatus.value === 'playing' || speechStatus.value === 'paused') {
    const originalText = currentFullText.value;
    const offset = currentCharIndex.value;
    if (originalText && offset < originalText.length) {
      const remainingText = originalText.substring(offset);
      synth.cancel();
      setTimeout(() => {
        const mode = readingMode.value; 
        speak(remainingText, mode);
      }, 50);
    }
  }
};

const captureSelection = () => {
  if (isEditing.value) return;
  const selection = window.getSelection();
  const previewDom = previewRef.value;

  if (!selection || selection.rangeCount === 0 || !previewDom) {
    selectedText.value = "";
    return;
  }
  if (!previewDom.contains(selection.anchorNode)) return;

  const text = selection.toString().trim();
  selectedText.value = text;

  // 跳转逻辑
  if (speechStatus.value === 'playing' || speechStatus.value === 'paused') {
    const userRange = selection.getRangeAt(0);
    const rangeToEnd = document.createRange();
    rangeToEnd.selectNodeContents(previewDom); 
    rangeToEnd.setStart(userRange.startContainer, userRange.startOffset); 
    
    const textToRead = rangeToEnd.toString();

    if (textToRead && textToRead.trim().length > 0) {
      if (synth.speaking) synth.cancel();
      readingMode.value = 'full';
      ElMessage.success("已跳转至选定位置播放");
      speak(textToRead, 'full'); 
    }
  }
};

const startFullReading = () => {
  const text = stripHtml(props.content);
  speak(text, 'full');
};

const startSelectedReading = () => {
  if (readingMode.value === 'full') return;
  if (!selectedText.value) return;
  speak(selectedText.value, 'selected');
};

const speak = (text: string, mode: ReadingMode) => {
  if (!text.trim()) {
    ElMessage.warning("没有可朗读的文本");
    return;
  }

  if (synth.speaking) synth.cancel();

  readingMode.value = mode;
  currentFullText.value = text; 
  currentCharIndex.value = 0;   

  utterance = new SpeechSynthesisUtterance(text);
  utterance.lang = 'zh-CN'; 
  utterance.rate = speechRate.value; 
  
  utterance.onboundary = (event) => {
    if (event.name === 'word' || event.name === 'sentence') {
      currentCharIndex.value = event.charIndex;
    }
  };

  utterance.onend = () => { 
    speechStatus.value = 'stopped'; 
    readingMode.value = 'none'; 
    currentCharIndex.value = 0;
  };
  
  utterance.onerror = (e) => { 
    if (e.error !== 'interrupted') {
      speechStatus.value = 'stopped'; 
      readingMode.value = 'none';
    }
  };
  
  synth.speak(utterance);
  speechStatus.value = 'playing';
};

// 复用 togglePauseResume, 这里保留 handleStop
const handleStop = () => { 
  synth.cancel(); 
  speechStatus.value = 'stopped'; 
  readingMode.value = 'none'; 
  currentCharIndex.value = 0;
  currentFullText.value = "";
};

// ... watch 和编辑器逻辑不变 ...
watch(() => props.content, (newVal) => {
  if (!isEditing.value) innerContent.value = newVal || "";
  handleStop();
}, { immediate: true });

watch(isEditing, (newVal) => {
  if (newVal) {
    handleStop();
    innerContent.value = props.content || "";
  }
});

const imgBaseUrl = import.meta.env.VITE_IMG_BASE_URL;
const toolbarConfig = {};
const editorConfig = {
  placeholder: "请输入内容...",
  MENU_CONF: {
    uploadImage: {
      async customUpload(file: File, insertFn: any) {
        try {
          const res = await uploadImage(file,props.pointId);
          if (res.data.code === 200) {
            const url = `${imgBaseUrl+res.data.data.path}`;
            insertFn(url, res.data.data.url, url);
          }
        } catch (e) {
          ElMessage.error("图片上传失败");
        }
      },
    },
  },
};

const handleCreated = (editor: any) => {
  editorRef.value = editor;
  if (innerContent.value) {
    editor.setHtml(innerContent.value);
  }
};

const startEdit = () => {
  innerContent.value = props.content || "";
  isEditing.value = true;
};
const cancelEdit = () => {
  isEditing.value = false;
  innerContent.value = props.content || "";
};
const saveEdit = async () => {
  try {
    await updatePoint(props.pointId, { content: innerContent.value });
    emit("update", innerContent.value);
    isEditing.value = false;
    ElMessage.success("保存成功");
  } catch (e) {
    ElMessage.error("保存失败");
  }
};
</script>

<style scoped>
.content-column { flex: 3; display: flex; flex-direction: column; padding-right: 5px; height: 100%; background: transparent; }

/* 头部布局 */
.section-header { 
  display: flex; 
  justify-content: space-between; 
  align-items: center; 
  margin-bottom: 15px; 
  flex-shrink: 0; 
  border-bottom: 1px solid rgba(0,0,0,0.05); 
  padding-bottom: 10px; 
}

/* 左侧组：标题 + 播放控制模块 */
.left-group {
  display: flex;
  align-items: center;
  gap: 20px;
}

.section-title { font-weight: bold; color: #303133; font-size: 18px; display: flex; align-items: center; }
.mr-1 { margin-right: 6px; }

/* 【核心修改】播放控制模块：固定胶囊样式 */
.player-module {
  display: flex;
  align-items: center;
  background: rgba(242, 243, 245, 0.8); /* 浅灰色背景 */
  border-radius: 20px;
  padding: 4px 12px;
  height: 32px;
  gap: 8px;
  transition: all 0.3s;
}
/* 停止时稍微变透明一点，表示非活跃 */
.player-module:hover {
  background: rgba(230, 232, 235, 0.9);
}

/* 语速区域 */
.speed-box { display: flex; align-items: center; }
.speed-label { font-size: 12px; color: #606266; margin-right: 8px; width: 24px; text-align: right; }
.custom-slider { width: 60px; }

.divider-small { width: 1px; height: 14px; background-color: #dcdfe6; margin: 0 4px; }

/* 控制按钮 */
.control-btn { padding: 0; height: auto; }
.control-btn.warning { color: #e6a23c; }
.control-btn.success { color: #67c23a; }
.control-btn.danger { color: #f56c6c; }
/* 禁用状态下的图标颜色 */
.control-btn.is-disabled { color: #c0c4cc; }


/* 右侧组：朗读触发 + 编辑 */
.action-area { display: flex; align-items: center; gap: 15px; }

.trigger-group { display: flex; gap: 10px; }
.trigger-btn { font-weight: 600; font-size: 14px; }
.primary-text { color: #606266; }
.primary-text:hover { color: #409eff; }
.highlight-text { color: #764ba2; font-weight: bold; }
.disabled-text { color: #c0c4cc; cursor: not-allowed; }


.divider-vertical { width: 1px; height: 16px; background-color: #e4e7ed; }
.gradient-btn { background: linear-gradient(90deg, #667eea, #764ba2); border: none; box-shadow: 0 2px 6px rgba(118, 75, 162, 0.3); padding: 8px 18px; font-weight: 600; }
.gradient-btn:hover { opacity: 0.9; transform: translateY(-1px); }
.cancel-btn { color: #909399; }
.cancel-btn:hover { color: #764ba2; background-color: rgba(118, 75, 162, 0.1); }
.edit-actions { display: flex; gap: 8px; }

.content-box { flex: 1; display: flex; flex-direction: column; min-height: 0; }
.editor-wrapper { flex: 1; display: flex; flex-direction: column; border: 1px solid rgba(0,0,0,0.1); border-radius: 8px; overflow: hidden; background: rgba(255,255,255,0.6); z-index: 10; }
:deep(.w-e-toolbar) { background-color: rgba(249, 250, 251, 0.9) !important; }
:deep(.w-e-text-container) { background-color: transparent !important; }
:deep(.w-e-bar-item button:hover) { color: #764ba2; }
.html-preview { flex: 1; padding: 5px; line-height: 1.8; color: #333; font-size: 15px; overflow-y: auto; cursor: text; }
.markdown-body :deep(p) { margin-bottom: 12px; }
.markdown-body :deep(img) { max-width: 100%; border-radius: 6px; box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
.markdown-body :deep(blockquote) { border-left: 4px solid #d3adf7; background: rgba(249, 240, 255, 0.5); padding: 10px 15px; margin: 10px 0; color: #666; border-radius: 4px; }
.markdown-body :deep(code) { background-color: rgba(0,0,0,0.05); padding: 2px 5px; border-radius: 4px; font-family: monospace; color: #c7254e; }
.markdown-body :deep(pre) { background-color: #f1f1f1; color: #333; padding: 10px; border-radius: 4px; border: 1px solid #ccc; }
.empty-tip { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%; color: rgba(0,0,0,0.3); margin-top: 40px; }
.empty-tip p { margin-top: 10px; font-size: 14px; }
.custom-scrollbar::-webkit-scrollbar { width: 6px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(0,0,0,0.1); border-radius: 3px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
</style>