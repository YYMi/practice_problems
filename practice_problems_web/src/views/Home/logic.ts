import { ref, reactive, computed, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { useRouter } from "vue-router";

// API 引用 (请确保路径正确)
import { getSubjects, createSubject, updateSubject, deleteSubject } from "../../api/subject";
import { getCategories, createCategory, updateCategory, deleteCategory, updateCategorySort } from "../../api/category";
import { getPoints, getPointDetail, createPoint, updatePoint, deletePoint, updatePointSort } from "../../api/point";
import request from "../../utils/request"; 
import type { Subject, Category, PointSummary, PointDetail } from "../../types";

export function useHomeLogic() {
  const router = useRouter();

  // --- State ---
  const subjects = ref<Subject[]>([]);
  const currentSubject = ref<Subject | null>(null);
  const categories = ref<Category[]>([]);
  const currentCategory = ref<Category | null>(null);
  const points = ref<PointSummary[]>([]);
  const currentPoint = ref<PointDetail | null>(null);

  const drawerVisible = ref(false);
  const categoryPracticeVisible = ref(false);

  // --- 用户信息 ---
  const userInfo = reactive({
    user_code: '',
    username: '',
    nickname: '',
    email: ''
  });

  const loadUserInfo = () => {
    const savedUser = localStorage.getItem('user_info');
    if (savedUser) {
      try {
        const u = JSON.parse(savedUser);
        Object.assign(userInfo, u);
      } catch (e) { console.error("读取用户信息失败", e); }
    }
  };

  const handleLogout = () => {
    ElMessageBox.confirm("确定要退出登录吗？", "提示", {
      confirmButtonText: "退出", cancelButtonText: "取消", type: "warning"
    }).then(() => {
      localStorage.removeItem('auth_token');
      localStorage.removeItem('user_info');
      localStorage.removeItem("question_bank_state");
      ElMessage.success("已退出登录");
      router.push('/login');
    });
  };

  // --- 用户信息修改弹窗 ---
  const profileDialog = reactive({ visible: false });
  const profileForm = reactive({ nickname: '', email: '', oldPassword: '', newPassword: '' });

  const openProfileDialog = () => {
    profileForm.nickname = userInfo.nickname;
    profileForm.email = userInfo.email;
    profileForm.oldPassword = '';
    profileForm.newPassword = '';
    profileDialog.visible = true;
  };

  const submitProfileUpdate = async () => {
    if (profileForm.newPassword && !profileForm.oldPassword) return ElMessage.warning("修改密码需要输入旧密码");
    try {
      const res: any = await request.put('/user/profile', {
        nickname: profileForm.nickname, email: profileForm.email,
        old_password: profileForm.oldPassword, new_password: profileForm.newPassword
      });
      if (res.data.code === 200) {
        ElMessage.success("修改成功");
        userInfo.nickname = profileForm.nickname;
        userInfo.email = profileForm.email;
        localStorage.setItem('user_info', JSON.stringify(userInfo));
        profileDialog.visible = false;
      }
    } catch (e) { console.error(e); }
  };

  // --- 水印逻辑 ---
  const isSubjectOwner = computed(() => {
    if (!currentSubject.value || !userInfo.user_code) return false;
    // @ts-ignore
    return currentSubject.value.creatorCode === userInfo.user_code;
  });

  const subjectWatermarkText = computed(() => {
    if (!currentSubject.value) return '';
    // @ts-ignore
    const code = currentSubject.value.creatorCode || 'Unknown';
    return `Created by ${code}`;
  });

  const isPointOwner = computed(() => {
    if (!currentPoint.value || !userInfo.user_code) return false;
    // @ts-ignore
    return currentPoint.value.creatorCode === userInfo.user_code; // 假设Point也有creatorCode
  });

  // --- 辅助函数 ---
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
  const getDifficultyLabel = (diff?: number) => {
    switch (diff) { case 1: return '中等'; case 2: return '困难'; case 3: return '重点'; default: return '简单'; }
  };
  const getDifficultyClass = (diff?: number) => {
    const d = diff || 0; return `diff-item-${d}`;
  };

  // --- API 逻辑 ---
  const loadSubjects = async () => { 
    const res = await getSubjects(); 
    if (res.data.code === 200) subjects.value = res.data.data;
  };

  const handleSelectSubject = async (item: Subject) => { 
    currentSubject.value = item; currentCategory.value = null; currentPoint.value = null;
    const res = await getCategories(item.id); 
    if (res.data.code === 200) categories.value = res.data.data;
  };

  const handleSelectCategory = async (item: Category) => { 
    currentCategory.value = item; currentPoint.value = null; 
    const res = await getPoints(item.id); 
    if (res.data.code === 200) points.value = res.data.data; 
  };

  const handleSelectPoint = async (id: number) => { 
    const res = await getPointDetail(id); 
    if (res.data.code === 200) currentPoint.value = res.data.data; 
  };

  // --- CRUD ---
  // Subject
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
      if (subjectDialog.isEdit) { await updateSubject(subjectDialog.id, { name: subjectForm.name, status: 1 }); } 
      else { await createSubject({ name: subjectForm.name, status: 1 }); }
      subjectDialog.visible = false; loadSubjects(); ElMessage.success("操作成功");
    } catch (e) { ElMessage.error("操作失败"); }
  };
  const handleDeleteSubject = (subject: Subject) => {
    ElMessageBox.confirm(`确定删除科目“${subject.name}”吗？`, "警告", { type: "warning" }).then(async () => {
      await deleteSubject(subject.id); loadSubjects(); currentSubject.value = null; ElMessage.success("删除成功");
    });
  };

  // Category
  const categoryDialog = reactive({ visible: false, isEdit: false, id: 0 });
  const categoryForm = reactive({ categoryName: "", difficulty: 0 });
  const openCategoryDialog = (category?: Category) => {
    if (category) { categoryDialog.isEdit = true; categoryDialog.id = category.id; categoryForm.categoryName = category.categoryName; categoryForm.difficulty = category.difficulty || 0; } 
    else { categoryDialog.isEdit = false; categoryForm.categoryName = ""; categoryForm.difficulty = 0; }
    categoryDialog.visible = true;
  };
  const submitCategory = async () => {
    if (!categoryForm.categoryName) return ElMessage.warning("请输入名称");
    try {
      if (categoryDialog.isEdit) { await updateCategory(categoryDialog.id, { categoryName: categoryForm.categoryName, difficulty: categoryForm.difficulty }); } 
      else { if (!currentSubject.value) return; await createCategory({ subjectId: currentSubject.value.id, categoryName: categoryForm.categoryName }); }
      categoryDialog.visible = false; 
      if (currentSubject.value) handleSelectSubject(currentSubject.value); // 刷新
      ElMessage.success("操作成功");
    } catch (e) { ElMessage.error("操作失败"); }
  };
  const handleDeleteCategory = (category: Category) => {
    ElMessageBox.confirm(`确定删除分类“${category.categoryName}”吗？`, "警告", { type: "warning" }).then(async () => {
      await deleteCategory(category.id); 
      if (currentSubject.value) handleSelectSubject(currentSubject.value);
      currentCategory.value = null; ElMessage.success("删除成功");
    });
  };
  const handleSortCategory = async (category: Category, action: 'top' | 'up' | 'down') => {
    await updateCategorySort(category.id, action);
    if (currentSubject.value) handleSelectSubject(currentSubject.value);
  };

  // Point
  const createPointDialog = reactive({ visible: false });
  const createPointForm = reactive({ title: "" });
  const openCreatePointDialog = () => { createPointForm.title = ""; createPointDialog.visible = true; };
  const submitCreatePoint = async () => { 
    if (!currentCategory.value) return; 
    await createPoint({ categoryId: currentCategory.value.id, title: createPointForm.title, }); 
    createPointDialog.visible = false; handleSelectCategory(currentCategory.value); ElMessage.success("创建成功"); 
  };
  const handleDeletePoint = (point?: PointSummary | PointDetail) => {
    const target = point || currentPoint.value; if (!target) return;
    ElMessageBox.confirm(`确定删除知识点“${target.title}”吗？`, "警告", { type: "warning" }).then(async () => {
      await deletePoint(target.id); 
      if (currentCategory.value) handleSelectCategory(currentCategory.value);
      currentPoint.value = null; ElMessage.success("删除成功");
    });
  };
  const handleSortPoint = async (point: PointSummary, action: 'top' | 'up' | 'down') => {
    await updatePointSort(point.id, action);
    if (currentCategory.value) handleSelectCategory(currentCategory.value);
  };
  
  const editTitleDialog = reactive({ visible: false, title: "", id: 0, difficulty: 0 });
  const openEditTitleDialog = (point?: PointSummary | PointDetail) => {
    const target = point || currentPoint.value; if (!target) return;
    editTitleDialog.id = target.id; editTitleDialog.title = target.title; editTitleDialog.difficulty = (target as any).difficulty || 0; editTitleDialog.visible = true;
  };
  const submitEditTitle = async () => {
    if (!editTitleDialog.id) return;
    await updatePoint(editTitleDialog.id, { title: editTitleDialog.title, difficulty: editTitleDialog.difficulty });
    if (currentCategory.value) handleSelectCategory(currentCategory.value);
    if (currentPoint.value?.id === editTitleDialog.id) { currentPoint.value.title = editTitleDialog.title; }
    editTitleDialog.visible = false; ElMessage.success("修改成功");
  };

  const addLink = () => { ElMessageBox.prompt("请输入链接地址", "添加链接", { confirmButtonText: "确定" }).then(async ({ value }) => { if (!value) return; const links = [...parsedLinks.value, value]; const jsonStr = JSON.stringify(links); await updatePoint(currentPoint.value!.id, { referenceLinks: jsonStr }); currentPoint.value!.referenceLinks = jsonStr; }); };
  const removeLink = async (index: number) => { const links = [...parsedLinks.value]; links.splice(index, 1); const jsonStr = JSON.stringify(links); await updatePoint(currentPoint.value!.id, { referenceLinks: jsonStr }); currentPoint.value!.referenceLinks = jsonStr; };

  const openCategoryPractice = () => { categoryPracticeVisible.value = true; };

  onMounted(() => { loadSubjects(); loadUserInfo(); });

  return {
    subjects, currentSubject, categories, currentCategory, points, currentPoint,
    drawerVisible, categoryPracticeVisible, parsedLinks, formatUrl,
    handleSelectSubject, handleSelectCategory, handleSelectPoint, openCategoryPractice,
    openSubjectDialog, submitSubject, handleDeleteSubject, subjectDialog, subjectForm,
    openCategoryDialog, submitCategory, handleDeleteCategory, handleSortCategory, categoryDialog, categoryForm,
    openCreatePointDialog, submitCreatePoint, handleDeletePoint, createPointDialog, createPointForm,
    openEditTitleDialog, submitEditTitle, editTitleDialog,
    addLink, removeLink, getDifficultyLabel, getDifficultyClass, handleSortPoint,
    userInfo, handleLogout, profileDialog, profileForm, openProfileDialog, submitProfileUpdate,
    isSubjectOwner, subjectWatermarkText, isPointOwner,loadSubjects,
  };
}
