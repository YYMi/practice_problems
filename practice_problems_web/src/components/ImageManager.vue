<template>
  <div class="images-column">
    <!-- 顶部标题栏 -->
    <div class="section-header">
      <div class="section-title">
        <el-icon class="mr-1"><Picture /></el-icon> 知识点图片
      </div>
      <!-- ★★★ 权限控制：新增按钮 (带 10MB 限制) ★★★ -->
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

    <!-- 图片列表区域 -->
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
      <div v-if="displayImages.length === 0" class="paste-hint">
        <el-icon :size="40" class="hint-icon"><Picture /></el-icon>
        <p class="hint-text">暂无图片</p>
        <p v-if="canEdit" class="sub-hint">点击此处激活后<br />按 Ctrl+V 粘贴截图</p>
      </div>

      <!-- 图片卡片循环 -->
      <div
        v-for="(imgItem, index) in displayImages"
        :key="index"
        class="img-card"
      >
        <!-- 1. 卡片头部：显示名字 & 重命名 -->
        <div class="card-header">
          <div class="name-tag-group">
            <el-tag size="small" effect="light" class="path-tag" :title="imgItem.name">
              {{ truncateName(imgItem.name) }}
            </el-tag>
          </div>
          
          <!-- ★★★ 修改：重命名按钮 (仅有权限时显示) ★★★ -->
          <el-button
            v-if="canEdit"
            link
            type="primary"
            size="small"
            icon="Edit"
            @click.stop="handleRename(index, imgItem.name)"
            class="rename-btn"
          >
            重命名
          </el-button>
        </div>

        <!-- ★★★ 修改后 ★★★ -->
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
            @click.stop="confirmDelete(index)"
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
  Edit, // 引入编辑图标
  Picture as IconPicture,
  Picture,
} from "@element-plus/icons-vue";
import { ElMessage, ElMessageBox, type UploadProps } from "element-plus";
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

// --- 数据结构定义 ---
interface ImageItem {
  name: string;
  url: string;
}

// --- 核心：解析并兼容旧数据 ---
// 以前存的是 ["/path/a.jpg"]，现在我们要转成 [{ name: "图片1", url: "/path/a.jpg" }]
const displayImages = computed<ImageItem[]>(() => {
  try {
    const raw = JSON.parse(props.imagesJson);
    if (!Array.isArray(raw)) return [];
    
    return raw.map((item: string | ImageItem, index: number) => {
      // 如果是旧格式(字符串)，转为对象
      if (typeof item === 'string') {
        return { name: `图片 ${index + 1}`, url: item };
      }
      // 如果已经是新格式(对象)，直接返回
      return item;
    });
  } catch {
    return [];
  }
});

const imgBaseUrl = import.meta.env.VITE_IMG_BASE_URL;

// 获取完整 URL
const getFullImageUrl = (path: string) => {
  if (!path) return '';
  // 如果路径已经是 http 开头的完整路径，直接返回
  if (path.startsWith('http') || path.startsWith('//')) {
    return path;
  }
  // 拼接配置的 Base URL 和图片路径
  return `${imgBaseUrl}${path}`;
};

// 截断过长的名字
const truncateName = (name: string) => {
  return name.length > 8 ? name.substring(0, 8) + '...' : name;
};

// 激活粘贴区域
const activateArea = () => {
  if (props.canEdit) isActive.value = true;
};

// --- ★★★ 限制上传大小 (10MB) ★★★ ---
const beforeUpload: UploadProps['beforeUpload'] = (rawFile) => {
  const isLt10M = rawFile.size / 1024 / 1024 < 10;
  if (!isLt10M) {
    ElMessage.error('上传失败：图片大小不能超过 10MB!');
    return false;
  }
  return true;
};

// --- ★★★ 功能 1: 重命名图片 ★★★ ---
const handleRename = (index: number, oldName: string) => {
  ElMessageBox.prompt('请输入新的图片名称', '重命名', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputValue: oldName,
    inputPattern: /\S/, // 不能为空
    inputErrorMessage: '名称不能为空'
  }).then(async ({ value }) => {
    // 1. 复制当前列表
    const newList = [...displayImages.value];
    // 2. 修改名字
    newList[index].name = value;
    // 3. 保存到后端
    await saveImages(newList);
    ElMessage.success('重命名成功');
  }).catch(() => {});
};

// --- 功能 2: 删除图片 ---
const confirmDelete = (index: number) => {
  if (!props.canEdit) return;
  ElMessageBox.confirm("确定要永久删除这张图片吗？", "删除确认", {
    confirmButtonText: "确定", cancelButtonText: "取消", type: "warning",
  }).then(() => {
    handleDelete(index);
  }).catch(() => {});
};

const handleDelete = async (index: number) => {
  try {
    const targetUrl = displayImages.value[index].url;
    // 调用后端删除文件接口
    await deletePointImage(props.pointId, targetUrl);
    
    // 从列表中移除
    const newList = [...displayImages.value];
    newList.splice(index, 1);
    
    await saveImages(newList);
    ElMessage.success("删除成功");
  } catch (e) {
  }
};

// --- 功能 3: 上传逻辑 ---
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
  if (file) {
    if (file.size / 1024 / 1024 > 10) {
      ElMessage.error('粘贴失败：图片大小不能超过 10MB!');
      return;
    }
    await doUpload(file);
  }
};

const handleUploadRequest = (options: any) => {
  doUpload(options.file);
};

const doUpload = async (file: File) => {
  try {
    const res = await uploadImage(file,props.pointId);
    if (res.data.code === 200) {
      const newPath = res.data.data.path;
      // 新增图片，默认名字为 "新图片"
      const newList = [...displayImages.value, { name: '新图片', url: newPath }];
      await saveImages(newList);
      ElMessage.success("上传成功");
    }
  } catch (e) {
   
  }
};

// --- 公共保存方法 ---
const saveImages = async (list: ImageItem[]) => {
  const jsonStr = JSON.stringify(list);
  await updatePoint(props.pointId, { localImageNames: jsonStr });
  emit("update", jsonStr);
};

const allImageUrls = computed(() => {
  return displayImages.value.map(item => getFullImageUrl(item.url));
});
</script>

<style scoped>
/* 最外层容器：透明 */
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
  color: #303133;
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
.glass-btn:hover { opacity: 0.9; transform: translateY(-1px); }

/* 图片列表容器 */
.image-list-container {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  padding: 10px;
  border: 2px dashed rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  outline: none;
  transition: all 0.3s;
  display: flex;
  flex-direction: column;
  gap: 20px;
  background-color: rgba(255, 255, 255, 0.3); 
}

.image-list-container:focus,
.image-list-container.is-active {
  border-color: #764ba2;
  background-color: rgba(255, 255, 255, 0.5);
}

/* 空状态 */
.paste-hint {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  height: 200px; color: rgba(0, 0, 0, 0.4); margin-top: 20px; flex-shrink: 0;
}
.hint-icon { color: rgba(0, 0, 0, 0.2); margin-bottom: 10px; }
.hint-text { font-size: 14px; font-weight: bold; color: rgba(0, 0, 0, 0.5); }
.sub-hint { font-size: 12px; text-align: center; margin-top: 5px; color: rgba(0, 0, 0, 0.4); }

/* --- 图片卡片样式 --- */
.img-card {
  background: rgba(255, 255, 255, 0.9);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.6);
  display: flex; flex-direction: column; overflow: hidden; transition: transform 0.2s; flex-shrink: 0; 
}

.img-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(118, 75, 162, 0.15);
  border-color: #d3adf7;
}

.card-header {
  padding: 8px 12px;
  background: rgba(249, 250, 251, 0.8);
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  display: flex; justify-content: space-between; align-items: center;
}

.path-tag { 
  background-color: rgba(118, 75, 162, 0.1); 
  border-color: rgba(118, 75, 162, 0.2); 
  color: #764ba2;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rename-btn { color: #764ba2; font-weight: bold; }
.rename-btn:hover { color: #5b3a85; }

.card-body {
  height: 200px;
  background: transparent;
  display: flex; align-items: center; justify-content: center; padding: 10px; cursor: zoom-in;
}

.main-img { width: 100%; height: 100%; display: block; border-radius: 4px; }

.card-footer {
  padding: 8px 12px;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
  display: flex; justify-content: flex-end;
  background: rgba(249, 250, 251, 0.8);
}

.image-slot {
  display: flex; flex-direction: column; justify-content: center; align-items: center;
  width: 100%; height: 100%; background: rgba(0, 0, 0, 0.05);
  color: #909399; font-size: 14px; border-radius: 4px;
}

/* 滚动条美化 */
.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(0, 0, 0, 0.1); border-radius: 4px; }
.custom-scrollbar::-webkit-scrollbar-thumb:hover { background: rgba(0, 0, 0, 0.2); }
</style>

<!-- 注意：不要加 scoped，因为查看器是渲染在 body 标签下的 -->
<style>
/* 强制提升 Element Plus 图片查看器的层级 */
.el-image-viewer__wrapper {
  z-index: 20000 !important; /* 必须比 el-dialog 的默认 2000 高很多 */
}

/* 可选：确保查看器的关闭按钮也能被点到 */
.el-image-viewer__btn {
  z-index: 20001 !important;
}
</style>
