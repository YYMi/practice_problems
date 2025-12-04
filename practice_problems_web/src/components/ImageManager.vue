<template>
  <div class="images-container-wrapper">
    <!-- 顶部标题栏 -->
    <div class="section-header">
      <div class="section-title">
        <el-icon class="mr-1"><Picture /></el-icon> 知识点图片
      </div>
      
      <!-- 只有上传按钮 -->
      <el-upload
        v-if="canEdit"
        action=""
        :http-request="handleUploadRequest"
        :before-upload="beforeUpload"
        :show-file-list="false"
        accept=".jpg,.jpeg,.png,.gif"
      >
        <el-button type="primary" size="small" icon="Plus" class="glass-btn">新增</el-button>
      </el-upload>
    </div>

    <!-- 核心内容区域 -->
    <div class="content-area">
      
      <!-- 左侧侧边栏 -->
      <div class="left-sidebar" v-if="localImageList.length > 0">
        
        <!-- 设置区域 (前后缀一行显示) -->
        <div v-if="canEdit" class="sidebar-settings">
           <el-tooltip content="复制前缀 (如：【 )" placement="top" :show-after="500">
             <el-input v-model="copyPrefix" size="small" class="tiny-input" placeholder="前" />
           </el-tooltip>
           <span class="sep">-</span>
           <el-tooltip content="复制后缀 (如：】 )" placement="top" :show-after="500">
             <el-input v-model="copySuffix" size="small" class="tiny-input" placeholder="后" />
           </el-tooltip>
        </div>

        <!-- 导航列表 (添加 ref 用于自动滚动) -->
        <div class="nav-list custom-scrollbar" ref="navListRef">
          <div 
            v-for="(imgItem, index) in localImageList" 
            :key="imgItem.url"
            class="nav-item"
            :class="{ 'is-active': activeIndex === index }"
            @click="handleNavClick(index, imgItem.name)"
            :title="canEdit ? '点击定位并复制文件名' : '点击定位'"
          >
            <span class="nav-text">{{ truncateName(imgItem.name, 4) }}</span>
            <el-icon v-if="canEdit" class="copy-icon"><CopyDocument /></el-icon>
          </div>
        </div>
      </div>

      <!-- 右侧：拖拽列表区域 -->
      <div
        ref="scrollContainerRef"
        class="image-list-container custom-scrollbar"
        :class="{ 'is-active': isActive && canEdit }"
        @click="activateArea"
        @paste="handlePaste"
        tabindex="0"
      >
        <!-- 空状态 -->
        <div v-if="localImageList.length === 0" class="paste-hint">
          <el-icon :size="40" class="hint-icon"><Picture /></el-icon>
          <p class="hint-text">暂无图片</p>
          <p v-if="canEdit" class="sub-hint">点击激活后 Ctrl+V 粘贴</p>
        </div>

        <!-- 拖拽列表 -->
        <VueDraggable
          v-model="localImageList"
          item-key="url"
          :animation="200"
          :disabled="!canEdit"
          handle=".drag-handle" 
          @end="onDragEnd" 
          class="draggable-list"
        >
          <template #item="{ element: imgItem, index }">
            <div
              :ref="(el) => setImageCardRef(el, index)"
              class="img-card"
              :class="{ 'highlight-card': activeIndex === index }"
              @mouseenter="handleMouseEnter(index)"
            >
              <div class="card-header drag-handle" :style="{ cursor: canEdit ? 'move' : 'default' }">
                <div class="name-tag-group">
                  <el-tag size="small" effect="light" class="path-tag">
                    {{ truncateName(imgItem.name) }}
                  </el-tag>
                </div>
                <el-icon v-if="canEdit" class="move-icon"><Rank /></el-icon>
                
                <div class="header-btns">
                   <el-button
                    v-if="canEdit"
                    link
                    type="primary"
                    size="small"
                    icon="Edit"
                    @click.stop="handleRename(index, imgItem.name)"
                  />
                  <el-button
                    v-if="canEdit"
                    type="danger"
                    link
                    size="small"
                    icon="Delete"
                    @click.stop="confirmDelete(index)"
                  />
                </div>
              </div>

              <div class="card-body">
                <el-image
                  :src="getFullImageUrl(imgItem.url)"
                  :preview-src-list="allImageUrls" 
                  :initial-index="index"
                  :preview-teleported="true"
                  :z-index="9999"
                  fit="contain"
                  class="main-img"
                  loading="lazy"
                  @click.stop
                >
                  <template #error>
                    <div class="image-slot">
                      <el-icon><icon-picture /></el-icon>
                    </div>
                  </template>
                </el-image>
              </div>
            </div>
          </template>
        </VueDraggable>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from "vue";
import {
  Delete, Plus, Edit, Picture as IconPicture, Picture, CopyDocument, Rank 
} from "@element-plus/icons-vue";
import { ElMessage, ElMessageBox, type UploadProps } from "element-plus";
import { uploadImage, deletePointImage, updatePoint } from "../api/point";
import useClipboard from 'vue-clipboard3';
import VueDraggable from 'vuedraggable'; 

// --- Props & Emits ---
const props = defineProps<{
  pointId: number;
  imagesJson: string;
  canEdit: boolean;
}>();

const emit = defineEmits(["update"]);

// --- 状态变量 ---
const isActive = ref(false);
const scrollContainerRef = ref<HTMLElement | null>(null);
const navListRef = ref<HTMLElement | null>(null); // 左侧列表的 DOM 引用
const imageCardRefs = ref<HTMLElement[]>([]);
const activeIndex = ref<number>(-1);
const copyPrefix = ref("【");
const copySuffix = ref("】");
const localImageList = ref<ImageItem[]>([]); // 本地数据
const { toClipboard } = useClipboard();

interface ImageItem {
  name: string;
  url: string;
}

// --- 初始化与数据同步 ---
const parseImages = (json: string): ImageItem[] => {
  try {
    const raw = JSON.parse(json);
    if (!Array.isArray(raw)) return [];
    return raw.map((item: string | ImageItem, index: number) => {
      if (typeof item === 'string') {
        return { name: `图片 ${index + 1}`, url: item };
      }
      return item;
    });
  } catch {
    return [];
  }
};

localImageList.value = parseImages(props.imagesJson);

watch(() => props.imagesJson, (newVal) => {
  const parsed = parseImages(newVal);
  if (JSON.stringify(parsed) !== JSON.stringify(localImageList.value)) {
     localImageList.value = parsed;
  }
});

// --- 交互逻辑 ---

// 鼠标滑入右侧图片，左侧联动
const handleMouseEnter = (index: number) => {
  activeIndex.value = index;
  // 滚动左侧导航
  if (navListRef.value) {
    const navItems = navListRef.value.querySelectorAll('.nav-item');
    if (navItems[index]) {
      navItems[index].scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }
  }
};

// 点击左侧导航，右侧定位 + 复制
const handleNavClick = async (index: number, name: string) => {
  activeIndex.value = index;
  const targetCard = imageCardRefs.value[index];
  if (targetCard && scrollContainerRef.value) {
    targetCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }
  if (props.canEdit) {
    const textToCopy = `${copyPrefix.value}${name}${copySuffix.value}`;
    try {
      await toClipboard(textToCopy);
      ElMessage.success(`已复制: ${textToCopy}`);
    } catch (e) {}
  }
};

// 拖拽结束
const onDragEnd = async () => {
  await saveImages(localImageList.value);
  imageCardRefs.value = []; // 清空缓存，防止定位偏移
};

// --- 数据操作 (CRUD) ---

const doUpload = async (file: File) => {
  try {
    const res = await uploadImage(file, props.pointId);
    if (res.data.code === 200) {
      const newName = `图${localImageList.value.length + 1}`; // 自动命名逻辑
      localImageList.value.push({ name: newName, url: res.data.data.path });
      await saveImages(localImageList.value);
      ElMessage.success("上传成功");
      // 滚动到底部
      nextTick(() => { if (scrollContainerRef.value) scrollContainerRef.value.scrollTop = scrollContainerRef.value.scrollHeight; });
    }
  } catch (e) {}
};

const handleRename = (index: number, oldName: string) => {
  ElMessageBox.prompt('请输入新名称', '重命名', { inputValue: oldName, inputPattern: /\S/ }).then(async ({ value }) => {
    localImageList.value[index].name = value;
    await saveImages(localImageList.value);
    ElMessage.success('重命名成功');
  }).catch(() => {});
};

const confirmDelete = (index: number) => {
  if (!props.canEdit) return;
  ElMessageBox.confirm("确定删除吗？", "提示", { type: "warning" }).then(() => handleDelete(index)).catch(() => {});
};

const handleDelete = async (index: number) => {
  try {
    await deletePointImage(props.pointId, localImageList.value[index].url);
    localImageList.value.splice(index, 1);
    await saveImages(localImageList.value);
    ElMessage.success("删除成功");
  } catch (e) {}
};

const saveImages = async (list: ImageItem[]) => {
  try {
    const jsonStr = JSON.stringify(list);
    await updatePoint(props.pointId, { localImageNames: jsonStr });
    emit("update", jsonStr);
  } catch (e) {
    console.error("保存失败", e);
  }
};

// --- 辅助函数 ---
const imgBaseUrl = import.meta.env.VITE_IMG_BASE_URL || ''; 
const getFullImageUrl = (path: string) => {
  if (!path) return '';
  if (path.startsWith('http') || path.startsWith('//')) return path;
  const safePath = path.startsWith('/') ? path : `/${path}`;
  return `${imgBaseUrl}${safePath}`;
};
const truncateName = (name: string, len: number = 8) => {
  if (!name) return '';
  return name.length > len ? name.substring(0, len) + '...' : name;
};
const activateArea = () => { if (props.canEdit) isActive.value = true; };
const setImageCardRef = (el: any, index: number) => { if (el) imageCardRefs.value[index] = el; };

const beforeUpload: UploadProps['beforeUpload'] = (rawFile) => {
  return rawFile.size / 1024 / 1024 < 10 || (ElMessage.error('大小不能超过 10MB!'), false);
};
const handleUploadRequest = (options: any) => doUpload(options.file);
const handlePaste = async (event: ClipboardEvent) => {
  if (!props.canEdit) return;
  const items = event.clipboardData?.items;
  if (items) {
    for (let i = 0; i < items.length; i++) {
      if (items[i].type.indexOf("image") !== -1) {
        await doUpload(items[i].getAsFile()!);
        break;
      }
    }
  }
};
const allImageUrls = computed(() => localImageList.value.map(item => getFullImageUrl(item.url)));
</script>

<style scoped>
.images-container-wrapper {
  flex: 1; display: flex; flex-direction: column; min-width: 300px; height: 100%; overflow: hidden; background-color: transparent; 
}
.section-header {
  display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; flex-shrink: 0; padding: 0 5px;
}
.section-title { font-weight: bold; color: #303133; font-size: 16px; display: flex; align-items: center; }
.mr-1 { margin-right: 5px; }
.glass-btn { background: linear-gradient(90deg, #667eea, #764ba2); border: none; box-shadow: 0 2px 6px rgba(118, 75, 162, 0.3); color: white; }
.glass-btn:hover { opacity: 0.9; transform: translateY(-1px); }

.content-area { flex: 1; display: flex; overflow: hidden; gap: 8px; }

/* --- 左侧侧边栏 --- */
.left-sidebar {
  width: 65px; /* 保持窄宽度 */
  flex-shrink: 0;
  display: flex; flex-direction: column;
  gap: 5px;
}

/* 设置区域 */
.sidebar-settings {
  background: rgba(255,255,255,0.8);
  padding: 4px;
  border-radius: 6px;
  border: 1px solid rgba(0,0,0,0.05);
  display: flex; flex-direction: row; align-items: center; justify-content: space-between;
  height: 32px;
}
.tiny-input { width: 24px; font-size: 12px; padding: 0; }
:deep(.tiny-input .el-input__wrapper) { padding: 0 2px !important; box-shadow: none !important; background: transparent; }
:deep(.tiny-input .el-input__inner) { text-align: center; height: 24px; line-height: 24px; }
.sep { color: #909399; font-size: 12px; margin: 0 1px; }

/* 导航列表 */
.nav-list {
  flex: 1; overflow-y: auto;
  background: rgba(255, 255, 255, 0.5);
  border-radius: 6px; padding: 4px;
  display: flex; flex-direction: column; gap: 4px;
  border: 1px solid rgba(0,0,0,0.05);
}
.nav-item {
  font-size: 11px; padding: 4px; border-radius: 4px; cursor: pointer;
  color: #606266; transition: all 0.2s;
  display: flex; justify-content: center; align-items: center;
  background: rgba(255,255,255,0.6); position: relative;
}
.nav-item:hover { background: rgba(118, 75, 162, 0.1); color: #764ba2; }
.nav-item.is-active { background: #764ba2; color: white; }
.nav-text { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.copy-icon { position: absolute; right: 2px; top: 2px; font-size: 8px; opacity: 0.5; }

/* 右侧列表 */
.image-list-container {
  flex: 1; overflow-y: auto; min-height: 0; padding: 10px;
  border: 2px dashed rgba(0, 0, 0, 0.1); border-radius: 8px;
  background-color: rgba(255, 255, 255, 0.3); 
}
.image-list-container:focus, .image-list-container.is-active { border-color: #764ba2; background-color: rgba(255, 255, 255, 0.5); }
.draggable-list { display: flex; flex-direction: column; gap: 15px; padding-bottom: 20px; }

/* 卡片样式 */
.img-card {
  background: rgba(255, 255, 255, 0.9); border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05); border: 1px solid rgba(255, 255, 255, 0.6);
  display: flex; flex-direction: column; overflow: hidden; 
  transition: all 0.3s ease; /* 平滑过渡 */
}
.img-card:hover { box-shadow: 0 8px 20px rgba(118, 75, 162, 0.15); }

/* 高亮呼吸灯效果 */
.highlight-card {
  border-color: #764ba2 !important;
  border-width: 2px !important;
  background-color: rgba(118, 75, 162, 0.05) !important;
  box-shadow: 0 0 15px rgba(118, 75, 162, 0.4) !important;
  transform: scale(1.02); z-index: 1;
  animation: pulse-purple 2s infinite;
}
@keyframes pulse-purple {
  0% { box-shadow: 0 0 0 0 rgba(118, 75, 162, 0.4); }
  70% { box-shadow: 0 0 0 10px rgba(118, 75, 162, 0); }
  100% { box-shadow: 0 0 0 0 rgba(118, 75, 162, 0); }
}

.card-header {
  padding: 5px 8px; background: rgba(249, 250, 251, 0.9); border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  display: flex; justify-content: space-between; align-items: center;
  user-select: none; 
}
.card-header:active { cursor: grabbing !important; }
.move-icon { color: #909399; margin-left: auto; margin-right: 10px; cursor: move; }
.path-tag { max-width: 100px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.header-btns { display: flex; gap: 0; }
.card-body { height: 180px; background: transparent; display: flex; align-items: center; justify-content: center; padding: 10px; }
.main-img { width: 100%; height: 100%; display: block; border-radius: 4px; cursor: zoom-in; }
.image-slot { display: flex; flex-direction: column; justify-content: center; align-items: center; width: 100%; height: 100%; background: rgba(0, 0, 0, 0.05); color: #909399; font-size: 14px; border-radius: 4px; }
.paste-hint { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 200px; color: rgba(0, 0, 0, 0.4); margin-top: 20px; flex-shrink: 0; }
.hint-icon { color: rgba(0, 0, 0, 0.2); margin-bottom: 10px; }
.hint-text { font-size: 14px; font-weight: bold; color: rgba(0, 0, 0, 0.5); }
.sub-hint { font-size: 12px; text-align: center; margin-top: 5px; color: rgba(0, 0, 0, 0.4); }
.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(0, 0, 0, 0.1); border-radius: 4px; }
</style>
<style>
.el-image-viewer__wrapper { z-index: 20000 !important; }
.el-image-viewer__btn { z-index: 20001 !important; }
</style>