<template>
  <el-dialog
    v-model="visible"
    title="AI 模拟面试"
    width="1000px" 
    class="ai-interviewer-dialog"
    @open="onDialogOpen"
    @close="onDialogClose"
    :modal="true"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    align-center
  >
    <div class="ai-interviewer-container">
      <!-- 头部信息 -->
      <div class="interviewer-header">
        <div class="point-info">
          <el-icon class="info-icon"><Service /></el-icon>
          <span class="point-title">{{ pointTitle }}</span>
        </div>
        
        <div class="header-right">
          <!-- 极简控制栏 -->
          <div class="simple-player-controls">
            <!-- 自动朗读开关 -->
            <el-checkbox v-model="autoRead" size="small" class="auto-read-checkbox">
              自动朗读
            </el-checkbox>
            
            <div class="divider-small" v-if="speechStatus === 'playing'"></div>

            <!-- 停止按钮 (只在播放时显示) -->
            <el-tooltip content="停止朗读" placement="top">
              <el-button 
                v-if="speechStatus === 'playing'"
                link 
                class="stop-btn is-playing"
                @click="handleStopSpeech"
              >
                <div class="playing-indicator">
                  <span></span><span></span><span></span>
                </div>
              </el-button>
            </el-tooltip>
          </div>

          <!-- 倒计时 -->
          <el-tag :type="quotaType" effect="dark" class="quota-tag">
            剩余: {{ formatTime(remainingQuota) }}
          </el-tag>
        </div>
      </div>

      <!-- 聊天内容区域 -->
      <div ref="chatContainerRef" class="chat-container">
        <!-- 空状态 -->
        <div v-if="messages.length === 0" class="empty-chat">
          <el-icon :size="40" color="#909399"><ChatDotRound /></el-icon>
          <p>等待 AI 面试官连接...</p>
        </div>
        
        <!-- 消息列表 -->
        <div 
          v-for="(msg, index) in messages" 
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
            <!-- Markdown 渲染 -->
            <div 
              class="message-bubble markdown-body" 
              v-html="renderMessage(msg.content)"
            ></div>
            <!-- 朗读按钮 -->
            <div v-if="msg.role === 'assistant'" class="msg-actions">
              <el-button link size="small" @click="speak(msg.content)">
                <el-icon><Microphone /></el-icon> 朗读
              </el-button>
               <!-- 2. 复制按钮 (新增在这里！！！) -->
              <el-button link size="small" @click="copyContent(msg.content)">
                <el-icon><CopyDocument /></el-icon> 复制
              </el-button>
            </div>
          </div>
        </div>
      </div>

      <!-- 底部输入区域 -->
      <div class="input-container">
        <el-input
          v-model="userInput"
          type="textarea"
          :rows="3"
          placeholder="请粘贴你的回答，或直接输入...(Tips: 你可以直接用: 豆包,ChatGPT等,工具让他们把你的语音转变成文字)"
          :disabled="isInputDisabled"
          resize="none"
          @keydown.enter="handleEnterKey"
        />
        
        <div class="input-actions">
          <!-- 发送按钮优化版 -->
          <el-button 
            @click="sendMessage" 
            :loading="isLoading" 
            :disabled="isInputDisabled || (!userInput.trim() && !isLoading)"
            class="gradient-btn"
          >
            <!-- 只有不加载时才显示小飞机图标，加载时 Element 会自动显示转圈 -->
            <el-icon v-if="!isLoading" class="mr-1"><Promotion /></el-icon>
            {{ isLoading ? '思考中...' : '发送' }}
          </el-button>

          <el-button @click="resetInterview" :disabled="isInputDisabled">
            <el-icon class="mr-1"><RefreshRight /></el-icon>
            重新开始
          </el-button>
        </div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed, nextTick, onUnmounted } from 'vue';
import { ElMessage } from 'element-plus';
import { ChatDotRound, Service, User, RefreshRight, Promotion, Microphone } from '@element-plus/icons-vue';
import MarkdownIt from 'markdown-it';

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  pointId: { type: Number, default: 0 },
  pointTitle: { type: String, default: '' },
  pointContent: { type: String, default: '' }
});
const emit = defineEmits(['update:modelValue']);

// --- 基础状态 ---
const visible = ref(false);
const messages = ref<any[]>([]);
const userInput = ref('');
const isLoading = ref(false);
const remainingQuota = ref(0);
const chatContainerRef = ref<HTMLElement | null>(null);

const md = new MarkdownIt({ html: false, linkify: true, breaks: true });
const aiWs = ref<WebSocket | null>(null);
const isConnected = ref(false);

// --- 语音状态 ---
const synth = window.speechSynthesis;
const speechStatus = ref<'stopped' | 'playing'>('stopped');
const autoRead = ref(true); 

// ==========================================
// 1. 弹窗打开/关闭时的音频管理 (核心修复)
// ==========================================
const onDialogOpen = () => {
  // 🔥 核心：一打开弹窗，立马把外面（父组件）正在读的声音掐断！
  synth.cancel();
  
  if (!isConnected.value) connectAIWebSocket();
};

const onDialogClose = () => {
  // 🔥 核心：一关闭弹窗，立马把自己正在读的 AI 声音掐断！
  handleStopSpeech();
  
  if (aiWs.value) {
    aiWs.value.close();
    aiWs.value = null;
  }
  
  isConnected.value = false;
  isLoading.value = false;
  messages.value = [];
};

// ==========================================
// 2. 朗读逻辑 (读缓存 + 极简)
// ==========================================
const speak = (rawText: string) => {
  const text = stripMarkdown(rawText);
  if (!text.trim()) return;

  // 先停掉当前的
  synth.cancel();

  setTimeout(() => {
    const newUtterance = new SpeechSynthesisUtterance(text);
    (window as any).currentUtterance = newUtterance; 

    // 读取本地缓存 (如果没有就用默认)
    const savedRate = localStorage.getItem('user-speech-rate');
    newUtterance.rate = savedRate ? parseFloat(savedRate) : 1.0;

    const savedVoiceURI = localStorage.getItem('user-speech-voice');
    if (savedVoiceURI) {
      const voices = synth.getVoices();
      const targetVoice = voices.find(v => v.voiceURI === savedVoiceURI);
      if (targetVoice) newUtterance.voice = targetVoice;
    }
    
    newUtterance.lang = 'zh-CN'; 
    newUtterance.onend = () => { speechStatus.value = 'stopped'; };
    newUtterance.onerror = () => { speechStatus.value = 'stopped'; };
    
    synth.speak(newUtterance);
    speechStatus.value = 'playing';
  }, 50);
};

// 2. 在 speak 函数附近添加这个复制函数
const copyContent = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text);
    ElMessage.success('内容已复制');
  } catch (err) {
    ElMessage.error('复制失败');
  }
};

const handleStopSpeech = () => {
  synth.cancel(); // 彻底停止所有声音
  speechStatus.value = 'stopped';
};

const stripMarkdown = (mdText: string) => {
  if (!mdText) return "";
  let text = mdText
    .replace(/\*\*(.*?)\*\*/g, '$1') 
    .replace(/__(.*?)__/g, '$1')
    .replace(/#+\s/g, '') 
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1') 
    .replace(/`{1,3}(.*?)`{1,3}/g, '$1') 
    .replace(/\n/g, '，'); 
  return text;
};

// ==========================================
// 3. WebSocket & 交互逻辑
// ==========================================
const isInputDisabled = computed(() => !isConnected.value || isLoading.value);
const quotaType = computed(() => remainingQuota.value < 60 ? 'danger' : 'success');

watch(() => props.modelValue, (val) => {
  visible.value = val;
  if (!val) onDialogClose();
});
watch(visible, (val) => emit('update:modelValue', val));

const connectAIWebSocket = () => {
  const token = localStorage.getItem('auth_token');
  if (!token) return;

  const title = props.pointTitle ? encodeURIComponent(props.pointTitle) : '';
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = process.env.NODE_ENV === 'development' ? 'localhost:19527' : window.location.host;
  const url = `${protocol}//${host}/api/v1/ws/ai-interview?token=${token}&point_title=${title}`;

  const ws = new WebSocket(url);
  aiWs.value = ws;

  ws.onopen = () => { isConnected.value = true; };
  ws.onmessage = (e) => {
    try {
      const msg = JSON.parse(e.data);
      handleAIMessage(msg);
    } catch (err) {}
  };
  ws.onerror = () => {
    if (!isConnected.value) ElMessage.error('连接失败：请检查登录或网络');
  };
  ws.onclose = () => {
    isConnected.value = false;
    handleStopSpeech();
  };
};

const handleAIMessage = (msg: any) => {
  switch (msg.type) {
    case 'init':
      remainingQuota.value = msg.content.quota || 0;
      break;
    case 'chat':
      isLoading.value = false;
      const content = msg.content;
      messages.value.push({ role: 'assistant', content: content });
      scrollToBottom();
      if (autoRead.value) speak(content);
      break;
    case 'quota_error':
    case 'error':
      ElMessage.error(msg.content.message || '发生错误');
      isLoading.value = false;
      break;
  }
};

const sendMessage = () => {
  const content = userInput.value.trim();
  if (!content) return;
  
  handleStopSpeech(); // 用户说话时，让 AI 闭嘴
  messages.value.push({ role: 'user', content });
  userInput.value = '';
  scrollToBottom();
  isLoading.value = true;

  const payload = { type: 'chat', content: { topic: props.pointTitle, content: content } };
  if (aiWs.value && aiWs.value.readyState === WebSocket.OPEN) {
    aiWs.value.send(JSON.stringify(payload));
  } else {
    ElMessage.error('连接已断开');
    isLoading.value = false;
  }
};

const resetInterview = () => {
  handleStopSpeech();
  messages.value = [];
  userInput.value = '';
  if (aiWs.value) aiWs.value.close();
  setTimeout(() => connectAIWebSocket(), 200);
};

const handleEnterKey = (e: KeyboardEvent) => {
  // 仅在非加载状态且有内容时，按 Enter 发送
  if (!e.ctrlKey && !e.shiftKey && !isLoading.value && userInput.value.trim()) {
    e.preventDefault();
    sendMessage();
  }
};

const scrollToBottom = () => {
  nextTick(() => {
    if (chatContainerRef.value) {
      chatContainerRef.value.scrollTop = chatContainerRef.value.scrollHeight;
    }
  });
};

const formatTime = (s: number) => {
  const m = Math.floor(s / 60);
  const sec = s % 60;
  return `${m}分${sec}秒`;
};

const renderMessage = (content: string) => md.render(content);

// 销毁组件时确保停止
onUnmounted(() => {
  onDialogClose();
});
</script>

<style scoped>
/* 弹窗容器 */
:global(.ai-interviewer-dialog) {
  position: absolute !important;
  top: 50% !important;
  left: 50% !important;
  transform: translate(-50%, -50%) !important;
  margin: 0 !important;
  height: 80vh !important;
  max-height: 90vh !important;
  display: flex !important;
  flex-direction: column !important;
  overflow: hidden !important; 
  box-shadow: 0 12px 32px 4px rgba(0, 0, 0, 0.12), 0 8px 20px rgba(0, 0, 0, 0.08);
}

:global(.ai-interviewer-dialog .el-dialog__body) {
  flex: 1 !important;
  height: 0 !important;
  min-height: 0 !important;
  padding: 0 !important;
  margin: 0 !important;
}

.ai-interviewer-container {
  height: 100% !important;
  width: 100%;
  display: flex;
  flex-direction: column;
  background-color: #fff;
}

/* 头部 */
.interviewer-header {
  flex-shrink: 0;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  border-bottom: 1px solid #ebeef5;
  background: #fff;
}

/* --- 极简播放控制条 --- */
.simple-player-controls {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-right: 16px;
  padding-right: 16px;
  border-right: 1px solid #ebeef5;
  height: 24px;
}
.auto-read-checkbox { height: 24px; }
:deep(.auto-read-checkbox .el-checkbox__label) { font-size: 13px; color: #606266; }
.stop-btn { padding: 0 8px; color: #f56c6c; display: flex; align-items: center; }

/* 播放动画 */
.playing-indicator span {
  display: inline-block;
  width: 3px;
  height: 12px;
  background-color: #f56c6c;
  margin: 0 1px;
  border-radius: 1px;
  animation: wave 1s infinite ease-in-out;
}
.playing-indicator span:nth-child(2) { animation-delay: 0.2s; }
.playing-indicator span:nth-child(3) { animation-delay: 0.4s; }
@keyframes wave {
  0%, 100% { transform: scaleY(0.6); }
  50% { transform: scaleY(1.2); }
}

/* 聊天内容 */
.chat-container {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  background: #f5f7fa;
  padding: 20px;
}
.chat-container::-webkit-scrollbar { width: 6px; }
.chat-container::-webkit-scrollbar-thumb { background: #c0c4cc; border-radius: 4px; }
.chat-container::-webkit-scrollbar-track { background: transparent; }

/* 消息气泡 */
.message-item { display: flex; gap: 12px; margin-bottom: 20px; }
.message-item.user { flex-direction: row-reverse; }
.message-avatar { width: 40px; height: 40px; border-radius: 50%; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.assistant-avatar { background: linear-gradient(135deg, #667eea, #764ba2); }
.user-avatar { background: linear-gradient(135deg, #11998e, #38ef7d); }
.message-content { max-width: 80%; display: flex; flex-direction: column; }
.message-bubble { padding: 12px 16px; border-radius: 12px; font-size: 14px; line-height: 1.6; }
.message-item.assistant .message-bubble { background: #fff; border: 1px solid #e4e7ed; border-top-left-radius: 2px; }
.message-item.user .message-bubble { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #fff; border-top-right-radius: 2px; }

/* 消息下方的朗读按钮 */
.msg-actions { margin-top: 4px; display: flex; gap: 8px; opacity: 0; transition: opacity 0.2s; }
.message-item.assistant:hover .msg-actions { opacity: 1; }

/* 底部输入 */
.input-container {
  flex-shrink: 0;
  background: #fff;
  border-top: 1px solid #ebeef5;
  padding: 16px 24px 24px;
  z-index: 10;
}
.input-actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 12px; }
.gradient-btn { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); border: none; color: #fff; }

.point-info { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 0; }
.info-icon { font-size: 18px; color: #764ba2; }
.point-title { font-size: 16px; font-weight: 600; color: #303133; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.header-right { display: flex; align-items: center; gap: 10px; }
.quota-tag { font-weight: 500; }
.empty-chat { height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #909399; }
</style>
