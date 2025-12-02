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
        
        <!-- ★★★ 权限控制：新增按钮 ★★★ -->
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
        <!-- 左上角：难度标签 (固定) -->
        <div class="corner-tag">{{ getDifficultyLabel(p.difficulty) }}</div>
        
        <!-- ★★★ 右上角：操作工具栏 (固定，有权限时显示) ★★★ -->
        <div class="ops-toolbar" v-if="hasPermission" @click.stop>
          <!-- 排序按钮组 -->
          <div class="tool-group">
            <el-button link size="small" :icon="Top" title="置顶" @click="$emit('sort', p, 'top')" />
            <el-button link size="small" :icon="ArrowUp" title="上移" @click="$emit('sort', p, 'up')" :disabled="index === 0" />
            <el-button link size="small" :icon="ArrowDown" title="下移" @click="$emit('sort', p, 'down')" :disabled="index === points.length - 1" />
          </div>
          
          <div class="toolbar-divider"></div>

          <!-- 编辑/删除按钮组 -->
          <div class="tool-group">
            <el-button link size="small" :icon="Edit" title="重命名" @click="$emit('open-edit-title', p)" />
            <el-button link size="small" type="danger" :icon="Delete" title="删除" @click="$emit('delete', p)" />
          </div>
        </div>

        <!-- 标题文字 (被 padding-top 挤到下方) -->
        <div class="item-title-box">{{ p.title }}</div>
      </div>
    </div>

    <!-- 新增弹窗 -->
    <el-dialog v-model="createPointDialog.visible" title="新增知识点" width="400px">
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

    <!-- 刷题抽屉 -->
    <CategoryPracticeDrawer 
      v-if="currentCategory" 
      :visible="categoryPracticeVisible" 
      @update:visible="(val) => $emit('update:categoryPracticeVisible', val)"
      :categoryId="currentCategory.id" 
      :title="currentCategory.categoryName"
      :viewMode="viewMode"         
      :isOwner="hasPermission"   
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

// 权限判断
const hasPermission = computed(() => {
  if (props.viewMode === 'read') return false;
  if (props.viewMode === 'dev') return true;
  if (!props.currentSubject || !props.userInfo) return false;
  return props.currentSubject.creatorCode === props.userInfo.user_code;
});
</script>

<style scoped>
.sidebar { display: flex; flex-direction: column; border-right: 1px solid #e4e7ed; transition: width 0.3s; width: 220px; background-color: #fff; }
.sidebar-header { height: 50px; display: flex; justify-content: space-between; align-items: center; padding: 0 15px; border-bottom: 1px solid #ebeef5; }
.sidebar-title { font-weight: 600; font-size: 14px; color: #303133; display: flex; align-items: center; }
.sidebar-actions { display: flex; align-items: center; gap: 5px; }
.mr-1 { margin-right: 6px; }
.list-container { flex: 1; overflow-y: auto; padding: 10px; }

/* --- 列表项卡片基础样式 --- */
.point-item { 
  position: relative;
  cursor: pointer; 
  /* 顶部留出空间，防止文字被标签遮挡 */
   padding: 10px 10px 8px 2px; 
  margin-bottom: 8px; 
  border-radius: 6px; 
  background: #fff; /* 默认白底 */
  
  /* 默认灰色边框，预留位置 */
  border: 2px solid #e4e7ed; 
  transition: all 0.2s; 
}

.point-item:hover { 
  box-shadow: 0 2px 8px rgba(0,0,0,0.05); 
}

/* ============================================================ */
/* ★★★ 1. 选中状态：只改边框和文字，不改背景 ★★★ */
/* ============================================================ */
.point-item.active { 
  /* 边框变紫色 */
  border-color: #722ed1 !important; 
  /* ★★★ 关键修复：背景设为透明(或继承)，让下面的 diff-x 颜色透出来 ★★★ */
  background-color: transparent; 
  z-index: 1;
}

/* 选中时，标题文字变紫色 */
.point-item.active .item-title-box {
  color: #2e39d1;
  font-weight: 600;
  padding: 20px 10px 10px 5px; 
}

/* ============================================================ */
/* ★★★ 2. 难度背景色 & 标签颜色 (强制生效) ★★★ */
/* ============================================================ */

/* 通用标签样式 */
.corner-tag { 
  position: absolute; 
  top: 0; 
  left: 0; 
  font-size: 10px; 
  padding: 2px 8px; 
  border-bottom-right-radius: 6px; 
  border-top-left-radius: 4px; /* 圆角贴合 */
  color: #fff; 
  z-index: 2; 
  font-weight: 500;
}

/* --- 简单 (绿色) --- */
/* 不论是否选中，背景都是浅绿 */
.point-item.diff-0 { background-color: #f0f9eb !important; border-color: #e1f3d8; }
/* 选中时，强制紫色边框 */
.point-item.active.diff-0 { border-color: #722ed1 !important; }
/* 标签颜色 */
.diff-0 .corner-tag { background-color: #67c23a !important; }

/* --- 中等 (黄色) --- */
.point-item.diff-1 { background-color: #fdf6ec !important; border-color: #faecd8; }
.point-item.active.diff-1 { border-color: #722ed1 !important; }
.diff-1 .corner-tag { background-color: #e6a23c !important; }

/* --- 困难 (粉色) --- */
.point-item.diff-2 { background-color: #fef0f0 !important; border-color: #fde2e2; }
.point-item.active.diff-2 { border-color: #722ed1 !important; }
.diff-2 .corner-tag { background-color: #f56c6c !important; }

/* --- 重点 (灰色) --- */
.point-item.diff-3 { background-color: #f4f4f5 !important; border-color: #e9e9eb; }
.point-item.active.diff-3 { border-color: #722ed1 !important; }
.diff-3 .corner-tag { background-color: #909399 !important; }


/* --- ★★★ 右上角：常驻操作栏 ★★★ --- */
.ops-toolbar {
  position: absolute; top: 0; right: 0; display: flex; align-items: center;
  background-color: transparent; height: 26px; padding: 0 2px; z-index: 2;
}
.tool-group { display: flex; align-items: center; }
.toolbar-divider { width: 1px; height: 14px; background-color: #dcdfe6; margin: 0 4px; }

/* 按钮样式 */
.ops-toolbar .el-button {
  padding: 4px; height: 20px; width: 20px; margin-left: 0 !important;
  font-size: 12px; color: #909399;
}
/* 悬停变成紫色 */
.ops-toolbar .el-button:hover {
  color: #722ed1; 
  background-color: rgba(114, 46, 209, 0.1);
}
.ops-toolbar .el-button--danger:hover {
  color: #f56c6c;
  background-color: rgba(245, 108, 108, 0.1);
}

/* 选中时工具栏适配 */
.point-item.active .ops-toolbar .el-button { color: #722ed1; opacity: 0.8; }
.point-item.active .ops-toolbar .el-button:hover { opacity: 1; }
.point-item.active .toolbar-divider { background-color: #d3adf7; }

/* --- 标题文字 --- */
.item-title-box { 
  font-size: 14px; 
  color: #303133;
  text-align: left; 
  word-break: break-all; 
  line-height: 1; 
  
}
</style>
