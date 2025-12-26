<template>
  <!-- 笔记弹窗 -->
  <el-dialog
    v-model="dialogVisible"
    title="我的笔记"
    width="1200px" 
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
      </div>
    </template>
    
    <!-- 副标题 + 工具栏 同一行 -->
    <div class="note-sub-header">
      <span class="point-title">{{ pointTitle }}</span>
      <!-- 工具栏 -->
      <div class="note-toolbar">
        <div class="toolbar-group">
          <button @click="execCommand('bold')" title="加粗" class="tool-btn"><strong>B</strong></button>
          <button @click="execCommand('italic')" title="斜体" class="tool-btn"><em>I</em></button>
          <button @click="execCommand('underline')" title="下划线" class="tool-btn"><u>U</u></button>
          <button @click="execCommand('strikeThrough')" title="删除线" class="tool-btn"><s>S</s></button>
        </div>
        <div class="toolbar-divider"></div>
        <div class="toolbar-group">
          <button @click="execCommand('insertUnorderedList')" title="无序列表" class="tool-btn">•</button>
          <button @click="execCommand('insertOrderedList')" title="有序列表" class="tool-btn">1.</button>
        </div>
        <div class="toolbar-divider"></div>
        <div class="toolbar-group">
          <button @click="insertHeading('h2')" title="标题" class="tool-btn">H₂</button>
          <button @click="insertHeading('h3')" title="小标题" class="tool-btn">H₃</button>
        </div>
      </div>
    </div>
    
    <div class="note-body-wrapper" v-loading="noteLoading">
      <!-- 富文本编辑区域 -->
      <div
        ref="editorRef"
        class="note-editor custom-scrollbar"
        contenteditable="true"
        @input="handleInput"
        @keydown="handleKeydown"
        @paste="handlePaste"
        :data-placeholder="'在此输入笔记内容...'"
      ></div>
    </div>
    
    <!-- 底部统计与操作 -->
    <div class="note-stats">
      <span>字数: {{ textLength }}</span>
    </div>
    
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="clearNote" class="clear-btn" type="danger" link>清空</el-button>
        <div class="spacer"></div>
        <el-button @click="closeDialog">关闭</el-button>
        <el-button type="primary" @click="saveNote" :loading="noteLoading" class="save-btn">
          保存笔记
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { getPointNote, savePointNote } from "../api/pointNote";

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  pointId: { type: Number, required: true },
  pointTitle: { type: String, default: '' }
});

const emit = defineEmits(['update:modelValue']);

// 内部状态
const noteContent = ref("");
const noteLoading = ref(false);
const editorRef = ref<HTMLElement | null>(null);

// 字数统计
const textLength = computed(() => {
  return editorRef.value?.innerText?.length || 0;
});

// 弹窗可见性 (双向绑定)
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
});

// 关闭弹窗
const closeDialog = () => {
  dialogVisible.value = false;
};

// 监听弹窗打开，加载笔记
watch(() => props.modelValue, async (visible) => {
  if (visible && props.pointId) {
    await loadNote();
  }
});

// 加载笔记
const loadNote = async () => {
  noteLoading.value = true;
  try {
    const res = await getPointNote(props.pointId);
    if (res.data?.code === 200) {
      noteContent.value = res.data.data.note || "";
    } else {
      noteContent.value = "";
    }
  } catch (error) {
    ElMessage.error("获取笔记失败");
    noteContent.value = "";
  } finally {
    noteLoading.value = false;
    // 设置编辑器内容
    nextTick(() => {
      if (editorRef.value) {
        editorRef.value.innerHTML = noteContent.value;
      }
    });
  }
};

// 处理输入
const handleInput = () => {
  if (editorRef.value) {
    noteContent.value = editorRef.value.innerHTML;
  }
};

// 处理键盘事件
const handleKeydown = (e: KeyboardEvent) => {
  // Tab 键插入制表符
  if (e.key === 'Tab') {
    e.preventDefault();
    document.execCommand('insertText', false, '\t');
  }
};

// 处理粘贴 - 去除复杂格式，保留基本样式
const handlePaste = (e: ClipboardEvent) => {
  e.preventDefault();
  const text = e.clipboardData?.getData('text/html') || e.clipboardData?.getData('text/plain') || '';
  // 简单清理，保留基本标签
  const cleanHtml = text
    .replace(/<script[^>]*>[\s\S]*?<\/script>/gi, '')
    .replace(/<style[^>]*>[\s\S]*?<\/style>/gi, '')
    .replace(/class="[^"]*"/gi, '')
    .replace(/style="[^"]*"/gi, '');
  document.execCommand('insertHTML', false, cleanHtml);
};

// 执行富文本命令
const execCommand = (command: string, value?: string) => {
  document.execCommand(command, false, value);
  editorRef.value?.focus();
};

// 插入标题
const insertHeading = (tag: string) => {
  document.execCommand('formatBlock', false, tag);
  editorRef.value?.focus();
};

// 保存笔记
const saveNote = async () => {
  try {
    noteLoading.value = true;
    const content = editorRef.value?.innerHTML || '';
    await savePointNote(props.pointId, content);
    ElMessage.success("笔记保存成功");
  } catch (error) {
    ElMessage.error("保存笔记失败");
  } finally {
    noteLoading.value = false;
  }
};

// 清空笔记
const clearNote = () => {
  ElMessageBox.confirm('确定要清空笔记内容吗？', '确认清空', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(() => {
    if (editorRef.value) {
      editorRef.value.innerHTML = '';
    }
    noteContent.value = '';
    ElMessage.success('笔记已清空');
  }).catch(() => {});
};
</script>

<style scoped>
/* 头部 */
.custom-dialog-header { display: flex; align-items: center; }
.dialog-title { font-size: 18px; font-weight: 600; color: #5b21b6; }

/* 副标题 + 工具栏 */
.note-sub-header { 
  display: flex; 
  align-items: center; 
  justify-content: space-between;
  padding:  0; 
  margin-bottom: 12px; 
  border-bottom: 1px solid #ebeef5; 
}
.point-title { 
  font-size: 15px; 
  font-weight: 600; 
  color: #303133; 
  padding-left: 10px; 
  border-left: 3px solid #764ba2; 
}

/* 工具栏 - 精致样式 */
.note-toolbar {
  display: flex;
  align-items: center;
  gap: 4px;
  background: #f8f9fa;
  padding: 4px 8px;
  border-radius: 8px;
  border: 1px solid #e8e8e8;
}
.toolbar-group {
  display: flex;
  gap: 2px;
}
.toolbar-divider {
  width: 1px;
  height: 20px;
  background: #dcdfe6;
  margin: 0 6px;
}
.tool-btn {
  width: 32px;
  height: 28px;
  border: none;
  background: transparent;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  color: #606266;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}
.tool-btn:hover {
  background: #e8e8e8;
  color: #303133;
}
.tool-btn:active {
  background: #dcdfe6;
}
.tool-btn strong { font-weight: 700; }
.tool-btn em { font-style: italic; }
.tool-btn u { text-decoration: underline; }
.tool-btn s { text-decoration: line-through; }

/* 内容区域 */
.note-body-wrapper { 
  flex: 1; 
  min-height: 0; 
  display: flex; 
  flex-direction: column; 
  border: 1px solid #e4e7ed; 
  border-radius: 8px; 
  overflow: hidden; 
  background: #fff; 
  box-shadow: inset 0 1px 3px rgba(0,0,0,0.03);
}

/* 富文本编辑器 */
.note-editor {
  flex: 1;
  padding: 20px 24px;
  overflow-y: auto;
  font-size: 15px;
  line-height: 1.7;
  color: #333;
  outline: none;
  background: #fafafa;
}
.note-editor:empty:before {
  content: attr(data-placeholder);
  color: #c0c4cc;
  pointer-events: none;
}
.note-editor:focus {
  background-color: #f5f5f5;
}

/* 编辑器内容样式 */
.note-editor h1, .note-editor h2, .note-editor h3 { margin: 0.4em 0 0.3em; font-weight: 600; color: #1a1a1a; }
.note-editor h2 { font-size: 1.25em; }
.note-editor h3 { font-size: 1.1em; }
.note-editor ul, .note-editor ol { padding-left: 24px; margin: 0.4em 0; }
.note-editor li { margin: 0.25em 0; }
.note-editor p { margin: 0.4em 0; }

/* 底部 */
.note-stats { 
  display: flex;
  justify-content: flex-end;
  font-size: 12px; 
  color: #909399; 
  margin-top: 10px; 
  padding-top: 5px;
}
.dialog-footer { 
  display: flex; 
  align-items: center;
  gap: 12px; 
  padding-top: 8px; 
}
.clear-btn { margin-right: auto; }
.save-btn { 
  padding: 10px 28px; 
  font-size: 14px; 
  font-weight: 500;
  border-radius: 8px; 
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); 
  border: none; 
  color: white; 
  box-shadow: 0 4px 12px rgba(118, 75, 162, 0.35); 
  transition: all 0.2s;
}
.save-btn:hover { 
  transform: translateY(-1px); 
  box-shadow: 0 6px 16px rgba(118, 75, 162, 0.4); 
}

.spacer { flex: 1; }
</style>

<!-- 全局样式 - 笔记弹窗 -->
<style>
/* 笔记弹窗样式修复 (全局覆盖 Element Plus) */
.simple-note-dialog { margin-top: 5vh !important; height: 85vh !important; display: flex !important; padding:0px; flex-direction: column !important; border-radius: 16px !important; box-shadow: 0 24px 48px rgba(0, 0, 0, 0.2) !important; overflow: hidden !important; border: none !important; }
.simple-note-dialog .el-dialog__header { background: linear-gradient(135deg, #f5f3ff 0%, #ede9fe 100%) !important; border-bottom: 1px solid #e9e4f5 !important; margin-right: 0 !important; padding: 10px 24px !important; flex-shrink: 0; display: flex; align-items: center; justify-content: space-between; }
.simple-note-dialog .el-dialog__headerbtn { top: 0 !important; position: static !important; width: 32px; height: 32px; display: flex; align-items: center; justify-content: center; border-radius: 50%; transition: background 0.2s; }
.simple-note-dialog .el-dialog__headerbtn:hover { background: #f5f5f5; }
.simple-note-dialog .el-dialog__headerbtn .el-dialog__close { color: #909399 !important; font-size: 20px !important; }
.simple-note-dialog .el-dialog__body { flex: 1 !important; height: 0 !important; min-height: 0 !important; padding: 10px 12px !important; display: flex !important; flex-direction: column !important; overflow: hidden !important; background-color: #884fdd11; }
.simple-note-dialog .el-dialog__footer { padding: 10px 12px 10px !important; border-top: 1px solid #f2f2f2; background: #fff; flex-shrink: 0; }
</style>
