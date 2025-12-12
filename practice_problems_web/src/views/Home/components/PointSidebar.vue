<template>
  <aside class="sidebar point-sidebar">
    <!-- 侧边栏头部 (保持不变) -->
    <div class="sidebar-header">
      <span class="sidebar-title">
        <el-icon class="mr-1"><Document /></el-icon> 知识点
      </span>
      <div class="sidebar-actions">
        <el-tooltip content="刷题" placement="top">
          <el-button link type="warning" icon="Trophy" @click="$emit('open-practice')" />
        </el-tooltip>
        
        <template v-if="hasPermission">
          <el-divider direction="vertical" />
          <el-button link icon="Plus" title="新增知识点" @click="$emit('open-create-dialog')" />
        </template>
      </div>
    </div>
    
    <!-- 列表区域 -->
    <div class="list-container custom-scrollbar">
      <el-empty v-if="points.length === 0" description="暂无" :image-size="60" />
      
      <div
        v-for="(p, index) in points"
        :key="p.id"
        class="list-item point-item"
        :class="[{ active: currentPoint?.id === p.id }, getDifficultyClass(p.difficulty)]"
        @click="$emit('select', p.id)"
      >
        <div class="corner-tag">{{ getDifficultyLabel(p.difficulty) }}</div>
        
        <div class="ops-toolbar" v-if="hasPermission" @click.stop>
          <div class="tool-group">
            <el-button link size="small" :icon="Top" title="置顶" @click="$emit('sort', p, 'top')" />
            <el-button link size="small" :icon="ArrowUp" title="上移" @click="$emit('sort', p, 'up')" :disabled="index === 0" />
            <el-button link size="small" :icon="ArrowDown" title="下移" @click="$emit('sort', p, 'down')" :disabled="index === points.length - 1" />
          </div>
          <div class="toolbar-divider"></div>
          <div class="tool-group">
            <!-- ★★★ 新增：分享按钮 ★★★ -->
            <el-button link size="small" :icon="Share" title="分享到集合" @click="openShareDialog(p)" />
            
            <!-- ★★★ 新增：移动按钮 ★★★ -->
            <el-button link size="small" :icon="Switch" title="移动到其他分类" @click="openMoveDialog(p)" />
            
            <el-button link size="small" :icon="Edit" title="重命名" @click="$emit('open-edit-title', p)" />
            <el-button link size="small" type="danger" :icon="Delete" title="删除" @click="$emit('delete', p)" />
          </div>
        </div>

        <template v-if="forceUpdate >= 0">
          <el-tooltip 
            v-if="overflowMap.get(p.id) === true"
            :content="p.title" 
            placement="top" 
            :show-after="100"
            effect="dark"
          >
            <div 
              class="item-title-box" 
              :ref="el => { if (el) checkOverflow(el, p.id); }"
            >
              {{ p.title }}
            </div>
          </el-tooltip>
          <div 
            v-else
            class="item-title-box" 
            :ref="el => { if (el) checkOverflow(el, p.id); }"
          >
            {{ p.title }}
          </div>
        </template>
      </div>
    </div>

    <!-- 分页组件 -->
    <div class="pagination-wrapper" v-if="pointTotal > pointPageSize">
      <el-pagination
        layout="prev, next"
        :current-page="pointPage"
        :page-size="pointPageSize"
        :total="pointTotal"
        @current-change="$emit('page-change', $event)"
        :hide-on-single-page="false"
        small
      />
    </div>

    <!-- 新增知识点弹窗 (保持不变) -->
    <el-dialog 
      v-model="createPointDialog.visible" 
      title="新增知识点" 
      width="400px"
      append-to-body 
    >
      <el-form :model="createPointForm" @submit.prevent>
        <el-form-item label="名称">
          <el-input v-model="createPointForm.title" @keydown.enter.prevent="$emit('submit-create')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createPointDialog.visible = false">取消</el-button>
        <el-button type="primary" v-reclick="() => $emit('submit-create')">确定</el-button>
      </template>
    </el-dialog>

    <!-- ★★★ 新增：移动分类弹窗 ★★★ -->
    <el-dialog
      v-model="moveDialogVisible"
      title="移动知识点"
      width="500px"
      append-to-body
      :close-on-click-modal="false"
    >
      <div style="padding: 4px 16px;">
        <!-- 当前知识标题 -->
        <div style="margin-bottom: 24px;">
          <el-tooltip :content="moveTargetPoint?.title" placement="top" :show-after="100">
            <div 
              style="padding: 12px 14px; background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%); border-left: 3px solid #667eea; border-radius: 6px; color: #495057; font-size: 14px; line-height: 1.6; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; text-overflow: ellipsis; white-space: normal; word-break: break-all; box-shadow: 0 1px 3px rgba(0,0,0,0.05);"
            >
              {{ moveTargetPoint?.title }}
            </div>
          </el-tooltip>
        </div>
        
        <!-- 选择目标分类 -->
        <div>
          <div style="color: #495057; font-size: 13px; margin-bottom: 10px; font-weight: 500;">选择目标分类</div>
          <el-select 
            v-model="selectedTargetCategoryId" 
            placeholder="请选择目标分类" 
            style="width: 100%" 
            size="large"
          >
            <el-option
              v-for="cat in allCategories"
              :key="cat.id"
              :label="cat.categoryName"
              :value="cat.id"
              :disabled="cat.id === currentCategory.id"
            />
          </el-select>
        </div>
      </div>
      
      <template #footer>
        <div style="display: flex; justify-content: flex-end; gap: 12px; padding: 0 8px;">
          <el-button @click="moveDialogVisible = false" size="large">取消</el-button>
          <el-button 
            type="primary" 
            @click="submitMove" 
            :disabled="!selectedTargetCategoryId"
            size="large"
          >
            确定移动
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- ★★★ 新增：分享到集合弹窗 ★★★ -->
    <el-dialog
      v-model="shareDialogVisible"
      title="分享到集合"
      width="500px"
      append-to-body
      :close-on-click-modal="false"
    >
      <div style="padding: 4px 16px;">
        <!-- 知识标题 -->
        <div style="margin-bottom: 24px;">
          <el-tooltip :content="shareTargetPoint?.title" placement="top" :show-after="300">
            <div 
              style="padding: 12px 14px; background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%); border-left: 3px solid #667eea; border-radius: 6px; color: #495057; font-size: 14px; line-height: 1.6; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; text-overflow: ellipsis; white-space: normal; word-break: break-all; box-shadow: 0 1px 3px rgba(0,0,0,0.05);"
            >
              {{ shareTargetPoint?.title }}
            </div>
          </el-tooltip>
        </div>
        
        <!-- 已绑定的集合 -->
        <div v-if="boundCollections.length > 0" style="margin-bottom: 24px;">
          <div style="color: #6c757d; font-size: 13px; margin-bottom: 10px; font-weight: 500;">
            <el-icon style="font-size: 14px; margin-right: 4px;"><Check /></el-icon>
            已在以下集合中
          </div>
          <div style="display: flex; flex-wrap: wrap; gap: 8px;">
            <el-tag 
              v-for="col in boundCollections" 
              :key="col.id"
              type="success"
              size="default"
              effect="light"
              style="border-radius: 4px;"
            >
              {{ col.name }}
            </el-tag>
          </div>
        </div>
        
        <!-- 选择集合 -->
        <div>
          <div style="color: #495057; font-size: 13px; margin-bottom: 10px; font-weight: 500;">选择目标集合</div>
          <el-select 
            v-model="selectedCollectionId" 
            placeholder="请选择要添加的集合" 
            style="width: 100%" 
            size="large"
          >
            <el-option
              v-for="col in collections"
              :key="col.id"
              :label="col.name"
              :value="col.id"
              :disabled="isCollectionBound(col.id)"
            >
              <span :style="{ color: isCollectionBound(col.id) ? '#adb5bd' : '#495057' }">
                {{ col.name }}
              </span>
              <span v-if="isCollectionBound(col.id)" style="color: #adb5bd; font-size: 12px; margin-left: 8px;">（已添加）</span>
            </el-option>
          </el-select>
        </div>
      </div>
      
      <template #footer>
        <div style="display: flex; justify-content: flex-end; gap: 12px; padding: 0 8px;">
          <el-button @click="shareDialogVisible = false" size="large">取消</el-button>
          <el-button 
            type="primary" 
            @click="submitShare" 
            :disabled="!selectedCollectionId"
            :loading="shareLoading"
            size="large"
          >
            确定分享
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 练习抽屉 (保持不变) -->
    <CategoryPracticeDrawer 
      v-if="currentCategory" 
      :visible="categoryPracticeVisible" 
      @update:visible="(val) => $emit('update:categoryPracticeVisible', val)"
      :categoryId="currentCategory.id" 
      :title="currentCategory.categoryName"
      :viewMode="viewMode"         
      :isOwner="hasPermission"
      append-to-body   
    />
  </aside>
</template>

<script setup lang="ts">
import { computed, ref, nextTick } from 'vue';
// ★★★ 引入 Switch 和 Share 图标 ★★★
import { Document, Trophy, Plus, Top, ArrowUp, ArrowDown, Edit, Delete, Switch, Share, Check } from "@element-plus/icons-vue";
import { ElMessage } from 'element-plus';
import CategoryPracticeDrawer from "../../../components/CategoryPracticeDrawer.vue";
import { getCollections, addPointToCollection, getPointCollections, type Collection } from '../../../api/collection';

const props = defineProps([
  'currentCategory', 
  'points', 
  'currentPoint', 
  'createPointDialog', 
  'createPointForm', 
  'categoryPracticeVisible', 
  'getDifficultyLabel', 
  'getDifficultyClass',
  'currentSubject', 
  'userInfo',       
  'viewMode',
  // ★★★ 新增 props：接收所有分类列表，用于下拉选择 ★★★
  'allCategories',
  // 分页相关
  'pointPage',
  'pointPageSize',
  'pointTotal'
]);

// ★★★ 新增 emit：move-point, page-change ★★★
const emit = defineEmits(['select', 'open-create-dialog', 'submit-create', 'delete', 'sort', 'open-edit-title', 'open-practice', 'update:categoryPracticeVisible', 'move-point', 'page-change']);

// ★★★ Tooltip显示控制 ★★★
const overflowMap = ref<Map<number, boolean>>(new Map());
const forceUpdate = ref(0); // 强制更新计数器

const checkOverflow = (el: any, id: number) => {
  nextTick(() => {
    if (!el) return;
    const element = el as HTMLElement;
    // 检查元素是否溢出（高度超过两行）
    // 添加2px的容差，避免浮点数误差
    const isOverflow = element.scrollHeight > element.clientHeight + 2;
    const oldValue = overflowMap.value.get(id);
    if (oldValue !== isOverflow) {
      overflowMap.value.set(id, isOverflow);
      forceUpdate.value++; // 触发重新渲染
    }
  });
};

const shouldShowTooltip = (id: number) => {
  // 如果还没检查过，默认禁用tooltip
  if (!overflowMap.value.has(id)) {
    return true; // 禁用
  }
  // 有溢出时启用tooltip (disabled=false)，没溢出时禁用 (disabled=true)
  const hasOverflow = overflowMap.value.get(id);
  return hasOverflow !== true;
};

const hasPermission = computed(() => {
  if (props.viewMode === 'read') return false;
  if (props.viewMode === 'dev') return true;
  if (!props.currentSubject || !props.userInfo) return false;
  return props.currentSubject.creatorCode === props.userInfo.user_code;
});

// ==========================================
// ★★★ 移动功能逻辑 ★★★
// ==========================================
const moveDialogVisible = ref(false);
const moveTargetPoint = ref<any>(null); // 当前要移动的知识点对象
const selectedTargetCategoryId = ref<number | null>(null);

// 计算可用的目标分类（排除当前分类）
const availableCategories = computed(() => {
  // 1. 调试日志：看看收到了什么
  console.log('父组件传来的所有分类:', props.allCategories);
  console.log('当前分类:', props.currentCategory);

  if (!props.allCategories || !props.currentCategory) return [];
  
  // 2. 过滤逻辑 (使用 != 而不是 !== 以兼容字符串/数字 ID)
  const list = props.allCategories.filter((c: any) => c.id != props.currentCategory.id);
  
  console.log('过滤后的可用分类:', list);
  return list;
});

// 打开弹窗
const openMoveDialog = (point: any) => {
  moveTargetPoint.value = point;
  selectedTargetCategoryId.value = null; // 重置选择
  moveDialogVisible.value = true;
};

// 提交移动
const submitMove = () => {
  if (!selectedTargetCategoryId.value || !moveTargetPoint.value) return;
  
  // 触发父组件事件，传递 知识点ID 和 目标分类ID
  emit('move-point', {
    pointId: moveTargetPoint.value.id,
    targetCategoryId: selectedTargetCategoryId.value
  });
  
  moveDialogVisible.value = false;
};

// ==========================================
// ★★★ 分享功能逻辑 ★★★
// ==========================================
const shareDialogVisible = ref(false);
const shareTargetPoint = ref<any>(null);
const collections = ref<Collection[]>([]);
const boundCollections = ref<Collection[]>([]); // 已绑定的集合列表
const selectedCollectionId = ref<number | null>(null);
const shareLoading = ref(false);

// 打开分享对话框
const openShareDialog = async (point: any) => {
  shareTargetPoint.value = point;
  selectedCollectionId.value = null;
  boundCollections.value = [];
  
  try {
    // 并行获取：1. 所有集合  2. 该知识点已绑定的集合
    const [collectionsRes, boundRes] = await Promise.all([
      getCollections(),
      getPointCollections(point.id)
    ]);
    
    if (collectionsRes.data.code === 200) {
      collections.value = collectionsRes.data.data || [];
      if (collections.value.length === 0) {
        ElMessage.warning('请先创建集合');
        return;
      }
    } else {
      ElMessage.error(collectionsRes.data.msg || '获取集合列表失败');
      return;
    }
    
    if (boundRes.data.code === 200) {
      boundCollections.value = boundRes.data.data || [];
    }
    
    shareDialogVisible.value = true;
  } catch (error) {
    console.error('获取数据失败:', error);
    ElMessage.error('获取数据失败');
  }
};

// 判断集合是否已绑定
const isCollectionBound = (collectionId: number): boolean => {
  return boundCollections.value.some(c => c.id === collectionId);
};

// 提交分享
const submitShare = async () => {
  if (!selectedCollectionId.value || !shareTargetPoint.value) return;
  
  shareLoading.value = true;
  try {
    const res = await addPointToCollection(selectedCollectionId.value, shareTargetPoint.value.id);
    if (res.data.code === 200) {
      ElMessage.success('分享成功');
      shareDialogVisible.value = false;
    } else {
      ElMessage.error(res.data.msg || '分享失败');
    }
  } catch (error: any) {
    console.error('分享知识点失败:', error);
    ElMessage.error(error.response?.data?.msg || '分享失败');
  } finally {
    shareLoading.value = false;
  }
};
</script>

<style scoped>
/* ... 原有样式保持不变 ... */

/* 侧边栏容器：悬浮毛玻璃 */
.sidebar { 
  display: flex; flex-direction: column; 
  width: 220px; 
  margin: 12px 0 12px 12px; 
  border-radius: 12px; 
  height: calc(100% - 24px);
  background-color: rgba(255, 255, 255, 0.9); 
  backdrop-filter: blur(16px);
  border-right: none; 
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.4);
  transition: width 0.3s; 
  overflow: hidden;
}

.sidebar-header { 
  height: 50px; display: flex; justify-content: space-between; align-items: center; padding: 0 15px; 
  border-bottom: 1px solid rgba(0, 0, 0, 0.05); 
  flex-shrink: 0; 
}
.sidebar-title { font-weight: 700; font-size: 14px; color: #2c3e50; display: flex; align-items: center; }
.sidebar-actions { display: flex; align-items: center; gap: 5px; }
.mr-1 { margin-right: 6px; }

.list-container { 
  flex: 1; 
  overflow-y: auto; 
  padding: 10px; 
}

/* 分页样式 */
.pagination-wrapper {
  padding: 12px;
  display: flex;
  justify-content: center;
  border-top: 1px solid rgba(118, 75, 162, 0.15);
  background: linear-gradient(to top, rgba(118, 75, 162, 0.03), transparent);
  flex-shrink: 0;
  margin-top: auto;
}

.pagination-wrapper :deep(.el-pagination) {
  justify-content: center;
}

.pagination-wrapper :deep(.btn-prev),
.pagination-wrapper :deep(.btn-next) {
  background-color: #fff !important;
  border: 2px solid #764ba2 !important;
  border-radius: 8px;
  color: #764ba2 !important;
  font-weight: 600;
  padding: 8px 16px !important;
  min-width: 60px;
  height: 32px !important;
  box-shadow: 0 2px 8px rgba(118, 75, 162, 0.15);
  transition: all 0.2s;
}

.pagination-wrapper :deep(.btn-prev:hover),
.pagination-wrapper :deep(.btn-next:hover) {
  background-color: #764ba2 !important;
  color: #fff !important;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(118, 75, 162, 0.3);
}

.pagination-wrapper :deep(.btn-prev:disabled),
.pagination-wrapper :deep(.btn-next:disabled) {
  color: #c0c4cc !important;
  background-color: #f5f5f5 !important;
  border-color: #dcdfe6 !important;
  box-shadow: none;
  cursor: not-allowed;
  opacity: 0.6;
}

/* 列表项卡片 */
.point-item { 
  position: relative;
  cursor: pointer; 
  padding: 12px 10px 10px 2px; 
  margin-bottom: 8px; 
  border-radius: 8px; 
  background: #fff; 
  border: 2px solid transparent; 
  transition: all 0.2s; 
  box-shadow: 0 2px 4px rgba(0,0,0,0.02);
}

.point-item:hover { 
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
}

/* 选中状态 */
.point-item.active { 
  border-color: #764ba2 !important; 
  z-index: 1;
  background-color: #fff; 
  box-shadow: 0 4px 12px rgba(118, 75, 162, 0.2);
}
.point-item.active .item-title-box { color: #764ba2; font-weight: 800; }

/* 难度标签（集合风格） */
.corner-tag { 
  position: absolute; top: 0; left: 0; font-size: 10px; padding: 1px 6px; 
  border-bottom-right-radius: 8px; border-top-left-radius: 6px; 
  color: #fff; z-index: 2; font-weight: 600;
}

/* 简单 - 鲜绿色 */
.diff-0, .point-item.active.diff-0 { 
  background: linear-gradient(135deg, #fff 0%, #f0f9ff 100%) !important; 
  border: 2px solid transparent;
  border-left: 4px solid #67c23a !important;
}
.point-item.active.diff-0 { 
  border: 2px solid #67c23a !important;
  border-left: 4px solid #67c23a !important;
}
.point-item.diff-0:hover { 
  box-shadow: 0 8px 24px rgba(103, 194, 58, 0.3);
  border-left-width: 6px !important;
}
.diff-0 .corner-tag { 
  background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%) !important; 
}
.diff-0 .item-title-box { color: #529b2e; }
.point-item.active.diff-0 .item-title-box { color: #529b2e; font-weight: 800; }

/* 中等 - 鲜蓝色 */
.diff-1, .point-item.active.diff-1 { 
  background: linear-gradient(135deg, #fff 0%, #f0f7ff 100%) !important; 
  border: 2px solid transparent;
  border-left: 4px solid #409eff !important;
}
.point-item.active.diff-1 { 
  border: 2px solid #409eff !important;
  border-left: 4px solid #409eff !important;
}
.point-item.diff-1:hover { 
  box-shadow: 0 8px 24px rgba(64, 158, 255, 0.3);
  border-left-width: 6px !important;
}
.diff-1 .corner-tag { 
  background: linear-gradient(135deg, #409eff 0%, #66b1ff 100%) !important; 
}
.diff-1 .item-title-box { color: #337ecc; }
.point-item.active.diff-1 .item-title-box { color: #337ecc; font-weight: 800; }

/* 困难 - 鲜橙色 */
.diff-2, .point-item.active.diff-2 { 
  background: linear-gradient(135deg, #fff 0%, #fef9f0 100%) !important; 
  border: 2px solid transparent;
  border-left: 4px solid #e6a23c !important;
}
.point-item.active.diff-2 { 
  border: 2px solid #e6a23c !important;
  border-left: 4px solid #e6a23c !important;
}
.point-item.diff-2:hover { 
  box-shadow: 0 8px 24px rgba(230, 162, 60, 0.3);
  border-left-width: 6px !important;
}
.diff-2 .corner-tag { 
  background: linear-gradient(135deg, #e6a23c 0%, #f0b969 100%) !important; 
}
.diff-2 .item-title-box { color: #b88230; }
.point-item.active.diff-2 .item-title-box { color: #b88230; font-weight: 800; }

/* 重点 - 鲜红色 */
.diff-3, .point-item.active.diff-3 { 
  background: linear-gradient(135deg, #fff 0%, #fff0f0 100%) !important; 
  border: 2px solid transparent;
  border-left: 4px solid #f56c6c !important;
}
.point-item.active.diff-3 { 
  border: 2px solid #f56c6c !important;
  border-left: 4px solid #f56c6c !important;
}
.point-item.diff-3:hover { 
  box-shadow: 0 8px 24px rgba(245, 108, 108, 0.3);
  border-left-width: 6px !important;
}
.diff-3 .corner-tag { 
  background: linear-gradient(135deg, #f56c6c 0%, #f89898 100%) !important; 
}
.diff-3 .item-title-box { color: #c45656; }
.point-item.active.diff-3 .item-title-box { color: #c45656; font-weight: 800; }

/* 标题文字布局 */
.item-title-box { 
  font-size: 14px; 
  color: #606266;
  text-align: left; 
  word-break: break-all; 
  line-height: 1; 
  margin-top: 6px; 
  text-indent: 1em; 
}

/* 操作栏 */
.ops-toolbar {
  position: absolute; top: 2px; right: 2px; display: flex; align-items: center;
  background-color: transparent; height: 20px; padding: 0; z-index: 2;
}
.tool-group { display: flex; align-items: center; }
.toolbar-divider { width: 1px; height: 10px; background-color: #dcdfe6; margin: 0 4px; }

.ops-toolbar .el-button {
  padding: 2px; height: 18px; width: 18px; min-height: 18px; margin-left: 0 !important;
  font-size: 12px; color: #909399;
}
.ops-toolbar .el-button:hover { color: #764ba2; background-color: rgba(118, 75, 162, 0.1); border-radius: 3px; }
.ops-toolbar .el-button--danger:hover { color: #f56c6c; background-color: rgba(245, 108, 108, 0.1); }

.point-item.active .ops-toolbar .el-button { color: #764ba2; opacity: 0.8; }
.point-item.active .ops-toolbar .el-button:hover { opacity: 1; }
.point-item.active .toolbar-divider { background-color: #d3adf7; }

/* 滚动条美化 */
.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(0, 0, 0, 0.1); border-radius: 4px; }
.custom-scrollbar::-webkit-scrollbar-thumb:hover { background: rgba(0, 0, 0, 0.2); }
</style>