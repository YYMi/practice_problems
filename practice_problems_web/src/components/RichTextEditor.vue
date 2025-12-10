<template>
  <div class="rich-text-editor">
    <div class="editor-toolbar">
      <Toolbar
        :editor="editorRef"
        :defaultConfig="toolbarConfig"
        mode="default"
      />
    </div>
    <div class="editor-container">
      <Editor
        v-model="localContent"
        :defaultConfig="editorConfig"
        mode="default"
        @onCreated="handleCreated"
        @onChange="handleChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onBeforeUnmount } from 'vue';
import { Editor, Toolbar } from '@wangeditor/editor-for-vue';
import '@wangeditor/editor/dist/css/style.css';
import { uploadImage } from '../api/point';
import { getResourceUrl } from '../utils/oss';
import { ElMessage } from 'element-plus';

const props = defineProps<{
  modelValue: string;
  pointId: number;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void;
}>();

const editorRef = ref<any>(null);
const localContent = ref(props.modelValue || '');

// 监听外部变化
watch(() => props.modelValue, (newVal) => {
  if (newVal !== localContent.value) {
    localContent.value = newVal;
    if (editorRef.value) {
      editorRef.value.setHtml(newVal);
    }
  }
});

const toolbarConfig = {};

const editorConfig = {
  placeholder: "请输入内容...",
  // 粘贴配置：完全保留原始格式
  PASTE_CONF: {
    pasteCode: false,
    pasteText: false,
    pasteHtml: true,
    pasteIgnoreImg: false,
    filterStyle: false,
  },
  hoverbarKeys: {
    text: {
      menuKeys: [],
    },
  },
  MENU_CONF: {
    codeSelectLang: {
      codeLangs: [
        { text: 'Java', value: 'java' },
        { text: 'JavaScript', value: 'javascript' },
        { text: 'Python', value: 'python' },
        { text: 'Go', value: 'go' },
        { text: 'HTML', value: 'html' },
        { text: 'CSS', value: 'css' },
        { text: 'SQL', value: 'sql' },
      ],
    },
    uploadImage: {
      async customUpload(file: File, insertFn: any) {
        try {
          const res = await uploadImage(file, props.pointId);
          if (res.data.code === 200) {
            const url = getResourceUrl(res.data.data.path);
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
  if (localContent.value) {
    editor.setHtml(localContent.value);
  }
  
  // 完全接管粘贴事件，保留所有原始格式 + 处理图片上传
  setTimeout(() => {
    const editorContainer = document.querySelector('.w-e-text-container');
    if (editorContainer) {
      editorContainer.addEventListener('paste', async (e: any) => {
        const clipboardData = e.clipboardData;
        if (!clipboardData) return;
        
        // 1. 优先检查是否有图片文件
        const items = clipboardData.items;
        let hasImage = false;
        
        if (items) {
          for (let i = 0; i < items.length; i++) {
            if (items[i].type.indexOf('image') !== -1) {
              e.preventDefault(); // 阻止默认粘贴
              hasImage = true;
              
              const file = items[i].getAsFile();
              if (file) {
                try {
                  const loadingMsg = ElMessage.info({ message: '图片上传中...', duration: 0 });
                  const res = await uploadImage(file, props.pointId);
                  loadingMsg.close();
                  if (res.data.code === 200) {
                    const url = getResourceUrl(res.data.data.path);
                    editor.dangerouslyInsertHtml(`<p><img src="${url}" alt="" style="max-width: 100%;"/></p>`);
                    ElMessage.success('图片上传成功');
                  } else {
                    ElMessage.error('图片上传失败');
                  }
                } catch (err) {
                  console.error('图片上传错误:', err);
                  ElMessage.error('图片上传失败');
                }
              }
              break;
            }
          }
        }
        
        // 2. 如果没有图片，处理HTML/文本
        if (!hasImage) {
          const html = clipboardData.getData('text/html');
          
          if (html) {
            // 检查HTML中是否包含base64图片
            if (html.includes('data:image')) {
              e.preventDefault();
              ElMessage.warning('检测到base64图片，建议使用截图工具直接粘贴或上传图片');
              return;
            }
            // 不阻止默认行为，让编辑器完整保留HTML格式（包括表格、列表、缩进等）
            return;
          }
          
          // 纯文本粘贴：保留空格和缩进
          const text = clipboardData.getData('text/plain');
          if (text) {
            e.preventDefault();
            const lines = text.split('\n');
            const htmlContent = lines.map(line => {
              if (line.trim() === '') {
                return '<p><br></p>';
              }
              // 将空格转换为&nbsp;以保留缩进
              const processedLine = line.replace(/^(\s+)/, (match) => {
                return '&nbsp;'.repeat(match.length);
              }).replace(/  /g, '&nbsp;&nbsp;');
              return `<p>${processedLine}</p>`;
            }).join('');
            editor.dangerouslyInsertHtml(htmlContent);
          }
        }
      }, true);
    }
  }, 500);
};

const handleChange = (editor: any) => {
  const html = editor.getHtml();
  localContent.value = html;
  emit('update:modelValue', html);
};

// 插入自定义分割线
const insertCustomDivider = () => {
  if (!editorRef.value) return;
  
  // 先尝试删除光标所在的空段落
  try {
    const editor = editorRef.value;
    const html = editor.getHtml();
    
    // 使用编辑器的 focus 确保有正确的选区
    editor.focus();
    
    // 获取当前选区
    const selection = editor.selection;
    if (selection) {
      // 获取当前节点
      const node = selection.getNode();
      if (node) {
        // 检查是否是空的 p 标签
        const parentP = node.closest?.('p') || (node.nodeName === 'P' ? node : null);
        if (parentP && parentP.textContent?.trim() === '') {
          // 删除空的 p 标签
          parentP.remove();
        }
      }
    }
  } catch (e) {
    console.log('清理空段落失败:', e);
  }
  
  const dividerHtml = `<hr style="margin: 0; border: none; border-top: 1px solid #ccc;"/><p style="text-align: center; margin: 0; padding: 8px 0; background-color: #67c23a; color: #ff0000; font-size: 15px; font-weight: bold; letter-spacing: 2px; line-height: 1;">---------------------------------------------------------- 我是分割线 ---------------------------------------------------------</p><hr style="margin: 0; border: none; border-top: 1px solid #ccc;"/>`;
  
  editorRef.value.dangerouslyInsertHtml(dividerHtml);
  ElMessage.success('已插入分割线');
};

// 暴露方法给父组件
defineExpose({
  insertCustomDivider,
  getEditor: () => editorRef.value,
});

onBeforeUnmount(() => {
  if (editorRef.value) {
    editorRef.value.destroy();
  }
});
</script>

<style scoped>
.rich-text-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  border: 1px solid #e8e8e8;
  border-radius: 4px;
  overflow: hidden;
}

.editor-toolbar {
  border-bottom: 1px solid #e8e8e8;
  background: #fff;
}

.editor-container {
  flex: 1;
  overflow-y: auto;
}

:deep(.w-e-text-container) {
  background-color: #fff !important;
}

:deep(.w-e-text-container [data-slate-editor]) {
  min-height: 400px;
  padding: 15px;
}
</style>
