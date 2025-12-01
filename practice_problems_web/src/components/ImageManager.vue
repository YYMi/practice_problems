<template>
  <div class="images-column">
    <!-- 顶部标题栏 -->
    <div class="section-header">
      <div class="section-title">知识点图片</div>
      <!-- ★★★ 权限控制：新增按钮 ★★★ -->
      <el-upload
        v-if="canEdit"
        action=""
        :http-request="handleUploadRequest"
        :show-file-list="false"
        accept=".jpg,.jpeg,.png,.gif"
      >
        <el-button type="primary" size="small" icon="Plus">新增</el-button>
      </el-upload>
    </div>

    <!-- 
      图片列表区域 
      支持 Ctrl+V 粘贴上传 (仅限有权限时)
    -->
    <div
      class="image-list-container"
      tabindex="0"
      @paste="handlePaste"
      @click="activateArea"
      :class="{ 'is-active': isActive && canEdit }"
      :title="canEdit ? '点击空白处激活后，可直接 Ctrl+V 粘贴截图' : '仅查看模式'"
      :style="{ cursor: canEdit ? 'text' : 'default' }"
    >
      <!-- 空状态提示 -->
      <div v-if="parsedImages.length === 0" class="paste-hint">
        <el-icon :size="30"><Picture /></el-icon>
        <p>暂无图片</p>
        <p v-if="canEdit" class="sub-hint">点击此处激活后<br />按 Ctrl+V 粘贴截图</p>
      </div>

      <!-- 图片卡片循环 -->
      <div
        v-for="(imgUrl, index) in parsedImages"
        :key="index"
        class="img-card"
      >
        <!-- 1. 卡片头部：复制链接 (所有人可见) -->
        <div class="card-header">
          <el-tag type="info" size="small" class="path-tag">图片 {{ index + 1 }}</el-tag>
          <el-button
            link
            type="primary"
            size="small"
            icon="CopyDocument"
            @click.stop="copyImgUrl(imgUrl)"
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

        <!-- 3. 卡片底部：删除按钮 (权限控制) -->
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
  canEdit: boolean; // ★★★ 新增：接收权限参数 ★★★
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
  // ★★★ 权限控制：只有有权限才能激活粘贴区域 ★★★
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
  if (!props.canEdit) return; // 双重保险
  
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
    .catch(() => {
      // 取消删除
    });
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
  // ★★★ 权限控制：没有权限直接拦截粘贴 ★★★
  if (!props.canEdit) {
    return;
  }

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
/* 最外层容器 */
.images-column {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 280px;
  height: 100%; 
  overflow: hidden; 
}

/* 顶部标题栏 */
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

/* 图片列表容器 */
.image-list-container {
  flex: 1;
  overflow-y: auto;
  min-height: 0;

  padding: 5px;
  border: 2px dashed transparent;
  border-radius: 8px;
  outline: none;
  transition: all 0.3s;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 美化滚动条 */
.image-list-container::-webkit-scrollbar {
  width: 6px;
}
.image-list-container::-webkit-scrollbar-thumb {
  background: #dcdfe6;
  border-radius: 4px;
}
.image-list-container::-webkit-scrollbar-track {
  background: transparent;
}

/* 激活状态样式 (仅当有权限时生效) */
.image-list-container:focus,
.image-list-container.is-active {
  border-color: #409eff;
  background-color: #f9faff;
}

/* 空状态 */
.paste-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px; 
  color: #909399;
  border: 1px dashed #dcdfe6;
  border-radius: 8px;
  margin-top: 20px;
  flex-shrink: 0;
}
.sub-hint {
  font-size: 12px;
  text-align: center;
  margin-top: 5px;
  color: #c0c4cc;
}

/* --- 图片卡片样式 --- */
.img-card {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  background: white;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: transform 0.2s;
  flex-shrink: 0; 
}

.img-card:hover {
  box-shadow: 0 4px 16px 0 rgba(0, 0, 0, 0.1);
  border-color: #c6e2ff;
}

.card-header {
  padding: 8px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #ebeef5;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-body {
  height: 200px;
  background: #fff;
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
}

.card-footer {
  padding: 8px 12px;
  border-top: 1px solid #ebeef5;
  display: flex;
  justify-content: flex-end;
  background: #fff;
}

.image-slot {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
  background: #f5f7fa;
  color: #909399;
  font-size: 14px;
}
</style>