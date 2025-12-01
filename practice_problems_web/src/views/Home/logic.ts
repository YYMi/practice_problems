import { ref, reactive, computed, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";

// API 引用
import { getSubjects, createSubject, updateSubject, deleteSubject } from "../../api/subject";
import { getCategories, createCategory, updateCategory, deleteCategory, updateCategorySort } from "../../api/category";
// ✅ 引入 updatePointSort
import { getPoints, getPointDetail, createPoint, updatePoint, deletePoint, updatePointSort } from "../../api/point";
import type { Subject, Category, PointSummary, PointDetail } from "../../types";

export function useHomeLogic() {
  // --- State ---
  const subjects = ref<Subject[]>([]);
  const currentSubject = ref<Subject | null>(null);
  const categories = ref<Category[]>([]);
  const currentCategory = ref<Category | null>(null);
  const points = ref<PointSummary[]>([]);
  const currentPoint = ref<PointDetail | null>(null);

  const drawerVisible = ref(false);
  const categoryPracticeVisible = ref(false);

  // --- Persistence ---
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
      if (state.subjectId) {
        const subject = subjects.value.find(s => s.id === state.subjectId);
        if (subject) {
          await handleSelectSubject(subject, false);
          if (state.categoryId) {
            const category = categories.value.find(c => c.id === state.categoryId);
            if (category) {
              await handleSelectCategory(category, false);
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

  // --- Helpers for Difficulty ---
  const getDifficultyLabel = (diff?: number) => {
    switch (diff) {
      case 1: return '中等';
      case 2: return '困难';
      case 3: return '重点';
      default: return '简单';
    }
  };

  const getDifficultyClass = (diff?: number) => {
    const d = diff || 0;
    return `diff-item-${d}`;
  };

  const getDifficultyType = (diff?: number) => {
    switch (diff) {
      case 1: return 'info';    
      case 2: return 'warning'; 
      case 3: return 'danger';  
      default: return 'success'; 
    }
  };

  // --- API Logic ---
  const loadSubjects = async () => { 
    const res = await getSubjects(); 
    if (res.data.code === 200) {
      subjects.value = res.data.data;
      await restoreState();
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
      saveState(); 
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
      if (subjectDialog.isEdit) { await updateSubject(subjectDialog.id, { name: subjectForm.name, status: 1 }); ElMessage.success("修改成功"); } 
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
  const categoryForm = reactive({ categoryName: "", difficulty: 0 });
  
  const openCategoryDialog = (category?: Category) => {
    if (category) { 
      categoryDialog.isEdit = true; 
      categoryDialog.id = category.id; 
      categoryForm.categoryName = category.categoryName; 
      categoryForm.difficulty = category.difficulty || 0; 
    } else { 
      categoryDialog.isEdit = false; 
      categoryForm.categoryName = ""; 
      categoryForm.difficulty = 0; 
    }
    categoryDialog.visible = true;
  };

  const submitCategory = async () => {
    if (!categoryForm.categoryName) return ElMessage.warning("请输入名称");
    try {
      if (categoryDialog.isEdit) { 
        await updateCategory(categoryDialog.id, { categoryName: categoryForm.categoryName, difficulty: categoryForm.difficulty }); 
        ElMessage.success("修改成功"); 
      } else { 
        if (!currentSubject.value) return; 
        await createCategory({ subjectId: currentSubject.value.id, categoryName: categoryForm.categoryName }); 
        ElMessage.success("创建成功"); 
      }
      categoryDialog.visible = false;
      if (currentSubject.value) { 
        const res = await getCategories(currentSubject.value.id); 
        if (res.data.code === 200) categories.value = res.data.data; 
      }
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

  const handleSortCategory = async (category: Category, action: 'top' | 'up' | 'down') => {
    try {
      await updateCategorySort(category.id, action);
      if (currentSubject.value) {
        const res = await getCategories(currentSubject.value.id);
        if (res.data.code === 200) {
          categories.value = res.data.data;
          ElMessage.success("排序成功");
        }
      }
    } catch (e) {
      ElMessage.error("排序失败");
    }
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

  // ✅ 新增：知识点排序
  const handleSortPoint = async (point: PointSummary, action: 'top' | 'up' | 'down') => {
    try {
      await updatePointSort(point.id, action);
      if (currentCategory.value) {
        const res = await getPoints(currentCategory.value.id);
        if (res.data.code === 200) points.value = res.data.data;
        ElMessage.success("排序成功");
      }
    } catch (e) {
      ElMessage.error("排序失败");
    }
  };

  // ✅ 修改：编辑知识点标题/难度
  const editTitleDialog = reactive({ 
    visible: false, 
    title: "", 
    id: 0,
    difficulty: 0 // 新增
  });

  const openEditTitleDialog = (point?: PointSummary | PointDetail) => {
    const target = point || currentPoint.value;
    if (!target) return;
    editTitleDialog.id = target.id;
    editTitleDialog.title = target.title;
    // 尝试获取 difficulty，若不存在则为 0
    editTitleDialog.difficulty = (target as any).difficulty || 0;
    editTitleDialog.visible = true;
  };

  const submitEditTitle = async () => {
    if (!editTitleDialog.id) return;
    // 提交标题和难度
    await updatePoint(editTitleDialog.id, { 
      title: editTitleDialog.title,
      difficulty: editTitleDialog.difficulty
    });
    
    // 刷新本地数据
    const p = points.value.find((i) => i.id === editTitleDialog.id);
    if (p) {
      p.title = editTitleDialog.title;
      (p as any).difficulty = editTitleDialog.difficulty;
    }
    if (currentPoint.value?.id === editTitleDialog.id) {
      currentPoint.value.title = editTitleDialog.title;
    }
    
    editTitleDialog.visible = false;
    ElMessage.success("修改成功");
  };

  const addLink = () => { ElMessageBox.prompt("请输入链接地址", "添加链接", { confirmButtonText: "确定", }).then(async ({ value }) => { if (!value) return; const links = [...parsedLinks.value, value]; const jsonStr = JSON.stringify(links); await updatePoint(currentPoint.value!.id, { referenceLinks: jsonStr }); currentPoint.value!.referenceLinks = jsonStr; ElMessage.success("添加成功"); }); };
  const removeLink = async (index: number) => { const links = [...parsedLinks.value]; links.splice(index, 1); const jsonStr = JSON.stringify(links); await updatePoint(currentPoint.value!.id, { referenceLinks: jsonStr }); currentPoint.value!.referenceLinks = jsonStr; ElMessage.success("删除成功"); };

  onMounted(() => loadSubjects());

  return {
    subjects, currentSubject, categories, currentCategory, points, currentPoint,
    drawerVisible, categoryPracticeVisible, parsedLinks, formatUrl,
    handleSelectSubject, handleSelectCategory, handleSelectPoint, openCategoryPractice,
    openSubjectDialog, submitSubject, handleDeleteSubject, subjectDialog, subjectForm,
    openCategoryDialog, submitCategory, handleDeleteCategory, handleSortCategory, categoryDialog, categoryForm,
    openCreatePointDialog, submitCreatePoint, handleDeletePoint, createPointDialog, createPointForm,
    openEditTitleDialog, submitEditTitle, editTitleDialog,
    addLink, removeLink,
    getDifficultyLabel, getDifficultyType, getDifficultyClass,
    handleSortPoint // 导出排序方法
  };
}
