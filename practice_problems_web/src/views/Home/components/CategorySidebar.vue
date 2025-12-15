<template>
  <aside class="sidebar category-sidebar">
    <!-- 侧边栏头部 -->
    <div class="sidebar-header">
      <span class="sidebar-title" :title="currentSubject.name">
        <el-icon class="mr-1"><Folder /></el-icon> {{ currentSubject.name }}
      </span>
      
      <div class="header-actions">
        <!-- 分享整个科目到合集 -->
        <el-button 
          v-if="hasPermission" 
          link 
          :icon="Share" 
          title="分享科目到合集"
          @click="handleShareSubject" 
        />
        <!-- 添加分类 -->
        <el-button 
          v-if="hasPermission" 
          link 
          icon="Plus" 
          title="添加分类"
          @click="$emit('open-dialog')" 
        />
      </div>
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
            <el-button link size="small" :icon="Share" title="分享到合集" @click="handleShare(cat)" />
            <el-button link size="small" :icon="Edit" title="重命名" @click="$emit('open-dialog', cat)" />
            <el-button link size="small" type="danger" :icon="Delete" title="删除" @click="$emit('delete', cat)" />
          </div>
        </div>

        <!-- 标题文字 -->
        <template v-if="forceUpdate >= 0">
          <el-tooltip 
            v-if="overflowMap.get(cat.id) === true"
            :content="cat.categoryName" 
            placement="top" 
            :show-after="100"
            effect="dark"
          >
            <div 
              class="item-title-box"
              :ref="el => { if (el) checkOverflow(el, cat.id); }"
            >
              {{ cat.categoryName }}
            </div>
          </el-tooltip>
          <div 
            v-else
            class="item-title-box"
            :ref="el => { if (el) checkOverflow(el, cat.id); }"
          >
            {{ cat.categoryName }}
          </div>
        </template>
      </div>
    </div>

    <!-- 分页组件 -->
    <div class="pagination-wrapper" v-if="categoryTotal > categoryPageSize">
      <el-pagination
        layout="prev, next"
        :current-page="categoryPage"
        :page-size="categoryPageSize"
        :total="categoryTotal"
        @current-change="$emit('page-change', $event)"
        :hide-on-single-page="false"
        small
      />
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
import { computed, ref, nextTick } from 'vue';
import { Folder, Plus, Top, ArrowUp, ArrowDown, Edit, Delete, Share } from "@element-plus/icons-vue";

const props = defineProps([
  'currentSubject', 
  'categories', 
  'currentCategory', 
  'categoryDialog', 
  'categoryForm', 
  'getDifficultyLabel', 
  'getDifficultyClass',
  'userInfo',   
  'viewMode',
  'categoryPage',
  'categoryPageSize',
  'categoryTotal'
]);

// 分享到合集
const emit = defineEmits(['select', 'open-dialog', 'submit', 'delete', 'sort', 'page-change', 'share']);
const handleShare = (cat: any) => {
  emit('share', { type: 'category', id: cat.id, name: cat.categoryName });
};
// 分享整个科目
const handleShareSubject = () => {
  emit('share', { type: 'subject', id: props.currentSubject.id, name: props.currentSubject.name });
};

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
  border-radius: 16px; 
  height: calc(100% - 24px);
  
  /* 毛玻璃背景 */
  background-color: rgba(255, 255, 255, 0.95); 
  backdrop-filter: blur(20px);
  
  /* 柔和边框 */
  border: 2px solid rgba(118, 75, 162, 0.2);
  box-shadow: 0 8px 32px rgba(118, 75, 162, 0.12), 
              0 2px 8px rgba(0, 0, 0, 0.05);
  
  transition: all 0.3s; 
  overflow: hidden;
}

.sidebar:hover {
  border-color: rgba(118, 75, 162, 0.3);
  box-shadow: 0 12px 40px rgba(118, 75, 162, 0.18), 
              0 4px 12px rgba(0, 0, 0, 0.08);
}

/* 头部样式 */
.sidebar-header { 
  height: 50px; 
  display: flex; 
  justify-content: space-between; 
  align-items: center; 
  padding: 0 15px; 
  border-bottom: 1px solid rgba(118, 75, 162, 0.15);
  background: linear-gradient(to bottom, rgba(118, 75, 162, 0.03), transparent);
  flex-shrink: 0; 
}

.sidebar-title { 
  font-weight: 700; 
  font-size: 14px; 
  color: #2c3e50;   
  display: flex; align-items: center; max-width: 140px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; 
}
.mr-1 { margin-right: 6px; }
.header-actions { display: flex; align-items: center; gap: 4px; }
.list-container { flex: 1; overflow-y: auto; padding: 10px; }

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
   4. 难度标签 & 颜色（集合风格）
   ============================================================ */
.corner-tag { 
  position: absolute; top: 0; left: 0; font-size: 10px; padding: 1px 6px; 
  border-bottom-right-radius: 8px; 
  border-top-left-radius: 6px; 
  z-index: 2; color: #fff;
  font-weight: 600;
}

/* 简单 - 鲜绿色 */
.diff-0, .category-item.active.diff-0 { 
  background: linear-gradient(135deg, #fff 0%, #f0f9ff 100%) !important; 
  border: 2px solid transparent;
  border-left: 4px solid #67c23a !important;
}
.category-item.active.diff-0 { 
  border: 2px solid #67c23a !important;
  border-left: 4px solid #67c23a !important;
}
.category-item.diff-0:hover { 
  box-shadow: 0 8px 24px rgba(103, 194, 58, 0.3);
  border-left-width: 6px !important;
}
.diff-0 .corner-tag { 
  background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%) !important; 
}
.diff-0 .item-title-box { color: #529b2e; }
.category-item.active.diff-0 .item-title-box { color: #529b2e; font-weight: 800; }

/* 中等 - 鲜蓝色 */
.diff-1, .category-item.active.diff-1 { 
  background: linear-gradient(135deg, #fff 0%, #f0f7ff 100%) !important; 
  border: 2px solid transparent;
  border-left: 4px solid #409eff !important;
}
.category-item.active.diff-1 { 
  border: 2px solid #409eff !important;
  border-left: 4px solid #409eff !important;
}
.category-item.diff-1:hover { 
  box-shadow: 0 8px 24px rgba(64, 158, 255, 0.3);
  border-left-width: 6px !important;
}
.diff-1 .corner-tag { 
  background: linear-gradient(135deg, #409eff 0%, #66b1ff 100%) !important; 
}
.diff-1 .item-title-box { color: #337ecc; }
.category-item.active.diff-1 .item-title-box { color: #337ecc; font-weight: 800; }

/* 困难 - 鲜橙色 */
.diff-2, .category-item.active.diff-2 { 
  background: linear-gradient(135deg, #fff 0%, #fef9f0 100%) !important; 
  border: 2px solid transparent;
  border-left: 4px solid #e6a23c !important;
}
.category-item.active.diff-2 { 
  border: 2px solid #e6a23c !important;
  border-left: 4px solid #e6a23c !important;
}
.category-item.diff-2:hover { 
  box-shadow: 0 8px 24px rgba(230, 162, 60, 0.3);
  border-left-width: 6px !important;
}
.diff-2 .corner-tag { 
  background: linear-gradient(135deg, #e6a23c 0%, #f0b969 100%) !important; 
}
.diff-2 .item-title-box { color: #b88230; }
.category-item.active.diff-2 .item-title-box { color: #b88230; font-weight: 800; }

/* 重点 - 鲜红色 */
.diff-3, .category-item.active.diff-3 { 
  background: linear-gradient(135deg, #fff 0%, #fff0f0 100%) !important; 
  border: 2px solid transparent;
  border-left: 4px solid #f56c6c !important;
}
.category-item.active.diff-3 { 
  border: 2px solid #f56c6c !important;
  border-left: 4px solid #f56c6c !important;
}
.category-item.diff-3:hover { 
  box-shadow: 0 8px 24px rgba(245, 108, 108, 0.3);
  border-left-width: 6px !important;
}
.diff-3 .corner-tag { 
  background: linear-gradient(135deg, #f56c6c 0%, #f89898 100%) !important; 
}
.diff-3 .item-title-box { color: #c45656; }
.category-item.active.diff-3 .item-title-box { color: #c45656; font-weight: 800; }

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
