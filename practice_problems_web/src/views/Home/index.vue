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
        
        @select="handleSelectCategory"
        @open-dialog="openCategoryDialog"
        @submit="submitCategory"
        @delete="handleDeleteCategory"
        @sort="handleSortCategory"
      />

      <!-- 3. 中间知识点侧边栏 -->
      <!-- ★★★ 将 isDevMode 替换为 viewMode ★★★ -->
      <PointSidebar 
        v-if="currentCategory"
        :currentCategory="currentCategory"
        :currentSubject="currentSubject" 
        :points="points"
        :currentPoint="currentPoint"
        :createPointDialog="createPointDialog"
        :createPointForm="createPointForm"
        :categoryPracticeVisible="categoryPracticeVisible"
        :getDifficultyLabel="getDifficultyLabel"
        :getDifficultyClass="getDifficultyClass"
        
        :userInfo="userInfo"
        :viewMode="viewMode"

        @select="handleSelectPoint"
        @open-create-dialog="openCreatePointDialog"
        @submit-create="submitCreatePoint"
        @delete="handleDeletePoint"
        @sort="handleSortPoint"
        @open-edit-title="openEditTitleDialog"
        @open-practice="openCategoryPractice"
        @update:categoryPracticeVisible="(val:any) => categoryPracticeVisible = val"
      />

      <!-- 4. 右侧详情面板 -->
      <DetailPanel 
        :currentPoint="currentPoint"
        :currentSubject="currentSubject"
        :isSubjectOwner="isSubjectOwner"
        :isPointOwner="isPointOwner"
        :subjectWatermarkText="subjectWatermarkText"
        :parsedLinks="parsedLinks"
        :drawerVisible="drawerVisible"
        :editTitleDialog="editTitleDialog"
        
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
      />

    </div>
    
    <!-- 删掉原来的右下角开关 -->
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useHomeLogic } from "./logic";
import HeaderSection from "./components/HeaderSection.vue";
import CategorySidebar from "./components/CategorySidebar.vue";
import PointSidebar from "./components/PointSidebar.vue";
import DetailPanel from "./components/DetailPanel.vue";

// ★★★ 模式状态管理 ★★★
const viewMode = ref('edit'); // 默认编辑模式

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

const logic = useHomeLogic();
// ... 解构 logic (保持不变) ...
const {
  subjects, currentSubject, categories, currentCategory, points, currentPoint,
  subjectDialog, subjectForm, profileDialog, profileForm, userInfo,
  categoryDialog, categoryForm, createPointDialog, createPointForm,
  editTitleDialog, drawerVisible, categoryPracticeVisible,
  parsedLinks, isSubjectOwner, subjectWatermarkText, isPointOwner,
  handleSelectSubject, openSubjectDialog, handleDeleteSubject, submitSubject,
  openProfileDialog, submitProfileUpdate, handleLogout,
  handleSelectCategory, openCategoryDialog, submitCategory, handleDeleteCategory, handleSortCategory,
  handleSelectPoint, openCreatePointDialog, submitCreatePoint, handleDeletePoint, handleSortPoint,
  openEditTitleDialog, submitEditTitle, openCategoryPractice,
  addLink, removeLink, formatUrl,
  getDifficultyLabel, getDifficultyClass, loadSubjects 
} = logic;
</script>

<style src="./style.css"></style>