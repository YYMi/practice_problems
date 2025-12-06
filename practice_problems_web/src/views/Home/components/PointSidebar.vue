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
            <!-- ★★★ 新增：移动按钮 ★★★ -->
            <el-button link size="small" :icon="Switch" title="移动到其他分类" @click="openMoveDialog(p)" />
            
            <el-button link size="small" :icon="Edit" title="重命名" @click="$emit('open-edit-title', p)" />
            <el-button link size="small" type="danger" :icon="Delete" title="删除" @click="$emit('delete', p)" />
          </div>
        </div>

        <div class="item-title-box">{{ p.title }}</div>
      </div>
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
      width="400px"
      append-to-body
    >
      <el-form label-width="80px">
        <el-form-item label="当前知识点">
          <el-tag type="info">{{ moveTargetPoint?.title }}</el-tag>
        </el-form-item>
        <el-form-item label="目标分类">
       <!-- 目标分类下拉框 -->
        <el-select v-model="selectedTargetCategoryId" placeholder="请选择目标分类" style="width: 100%">
          <!-- 直接循环 allCategories -->
          <!-- 加上 :disabled 来禁用当前分类，而不是过滤掉它 -->
          <el-option
            v-for="cat in allCategories"
            :key="cat.id"
            :label="cat.categoryName"
            :value="cat.id"
            :disabled="cat.id === currentCategory.id"
          />
        </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="moveDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitMove" :disabled="!selectedTargetCategoryId">
          确定移动
        </el-button>
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
import { computed, ref } from 'vue';
// ★★★ 引入 Switch 图标 ★★★
import { Document, Trophy, Plus, Top, ArrowUp, ArrowDown, Edit, Delete, Switch } from "@element-plus/icons-vue";
import { ElMessage } from 'element-plus';
import CategoryPracticeDrawer from "../../../components/CategoryPracticeDrawer.vue";

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
  'allCategories' 
]);

// ★★★ 新增 emit：move-point ★★★
const emit = defineEmits(['select', 'open-create-dialog', 'submit-create', 'delete', 'sort', 'open-edit-title', 'open-practice', 'update:categoryPracticeVisible', 'move-point']);

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

/* 难度标签 */
.corner-tag { 
  position: absolute; top: 0; left: 0; font-size: 10px; padding: 1px 6px; 
  border-bottom-right-radius: 8px; border-top-left-radius: 6px; 
  color: #fff; z-index: 2; font-weight: 600;
}
.diff-0, .point-item.active.diff-0 { background-color: #f0f9eb !important; border-color: #e1f3d8; }
.point-item.active.diff-0 { border-color: #764ba2 !important; }
.diff-0 .corner-tag { background-color: #67c23a !important; }
.diff-1, .point-item.active.diff-1 { background-color: #fdf6ec !important; border-color: #faecd8; }
.point-item.active.diff-1 { border-color: #764ba2 !important; }
.diff-1 .corner-tag { background-color: #e6a23c !important; }
.diff-2, .point-item.active.diff-2 { background-color: #fef0f0 !important; border-color: #fde2e2; }
.point-item.active.diff-2 { border-color: #764ba2 !important; }
.diff-2 .corner-tag { background-color: #f56c6c !important; }
.diff-3, .point-item.active.diff-3 { background-color: #f4f4f5 !important; border-color: #e9e9eb; }
.point-item.active.diff-3 { border-color: #764ba2 !important; }
.diff-3 .corner-tag { background-color: #909399 !important; }

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