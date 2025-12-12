<template>
  <div class="rich-text-editor">
    <!-- 工具栏独立容器 - 如果使用外部工具栏，就隐藏这个 -->
    <div class="editor-toolbar" ref="toolbarContainer" v-show="!externalToolbar"></div>
    <!-- 编辑内容 -->
    <div class="editor-content" ref="editorContainer"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, markRaw } from 'vue';
import { uploadImage } from '../api/point';
import { getResourceUrl } from '../utils/oss';
import { ElMessage } from 'element-plus';

// 使用 CDN 的 CKEditor
declare global {
  interface Window {
    DecoupledEditor: any;
  }
}

const DecoupledEditor = window.DecoupledEditor;

const props = defineProps<{
  modelValue: string;
  pointId: number;
  externalToolbar?: boolean; // 是否使用外部工具栏
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void;
  (e: 'ready'): void; // 编辑器初始化完成
}>();

const toolbarContainer = ref<HTMLElement | null>(null);
const editorContainer = ref<HTMLElement | null>(null);
const editorInstance = ref<any>(null);
const localContent = ref(props.modelValue || '');

// 初始化编辑器
onMounted(async () => {
  if (!editorContainer.value || !toolbarContainer.value) return;
  
  try {
    const editor = await DecoupledEditor.create(editorContainer.value, {
      language: 'zh-cn',
      toolbar: {
        items: [
          'heading',
          '|',
          'fontSize',
          'fontFamily',
          'fontColor',
          'fontBackgroundColor',
          '|',
          'bold',
          'italic',
          'underline',
          'strikethrough',
          '|',
          'alignment',
          '|',
          'bulletedList',
          'numberedList',
          '|',
          'outdent',
          'indent',
          '|',
          'link',
          'imageUpload',
          'blockQuote',
          'insertTable',
          '|',
          'undo',
          'redo'
        ]
      },
      image: {
        toolbar: [
          'imageTextAlternative',
          '|',
          'imageStyle:inline',
          'imageStyle:block',
          'imageStyle:side'
        ]
      },
      table: {
        contentToolbar: [
          'tableColumn',
          'tableRow',
          'mergeTableCells',
          'tableCellProperties',
          'tableProperties'
        ]
      },
      heading: {
        options: [
          { model: 'paragraph', title: '段落', class: 'ck-heading_paragraph' },
          { model: 'heading1', view: 'h1', title: '标题1', class: 'ck-heading_heading1' },
          { model: 'heading2', view: 'h2', title: '标题2', class: 'ck-heading_heading2' },
          { model: 'heading3', view: 'h3', title: '标题3', class: 'ck-heading_heading3' }
        ]
      },
      fontSize: {
        options: [
          10, 11, 12, 13, 14, 15, 16, 18, 20, 22, 24, 26, 28, 32, 36, 40, 48, 56, 64, 72
        ],
        supportAllValues: true
      },
      fontFamily: {
        supportAllValues: true
      }
    });
    
    // 关键：把工具栏插入到独立的容器
    toolbarContainer.value.appendChild(editor.ui.view.toolbar.element!);
    
    // 使用 markRaw 避免 Vue 响应式代理导致的错误
    editorInstance.value = markRaw(editor);
    
    // 设置初始内容
    if (props.modelValue) {
      editor.setData(props.modelValue);
    }
    
    // 监听内容变化
    editor.model.document.on('change:data', () => {
      const data = editor.getData();
      localContent.value = data;
      emit('update:modelValue', data);
    });
    
    // 监听选区变化，显示实际字号
    editor.model.document.selection.on('change', () => {
      const selection = editor.model.document.selection;
      const fontSize = selection.getAttribute('fontSize');
      
      // 如果没有设置字号属性，尝试获取实际渲染的字号
      if (!fontSize) {
        // 获取当前光标位置的 DOM 元素
        const domSelection = window.getSelection();
        if (domSelection && domSelection.rangeCount > 0) {
          const range = domSelection.getRangeAt(0);
          let element: Node | null = range.startContainer;
          
          // 如果是文本节点，获取其父元素
          if (element.nodeType === Node.TEXT_NODE) {
            element = element.parentElement;
          }
          
          // 获取计算后的字号
          if (element && element instanceof HTMLElement) {
            const computedStyle = window.getComputedStyle(element);
            const actualFontSize = computedStyle.fontSize;
            console.log('当前选中文字的字号:', fontSize || '未设置', '实际显示字号:', actualFontSize);
          }
        }
      } else {
        console.log('当前选中文字的字号:', fontSize);
      }
    });
    
    // 自定义图片上传
    editor.plugins.get('FileRepository').createUploadAdapter = (loader: any) => {
      return {
        upload: async () => {
          const file = await loader.file;
          try {
            const loadingMsg = ElMessage.info({ message: '图片上传中...', duration: 0 });
            const res = await uploadImage(file, props.pointId);
            loadingMsg.close();
            
            if (res.data.code === 200) {
              const url = getResourceUrl(res.data.data.path);
              ElMessage.success('图片上传成功');
              return { default: url };
            } else {
              ElMessage.error('图片上传失败');
              throw new Error('上传失败');
            }
          } catch (err) {
            console.error('图片上传错误:', err);
            ElMessage.error('图片上传失败');
            throw err;
          }
        }
      };
    };
    
    // 编辑器初始化完成，通知父组件
    emit('ready');
  } catch (err) {
    console.error('CKEditor 初始化失败:', err);
  }
});

// 监听外部变化
watch(() => props.modelValue, (newVal) => {
  if (newVal !== localContent.value && editorInstance.value) {
    editorInstance.value.setData(newVal);
    localContent.value = newVal;
  }
});

// 销毁编辑器
onBeforeUnmount(() => {
  if (editorInstance.value) {
    try {
      editorInstance.value.destroy().catch(() => {
        // 忽略销毁错误
      });
    } catch (e) {
      // 忽略错误
    }
    editorInstance.value = null;
  }
});

// 暴露方法给父组件
const insertCustomDivider = () => {
  if (!editorInstance.value) return;
  
  const dividerHtml = `
    <hr style="margin: 0; border: none; border-top: 1px solid #ccc;"/>
    <p style="text-align: center; margin: 0; padding: 8px 0; background-color: #67c23a; color: #ffffff; font-size: 15px; font-weight: bold; letter-spacing: 2px; line-height: 1;">
      ---------------------------------------------------------- 我是分割线 ---------------------------------------------------------
    </p>
    <hr style="margin: 0; border: none; border-top: 1px solid #ccc;"/>
  `;
  
  const viewFragment = editorInstance.value.data.processor.toView(dividerHtml);
  const modelFragment = editorInstance.value.data.toModel(viewFragment);
  editorInstance.value.model.change((writer: any) => {
    editorInstance.value.model.insertContent(modelFragment, editorInstance.value.model.document.selection);
  });
  
  ElMessage.success('已插入分割线');
};

defineExpose({
  insertCustomDivider,
  getEditor: () => editorInstance.value,
});
</script>

<style scoped>
.rich-text-editor {
  display: flex;
  flex-direction: column;
  background: #fff;
}

/* 工具栏独立容器 - 固定在顶部 */
.editor-toolbar {
  position: sticky;
  top: 0;
  z-index: 100;
  background: #fff;
  border: 1px solid #e8e8e8;
  border-bottom: 2px solid #e8e8e8;
  border-radius: 4px 4px 0 0;
  /* 当工具栏被移走后，隐藏这个容器 */
  min-height: 0;
}

/* 编辑内容区 */
.editor-content {
  flex: 1;
  border: 1px solid #e8e8e8;
  border-top: 1px solid #e8e8e8; /* 改为1px，因为工具栏已经移走 */
  border-radius: 4px; /* 四个角都是圆角 */
  min-height: 500px;
  overflow: visible;
}

:deep(.ck-editor) {
  height: auto;
  overflow: visible;
}

:deep(.ck-editor__main) {
  height: auto;
  overflow: visible;
}

:deep(.ck-editor__editable_inline) {
  min-height: 500px !important;
  border: none !important;
  box-shadow: none !important;
  overflow: visible !important;
  max-height: none !important;
}

:deep(.ck-editor__editable) {
  min-height: 500px !important;
  border: none !important;
  box-shadow: none !important;
  overflow: visible !important;
  max-height: none !important;
}

:deep(.ck-content img) {
  max-width: 100%;
  height: auto;
}

/* 隐藏浮动工具栏 */
:deep(.ck-balloon-panel) {
  display: none !important;
}
</style>

<!-- 全局样式 - 修复字号下拉菜单 -->
<style>
/* 修复字号下拉菜单显示问题 - 强制所有字号选项以14px显示 */
/* 使用全局样式确保优先级足够高 */
.ck.ck-dropdown__panel .ck-list__item,
.ck.ck-dropdown__panel .ck-list__item *,
.ck.ck-dropdown__panel .ck-button__label {
  font-size: 14px !important;
  line-height: 1.5 !important;
}

.ck.ck-dropdown__panel {
  font-size: 14px !important;
}

/* 特别针对字号选项 */
.ck.ck-fontsize-option,
.ck.ck-fontsize-option * {
  font-size: 14px !important;
}
</style>
