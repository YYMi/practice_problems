<template>
  <div class="content-column">
    <div class="section-header">
      <div class="left-group">
        <div class="section-title">
          <el-icon class="mr-1"><Reading /></el-icon> 知识点详解
          <!-- 绑定数量标签（和标题融合） -->
          <el-popover
            v-if="bindings && bindings.length > 0"
            v-model:visible="bindingsPopoverVisible"
            placement="bottom"
            :width="320"
            trigger="click"
          >
            <template #reference>
              <span class="binding-count-inline">(<el-icon><Link /></el-icon>{{ bindings.length }})</span>
            </template>
            <div class="all-bindings-popover">
              <div class="all-bindings-header">全部绑定关系 ({{ bindings.length }})</div>
              <div class="all-bindings-list custom-scrollbar">
                <div 
                  v-for="b in (bindings as any[])" 
                  :key="b.id" 
                  class="binding-row"
                  @click="handleClickBinding(b)"
                >
                  <div class="binding-text">{{ b.bindText }}</div>
                  <el-icon class="arrow-icon"><ArrowRight /></el-icon>
                  <div class="binding-target">{{ getPointDisplayName(b.targetPointId) }}</div>
                </div>
              </div>
            </div>
          </el-popover>
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
          <!-- 选中文字展示（点击翻译） -->
          <el-tooltip content="点击翻译" placement="top" v-if="selectedText">
            <div class="selected-text-display clickable" @click="openTranslateDialog">
              <span class="selected-label">选中:</span>
              <span class="selected-content">{{ displaySelectedText }}</span>
            </div>
          </el-tooltip>
          
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

          <!-- 绑定/知识点按钮 -->
          <template v-if="selectedTextBindings && selectedTextBindings.length > 0">
            <el-popover
              placement="bottom"
              :width="280"
              trigger="click"
            >
              <template #reference>
                <el-button 
                  link 
                  class="trigger-btn highlight-text"
                  :disabled="!selectedText"
                >
                  <el-icon class="mr-1"><Connection /></el-icon> 知识点 ({{ selectedTextBindings.length }})
                </el-button>
              </template>
              <div class="binding-list-popover">
                <div class="binding-list-header">
                  <span>「{{ displaySelectedText }}」 已绑定:</span>
                  <el-button v-if="canEdit" link type="primary" size="small" @click="openBindingDialog">
                    <el-icon><Plus /></el-icon> 新增绑定
                  </el-button>
                </div>
                <div class="binding-list-items">
                  <div 
                    v-for="b in selectedTextBindings" 
                    :key="b.id" 
                    class="binding-item"
                  >
                    <div class="item-left" @click="handleClickBinding(b)">
                      <el-icon class="item-icon"><Link /></el-icon>
                      <span class="item-title">{{ getPointDisplayName(b.targetPointId) }}</span>
                    </div>
                    <el-button 
                      v-if="canEdit"
                      link 
                      type="danger" 
                      size="small" 
                      @click.stop="handleDeleteBinding(b.id)"
                    >
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                </div>
              </div>
            </el-popover>
          </template>
          <template v-else-if="canEdit">
            <el-tooltip :content="selectedText ? '绑定到其他知识点' : '请先在下方选择文字'" placement="top">
              <el-button 
                link 
                class="trigger-btn"
                :class="!selectedText ? 'disabled-text' : 'primary-text'"
                :disabled="!selectedText"
                @click="openBindingDialog"
              >
                <el-icon class="mr-1"><Connection /></el-icon> 绑定
              </el-button>
            </el-tooltip>
          </template>

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
            <el-input 
              v-model="currentFontSize" 
              placeholder="字号" 
              size="small" 
              style="width: 80px; margin-right: 8px;"
              @keyup.enter="applyFontSize"
              @focus="isFontSizeInputFocused = true"
              @blur="isFontSizeInputFocused = false"
              clearable
            >
              <template #suffix>
                <span style="font-size: 12px; color: #909399;">px</span>
              </template>
            </el-input>
            <el-button size="small" @click="insertCustomDivider" style="margin-right: 8px;">插入分割线</el-button>
            <el-button size="small" @click="cancelEdit" class="cancel-btn">取消</el-button>
            <el-button type="primary" size="small" icon="Check" class="gradient-btn" @click="saveEdit">保存</el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 编辑器工具栏（和标题栏平级） -->
    <div v-if="isEditing" class="editor-toolbar-container" ref="editorToolbarContainerRef"></div>

    <!-- 可滚动内容区域 -->
    <div class="scrollable-wrapper" ref="scrollableWrapperRef">
      <div class="content-box custom-scrollbar">
      <div v-if="isEditing" class="editor-wrapper" :style="{ opacity: editorReady ? 1 : 0 }">
        <RichTextEditor 
          ref="richTextEditorRef"
          :model-value="innerContent"
          @update:model-value="innerContent = $event"
          :point-id="pointId"
          :external-toolbar="true"
          @ready="onEditorReady"
        />
      </div>

      <div v-else class="html-preview ck ck-content ck-editor__editable" ref="previewRef" @mouseup="captureSelection" @touchend="captureSelection">
        <div v-if="content" v-html="processedContent" class="markdown-body"></div>
        <div v-else class="empty-tip">
          <el-icon :size="40"><Edit /></el-icon>
          <p>暂无详细内容哦(⊙o⊙)？</p>
        </div>
      </div>
    </div>
    <!-- /scrollable-wrapper -->
  </div>

    <!-- 绑定知识点弹窗 -->
    <el-dialog
      v-model="bindingDialogVisible"
      title="绑定到其他知识点"
      width="500px"
      append-to-body
      destroy-on-close
    >
      <div class="binding-form" v-loading="bindingLoading">
        <div class="binding-text-preview">
          <span class="binding-label">绑定文字：</span>
          <span class="binding-text">{{ selectedText }}</span>
        </div>
        
        <el-form label-width="80px" class="binding-selects">
          <el-form-item label="目标分类">
            <el-select
              v-model="selectedBindCategory"
              placeholder="请选择分类"
              style="width: 100%"
              @change="handleBindCategoryChange"
              filterable
            >
              <el-option
                v-for="category in bindingCategories"
                :key="category.id"
                :label="category.name"
                :value="category.id"
              />
            </el-select>
          </el-form-item>
          
          <el-form-item label="目标知识点">
            <el-select
              v-model="selectedBindPoint"
              placeholder="请先选择分类"
              style="width: 100%"
              :disabled="!selectedBindCategory"
              filterable
            >
              <el-option
                v-for="point in bindingPoints"
                :key="point.id"
                :label="point.title"
                :value="point.id"
              />
            </el-select>
          </el-form-item>
        </el-form>
      </div>
      
      <template #footer>
        <el-button @click="bindingDialogVisible = false">取消</el-button>
        <el-button 
          type="primary" 
          @click="submitBinding" 
          :loading="bindingSubmitting"
          :disabled="!selectedBindCategory || !selectedBindPoint"
        >
          确认绑定
        </el-button>
      </template>
    </el-dialog>

    <!-- 谷歌翻译弹窗 -->
    <el-dialog
      v-model="translateDialogVisible"
      title="谷歌翻译 (右下角可拖拽大小)"
      width="auto"
      class="resizable-translate-dialog"
      append-to-body
      draggable
      align-center
      destroy-on-close
      show-close
      :modal="false"
      :lock-scroll="false"
      :close-on-click-modal="false"
      modal-class="translate-overlay-transparent"
    >
      <div 
        class="translate-resizable-wrapper" 
        @mousedown="isTranslateResizing = true" 
        @mouseup="isTranslateResizing = false"
        @mouseleave="isTranslateResizing = false"
      >
        <div v-show="isTranslateResizing" class="resize-mask"></div>
        <iframe 
          :src="translateUrl" 
          class="translate-iframe"
          sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
        ></iframe>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, onBeforeUnmount, watch, onMounted, computed } from "vue";
import { ElMessage } from "element-plus";
import { Microphone, VideoPause, VideoPlay, SwitchButton, Reading, Edit, Check, ChatLineSquare, Headset, Document, Connection, Link, Plus, Delete, ArrowRight } from '@element-plus/icons-vue';
import { updatePoint, getPoints } from "../api/point";
import { createBinding, getCategoriesBySubjectForBinding, getPointsByCategoryForBinding, deleteBinding } from "../api/binding";
import RichTextEditor from './RichTextEditor.vue'; 

// ------------------------------------------------------------------
// 逻辑完全保持不变
// ------------------------------------------------------------------

const props = defineProps({
  pointId: { type: Number, required: true },
  subjectId: { type: Number, default: 0 }, // 当前科目 ID
  content: { type: String, default: '' },
  canEdit: { type: Boolean, default: false },
  bindings: { type: Array, default: () => [] }, // 绑定列表: [{id, bindText, targetPointId}]
  pointsInfoMap: { type: Map, default: () => new Map() } // 知识点信息缓存: pointId -> {title, categoryName}
});

const emit = defineEmits(["update", "goto-point", "refresh-bindings", "cache-point", "navigate-to-point"]);

const richTextEditorRef = ref<any>(null);
const scrollableWrapperRef = ref<HTMLElement | null>(null);
const editorToolbarContainerRef = ref<HTMLElement | null>(null);
const editorReady = ref(false); // 编辑器是否准备好（工具栏移动完成）
const mode = "default";
const isEditing = ref(false);
const innerContent = ref("");
const previewRef = ref<HTMLElement | null>(null);
const savedScrollTop = ref(0); // 保存切换前的滚动位置
const currentFontSize = ref<string>(''); // 当前选中文字的字号
const isFontSizeInputFocused = ref(false); // 字号输入框是否聚焦 

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

// 翻译弹窗状态
const translateDialogVisible = ref(false);
const translateUrl = ref('');
const isTranslateResizing = ref(false);

// 绑定弹窗状态
const bindingDialogVisible = ref(false);
const bindingsPopoverVisible = ref(false); // 绑定列表弹出框
const bindingCategories = ref<{id: number; name: string}[]>([]);
const bindingPoints = ref<{id: number; title: string}[]>([]);
const selectedBindCategory = ref<number | null>(null);
const selectedBindPoint = ref<number | null>(null);
const bindingLoading = ref(false);
const bindingSubmitting = ref(false);

// 处理内容：把绑定文字高亮显示
const processedContent = computed(() => {
  if (!props.content) return '';
  if (!props.bindings || props.bindings.length === 0) return props.content;
  
  let result = props.content;
  // 按 bindText 长度降序排列，避免短字符串先被替换影响长字符串
  const sortedBindings = [...props.bindings].sort((a: any, b: any) => 
    (b.bindText?.length || 0) - (a.bindText?.length || 0)
  );
  
  for (const binding of sortedBindings) {
    const text = (binding as any).bindText;
    const targetPointId = (binding as any).targetPointId;
    if (!text) continue;
    
    // 转义特殊字符
    const escapedText = text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const regex = new RegExp(`(?<!<[^>]*)${escapedText}(?![^<]*>)`, 'g');
    result = result.replace(regex, 
      `<span class="binding-link" data-target-point-id="${targetPointId}">${text}</span>`
    );
  }
  return result;
});

// 为代码块添加 Copy 按钮
const addCopyButtonsToCodeBlocks = () => {
  if (!previewRef.value) return;
  
  // 查找所有代码块（pre 标签）
  const codeBlocks = previewRef.value.querySelectorAll('pre');
  
  codeBlocks.forEach((pre: HTMLElement) => {
    // 避免重复添加
    if (pre.parentElement?.classList.contains('code-block-wrapper')) return;
    
    // 创建包装层
    const wrapper = document.createElement('div');
    wrapper.className = 'code-block-wrapper';
    wrapper.style.position = 'relative';
    
    // 创建 Copy 按钮
    const btnContainer = document.createElement('div');
    btnContainer.className = 'code-copy-btn';
    btnContainer.innerHTML = 'Copy';
    btnContainer.title = '复制代码';
    
    // 点击复制（只复制 pre 的文本，不包括按钮）
    btnContainer.addEventListener('click', () => {
      const code = pre.innerText;
      navigator.clipboard.writeText(code).then(() => {
        btnContainer.innerHTML = '✓ 已复制';
        btnContainer.style.color = '#67c23a';
        setTimeout(() => {
          btnContainer.innerHTML = 'Copy';
          btnContainer.style.color = '';
        }, 2000);
      }).catch(() => {
        ElMessage.error('复制失败');
      });
    });
    
    // 用包装层替换原 pre 元素
    pre.parentNode?.insertBefore(wrapper, pre);
    wrapper.appendChild(pre);
    wrapper.appendChild(btnContainer);
    
    // 确保 pre 元素样式不受影响
    pre.style.margin = '0';
  });
};

// 当前朗读文字显示（截取显示）
const displayReadingText = computed(() => {
  const text = currentFullText.value;
  if (!text) return '';
  // 从当前位置开始截取一段文字
  const start = Math.max(0, currentCharIndex.value);
  const end = Math.min(text.length, start + 30);
  const snippet = text.substring(start, end);
  return snippet + (end < text.length ? '...' : '');
});

// 选中文字显示（截取显示）
const displaySelectedText = computed(() => {
  const text = selectedText.value;
  if (!text) return '';
  if (text.length <= 20) return text;
  return text.substring(0, 20) + '...';
}); 

// 已加载过缓存的分类ID集合
const loadedCategoryIds = ref<Set<number>>(new Set());

// 根据targetPointId获取显示名称
const getPointDisplayName = (targetPointId: number): string => {
  // 1. 优先从 bindings 中查找（后端直接返回了标题和分类名）
  const binding = props.bindings?.find((b: any) => b.targetPointId === targetPointId) as any;
  if (binding?.targetPointTitle && binding?.targetCategoryName) {
    return `${binding.targetCategoryName} → ${binding.targetPointTitle}`;
  }
  
  // 2. 其次从缓存中查找
  const info = props.pointsInfoMap?.get(targetPointId) as {title: string; categoryName: string} | undefined;
  if (info) {
    return `${info.categoryName} → ${info.title}`;
  }
  
  // 3. 都找不到才显示默认格式
  return `知识点 #${targetPointId}`;
};

// 当前选中文字的所有绑定
const selectedTextBindings = computed((): {id: number; bindText: string; targetPointId: number; targetCategoryId?: number}[] => {
  if (!selectedText.value || !props.bindings) return [];
  return props.bindings.filter((b: any) => b.bindText === selectedText.value) as any[];
});

// 按需加载分类下的知识点缓存
const ensureBindingsCached = async () => {
  if (!props.bindings || props.bindings.length === 0) return;
  
  // 找出所有未缓存的分类ID
  const missingCategoryIds = new Set<number>();
  for (const b of props.bindings as any[]) {
    const targetPointId = b.targetPointId;
    const targetCategoryId = b.targetCategoryId;
    // 如果该知识点不在缓存中，且有分类ID，且该分类未加载过
    if (targetCategoryId && !props.pointsInfoMap?.has(targetPointId) && !loadedCategoryIds.value.has(targetCategoryId)) {
      missingCategoryIds.add(targetCategoryId);
    }
  }
  
  // 加载缺失的分类
  for (const categoryId of missingCategoryIds) {
    try {
      const res = await getPoints(categoryId);
      if (res.data?.code === 200 && res.data.data?.list) {
        // 获取分类名称（从绑定列表匹配不到，暂时用空）
        // 实际上我们需要从 categories 中获取名称
        for (const p of res.data.data.list) {
          if (!props.pointsInfoMap?.has(p.id)) {
            // 添加到缓存
            emit('cache-point', { pointId: p.id, title: p.title, categoryId });
          }
        }
        loadedCategoryIds.value.add(categoryId);
      }
    } catch (e) {
      console.error('Failed to load category points cache:', e);
    }
  }
};

// 监听bindings变化，自动加载缺失的缓存
watch(() => props.bindings, () => {
  ensureBindingsCached();
}, { immediate: true }); 

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

// 打开绑定弹窗
const openBindingDialog = async () => {
  if (!selectedText.value) {
    ElMessage.warning('请先选择要绑定的文字');
    return;
  }
  if (!props.subjectId) {
    ElMessage.warning('无法获取当前科目');
    return;
  }
  bindingDialogVisible.value = true;
  bindingLoading.value = true;
  selectedBindCategory.value = null;
  selectedBindPoint.value = null;
  bindingPoints.value = [];
  
  try {
    const res = await getCategoriesBySubjectForBinding(props.subjectId);
    bindingCategories.value = res.data.data || [];
  } catch (e) {
    ElMessage.error('获取分类列表失败');
  } finally {
    bindingLoading.value = false;
  }
};

// 分类变更时加载知识点
const handleBindCategoryChange = async (categoryId: number) => {
  selectedBindPoint.value = null;
  bindingPoints.value = [];
  if (!categoryId) return;
  
  bindingLoading.value = true;
  try {
    const res = await getPointsByCategoryForBinding(categoryId);
    bindingPoints.value = res.data.data || [];
  } catch (e) {
    ElMessage.error('获取知识点列表失败');
  } finally {
    bindingLoading.value = false;
  }
};

// 提交绑定
const submitBinding = async () => {
  if (!selectedBindCategory.value || !selectedBindPoint.value) {
    ElMessage.warning('请选择目标分类和知识点');
    return;
  }
  
  // 不允许自己绑定自己
  if (selectedBindPoint.value === props.pointId) {
    ElMessage.warning('不能绑定到自己');
    return;
  }
  
  bindingSubmitting.value = true;
  try {
    await createBinding({
      sourceSubjectId: props.subjectId,
      sourcePointId: props.pointId,
      targetSubjectId: props.subjectId, // 同一科目内绑定
      targetPointId: selectedBindPoint.value,
      bindText: selectedText.value
    });
    ElMessage.success('绑定成功');
    bindingDialogVisible.value = false;
    emit('refresh-bindings'); // 通知父组件刷新
  } catch (e) {
    ElMessage.error('绑定失败');
  } finally {
    bindingSubmitting.value = false;
  }
};

// 删除绑定
const handleDeleteBinding = async (bindingId: number) => {
  try {
    await deleteBinding(bindingId);
    ElMessage.success('已删除绑定');
    emit('refresh-bindings'); // 通知父组件刷新
  } catch (e) {
    ElMessage.error('删除失败');
  }
};

// 点击绑定链接，跳转到目标知识点
const handleClickBinding = (binding: {targetPointId: number; targetCategoryId?: number}) => {
  emit('navigate-to-point', {
    pointId: binding.targetPointId,
    categoryId: binding.targetCategoryId || 0
  });
};

// 打开翻译（新窗口）
const openTranslateDialog = () => {
  if (!selectedText.value) return;
  const text = encodeURIComponent(selectedText.value);
  const url = `https://translate.google.com/?sl=auto&tl=zh-CN&text=${text}&op=translate`;
  // 新窗口打开，设置窗口大小和位置
  const width = 800;
  const height = 600;
  const left = (screen.width - width) / 2;
  const top = (screen.height - height) / 2;
  window.open(url, 'GoogleTranslate', `width=${width},height=${height},left=${left},top=${top},resizable=yes,scrollbars=yes`);
};

watch(() => props.content, (newVal) => {
  if (!isEditing.value) innerContent.value = newVal || "";
  handleStop();
  // 内容更新后，为代码块添加 Copy 按钮
  setTimeout(() => {
    addCopyButtonsToCodeBlocks();
  }, 100);
}, { immediate: true });

// 切换知识点时关闭弹窗
watch(() => props.pointId, () => {
  bindingsPopoverVisible.value = false;
  selectedText.value = '';
  // 如果当前在编辑模式，切换知识点时自动退出编辑模式
  if (isEditing.value) {
    isEditing.value = false;
    editorReady.value = false;
    innerContent.value = props.content || '';
  }
});

watch(isEditing, (newVal) => {
  if (newVal) {
    handleStop();
    innerContent.value = props.content || "";
  } else {
    // 切换到预览模式时，等待编辑器销毁
    setTimeout(() => {
      richTextEditorRef.value = null;
    }, 100);
  }
});

const startEdit = () => {
  // 保存当前滚动位置（保存scrollableWrapper的滚动位置）
  if (scrollableWrapperRef.value) {
    savedScrollTop.value = scrollableWrapperRef.value.scrollTop;
  }
  innerContent.value = props.content || "";
  isEditing.value = true;
  editorReady.value = false; // 先隐藏编辑器
};

// 编辑器初始化完成回调
const onEditorReady = () => {
  // 等待下一帧，确保 DOM 完全渲染
  requestAnimationFrame(() => {
    if (richTextEditorRef.value && editorToolbarContainerRef.value) {
      const editor = richTextEditorRef.value.getEditor();
      
      if (editor && editor.ui && editor.ui.view && editor.ui.view.toolbar && editor.ui.view.toolbar.element) {
        // 清空顶部容器
        editorToolbarContainerRef.value.innerHTML = '';
        // 移动工具栏到顶部容器
        editorToolbarContainerRef.value.appendChild(editor.ui.view.toolbar.element);
        
        // 监听编辑器选区变化，实时更新字号显示
        editor.model.document.on('change:data', () => {
          updateFontSizeFromEditor();
        });
        
        // 监听选区变化
        editor.model.document.selection.on('change', () => {
          updateFontSizeFromEditor();
        });
      }
    }
    // 恢复滚动位置（而不是滚动到顶部）
    if (scrollableWrapperRef.value && savedScrollTop.value > 0) {
      scrollableWrapperRef.value.scrollTop = savedScrollTop.value;
    }
    // 显示编辑器
    editorReady.value = true;
  });
};

const cancelEdit = () => {
  isEditing.value = false;
  editorReady.value = false; // 重置状态
  innerContent.value = props.content || "";
  // 恢复滚动位置
  restoreScrollPosition();
};

const saveEdit = async () => {
  try {
    await updatePoint(props.pointId, { content: innerContent.value });
    emit("update", innerContent.value);
    emit('refresh-bindings'); // 刷新绑定列表（后端可能自动清理了不匹配的绑定）
    isEditing.value = false;
    editorReady.value = false; // 重置状态
    ElMessage.success("保存成功");
    // 恢复滚动位置
    restoreScrollPosition();
  } catch (e) {
    ElMessage.error("保存失败");
  }
};

// 恢复预览区滚动位置
const restoreScrollPosition = () => {
  if (savedScrollTop.value > 0) {
    setTimeout(() => {
      if (scrollableWrapperRef.value) {
        scrollableWrapperRef.value.scrollTop = savedScrollTop.value;
      }
    }, 50);
  }
};

// 插入自定义分割线
const insertCustomDivider = () => {
  if (!richTextEditorRef.value) {
    ElMessage.warning('编辑器未初始化');
    return;
  }
  richTextEditorRef.value.insertCustomDivider();
};

// 应用自定义字号到选中文字
// 从编辑器获取当前选区的字号
const updateFontSizeFromEditor = () => {
  if (!isEditing.value) return;
  if (isFontSizeInputFocused.value) return; // 输入框聚焦时不更新
  if (!richTextEditorRef.value) return;
  
  const editor = richTextEditorRef.value.getEditor();
  if (!editor) return;
  
  // 获取当前选区的 fontSize 属性
  const selection = editor.model.document.selection;
  const fontSize = selection.getAttribute('fontSize');
  
  if (fontSize) {
    // fontSize 格式通常是 "16px" 这样的字符串
    const sizeNumber = parseInt(fontSize);
    if (!isNaN(sizeNumber)) {
      currentFontSize.value = `${sizeNumber}`;
    }
  } else {
    // 如果没有设置字号，获取默认字号
    currentFontSize.value = '';
  }
};

const applyFontSize = () => {
  if (!currentFontSize.value) {
    return;
  }
  
  const fontSize = parseInt(currentFontSize.value);
  if (isNaN(fontSize) || fontSize < 1 || fontSize > 200) {
    return;
  }
  
  if (!richTextEditorRef.value) {
    return;
  }
  
  const editor = richTextEditorRef.value.getEditor();
  if (!editor) return;
  
  // 设置选中文字的字号
  editor.execute('fontSize', { value: `${fontSize}px` });
};

// 监听编辑器内的选区变化，自动更新字号显示
watch(isEditing, (newVal) => {
  if (newVal) {
    // 进入编辑模式时，监听选区变化
    setTimeout(() => {
      document.addEventListener('selectionchange', updateFontSizeOnSelection);
    }, 500);
  } else {
    // 退出编辑模式时，移除监听
    document.removeEventListener('selectionchange', updateFontSizeOnSelection);
    currentFontSize.value = '';
  }
});

const updateFontSizeOnSelection = () => {
  if (!isEditing.value) return;
  if (isFontSizeInputFocused.value) return; // 输入框聚焦时不更新
  
  const domSelection = window.getSelection();
  if (!domSelection || domSelection.rangeCount === 0 || domSelection.toString().trim() === '') {
    // 不清空输入框，允许用户继续编辑
    return;
  }
  
  const range = domSelection.getRangeAt(0);
  let element: Node | null = range.startContainer;
  
  if (element.nodeType === Node.TEXT_NODE) {
    element = element.parentElement;
  }
  
  if (element && element instanceof HTMLElement) {
    const computedStyle = window.getComputedStyle(element);
    const actualFontSize = computedStyle.fontSize;
    const fontSizeNumber = parseInt(actualFontSize);
    currentFontSize.value = `${fontSizeNumber}`; // 只显示数字，不px在输入框后缀
  }
};
</script>

<style scoped>
.content-column { 
  display: flex; 
  flex-direction: column; 
  height: 100%; 
  background: transparent;
  overflow: hidden;
}

.section-header { 
  display: flex; 
  justify-content: space-between; 
  align-items: center; 
  padding: 14px 20px;
  flex-shrink: 0; 
  border-bottom: 2px solid #e4e7ed;
  background: linear-gradient(to bottom, #fafbfc 0%, #f5f7fa 100%);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  z-index: 10;
}

/* 编辑器工具栏容器（和标题栏平级） */
.editor-toolbar-container {
  flex-shrink: 0;
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  z-index: 100;
  position: relative;
}

/* 可滚动wrapper */
.scrollable-wrapper {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
}

.content-box { 
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 0;
}

.left-group { display: flex; align-items: center; gap: 20px; }
.section-title { 
  font-weight: 600;
  color: #303133;
  font-size: 16px;
  display: flex;
  align-items: center;
  letter-spacing: 0.5px;
}
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
.editor-wrapper { 
  width: 100%;
  border: 1px solid rgba(0,0,0,0.1);
  border-radius: 8px;
  overflow: visible;
  background: rgba(255,255,255,0.6);
  z-index: 10;
  transition: opacity 0.2s ease-in-out; /* 添加平滑过渡 */
}
:deep(.w-e-toolbar) { background-color: rgba(249, 250, 251, 0.9) !important; }
:deep(.w-e-text-container) { background-color: transparent !important; }
:deep(.w-e-bar-item button:hover) { color: #764ba2; }
.html-preview { 
  flex: 1;
  padding: var(--ck-spacing-large); /* 使用CKEditor的间距变量，约24px */
  line-height: var(--ck-line-height-base, 1.6);
  color: var(--ck-color-text, #000);
  font-size: var(--ck-font-size-base, 16px);
  overflow-y: auto;
  cursor: text;
  font-family: var(--ck-font-face, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif);
  /* 添加CKEditor内容区域的基础样式 */
  word-wrap: break-word;
}
/* 使用CKEditor的样式类名 */
.markdown-body :deep(*) {
  font-size: inherit;
  line-height: inherit;
}
.markdown-body :deep(p) { 
  margin-bottom: 1em; /* 与CKEditor保持一致 */
}
.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4),
.markdown-body :deep(h5),
.markdown-body :deep(h6) {
  margin-top: 1.5em;
  margin-bottom: 0.5em;
  font-weight: bold;
  line-height: 1.3;
}
.markdown-body :deep(h1) { font-size: 2em; }
.markdown-body :deep(h2) { font-size: 1.5em; }
.markdown-body :deep(h3) { font-size: 1.25em; }
.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  padding-left: 40px;
  margin: 1em 0;
}
.markdown-body :deep(li) {
  margin: 0.5em 0;
}
.markdown-body :deep(img) { max-width: 100%; border-radius: 6px; box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
.markdown-body :deep(blockquote) { border-left: 4px solid #d3adf7; background: rgba(249, 240, 255, 0.5); padding: 10px 15px; margin: 10px 0; color: #666; border-radius: 4px; }
.markdown-body :deep(code) { background-color: rgba(0,0,0,0.05); padding: 2px 5px; border-radius: 4px; font-family: monospace; color: #c7254e; }
.markdown-body :deep(pre) { background-color: #f1f1f1; color: #333; padding: 10px; border-radius: 4px; border: 1px solid #ccc; }

/* 阅读模式下的表格样式 */
.markdown-body :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 10px 0;
  border: 1px solid #dcdfe6;
}
.markdown-body :deep(table th),
.markdown-body :deep(table td) {
  border: 1px solid #dcdfe6;
  padding: 8px 12px;
  text-align: left;
}
.markdown-body :deep(table th) {
  background-color: #f5f7fa;
  font-weight: 600;
  color: #303133;
}
.markdown-body :deep(table tr:hover) {
  background-color: #f5f7fa;
}

/* 代码块包装层 */
.markdown-body :deep(.code-block-wrapper) {
  position: relative;
  margin: 10px 0;
}

/* 代码块 Copy 按钮样式 */
.markdown-body :deep(.code-copy-btn) {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 4px 12px;
  background: rgba(64, 158, 255, 0.1);
  border: 1px solid rgba(64, 158, 255, 0.3);
  border-radius: 4px;
  font-size: 12px;
  color: #409eff;
  cursor: pointer;
  user-select: none;
  transition: all 0.2s;
  font-weight: 500;
  z-index: 10;
}
.markdown-body :deep(.code-copy-btn:hover) {
  background: rgba(64, 158, 255, 0.2);
  border-color: #409eff;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(64, 158, 255, 0.2);
}

/* 绑定链接样式：蓝色+下划线 */
.markdown-body :deep(.binding-link) {
  color: #409eff;
  text-decoration: underline;
  cursor: pointer;
  transition: all 0.2s;
}
.markdown-body :deep(.binding-link:hover) {
  color: #764ba2;
  text-decoration-color: #764ba2;
}
.empty-tip { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%; color: rgba(0,0,0,0.3); margin-top: 40px; }
.empty-tip p { margin-top: 10px; font-size: 14px; }
.custom-scrollbar::-webkit-scrollbar { width: 6px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(0,0,0,0.1); border-radius: 3px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }

/* 朗读文字展示 */
.reading-text-display {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: linear-gradient(90deg, rgba(102, 126, 234, 0.1), rgba(118, 75, 162, 0.1));
  border-radius: 12px;
  max-width: 200px;
  animation: pulse-glow 2s ease-in-out infinite;
}

/* 选中文字展示 */
.selected-text-display {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  background: linear-gradient(90deg, rgba(64, 158, 255, 0.1), rgba(102, 126, 234, 0.1));
  border: 1px solid rgba(64, 158, 255, 0.2);
  border-radius: 14px;
  max-width: 220px;
  transition: all 0.2s;
}

.selected-text-display.clickable {
  cursor: pointer;
}

.selected-text-display.clickable:hover {
  background: linear-gradient(90deg, rgba(64, 158, 255, 0.2), rgba(102, 126, 234, 0.2));
  border-color: rgba(64, 158, 255, 0.4);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
}

.selected-label {
  font-size: 11px;
  color: #409eff;
  font-weight: 600;
  white-space: nowrap;
}

.selected-content {
  font-size: 12px;
  color: #303133;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 160px;
}

/* 绑定弹窗样式 */
.binding-form {
  min-height: 150px;
}

.binding-text-preview {
  background: linear-gradient(90deg, rgba(102, 126, 234, 0.1), rgba(118, 75, 162, 0.1));
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 20px;
  border: 1px solid rgba(118, 75, 162, 0.15);
}

.binding-text-preview .binding-label {
  font-size: 13px;
  color: #764ba2;
  font-weight: 600;
  margin-right: 8px;
}

.binding-text-preview .binding-text {
  font-size: 14px;
  color: #303133;
  word-break: break-all;
}

.binding-selects {
  margin-top: 10px;
}

/* 绑定列表弹出框样式 */
.binding-list-popover {
  padding: 4px 0;
}

.binding-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 4px 8px;
  border-bottom: 1px solid #eee;
  margin-bottom: 8px;
}

.binding-list-header span {
  font-size: 13px;
  color: #606266;
  font-weight: 500;
}

.binding-list-items {
  max-height: 200px;
  overflow-y: auto;
}

.binding-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.binding-item:hover {
  background: linear-gradient(90deg, rgba(64, 158, 255, 0.1), rgba(118, 75, 162, 0.1));
}

.binding-item .item-left {
  display: flex;
  align-items: center;
  flex: 1;
  cursor: pointer;
}

.binding-item .item-icon {
  color: #409eff;
  margin-right: 8px;
  font-size: 14px;
}

.binding-item .item-title {
  font-size: 13px;
  color: #303133;
}

/* 绑定数量（内联样式） */
.binding-count-inline {
  display: inline-flex;
  align-items: center;
  color: #409eff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  margin-left: 2px;
  transition: all 0.2s;
}
.binding-count-inline:hover {
  color: #764ba2;
}
.binding-count-inline .el-icon {
  font-size: 12px;
  margin-right: 2px;
}

/* 全部绑定关系弹出框 */
.all-bindings-popover {
  padding: 0;
}
.all-bindings-header {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
  margin-bottom: 8px;
}
.all-bindings-list {
  max-height: 240px;
  overflow-y: auto;
}
.binding-row {
  display: flex;
  align-items: center;
  padding: 8px 6px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  gap: 8px;
}
.binding-row:hover {
  background: linear-gradient(90deg, rgba(64, 158, 255, 0.1), rgba(118, 75, 162, 0.1));
}
.binding-row .binding-text {
  font-size: 12px;
  color: #764ba2;
  background: rgba(118, 75, 162, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex-shrink: 0;
}
.binding-row .arrow-icon {
  color: #c0c4cc;
  font-size: 12px;
  flex-shrink: 0;
}
.binding-row .binding-target {
  font-size: 13px;
  color: #409eff;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

@keyframes pulse-glow {
  0%, 100% { box-shadow: 0 0 4px rgba(118, 75, 162, 0.2); }
  50% { box-shadow: 0 0 8px rgba(118, 75, 162, 0.4); }
}

.reading-label {
  font-size: 11px;
  color: #764ba2;
  font-weight: 600;
  white-space: nowrap;
}

.reading-content {
  font-size: 12px;
  color: #606266;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 140px;
}
</style>

<!-- 翻译弹窗全局样式 -->
<style>
/* 透明遮罩层 */
.translate-overlay-transparent {
  pointer-events: none !important;
  background-color: transparent !important;
  overflow: hidden !important;
}

/* 弹窗本体 */
.translate-overlay-transparent .el-dialog {
  pointer-events: auto !important;
  margin: 0 !important;
  background: #fff !important;
  border-radius: 8px !important;
  box-shadow: 0 10px 40px rgba(0,0,0,0.3) !important;
  display: flex !important;
  flex-direction: column !important;
  width: auto !important;
}

/* 标题栏 */
.translate-overlay-transparent .el-dialog__header {
  padding: 12px 16px !important;
  background: #fff !important;
  border-bottom: 1px solid #eee !important;
  margin: 0 !important;
  flex-shrink: 0;
  cursor: move !important;
  user-select: none;
}

.translate-overlay-transparent .el-dialog__title {
  color: #303133 !important;
  font-size: 15px !important;
  font-weight: 600 !important;
}

.translate-overlay-transparent .el-dialog__headerbtn {
  top: 14px !important;
}

.translate-overlay-transparent .el-dialog__headerbtn .el-dialog__close {
  color: #909399 !important;
}

.translate-overlay-transparent .el-dialog__headerbtn:hover .el-dialog__close {
  color: #409eff !important;
}

/* 内容区 */
.translate-overlay-transparent .el-dialog__body {
  padding: 8px !important;
  margin: 0 !important;
  background: #fff !important;
  flex: 1;
  display: flex;
  height: auto !important;
}

/* Flex 布局容器 */
.translate-overlay-transparent .el-overlay-dialog {
  pointer-events: none !important;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 可调整大小的包装器 */
.translate-resizable-wrapper {
  width: 800px;
  height: 600px;
  min-width: 400px;
  min-height: 300px;
  background: #f5f5f5;
  position: relative;
  display: flex;
  resize: both;
  overflow: auto;
}

/* 调整大小时的透明遮罩 */
.translate-resizable-wrapper .resize-mask {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  z-index: 998;
  background: transparent;
}

/* 右下角拖拽手柄 */
.translate-resizable-wrapper::after {
  content: '';
  position: absolute;
  bottom: 0;
  right: 0;
  width: 15px;
  height: 15px;
  cursor: se-resize;
  z-index: 999;
  background: linear-gradient(135deg, transparent 50%, rgba(0,0,0,0.1) 50%);
  pointer-events: auto;
}

/* 翻译 iframe */
.translate-iframe {
  width: 100%;
  height: 100%;
  border: none;
  background: #fff;
}
</style>