<template>
  <div class="content-column">
    <div class="section-header">
      <div class="section-title">知识点详解</div>
      
      <!-- 右侧操作区 -->
      <div class="action-area">
        <!-- 语音控制组 -->
        <div class="speech-controls" v-if="!isEditing && content">
          
          <!-- 语速调节滑块 -->
          <div class="speed-slider-box">
            <span class="speed-label">语速 {{ speechRate.toFixed(1) }}x</span>
            <el-slider 
              v-model="speechRate" 
              :min="0.1" 
              :max="3.0" 
              :step="0.1" 
              size="small"
              style="width: 100px; margin-left: 10px;"
              @change="handleRateChange"
            />
          </div>

          <el-divider direction="vertical" />

          <el-tooltip content="朗读全文 (或选中文字从该处开始)" placement="top" v-if="speechStatus === 'stopped'">
            <el-button type="primary" link @click="handleSpeak()">
              <el-icon><Microphone /></el-icon> 朗读
            </el-button>
          </el-tooltip>

          <el-tooltip content="暂停" placement="top" v-if="speechStatus === 'playing'">
            <el-button type="warning" link @click="handlePause">
              <el-icon><VideoPause /></el-icon> 暂停
            </el-button>
          </el-tooltip>

          <el-tooltip content="继续" placement="top" v-if="speechStatus === 'paused'">
            <el-button type="success" link @click="handleResume">
              <el-icon><VideoPlay /></el-icon> 继续
            </el-button>
          </el-tooltip>

          <el-tooltip content="停止" placement="top" v-if="speechStatus !== 'stopped'">
            <el-button type="danger" link @click="handleStop">
              <el-icon><SwitchButton /></el-icon> 停止
            </el-button>
          </el-tooltip>
          
          <el-divider direction="vertical" />
        </div>

        <!-- 编辑按钮组 -->
        <el-button
          type="primary"
          size="small"
          v-if="!isEditing"
          @click="startEdit"
        >编辑</el-button>
        <div v-else>
          <el-button size="small" @click="cancelEdit">取消</el-button>
          <el-button type="success" size="small" @click="saveEdit">保存</el-button>
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
        <Editor
          style="height: 100%; overflow-y: hidden"
          v-model="innerContent"
          :defaultConfig="editorConfig"
          :mode="mode"
          @onCreated="handleCreated"
        />
      </div>
      <!-- 预览模式 -->
      <!-- 绑定 mouseup 事件来捕获选区 -->
      <div 
        v-else 
        class="html-preview" 
        ref="previewRef"
        @mouseup="handleTextSelection"
      >
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

const props = defineProps<{
  pointId: number;
  content: string;
}>();

const emit = defineEmits(["update"]);

// --- 编辑器相关 ---
const editorRef = shallowRef();
const mode = "default";
const isEditing = ref(false);
const innerContent = ref("");
const previewRef = ref<HTMLElement | null>(null); 

// --- 语音朗读相关逻辑 ---
type SpeechStatus = 'stopped' | 'playing' | 'paused';
const speechStatus = ref<SpeechStatus>('stopped');
const speechRate = ref(1.0);
const synth = window.speechSynthesis;
let utterance: SpeechSynthesisUtterance | null = null;

onMounted(() => {
  const savedRate = localStorage.getItem('user-speech-rate');
  if (savedRate) {
    speechRate.value = parseFloat(savedRate);
  }
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

// ------------------------------------------------------
//  修改后的核心选段逻辑
// ------------------------------------------------------
const handleTextSelection = () => {
  if (isEditing.value) return;

  const selection = window.getSelection();
  if (!selection || selection.rangeCount === 0) return;
  
  const previewDom = previewRef.value;
  if (!previewDom) return;

  // 防止点击了编辑器以外的地方触发
  if (!previewDom.contains(selection.anchorNode)) return;

  // 1. 拿到用户选中的那个范围（即你图中蓝色的部分）
  const userRange = selection.getRangeAt(0);

  // 2. 创建一个"从这里到文档结束"的新范围
  const rangeToEnd = document.createRange();
  rangeToEnd.selectNodeContents(previewDom); // 先选中所有
  rangeToEnd.setStart(userRange.startContainer, userRange.startOffset); // 把起点移到你选中的开头

  // 3. 提取文字
  const textToRead = rangeToEnd.toString();

  // 4. 如果有字，就开始读
  if (textToRead && textToRead.trim().length > 1) {
    handleStop(); // 切歌
    setTimeout(() => {
      handleSpeak(textToRead); // 播放剩下的所有内容
      ElMessage.success("已跳转至选定位置播放");
    }, 200);
  }
};

const handleSpeak = (specificText?: string) => {
  // 如果传入了特定文本（比如从中间开始的），就读特定的；否则读全文
  const textToRead = specificText || stripHtml(props.content);
  
  if (!textToRead.trim()) {
    if (!specificText) ElMessage.warning("没有可朗读的文本内容");
    return;
  }

  if (synth.speaking) synth.cancel();

  utterance = new SpeechSynthesisUtterance(textToRead);
  utterance.lang = 'zh-CN'; 
  utterance.rate = speechRate.value; 
  utterance.pitch = 1; 

  utterance.onend = () => {
    speechStatus.value = 'stopped';
  };
  
  utterance.onerror = (e) => {
    if (e.error !== 'interrupted' && e.error !== 'canceled') {
       speechStatus.value = 'stopped';
    }
  };

  synth.speak(utterance);
  speechStatus.value = 'playing';
};

const handlePause = () => {
  if (synth.speaking && !synth.paused) {
    synth.pause();
    speechStatus.value = 'paused';
  }
};

const handleResume = () => {
  if (synth.paused) {
    synth.resume();
    speechStatus.value = 'playing';
  }
};

const handleStop = () => {
  synth.cancel(); 
  speechStatus.value = 'stopped';
};

// --- 监听与生命周期 ---
watch(() => props.content, (newVal) => {
  if (!isEditing.value) innerContent.value = newVal || "";
  handleStop();
}, { immediate: true });

watch(isEditing, (newVal) => {
  if (newVal) handleStop();
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
};

const startEdit = () => {
  innerContent.value = props.content || "";
  isEditing.value = true;
};
const cancelEdit = () => {
  isEditing.value = false;
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
.content-column {
  flex: 3;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #eee;
  padding-right: 20px;
  height: 100%;
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  flex-shrink: 0;
}
.section-title {
  font-weight: bold;
  color: #333;
  font-size: 20px;
}

.action-area {
  display: flex;
  align-items: center;
  gap: 15px;
}
.speech-controls {
  display: flex;
  align-items: center;
  gap: 10px;
}

.speed-slider-box {
  display: flex;
  align-items: center;
  padding: 0 10px;
  background-color: #f5f7fa;
  border-radius: 16px;
  height: 32px;
}
.speed-label {
  font-size: 12px;
  color: #606266;
  white-space: nowrap;
  width: 60px; 
}

.content-box {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.editor-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  border: 1px solid #ccc;
  z-index: 100;
}
.html-preview {
  flex: 1;
  padding: 10px;
  line-height: 1.8;
  color: #333;
  font-size: 15px;
  overflow-y: auto;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  cursor: text; 
}
.html-preview :deep(img) {
  max-width: 100%;
}
</style>