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
        
       
        :all-categories="categories"

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
        
       
        @move-point="handleMovePoint"

        @open-edit-title="openEditTitleDialog"
        @open-practice="openCategoryPractice"
        @update:categoryPracticeVisible="(val:any) => categoryPracticeVisible = val"
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
import { ref, onMounted, onBeforeUnmount } from 'vue';
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
  getDifficultyLabel, getDifficultyClass, loadSubjects, handleMovePoint, handleSaveVideo,
  navigateToPoint, navigateBack, canGoBack,
} = logic;

// Esc 键长按计时器
let escPressTimer: ReturnType<typeof setTimeout> | null = null;
let escPressStartTime = 0;
let lastSwitchTime = 0; // 记录上次切换的时间
const LONG_PRESS_DURATION = 1000; // 长按 1 秒
const SWITCH_COOLDOWN = 200; // 切换后的冷却时间

// 键盘按下事件
const handleKeyDown = (e: KeyboardEvent) => {
  if (e.key !== 'Escape') return;
  
  // 如果正在输入框中，不处理
  const activeEl = document.activeElement;
  if (activeEl && (activeEl.tagName === 'INPUT' || activeEl.tagName === 'TEXTAREA' || (activeEl as HTMLElement).isContentEditable)) {
    return;
  }

  // 记录按下时间
  if (!escPressStartTime) {
    escPressStartTime = Date.now();
  }

  // 如果是阅读模式，需要长按
  if (viewMode.value === 'read') {
    if (!escPressTimer) {
      escPressTimer = setTimeout(() => {
        // 长按 1 秒后切换到编辑模式
        changeMode('edit');
        lastSwitchTime = Date.now(); // 记录切换时间
        ElMessage.success('已切换到编辑模式');
        resetEscTimer();
      }, LONG_PRESS_DURATION);
    }
  }
};

// 键盘释放事件
const handleKeyUp = (e: KeyboardEvent) => {
  if (e.key !== 'Escape') return;

  // 如果刚刚切换过（200ms内），跳过这次释放
  if (Date.now() - lastSwitchTime < SWITCH_COOLDOWN) {
    resetEscTimer();
    return;
  }

  const pressDuration = Date.now() - escPressStartTime;
  
  // 如果是编辑/开发模式，短按即可切换
  if (viewMode.value === 'edit' || viewMode.value === 'dev') {
    changeMode('read');
    lastSwitchTime = Date.now();
    ElMessage.info('已切换到阅读模式');
  } else if (viewMode.value === 'read' && pressDuration < LONG_PRESS_DURATION) {
    // 阅读模式短按，提示需要长按
    ElMessage.warning('长按 Esc 1秒 切换到编辑模式');
  }

  resetEscTimer();
};

// 重置计时器
const resetEscTimer = () => {
  if (escPressTimer) {
    clearTimeout(escPressTimer);
    escPressTimer = null;
  }
  escPressStartTime = 0;
};

// 初始化：从本地缓存读取 + 注册键盘事件
onMounted(() => {
  const cachedMode = localStorage.getItem('app_view_mode');
  if (cachedMode && ['read', 'edit', 'dev'].includes(cachedMode)) {
    viewMode.value = cachedMode;
  }
  
  // 注册键盘事件
  document.addEventListener('keydown', handleKeyDown);
  document.addEventListener('keyup', handleKeyUp);
});

// 清理键盘事件
onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeyDown);
  document.removeEventListener('keyup', handleKeyUp);
  resetEscTimer();
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