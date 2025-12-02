<template>
  <div class="content-column">
    <!-- 顶部标题与操作栏 -->
    <div class="section-header">
      <div class="section-title">
        <el-icon class="mr-1"><Reading /></el-icon> 知识点详解
      </div>
      
      <!-- 右侧操作区 -->
      <div class="action-area">
        
        <!-- 1. 语音控制组 (非编辑状态且有内容时显示) -->
        <div class="speech-controls" v-if="!isEditing && content">
          <div class="speed-slider-box">
            <span class="speed-label">语速 {{ speechRate.toFixed(1) }}x</span>
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
          
          <div class="speech-btns">
            <el-tooltip content="朗读全文" placement="top" v-if="speechStatus === 'stopped'">
              <el-button link class="icon-btn primary" @click="handleSpeak()">
                <el-icon><Microphone /></el-icon> 朗读
              </el-button>
            </el-tooltip>
            <el-tooltip content="暂停" placement="top" v-if="speechStatus === 'playing'">
              <el-button link class="icon-btn warning" @click="handlePause">
                <el-icon><VideoPause /></el-icon> 暂停
              </el-button>
            </el-tooltip>
            <el-tooltip content="继续" placement="top" v-if="speechStatus === 'paused'">
              <el-button link class="icon-btn success" @click="handleResume">
                <el-icon><VideoPlay /></el-icon> 继续
              </el-button>
            </el-tooltip>
            <el-tooltip content="停止" placement="top" v-if="speechStatus !== 'stopped'">
              <el-button link class="icon-btn danger" @click="handleStop">
                <el-icon><SwitchButton /></el-icon> 停止
              </el-button>
            </el-tooltip>
          </div>
        </div>

        <!-- 分割线 -->
        <div class="divider-vertical" v-if="(!isEditing && content) && canEdit"></div>

        <!-- 2. 编辑按钮组 -->
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
       <!-- A. 编辑模式 (WangEditor) -->
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

      <!-- B. 预览模式 (HTML) -->
      <div 
        v-else 
        class="html-preview" 
        ref="previewRef" 
        @mouseup="handleTextSelection"
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
import { Microphone, VideoPause, VideoPlay, SwitchButton, Reading, Edit, Check } from '@element-plus/icons-vue';
import { uploadImage, updatePoint } from "../api/point";
import '@wangeditor/editor/dist/css/style.css'; // 引入 css

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
const speechStatus = ref<SpeechStatus>('stopped');
const speechRate = ref(1.0);
const synth = window.speechSynthesis;
let utterance: SpeechSynthesisUtterance | null = null;

onMounted(() => {
  const savedRate = localStorage.getItem('user-speech-rate');
  if (savedRate) speechRate.value = parseFloat(savedRate);
});

const stripHtml = (html: string) => {
  const tmp = document.createElement("DIV");
  tmp.innerHTML = html;
  return tmp.textContent || tmp.innerText || "";
};

const handleRateChange = (val: number) => {
  localStorage.setItem('user-speech-rate', val.toString()); 
  if (speechStatus.value === 'playing' || speechStatus.value === 'paused') {
    handleStop();
    setTimeout(() => handleSpeak(), 100); 
  }
};

const handleTextSelection = () => {
  if (isEditing.value || speechStatus.value === 'stopped') return;
  const selection = window.getSelection();
  if (!selection || selection.rangeCount === 0) return;
  const previewDom = previewRef.value;
  if (!previewDom || !previewDom.contains(selection.anchorNode)) return;
  const userRange = selection.getRangeAt(0);
  const rangeToEnd = document.createRange();
  rangeToEnd.selectNodeContents(previewDom); 
  rangeToEnd.setStart(userRange.startContainer, userRange.startOffset); 
  const textToRead = rangeToEnd.toString();
  if (textToRead && textToRead.trim().length > 1) {
    handleStop(); 
    setTimeout(() => {
      handleSpeak(textToRead); 
      ElMessage.success("已跳转至选定位置播放");
    }, 200);
  }
};

const handleSpeak = (specificText?: string) => {
  const textToRead = specificText || stripHtml(props.content);
  if (!textToRead.trim()) {
    if (!specificText) ElMessage.warning("没有可朗读的文本内容");
    return;
  }
  if (synth.speaking) synth.cancel();
  utterance = new SpeechSynthesisUtterance(textToRead);
  utterance.lang = 'zh-CN'; 
  utterance.rate = speechRate.value; 
  utterance.onend = () => { speechStatus.value = 'stopped'; };
  utterance.onerror = (e) => { if (e.error !== 'interrupted') speechStatus.value = 'stopped'; };
  synth.speak(utterance);
  speechStatus.value = 'playing';
};

const handlePause = () => { if (synth.speaking && !synth.paused) { synth.pause(); speechStatus.value = 'paused'; } };
const handleResume = () => { if (synth.paused) { synth.resume(); speechStatus.value = 'playing'; } };
const handleStop = () => { synth.cancel(); speechStatus.value = 'stopped'; };

// 监听内容变化
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

onBeforeUnmount(() => {
  handleStop();
  const editor = editorRef.value;
  if (editor == null) return;
  editor.destroy();
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
/* 整体容器 */
.content-column { 
  flex: 3; 
  display: flex; 
  flex-direction: column; 
  /* 移除右边框，用间距代替，更通透 */
  padding-right: 5px; 
  height: 100%; 
  background: transparent;
}

/* 顶部标题栏 */
.section-header { 
  display: flex; 
  justify-content: space-between; 
  align-items: center; 
  margin-bottom: 15px; 
  flex-shrink: 0; 
  border-bottom: 1px solid rgba(0,0,0,0.05);
  padding-bottom: 10px;
}
.section-title { 
  font-weight: bold; 
  color: #303133; 
  font-size: 18px; 
  display: flex; align-items: center;
}
.mr-1 { margin-right: 6px; }

/* 操作区域 */
.action-area { display: flex; align-items: center; gap: 15px; }

/* 语音控制 */
.speech-controls { display: flex; align-items: center; gap: 12px; }
.speed-slider-box { 
  display: flex; align-items: center; 
  padding: 0 12px; 
  background-color: rgba(245, 247, 250, 0.8); 
  border-radius: 20px; 
  height: 32px; 
}
.speed-label { font-size: 12px; color: #909399; white-space: nowrap; margin-right: 10px; }
.custom-slider { width: 80px; }

.speech-btns { display: flex; gap: 5px; }
.icon-btn { font-weight: bold; font-size: 13px; }
.icon-btn.primary { color: #764ba2; }
.icon-btn.primary:hover { color: #5b3a85; }
.icon-btn.warning { color: #e6a23c; }
.icon-btn.success { color: #67c23a; }
.icon-btn.danger { color: #f56c6c; }

.divider-vertical { width: 1px; height: 16px; background-color: #e4e7ed; }

/* 按钮样式 */
.gradient-btn {
  background: linear-gradient(90deg, #667eea, #764ba2);
  border: none;
  box-shadow: 0 2px 6px rgba(118, 75, 162, 0.3);
  padding: 8px 18px;
  font-weight: 600;
}
.gradient-btn:hover { opacity: 0.9; transform: translateY(-1px); }
.cancel-btn { color: #909399; }
.cancel-btn:hover { color: #764ba2; background-color: rgba(118, 75, 162, 0.1); }
.edit-actions { display: flex; gap: 8px; }

/* 内容区域 */
.content-box { flex: 1; display: flex; flex-direction: column; min-height: 0; }

/* 编辑器容器 */
.editor-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  border: 1px solid rgba(0,0,0,0.1);
  border-radius: 8px;
  overflow: hidden;
  background: rgba(255,255,255,0.6); /* 微透白 */
  z-index: 10;
}

/* WangEditor 样式穿透 */
:deep(.w-e-toolbar) { background-color: rgba(249, 250, 251, 0.9) !important; }
:deep(.w-e-text-container) { background-color: transparent !important; }
:deep(.w-e-bar-item button:hover) { color: #764ba2; }

/* 预览模式 */
.html-preview { 
  flex: 1; 
  padding: 5px; 
  line-height: 1.8; 
  color: #333; 
  font-size: 15px; 
  overflow-y: auto; 
  cursor: text; 
}

/* 预览内容样式优化 */
.markdown-body :deep(p) { margin-bottom: 12px; }
.markdown-body :deep(img) { max-width: 100%; border-radius: 6px; box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
.markdown-body :deep(blockquote) { border-left: 4px solid #d3adf7; background: rgba(249, 240, 255, 0.5); padding: 10px 15px; margin: 10px 0; color: #666; border-radius: 4px; }
.markdown-body :deep(code) { background-color: rgba(0,0,0,0.05); padding: 2px 5px; border-radius: 4px; font-family: monospace; color: #c7254e; }
/* 改为类似 WangEditor 的默认样式 */
.markdown-body :deep(pre) {
  background-color: #f1f1f1; /* 浅灰背景 */
  color: #333;
  padding: 10px;
  border-radius: 4px;
  border: 1px solid #ccc;
}

/* 空状态 */
.empty-tip {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  height: 100%; color: rgba(0,0,0,0.3); margin-top: 40px;
}
.empty-tip p { margin-top: 10px; font-size: 14px; }

/* 滚动条 */
.custom-scrollbar::-webkit-scrollbar { width: 6px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(0,0,0,0.1); border-radius: 3px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }



</style>
