<template>
  <aside class="sidebar category-sidebar">
    <!-- 侧边栏头部 -->
    <div class="sidebar-header">
      <span class="sidebar-title" :title="currentSubject.name">
        <el-icon class="mr-1"><Folder /></el-icon> {{ currentSubject.name }}
      </span>
      
      <!-- ★★★ 权限控制：添加按钮 ★★★ -->
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
        
        <!-- ★★★ 右上角：常驻透明工具栏 (有权限时显示) ★★★ -->
        <div class="ops-toolbar" v-if="hasPermission" @click.stop>
          <!-- 排序组 -->
          <div class="tool-group">
            <el-button link size="small" :icon="Top" title="置顶" @click="$emit('sort', cat, 'top')" />
            <el-button link size="small" :icon="ArrowUp" title="上移" @click="$emit('sort', cat, 'up')" :disabled="index === 0" />
            <el-button link size="small" :icon="ArrowDown" title="下移" @click="$emit('sort', cat, 'down')" :disabled="index === categories.length - 1" />
          </div>
          
          <div class="toolbar-divider"></div>

          <!-- 编辑/删除组 -->
          <div class="tool-group">
            <el-button link size="small" :icon="Edit" title="重命名" @click="$emit('open-dialog', cat)" />
            <el-button link size="small" type="danger" :icon="Delete" title="删除" @click="$emit('delete', cat)" />
          </div>
        </div>

        <!-- 标题文字 -->
        <div class="item-title-box">{{ cat.categoryName }}</div>
      </div>
    </div>

    <!-- 弹窗保持不变 -->
    <el-dialog v-model="categoryDialog.visible" :title="categoryDialog.isEdit ? '修改分类' : '添加分类'" width="400px">
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
.sidebar { display: flex; flex-direction: column; border-right: 1px solid #bf1fdf; transition: width 0.3s; width: 200px; background-color: #f7f8fa; }
.sidebar-header { height: 50px; display: flex; justify-content: space-between; align-items: center; padding: 0 15px; border-bottom: 1px solid #ebeef5; }
.sidebar-title { font-weight: 600; font-size: 14px; color: #303133; display: flex; align-items: center; max-width: 140px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.mr-1 { margin-right: 6px; }
.list-container { flex: 1; overflow-y: auto; padding: 10px; }

.category-item { 
  position: relative; 
  cursor: pointer; 
  padding: 10px 10px 10px 2px; 
  margin-bottom: 8px; 
  border-radius: 6px; 
  /* ★★★ 修改1：预留 2px 的边框位置，防止选中时抖动 ★★★ */
  border: 2px solid transparent; 
  transition: all 0.2s; 
  background-color: #fff; 
}
.category-item:hover { filter: brightness(0.95); }

/* ============================================================ */
/* ★★★ 选中状态：直接变色，不加阴影，保证贴合 ★★★ */
/* ============================================================ */
.category-item.active { 
  /* 把边框变成蓝色，它会包裹住里面的标签 */
  border-color: #409eff !important;
  z-index: 1; 
}

/* 选中时，文字变蓝 */
.category-item.active .item-title-box {
  color: #409eff;
  font-weight: bold;
}

/* ============================================================ */
/* ★★★ 难度背景色 & 标签样式 ★★★ */
/* ============================================================ */

/* --- 标签样式微调 --- */
.corner-tag { 
  position: absolute; 
  top: 0; 
  left: 0; 
  font-size: 10px; 
  padding: 1px 4px; 
  border-bottom-right-radius: 6px; 
  /* ★★★ 修改2：给左上角加一点圆角，完美匹配外面的圆框 ★★★ */
  border-top-left-radius: 4px; 
  z-index: 2; 
}

/* 简单 */
.diff-0, .category-item.active.diff-0 { background-color: #f0f9eb !important; border-color: #e1f3d8; }
.category-item.active.diff-0 { border-color: #409eff !important; } /* 选中时强制蓝框 */
.diff-0 .corner-tag { background-color: #67c23a !important; color: #fff; }

/* 中等 */
.diff-1, .category-item.active.diff-1 { background-color: #fdf6ec !important; border-color: #faecd8; }
.category-item.active.diff-1 { border-color: #409eff !important; }
.diff-1 .corner-tag { background-color: #e6a23c !important; color: #fff; }

/* 困难 */
.diff-2, .category-item.active.diff-2 { background-color: #fef0f0 !important; border-color: #fde2e2; }
.category-item.active.diff-2 { border-color: #409eff !important; }
.diff-2 .corner-tag { background-color: #f56c6c !important; color: #fff; }

/* 重点 */
.diff-3, .category-item.active.diff-3 { background-color: #f4f4f5 !important; border-color: #e9e9eb; }
.category-item.active.diff-3 { border-color: #409eff !important; }
.diff-3 .corner-tag { background-color: #909399 !important; color: #fff; }


/* --- 其他通用样式 --- */
.item-title-box { margin-top: 10px; font-size: 14px; text-align: center; word-break: break-all; line-height: 1; }

/* 右上角工具栏 */
.ops-toolbar {
  position: absolute; top: 2px; right: 2px; display: flex; align-items: center;
  background-color: transparent; height: 20px; padding: 0; z-index: 2;
}
.toolbar-divider { width: 1px; height: 10px; background-color: #dcdfe6; margin: 0 3px; }
.ops-toolbar .el-button { padding: 2px; height: 18px; width: 18px; min-height: 18px; margin-left: 0 !important; font-size: 12px; color: #909399; }
.ops-toolbar .el-button:hover { color: #409eff; background-color: rgba(0,0,0,0.05); border-radius: 3px; }
.ops-toolbar .el-button--danger:hover { color: #f56c6c; background-color: rgba(245, 108, 108, 0.1); }

/* 选中时工具栏微调 */
.category-item.active .ops-toolbar .el-button { color: #409eff; }
.category-item.active .toolbar-divider { background-color: #a0cfff; }
</style>
