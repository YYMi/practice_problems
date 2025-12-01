<template>
  <main class="content-viewport">
    <div v-if="!currentPoint" class="empty-state">
      <img src="https://gw.alipayobjects.com/zos/antfincdn/ZHrcdLPrvN/empty.svg" width="200">
      <p>请选择左侧知识点开始编辑</p>
    </div>
    
    <div v-else class="detail-panel custom-scrollbar">
      <div class="detail-header-card">
        <div class="header-top">
          <div class="title-wrapper">
            <h1 class="point-title">{{ currentPoint.title }}</h1>
            <el-icon v-if="hasPermission" class="edit-title-icon" @click="$emit('open-edit-title')"><EditPen /></el-icon>
          </div>
          
          <div class="actions-wrapper">
            <el-button v-if="hasPermission" type="danger" plain icon="Delete" @click="$emit('delete')">删除</el-button>
            <el-button type="primary" class="shua-ti-btn" icon="VideoPlay" @click="$emit('update:drawerVisible', true)">练习 & 管理</el-button>
          </div>
        </div>
        
        <div class="links-section">
          <div class="link-label"><el-icon><Link /></el-icon> 参考资料：</div>
          <div class="link-list">
            <a 
              v-for="(link, index) in parsedLinks" 
              :key="index" 
              :href="formatUrl(link)" 
              target="_blank" 
              class="link-chip"
            >
              {{ link }}
              <el-icon v-if="hasPermission" class="close-link" @click.prevent="$emit('remove-link', index)"><Close /></el-icon>
            </a>
            <el-button v-if="hasPermission" size="small" link type="primary" icon="Plus" @click="$emit('add-link')">添加</el-button>
          </div>
        </div>
      </div>
      
      <div class="detail-body-layout">
        <div 
          class="panel-column editor-column"
          :class="{ 'is-mine': isPointOwner, 'is-others': !isPointOwner }"
        >
          <div class="column-header">
            <span class="col-title">知识详解</span>
            <el-tag v-if="isPointOwner" size="small" effect="dark">原创</el-tag>
            <el-tag v-else size="small" type="info" effect="plain">引用</el-tag>
          </div>
          <div class="column-content">
            <PointEditor 
              :pointId="currentPoint.id" 
              :content="currentPoint.content" 
              :canEdit="hasPermission"
              @update="(val) => { if(currentPoint) currentPoint.content = val }" 
            />
          </div>
        </div>
        
        <div class="panel-column image-column">
          <div class="column-header">
            <span class="col-title">关联图片</span>
            <el-tag size="small" type="success" effect="plain">Assets</el-tag>
          </div>
          <div class="column-content">
            <ImageManager 
              :pointId="currentPoint.id" 
              :imagesJson="currentPoint.localImageNames" 
              :canEdit="hasPermission"
              @update="(val) => { if(currentPoint) currentPoint.localImageNames = val }" 
            />
          </div>
        </div>
      </div>
    </div>

    <!-- ★★★★★ 核心修复：传递权限参数给子组件 ★★★★★ -->
    <QuestionDrawer 
      v-if="currentPoint" 
      :visible="drawerVisible" 
      @update:visible="(val) => $emit('update:drawerVisible', val)" 
      :pointId="currentPoint.id" 
      :title="currentPoint.title"
      
      :viewMode="viewMode"       
      :userInfo="userInfo"       
      :isOwner="hasPermission"   
    />
    
    <el-dialog v-if="editTitleDialog" v-model="editTitleDialog.visible" title="修改知识点" width="400px">
      <el-form @submit.prevent label-width="50px">
        <el-form-item label="标题"><el-input v-model="editTitleDialog.title" @keydown.enter.prevent="$emit('submit-edit-title')" /></el-form-item>
        <el-form-item label="难度">
          <el-radio-group v-model="editTitleDialog.difficulty">
            <el-radio-button :label="0">简单</el-radio-button>
            <el-radio-button :label="1">中等</el-radio-button>
            <el-radio-button :label="2">困难</el-radio-button>
            <el-radio-button :label="3">重点</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editTitleDialog.visible = false">取消</el-button>
        <el-button type="primary" v-reclick="() => $emit('submit-edit-title')">保存</el-button>
      </template>
    </el-dialog>
  </main>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { EditPen, Delete, VideoPlay, Link, Close, Plus } from "@element-plus/icons-vue";
import PointEditor from "../../../components/PointEditor.vue";
import ImageManager from "../../../components/ImageManager.vue";
import QuestionDrawer from "../../../components/QuestionDrawer.vue";

const props = defineProps([
  'currentPoint', 'currentSubject', 'isSubjectOwner', 'isPointOwner', 
  'subjectWatermarkText', 'parsedLinks', 'drawerVisible', 'editTitleDialog',
  'userInfo', 'viewMode' 
]);

defineEmits(['update:drawerVisible', 'update:currentPoint', 'open-edit-title', 'submit-edit-title', 'delete', 'add-link', 'remove-link']);

// 计算权限：如果是开发模式，或者 是拥有者
const hasPermission = computed(() => {
  if (props.viewMode === 'read') return false;
  if (props.viewMode === 'dev') return true;
  // 只要是知识点作者 或者 科目作者，都有权
  return !!props.isPointOwner || !!props.isSubjectOwner;
});

const formatUrl = (url: string) => {
  if (!url) return '#';
  url = url.trim();
  if (!/^https?:\/\//i.test(url)) {
    return 'http://' + url;
  }
  return url;
};
</script>

<style scoped>
.content-viewport { flex: 1; padding: 12px; background-color: #f0f2f5; overflow-y: auto; display: flex; flex-direction: column; position: relative; }
.empty-state { margin: auto; text-align: center; color: #909399; }
.detail-panel { background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); padding: 20px; height: 100%; display: flex; flex-direction: column; box-sizing: border-box; position: relative; z-index: 2; }
.detail-header-card { border-bottom: 1px solid #ebeef5; padding-bottom: 20px; margin-bottom: 20px; }
.header-top { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 15px; }
.title-wrapper { display: flex; align-items: center; gap: 10px; }
.point-title { margin: 0; font-size: 22px; color: #1f2f3d; font-weight: 700; }
.edit-title-icon { cursor: pointer; color: #909399; transition: color 0.2s; }
.edit-title-icon:hover { color: #409eff; }
.shua-ti-btn { background: linear-gradient(90deg, #409eff, #36a3f7); border: none; box-shadow: 0 4px 10px rgba(64, 158, 255, 0.3); padding: 8px 20px; font-weight: 600; }
.shua-ti-btn:hover { transform: translateY(-1px); box-shadow: 0 6px 12px rgba(64, 158, 255, 0.4); }
.links-section { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.link-label { font-size: 13px; color: #909399; display: flex; align-items: center; gap: 4px; }
.link-list { display: flex; flex-wrap: wrap; gap: 8px; }
.link-chip { display: inline-flex; align-items: center; padding: 4px 10px; background: #f2f6fc; border-radius: 14px; color: #409eff; text-decoration: none; font-size: 12px; transition: all 0.2s; border: 1px solid transparent; }
.link-chip:hover { background: #ecf5ff; border-color: #b3d8ff; }
.close-link { margin-left: 6px; font-size: 12px; color: #a8abb2; cursor: pointer; }
.detail-body-layout { display: flex; flex: 1; gap: 15px; min-height: 0; }
.panel-column { display: flex; flex-direction: column; border: 1px solid #ebeef5; border-radius: 6px; background: #fff; overflow: hidden; }
.editor-column { flex: 3; min-width: 0; }
.image-column { flex: 1; min-width: 300px; max-width: 400px; border-left: 1px solid #ebeef5; }
.column-header { height: 40px; background: #f9fafc; border-bottom: 1px solid #ebeef5; display: flex; align-items: center; justify-content: space-between; padding: 0 15px; }
.col-title { font-weight: 600; font-size: 14px; color: #606266; }
.column-content { flex: 1; overflow-y: auto; padding: 15px; position: relative; }
.editor-column.is-mine { border-color: #b3d8ff; background-color: #f0f9ff; }
.editor-column.is-mine .column-header { background-color: #ecf5ff; border-bottom-color: #d9ecff; }
.editor-column.is-others { border-color: #e4e7ed; background-color: #fff; }
</style>