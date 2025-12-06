<template>
  <div class="content-column">
    <div class="section-header">
      <div class="left-group">
        <div class="section-title">
          <el-icon class="mr-1"><Reading /></el-icon> 知识点详解
        </div>

        <!-- 【固定播放控制条】 -->
        <div class="player-module" v-if="!isEditing && content">
          
          <!-- 1. 语音选择下拉框 (逻辑不变，仅增加图标和样式优化) -->
          <el-select 
            v-model="selectedVoiceURI" 
            placeholder="选择语音" 
            size="small" 
            class="voice-select"
            @change="handleVoiceChange"
            :teleported="false" 
          >
            <!-- 增加一个前缀图标，美观度+1 -->
            <template #prefix>
              <el-icon class="select-icon"><Headset /></el-icon>
            </template>
            
            <el-option
              v-for="voice in voiceList"
              :key="voice.voiceURI"
              :label="voice.name" 
              :value="voice.voiceURI"
            >
              <!-- 下拉列表里的样式保持原样 -->
              <span style="float: left">{{ voice.name }}</span>
              <span style="float: right; color: #8492a6; font-size: 12px; margin-left: 10px;">{{ voice.lang }}</span>
            </el-option>
          </el-select>

          <div class="divider-small"></div>

          <!-- 2. 语速滑块 -->
          <div class="speed-box">
            <span class="speed-label">{{ speechRate.toFixed(1) }}x</span>
            <el-slider 
              v-model="speechRate" 
              :min="0.5" 
              :max="2.0" 
              :step="0.1" 
              size="small" 
              class="custom-slider"
              @change="handleRateChange" 
            />
          </div>

          <div class="divider-small"></div>

          <!-- 3. 暂停/继续 -->
          <el-tooltip :content="speechStatus === 'paused' ? '继续' : '暂停'" placement="top">
            <el-button 
              link 
              class="control-btn"
              :class="speechStatus === 'playing' ? 'warning' : 'success'"
              :disabled="speechStatus === 'stopped'"
              @click="togglePauseResume"
            >
              <el-icon size="18">
                <VideoPlay v-if="speechStatus === 'paused'" />
                <VideoPause v-else />
              </el-icon>
            </el-button>
          </el-tooltip>

          <!-- 4. 停止 -->
          <el-tooltip content="停止" placement="top">
            <el-button 
              link 
              class="control-btn danger"
              :disabled="speechStatus === 'stopped'"
              @click="handleStop"
            >
              <el-icon size="18"><SwitchButton /></el-icon>
            </el-button>
          </el-tooltip>
        </div>
      </div>
      
      <!-- 右侧操作区 -->
      <div class="action-area">
        <div class="trigger-group" v-if="!isEditing && content">
          <!-- 朗读选中 -->
          <el-tooltip :content="selectedText ? '朗读选中的文字' : '请先在下方选择文字'" placement="top">
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

          <!-- 全文朗读 -->
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

        <div class="divider-vertical" v-if="(!isEditing && content) && canEdit"></div>

        <div class="edit-controls" v-if="canEdit">
          <el-button v-if="!isEditing" type="primary" size="small" icon="Edit" class="gradient-btn" @click="startEdit">编辑内容</el-button>
          <div v-else class="edit-actions">
            <el-button size="small" @click="cancelEdit" class="cancel-btn">取消</el-button>
            <el-button type="primary" size="small" icon="Check" class="gradient-btn" @click="saveEdit">保存</el-button>
          </div>
        </div>
      </div>
    </div>

    <div class="content-box custom-scrollbar">
      <div v-if="isEditing" class="editor-wrapper">
        <Toolbar style="border-bottom: 1px solid rgba(0,0,0,0.05)" :editor="editorRef" :defaultConfig="toolbarConfig" :mode="mode" />
        <Editor style="flex: 1; overflow-y: hidden;" v-model="innerContent" :defaultConfig="editorConfig" :mode="mode" @onCreated="handleCreated" />
      </div>

      <div v-else class="html-preview" ref="previewRef" @mouseup="captureSelection" @touchend="captureSelection">
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
import { Microphone, VideoPause, VideoPlay, SwitchButton, Reading, Edit, Check, ChatLineSquare, Headset } from '@element-plus/icons-vue';
import { uploadImage, updatePoint } from "../api/point";
import '@wangeditor/editor/dist/css/style.css'; 

// ------------------------------------------------------------------
// 逻辑完全保持不变
// ------------------------------------------------------------------

const props = defineProps({
  pointId: { type: Number, required: true },
  content: { type: String, default: '' },
  canEdit: { type: Boolean, default: false }
});

const emit = defineEmits(["update"]);

const editorRef = shallowRef();
const mode = "default";
const isEditing = ref(false);
const innerContent = ref("");
const previewRef = ref<HTMLElement | null>(null); 

type SpeechStatus = 'stopped' | 'playing' | 'paused';
type ReadingMode = 'full' | 'selected' | 'none'; 

const speechStatus = ref<SpeechStatus>('stopped');
const readingMode = ref<ReadingMode>('none'); 
const speechRate = ref(1.0);
const selectedText = ref(""); 
const synth = window.speechSynthesis;
let utterance: SpeechSynthesisUtterance | null = null;

const voiceList = ref<SpeechSynthesisVoice[]>([]); 
const selectedVoiceURI = ref(""); 

const currentCharIndex = ref(0); 
const currentFullText = ref(""); 

onMounted(() => {
  const savedRate = localStorage.getItem('user-speech-rate');
  if (savedRate) speechRate.value = parseFloat(savedRate);
  
  initVoices();
  if (speechSynthesis.onvoiceschanged !== undefined) {
    speechSynthesis.onvoiceschanged = initVoices;
  }

  document.addEventListener('click', handleGlobalClick);
});

const initVoices = () => {
  const allVoices = synth.getVoices();
  voiceList.value = allVoices.sort((a, b) => {
      const aZh = a.lang.includes('zh');
      const bZh = b.lang.includes('zh');
      if (aZh && !bZh) return -1;
      if (!aZh && bZh) return 1;
      return 0;
  });

  const savedVoice = localStorage.getItem('user-speech-voice');
  if (savedVoice && voiceList.value.find(v => v.voiceURI === savedVoice)) {
    selectedVoiceURI.value = savedVoice;
  } else {
    const zhVoice = voiceList.value.find(v => v.lang.includes('zh-CN'));
    if (zhVoice) selectedVoiceURI.value = zhVoice.voiceURI;
    else if (voiceList.value.length > 0) selectedVoiceURI.value = voiceList.value[0].voiceURI;
  }
};

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

const togglePauseResume = () => {
  if (speechStatus.value === 'playing') {
    synth.pause();
    speechStatus.value = 'paused';
  } else if (speechStatus.value === 'paused') {
    synth.resume();
    speechStatus.value = 'playing';
  }
};

const handleVoiceChange = (val: string) => {
  localStorage.setItem('user-speech-voice', val);
  restartSpeechIfPlaying();
};

const handleRateChange = (val: number) => {
  localStorage.setItem('user-speech-rate', val.toString()); 
  restartSpeechIfPlaying();
};

const restartSpeechIfPlaying = () => {
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

// 核心播放方法 (修复版)
const speak = (text: string, mode: ReadingMode) => {
  if (!text.trim()) {
    ElMessage.warning("没有可朗读的文本");
    return;
  }

  // 1. 先彻底停止当前正在读的
  synth.cancel();

  // 更新 UI 状态
  readingMode.value = mode;
  currentFullText.value = text; 
  currentCharIndex.value = 0; 

  // 2. 【关键修复】使用 setTimeout 延时启动
  // 解决 Chrome/Edge 中立即 speak 导致的 interrupted 错误
  setTimeout(() => {
    const newUtterance = new SpeechSynthesisUtterance(text);
    
    // ------------------------------------------------------
    // 【核心修复】防止浏览器垃圾回收机制(GC)杀掉朗读进程
    // 必须把它挂载到全局变量上，只要它还活着，浏览器就不敢杀
    // ------------------------------------------------------
    (window as any).currentUtterance = newUtterance; 

    // 3. 强制从浏览器最新列表中匹配音色
    if (selectedVoiceURI.value) {
      const voices = synth.getVoices();
      // 优先匹配 ID
      let targetVoice = voices.find(v => v.voiceURI === selectedVoiceURI.value);
      
      // 兜底：匹配名字 (防止 ID 变动)
      if (!targetVoice && voiceList.value.length > 0) {
         const cached = voiceList.value.find(v => v.voiceURI === selectedVoiceURI.value);
         if (cached) targetVoice = voices.find(v => v.name === cached.name);
      }

      if (targetVoice) {
        newUtterance.voice = targetVoice;
        console.log("正在使用音色:", targetVoice.name);
      }
    }
    
    newUtterance.lang = 'zh-CN'; 
    newUtterance.rate = speechRate.value; 
    
    // 监听进度
    newUtterance.onboundary = (event) => {
      if (event.name === 'word' || event.name === 'sentence') {
        currentCharIndex.value = event.charIndex;
      }
    };

    // 监听结束
    newUtterance.onend = () => { 
      speechStatus.value = 'stopped'; 
      readingMode.value = 'none'; 
      currentCharIndex.value = 0;
    };
    
    // 监听错误
    newUtterance.onerror = (e) => { 
      // 忽略 interrupted 错误，这通常是我们手动切换导致的，不是真的错
      if (e.error === 'interrupted' || e.error === 'canceled') {
        return; 
      }
      console.error("朗读出错详情:", e);
      speechStatus.value = 'stopped'; 
      readingMode.value = 'none';
    };
    
    // 4. 开始播放
    synth.speak(newUtterance);
    speechStatus.value = 'playing';
    
    // 更新外部引用
    utterance = newUtterance;

  }, 50); // 延时 50ms 足够解决冲突
};


const handleStop = () => { 
  synth.cancel(); 
  speechStatus.value = 'stopped'; 
  readingMode.value = 'none'; 
  currentCharIndex.value = 0;
  currentFullText.value = "";
};

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

.section-header { 
  display: flex; 
  justify-content: space-between; 
  align-items: center; 
  margin-bottom: 15px; 
  flex-shrink: 0; 
  border-bottom: 1px solid rgba(0,0,0,0.05); 
  padding-bottom: 10px; 
}

.left-group { display: flex; align-items: center; gap: 20px; }
.section-title { font-weight: bold; color: #303133; font-size: 18px; display: flex; align-items: center; }
.mr-1 { margin-right: 6px; }

/* 播放控制模块 */
.player-module {
  display: flex;
  align-items: center;
  
  /* 【修改点】：背景色改为淡紫色，带一点透明度 */
  background: rgba(242, 235, 255, 0.9); 
  border: 1px solid rgba(118, 75, 162, 0.1); /* 可选：加一个极淡的紫色边框增加精致感 */

  border-radius: 20px;
  padding: 4px 12px;
  height: 32px;
  gap: 8px;
  transition: all 0.3s;
  box-shadow: 0 1px 3px rgba(118, 75, 162, 0.1); /* 阴影也带一点点紫 */
}

/* 悬停时稍微加深一点点 */
.player-module:hover { 
  background: rgba(233, 222, 255, 0.95); 
  box-shadow: 0 2px 8px rgba(118, 75, 162, 0.15);
}

/* 
   ----------------------------------------------------------------
   【核心样式优化】：让 el-select 变透明、无边框、融合进胶囊 
   ----------------------------------------------------------------
*/
.voice-select {
  width: 50px; /* 宽度适中 */
}

/* 去掉输入框的阴影（边框） */
:deep(.voice-select .el-input__wrapper) {
  background-color: transparent !important;
  box-shadow: none !important;
  padding: 0 0 0 0px; /* 紧凑一点 */
}

/* 聚焦时也不要有边框 */
:deep(.voice-select .el-input__wrapper.is-focus) {
  box-shadow: none !important;
}

/* 调整文字样式 */
:deep(.voice-select .el-input__inner) {
  font-size: 12px;
  color: #606266;
  font-weight: 500;
  height: 32px;
}

/* 图标颜色 */
.select-icon {
  color: #606266;
  margin-right: 0px;
  font-size: 14px;
}

/* ---------------------------------------------------------------- */

.speed-box { display: flex; align-items: center; }
.speed-label { font-size: 12px; color: #606266; margin-right: 8px; width: 24px; text-align: right; }
.custom-slider { width: 60px; }

.divider-small { width: 1px; height: 14px; background-color: #dcdfe6; margin: 0 4px; }

.control-btn { padding: 0; height: auto; }
.control-btn.warning { color: #e6a23c; }
.control-btn.success { color: #67c23a; }
.control-btn.danger { color: #f56c6c; }
.control-btn.is-disabled { color: #c0c4cc; }

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