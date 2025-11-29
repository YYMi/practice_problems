<template>
  <div class="app-wrapper">
    <!-- 1. 顶部 Header -->
    <header class="app-header">
      <div class="brand">
        <div class="logo-box">
          <el-icon><Collection /></el-icon>
        </div>
        <div class="brand-text">
          <span class="main-name">题库</span>
          <span class="sub-name">Manager</span>
        </div>
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
          
          <!-- 科目操作按钮 (Hover显示) -->
          <div class="subject-actions">
            <el-icon class="sub-action edit" @click.stop="openSubjectDialog(item)"><Edit /></el-icon>
            <el-icon class="sub-action del" @click.stop="handleDeleteSubject(item)"><Delete /></el-icon>
          </div>
        </div>
        
        <el-button
          class="add-subject-btn"
          type="primary"
          icon="Plus"
          circle
          plain
          @click="openSubjectDialog()"
          title="添加科目"
        />
      </div>
      
      <div class="header-user">
        <el-avatar :size="32" style="background-color: #409eff">Admin</el-avatar>
      </div>
    </header>

    <!-- 2. 主体区域 -->
    <div class="main-body">
      
      <!-- 第一栏：分类列表 -->
      <aside class="sidebar category-sidebar" v-if="currentSubject">
        <div class="sidebar-header">
          <span class="sidebar-title">
            <el-icon class="mr-1"><Folder /></el-icon> 分类目录
          </span>
          <el-button link icon="Plus" @click="openCategoryDialog()" />
        </div>
        <div class="list-container custom-scrollbar">
          <el-empty v-if="categories.length === 0" description="暂无" :image-size="60" />
          <div
            v-for="cat in categories"
            :key="cat.id"
            class="list-item category-item"
            :class="{ active: currentCategory?.id === cat.id }"
            @click="handleSelectCategory(cat)"
          >
            <span class="item-text">{{ cat.categoryName }}</span>
            <div class="item-actions">
              <el-icon class="action-icon edit" @click.stop="openCategoryDialog(cat)"><Edit /></el-icon>
              <el-icon class="action-icon del" @click.stop="handleDeleteCategory(cat)"><Delete /></el-icon>
            </div>
          </div>
        </div>
      </aside>

      <!-- 第二栏：知识点列表 -->
      <aside class="sidebar point-sidebar" v-if="currentCategory">
        <div class="sidebar-header">
          <span class="sidebar-title">
            <el-icon class="mr-1"><Document /></el-icon> 知识点
          </span>
          <div class="sidebar-actions">
            <el-tooltip content="本分类综合刷题 (只读)" placement="top">
              <el-button link type="warning" icon="Trophy" @click="openCategoryPractice" />
            </el-tooltip>
            <el-divider direction="vertical" />
            <el-button link icon="Plus" @click="openCreatePointDialog" />
          </div>
        </div>
        <div class="list-container custom-scrollbar">
          <el-empty v-if="points.length === 0" description="暂无" :image-size="60" />
          <div
            v-for="p in points"
            :key="p.id"
            class="list-item point-item"
            :class="{ active: currentPoint?.id === p.id }"
            @click="handleSelectPoint(p.id)"
          >
            <div class="point-left">
              <div class="point-dot"></div>
              <div class="item-name text-ellipsis" :title="p.title">{{ p.title }}</div>
            </div>
            <div class="item-actions">
              <el-icon class="action-icon edit" @click.stop="openEditTitleDialog(p)"><Edit /></el-icon>
              <el-icon class="action-icon del" @click.stop="handleDeletePoint(p)"><Delete /></el-icon>
            </div>
          </div>
        </div>
      </aside>

      <!-- 第三栏：知识点详情 -->
      <main class="content-viewport">
        <div v-if="!currentPoint" class="empty-state">
          <img src="https://gw.alipayobjects.com/zos/antfincdn/ZHrcdLPrvN/empty.svg" alt="empty" width="200">
          <p>请选择左侧知识点开始编辑</p>
        </div>

        <div v-else class="detail-panel custom-scrollbar">
          <!-- 顶部：标题与工具栏 -->
          <div class="detail-header-card">
            <div class="header-top">
              <div class="title-wrapper">
                <h1 class="point-title">{{ currentPoint.title }}</h1>
                <el-icon class="edit-title-icon" @click="openEditTitleDialog()"><EditPen /></el-icon>
              </div>
              <div class="actions-wrapper">
                <el-button type="danger" plain icon="Delete" @click="handleDeletePoint()">删除</el-button>
                <el-button type="primary" class="shua-ti-btn" icon="VideoPlay" @click="drawerVisible = true">
                  专项练习 & 管理
                </el-button>
              </div>
            </div>

            <!-- 外部链接区域 -->
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
                  <el-icon class="close-link" @click.prevent="removeLink(index)"><Close /></el-icon>
                </a>
                <el-button size="small" link type="primary" icon="Plus" @click="addLink">添加</el-button>
              </div>
            </div>
          </div>

          <!-- 左右分栏布局 -->
          <div class="detail-body-layout">
            <!-- 左侧：编辑器 -->
            <div class="panel-column editor-column">
              <div class="column-header">
                <span class="col-title">知识详解</span>
                <el-tag size="small" effect="plain">Markdown</el-tag>
              </div>
              <div class="column-content">
                <PointEditor
                  :pointId="currentPoint.id"
                  :content="currentPoint.content"
                  @update="(val: string) => { if(currentPoint) currentPoint.content = val }"
                />
              </div>
            </div>

            <!-- 右侧：图片管理 -->
            <div class="panel-column image-column">
              <div class="column-header">
                <span class="col-title">关联图片</span>
                <el-tag size="small" type="success" effect="plain">Assets</el-tag>
              </div>
              <div class="column-content">
                <ImageManager
                  :pointId="currentPoint.id"
                  :imagesJson="currentPoint.localImageNames"
                  @update="(val: string) => { if(currentPoint) currentPoint.localImageNames = val }"
                />
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>

    <!-- 1. 单个知识点 刷题/管理 抽屉 -->
    <QuestionDrawer
      v-if="currentPoint"
      v-model:visible="drawerVisible"
      :pointId="currentPoint.id"
      :title="currentPoint.title"
    />

    <!-- 2. 全分类 纯刷题 抽屉 -->
    <CategoryPracticeDrawer
      v-if="currentCategory"
      v-model:visible="categoryPracticeVisible"
      :categoryId="currentCategory.id"
      :title="currentCategory.categoryName"
    />
    
    <!-- 科目弹窗 -->
    <el-dialog v-model="subjectDialog.visible" :title="subjectDialog.isEdit ? '修改科目' : '添加科目'" width="400px">
      <el-form :model="subjectForm" @submit.prevent>
        <el-form-item label="名称"><el-input v-model="subjectForm.name" @keydown.enter.prevent="submitSubject" /></el-form-item>
      </el-form>
      <template #footer><el-button type="primary" v-reclick="submitSubject">确定</el-button></template>
    </el-dialog>

    <!-- 分类弹窗 -->
    <el-dialog v-model="categoryDialog.visible" :title="categoryDialog.isEdit ? '修改分类' : '添加分类'" width="400px">
      <el-form :model="categoryForm" @submit.prevent>
        <el-form-item label="名称"><el-input v-model="categoryForm.categoryName" @keydown.enter.prevent="submitCategory" /></el-form-item>
      </el-form>
      <template #footer><el-button type="primary" v-reclick="submitCategory">确定</el-button></template>
    </el-dialog>

    <!-- 知识点弹窗 -->
    <el-dialog v-model="createPointDialog.visible" title="新增知识点" width="400px">
      <el-form :model="createPointForm" @submit.prevent>
        <el-form-item label="名称"><el-input v-model="createPointForm.title" @keydown.enter.prevent="submitCreatePoint" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createPointDialog.visible = false">取消</el-button>
        <el-button type="primary" v-reclick="submitCreatePoint">确定</el-button>
      </template>
    </el-dialog>

    <!-- 标题修改弹窗 -->
    <el-dialog v-model="editTitleDialog.visible" title="修改名称" width="400px">
      <el-form @submit.prevent>
        <el-input v-model="editTitleDialog.title" @keydown.enter.prevent="submitEditTitle" />
      </el-form>
      <template #footer>
        <el-button @click="editTitleDialog.visible = false">取消</el-button>
        <el-button type="primary" v-reclick="submitEditTitle">保存</el-button>
      </template>
    </el-dialog>

  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus, Edit, Delete, VideoPlay, Collection, Folder, Document, EditPen, Link, Close, Trophy } from "@element-plus/icons-vue";

import PointEditor from "../components/PointEditor.vue";
import ImageManager from "../components/ImageManager.vue";
import QuestionDrawer from "../components/QuestionDrawer.vue";
import CategoryPracticeDrawer from "../components/CategoryPracticeDrawer.vue";

import { getSubjects, createSubject, updateSubject, deleteSubject } from "../api/subject";
import { getCategories, createCategory, updateCategory, deleteCategory } from "../api/category";
import { getPoints, getPointDetail, createPoint, updatePoint, deletePoint } from "../api/point";
import type { Subject, Category, PointSummary, PointDetail } from "../types";

// --- State ---
const subjects = ref<Subject[]>([]);
const currentSubject = ref<Subject | null>(null);
const categories = ref<Category[]>([]);
const currentCategory = ref<Category | null>(null);
const points = ref<PointSummary[]>([]);
const currentPoint = ref<PointDetail | null>(null);

const drawerVisible = ref(false);
const categoryPracticeVisible = ref(false);

// --- Persistence (状态持久化) ---
const STORAGE_KEY = "question_bank_state";

const saveState = () => {
  const state = {
    subjectId: currentSubject.value?.id,
    categoryId: currentCategory.value?.id,
    pointId: currentPoint.value?.id
  };
  localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
};

const restoreState = async () => {
  const saved = localStorage.getItem(STORAGE_KEY);
  if (!saved) return;
  try {
    const state = JSON.parse(saved);
    // 1. 恢复 Subject
    if (state.subjectId) {
      const subject = subjects.value.find(s => s.id === state.subjectId);
      if (subject) {
        await handleSelectSubject(subject, false); // false 表示暂不保存，防止覆盖
        // 2. 恢复 Category
        if (state.categoryId) {
          const category = categories.value.find(c => c.id === state.categoryId);
          if (category) {
            await handleSelectCategory(category, false);
            // 3. 恢复 Point
            if (state.pointId) {
              const point = points.value.find(p => p.id === state.pointId);
              if (point) await handleSelectPoint(point.id);
            }
          }
        }
      }
    }
  } catch (e) { localStorage.removeItem(STORAGE_KEY); }
};

// --- Computed ---
const parsedLinks = computed(() => {
  if (!currentPoint.value || !currentPoint.value.referenceLinks) return [];
  try { return JSON.parse(currentPoint.value.referenceLinks); } catch { return []; }
});
const formatUrl = (url: string) => {
  if (!url) return "";
  url = url.trim();
  if (!url.startsWith("http://") && !url.startsWith("https://")) { return `http://${url}`; }
  return url;
};

// --- API Logic ---
const loadSubjects = async () => { 
  const res = await getSubjects(); 
  if (res.data.code === 200) {
    subjects.value = res.data.data;
    await restoreState(); // 加载完科目后尝试恢复状态
  }
};

const handleSelectSubject = async (item: Subject, autoSave = true) => { 
  currentSubject.value = item; 
  if (autoSave) { currentCategory.value = null; currentPoint.value = null; }
  const res = await getCategories(item.id); 
  if (res.data.code === 200) categories.value = res.data.data;
  if (autoSave) saveState();
};

const handleSelectCategory = async (item: Category, autoSave = true) => { 
  currentCategory.value = item; 
  if (autoSave) currentPoint.value = null; 
  await loadPoints();
  if (autoSave) saveState();
};

const loadPoints = async () => { 
  if (!currentCategory.value) return; 
  const res = await getPoints(currentCategory.value.id); 
  if (res.data.code === 200) points.value = res.data.data; 
};

const handleSelectPoint = async (id: number) => { 
  const res = await getPointDetail(id); 
  if (res.data.code === 200) { 
    currentPoint.value = res.data.data; 
    saveState(); // 选中知识点必保存
  } 
};

// --- Actions ---
const openCategoryPractice = () => { categoryPracticeVisible.value = true; };

// --- CRUD: Subject ---
const subjectDialog = reactive({ visible: false, isEdit: false, id: 0 });
const subjectForm = reactive({ name: "" });
const openSubjectDialog = (subject?: Subject) => {
  if (subject) { subjectDialog.isEdit = true; subjectDialog.id = subject.id; subjectForm.name = subject.name; } 
  else { subjectDialog.isEdit = false; subjectForm.name = ""; }
  subjectDialog.visible = true;
};
const submitSubject = async () => {
  if (!subjectForm.name) return ElMessage.warning("请输入名称");
  try {
    if (subjectDialog.isEdit) { await updateSubject(subjectDialog.id, { name: subjectForm.name,status:1 }); ElMessage.success("修改成功"); } 
    else { await createSubject({ name: subjectForm.name, status: 1 }); ElMessage.success("创建成功"); }
    subjectDialog.visible = false; loadSubjects();
  } catch (e) { ElMessage.error("操作失败"); }
};
const handleDeleteSubject = (subject: Subject) => {
  ElMessageBox.confirm(`确定删除科目“${subject.name}”吗？`, "警告", { type: "warning", confirmButtonText: "删除", cancelButtonText: "取消" }).then(async () => {
    try {
      await deleteSubject(subject.id);
      ElMessage.success("删除成功");
      if (currentSubject.value?.id === subject.id) { currentSubject.value = null; categories.value = []; points.value = []; currentPoint.value = null; saveState(); }
      loadSubjects();
    } catch (e) { ElMessage.error("删除失败"); }
  });
};

// --- CRUD: Category ---
const categoryDialog = reactive({ visible: false, isEdit: false, id: 0 });
const categoryForm = reactive({ categoryName: "" });
const openCategoryDialog = (category?: Category) => {
  if (category) { categoryDialog.isEdit = true; categoryDialog.id = category.id; categoryForm.categoryName = category.categoryName; } 
  else { categoryDialog.isEdit = false; categoryForm.categoryName = ""; }
  categoryDialog.visible = true;
};
const submitCategory = async () => {
  if (!categoryForm.categoryName) return ElMessage.warning("请输入名称");
  try {
    if (categoryDialog.isEdit) { await updateCategory(categoryDialog.id, { categoryName: categoryForm.categoryName }); ElMessage.success("修改成功"); } 
    else { if (!currentSubject.value) return; await createCategory({ subjectId: currentSubject.value.id, categoryName: categoryForm.categoryName }); ElMessage.success("创建成功"); }
    categoryDialog.visible = false;
    if (currentSubject.value) { const res = await getCategories(currentSubject.value.id); if (res.data.code === 200) categories.value = res.data.data; }
  } catch (e) { ElMessage.error("操作失败"); }
};
const handleDeleteCategory = (category: Category) => {
  ElMessageBox.confirm(`确定删除分类“${category.categoryName}”吗？`, "警告", { type: "warning", confirmButtonText: "删除", cancelButtonText: "取消" }).then(async () => {
    try {
      await deleteCategory(category.id);
      ElMessage.success("删除成功");
      if (currentCategory.value?.id === category.id) { currentCategory.value = null; points.value = []; currentPoint.value = null; saveState(); }
      if (currentSubject.value) { const res = await getCategories(currentSubject.value.id); if (res.data.code === 200) categories.value = res.data.data; }
    } catch (e) { ElMessage.error("删除失败"); }
  });
};

// --- CRUD: Point ---
const createPointDialog = reactive({ visible: false });
const createPointForm = reactive({ title: "" });
const openCreatePointDialog = () => { createPointForm.title = ""; createPointDialog.visible = true; };
const submitCreatePoint = async () => { if (!currentCategory.value) return; await createPoint({ categoryId: currentCategory.value.id, title: createPointForm.title, }); createPointDialog.visible = false; loadPoints(); ElMessage.success("创建成功"); };

const handleDeletePoint = (point?: PointSummary | PointDetail) => {
  const target = point || currentPoint.value;
  if (!target) return;
  ElMessageBox.confirm(`确定删除知识点“${target.title}”吗？`, "警告", { type: "warning", confirmButtonText: "删除", cancelButtonText: "取消" }).then(async () => {
    try {
      await deletePoint(target.id);
      ElMessage.success("删除成功");
      if (currentPoint.value?.id === target.id) { currentPoint.value = null; saveState(); }
      loadPoints();
    } catch (e) { ElMessage.error("删除失败"); }
  });
};

const editTitleDialog = reactive({ visible: false, title: "", id: 0 });
const openEditTitleDialog = (point?: PointSummary | PointDetail) => {
  const target = point || currentPoint.value;
  if (!target) return;
  editTitleDialog.id = target.id;
  editTitleDialog.title = target.title;
  editTitleDialog.visible = true;
};
const submitEditTitle = async () => {
  if (!editTitleDialog.id) return;
  await updatePoint(editTitleDialog.id, { title: editTitleDialog.title });
  const p = points.value.find((i) => i.id === editTitleDialog.id);
  if (p) p.title = editTitleDialog.title;
  if (currentPoint.value?.id === editTitleDialog.id) currentPoint.value.title = editTitleDialog.title;
  editTitleDialog.visible = false;
  ElMessage.success("修改成功");
};

// --- Links ---
const addLink = () => { ElMessageBox.prompt("请输入链接地址", "添加链接", { confirmButtonText: "确定", }).then(async ({ value }) => { if (!value) return; const links = [...parsedLinks.value, value]; const jsonStr = JSON.stringify(links); await updatePoint(currentPoint.value!.id, { referenceLinks: jsonStr }); currentPoint.value!.referenceLinks = jsonStr; ElMessage.success("添加成功"); }); };
const removeLink = async (index: number) => { const links = [...parsedLinks.value]; links.splice(index, 1); const jsonStr = JSON.stringify(links); await updatePoint(currentPoint.value!.id, { referenceLinks: jsonStr }); currentPoint.value!.referenceLinks = jsonStr; ElMessage.success("删除成功"); };

onMounted(() => loadSubjects());
</script>

<style scoped>
/* --- 全局与布局 --- */
.app-wrapper {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: #f0f2f5;
  color: #333;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
}

/* --- Header --- */
.app-header {
  height: 64px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.03);
  z-index: 10;
}
.brand {
  display: flex;
  align-items: center;
  margin-right: 40px;
}
.logo-box {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #409eff, #36cfc9);
  color: #fff;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  margin-right: 10px;
  box-shadow: 0 2px 6px rgba(64, 158, 255, 0.3);
}
.brand-text {
  display: flex;
  flex-direction: column;
  line-height: 1.1;
}
.main-name { font-weight: 800; font-size: 16px; color: #2c3e50; }
.sub-name { font-size: 10px; color: #909399; text-transform: uppercase; letter-spacing: 1px; }

.subject-scroll-area {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  overflow-x: auto;
  padding-bottom: 2px;
}
.subject-scroll-area::-webkit-scrollbar { display: none; }

.subject-pill {
  padding: 6px 16px;
  background: #f5f7fa;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  color: #606266;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 6px;
  border: 1px solid transparent;
  position: relative;
}
.subject-pill:hover { background: #e6e8eb; color: #303133; }
.subject-pill.active {
  background: #ecf5ff;
  color: #409eff;
  border-color: #b3d8ff;
  font-weight: 600;
  box-shadow: 0 2px 4px rgba(64, 158, 255, 0.1);
}
.subject-pill .dot { width: 6px; height: 6px; border-radius: 50%; background: #409eff; }

.subject-actions { display: none; gap: 4px; margin-left: 6px; }
.subject-pill:hover .subject-actions { display: flex; }
.sub-action { font-size: 12px; padding: 2px; border-radius: 4px; }
.sub-action:hover { background: #fff; }
.sub-action.del:hover { color: #f56c6c; }

/* --- 主体 --- */
.main-body { flex: 1; display: flex; overflow: hidden; }

.sidebar { display: flex; flex-direction: column; border-right: 1px solid #e4e7ed; transition: width 0.3s; }
.category-sidebar { width: 200px; background-color: #f7f8fa; }
.point-sidebar { width: 220px; background-color: #fff; }

.sidebar-header {
  height: 50px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 15px;
  border-bottom: 1px solid #ebeef5;
}
.sidebar-title { font-weight: 600; font-size: 14px; color: #303133; display: flex; align-items: center; }
.sidebar-actions { display: flex; align-items: center; gap: 5px; }
.mr-1 { margin-right: 6px; }

.list-container { flex: 1; overflow-y: auto; padding: 10px; }
.custom-scrollbar::-webkit-scrollbar { width: 6px; height: 6px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #dcdfe6; border-radius: 3px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }

.list-item {
  padding: 10px 12px; margin-bottom: 4px; border-radius: 6px; cursor: pointer;
  font-size: 14px; color: #606266; transition: all 0.2s; position: relative;
}
.list-item:hover { background-color: rgba(0,0,0,0.04); color: #303133; }
.list-item.active { background-color: #e6f7ff; color: #1890ff; font-weight: 500; }

.category-item, .point-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.item-actions { display: none; gap: 4px; }
.list-item:hover .item-actions { display: flex; }
.action-icon { padding: 4px; border-radius: 4px; font-size: 12px; }
.action-icon:hover { background: #fff; }
.action-icon.del:hover { color: #f56c6c; }

.point-left { display: flex; align-items: center; gap: 8px; flex: 1; overflow: hidden; }
.point-dot { width: 6px; height: 6px; border-radius: 50%; background: #dcdfe6; transition: background 0.3s; flex-shrink: 0; }
.point-item.active .point-dot { background: #409eff; box-shadow: 0 0 4px #409eff; }
.text-ellipsis { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

/* --- 内容详情区 --- */
.content-viewport {
  flex: 1;
  padding: 12px;
  background-color: #f0f2f5;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}
.empty-state { margin: auto; text-align: center; color: #909399; }

.detail-panel {
  background: #fff; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  padding: 20px; height: 100%; display: flex; flex-direction: column; box-sizing: border-box;
}

.detail-header-card { border-bottom: 1px solid #ebeef5; padding-bottom: 20px; margin-bottom: 20px; }
.header-top { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 15px; }
.title-wrapper { display: flex; align-items: center; gap: 10px; }
.point-title { margin: 0; font-size: 22px; color: #1f2f3d; font-weight: 700; }
.edit-title-icon { cursor: pointer; color: #909399; transition: color 0.2s; }
.edit-title-icon:hover { color: #409eff; }

.shua-ti-btn {
  background: linear-gradient(90deg, #409eff, #36a3f7); border: none;
  box-shadow: 0 4px 10px rgba(64, 158, 255, 0.3); padding: 8px 20px; font-weight: 600;
}
.shua-ti-btn:hover { transform: translateY(-1px); box-shadow: 0 6px 12px rgba(64, 158, 255, 0.4); }

.links-section { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.link-label { font-size: 13px; color: #909399; display: flex; align-items: center; gap: 4px; }
.link-list { display: flex; flex-wrap: wrap; gap: 8px; }
.link-chip {
  display: inline-flex; align-items: center; padding: 4px 10px; background: #f2f6fc;
  border-radius: 14px; color: #409eff; text-decoration: none; font-size: 12px;
  transition: all 0.2s; border: 1px solid transparent;
}
.link-chip:hover { background: #ecf5ff; border-color: #b3d8ff; }
.close-link { margin-left: 6px; font-size: 12px; color: #a8abb2; cursor: pointer; }
.close-link:hover { color: #f56c6c; }

/* --- 左右分栏内容区 --- */
.detail-body-layout {
  display: flex;
  flex: 1;
  gap: 15px;
  min-height: 0;
}
.panel-column {
  display: flex;
  flex-direction: column;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  background: #fff;
  overflow: hidden;
}

/* 左侧：编辑器 (占3份) */
.editor-column {
  flex: 3;
  min-width: 0;
}

/* 右侧：图片管理 (占1份，且限制宽度) */
.image-column {
  flex: 1;
  min-width: 300px;
  max-width: 400px;
  border-left: 1px solid #ebeef5;
}

.column-header {
  height: 40px; background: #f9fafc; border-bottom: 1px solid #ebeef5;
  display: flex; align-items: center; justify-content: space-between; padding: 0 15px;
}
.col-title { font-weight: 600; font-size: 14px; color: #606266; }
.column-content { flex: 1; overflow-y: auto; padding: 15px; position: relative; }
</style>
