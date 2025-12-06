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
              :title="link" 
            >
              {{ formatLinkText(link) }}
              <el-icon v-if="hasPermission" class="close-link" @click.prevent="$emit('remove-link', index)"><Close /></el-icon>
            </a>
            <!-- ★★★ 添加链接按钮 ★★★ -->
            <el-button 
              v-if="hasPermission" 
              class="add-link-btn" 
              size="small" 
              plain 
              icon="Plus" 
              @click="$emit('add-link')"
            >
              添加链接
            </el-button>
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

// 格式化链接文本：超过30字符则中间省略
const formatLinkText = (link: string) => {
  if (!link) return '';
  if (link.length <= 30) return link;
  
  // 取前15个字符
  const start = link.substring(0, 15);
  // 取后15个字符
  const end = link.substring(link.length - 15);
  
  return `${start}...${end}`;
};

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
/* ============================================================
   1. 外层容器：强制透明，显示全局紫色背景
   ============================================================ */
.content-viewport { 
  flex: 1; 
  padding: 12px; 
  /* ★★★ 核心：强制背景透明，防止被全局样式覆盖 ★★★ */
  background: transparent !important; 
  background-color: transparent !important;
  
  overflow-y: auto; 
  display: flex; 
  flex-direction: column; 
  position: relative; 
  min-width: 0;
}

/* 空状态文字：适应深色背景 */
.empty-state { 
  margin: auto; 
  text-align: center; 
  color: rgba(255, 255, 255, 0.9); 
}
.empty-state p { margin-top: 10px; font-size: 16px; }

/* ============================================================
   2. 详情卡片：微透毛玻璃 (95%不透明度)
   ============================================================ */
.detail-panel { 
  /* ★★★ 关键修改：不再是死板的纯白，而是微透 ★★★ */
  background: rgba(255, 255, 255, 0.8); 
  backdrop-filter: blur(20px);
  
  border-radius: 12px; 
  /* 阴影加深，增强悬浮感 */
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15); 
  border: 1px solid rgba(255, 255, 255, 0.5);
  
  padding: 20px; 
  height: 100%; 
  display: flex; 
  flex-direction: column; 
  box-sizing: border-box; 
  position: relative; 
  z-index: 2; 
}

/* ============================================================
   3. 内部元素样式
   ============================================================ */
.detail-header-card { 
  border-bottom: 1px solid rgba(0, 0, 0, 0.06); /* 分割线变淡 */
  padding-bottom: 20px; 
  margin-bottom: 20px; 
  flex-shrink: 0; 
}

.header-top { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 15px; }
.title-wrapper { display: flex; align-items: center; gap: 10px; }
.point-title { margin: 0; font-size: 22px; color: #1f2f3d; font-weight: 700; }

/* 编辑图标 hover 变紫 */
.edit-title-icon { cursor: pointer; color: #909399; transition: color 0.2s; }
.edit-title-icon:hover { color: #764ba2; }

/* 刷题按钮：渐变紫 */
.shua-ti-btn { 
  background: linear-gradient(90deg, #667eea, #764ba2); 
  border: none; 
  box-shadow: 0 4px 10px rgba(118, 75, 162, 0.3); 
  padding: 8px 20px; font-weight: 600; 
}
.shua-ti-btn:hover { transform: translateY(-1px); box-shadow: 0 6px 12px rgba(118, 75, 162, 0.4); }

.links-section { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.link-label { font-size: 13px; color: #909399; display: flex; align-items: center; gap: 4px; }
.link-list { display: flex; flex-wrap: wrap; gap: 8px; }

/* 链接标签：淡紫背景 */
.link-chip { 
  display: inline-flex; align-items: center; padding: 4px 10px; 
  background: #f9f0ff; 
  border-radius: 14px; 
  color: #764ba2; 
  text-decoration: none; font-size: 12px; transition: all 0.2s; border: 1px solid transparent; 
}
.link-chip:hover { background: #f3eaff; border-color: #d3adf7; }
.close-link { margin-left: 6px; font-size: 12px; color: #a8abb2; cursor: pointer; }

/* 添加链接按钮样式 */
.add-link-btn {
  height: 26px;
  padding: 4px 12px;
  border-radius: 14px;
  font-size: 12px;
  border-color: #d3adf7;
  color: #764ba2;
}
.add-link-btn:hover {
  background: #f9f0ff;
  border-color: #764ba2;
  color: #764ba2;
}

/* 编辑器布局 */
.detail-body-layout { display: flex; flex: 1; gap: 15px; min-height: 0; }
.panel-column { display: flex; flex-direction: column; border: 1px solid #ebeef5; border-radius: 6px; background: #fff; overflow: hidden; }
.editor-column { flex: 3; min-width: 0; }
.image-column { flex: 1; min-width: 300px; max-width: 400px; border-left: 1px solid #ebeef5; }
.column-header { height: 40px; background: #f9fafc; border-bottom: 1px solid #ebeef5; display: flex; align-items: center; justify-content: space-between; padding: 0 15px; flex-shrink: 0; }
.col-title { font-weight: 600; font-size: 14px; color: #606266; }
.column-content { flex: 1; overflow-y: auto; padding: 15px; position: relative; }

/* 原创/引用 标签样式微调 */
.editor-column.is-mine { border-color: #b3d8ff; background-color: #fff; }
.editor-column.is-mine .column-header { background-color: #ecf5ff; border-bottom-color: #d9ecff; }
.editor-column.is-others { border-color: #e4e7ed; background-color: #fff; }

/* 滚动条美化 */
.custom-scrollbar::-webkit-scrollbar { width: 6px; height: 6px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(0,0,0,0.15); border-radius: 3px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
</style>
