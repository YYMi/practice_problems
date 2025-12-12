<template>
  <div class="app-wrapper">
    <!-- 1. 顶部 Header -->
    <!-- 传入 viewMode 并监听更新 -->
    <HeaderSection 
      :subjects="subjects"
      :currentSubject="currentSubject"
      :userInfo="userInfo"
      :subjectDialog="subjectDialog"
      :subjectForm="subjectForm"
      :profileDialog="profileDialog"
      :profileForm="profileForm"
      
      :viewMode="viewMode"
      @update:viewMode="changeMode"

      @select="handleSelectSubject"
      @open-dialog="openSubjectDialog"
      @delete="handleDeleteSubject"
      @submit-subject="submitSubject"
      @open-profile="openProfileDialog"
      @submit-profile="submitProfileUpdate"
      @logout="handleLogout"
      @refresh-subjects="loadSubjects" 
      @toggle-wordbook="toggleWordbook"
    />

    <div class="main-body">
      
      <!-- 2. 左侧分类侧边栏 -->
      <!-- ★★★ 将 isDevMode 替换为 viewMode ★★★ -->
      <CategorySidebar 
        v-if="currentSubject"
        :currentSubject="currentSubject"
        :categories="categories"
        :currentCategory="currentCategory"
        :categoryDialog="categoryDialog"
        :categoryForm="categoryForm"
        :getDifficultyLabel="getDifficultyLabel"
        :getDifficultyClass="getDifficultyClass"
        
        :userInfo="userInfo"
        :viewMode="viewMode"
        
        :categoryPage="categoryPage"
        :categoryPageSize="categoryPageSize"
        :categoryTotal="categoryTotal"
        
        @select="handleSelectCategory"
        @open-dialog="openCategoryDialog"
        @submit="submitCategory"
        @delete="handleDeleteCategory"
        @sort="handleSortCategory"
        @page-change="handleCategoryPageChange"
      />

      <!-- 3. 中间知识点侧边栏 -->
      <!-- ★★★ 将 isDevMode 替换为 viewMode ★★★ -->
      <PointSidebar 
        v-if="currentCategory"
        :currentCategory="currentCategory"
        :currentSubject="currentSubject" 
        :points="points"
        :currentPoint="currentPoint"
        
       
        :all-categories="categories"

        :createPointDialog="createPointDialog"
        :createPointForm="createPointForm"
        :categoryPracticeVisible="categoryPracticeVisible"
        :getDifficultyLabel="getDifficultyLabel"
        :getDifficultyClass="getDifficultyClass"
        
        :userInfo="userInfo"
        :viewMode="viewMode"
        
        :pointPage="pointPage"
        :pointPageSize="pointPageSize"
        :pointTotal="pointTotal"

        @select="handleSelectPoint"
        @open-create-dialog="openCreatePointDialog"
        @submit-create="submitCreatePoint"
        @delete="handleDeletePoint"
        @sort="handleSortPoint"
        
       
        @move-point="handleMovePoint"

        @open-edit-title="openEditTitleDialog"
        @open-practice="openCategoryPractice"
        @update:categoryPracticeVisible="(val:any) => categoryPracticeVisible = val"
        @page-change="handlePointPageChange"
      />

      <!-- 4. 右侧详情面板 -->
      <DetailPanel 
        :currentPoint="currentPoint"
        :currentSubject="currentSubject"
        :currentPointBindings="currentPointBindings"
        :pointsInfoMap="pointsInfoMap"
        :isSubjectOwner="isSubjectOwner"
        :isPointOwner="isPointOwner"
        :subjectWatermarkText="subjectWatermarkText"
        :parsedLinks="parsedLinks"
        :drawerVisible="drawerVisible"
        :editTitleDialog="editTitleDialog"
        :canGoBack="canGoBack"
        
        :userInfo="userInfo"
        :viewMode="viewMode"

        @update:drawerVisible="(val:any) => drawerVisible = val"
        @update:currentPoint="(val:any) => currentPoint = val"
        @open-edit-title="openEditTitleDialog"
        @submit-edit-title="submitEditTitle"
        @delete="handleDeletePoint"
        @add-link="addLink"
        @remove-link="removeLink"
        @format-url="formatUrl"
        @save-video="handleSaveVideo"
        @open-practice="() => drawerVisible = true"
        @refresh-bindings="handleRefreshBindings"
        @cache-point="handleCachePoint"
        @navigate-to-point="handleNavigateToPoint"
        @navigate-back="navigateBack"
      />

    </div>
    
    <!-- 删掉原来的右下角开关 -->
    <!-- 单词本 -->
    <WordBook v-model:visible="wordbookVisible" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { useHomeLogic } from "./logic";
import HeaderSection from "./components/HeaderSection.vue";
import CategorySidebar from "./components/CategorySidebar.vue";
import PointSidebar from "./components/PointSidebar.vue";
import DetailPanel from "./components/DetailPanel.vue";
import WordBook from "../../components/WordBook.vue";

// ★★★ 模式状态管理 ★★★
const viewMode = ref('edit'); // 默认编辑模式

// 先初始化 logic，以便快捷键可以使用
const logic = useHomeLogic();
const wordbookVisible = ref(false);

const toggleWordbook = () => {
  wordbookVisible.value = !wordbookVisible.value;
};

const {
  subjects, currentSubject, categories, currentCategory, points, currentPoint, currentPointBindings, pointsInfoMap,
  categoryPage, categoryPageSize, categoryTotal, // 分类分页状态
  pointPage, pointPageSize, pointTotal, // 知识点分页状态
  subjectDialog, subjectForm, profileDialog, profileForm, userInfo,
  categoryDialog, categoryForm, createPointDialog, createPointForm,
  editTitleDialog, drawerVisible, categoryPracticeVisible,
  parsedLinks, isSubjectOwner, subjectWatermarkText, isPointOwner,
  handleSelectSubject, openSubjectDialog, handleDeleteSubject, submitSubject,
  openProfileDialog, submitProfileUpdate, handleLogout,
  handleSelectCategory, openCategoryDialog, submitCategory, handleDeleteCategory, handleSortCategory,
  handleCategoryPageChange, // 分类分页方法
  handleSelectPoint, openCreatePointDialog, submitCreatePoint, handleDeletePoint, handleSortPoint,
  handlePointPageChange, // 知识点分页方法
  openEditTitleDialog, submitEditTitle, openCategoryPractice,
  addLink, removeLink, formatUrl,
  getDifficultyLabel, getDifficultyClass, loadSubjects, handleMovePoint, handleSaveVideo,
  navigateToPoint, navigateBack, canGoBack,
} = logic;

// 初始化：从本地缓存读取
onMounted(() => {
  const cachedMode = localStorage.getItem('app_view_mode');
  if (cachedMode && ['read', 'edit', 'dev'].includes(cachedMode)) {
    viewMode.value = cachedMode;
  }
});

// 切换模式并缓存
const changeMode = (mode: string) => {
  viewMode.value = mode;
  localStorage.setItem('app_view_mode', mode);
};

// 刷新绑定列表
const handleRefreshBindings = () => {
  if (currentPoint.value?.id) {
    handleSelectPoint(currentPoint.value.id);
  }
};

// 缓存知识点信息
const handleCachePoint = (data: {pointId: number; title: string; categoryId: number}) => {
  // 从 categories 中获取分类名称
  const category = categories.value.find((c: any) => c.id === data.categoryId);
  const categoryName = category?.categoryName || '';
  pointsInfoMap.value.set(data.pointId, { title: data.title, categoryName });
};

// 跳转到知识点
const handleNavigateToPoint = (data: {pointId: number; categoryId: number}) => {
  navigateToPoint(data.pointId, data.categoryId);
};
</script>

<style src="./style.css"></style>