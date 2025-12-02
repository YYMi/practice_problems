<template>
  <aside class="sidebar category-sidebar">
    <!-- 侧边栏头部 -->
    <div class="sidebar-header">
      <span class="sidebar-title" :title="currentSubject.name">
        <el-icon class="mr-1"><Folder /></el-icon> {{ currentSubject.name }}
      </span>
      
      <!-- 权限控制：添加按钮 -->
      <el-button 
        v-if="hasPermission" 
        link 
        icon="Plus" 
        title="添加分类"
        @click="$emit('open-dialog')" 
      />
    </div>
    
    <!-- 列表区域 -->
    <div class="list-container custom-scrollbar">
      <el-empty v-if="categories.length === 0" description="暂无" :image-size="60" />
      
      <div
        v-for="(cat, index) in categories"
        :key="cat.id"
        class="list-item category-item"
        :class="[{ active: currentCategory?.id === cat.id }, getDifficultyClass(cat.difficulty)]"
        @click="$emit('select', cat)"
      >
        <!-- 左上角：难度标签 -->
        <div class="corner-tag">{{ getDifficultyLabel(cat.difficulty) }}</div>
        
        <!-- 右上角：常驻透明工具栏 -->
        <div class="ops-toolbar" v-if="hasPermission" @click.stop>
          <div class="tool-group">
            <el-button link size="small" :icon="Top" title="置顶" @click="$emit('sort', cat, 'top')" />
            <el-button link size="small" :icon="ArrowUp" title="上移" @click="$emit('sort', cat, 'up')" :disabled="index === 0" />
            <el-button link size="small" :icon="ArrowDown" title="下移" @click="$emit('sort', cat, 'down')" :disabled="index === categories.length - 1" />
          </div>
          
          <div class="toolbar-divider"></div>

          <div class="tool-group">
            <el-button link size="small" :icon="Edit" title="重命名" @click="$emit('open-dialog', cat)" />
            <el-button link size="small" type="danger" :icon="Delete" title="删除" @click="$emit('delete', cat)" />
          </div>
        </div>

        <!-- 标题文字 -->
        <div class="item-title-box">{{ cat.categoryName }}</div>
      </div>
    </div>

    <!-- ★★★ 核心修复：添加 append-to-body ★★★ -->
    <el-dialog 
      v-model="categoryDialog.visible" 
      :title="categoryDialog.isEdit ? '修改分类' : '添加分类'" 
      width="400px"
      append-to-body
    >
      <el-form :model="categoryForm" @submit.prevent label-width="50px">
        <el-form-item label="名称"><el-input v-model="categoryForm.categoryName" @keydown.enter.prevent="$emit('submit')" /></el-form-item>
        <el-form-item label="难度" v-if="categoryDialog.isEdit">
          <el-radio-group v-model="categoryForm.difficulty">
            <el-radio-button :label="0">简单</el-radio-button>
            <el-radio-button :label="1">中等</el-radio-button>
            <el-radio-button :label="2">困难</el-radio-button>
            <el-radio-button :label="3">重点</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer><el-button type="primary" v-reclick="() => $emit('submit')">确定</el-button></template>
    </el-dialog>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Folder, Plus, Top, ArrowUp, ArrowDown, Edit, Delete } from "@element-plus/icons-vue";

const props = defineProps([
  'currentSubject', 
  'categories', 
  'currentCategory', 
  'categoryDialog', 
  'categoryForm', 
  'getDifficultyLabel', 
  'getDifficultyClass',
  'userInfo',   
  'viewMode'
]);

defineEmits(['select', 'open-dialog', 'submit', 'delete', 'sort']);

const hasPermission = computed(() => {
  if (props.viewMode === 'read') return false;
  if (props.viewMode === 'dev') return true;
  if (!props.currentSubject || !props.userInfo) return false;
  return props.currentSubject.creatorCode === props.userInfo.user_code;
});
</script>

<style scoped>
/* ============================================================
   1. 侧边栏容器：悬浮毛玻璃卡片
   ============================================================ */
.sidebar { 
  display: flex; 
  flex-direction: column; 
  width: 200px; 
  
  /* 悬浮卡片设计 */
  margin: 12px 0 12px 12px; 
  border-radius: 12px; 
  height: calc(100% - 24px);
  
  /* 毛玻璃背景 */
  background-color: rgba(255, 255, 255, 0.65); 
  backdrop-filter: blur(16px);
  
  /* 去掉右边框，改用柔和的阴影 */
  border-right: none; 
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.4);
  
  transition: width 0.3s; 
  /* 防止内部元素溢出圆角 */
  overflow: hidden;
}

/* 头部样式 */
.sidebar-header { 
  height: 50px; 
  display: flex; 
  justify-content: space-between; 
  align-items: center; 
  padding: 0 15px; 
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  flex-shrink: 0; 
}

.sidebar-title { 
  font-weight: 700; 
  font-size: 14px; 
  color: #2c3e50;   
  display: flex; align-items: center; max-width: 140px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; 
}
.mr-1 { margin-right: 6px; }
.list-container { flex: 1; overflow-y: auto; padding: 10px; }

/* ============================================================
   2. 列表项卡片
   ============================================================ */
.category-item { 
  position: relative; 
  cursor: pointer; 
  padding: 12px 10px 10px 2px; 
  margin-bottom: 8px; 
  border-radius: 8px; 
  border: 2px solid transparent; 
  transition: all 0.2s; 
  
  /* 卡片背景稍微白一点 */
  background-color: rgba(255, 255, 255, 0.8); 
  box-shadow: 0 2px 4px rgba(0,0,0,0.02);
}

.category-item:hover { 
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
}

/* ============================================================
   3. 选中状态：紫色系
   ============================================================ */
.category-item.active { 
  border-color: #764ba2 !important;
  z-index: 1; 
  background-color: #fff; 
  box-shadow: 0 4px 12px rgba(118, 75, 162, 0.2);
}

.category-item.active .item-title-box {
  color: #764ba2;
  font-weight: 800;
}

/* ============================================================
   4. 难度标签 & 颜色
   ============================================================ */
.corner-tag { 
  position: absolute; top: 0; left: 0; font-size: 10px; padding: 1px 6px; 
  border-bottom-right-radius: 8px; 
  border-top-left-radius: 6px; 
  z-index: 2; color: #fff;
  font-weight: 600;
}

/* 简单 */
.diff-0, .category-item.active.diff-0 { background-color: #f0f9eb !important; border-color: #e1f3d8; }
.category-item.active.diff-0 { border-color: #764ba2 !important; }
.diff-0 .corner-tag { background-color: #67c23a !important; }

/* 中等 */
.diff-1, .category-item.active.diff-1 { background-color: #fdf6ec !important; border-color: #faecd8; }
.category-item.active.diff-1 { border-color: #764ba2 !important; }
.diff-1 .corner-tag { background-color: #e6a23c !important; }

/* 困难 */
.diff-2, .category-item.active.diff-2 { background-color: #fef0f0 !important; border-color: #fde2e2; }
.category-item.active.diff-2 { border-color: #764ba2 !important; }
.diff-2 .corner-tag { background-color: #f56c6c !important; }

/* 重点 */
.diff-3, .category-item.active.diff-3 { background-color: #f4f4f5 !important; border-color: #e9e9eb; }
.category-item.active.diff-3 { border-color: #764ba2 !important; }
.diff-3 .corner-tag { background-color: #909399 !important; }

/* ============================================================
   5. 其他
   ============================================================ */
.item-title-box { 
  margin-top: 6px; 
  font-size: 14px; 
  text-align: center; 
  word-break: break-all; 
  line-height: 1.4; 
  color: #606266;
}

.ops-toolbar {
  position: absolute; top: 2px; right: 2px; display: flex; align-items: center;
  background-color: transparent; height: 20px; padding: 0; z-index: 2;
}
.toolbar-divider { width: 1px; height: 10px; background-color: #dcdfe6; margin: 0 3px; }

.ops-toolbar .el-button { 
  padding: 2px; height: 18px; width: 18px; min-height: 18px; margin-left: 0 !important; 
  font-size: 12px; color: #909399; 
}
.ops-toolbar .el-button:hover { 
  color: #764ba2; 
  background-color: rgba(118, 75, 162, 0.1); 
  border-radius: 3px; 
}
.ops-toolbar .el-button--danger:hover { color: #f56c6c; background-color: rgba(245, 108, 108, 0.1); }

.category-item.active .ops-toolbar .el-button { color: #764ba2; opacity: 0.8; }
.category-item.active .ops-toolbar .el-button:hover { opacity: 1; }
.category-item.active .toolbar-divider { background-color: #d3adf7; }

/* ============================================================
   6. 滚动条美化 (细长风格)
   ============================================================ */
.custom-scrollbar::-webkit-scrollbar {
  width: 4px; /* 极细 */
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 4px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 0, 0, 0.2);
}
</style>
