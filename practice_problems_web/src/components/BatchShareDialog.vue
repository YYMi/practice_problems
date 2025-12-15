<template>
  <el-dialog
    v-model="visible"
    :title="dialogTitle"
    width="450px"
    append-to-body
    @closed="handleClose"
  >
    <div class="share-content">
      <div class="share-info">
        <el-icon><InfoFilled /></el-icon>
        <span>{{ shareInfo }}</span>
      </div>
      
      <el-divider content-position="left">选择目标合集</el-divider>
      
      <div v-if="loading" class="loading-box">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>加载中...</span>
      </div>
      
      <div v-else-if="collections.length === 0" class="empty-box">
        <el-empty description="暂无合集，请先创建合集" :image-size="60" />
      </div>
      
      <div v-else class="collection-list custom-scrollbar">
        <div
          v-for="col in collections"
          :key="col.id"
          class="collection-item"
          :class="{ selected: selectedCollectionId === col.id }"
          @click="selectedCollectionId = col.id"
        >
          <el-icon><FolderOpened /></el-icon>
          <span class="collection-name">{{ col.name }}</span>
          <el-icon v-if="selectedCollectionId === col.id" class="check-icon"><Check /></el-icon>
        </div>
      </div>
    </div>
    
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button 
        type="primary" 
        :loading="submitting" 
        :disabled="!selectedCollectionId || collections.length === 0"
        @click="handleSubmit"
      >
        确认分享
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { InfoFilled, Loading, FolderOpened, Check } from '@element-plus/icons-vue';
import { getCollections, batchAddPointsToCollection, type Collection } from '../api/collection';

const props = defineProps<{
  modelValue: boolean;
  shareType: 'subject' | 'category'; // 分享类型
  shareId: number; // 科目ID或分类ID
  shareName: string; // 科目名或分类名
}>();

const emit = defineEmits(['update:modelValue', 'success']);

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
});

const loading = ref(false);
const submitting = ref(false);
const collections = ref<Collection[]>([]);
const selectedCollectionId = ref<number | null>(null);

const dialogTitle = computed(() => {
  return props.shareType === 'subject' ? '分享科目到合集' : '分享分类到合集';
});

const shareInfo = computed(() => {
  const typeName = props.shareType === 'subject' ? '科目' : '分类';
  return `将 "${props.shareName}" 下的所有知识点分享到选中的合集（已存在的知识点会自动跳过）`;
});

// 监听弹窗打开，加载合集列表
watch(visible, async (val) => {
  if (val) {
    selectedCollectionId.value = null;
    await loadCollections();
  }
});

const loadCollections = async () => {
  loading.value = true;
  try {
    const res = await getCollections();
    if (res.data.code === 200) {
      // 只显示自己拥有的合集
      collections.value = (res.data.data || []).filter((c: Collection) => c.isOwner);
    }
  } catch (err) {
    console.error('加载合集列表失败:', err);
    ElMessage.error('加载合集列表失败');
  } finally {
    loading.value = false;
  }
};

const handleSubmit = async () => {
  if (!selectedCollectionId.value) return;
  
  submitting.value = true;
  try {
    const options = props.shareType === 'subject' 
      ? { subjectId: props.shareId }
      : { categoryId: props.shareId };
    
    const res = await batchAddPointsToCollection(selectedCollectionId.value, options);
    
    if (res.data.code === 200) {
      const added = res.data.data?.added || 0;
      if (added > 0) {
        ElMessage.success(`成功分享 ${added} 个知识点到合集`);
      } else {
        ElMessage.info('没有新的知识点需要添加（可能已全部在合集中）');
      }
      emit('success', added);
      visible.value = false;
    } else {
      ElMessage.error(res.data.msg || '分享失败');
    }
  } catch (err) {
    console.error('分享失败:', err);
    ElMessage.error('分享失败');
  } finally {
    submitting.value = false;
  }
};

const handleClose = () => {
  selectedCollectionId.value = null;
  collections.value = [];
};
</script>

<style scoped>
.share-content {
  min-height: 200px;
}

.share-info {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px;
  background: #f0f9eb;
  border-radius: 8px;
  color: #67c23a;
  font-size: 14px;
  line-height: 1.5;
}

.share-info .el-icon {
  margin-top: 2px;
  flex-shrink: 0;
}

.loading-box,
.empty-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 0;
  color: #909399;
}

.loading-box .el-icon {
  font-size: 24px;
  margin-bottom: 8px;
}

.collection-list {
  max-height: 300px;
  overflow-y: auto;
}

.collection-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: 2px solid transparent;
}

.collection-item:hover {
  background: #f5f7fa;
}

.collection-item.selected {
  background: #ecf5ff;
  border-color: #409eff;
}

.collection-item .el-icon {
  color: #909399;
  font-size: 18px;
}

.collection-item.selected .el-icon {
  color: #409eff;
}

.collection-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.check-icon {
  color: #67c23a !important;
  font-size: 16px;
}
</style>
