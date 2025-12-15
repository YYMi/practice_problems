import { ref, reactive, onMounted, computed } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { useRouter } from 'vue-router';

// API 引用
import { getSubjects, createSubject, updateSubject, deleteSubject } from '../../api/subject';
import { getCategories, createCategory, updateCategory, deleteCategory, updateCategorySort } from '../../api/category';
import { getPoints, getPointDetail, createPoint, updatePoint, deletePoint, updatePointSort } from '../../api/point';
// ★★★ 1. 引入通用 request 工具，不再需要 api/user.ts ★★★
import request from '../../utils/request'; 
import type { Subject, Category, PointSummary, PointDetail } from '../../types';

export function useHomeLogic() {
  const router = useRouter();

  // ================= 数据状态 =================
  const subjects = ref<Subject[]>([]);
  const currentSubject = ref<Subject | null>(null);
  
  const categories = ref<Category[]>([]);
  const currentCategory = ref<Category | null>(null);
  
  // 分类分页状态
  const categoryPage = ref(1);
  const categoryPageSize = ref(11);
  const categoryTotal = ref(0);
  
  // 知识点分页状态
  const pointPage = ref(1);
  const pointPageSize = ref(11);
  const pointTotal = ref(0);
  
  const points = ref<PointSummary[]>([]);
  const currentPoint = ref<PointDetail | null>(null);
  const currentPointBindings = ref<any[]>([]); // 当前知识点的绑定列表
  
  // 知识点信息缓存: pointId -> {title, categoryName}
  const pointsInfoMap = ref<Map<number, {title: string; categoryName: string}>>(new Map());
  
  // 导航历史栈: 用于知识点链接跳转的返回功能（包含滚动位置）
  const navigationStack = ref<{categoryId: number; pointId: number; scrollTop: number}[]>([]);

  // 弹窗控制状态
  const drawerVisible = ref(false); // 题目抽屉
  const categoryPracticeVisible = ref(false); // 分类刷题抽屉

  // 用户信息
  const userInfo = ref<any>({});

  // ================= 弹窗表单状态 =================
  // 科目
  const subjectDialog = reactive({ visible: false, isEdit: false, id: 0 });
  const subjectForm = reactive({ name: '' });
  
  // 分类
  const categoryDialog = reactive({ visible: false, isEdit: false, id: 0 });
  const categoryForm = reactive({ categoryName: '', difficulty: 0 });
  
  // 知识点创建
  const createPointDialog = reactive({ visible: false });
  const createPointForm = reactive({ title: '', difficulty: 0 });
  
  // 知识点编辑
  const editTitleDialog = reactive({ visible: false, id: 0, title: '', difficulty: 0 });
  
  // 用户资料
  const profileDialog = reactive({ visible: false });
  const profileForm = reactive({ nickname: '', email: '', oldPassword: '', newPassword: '' });

  // ================= 核心逻辑：加载与恢复 =================

  // 1. 加载科目 (入口)
  const loadSubjects = async () => {
    try {
      const res = await getSubjects();
      if (res.data && res.data.code === 200) {
        subjects.value = res.data.data;
        
        // ★★★ 自动恢复逻辑：科目 ★★★
        const lastSubId = localStorage.getItem('last_subject_id');
        if (lastSubId) {
          const target = subjects.value.find((s: any) => s.id == lastSubId);
          if (target) {
            currentSubject.value = target;
            await loadCategories(true); // 链式恢复下一级
          } else {
            clearStorageFrom('subject'); // 没找到(过期了)，清除缓存
          }
        }
      }
    } catch (e) { console.error(e); }
  };

  // 2. 加载分类（支持分页）
  const loadCategories = async (isRestore = false) => {
    if (!currentSubject.value) return;
    try {
      const res = await getCategories(currentSubject.value.id, categoryPage.value, categoryPageSize.value);
      if (res.data && res.data.code === 200) {
        const responseData = res.data.data;
        categories.value = responseData.list || [];
        categoryTotal.value = responseData.total || 0;
        
        // ★★★ 自动恢复逻辑：分类 ★★★
        if (isRestore) {
          const lastCatId = localStorage.getItem('last_category_id');
          if (lastCatId) {
            const target = categories.value.find((c: any) => c.id == lastCatId);
            if (target) {
              currentCategory.value = target;
              await loadPoints(true); // 链式恢复下一级
            } else {
              clearStorageFrom('category');
            }
          }
        }
      }
    } catch (e) { console.error(e); }
  };

  // 分类分页处理
  const handleCategoryPageChange = (page: number) => {
    categoryPage.value = page;
    loadCategories(false);
  };

  // 3. 加载知识点列表（支持分页）
  const loadPoints = async (isRestore = false) => {
    if (!currentCategory.value) return;
    try {
      const res = await getPoints(currentCategory.value.id, pointPage.value, pointPageSize.value);
      if (res.data && res.data.code === 200) {
        const responseData = res.data.data;
        points.value = responseData.list || [];
        pointTotal.value = responseData.total || 0;
        
        // 填充知识点缓存
        const categoryName = currentCategory.value.categoryName;
        for (const p of points.value) {
          pointsInfoMap.value.set(p.id, { title: p.title, categoryName });
        }

        // ★★★ 自动恢复逻辑：知识点 ★★★
        if (isRestore) {
          const lastPointId = localStorage.getItem('last_point_id');
          if (lastPointId) {
            // 注意：列表里的 point 只是摘要，选中后需要调详情接口
            const targetSummary = points.value.find((p: any) => p.id == lastPointId);
            if (targetSummary) {
              handleSelectPoint(targetSummary.id); // 调用详情接口
            } else {
              clearStorageFrom('point');
            }
          }
        }
      }
    } catch (e) { console.error(e); }
  };

  // 知识点分页处理
  const handlePointPageChange = (page: number) => {
    pointPage.value = page;
    loadPoints(false);
  };

  // 辅助：清除缓存 (级联清除)
  const clearStorageFrom = (level: 'subject' | 'category' | 'point') => {
    if (level === 'subject') {
      localStorage.removeItem('last_subject_id');
      currentSubject.value = null;
    }
    if (level === 'subject' || level === 'category') {
      localStorage.removeItem('last_category_id');
      categories.value = [];
      currentCategory.value = null;
    }
    if (level === 'subject' || level === 'category' || level === 'point') {
      localStorage.removeItem('last_point_id');
      points.value = [];
      currentPoint.value = null;
      currentPointBindings.value = [];
    }
  };

  // ================= 用户交互事件 (点击选择) =================

  // 选择科目
  const handleSelectSubject = (item: Subject) => {
    if (currentSubject.value?.id === item.id) return;
    currentSubject.value = item;
    localStorage.setItem('last_subject_id', String(item.id)); // 存
    clearStorageFrom('category'); // 清除下级状态
    loadCategories(false);
  };

  // 选择分类
  const handleSelectCategory = (item: Category) => {
    if (currentCategory.value?.id === item.id) return;
    currentCategory.value = item;
    localStorage.setItem('last_category_id', String(item.id)); // 存
    clearStorageFrom('point'); // 清除下级状态
    navigationStack.value = []; // 清空导航栈
    loadPoints(false);
  };

  // 选择知识点 (获取详情) - 主动点击，清空导航栈
  const handleSelectPoint = async (id: number, fromNavigation = false) => {
    // 如果不是从导航跳转来的，清空导航栈
    if (!fromNavigation) {
      navigationStack.value = [];
    }
    
    try {
      const res = await getPointDetail(id);
      if (res.data && res.data.code === 200) {
        const data = res.data.data as any;
        // 兼容新旧两种返回结构
        if (data.point) {
          // 新结构: { point: {...}, bindings: [...] }
          currentPoint.value = data.point;
          currentPointBindings.value = data.bindings || [];
        } else {
          // 旧结构: 直接就是 point 对象
          currentPoint.value = data;
          currentPointBindings.value = [];
        }
        localStorage.setItem('last_point_id', String(id)); // 存
      }
    } catch (e) { console.error(e); }
  };
  
  // 通过绑定链接跳转到知识点
  const navigateToPoint = async (targetPointId: number, targetCategoryId: number) => {
    // 1. 记录当前位置到栈中（包含滚动位置）
    if (currentCategory.value && currentPoint.value) {
      // 获取当前滚动位置
      const scrollContainer = document.querySelector('.html-preview') || document.querySelector('.content-box');
      const scrollTop = scrollContainer?.scrollTop || 0;
      
      const currentPos = { categoryId: currentCategory.value.id, pointId: currentPoint.value.id, scrollTop };
      // 检查栈顶是否已经是当前位置
      const lastInStack = navigationStack.value[navigationStack.value.length - 1];
      if (!lastInStack || lastInStack.pointId !== currentPos.pointId) {
        navigationStack.value.push(currentPos);
      }
    }
    
    // 2. 切换到目标分类（如果需要）
    if (targetCategoryId && currentCategory.value?.id !== targetCategoryId) {
      const targetCategory = categories.value.find((c: any) => c.id === targetCategoryId);
      if (targetCategory) {
        currentCategory.value = targetCategory;
        localStorage.setItem('last_category_id', String(targetCategoryId));
        // 加载目标分类的知识点列表
        const res = await getPoints(targetCategoryId, pointPage.value, pointPageSize.value);
        if (res.data && res.data.code === 200) {
          const responseData = res.data.data;
          points.value = responseData.list || [];
          pointTotal.value = responseData.total || 0;
          // 填充缓存
          const categoryName = targetCategory.categoryName;
          for (const p of points.value) {
            pointsInfoMap.value.set(p.id, { title: p.title, categoryName });
          }
        }
      }
    }
    
    // 3. 跳转到目标知识点
    await handleSelectPoint(targetPointId, true);
  };
  
  // 返回上一个知识点
  const navigateBack = async () => {
    if (navigationStack.value.length === 0) return;
    
    // 弹出栈顶（返回目标）
    const target = navigationStack.value.pop()!;
    
    // 切换到目标分类（如果需要）
    if (target.categoryId !== currentCategory.value?.id) {
      const targetCategory = categories.value.find((c: any) => c.id === target.categoryId);
      if (targetCategory) {
        currentCategory.value = targetCategory;
        localStorage.setItem('last_category_id', String(target.categoryId));
        // 加载目标分类的知识点列表
        const res = await getPoints(target.categoryId, pointPage.value, pointPageSize.value);
        if (res.data && res.data.code === 200) {
          const responseData = res.data.data;
          points.value = responseData.list || [];
          pointTotal.value = responseData.total || 0;
          const categoryName = targetCategory.categoryName;
          for (const p of points.value) {
            pointsInfoMap.value.set(p.id, { title: p.title, categoryName });
          }
        }
      }
    }
    
    // 跳转到目标知识点
    await handleSelectPoint(target.pointId, true);
    
    // 恢复滚动位置
    if (target.scrollTop) {
      setTimeout(() => {
        const scrollContainer = document.querySelector('.html-preview') || document.querySelector('.content-box');
        if (scrollContainer) {
          scrollContainer.scrollTop = target.scrollTop;
        }
      }, 100); // 延迟等待DOM更新
    }
  };
  
  // 是否可以返回
  const canGoBack = computed(() => navigationStack.value.length > 0);

  // ================= CRUD 操作 =================

  // --- 科目 ---
  const openSubjectDialog = (item?: Subject) => {
    if (item) { subjectForm.name = item.name; subjectDialog.id = item.id; subjectDialog.isEdit = true; }
    else { subjectForm.name = ''; subjectDialog.isEdit = false; }
    subjectDialog.visible = true;
  };
  
  const submitSubject = async () => {
    if (!subjectForm.name) return ElMessage.warning('请输入名称');
    try {
      if (subjectDialog.isEdit) {
        await updateSubject(subjectDialog.id, { name: subjectForm.name, status: 1 });
      } else {
        await createSubject({ name: subjectForm.name, status: 1 });
      }
      subjectDialog.visible = false;
      loadSubjects(); // 重新加载列表
      ElMessage.success('操作成功');
    } catch (e) { ElMessage.error('操作失败'); }
  };

  const handleDeleteSubject = (item: Subject) => {
    ElMessageBox.confirm(`确定删除科目“${item.name}”吗？`, '提示', { type: 'warning' }).then(async () => {
      await deleteSubject(item.id);
      ElMessage.success('已删除');
      if (currentSubject.value?.id === item.id) clearStorageFrom('subject');
      loadSubjects();
    });
  };

  // --- 分类 ---
  const openCategoryDialog = (item?: Category) => {
    if (item) { categoryForm.categoryName = item.categoryName; categoryForm.difficulty = item.difficulty || 0; categoryDialog.id = item.id; categoryDialog.isEdit = true; }
    else { categoryForm.categoryName = ''; categoryForm.difficulty = 0; categoryDialog.isEdit = false; }
    categoryDialog.visible = true;
  };

  const submitCategory = async () => {
    if (!categoryForm.categoryName) return ElMessage.warning('请输入名称');
    if (!currentSubject.value) return;
    try {
      if (categoryDialog.isEdit) {
        await updateCategory(categoryDialog.id, { ...categoryForm, subjectId: currentSubject.value.id });
      } else {
        await createCategory({ ...categoryForm, subjectId: currentSubject.value.id });
      }
      categoryDialog.visible = false;
      ElMessage.success('操作成功');
      loadCategories(false);
    } catch (e) { ElMessage.error('操作失败'); }
  };

  const handleDeleteCategory = (item: Category) => {
    ElMessageBox.confirm(`确定删除分类“${item.categoryName}”吗？`, '提示', { type: 'warning' }).then(async () => {
      await deleteCategory(item.id);
      ElMessage.success('已删除');
      if (currentCategory.value?.id === item.id) clearStorageFrom('category');
      loadCategories(false);
    });
  };

  const handleSortCategory = async (item: Category, action: 'top' | 'up' | 'down') => {
    await updateCategorySort(item.id, action);
    loadCategories(false);
  };

  // --- 知识点 ---
  const openCreatePointDialog = () => { createPointForm.title = ''; createPointForm.difficulty = 0; createPointDialog.visible = true; };
  
  const submitCreatePoint = async () => {
    if (!createPointForm.title || !currentCategory.value) return;
    try {
      await createPoint({ categoryId: currentCategory.value.id, title: createPointForm.title, content: '', difficulty: createPointForm.difficulty });
      ElMessage.success('创建成功');
      createPointDialog.visible = false;
      loadPoints(false);
    } catch (e) { ElMessage.error('失败'); }
  };

  const handleDeletePoint = (item?: any) => {
    const target = item || currentPoint.value;
    if (!target) return;
    ElMessageBox.confirm(`确定删除知识点“${target.title}”吗？`, '提示', { type: 'warning' }).then(async () => {
      await deletePoint(target.id);
      ElMessage.success('已删除');
      if (currentPoint.value?.id === target.id) clearStorageFrom('point');
      loadPoints(false);
    });
  };

  const handleSortPoint = async (item: any, action: 'top' | 'up' | 'down') => {
    await updatePointSort(item.id, action);
    loadPoints(false);
  };

  // 知识点标题编辑
  const openEditTitleDialog = (p?: any) => {
    const target = p || currentPoint.value;
    if (target) { editTitleDialog.id = target.id; editTitleDialog.title = target.title; editTitleDialog.difficulty = target.difficulty || 0; editTitleDialog.visible = true; }
  };

  const submitEditTitle = async () => {
    if (!editTitleDialog.id) return;
    await updatePoint(editTitleDialog.id, { title: editTitleDialog.title, difficulty: editTitleDialog.difficulty });
    editTitleDialog.visible = false;
    ElMessage.success('修改成功');
    loadPoints(false); // 刷新列表
    if (currentPoint.value?.id === editTitleDialog.id) {
      // 同步更新当前详情页
      currentPoint.value.title = editTitleDialog.title;
      (currentPoint.value as any).difficulty = editTitleDialog.difficulty;
    }
  };

    // ============================================================
  // ★★★ 新增：移动知识点逻辑 ★★★
  // ============================================================
  const handleMovePoint = async ({ pointId, targetCategoryId }: { pointId: number, targetCategoryId: number }) => {
    try {
      // 1. 调用接口更新分类ID
      await updatePoint(pointId, { categoryId: targetCategoryId });
      
      ElMessage.success('移动成功');

      // 2. 刷新当前列表（因为移走了，它应该从当前列表中消失）
      // loadPoints(false) 会重新拉取当前分类下的数据
      await loadPoints(false);

      // 3. 边界情况处理：
      // 如果当前选中的知识点正是被移走的那个，我们需要清空选中状态，或者重新获取详情
      if (currentPoint.value?.id === pointId) {
        // 简单处理：直接清空详情，避免数据显示不一致
        currentPoint.value = null;
      }
    } catch (e) {
      console.error(e);
      ElMessage.error('移动失败');
    }
  };

  // --- 用户相关 ---
  const openProfileDialog = () => {
    profileForm.nickname = userInfo.value.nickname;
    profileForm.email = userInfo.value.email;
    profileForm.oldPassword = '';
    profileForm.newPassword = '';
    profileDialog.visible = true;
  };

  // ★★★ 修改个人信息 ★★★
  // 注意：这里增加了参数 payload，这是 Header 组件传过来的数据
  const submitProfileUpdate = async (payload: any) => {
    
    // 1. 移除这里的校验逻辑
    // 因为 Header.vue 里的 handleSaveProfile 已经校验过长度和空值了
    // if (profileForm.newPassword && !profileForm.oldPassword) ... (这行删掉)
    
    try {
      // 2. 直接发送 payload
      // payload 里已经包含了：nickname, email, old_password(MD5), new_password(MD5)
      const res: any = await request.put('/user/profile', payload);

      const code = res.code || (res.data && res.data.code);
      
      if (code === 200) {
        ElMessage.success("修改成功");
        
        // 3. 更新内存状态 (使用提交的数据更新显示)
        userInfo.value.nickname = payload.nickname;
        userInfo.value.email = payload.email;
        
        // 4. 同步更新缓存
        const storedUser = JSON.parse(localStorage.getItem('user_info') || '{}');
        storedUser.nickname = payload.nickname;
        storedUser.email = payload.email;
        localStorage.setItem('user_info', JSON.stringify(storedUser));
        
        // 关闭弹窗
        profileDialog.visible = false;
      } else {
        // 错误提示交给拦截器或这里
        ElMessage.error(res.msg || "修改失败");
      }
    } catch (e) { 
      console.error(e); 
    }
  };

  const handleLogout = () => {
    ElMessageBox.confirm("确定要退出登录吗？", "提示", { type: "warning" }).then(() => {
      localStorage.clear();
      router.push('/login');
      ElMessage.success("已退出登录");
    });
  };

  // ================= 辅助计算属性 =================
  
  const isSubjectOwner = computed(() => currentSubject.value?.creatorCode === userInfo.value.user_code);
  
  const isPointOwner = computed(() => {
    if (currentPoint.value && (currentPoint.value as any).creatorCode) {
      return (currentPoint.value as any).creatorCode === userInfo.value.user_code;
    }
    return isSubjectOwner.value;
  });

  const subjectWatermarkText = computed(() => currentSubject.value ? `Created by ${currentSubject.value.creatorCode}` : '');

  const parsedLinks = computed(() => {
    if (!currentPoint.value || !currentPoint.value.referenceLinks) return [];
    try { return JSON.parse(currentPoint.value.referenceLinks); } catch { return []; }
  });

  // 辅助函数
  const addLink = () => {
    ElMessageBox.prompt("请输入链接地址", "添加链接", { confirmButtonText: "确定" }).then(async ({ value }) => {
      if (!value) return;
      const links = [...parsedLinks.value, value];
      const jsonStr = JSON.stringify(links);
      await updatePoint(currentPoint.value!.id, { referenceLinks: jsonStr });
      currentPoint.value!.referenceLinks = jsonStr;
    });
  };

  const removeLink = async (index: number) => {
    const links = [...parsedLinks.value];
    links.splice(index, 1);
    const jsonStr = JSON.stringify(links);
    await updatePoint(currentPoint.value!.id, { referenceLinks: jsonStr });
    currentPoint.value!.referenceLinks = jsonStr;
  };

  const formatUrl = (url: string) => {
    if (!url) return "";
    url = url.trim();
    if (!url.startsWith("http://") && !url.startsWith("https://")) return `http://${url}`;
    return url;
  };

  // ============================================================
  // ★★★ 新增：保存视频逻辑 (请插入在 formatUrl 函数下方) ★★★
  // ============================================================
  const handleSaveVideo = async (videoUrl: string) => {
    // 1. 安全检查：如果没有选中知识点，直接退出
    if (!currentPoint.value) return;
    
    try {
      // 2. 调用后端 API 更新数据库
      // 注意：这里传给后端的字段是 videoUrl，请确保后端 Point 实体类里有这个字段
      await updatePoint(currentPoint.value.id, { videoUrl: videoUrl });

      // 3. 更新本地前端状态 (这样不需要刷新页面，界面上的视频就会立刻出现)
      // 注意：TypeScript 可能会报 videoUrl 不存在，需要去 types.ts 里给 PointDetail 加这个字段
      // 如果报错，可以暂时用 (currentPoint.value as any).videoUrl = videoUrl 绕过
      (currentPoint.value as any).videoUrl = videoUrl;

      ElMessage.success(videoUrl ? '视频设置成功' : '视频已移除');
    } catch (e) {
      console.error(e);
      ElMessage.error('视频保存失败');
    }
  };

  const getDifficultyLabel = (diff?: number) => ['简单', '中等', '困难', '重点'][diff || 0] || '简单';
  const getDifficultyClass = (diff?: number) => `diff-${diff || 0}`;
  const openCategoryPractice = () => { categoryPracticeVisible.value = true; };

  // ================= 初始化 =================
  onMounted(() => {
    const userStr = localStorage.getItem('user_info');
    if (userStr) {
      userInfo.value = JSON.parse(userStr);
      loadSubjects(); // 启动应用
    } else {
      // 没有用户信息，踢回登录
      router.push('/login');
    }
  });

  // 返回所有状态和方法
  return {
    subjects, currentSubject, categories, currentCategory, points, currentPoint, currentPointBindings, pointsInfoMap,
    categoryPage, categoryPageSize, categoryTotal, // 分类分页状态
    pointPage, pointPageSize, pointTotal, // 知识点分页状态
    subjectDialog, subjectForm, categoryDialog, categoryForm, createPointDialog, createPointForm,
    profileDialog, profileForm, editTitleDialog, drawerVisible, categoryPracticeVisible, userInfo,
    parsedLinks, isSubjectOwner, isPointOwner, subjectWatermarkText,
    handleSelectSubject, openSubjectDialog, submitSubject, handleDeleteSubject,
    handleSelectCategory, openCategoryDialog, submitCategory, handleDeleteCategory, handleSortCategory,
    handleCategoryPageChange, // 分类分页方法
    handleSelectPoint, openCreatePointDialog, submitCreatePoint, handleDeletePoint, handleSortPoint,
    handlePointPageChange, // 知识点分页方法
    openProfileDialog, submitProfileUpdate, handleLogout,
    openEditTitleDialog, submitEditTitle, openCategoryPractice,
    addLink, removeLink, formatUrl, getDifficultyLabel, getDifficultyClass, loadSubjects, handleMovePoint, handleSaveVideo,
    navigateToPoint, navigateBack, canGoBack, // 导航历史
  };
}
