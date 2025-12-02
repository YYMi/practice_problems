<template>
  <aside class="sidebar point-sidebar">
    <!-- 侧边栏头部 -->
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
            <el-button link size="small" :icon="Edit" title="重命名" @click="$emit('open-edit-title', p)" />
            <el-button link size="small" type="danger" :icon="Delete" title="删除" @click="$emit('delete', p)" />
          </div>
        </div>

        <div class="item-title-box">{{ p.title }}</div>
      </div>
    </div>

    <!-- ★★★ 核心修复：添加 append-to-body，让弹窗挂载到 body 上，覆盖全屏 ★★★ -->
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

    <!-- ★★★ 核心修复：添加 append-to-body ★★★ -->
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
import { computed } from 'vue';
import { Document, Trophy, Plus, Top, ArrowUp, ArrowDown, Edit, Delete } from "@element-plus/icons-vue";
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
  'viewMode'
]);

defineEmits(['select', 'open-create-dialog', 'submit-create', 'delete', 'sort', 'open-edit-title', 'open-practice', 'update:categoryPracticeVisible']);

const hasPermission = computed(() => {
  if (props.viewMode === 'read') return false;
  if (props.viewMode === 'dev') return true;
  if (!props.currentSubject || !props.userInfo) return false;
  return props.currentSubject.creatorCode === props.userInfo.user_code;
});
</script>

<style scoped>
/* ============================================================
   1. 侧边栏容器：悬浮毛玻璃
   ============================================================ */
.sidebar { 
  display: flex; flex-direction: column; 
  width: 220px; 
  
  /* 悬浮卡片设计 */
  margin: 12px 0 12px 12px; 
  border-radius: 12px; 
  height: calc(100% - 24px);
  
  /* 毛玻璃背景 */
  background-color: rgba(255, 255, 255, 0.9); 
  backdrop-filter: blur(16px);
  
  border-right: none; 
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.4);
  
  transition: width 0.3s; 
  /* ★★★ 确保内部元素不会溢出圆角 ★★★ */
  overflow: hidden;
}

.sidebar-header { 
  height: 50px; display: flex; justify-content: space-between; align-items: center; padding: 0 15px; 
  border-bottom: 1px solid rgba(0, 0, 0, 0.05); 
  flex-shrink: 0; /* 防止头部被压缩 */
}
.sidebar-title { font-weight: 700; font-size: 14px; color: #2c3e50; display: flex; align-items: center; }
.sidebar-actions { display: flex; align-items: center; gap: 5px; }
.mr-1 { margin-right: 6px; }

.list-container { 
  flex: 1; 
  overflow-y: auto; 
  padding: 10px; 
}

/* ============================================================
   2. 列表项卡片
   ============================================================ */
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

/* ============================================================
   3. 选中状态
   ============================================================ */
.point-item.active { 
  border-color: #764ba2 !important; 
  z-index: 1;
  background-color: #fff; 
  box-shadow: 0 4px 12px rgba(118, 75, 162, 0.2);
}
.point-item.active .item-title-box { color: #764ba2; font-weight: 800; }

/* ============================================================
   4. 难度标签
   ============================================================ */
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

/* ============================================================
   5. 标题文字布局
   ============================================================ */
.item-title-box { 
  font-size: 14px; 
  color: #606266;
  text-align: left; 
  word-break: break-all; 
  line-height: 1; 
  margin-top: 6px; 
  text-indent: 1em; 
}

/* ============================================================
   6. 操作栏
   ============================================================ */
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

/* ============================================================
   7. 滚动条美化 (细长风格)
   ============================================================ */
.custom-scrollbar::-webkit-scrollbar {
  width: 4px; /* 极细 */
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.1); /* 默认很淡 */
  border-radius: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 0, 0, 0.2); /* 鼠标放上去稍微深一点 */
}
</style>
