<template>
  <div class="images-column">
    <!-- 顶部标题栏 -->
    <div class="section-header">
      <div class="section-title">
        <el-icon class="mr-1"><Picture /></el-icon> 知识点图片
      </div>
      <!-- ★★★ 权限控制：新增按钮 ★★★ -->
      <el-upload
        v-if="canEdit"
        action=""
        :http-request="handleUploadRequest"
        :show-file-list="false"
        accept=".jpg,.jpeg,.png,.gif"
      >
        <el-button type="primary" size="small" icon="Plus" class="glass-btn">新增</el-button>
      </el-upload>
    </div>

    <!-- 
      图片列表区域 
      支持 Ctrl+V 粘贴上传 (仅限有权限时)
    -->
    <div
      class="image-list-container custom-scrollbar"
      tabindex="0"
      @paste="handlePaste"
      @click="activateArea"
      :class="{ 'is-active': isActive && canEdit }"
      :title="canEdit ? '点击空白处激活后，可直接 Ctrl+V 粘贴截图' : '仅查看模式'"
      :style="{ cursor: canEdit ? 'text' : 'default' }"
    >
      <!-- 空状态提示 -->
      <div v-if="parsedImages.length === 0" class="paste-hint">
        <el-icon :size="40" class="hint-icon"><Picture /></el-icon>
        <p class="hint-text">暂无图片</p>
        <p v-if="canEdit" class="sub-hint">点击此处激活后<br />按 Ctrl+V 粘贴截图</p>
      </div>

      <!-- 图片卡片循环 -->
      <div
        v-for="(imgUrl, index) in parsedImages"
        :key="index"
        class="img-card"
      >
        <!-- 1. 卡片头部：复制链接 -->
        <div class="card-header">
          <el-tag size="small" effect="light" class="path-tag">图片 {{ index + 1 }}</el-tag>
          <el-button
            link
            type="primary"
            size="small"
            icon="CopyDocument"
            @click.stop="copyImgUrl(imgUrl)"
            class="copy-btn"
          >
            复制地址
          </el-button>
        </div>

        <!-- 2. 卡片中间：图片预览 -->
        <div class="card-body">
          <el-image
            :src="getFullImageUrl(imgUrl)"
            :preview-src-list="[getFullImageUrl(imgUrl)]"
            fit="contain"
            class="main-img"
            :initial-index="0"
            loading="lazy"
          >
            <template #error>
              <div class="image-slot">
                <el-icon><icon-picture /></el-icon>
                <span>加载失败</span>
              </div>
            </template>
          </el-image>
        </div>

        <!-- 3. 卡片底部：删除按钮 -->
        <div class="card-footer" v-if="canEdit">
          <el-button
            type="danger"
            link
            size="small"
            icon="Delete"
            @click.stop="confirmDelete(imgUrl)"
          >
            删除图片
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import {
  Delete,
  Plus,
  CopyDocument,
  Picture as IconPicture,
  Picture,
} from "@element-plus/icons-vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { uploadImage, deletePointImage, updatePoint } from "../api/point";

// --- Props & Emits ---
const props = defineProps<{
  pointId: number;
  imagesJson: string;
  canEdit: boolean;
}>();

const emit = defineEmits(["update"]);

// --- 状态 ---
const isActive = ref(false);

// 解析 JSON
const parsedImages = computed(() => {
  try {
    return JSON.parse(props.imagesJson) || [];
  } catch {
    return [];
  }
});

// 获取完整 URL
const getFullImageUrl = (path: string) => `http://localhost:8080${path}`;

// 激活粘贴区域样式
const activateArea = () => {
  if (props.canEdit) {
    isActive.value = true;
  }
};

// --- 功能 1: 复制图片地址 ---
const copyImgUrl = async (path: string) => {
  const fullUrl = getFullImageUrl(path);
  try {
    await navigator.clipboard.writeText(fullUrl);
    ElMessage.success({
      message: "图片地址已复制！可去左侧粘贴",
      duration: 2000,
    });
  } catch (err) {
    ElMessage.error("复制失败，请手动复制");
  }
};

// --- 功能 2: 删除确认 ---
const confirmDelete = (path: string) => {
  if (!props.canEdit) return;
  
  ElMessageBox.confirm(
    "确定要永久删除这张图片吗？此操作不可恢复。",
    "删除确认",
    {
      confirmButtonText: "确定删除",
      cancelButtonText: "取消",
      type: "warning",
    }
  )
    .then(() => {
      handleDelete(path);
    })
    .catch(() => {});
};

// 执行删除逻辑
const handleDelete = async (path: string) => {
  try {
    await deletePointImage(props.pointId, path);
    const newImages = parsedImages.value.filter((i: string) => i !== path);
    const jsonStr = JSON.stringify(newImages);
    emit("update", jsonStr);
    ElMessage.success("删除成功");
  } catch (e) {
    ElMessage.error("删除失败");
  }
};

// --- 功能 3: 粘贴上传 & 按钮上传 ---
const handlePaste = async (event: ClipboardEvent) => {
  if (!props.canEdit) return;

  const items = event.clipboardData?.items;
  if (!items) return;
  let file: File | null = null;
  for (let i = 0; i < items.length; i++) {
    if (items[i].type.indexOf("image") !== -1) {
      file = items[i].getAsFile();
      break;
    }
  }
  if (file) await doUpload(file);
};

const handleUploadRequest = (options: any) => {
  doUpload(options.file);
};

const doUpload = async (file: File) => {
  try {
    const res = await uploadImage(file);
    if (res.data.code === 200) {
      const newPath = res.data.data.path;
      const newImages = [...parsedImages.value, newPath];
      const jsonStr = JSON.stringify(newImages);
      await updatePoint(props.pointId, { localImageNames: jsonStr });
      emit("update", jsonStr);
      ElMessage.success("上传成功");
    }
  } catch (e) {
    ElMessage.error("上传失败");
  }
};
</script>

<style scoped>
/* 最外层容器：透明，显示父级紫色 */
.images-column {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 280px;
  height: 100%; 
  overflow: hidden; 
  background-color: transparent; 
}

/* 顶部标题栏 */
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  flex-shrink: 0; 
  padding: 0 5px;
}
.section-title {
  font-weight: bold;
  color: #303133; /* 深色文字 */
  font-size: 16px;
  display: flex;
  align-items: center;
}
.mr-1 { margin-right: 5px; }

/* 玻璃质感按钮 */
.glass-btn {
  background: linear-gradient(90deg, #667eea, #764ba2);
  border: none;
  box-shadow: 0 2px 6px rgba(118, 75, 162, 0.3);
}
.glass-btn:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}

/* 图片列表容器 */
.image-list-container {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  padding: 10px;
  
  /* 边框改为虚线，颜色淡化 */
  border: 2px dashed rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  outline: none;
  transition: all 0.3s;
  display: flex;
  flex-direction: column;
  gap: 20px;
  
  /* 背景透明 */
  background-color: rgba(255, 255, 255, 0.3); 
}

/* 激活状态样式 (紫色边框) */
.image-list-container:focus,
.image-list-container.is-active {
  border-color: #764ba2;
  background-color: rgba(255, 255, 255, 0.5);
}

/* 空状态 */
.paste-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px; 
  color: rgba(0, 0, 0, 0.4); /* 半透明深色文字 */
  margin-top: 20px;
  flex-shrink: 0;
}
.hint-icon { color: rgba(0, 0, 0, 0.2); margin-bottom: 10px; }
.hint-text { font-size: 14px; font-weight: bold; color: rgba(0, 0, 0, 0.5); }
.sub-hint {
  font-size: 12px;
  text-align: center;
  margin-top: 5px;
  color: rgba(0, 0, 0, 0.4);
}

/* --- 图片卡片样式 (悬浮白卡) --- */
.img-card {
  /* ★★★ 关键：白色微透卡片 ★★★ */
  background: rgba(255, 255, 255, 0.9);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.6);
  
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: transform 0.2s;
  flex-shrink: 0; 
}

.img-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(118, 75, 162, 0.15); /* 紫色阴影 */
  border-color: #d3adf7;
}

.card-header {
  padding: 8px 12px;
  background: rgba(249, 250, 251, 0.8);
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.path-tag { 
  background-color: rgba(118, 75, 162, 0.1); 
  border-color: rgba(118, 75, 162, 0.2); 
  color: #764ba2; 
}

.copy-btn { color: #764ba2; }
.copy-btn:hover { color: #5b3a85; }

.card-body {
  height: 200px;
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 10px;
  cursor: zoom-in;
}

.main-img {
  width: 100%;
  height: 100%;
  display: block;
  border-radius: 4px;
}

.card-footer {
  padding: 8px 12px;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
  display: flex;
  justify-content: flex-end;
  background: rgba(249, 250, 251, 0.8);
}

.image-slot {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.05);
  color: #909399;
  font-size: 14px;
  border-radius: 4px;
}

/* 滚动条美化 */
.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(0, 0, 0, 0.1); border-radius: 4px; }
.custom-scrollbar::-webkit-scrollbar-thumb:hover { background: rgba(0, 0, 0, 0.2); }
</style>