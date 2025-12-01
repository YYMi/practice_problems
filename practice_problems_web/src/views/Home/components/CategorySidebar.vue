<template>
  <aside class="sidebar category-sidebar">
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
    
    <div class="list-container custom-scrollbar">
      <el-empty v-if="categories.length === 0" description="暂无" :image-size="60" />
      <div
        v-for="(cat, index) in categories"
        :key="cat.id"
        class="list-item category-item"
        :class="[{ active: currentCategory?.id === cat.id }, getDifficultyClass(cat.difficulty)]"
        @click="$emit('select', cat)"
      >
        <div class="corner-tag">{{ getDifficultyLabel(cat.difficulty) }}</div>
        
        <!-- ★★★ 权限控制：三个点操作 ★★★ -->
        <div class="corner-actions" v-if="hasPermission">
          <el-popover placement="bottom-end" :width="220" trigger="click" popper-class="category-ops-popover">
            <template #reference>
              <el-icon class="action-icon" @click.stop><MoreFilled /></el-icon>
            </template>
            
            <div class="ops-container">
              <div class="ops-row">
                <span class="ops-label">排序</span>
                <el-button-group size="small">
                  <el-button :icon="Top" title="置顶" @click="$emit('sort', cat, 'top')" />
                  <el-button :icon="ArrowUp" title="上移" @click="$emit('sort', cat, 'up')" :disabled="index === 0" />
                  <el-button :icon="ArrowDown" title="下移" @click="$emit('sort', cat, 'down')" :disabled="index === categories.length - 1" />
                </el-button-group>
              </div>
              <el-divider style="margin: 8px 0" />
              <div class="ops-row actions">
                <el-button size="small" text bg :icon="Edit" @click="$emit('open-dialog', cat)">重命名</el-button>
                <el-button size="small" text bg type="danger" :icon="Delete" @click="$emit('delete', cat)">删除</el-button>
              </div>
            </div>
          </el-popover>
        </div>

        <div class="item-title-box">{{ cat.categoryName }}</div>
      </div>
    </div>

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
import { Folder, Plus, MoreFilled, Top, ArrowUp, ArrowDown, Edit, Delete } from "@element-plus/icons-vue";

const props = defineProps([
  'currentSubject', 
  'categories', 
  'currentCategory', 
  'categoryDialog', 
  'categoryForm', 
  'getDifficultyLabel', 
  'getDifficultyClass',
  'userInfo',   
  'viewMode' // <--- 接收模式
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
.sidebar { display: flex; flex-direction: column; border-right: 1px solid #e4e7ed; transition: width 0.3s; width: 200px; background-color: #f7f8fa; }
.sidebar-header { height: 50px; display: flex; justify-content: space-between; align-items: center; padding: 0 15px; border-bottom: 1px solid #ebeef5; }
.sidebar-title { font-weight: 600; font-size: 14px; color: #303133; display: flex; align-items: center; max-width: 140px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.mr-1 { margin-right: 6px; }
.list-container { flex: 1; overflow-y: auto; padding: 10px; }

.category-item { position: relative; cursor: pointer; padding: 10px; margin-bottom: 8px; border-radius: 6px; background: #fff; border: 1px solid #e4e7ed; transition: all 0.2s; }
.category-item:hover { box-shadow: 0 2px 8px rgba(0,0,0,0.05); }
.category-item.active { border-color: #409eff; background-color: #ecf5ff; }
.corner-tag { position: absolute; top: 0; left: 0; font-size: 10px; padding: 1px 4px; border-bottom-right-radius: 6px; background: #909399; color: #fff; }
.corner-actions { position: absolute; top: 2px; right: 2px; display: none; }
.category-item:hover .corner-actions { display: block; }
.action-icon { font-size: 14px; color: #909399; padding: 2px; }
.action-icon:hover { color: #409eff; background: #f0f2f5; border-radius: 4px; }
.item-title-box { margin-top: 10px; font-size: 14px; text-align: center; word-break: break-all; }

.diff-0 .corner-tag { background-color: #67c23a; }
.diff-1 .corner-tag { background-color: #e6a23c; }
.diff-2 .corner-tag { background-color: #f56c6c; }
.diff-3 .corner-tag { background-color: #909399; }
</style>