<template>
  <div class="content-column">
    <div class="section-header">
      <div class="section-title">知识点详解</div>
      
      <!-- 右侧操作区 -->
      <div class="action-area">
        
        <!-- 1. 语音控制组 (只有在非编辑状态且有内容时显示) -->
        <div class="speech-controls" v-if="!isEditing && content">
          <div class="speed-slider-box">
            <span class="speed-label">语速 {{ speechRate.toFixed(1) }}x</span>
            <el-slider v-model="speechRate" :min="0.1" :max="3.0" :step="0.1" size="small" style="width: 80px; margin-left: 10px;" @change="handleRateChange" />
          </div>
          
          <el-divider direction="vertical" />

          <el-tooltip content="朗读全文" placement="top" v-if="speechStatus === 'stopped'">
            <el-button type="primary" link @click="handleSpeak()"><el-icon><Microphone /></el-icon> 朗读</el-button>
          </el-tooltip>
          <el-tooltip content="暂停" placement="top" v-if="speechStatus === 'playing'">
            <el-button type="warning" link @click="handlePause"><el-icon><VideoPause /></el-icon> 暂停</el-button>
          </el-tooltip>
          <el-tooltip content="继续" placement="top" v-if="speechStatus === 'paused'">
            <el-button type="success" link @click="handleResume"><el-icon><VideoPlay /></el-icon> 继续</el-button>
          </el-tooltip>
          <el-tooltip content="停止" placement="top" v-if="speechStatus !== 'stopped'">
            <el-button type="danger" link @click="handleStop"><el-icon><SwitchButton /></el-icon> 停止</el-button>
          </el-tooltip>
        </div>

        <!-- 分割线：只有当既有语音控件，又有编辑按钮时才显示 -->
        <el-divider direction="vertical" v-if="(!isEditing && content) && canEdit" />

        <!-- 2. ★★★ 编辑按钮组 (核心修复) ★★★ -->
        <div class="edit-controls" v-if="canEdit">
          <el-button type="primary" size="small" v-if="!isEditing" @click="startEdit">编辑</el-button>
          <div v-else>
            <el-button size="small" @click="cancelEdit">取消</el-button>
            <el-button type="success" size="small" @click="saveEdit">保存</el-button>
          </div>
        </div>

      </div>
    </div>

    <div class="content-box">
       <!-- 编辑模式 -->
      <div v-if="isEditing" class="editor-wrapper">
        <Toolbar
          style="border-bottom: 1px solid #ccc"
          :editor="editorRef"
          :defaultConfig="toolbarConfig"
          :mode="mode"
        />
        <!-- ★★★ 修复点：改为 flex: 1，让它自动占据工具栏剩下的空间 ★★★ -->
        <Editor
          style="flex: 1; overflow-y: hidden;"
          v-model="innerContent"
          :defaultConfig="editorConfig"
          :mode="mode"
          @onCreated="handleCreated"
        />
      </div>
      <!-- 预览模式 -->
      <div v-else class="html-preview" ref="previewRef" @mouseup="handleTextSelection">
        <div v-if="content" v-html="content"></div>
        <el-empty v-else description="暂无详细内容" :image-size="60"></el-empty>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, onBeforeUnmount, watch, onMounted } from "vue";
import { ElMessage } from "element-plus";
import { Editor, Toolbar } from "@wangeditor/editor-for-vue";
import { Microphone, VideoPause, VideoPlay, SwitchButton } from '@element-plus/icons-vue';
import { uploadImage, updatePoint } from "../api/point";

// ★★★ 修复点：使用运行时 Props 定义，确保默认值 ★★★
const props = defineProps({
  pointId: { type: Number, required: true },
  content: { type: String, default: '' },
  canEdit: { type: Boolean, default: false } // 默认为 false，防止 undefined 报错
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

const toolbarConfig = {};
const editorConfig = {
  placeholder: "请输入内容...",
  MENU_CONF: {
    uploadImage: {
      async customUpload(file: File, insertFn: any) {
        try {
          const res = await uploadImage(file);
          if (res.data.code === 200) {
            const url = `http://localhost:8080${res.data.data.path}`;
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
.content-column { flex: 3; display: flex; flex-direction: column; border-right: 1px solid #eee; padding-right: 20px; height: 100%; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; flex-shrink: 0; }
.section-title { font-weight: bold; color: #333; font-size: 20px; }
.action-area { display: flex; align-items: center; gap: 15px; }
.speech-controls { display: flex; align-items: center; gap: 10px; }
.speed-slider-box { display: flex; align-items: center; padding: 0 10px; background-color: #f5f7fa; border-radius: 16px; height: 32px; }
.speed-label { font-size: 12px; color: #606266; white-space: nowrap; width: 60px; }
.content-box { flex: 1; display: flex; flex-direction: column; min-height: 0; }
.editor-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  border: 1px solid #ccc;
  z-index: 100;
  
  /* ★★★ 核心修复点：加上这行，防止编辑器无限撑高 ★★★ */
  overflow: hidden; 
}
.html-preview { flex: 1; padding: 10px; line-height: 1.8; color: #333; font-size: 15px; overflow-y: auto; border: 1px solid #dcdfe6; border-radius: 4px; cursor: text; }
.html-preview :deep(img) { max-width: 100%; }
</style>