<template>
  <div class="app-wrapper">
    <!-- 1. 顶部 Header -->
    <HeaderSection 
      :subjects="subjects"
      :currentSubject="currentSubject"
      :userInfo="userInfo"
      :subjectDialog="subjectDialog"
      :subjectForm="subjectForm"
      :profileDialog="profileDialog"
      :profileForm="profileForm"
      @select="handleSelectSubject"
      @open-dialog="openSubjectDialog"
      @delete="handleDeleteSubject"
      @submit-subject="submitSubject"
      @open-profile="openProfileDialog"
      @submit-profile="submitProfileUpdate"
      @logout="handleLogout"
      @refresh-subjects="loadSubjects" 
    />

    <!-- 主体内容区 -->
    <div class="main-body">
      
      <!-- 2. 左侧分类侧边栏 -->
      <CategorySidebar 
        v-if="currentSubject"
        :currentSubject="currentSubject"
        :categories="categories"
        :currentCategory="currentCategory"
        :categoryDialog="categoryDialog"
        :categoryForm="categoryForm"
        :getDifficultyLabel="getDifficultyLabel"
        :getDifficultyClass="getDifficultyClass"
        @select="handleSelectCategory"
        @open-dialog="openCategoryDialog"
        @submit="submitCategory"
        @delete="handleDeleteCategory"
        @sort="handleSortCategory"
      />

      <!-- 3. 中间知识点侧边栏 -->
      <PointSidebar 
        v-if="currentCategory"
        :currentCategory="currentCategory"
        :points="points"
        :currentPoint="currentPoint"
        :createPointDialog="createPointDialog"
        :createPointForm="createPointForm"
        :categoryPracticeVisible="categoryPracticeVisible"
        :getDifficultyLabel="getDifficultyLabel"
        :getDifficultyClass="getDifficultyClass"
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
  </div>
</template>

<script setup lang="ts">
import { useHomeLogic } from "./logic";
import HeaderSection from "./components/HeaderSection.vue";
import CategorySidebar from "./components/CategorySidebar.vue";
import PointSidebar from "./components/PointSidebar.vue";
import DetailPanel from "./components/DetailPanel.vue";

const logic = useHomeLogic();

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
  getDifficultyLabel, getDifficultyClass,loadSubjects 
} = logic;
</script>

<style src="./style.css"></style>
