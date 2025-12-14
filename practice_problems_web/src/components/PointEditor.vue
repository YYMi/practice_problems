<template>
  <div class="content-column">
    <!-- ... 顶部的 header 代码完全不变 ... -->
    <div class="section-header">
      <div class="left-group">
        <div class="section-title">
          <el-icon class="mr-1"><Reading /></el-icon> 知识详解
          <el-popover v-if="bindings && bindings.length > 0" v-model:visible="bindingsPopoverVisible" placement="bottom" :width="320" trigger="click">
            <template #reference><span class="binding-count-inline">(<el-icon><Link /></el-icon>{{ bindings.length }})</span></template>
            <div class="all-bindings-popover">
              <div class="all-bindings-header">全部绑定关系 ({{ bindings.length }})</div>
              <div class="all-bindings-list custom-scrollbar">
                <div v-for="b in (bindings as any[])" :key="b.id" class="binding-row" @click="handleClickBinding(b)">
                  <div class="binding-text">{{ b.bindText }}</div>
                  <el-icon class="arrow-icon"><ArrowRight /></el-icon>
                  <div class="binding-target">{{ getPointDisplayName(b.targetPointId) }}</div>
                </div>
              </div>
            </div>
          </el-popover>
          <el-tooltip content="查看笔记" placement="top">
            <el-button link class="note-btn" @click="openNoteDialog">
              <el-icon class="note-icon"><Document /></el-icon>
            </el-button>
          </el-tooltip>
        </div>
        <!-- 播放模块 (保持不变) -->
        <div class="player-module" v-if="!isEditing && content">
          <el-select v-model="selectedVoiceURI" placeholder="选择语音" size="small" class="voice-select" @change="handleVoiceChange" :teleported="false">
            <template #prefix><el-icon class="select-icon"><Headset /></el-icon></template>
            <el-option v-for="voice in voiceList" :key="voice.voiceURI" :label="voice.name" :value="voice.voiceURI">
              <span style="float: left">{{ voice.name }}</span><span style="float: right; color: #8492a6; font-size: 12px; margin-left: 10px;">{{ voice.lang }}</span>
            </el-option>
          </el-select>
          <div class="divider-small"></div>
          <div class="speed-box"><span class="speed-label">{{ speechRate.toFixed(1) }}x</span><el-slider v-model="speechRate" :min="0.5" :max="2.0" :step="0.1" size="small" class="custom-slider" @change="handleRateChange" /></div>
          <div class="divider-small"></div>
          <el-tooltip :content="speechStatus === 'paused' ? '继续' : '暂停'" placement="top">
            <el-button link class="control-btn" :class="speechStatus === 'playing' ? 'warning' : 'success'" :disabled="speechStatus === 'stopped'" @click="togglePauseResume">
              <el-icon size="18"><VideoPlay v-if="speechStatus === 'paused'" /><VideoPause v-else /></el-icon>
            </el-button>
          </el-tooltip>
          <el-tooltip content="停止" placement="top">
            <el-button link class="control-btn danger" :disabled="speechStatus === 'stopped'" @click="handleStop"><el-icon size="18"><SwitchButton /></el-icon></el-button>
          </el-tooltip>
        </div>
      </div>
      <!-- 右侧操作区 (保持不变) -->
      <div class="action-area">
        <div class="trigger-group" v-if="!isEditing && content">
          <el-tooltip content="点击翻译" placement="top" v-if="selectedText">
            <div class="selected-text-display clickable" @click="openTranslateDialog">
              <span class="selected-label">选中:</span><span class="selected-content">{{ displaySelectedText }}</span>
            </div>
          </el-tooltip>
          <el-tooltip :content="selectedText ? '朗读选中的文字' : '请先在下方选择文字'" placement="top">
            <el-button link class="trigger-btn" :class="(readingMode === 'full' || !selectedText) ? 'disabled-text' : 'primary-text'" :disabled="readingMode === 'full' || !selectedText" @click="startSelectedReading">
              <el-icon class="mr-1"><ChatLineSquare /></el-icon> 朗读
            </el-button>
          </el-tooltip>
          <template v-if="selectedTextBindings && selectedTextBindings.length > 0">
            <el-popover placement="bottom" :width="280" trigger="click">
              <template #reference>
                <el-button link class="trigger-btn highlight-text" :disabled="!selectedText"><el-icon class="mr-1"><Connection /></el-icon> 知识点 ({{ selectedTextBindings.length }})</el-button>
              </template>
              <div class="binding-list-popover">
                <div class="binding-list-header"><span>「{{ displaySelectedText }}」 已绑定:</span><el-button v-if="canEdit" link type="primary" size="small" @click="openBindingDialog"><el-icon><Plus /></el-icon> 新增绑定</el-button></div>
                <div class="binding-list-items">
                  <div v-for="b in selectedTextBindings" :key="b.id" class="binding-item"><div class="item-left" @click="handleClickBinding(b)"><el-icon class="item-icon"><Link /></el-icon><span class="item-title">{{ getPointDisplayName(b.targetPointId) }}</span></div><el-button v-if="canEdit" link type="danger" size="small" @click.stop="handleDeleteBinding(b.id)"><el-icon><Delete /></el-icon></el-button></div>
                </div>
              </div>
            </el-popover>
          </template>
          <template v-else-if="canEdit">
            <el-tooltip :content="selectedText ? '绑定到其他知识点' : '请先在下方选择文字'" placement="top">
              <el-button link class="trigger-btn" :class="!selectedText ? 'disabled-text' : 'primary-text'" :disabled="!selectedText" @click="openBindingDialog"><el-icon class="mr-1"><Connection /></el-icon> 绑定</el-button>
            </el-tooltip>
          </template>
          <el-tooltip content="从头朗读全文" placement="top">
            <el-button link class="trigger-btn" :class="readingMode === 'full' ? 'highlight-text' : 'primary-text'" @click="startFullReading"><el-icon class="mr-1"><Microphone /></el-icon> 全文朗读</el-button>
          </el-tooltip>
          <el-tooltip content="AI 模拟面试" placement="top">
            <el-button link class="trigger-btn primary-text" @click="openAIInterviewer"><el-icon class="mr-1"><Service /></el-icon> AI面试官</el-button>
          </el-tooltip>
        </div>
        <div class="divider-vertical" v-if="(!isEditing && content) && canEdit"></div>
        <div class="edit-controls" v-if="canEdit">
          <el-button v-if="!isEditing" type="primary" size="small" icon="Edit" class="gradient-btn" @click="startEdit">编辑内容</el-button>
          <div v-else class="edit-actions">
            <el-input v-model="currentFontSize" placeholder="字号" size="small" style="width: 80px; margin-right: 8px;" @keyup.enter="applyFontSize" @focus="isFontSizeInputFocused = true" @blur="isFontSizeInputFocused = false" clearable><template #suffix><span style="font-size: 12px; color: #909399;">px</span></template></el-input>
            <el-button size="small" @click="insertCustomDivider" style="margin-right: 8px;">插入分割线</el-button>
            <el-button size="small" @click="cancelEdit" class="cancel-btn">取消</el-button>
            <el-button type="primary" size="small" icon="Check" class="gradient-btn" @click="saveEdit">保存</el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 编辑器工具栏 -->
    <div v-if="isEditing" class="editor-toolbar-container" ref="editorToolbarContainerRef"></div>

    <!-- 可滚动内容区域 -->
    <div class="scrollable-wrapper" ref="scrollableWrapperRef">
      <div class="content-box custom-scrollbar">
        <div v-if="isEditing" class="editor-wrapper" :style="{ opacity: editorReady ? 1 : 0 }">
          <RichTextEditor ref="richTextEditorRef" :model-value="innerContent" @update:model-value="innerContent = $event" :point-id="pointId" :external-toolbar="true" @ready="onEditorReady" />
        </div>
        <div v-else class="html-preview ck ck-content ck-editor__editable" ref="previewRef" @mouseup="captureSelection" @touchend="captureSelection">
          <div v-if="content" v-html="processedContent" class="markdown-body"></div>
          <div v-else class="empty-tip"><el-icon :size="40"><Edit /></el-icon><p>暂无详细内容哦(⊙o⊙)？</p></div>
        </div>
      </div>
    </div>

    <!-- 弹窗部分 (绑定、翻译、AI) 保持不变 -->
    <el-dialog v-model="bindingDialogVisible" title="绑定到其他知识点" width="500px" append-to-body destroy-on-close>
      <div class="binding-form" v-loading="bindingLoading">
        <div class="binding-text-preview"><span class="binding-label">绑定文字：</span><span class="binding-text">{{ selectedText }}</span></div>
        <el-form label-width="80px" class="binding-selects">
          <el-form-item label="目标分类"><el-select v-model="selectedBindCategory" placeholder="请选择分类" style="width: 100%" @change="handleBindCategoryChange" filterable><el-option v-for="category in bindingCategories" :key="category.id" :label="category.name" :value="category.id" /></el-select></el-form-item>
          <el-form-item label="目标知识点"><el-select v-model="selectedBindPoint" placeholder="请先选择分类" style="width: 100%" :disabled="!selectedBindCategory" filterable><el-option v-for="point in bindingPoints" :key="point.id" :label="point.title" :value="point.id" /></el-select></el-form-item>
        </el-form>
      </div>
      <template #footer><el-button @click="bindingDialogVisible = false">取消</el-button><el-button type="primary" @click="submitBinding" :loading="bindingSubmitting" :disabled="!selectedBindCategory || !selectedBindPoint">确认绑定</el-button></template>
    </el-dialog>
    <el-dialog v-model="translateDialogVisible" title="谷歌翻译 (右下角可拖拽大小)" width="auto" class="resizable-translate-dialog" append-to-body draggable align-center destroy-on-close show-close :modal="false" :lock-scroll="false" :close-on-click-modal="false" modal-class="translate-overlay-transparent">
      <div class="translate-resizable-wrapper" @mousedown="isTranslateResizing = true" @mouseup="isTranslateResizing = false" @mouseleave="isTranslateResizing = false"><div v-show="isTranslateResizing" class="resize-mask"></div><iframe :src="translateUrl" class="translate-iframe" sandbox="allow-scripts allow-same-origin allow-forms allow-popups"></iframe></div>
    </el-dialog>
    <AIInterviewer v-model="aiInterviewerVisible" :point-id="pointId" :point-title="pointTitle" :point-content="content" />
    
    <!-- 【修改后】笔记弹窗：明确区分 阅读模式 vs 编辑模式 -->
    <el-dialog
      v-model="noteDialogVisible"
      title="我的笔记"
      width="800px" 
      append-to-body
      destroy-on-close
      class="simple-note-dialog"
      :close-on-click-modal="false"
      :show-close="true"
    >
      <!-- 自定义 Header -->
      <template #header>
        <div class="custom-dialog-header">
          <span class="dialog-title">我的笔记</span>
          
          <!-- 核心修改：模式切换按钮组 -->
          <div class="mode-switch-group">
            <el-button-group>
              <el-button 
                :type="!isNoteEditing ? 'primary' : 'default'" 
                size="small"
                @click="isNoteEditing = false"
              >
                <el-icon class="mr-1"><View /></el-icon> 阅读模式
              </el-button>
              <el-button 
                :type="isNoteEditing ? 'primary' : 'default'" 
                size="small"
                @click="startEditingNote"
              >
                <el-icon class="mr-1"><EditPen /></el-icon> 编辑模式
              </el-button>
            </el-button-group>
          </div>
        </div>
      </template>
      
      <!-- 副标题：显示当前知识点 -->
      <div class="note-sub-header">
        <span class="point-title">{{ pointTitle }}</span>
      </div>
      
      <div class="note-body-wrapper" v-loading="noteLoading">
        
        <!-- 1. 阅读模式：只显示漂亮的渲染结果 -->
        <div 
          v-show="!isNoteEditing" 
          class="preview-container markdown-body custom-scrollbar"
        >
          <div v-if="noteContent" v-html="renderedNote"></div>
          <!-- 空状态 -->
          <div v-else class="empty-preview-static">
            <el-icon :size="48" color="#dcdfe6"><Document /></el-icon>
            <p>暂无笔记内容，请点击右上角“编辑模式”开始撰写</p>
          </div>
        </div>

        <!-- 2. 编辑模式：只显示输入框 -->
        <div v-show="isNoteEditing" class="edit-container">
          <el-input
            ref="noteInputRef"
            v-model="noteContent"
            type="textarea"
            placeholder="在此输入 Markdown 内容..."
            resize="none"
            class="simple-textarea"
          />
        </div>
      </div>
      
      <!-- 底部统计与操作 -->
      <div class="note-stats">
        <span>当前模式: {{ isNoteEditing ? '编辑中' : '阅读中' }}</span>
        <span class="ml-3">字数: {{ noteContent.length }}</span>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="clearNote" class="clear-btn" type="danger" link>清空</el-button>
          <div class="spacer"></div>
          <el-button @click="noteDialogVisible = false">关闭</el-button>
          <!-- 只有在编辑模式下，保存按钮才高亮，或者一直高亮也行 -->
          <el-button type="primary" @click="saveNote" :loading="noteLoading" class="save-btn">
            保存笔记
          </el-button>
        </div>
      </template>
    </el-dialog>


  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, onBeforeUnmount, watch, onMounted, computed, nextTick } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import MarkdownIt from 'markdown-it';
import { Microphone, VideoPause, VideoPlay, SwitchButton, Reading, Edit, Check, ChatLineSquare, Headset, Document, Connection, Link, Plus, Delete, ArrowRight, Service, Loading, EditPen } from '@element-plus/icons-vue';
import { updatePoint, getPoints } from "../api/point";
import { createBinding, getCategoriesBySubjectForBinding, getPointsByCategoryForBinding, deleteBinding } from "../api/binding";
import { getPointNote, savePointNote } from "../api/pointNote";
import RichTextEditor from './RichTextEditor.vue'; 
import AIInterviewer from './AIInterviewer.vue'; 

const props = defineProps({
  pointId: { type: Number, required: true },
  pointTitle: { type: String, default: '' }, 
  subjectId: { type: Number, default: 0 }, 
  content: { type: String, default: '' },
  canEdit: { type: Boolean, default: false },
  bindings: { type: Array, default: () => [] }, 
  pointsInfoMap: { type: Map, default: () => new Map() } 
});
const emit = defineEmits(["update", "goto-point", "refresh-bindings", "cache-point", "navigate-to-point"]);

// RichText Editor & Controls
const richTextEditorRef = ref<any>(null);
const scrollableWrapperRef = ref<HTMLElement | null>(null);
const editorToolbarContainerRef = ref<HTMLElement | null>(null);
const editorReady = ref(false); 
const isEditing = ref(false);
const innerContent = ref("");
const previewRef = ref<HTMLElement | null>(null);
const savedScrollTop = ref(0); 
const currentFontSize = ref<string>(''); 
const isFontSizeInputFocused = ref(false); 

// Speech
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

// Dialogs
const translateDialogVisible = ref(false);
const translateUrl = ref('');
const isTranslateResizing = ref(false);
const bindingDialogVisible = ref(false);
const bindingsPopoverVisible = ref(false); 
const bindingCategories = ref<{id: number; name: string}[]>([]);
const bindingPoints = ref<{id: number; title: string}[]>([]);
const selectedBindCategory = ref<number | null>(null);
const selectedBindPoint = ref<number | null>(null);
const bindingLoading = ref(false);
const bindingSubmitting = ref(false);
const aiInterviewerVisible = ref(false);

// Note (Click-to-Edit Markdown)
const noteDialogVisible = ref(false);
const noteContent = ref(""); 
const noteLoading = ref(false); 
const isNoteEditing = ref(false); // 控制当前是显示预览还是编辑框
const noteInputRef = ref<any>(null);

const mdNote = new MarkdownIt({ 
  html: true, 
  linkify: true, 
  breaks: true 
});

const renderedNote = computed(() => {
  return mdNote.render(noteContent.value || "");
});

// Main content markdown renderer
const md = new MarkdownIt({ html: false, linkify: true, breaks: true });

// Actions
const openAIInterviewer = () => { aiInterviewerVisible.value = true; };

const openNoteDialog = async () => {
  noteLoading.value = true;
  noteDialogVisible.value = true;
  isNoteEditing.value = false; // 默认预览模式
  try {
    const res = await getPointNote(props.pointId);
    if (res.data?.code === 200) {
      noteContent.value = res.data.data.note || "";
      // 如果没内容，自动进入编辑模式
      if (!noteContent.value) {
        isNoteEditing.value = true;
      }
    } else {
      noteContent.value = "";
      isNoteEditing.value = true;
    }
  } catch (error) {
    ElMessage.error("获取笔记失败");
    noteContent.value = "";
  } finally {
    noteLoading.value = false;
  }
};

// 点击预览区域 -> 进入编辑模式
const startEditingNote = () => {
 isNoteEditing.value = true;
  nextTick(() => {
    noteInputRef.value?.focus(); // 切换到编辑模式后，光标自动进去
  });
};

// 失去焦点(点旁边) -> 回到预览模式
const stopEditingNote = () => {
  isNoteEditing.value = false;
};

const saveNote = async () => {
  try {
    noteLoading.value = true; // 开启按钮转圈，防止重复点击
    
    // 调用后台保存接口
    await savePointNote(props.pointId, noteContent.value);
    
    ElMessage.success("笔记保存成功");

    // =========== 核心修改 ===========
    // noteDialogVisible.value = false;  <-- 这行代码我删掉了！
    // ===============================
    
    // 保存后，自动切回“阅读模式”让你看看渲染效果（如果你想保存后继续编辑，把下面这行也注释掉即可）
    isNoteEditing.value = false; 

  } catch (error) {
    ElMessage.error("保存笔记失败");
  } finally {
    noteLoading.value = false; // 关闭转圈
  }
};

const clearNote = () => {
  ElMessageBox.confirm('确定要清空笔记内容吗？', '确认清空', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(() => {
    noteContent.value = '';
    isNoteEditing.value = true; // 清空后自动进入编辑
    ElMessage.success('笔记已清空');
  }).catch(() => {});
};

// Bindings Logic
const processedContent = computed(() => {
  if (!props.content) return '';
  if (!props.bindings || props.bindings.length === 0) return props.content;
  let result = props.content;
  const sortedBindings = [...props.bindings].sort((a: any, b: any) => (b.bindText?.length || 0) - (a.bindText?.length || 0));
  for (const binding of sortedBindings) {
    const text = (binding as any).bindText;
    const targetPointId = (binding as any).targetPointId;
    if (!text) continue;
    const escapedText = text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const regex = new RegExp(`(?<!<[^>]*)${escapedText}(?![^<]*>)`, 'g');
    result = result.replace(regex, `<span class="binding-link" data-target-point-id="${targetPointId}">${text}</span>`);
  }
  return result;
});

const addCopyButtonsToCodeBlocks = () => {
  if (!previewRef.value) return;
  const codeBlocks = previewRef.value.querySelectorAll('pre');
  codeBlocks.forEach((pre: HTMLElement) => {
    if (pre.parentElement?.classList.contains('code-block-wrapper')) return;
    const wrapper = document.createElement('div');
    wrapper.className = 'code-block-wrapper';
    wrapper.style.position = 'relative';
    const btnContainer = document.createElement('div');
    btnContainer.className = 'code-copy-btn';
    btnContainer.innerHTML = 'Copy';
    btnContainer.title = '复制代码';
    btnContainer.addEventListener('click', () => {
      const code = pre.innerText;
      navigator.clipboard.writeText(code).then(() => {
        btnContainer.innerHTML = '✓ 已复制';
        btnContainer.style.color = '#67c23a';
        setTimeout(() => { btnContainer.innerHTML = 'Copy'; btnContainer.style.color = ''; }, 2000);
      });
    });
    pre.parentNode?.insertBefore(wrapper, pre);
    wrapper.appendChild(pre);
    wrapper.appendChild(btnContainer);
    pre.style.margin = '0';
  });
};

const displayReadingText = computed(() => {
  const text = currentFullText.value;
  if (!text) return '';
  const start = Math.max(0, currentCharIndex.value);
  const end = Math.min(text.length, start + 30);
  return text.substring(start, end) + (end < text.length ? '...' : '');
});

const displaySelectedText = computed(() => {
  const text = selectedText.value;
  if (!text) return '';
  return text.length <= 20 ? text : text.substring(0, 20) + '...';
}); 

const loadedCategoryIds = ref<Set<number>>(new Set());
const getPointDisplayName = (targetPointId: number): string => {
  const binding = props.bindings?.find((b: any) => b.targetPointId === targetPointId) as any;
  if (binding?.targetPointTitle && binding?.targetCategoryName) return `${binding.targetCategoryName} → ${binding.targetPointTitle}`;
  const info = props.pointsInfoMap?.get(targetPointId) as {title: string; categoryName: string} | undefined;
  return info ? `${info.categoryName} → ${info.title}` : `知识点 #${targetPointId}`;
};

const selectedTextBindings = computed(() => {
  if (!selectedText.value || !props.bindings) return [];
  return props.bindings.filter((b: any) => b.bindText === selectedText.value) as any[];
});

const ensureBindingsCached = async () => {
  if (!props.bindings || props.bindings.length === 0) return;
  const missingCategoryIds = new Set<number>();
  for (const b of props.bindings as any[]) {
    const targetCategoryId = b.targetCategoryId;
    if (targetCategoryId && !props.pointsInfoMap?.has(b.targetPointId) && !loadedCategoryIds.value.has(targetCategoryId)) {
      missingCategoryIds.add(targetCategoryId);
    }
  }
  for (const categoryId of missingCategoryIds) {
    try {
      const res = await getPoints(categoryId);
      if (res.data?.code === 200 && res.data.data?.list) {
        for (const p of res.data.data.list) {
          if (!props.pointsInfoMap?.has(p.id)) emit('cache-point', { pointId: p.id, title: p.title, categoryId });
        }
        loadedCategoryIds.value.add(categoryId);
      }
    } catch (e) {}
  }
};

watch(() => props.bindings, () => { ensureBindingsCached(); }, { immediate: true }); 

onMounted(() => {
  const savedRate = localStorage.getItem('user-speech-rate');
  if (savedRate) speechRate.value = parseFloat(savedRate);
  initVoices();
  if (speechSynthesis.onvoiceschanged !== undefined) speechSynthesis.onvoiceschanged = initVoices;
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
  if (savedVoice && voiceList.value.find(v => v.voiceURI === savedVoice)) selectedVoiceURI.value = savedVoice;
  else {
    const zhVoice = voiceList.value.find(v => v.lang.includes('zh-CN'));
    selectedVoiceURI.value = zhVoice ? zhVoice.voiceURI : (voiceList.value.length > 0 ? voiceList.value[0].voiceURI : "");
  }
};

onBeforeUnmount(() => { handleStop(); document.removeEventListener('click', handleGlobalClick); });
const handleGlobalClick = (e: MouseEvent) => {};
const stripHtml = (html: string) => { const tmp = document.createElement("DIV"); tmp.innerHTML = html; return tmp.textContent || tmp.innerText || ""; };

const togglePauseResume = () => {
  if (speechStatus.value === 'playing') { synth.pause(); speechStatus.value = 'paused'; }
  else if (speechStatus.value === 'paused') { synth.resume(); speechStatus.value = 'playing'; }
};
const handleVoiceChange = (val: string) => { localStorage.setItem('user-speech-voice', val); restartSpeechIfPlaying(); };
const handleRateChange = (val: number) => { localStorage.setItem('user-speech-rate', val.toString()); restartSpeechIfPlaying(); };
const restartSpeechIfPlaying = () => {
  if (speechStatus.value === 'playing' || speechStatus.value === 'paused') {
    const remainingText = currentFullText.value.substring(currentCharIndex.value);
    synth.cancel();
    setTimeout(() => { speak(remainingText, readingMode.value); }, 50);
  }
};
const captureSelection = () => {
  if (isEditing.value) return;
  const selection = window.getSelection();
  const previewDom = previewRef.value;
  if (!selection || selection.rangeCount === 0 || !previewDom || !previewDom.contains(selection.anchorNode)) { selectedText.value = ""; return; }
  selectedText.value = selection.toString().trim();
  if (speechStatus.value === 'playing' || speechStatus.value === 'paused') {
    const rangeToEnd = document.createRange();
    rangeToEnd.selectNodeContents(previewDom); 
    rangeToEnd.setStart(selection.getRangeAt(0).startContainer, selection.getRangeAt(0).startOffset); 
    const textToRead = rangeToEnd.toString();
    if (textToRead && textToRead.trim().length > 0) {
      if (synth.speaking) synth.cancel();
      readingMode.value = 'full';
      ElMessage.success("已跳转至选定位置播放");
      speak(textToRead, 'full'); 
    }
  }
};
const startFullReading = () => { speak(stripHtml(props.content), 'full'); };
const startSelectedReading = () => { if (readingMode.value !== 'full' && selectedText.value) speak(selectedText.value, 'selected'); };
const speak = (text: string, mode: ReadingMode) => {
  if (!text.trim()) { ElMessage.warning("没有可朗读的文本"); return; }
  synth.cancel(); readingMode.value = mode; currentFullText.value = text; currentCharIndex.value = 0; 
  setTimeout(() => {
    const newUtterance = new SpeechSynthesisUtterance(text);
    (window as any).currentUtterance = newUtterance; 
    if (selectedVoiceURI.value) {
      const v = synth.getVoices().find(v => v.voiceURI === selectedVoiceURI.value) || voiceList.value.find(v => v.voiceURI === selectedVoiceURI.value);
      if (v) newUtterance.voice = v as any;
    }
    newUtterance.lang = 'zh-CN'; newUtterance.rate = speechRate.value; 
    newUtterance.onboundary = (e) => { if (e.name === 'word' || e.name === 'sentence') currentCharIndex.value = e.charIndex; };
    newUtterance.onend = () => { speechStatus.value = 'stopped'; readingMode.value = 'none'; currentCharIndex.value = 0; };
    newUtterance.onerror = (e) => { if (e.error !== 'interrupted' && e.error !== 'canceled') { speechStatus.value = 'stopped'; readingMode.value = 'none'; } };
    synth.speak(newUtterance); speechStatus.value = 'playing'; utterance = newUtterance;
  }, 50);
};
const handleStop = () => { synth.cancel(); speechStatus.value = 'stopped'; readingMode.value = 'none'; currentCharIndex.value = 0; currentFullText.value = ""; };

const openBindingDialog = async () => {
  if (!selectedText.value) { ElMessage.warning('请先选择要绑定的文字'); return; }
  if (!props.subjectId) { ElMessage.warning('无法获取当前科目'); return; }
  bindingDialogVisible.value = true; bindingLoading.value = true; selectedBindCategory.value = null; selectedBindPoint.value = null; bindingPoints.value = [];
  try { const res = await getCategoriesBySubjectForBinding(props.subjectId); bindingCategories.value = res.data.data || []; } catch (e) { ElMessage.error('获取分类列表失败'); } finally { bindingLoading.value = false; }
};
const handleBindCategoryChange = async (categoryId: number) => {
  selectedBindPoint.value = null; bindingPoints.value = []; if (!categoryId) return;
  bindingLoading.value = true;
  try { const res = await getPointsByCategoryForBinding(categoryId); bindingPoints.value = res.data.data || []; } catch (e) { ElMessage.error('获取知识点列表失败'); } finally { bindingLoading.value = false; }
};
const submitBinding = async () => {
  if (!selectedBindCategory.value || !selectedBindPoint.value) { ElMessage.warning('请选择目标分类和知识点'); return; }
  if (selectedBindPoint.value === props.pointId) { ElMessage.warning('不能绑定到自己'); return; }
  bindingSubmitting.value = true;
  try { await createBinding({ sourceSubjectId: props.subjectId, sourcePointId: props.pointId, targetSubjectId: props.subjectId, targetPointId: selectedBindPoint.value, bindText: selectedText.value }); ElMessage.success('绑定成功'); bindingDialogVisible.value = false; emit('refresh-bindings'); } catch (e) { ElMessage.error('绑定失败'); } finally { bindingSubmitting.value = false; }
};
const handleDeleteBinding = async (bindingId: number) => { try { await deleteBinding(bindingId); ElMessage.success('已删除绑定'); emit('refresh-bindings'); } catch (e) { ElMessage.error('删除失败'); } };
const handleClickBinding = (b: any) => { emit('navigate-to-point', { pointId: b.targetPointId, categoryId: b.targetCategoryId || 0 }); };
const openTranslateDialog = () => {
  if (!selectedText.value) return;
  const url = `https://translate.google.com/?sl=auto&tl=zh-CN&text=${encodeURIComponent(selectedText.value)}&op=translate`;
  const w = 800, h = 600, l = (screen.width - w) / 2, t = (screen.height - h) / 2;
  window.open(url, 'GoogleTranslate', `width=${w},height=${h},left=${l},top=${t},resizable=yes,scrollbars=yes`);
};
watch(() => props.content, (v) => { if (!isEditing.value) innerContent.value = v || ""; handleStop(); setTimeout(addCopyButtonsToCodeBlocks, 100); }, { immediate: true });
watch(() => props.pointId, () => { bindingsPopoverVisible.value = false; selectedText.value = ''; if (isEditing.value) { isEditing.value = false; editorReady.value = false; innerContent.value = props.content || ''; } });
watch(isEditing, (v) => { if (v) { handleStop(); innerContent.value = props.content || ""; } else setTimeout(() => { richTextEditorRef.value = null; }, 100); });
const startEdit = () => { if (scrollableWrapperRef.value) savedScrollTop.value = scrollableWrapperRef.value.scrollTop; innerContent.value = props.content || ""; isEditing.value = true; editorReady.value = false; };
const onEditorReady = () => {
  requestAnimationFrame(() => {
    if (richTextEditorRef.value && editorToolbarContainerRef.value) {
      const editor = richTextEditorRef.value.getEditor();
      if (editor && editor.ui && editor.ui.view) { editorToolbarContainerRef.value.innerHTML = ''; editorToolbarContainerRef.value.appendChild(editor.ui.view.toolbar.element); }
    }
    if (scrollableWrapperRef.value && savedScrollTop.value > 0) scrollableWrapperRef.value.scrollTop = savedScrollTop.value;
    editorReady.value = true;
  });
};
const cancelEdit = () => { isEditing.value = false; editorReady.value = false; innerContent.value = props.content || ""; if (savedScrollTop.value > 0) setTimeout(() => { if (scrollableWrapperRef.value) scrollableWrapperRef.value.scrollTop = savedScrollTop.value; }, 50); };
const saveEdit = async () => { try { await updatePoint(props.pointId, { content: innerContent.value }); emit("update", innerContent.value); emit('refresh-bindings'); isEditing.value = false; editorReady.value = false; ElMessage.success("保存成功"); if (savedScrollTop.value > 0) setTimeout(() => { if (scrollableWrapperRef.value) scrollableWrapperRef.value.scrollTop = savedScrollTop.value; }, 50); } catch (e) { ElMessage.error("保存失败"); } };
const insertCustomDivider = () => { if (richTextEditorRef.value) richTextEditorRef.value.insertCustomDivider(); };
const updateFontSizeFromEditor = () => { if (!isEditing.value || isFontSizeInputFocused.value || !richTextEditorRef.value) return; const sel = richTextEditorRef.value.getEditor().model.document.selection; const fs = sel.getAttribute('fontSize'); currentFontSize.value = fs ? parseInt(fs).toString() : ''; };
const applyFontSize = () => { if (currentFontSize.value && richTextEditorRef.value) richTextEditorRef.value.getEditor().execute('fontSize', { value: `${parseInt(currentFontSize.value)}px` }); };
watch(isEditing, (v) => { if (v) setTimeout(() => document.addEventListener('selectionchange', updateFontSizeOnSelection), 500); else { document.removeEventListener('selectionchange', updateFontSizeOnSelection); currentFontSize.value = ''; } });
const updateFontSizeOnSelection = () => { if (!isEditing.value || isFontSizeInputFocused.value) return; 
const sel = window.getSelection(); if (!sel || sel.rangeCount === 0 || !sel.toString().trim()) return; 
const r = sel.getRangeAt(0); let el = r.startContainer; if (el.nodeType === Node.TEXT_NODE) el = el.parentElement; if (el instanceof HTMLElement) currentFontSize.value = parseInt(window.getComputedStyle(el).fontSize).toString(); };
</script>

<style scoped>
/* =========================================
   1. 组件独有样式 (只在当前组件生效)
   ========================================= */

/* 播放器模块 (中间独有) */
.player-module {
  display: flex;
  align-items: center;
  background: rgba(242, 235, 255, 0.9); 
  border: 1px solid rgba(118, 75, 162, 0.1); 
  border-radius: 20px;
  padding: 4px 12px;
  height: 32px;
  gap: 8px;
  transition: all 0.3s;
  box-shadow: 0 1px 3px rgba(118, 75, 162, 0.1);
}
.player-module:hover { 
  background: rgba(233, 222, 255, 0.95); 
  box-shadow: 0 2px 8px rgba(118, 75, 162, 0.15);
}

.voice-select { width: 50px; }
:deep(.voice-select .el-input__wrapper) { background-color: transparent !important; box-shadow: none !important; padding: 0; }
:deep(.voice-select .el-input__inner) { font-size: 12px; color: #606266; font-weight: 500; height: 32px; }
.select-icon { color: #606266; margin-right: 0px; font-size: 14px; }
.speed-box { display: flex; align-items: center; }
.speed-label { font-size: 12px; color: #606266; margin-right: 8px; width: 24px; text-align: right; }
.custom-slider { width: 60px; }
.divider-small { width: 1px; height: 14px; background-color: #dcdfe6; margin: 0 4px; }
.control-btn { padding: 0; height: auto; }
.control-btn.warning { color: #e6a23c; }
.control-btn.success { color: #67c23a; }
.control-btn.danger { color: #f56c6c; }
.control-btn.is-disabled { color: #c0c4cc; }

/* 顶部右侧操作区 */
.action-area { display: flex; align-items: center; gap: 15px; }
.trigger-group { display: flex; gap: 10px; }
.trigger-btn { font-weight: 600; font-size: 14px; }
.primary-text { color: #606266; }
.primary-text:hover { color: #409eff; }
.highlight-text { color: #764ba2; font-weight: bold; }
.disabled-text { color: #c0c4cc; cursor: not-allowed; }
.divider-vertical { width: 1px; height: 16px; background-color: #e4e7ed; }
.note-btn { margin-left: 10px; color: #909399; font-size: 16px; transition: all 0.3s; }
.note-btn:hover { color: #409eff; transform: scale(1.1); }
.note-icon { font-size: 18px; }

/* 编辑器容器 */
.editor-toolbar-container { flex-shrink: 0; background: #fff; border-bottom: 1px solid #e8e8e8; z-index: 100; position: relative; }
.editor-wrapper { width: 100%; border: 1px solid rgba(0,0,0,0.1); border-radius: 8px; background: rgba(255,255,255,0.6); z-index: 10; transition: opacity 0.2s ease-in-out; }

/* 预览内容样式 */
.html-preview { flex: 1; padding: 24px; line-height: 1.6; color: #333; font-size: 16px; overflow-y: auto; word-wrap: break-word; }
.empty-tip { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%; color: rgba(0,0,0,0.3); margin-top: 40px; }

/* 交互组件 */
.selected-text-display { display: flex; align-items: center; gap: 6px; padding: 4px 12px; background: linear-gradient(90deg, rgba(64, 158, 255, 0.1), rgba(102, 126, 234, 0.1)); border: 1px solid rgba(64, 158, 255, 0.2); border-radius: 14px; max-width: 220px; }
.selected-text-display.clickable { cursor: pointer; }
.selected-label { font-size: 11px; color: #409eff; font-weight: 600; white-space: nowrap; }
.selected-content { font-size: 12px; color: #303133; font-weight: 500; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.binding-list-items { max-height: 200px; overflow-y: auto; }
.binding-item { display: flex; align-items: center; justify-content: space-between; padding: 8px 10px; border-radius: 6px; cursor: pointer; }
.binding-item:hover { background: linear-gradient(90deg, rgba(64, 158, 255, 0.1), rgba(118, 75, 162, 0.1)); }
.binding-count-inline { display: inline-flex; align-items: center; color: #409eff; font-size: 14px; font-weight: 600; cursor: pointer; margin-left: 2px; }

/* 笔记弹窗内部 - 编辑器 */
.edit-container { flex: 1; display: flex; flex-direction: column; background: #fff; }
.simple-textarea { height: 100%; display: flex; }
.simple-textarea :deep(.el-textarea__inner) { height: 100% !important; border: none !important; border-radius: 0; padding: 24px 32px; background: transparent; font-family: 'Menlo', 'Monaco', monospace; font-size: 15px; line-height: 1.8; resize: none; box-shadow: none !important; }
.simple-textarea :deep(.el-textarea__inner:focus) { box-shadow: none !important; }

/* 笔记弹窗内部 - 预览 */
.preview-container { flex: 1; padding: 24px 32px; overflow-y: auto; cursor: text; }
.preview-container:hover { background-color: #fafbfc; }
.empty-preview-clickable { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #c0c4cc; cursor: pointer; background: #fbfbfb; transition: all 0.3s; }
.empty-preview-clickable:hover { color: #764ba2; background: #f5f0fa; }
.empty-preview-clickable p { margin-top: 10px; font-size: 14px; }

/* 笔记弹窗内部 - 头部 */
.custom-dialog-header { display: flex; align-items: center; gap: 12px; }
.dialog-title { font-size: 18px; font-weight: 600; color: #fff; text-shadow: 0 1px 2px rgba(0,0,0,0.1); }
.header-tips { display: flex; align-items: center; background: rgba(255, 255, 255, 0.2); padding: 2px 10px; border-radius: 12px; }
.tip-text { font-size: 12px; color: rgba(255, 255, 255, 0.95); display: flex; align-items: center; gap: 4px; }
.tip-text.editing { color: #fff; font-weight: 600; }
.note-sub-header { padding-bottom: 12px; margin-bottom: 10px; border-bottom: 1px dashed #ebeef5; }
.point-title { font-size: 16px; font-weight: 600; color: #303133; padding-left: 8px; border-left: 4px solid #764ba2; line-height: 1.4; }
.note-body-wrapper { flex: 1; min-height: 0; display: flex; flex-direction: column; border: 1px solid #dcdfe6; border-radius: 8px; overflow: hidden; background: #fff; position: relative; margin-top: 0; }
.note-stats { text-align: right; font-size: 12px; color: #909399; margin-top: 8px; flex-shrink: 0; }
.dialog-footer { display: flex; justify-content: flex-end; gap: 12px; padding-top: 10px; }
.clear-btn, .cancel-btn { padding: 10px 24px; font-size: 14px; border-radius: 6px; }
.save-btn { padding: 10px 24px; font-size: 14px; border-radius: 6px; background: linear-gradient(90deg, #667eea, #764ba2); border: none; color: white; box-shadow: 0 4px 12px rgba(118, 75, 162, 0.3); }
.save-btn:hover { opacity: 0.9; transform: translateY(-1px); }

/* --- 模式切换按钮样式 --- */
.mode-switch-group {
  margin-left: auto; /* 靠右对齐 */
  margin-right: 32px; /* 距离关闭按钮一点距离 */
}
.mode-switch-group .el-button {
  border-radius: 4px;
  font-weight: 600;
  padding: 8px 16px;
}
/* 选中状态不仅变色，还加点阴影 */
.mode-switch-group .el-button--primary {
  background: #fff;
  color: #764ba2;
  border-color: #fff;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
/* 未选中状态半透明 */
.mode-switch-group .el-button--default {
  background: rgba(255,255,255,0.2);
  color: rgba(255,255,255,0.9);
  border: 1px solid rgba(255,255,255,0.3);
}

/* 静态空状态 (阅读模式下没内容) */
.empty-preview-static {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
}
.empty-preview-static p { margin-top: 16px; font-size: 14px; }

.ml-3 { margin-left: 12px; }
</style>

<!-- 
  =========================================
  全局样式 (Global Styles)
  注意：这里包含了通用的布局类，这样右侧栏也能用到！
  ========================================= 
-->
<style>
/* 1. 通用布局类 (解决右侧栏样式丢失问题) */
.content-column { display: flex; flex-direction: column; height: 100%; background: transparent; overflow: hidden; }
.section-header { display: flex; justify-content: space-between; align-items: center; padding: 14px 20px; flex-shrink: 0; border-bottom: 2px solid #e4e7ed; background: linear-gradient(to bottom, #fafbfc 0%, #f5f7fa 100%); box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05); z-index: 10; }
.scrollable-wrapper { flex: 1; overflow-y: auto; overflow-x: hidden; }
.content-box { flex: 1; display: flex; flex-direction: column; min-height: 0; padding: 0; }
.left-group { display: flex; align-items: center; gap: 20px; }
.section-title { font-weight: 600; color: #303133; font-size: 16px; display: flex; align-items: center; letter-spacing: 0.5px; }
.mr-1 { margin-right: 6px; }

/* 2. 滚动条美化 (全局生效，解决右侧栏滚动条丑的问题) */
.custom-scrollbar::-webkit-scrollbar { width: 6px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(0,0,0,0.1); border-radius: 3px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }

/* 3. Markdown 内容样式 (全局生效，解决预览格式) */
.markdown-body h1, .markdown-body h2, .markdown-body h3 { margin-top: 1.2em; margin-bottom: 0.5em; font-weight: bold; }
.markdown-body ul, .markdown-body ol { padding-left: 20px; margin: 1em 0; }
.markdown-body img { max-width: 100%; border-radius: 6px; box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
.markdown-body blockquote { border-left: 4px solid #d3adf7; background: rgba(249, 240, 255, 0.5); padding: 10px 15px; margin: 10px 0; color: #666; }
.markdown-body code { background-color: rgba(0,0,0,0.05); padding: 2px 5px; border-radius: 4px; font-family: monospace; color: #c7254e; }
.markdown-body pre { background-color: #f6f8fa; padding: 12px; border-radius: 6px; overflow-x: auto; }
.markdown-body table { width: 100%; border-collapse: collapse; margin: 10px 0; }
.markdown-body th, .markdown-body td { border: 1px solid #dcdfe6; padding: 8px 12px; }
.markdown-body th { background-color: #f5f7fa; font-weight: 600; }
.markdown-body .code-block-wrapper { position: relative; margin: 10px 0; }
.markdown-body .code-copy-btn { position: absolute; top: 8px; right: 8px; padding: 4px 12px; background: rgba(64, 158, 255, 0.1); border: 1px solid rgba(64, 158, 255, 0.3); border-radius: 4px; font-size: 12px; color: #409eff; cursor: pointer; z-index: 10; }
.markdown-body .binding-link { color: #409eff; text-decoration: underline; cursor: pointer; }

/* 4. 笔记弹窗样式修复 (全局覆盖 Element Plus) */
.simple-note-dialog { margin-top: 8vh !important; height: 80vh !important; display: flex !important; flex-direction: column !important; border-radius: 16px !important; box-shadow: 0 24px 48px rgba(0, 0, 0, 0.2) !important; overflow: hidden !important; border: none !important; }
.simple-note-dialog .el-dialog__header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important; margin-right: 0 !important; padding: 16px 24px !important; flex-shrink: 0; display: flex; align-items: center; justify-content: space-between; }
.simple-note-dialog .el-dialog__headerbtn { top: 0 !important; position: static !important; width: 32px; height: 32px; display: flex; align-items: center; justify-content: center; border-radius: 50%; transition: background 0.2s; }
.simple-note-dialog .el-dialog__headerbtn:hover { background: rgba(255,255,255,0.2); }
.simple-note-dialog .el-dialog__headerbtn .el-dialog__close { color: white !important; font-size: 20px !important; }
.simple-note-dialog .el-dialog__body { flex: 1 !important; height: 0 !important; min-height: 0 !important; padding: 24px 28px !important; display: flex !important; flex-direction: column !important; overflow: hidden !important; background-color: #fff; }
.simple-note-dialog .el-dialog__footer { padding: 16px 28px 20px !important; border-top: 1px solid #f2f2f2; background: #fff; flex-shrink: 0; }

/* 5. 翻译弹窗透明遮罩 */
.translate-overlay-transparent { pointer-events: none !important; background-color: transparent !important; overflow: hidden !important; }
.resizable-translate-dialog { pointer-events: auto !important; background-color: white !important; border-radius: 8px !important; box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1) !important; }
.translate-resizable-wrapper { position: relative; width: 100%; height: 100%; }
.translate-iframe { width: 100%; height: 100%; border: none; }
.resize-mask { position: absolute; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0, 0, 0, 0.1); cursor: se-resize; }
</style>


