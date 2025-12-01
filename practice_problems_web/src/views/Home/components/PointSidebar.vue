<template>
  <aside class="sidebar point-sidebar">
    <div class="sidebar-header">
      <span class="sidebar-title"><el-icon class="mr-1"><Document /></el-icon> 知识点</span>
      <div class="sidebar-actions">
        <el-tooltip content="刷题" placement="top">
          <el-button link type="warning" icon="Trophy" @click="$emit('open-practice')" />
        </el-tooltip>
        
        <!-- ★★★ 权限控制：添加按钮 ★★★ -->
        <template v-if="hasPermission">
          <el-divider direction="vertical" />
          <el-button link icon="Plus" title="新增知识点" @click="$emit('open-create-dialog')" />
        </template>
      </div>
    </div>
    
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
        
        <!-- ★★★ 权限控制：三个点操作 ★★★ -->
        <div class="corner-actions" v-if="hasPermission">
          <el-popover placement="bottom-end" :width="220" trigger="click" popper-class="category-ops-popover">
            <template #reference><el-icon class="action-icon" @click.stop><MoreFilled /></el-icon></template>
            <div class="ops-container">
              <div class="ops-row">
                <span class="ops-label">排序</span>
                <el-button-group size="small">
                  <el-button :icon="Top" title="置顶" @click="$emit('sort', p, 'top')" />
                  <el-button :icon="ArrowUp" title="上移" @click="$emit('sort', p, 'up')" :disabled="index === 0" />
                  <el-button :icon="ArrowDown" title="下移" @click="$emit('sort', p, 'down')" :disabled="index === points.length - 1" />
                </el-button-group>
              </div>
              <el-divider style="margin: 8px 0" />
              <div class="ops-row actions">
                <el-button size="small" text bg :icon="Edit" @click="$emit('open-edit-title', p)">编辑</el-button>
                <el-button size="small" text bg type="danger" :icon="Delete" @click="$emit('delete', p)">删除</el-button>
              </div>
            </div>
          </el-popover>
        </div>

        <div class="item-title-box">{{ p.title }}</div>
      </div>
    </div>

    <el-dialog v-model="createPointDialog.visible" title="新增知识点" width="400px">
      <el-form :model="createPointForm" @submit.prevent><el-form-item label="名称"><el-input v-model="createPointForm.title" @keydown.enter.prevent="$emit('submit-create')" /></el-form-item></el-form>
      <template #footer><el-button @click="createPointDialog.visible = false">取消</el-button><el-button type="primary" v-reclick="() => $emit('submit-create')">确定</el-button></template>
    </el-dialog>

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
import { Document, Trophy, Plus, MoreFilled, Top, ArrowUp, ArrowDown, Edit, Delete } from "@element-plus/icons-vue";
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
  'viewMode' // <--- 接收模式
]);

defineEmits(['select', 'open-create-dialog', 'submit-create', 'delete', 'sort', 'open-edit-title', 'open-practice', 'update:categoryPracticeVisible']);



// 这个 hasPermission 是计算好的（是否是作者 OR 开发模式）
// 我们直接把它传给子组件作为 isOwner 即可
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

.point-item { position: relative; cursor: pointer; padding: 10px; margin-bottom: 8px; border-radius: 6px; background: #fff; border: 1px solid #e4e7ed; transition: all 0.2s; }
.point-item:hover { box-shadow: 0 2px 8px rgba(0,0,0,0.05); }
.point-item.active { border-color: #409eff; background-color: #ecf5ff; }
.corner-tag { position: absolute; top: 0; left: 0; font-size: 10px; padding: 1px 4px; border-bottom-right-radius: 6px; background: #909399; color: #fff; }
.corner-actions { position: absolute; top: 2px; right: 2px; display: none; }
.point-item:hover .corner-actions { display: block; }
.action-icon { font-size: 14px; color: #909399; padding: 2px; }
.action-icon:hover { color: #409eff; background: #f0f2f5; border-radius: 4px; }
.item-title-box { margin-top: 10px; font-size: 14px; text-align: left; word-break: break-all; line-height: 1.4; }

.diff-0 .corner-tag { background-color: #67c23a; }
.diff-1 .corner-tag { background-color: #e6a23c; }
.diff-2 .corner-tag { background-color: #f56c6c; }
.diff-3 .corner-tag { background-color: #909399; }
</style>