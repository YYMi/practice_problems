<template>
  <div class="app-wrapper">
    <!-- Header -->
    <header class="app-header">
      <div class="brand">
        <div class="logo-box"><el-icon><Collection /></el-icon></div>
        <div class="brand-text"><span class="main-name">题库</span><span class="sub-name">Manager</span></div>
      </div>
      
      <div class="subject-scroll-area">
        <div
          v-for="item in subjects"
          :key="item.id"
          class="subject-pill"
          :class="{ active: currentSubject?.id === item.id }"
          @click="handleSelectSubject(item)"
        >
          <span class="dot" v-if="currentSubject?.id === item.id"></span>
          <span class="subject-name">{{ item.name }}</span>
          <div class="subject-actions">
            <el-icon class="sub-action edit" @click.stop="openSubjectDialog(item)"><Edit /></el-icon>
            <el-icon class="sub-action del" @click.stop="handleDeleteSubject(item)"><Delete /></el-icon>
          </div>
        </div>
        <el-button class="add-subject-btn" type="primary" icon="Plus" circle plain @click="openSubjectDialog()" />
      </div>
      <div class="header-user"><el-avatar :size="32" style="background-color: #409eff">Admin</el-avatar></div>
    </header>

    <!-- Body -->
    <div class="main-body">
      <!-- Sidebar: Categories -->
      <aside class="sidebar category-sidebar" v-if="currentSubject">
        <div class="sidebar-header">
          <span class="sidebar-title"><el-icon class="mr-1"><Folder /></el-icon> 分类目录</span>
          <el-button link icon="Plus" @click="openCategoryDialog()" />
        </div>
        <div class="list-container custom-scrollbar">
          <el-empty v-if="categories.length === 0" description="暂无" :image-size="60" />
          <!-- ✅ 分类列表循环 -->
          <div
            v-for="(cat, index) in categories"
            :key="cat.id"
            class="list-item category-item"
            :class="[
              { active: currentCategory?.id === cat.id }, 
              getDifficultyClass(cat.difficulty) 
            ]"
            @click="handleSelectCategory(cat)"
          >
            <!-- 左上角：难度 -->
            <div class="corner-tag">{{ getDifficultyLabel(cat.difficulty) }}</div>

            <!-- 右上角：操作 -->
            <div class="corner-actions">
              <el-popover placement="bottom-end" :width="220" trigger="click" popper-class="category-ops-popover">
                <template #reference>
                  <el-icon class="action-icon" @click.stop><MoreFilled /></el-icon>
                </template>
                <div class="ops-container">
                  <div class="ops-row">
                    <span class="ops-label">排序</span>
                    <el-button-group size="small">
                      <el-button :icon="Top" title="置顶" @click="handleSortCategory(cat, 'top')" />
                      <el-button :icon="ArrowUp" title="上移" @click="handleSortCategory(cat, 'up')" :disabled="index === 0" />
                      <el-button :icon="ArrowDown" title="下移" @click="handleSortCategory(cat, 'down')" :disabled="index === categories.length - 1" />
                    </el-button-group>
                  </div>
                  <el-divider style="margin: 8px 0" />
                  <div class="ops-row actions">
                    <el-button size="small" text bg :icon="Edit" @click="openCategoryDialog(cat)">重命名</el-button>
                    <el-button size="small" text bg type="danger" :icon="Delete" @click="handleDeleteCategory(cat)">删除</el-button>
                  </div>
                </div>
              </el-popover>
            </div>

            <!-- 中间：标题 -->
            <div class="item-title-box" :title="cat.categoryName">
              {{ cat.categoryName }}
            </div>
          </div>
        </div>
      </aside>

      <!-- Sidebar: Points -->
      <aside class="sidebar point-sidebar" v-if="currentCategory">
        <div class="sidebar-header">
          <span class="sidebar-title"><el-icon class="mr-1"><Document /></el-icon> 知识点</span>
          <div class="sidebar-actions">
            <el-tooltip content="刷题" placement="top"><el-button link type="warning" icon="Trophy" @click="openCategoryPractice" /></el-tooltip>
            <el-divider direction="vertical" />
            <el-button link icon="Plus" @click="openCreatePointDialog" />
          </div>
        </div>
        <div class="list-container custom-scrollbar">
          <el-empty v-if="points.length === 0" description="暂无" :image-size="60" />
          <!-- ✅ 知识点列表循环 -->
          <div
            v-for="(p, index) in points"
            :key="p.id"
            class="list-item point-item"
            :class="[
              { active: currentPoint?.id === p.id },
              getDifficultyClass(p.difficulty) 
            ]"
            @click="handleSelectPoint(p.id)"
          >
            <!-- 左上角：难度 -->
            <div class="corner-tag">{{ getDifficultyLabel(p.difficulty) }}</div>

            <!-- 右上角：操作 -->
            <div class="corner-actions">
              <el-popover placement="bottom-end" :width="220" trigger="click" popper-class="category-ops-popover">
                <template #reference>
                  <el-icon class="action-icon" @click.stop><MoreFilled /></el-icon>
                </template>
                <div class="ops-container">
                  <div class="ops-row">
                    <span class="ops-label">排序</span>
                    <el-button-group size="small">
                      <el-button :icon="Top" title="置顶" @click="handleSortPoint(p, 'top')" />
                      <el-button :icon="ArrowUp" title="上移" @click="handleSortPoint(p, 'up')" :disabled="index === 0" />
                      <el-button :icon="ArrowDown" title="下移" @click="handleSortPoint(p, 'down')" :disabled="index === points.length - 1" />
                    </el-button-group>
                  </div>
                  <el-divider style="margin: 8px 0" />
                  <div class="ops-row actions">
                    <el-button size="small" text bg :icon="Edit" @click="openEditTitleDialog(p)">编辑</el-button>
                    <el-button size="small" text bg type="danger" :icon="Delete" @click="handleDeletePoint(p)">删除</el-button>
                  </div>
                </div>
              </el-popover>
            </div>

            <!-- 中间：标题 -->
            <div class="item-title-box" :title="p.title">
              {{ p.title }}
            </div>
          </div>
        </div>
      </aside>

      <!-- Main Content -->
      <main class="content-viewport">
        <div v-if="!currentPoint" class="empty-state"><img src="https://gw.alipayobjects.com/zos/antfincdn/ZHrcdLPrvN/empty.svg" width="200"><p>请选择左侧知识点开始编辑</p></div>
        <div v-else class="detail-panel custom-scrollbar">
          <div class="detail-header-card">
            <div class="header-top">
              <div class="title-wrapper"><h1 class="point-title">{{ currentPoint.title }}</h1><el-icon class="edit-title-icon" @click="openEditTitleDialog()"><EditPen /></el-icon></div>
              <div class="actions-wrapper"><el-button type="danger" plain icon="Delete" @click="handleDeletePoint()">删除</el-button><el-button type="primary" class="shua-ti-btn" icon="VideoPlay" @click="drawerVisible = true">练习 & 管理</el-button></div>
            </div>
            <div class="links-section">
              <div class="link-label"><el-icon><Link /></el-icon> 参考资料：</div>
              <div class="link-list">
                <a v-for="(link, index) in parsedLinks" :key="index" :href="formatUrl(link)" target="_blank" class="link-chip">{{ link }}<el-icon class="close-link" @click.prevent="removeLink(index)"><Close /></el-icon></a>
                <el-button size="small" link type="primary" icon="Plus" @click="addLink">添加</el-button>
              </div>
            </div>
          </div>
          <div class="detail-body-layout">
            <div class="panel-column editor-column"><div class="column-header"><span class="col-title">知识详解</span><el-tag size="small" effect="plain">Markdown</el-tag></div><div class="column-content"><PointEditor :pointId="currentPoint.id" :content="currentPoint.content" @update="(val) => { if(currentPoint) currentPoint.content = val }" /></div></div>
            <div class="panel-column image-column"><div class="column-header"><span class="col-title">关联图片</span><el-tag size="small" type="success" effect="plain">Assets</el-tag></div><div class="column-content"><ImageManager :pointId="currentPoint.id" :imagesJson="currentPoint.localImageNames" @update="(val) => { if(currentPoint) currentPoint.localImageNames = val }" /></div></div>
          </div>
        </div>
      </main>
    </div>

    <!-- Dialogs -->
    <QuestionDrawer v-if="currentPoint" v-model:visible="drawerVisible" :pointId="currentPoint.id" :title="currentPoint.title" />
    <CategoryPracticeDrawer v-if="currentCategory" v-model:visible="categoryPracticeVisible" :categoryId="currentCategory.id" :title="currentCategory.categoryName" />
    
    <el-dialog v-model="subjectDialog.visible" :title="subjectDialog.isEdit ? '修改科目' : '添加科目'" width="400px">
      <el-form :model="subjectForm" @submit.prevent><el-form-item label="名称"><el-input v-model="subjectForm.name" @keydown.enter.prevent="submitSubject" /></el-form-item></el-form>
      <template #footer><el-button type="primary" v-reclick="submitSubject">确定</el-button></template>
    </el-dialog>

    <el-dialog v-model="categoryDialog.visible" :title="categoryDialog.isEdit ? '修改分类' : '添加分类'" width="400px">
      <el-form :model="categoryForm" @submit.prevent label-width="50px">
        <el-form-item label="名称"><el-input v-model="categoryForm.categoryName" @keydown.enter.prevent="submitCategory" /></el-form-item>
        <el-form-item label="难度" v-if="categoryDialog.isEdit">
          <el-radio-group v-model="categoryForm.difficulty">
            <el-radio-button :label="0">简单</el-radio-button>
            <el-radio-button :label="1">中等</el-radio-button>
            <el-radio-button :label="2">困难</el-radio-button>
            <el-radio-button :label="3">重点</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer><el-button type="primary" v-reclick="submitCategory">确定</el-button></template>
    </el-dialog>

    <el-dialog v-model="createPointDialog.visible" title="新增知识点" width="400px">
      <el-form :model="createPointForm" @submit.prevent><el-form-item label="名称"><el-input v-model="createPointForm.title" @keydown.enter.prevent="submitCreatePoint" /></el-form-item></el-form>
      <template #footer><el-button @click="createPointDialog.visible = false">取消</el-button><el-button type="primary" v-reclick="submitCreatePoint">确定</el-button></template>
    </el-dialog>

    <!-- 知识点修改弹窗 -->
    <el-dialog v-model="editTitleDialog.visible" title="修改知识点" width="400px">
      <el-form @submit.prevent label-width="50px">
        <el-form-item label="标题">
          <el-input v-model="editTitleDialog.title" @keydown.enter.prevent="submitEditTitle" />
        </el-form-item>
        <el-form-item label="难度">
          <el-radio-group v-model="editTitleDialog.difficulty">
            <el-radio-button :label="0">简单</el-radio-button>
            <el-radio-button :label="1">中等</el-radio-button>
            <el-radio-button :label="2">困难</el-radio-button>
            <el-radio-button :label="3">重点</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer><el-button @click="editTitleDialog.visible = false">取消</el-button><el-button type="primary" v-reclick="submitEditTitle">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { useHomeLogic } from "./logic";
import { Plus, Edit, Delete, VideoPlay, Collection, Folder, Document, EditPen, Link, Close, Trophy, Top, ArrowUp, ArrowDown, MoreFilled } from "@element-plus/icons-vue";
import PointEditor from "../../components/PointEditor.vue";
import ImageManager from "../../components/ImageManager.vue";
import QuestionDrawer from "../../components/QuestionDrawer.vue";
import CategoryPracticeDrawer from "../../components/CategoryPracticeDrawer.vue";

const {
  subjects, currentSubject, categories, currentCategory, points, currentPoint,
  drawerVisible, categoryPracticeVisible, parsedLinks, formatUrl,
  handleSelectSubject, handleSelectCategory, handleSelectPoint, openCategoryPractice,
  openSubjectDialog, submitSubject, handleDeleteSubject, subjectDialog, subjectForm,
  openCategoryDialog, submitCategory, handleDeleteCategory, handleSortCategory, categoryDialog, categoryForm,
  openCreatePointDialog, submitCreatePoint, handleDeletePoint, createPointDialog, createPointForm,
  openEditTitleDialog, submitEditTitle, editTitleDialog,
  addLink, removeLink, getDifficultyLabel, getDifficultyType, getDifficultyClass, handleSortPoint
} = useHomeLogic();
</script>

<style src="./style.css"></style>