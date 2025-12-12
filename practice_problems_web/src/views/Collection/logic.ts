import { ref, computed, watch } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { 
  getCollections, 
  createCollection, 
  updateCollection, 
  deleteCollection, 
  getCollectionPoints, 
  getCollectionPointDetail, 
  removePointFromCollection, 
  updateCollectionItemsOrder,
  setCollectionPermission,
  addCollectionPermission,
  getCollectionPermissions,
  updateCollectionPermission,
  deleteCollectionPermission,
  findPointInCollections,
  type Collection, 
  type CollectionPoint,
  type CollectionPermission
} from '../../api/collection';

// 颜色调色板：50种协调的中等饱和度颜色（文字清晰，视觉舒适）
const COLOR_PALETTE = [
  '#E8B4D4', '#DFA4C8', '#D694BC', '#CD84B0', '#C474A4',
  '#D4A5D4', '#C895C8', '#BC85BC', '#B075B0', '#A465A4',
  '#A4B4E8', '#94A4DF', '#8494D6', '#7484CD', '#6474C4',
  '#94C4E8', '#84B4DF', '#74A4D6', '#6494CD', '#5484C4',
  '#94D4C4', '#84C8B8', '#74BCAC', '#64B0A0', '#54A494',
  '#B4E8D4', '#A4DFC8', '#94D6BC', '#84CDB0', '#74C4A4',
  '#E8D4A4', '#DFC894', '#D6BC84', '#CDB074', '#C4A464',
  '#E8C4A4', '#DFB894', '#D6AC84', '#CDA074', '#C49464',
  '#D4B4A4', '#C8A894', '#BC9C84', '#B09074', '#A48464',
  '#E8A4B4', '#DF94A4', '#D68494', '#CD7484', '#C46474'
];

// 字符串哈希函数：将字符串转换为索引
const stringToColorIndex = (str: string): number => {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  return Math.abs(hash) % COLOR_PALETTE.length;
};

// 将 hex 颜色转换为 rgba（带透明度）
const hexToRgba = (hex: string, alpha: number): string => {
  const r = parseInt(hex.slice(1, 3), 16);
  const g = parseInt(hex.slice(3, 5), 16);
  const b = parseInt(hex.slice(5, 7), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
};

// 获取科目颜色（60% 透明度）
const getSubjectColor = (subjectName: string): string => {
  if (!subjectName) return 'rgba(245, 247, 250, 0.6)';
  const hex = COLOR_PALETTE[stringToColorIndex(subjectName)];
  return hexToRgba(hex, 0.6);
};

// 获取分类颜色（60% 透明度）
const getCategoryColor = (categoryName: string): string => {
  if (!categoryName) return 'rgba(245, 247, 250, 0.6)';
  const hex = COLOR_PALETTE[stringToColorIndex(categoryName)];
  return hexToRgba(hex, 0.6);
};

export function useCollectionLogic() {
  const router = useRouter();

  // 集合数据
  const collections = ref<Collection[]>([]);
  const currentCollectionId = ref(0);
  const currentCollectionName = ref('');
  const loading = ref(false);

  // 单词本状态
  const wordbookVisible = ref(false);

  // 计算属性：当前集合是否属于当前用户
  const isCollectionOwner = computed(() => {
    const collection = collections.value.find(c => c.id === currentCollectionId.value);
    return collection ? collection.isOwner || false : false;
  });

  // 知识点列表数据
  const collectionPoints = ref<CollectionPoint[]>([]);
  const pointsLoading = ref(false);
  const pointsPage = ref(1);
  const pointsPageSize = ref(15);
  const pointsTotal = ref(0);
  const isEditMode = ref(false); // 列表编辑模式
  
  // 模式切换：普通模式 / 随机模式
  const isRandomMode = ref(false);
  const allPointsCache = ref<CollectionPoint[]>([]); // 随机模式下缓存所有知识点
  
  // 排序状态：'asc' 升序，'desc' 降序，null 默认（不排序）
  const categorySortOrder = ref<'asc' | 'desc' | null>(null); // 分类排序
  const pointSortOrder = ref<'asc' | 'desc' | null>(null); // 知识点排序

  // 选中的知识点详情
  const selectedPointDetail = ref<any>(null);
  const selectedPointId = ref<number>(0);
  const currentPointBindings = ref<any[]>([]); // 当前知识点的绑定关系
  
  // 导航栈（用于返回功能）
  interface NavigationState {
    collectionId: number;
    pointId: number;
    page: number;
    scrollTop: number;
  }
  const navigationStack = ref<NavigationState[]>([]);

  // 练习抽屉状态
  const practiceDrawerVisible = ref(false);
  const pointPracticeDrawerVisible = ref(false);

  // 开始刷题
  const handleStartPractice = () => {
    if (!currentCollectionId.value || collectionPoints.value.length === 0) {
      ElMessage.warning('当前集合没有知识点，无法开始刷题');
      return;
    }
    practiceDrawerVisible.value = true;
  };

  // 创建集合相关
  const showCreateDialog = ref(false);
  const createLoading = ref(false);
  const createForm = ref({
    name: ''
  });

  // 编辑集合相关
  const showEditDialog = ref(false);
  const editLoading = ref(false);
  const editForm = ref({
    id: 0,
    name: ''
  });

  // 获取集合列表
  const fetchCollections = async () => {
    loading.value = true;
    try {
      const res = await getCollections();
      if (res.data.code === 200) {
        collections.value = res.data.data || [];
        // 如果有集合，默认选中第一个
        if (collections.value.length > 0) {
          currentCollectionId.value = collections.value[0].id;
        }
      } else {
        ElMessage.error(res.data.msg || '获取集合列表失败');
      }
    } catch (error: any) {
      console.error('获取集合列表失败:', error);
      ElMessage.error(error.response?.data?.msg || '获取集合列表失败，请稍后重试');
    } finally {
      loading.value = false;
    }
  };

  const selectCollection = (id: number) => {
    currentCollectionId.value = id;
    // 找到当前集合的名称
    const collection = collections.value.find(c => c.id === id);
    currentCollectionName.value = collection ? collection.name : '';
    // 切换集合时，重置分页并加载知识点列表
    pointsPage.value = 1;
    // 清空缓存
    allPointsCache.value = [];
    fetchCollectionPoints();
  };

  // 获取集合的知识点列表
  const fetchCollectionPoints = async () => {
    if (!currentCollectionId.value) {
      collectionPoints.value = [];
      pointsTotal.value = 0;
      return;
    }

    pointsLoading.value = true;
    try {
      // 随机模式：获取所有知识点（不分页，后端最大支持 10000 条）
      if (isRandomMode.value) {
        const res = await getCollectionPoints(currentCollectionId.value, 1, 10000);
        
        if (res.data.code === 200) {
          allPointsCache.value = res.data.data.list || [];
          // 随机打乱顺序
          collectionPoints.value = shuffleArray([...allPointsCache.value]);
          pointsTotal.value = 0; // 随机模式不显示分页
          // 重置排序状态
          categorySortOrder.value = null;
          pointSortOrder.value = null;
        } else {
          ElMessage.error(res.data.msg || '获取知识点列表失败');
        }
      } else {
        // 普通模式：分页获取
        const res = await getCollectionPoints(currentCollectionId.value, pointsPage.value, pointsPageSize.value);
        
        if (res.data.code === 200) {
          collectionPoints.value = res.data.data.list || [];
          pointsTotal.value = res.data.data.total || 0;
          // 重置排序状态
          categorySortOrder.value = null;
          pointSortOrder.value = null;
        } else {
          ElMessage.error(res.data.msg || '获取知识点列表失败');
        }
      }
    } catch (error: any) {
      console.error('获取知识点列表失败:', error);
      ElMessage.error(error.response?.data?.msg || '获取知识点列表失败，请稍后重试');
    } finally {
      pointsLoading.value = false;
    }
  };
  
  //  应用排序
  const applySorting = () => {
    if (!categorySortOrder.value && !pointSortOrder.value) return;
    
    collectionPoints.value.sort((a, b) => {
      // 第一优先级：按分类排序
      if (categorySortOrder.value) {
        const catA = a.categoryName.toLowerCase();
        const catB = b.categoryName.toLowerCase();
        const catCompare = categorySortOrder.value === 'asc' 
          ? catA.localeCompare(catB, 'zh-CN')
          : catB.localeCompare(catA, 'zh-CN');
        
        // 如果分类不同，直接返回分类比较结果
        if (catCompare !== 0) return catCompare;
        
        // 如果分类相同，继续比较知识点标题
      }
      
      // 第二优先级：按知识点标题排序（同分类内或仅知识点排序时）
      if (pointSortOrder.value) {
        const titleA = a.title.toLowerCase();
        const titleB = b.title.toLowerCase();
        return pointSortOrder.value === 'asc'
          ? titleA.localeCompare(titleB, 'zh-CN')
          : titleB.localeCompare(titleA, 'zh-CN');
      }
      
      return 0;
    });
  };
  
  // 切换分类排序
  const toggleCategorySort = () => {
    if (categorySortOrder.value === null) {
      categorySortOrder.value = 'asc';
      applySorting();
    } else if (categorySortOrder.value === 'asc') {
      categorySortOrder.value = 'desc';
      applySorting();
    } else {
      categorySortOrder.value = null;
      // 取消分类排序时，同时取消知识点排序
      pointSortOrder.value = null;
      fetchCollectionPoints();
    }
  };
  
  // 切换知识点排序
  const togglePointSort = () => {
    // 必须先启用分类排序，才能启用知识点排序
    if (!categorySortOrder.value) {
      ElMessage.warning('请先启用分类排序');
      return;
    }
    
    if (pointSortOrder.value === null) {
      pointSortOrder.value = 'asc';
    } else if (pointSortOrder.value === 'asc') {
      pointSortOrder.value = 'desc';
    } else {
      pointSortOrder.value = null;
    }
    applySorting();
  };
  
  // 随机打乱数组
  const shuffleArray = <T>(array: T[]): T[] => {
    const result = [...array];
    for (let i = result.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [result[i], result[j]] = [result[j], result[i]];
    }
    return result;
  };
  
  // 切换模式
  const toggleMode = () => {
    isRandomMode.value = !isRandomMode.value;
    // 切换模式时重新加载
    if (isRandomMode.value) {
      pointsPage.value = 1; // 重置分页
    }
    // 清空当前选中的知识点
    selectedPointDetail.value = null;
    selectedPointId.value = 0;
    fetchCollectionPoints();
  };
  
  // 刷新随机顺序（仅在随机模式下）
  const refreshRandomOrder = () => {
    if (!isRandomMode.value || allPointsCache.value.length === 0) return;
    collectionPoints.value = shuffleArray([...allPointsCache.value]);
    // 清空当前选中的知识点
    selectedPointDetail.value = null;
    selectedPointId.value = 0;
    // 重置排序状态
    categorySortOrder.value = null;
    pointSortOrder.value = null;
    ElMessage.success('已刷新顺序');
  };
  
  // 重新加载（从后端重新获取）
  const reloadPoints = () => {
    // 清空当前选中的知识点
    selectedPointDetail.value = null;
    selectedPointId.value = 0;
    fetchCollectionPoints();
    ElMessage.success('已重新加载');
  };

  // 选中知识点，加载详情（使用集合专用接口）
  const selectPoint = async (pointId: number) => {
    selectedPointId.value = pointId; // 设置选中ID
    
    if (!currentCollectionId.value) {
      ElMessage.error('请先选择集合');
      return;
    }
    
    try {
      const res = await getCollectionPointDetail(currentCollectionId.value, pointId);
      console.log('获取知识点详情响应:', res.data);
      if (res.data.code === 200) {
        // 后端返回的是 { point: {...}, bindings: [...] }
        const pointData = res.data.data.point;
        const bindingsData = res.data.data.bindings || [];
        console.log('知识点原始数据:', pointData);
        console.log('绑定数据:', bindingsData);
        
        // 确保必要字段有默认值（兼容驼峰和下划线命名）
        selectedPointDetail.value = {
          ...pointData,
          id: pointData.id || 0,
          categoryId: pointData.categoryId || pointData.category_id || 0,
          localImageNames: pointData.localImageNames || pointData.local_image_names || '[]',
          content: pointData.content || '',
          videoUrl: pointData.videoUrl || pointData.video_url || '[]',
          referenceLinks: pointData.referenceLinks || pointData.reference_links || '[]',
          title: pointData.title || '未命名知识点',
          difficulty: pointData.difficulty || 0
        };
        
        // 保存绑定数据
        currentPointBindings.value = bindingsData;
        
        console.log('处理后的知识点数据:', selectedPointDetail.value);
        console.log('绑定关系数据:', currentPointBindings.value);
      } else {
        ElMessage.error(res.data.msg || '获取知识点详情失败');
      }
    } catch (error: any) {
      console.error('获取知识点详情失败:', error);
      ElMessage.error(error.response?.data?.msg || '获取详情失败');
    }
  };

  // 通过绑定链接跳转到知识点（集合页面版本）
  const navigateToPointFromBinding = async (targetPointId: number) => {
    console.log('点击绑定跳转，目标知识点ID:', targetPointId);
    
    // 1. 先查本地缓存：判断目标知识点是否在当前显示的知识点列表中
    const foundInList = collectionPoints.value.find(p => p.pointId === targetPointId);
    
    if (foundInList) {
      console.log('在当前列表中找到了，直接跳转');
      // 记录当前位置到导航栈
      if (currentCollectionId.value && selectedPointId.value) {
        const scrollContainer = document.querySelector('.html-preview') || document.querySelector('.content-box');
        const scrollTop = scrollContainer ? scrollContainer.scrollTop : 0;
        
        navigationStack.value.push({
          collectionId: currentCollectionId.value,
          pointId: selectedPointId.value,
          page: pointsPage.value,
          scrollTop
        });
      }
      
      // 直接跳转
      await selectPoint(targetPointId);
      return;
    }
    
    console.log('当前列表没找到，请求后端查找');
    
    // 2. 本地没找到，请求后端查找
    try {
      const res = await findPointInCollections(targetPointId, currentCollectionId.value);
      
      if (res.data.code === 200) {
        const data = res.data.data;
        
        if (!data.found) {
          // 没找到
          ElMessage.warning(data.message || '知识点不存在或作者未分享');
          return;
        }
        
        console.log('后端返回数据:', data);
        
        // 记录当前位置到导航栈
        if (currentCollectionId.value && selectedPointId.value) {
          const scrollContainer = document.querySelector('.html-preview') || document.querySelector('.content-box');
          const scrollTop = scrollContainer ? scrollContainer.scrollTop : 0;
          
          navigationStack.value.push({
            collectionId: currentCollectionId.value,
            pointId: selectedPointId.value,
            page: pointsPage.value,
            scrollTop
          });
        }
        
        // 判断是否需要切换集合
        if (data.collectionId !== currentCollectionId.value) {
          console.log('需要切换集合，从id:', currentCollectionId.value, '到:', data.collectionId);
          // 切换集合
          currentCollectionId.value = data.collectionId!;
          const collection = collections.value.find(c => c.id === data.collectionId);
          currentCollectionName.value = collection ? collection.name : '';
        }
        
        // 更新分页和列表
        if (data.points && data.points.length > 0) {
          pointsPage.value = data.page!;
          collectionPoints.value = data.points;
          pointsTotal.value = data.total!;
          
          // 跳转到目标知识点
          await selectPoint(targetPointId);
        }
      } else {
        ElMessage.error(res.data.msg || '查找知识点失败');
      }
    } catch (error: any) {
      console.error('查找知识点失败:', error);
      ElMessage.error(error.response?.data?.msg || '查找失败');
    }
  };
  
  // 返回上一个知识点（使用导航栈）
  const goBackToPreviousPoint = async () => {
    if (navigationStack.value.length === 0) {
      ElMessage.info('已经是最初位置');
      return;
    }
    
    const prevState = navigationStack.value.pop()!;
    
    // 判断是否需要切换集合
    if (prevState.collectionId !== currentCollectionId.value) {
      currentCollectionId.value = prevState.collectionId;
      const collection = collections.value.find(c => c.id === prevState.collectionId);
      currentCollectionName.value = collection ? collection.name : '';
    }
    
    // 判断是否需要切换分页
    if (prevState.page !== pointsPage.value) {
      pointsPage.value = prevState.page;
      await fetchCollectionPoints();
    }
    
    // 跳转到之前的知识点
    await selectPoint(prevState.pointId);
    
    // 恢复滚动位置
    setTimeout(() => {
      const scrollContainer = document.querySelector('.html-preview') || document.querySelector('.content-box');
      if (scrollContainer && prevState.scrollTop > 0) {
        scrollContainer.scrollTop = prevState.scrollTop;
      }
    }, 100);
  };

  const handleBack = () => {
    router.push('/');
  };

  // 切换单词本显示
  const goToWordbook = () => {
    wordbookVisible.value = !wordbookVisible.value;
  };

  const handleCreateCollection = () => {
    createForm.value.name = '';
    showCreateDialog.value = true;
  };

  const confirmCreate = async () => {
    if (!createForm.value.name.trim()) {
      ElMessage.warning('请输入集合名称');
      return;
    }

    // 检查名称是否重复（比较时提取名称部分）
    const isDuplicate = collections.value.some(c => {
      const match = c.name.match(/^\d+\.\s*(.*)$/);
      const existingName = match ? match[1] : c.name;
      return existingName === createForm.value.name.trim();
    });
    if (isDuplicate) {
      ElMessage.warning('集合名称已存在，请使用其他名称');
      return;
    }

    createLoading.value = true;
    try {
      const res = await createCollection({ name: createForm.value.name.trim() });

      if (res.data.code === 200) {
        ElMessage.success('创建成功');
        // 重新获取列表
        await fetchCollections();
        // 切换到新创建的集合
        currentCollectionId.value = res.data.data.id;
        showCreateDialog.value = false;
      } else {
        ElMessage.error(res.data.msg || '创建失败');
      }
    } catch (error: any) {
      console.error('创建集合失败:', error);
      ElMessage.error(error.response?.data?.msg || '创建失败，请稍后重试');
    } finally {
      createLoading.value = false;
    }
  };

  // 难度工具方法
  const getDifficultyText = (difficulty: number): string => {
    const difficultyMap: Record<number, string> = {
      0: '简单',
      1: '普通',
      2: '困难',
      3: '地狱'
    };
    return difficultyMap[difficulty] || '普通';
  };

  const getDifficultyClass = (difficulty: number): string => {
    const classMap: Record<number, string> = {
      0: 'easy',
      1: 'normal',
      2: 'hard',
      3: 'hell'
    };
    return classMap[difficulty] || 'normal';
  };

  // 处理集合菜单操作
  const handleCollectionMenu = (command: string, collection: Collection) => {
    if (command === 'edit') {
      // 提取序号后的名称部分（如 "1. 我的集合" 提取 "我的集合"）
      let nameWithoutPrefix = collection.name;
      const match = collection.name.match(/^\d+\.\s*(.*)$/);
      if (match) {
        nameWithoutPrefix = match[1]; // 提取序号后的名称
      }
      
      editForm.value = {
        id: collection.id,
        name: nameWithoutPrefix // 只显示名称部分
      };
      showEditDialog.value = true;
    } else if (command === 'manage') {
      handleManageCollection(collection);
    } else if (command === 'delete') {
      handleDeleteCollection(collection);
    }
  };

  // 集合管理相关
  const showManageDialog = ref(false);
  const manageTabActive = ref('permission');
  const manageLoading = ref(false);
  const currentManageCollection = ref<Collection | null>(null);

  // 权限设置表单
  const permissionForm = ref({
    isPublic: false
  });

  // 授权表单
  const authForm = ref({
    userCode: '',
    expireTime: ''
  });

  // 授权列表
  const authorizations = ref<CollectionPermission[]>([]);
  const authPage = ref(1);
  const authPageSize = ref(10);
  const authTotal = ref(0);
  const authSearchKeyword = ref('');

  // 修改授权表单
  const showEditAuthDialog = ref(false);
  const editAuthForm = ref({
    userCode: '',
    expireTime: ''
  });

  // 管理集合
  const handleManageCollection = async (collection: Collection) => {
    currentManageCollection.value = collection;
    permissionForm.value.isPublic = collection.isPublic || false;
    showManageDialog.value = true;
    
    // 如果是私有集合且是所有者，加载授权列表
    if (!collection.isPublic && collection.isOwner) {
      await loadAuthorizations();
    }
  };

  // 关闭管理对话框时的清理函数
  const handleManageDialogClosed = () => {
    // 重置分页
    authPage.value = 1;
    authPageSize.value = 10;
    authTotal.value = 0;
  };

  // 确认编辑
  const confirmEdit = async () => {
    if (!editForm.value.name.trim()) {
      ElMessage.warning('请输入集合名称');
      return;
    }

    // 检查名称是否与其他集合重复（排除自己，比较时提取名称部分）
    const isDuplicate = collections.value.some(c => {
      if (c.id === editForm.value.id) return false;
      const match = c.name.match(/^\d+\.\s*(.*)$/);
      const existingName = match ? match[1] : c.name;
      return existingName === editForm.value.name.trim();
    });
    if (isDuplicate) {
      ElMessage.warning('集合名称已存在，请使用其他名称');
      return;
    }

    editLoading.value = true;
    try {
      const res = await updateCollection(editForm.value.id, { name: editForm.value.name.trim() });

      if (res.data.code === 200) {
        ElMessage.success('修改成功');
        // 直接更新列表中的集合名称（使用后端返回的带序号的名称）
        const index = collections.value.findIndex(c => c.id === editForm.value.id);
        if (index !== -1 && res.data.data?.name) {
          collections.value[index].name = res.data.data.name;
        }
        showEditDialog.value = false;
      } else {
        ElMessage.error(res.data.msg || '修改失败');
      }
    } catch (error: any) {
      console.error('修改集合失败:', error);
      ElMessage.error(error.response?.data?.msg || '修改失败，请稍后重试');
    } finally {
      editLoading.value = false;
    }
  };

  // 删除集合
  const handleDeleteCollection = (collection: Collection) => {
    ElMessageBox.confirm(
      `确定要删除集合"${collection.name}"吗？删除后将无法恢复！`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger'
      }
    ).then(async () => {
      try {
        const res = await deleteCollection(collection.id);
        if (res.data.code === 200) {
          ElMessage.success('删除成功');
          // 重新获取列表
          await fetchCollections();
          // 如果删除的是当前选中的集合，切换到第一个
          if (currentCollectionId.value === collection.id && collections.value.length > 0) {
            currentCollectionId.value = collections.value[0].id;
          }
        } else {
          ElMessage.error(res.data.msg || '删除失败');
        }
      } catch (error: any) {
        console.error('删除集合失败:', error);
        ElMessage.error(error.response?.data?.msg || '删除失败，请稍后重试');
      }
    }).catch(() => {
      // 用户取消删除
    });
  };

  // 获取当前选中集合的所有权信息
  const getCurrentCollectionOwnership = () => {
    const collection = collections.value.find(c => c.id === currentCollectionId.value);
    return collection ? { isOwner: collection.isOwner || false } : { isOwner: false };
  };

  // 置顶
  const movePointToTop = async (index: number) => {
    // 检查是否是所有者
    const { isOwner } = getCurrentCollectionOwnership();
    if (!isOwner) {
      ElMessage.warning('您没有权限修改此集合');
      return;
    }
    
    if (index === 0) return;
    
    const item = collectionPoints.value.splice(index, 1)[0];
    collectionPoints.value.unshift(item);
    
    // 更新所有项的 sort_order
    await updateSortOrderToBackend();
    ElMessage.success('已置顶');
  };

  // 上移
  const movePointUp = async (index: number) => {
    // 检查是否是所有者
    const { isOwner } = getCurrentCollectionOwnership();
    if (!isOwner) {
      ElMessage.warning('您没有权限修改此集合');
      return;
    }
    
    if (index === 0) return;
    
    const temp = collectionPoints.value[index];
    collectionPoints.value[index] = collectionPoints.value[index - 1];
    collectionPoints.value[index - 1] = temp;
    
    // 更新排序
    await updateSortOrderToBackend();
    ElMessage.success('已上移');
  };

  // 下移
  const movePointDown = async (index: number) => {
    // 检查是否是所有者
    const { isOwner } = getCurrentCollectionOwnership();
    if (!isOwner) {
      ElMessage.warning('您没有权限修改此集合');
      return;
    }
    
    if (index === collectionPoints.value.length - 1) return;
    
    const temp = collectionPoints.value[index];
    collectionPoints.value[index] = collectionPoints.value[index + 1];
    collectionPoints.value[index + 1] = temp;
    
    // 更新排序
    await updateSortOrderToBackend();
    ElMessage.success('已下移');
  };

  // 移除知识点
  const removePoint = (point: CollectionPoint) => {
    // 检查是否是所有者
    const { isOwner } = getCurrentCollectionOwnership();
    if (!isOwner) {
      ElMessage.warning('您没有权限修改此集合');
      return;
    }
    
    ElMessageBox.confirm(
      `确定要从集合中移除知识点"${point.title}"吗？`,
      '移除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    ).then(async () => {
      try {
        const res = await removePointFromCollection(point.id);
        if (res.data.code === 200) {
          ElMessage.success('移除成功');
          // 重新加载列表
          await fetchCollectionPoints();
        } else {
          ElMessage.error(res.data.msg || '移除失败');
        }
      } catch (error: any) {
        console.error('移除知识点失败:', error);
        ElMessage.error(error.response?.data?.msg || '移除失败，请稍后重试');
      }
    }).catch(() => {
      // 用户取消
    });
  };

  // 更新排序到后端
  const updateSortOrderToBackend = async () => {
    // 检查是否是所有者
    const { isOwner } = getCurrentCollectionOwnership();
    if (!isOwner) {
      ElMessage.warning('您没有权限修改此集合');
      return;
    }
    
    if (!currentCollectionId.value || collectionPoints.value.length === 0) {
      return;
    }

    try {
      // 按当前顺序生成 sort_order（递减，保证第一个最大）
      const items = collectionPoints.value.map((point, index) => ({
        id: point.id,
        sort_order: collectionPoints.value.length - index
      }));

      const res = await updateCollectionItemsOrder(currentCollectionId.value, items);
      
      if (res.data.code !== 200) {
        ElMessage.error(res.data.msg || '更新排序失败');
      }
    } catch (error: any) {
      console.error('更新排序失败:', error);
      ElMessage.error(error.response?.data?.msg || '更新排序失败');
    }
  };

  // 搜索处理
  const handleAuthSearch = () => {
    // 重置到第一页
    authPage.value = 1;
    loadAuthorizations();
  };

  // 加载授权列表
  const loadAuthorizations = async () => {
    if (!currentManageCollection.value || !currentManageCollection.value.isOwner) return;
    
    manageLoading.value = true;
    try {
      const res = await getCollectionPermissions(
        currentManageCollection.value.id, 
        authPage.value, 
        authPageSize.value,
        authSearchKeyword.value.trim()
      );
      if (res.data.code === 200) {
        authorizations.value = res.data.data.list || [];
        authTotal.value = res.data.data.total || 0;
      } else {
        ElMessage.error(res.data.msg || '获取授权列表失败');
      }
    } catch (error: any) {
      console.error('获取授权列表失败:', error);
      ElMessage.error(error.response?.data?.msg || '获取授权列表失败，请稍后重试');
    } finally {
      manageLoading.value = false;
    }
  };

  // 保存权限设置
  const savePermission = async () => {
    if (!currentManageCollection.value || !currentManageCollection.value.isOwner) return;
    
    manageLoading.value = true;
    try {
      const res = await setCollectionPermission(currentManageCollection.value.id, permissionForm.value.isPublic);
      if (res.data.code === 200) {
        ElMessage.success('权限设置保存成功');
        
        // 更新本地集合列表中的权限状态
        const index = collections.value.findIndex(c => c.id === currentManageCollection.value!.id);
        if (index !== -1) {
          collections.value[index].isPublic = permissionForm.value.isPublic;
        }
        
        // 如果改为公有，清空授权列表
        if (permissionForm.value.isPublic) {
          authorizations.value = [];
        }
      } else {
        ElMessage.error(res.data.msg || '保存权限设置失败');
      }
    } catch (error: any) {
      console.error('保存权限设置失败:', error);
      ElMessage.error(error.response?.data?.msg || '保存权限设置失败，请稍后重试');
    } finally {
      manageLoading.value = false;
    }
  };

  // 添加授权
  const addAuthorization = async () => {
    if (!currentManageCollection.value || !currentManageCollection.value.isOwner) return;
    if (!authForm.value.userCode.trim()) {
      ElMessage.warning('请输入用户Code');
      return;
    }
    
    manageLoading.value = true;
    try {
      const res = await addCollectionPermission(
        currentManageCollection.value.id, 
        authForm.value.userCode, 
        authForm.value.expireTime || undefined
      );
      if (res.data.code === 200) {
        ElMessage.success('授权添加成功');
        authForm.value.userCode = '';
        authForm.value.expireTime = '';
        await loadAuthorizations(); // 重新加载授权列表
      } else {
        ElMessage.error(res.data.msg || '添加授权失败');
      }
    } catch (error: any) {
      console.error('添加授权失败:', error);
      ElMessage.error(error.response?.data?.msg || '添加授权失败，请稍后重试');
    } finally {
      manageLoading.value = false;
    }
  };

  // 编辑授权
  const editAuthorization = (auth: CollectionPermission) => {
    editAuthForm.value.userCode = auth.userCode;
    editAuthForm.value.expireTime = auth.expireTime;
    showEditAuthDialog.value = true;
  };

  // 保存编辑授权
  const saveEditAuthorization = async () => {
    if (!currentManageCollection.value || !currentManageCollection.value.isOwner) return;
    
    manageLoading.value = true;
    try {
      const res = await updateCollectionPermission(
        currentManageCollection.value.id, 
        editAuthForm.value.userCode, 
        editAuthForm.value.expireTime || undefined
      );
      if (res.data.code === 200) {
        ElMessage.success('授权时间更新成功');
        showEditAuthDialog.value = false;
        await loadAuthorizations(); // 重新加载授权列表
      } else {
        ElMessage.error(res.data.msg || '更新授权时间失败');
      }
    } catch (error: any) {
      console.error('更新授权时间失败:', error);
      ElMessage.error(error.response?.data?.msg || '更新授权时间失败，请稍后重试');
    } finally {
      manageLoading.value = false;
    }
  };

  // 删除授权
  const deleteAuthorization = (auth: CollectionPermission) => {
    if (!currentManageCollection.value || !currentManageCollection.value.isOwner) return;
    
    ElMessageBox.confirm(
      `确定要删除用户 ${auth.userCode} 的授权吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    ).then(async () => {
      manageLoading.value = true;
      try {
        const res = await deleteCollectionPermission(currentManageCollection.value!.id, auth.userCode);
        if (res.data.code === 200) {
          ElMessage.success('授权删除成功');
          await loadAuthorizations(); // 重新加载授权列表
        } else {
          ElMessage.error(res.data.msg || '删除授权失败');
        }
      } catch (error: any) {
        console.error('删除授权失败:', error);
        ElMessage.error(error.response?.data?.msg || '删除授权失败，请稍后重试');
      } finally {
        manageLoading.value = false;
      }
    }).catch(() => {
      // 用户取消删除
    });
  };

  // 根据难度获取知识点数量
  const getPointsByDifficulty = (difficulty: number) => {
    return collectionPoints.value.filter(point => point.pointDifficulty === difficulty).length;
  };

  // 监听 currentCollectionId 变化，自动加载知识点
  watch(currentCollectionId, (newId) => {
    if (newId) {
      pointsPage.value = 1;
      fetchCollectionPoints();
    }
  });

  return {
    // 状态
    collections,
    currentCollectionId,
    currentCollectionName,
    loading,
    wordbookVisible,
    isCollectionOwner,
    collectionPoints,
    pointsLoading,
    pointsPage,
    pointsPageSize,
    pointsTotal,
    isEditMode,
    isRandomMode,
    selectedPointDetail,
    selectedPointId,
    currentPointBindings,
    practiceDrawerVisible,
    pointPracticeDrawerVisible,
    showCreateDialog,
    createLoading,
    createForm,
    showEditDialog,
    editLoading,
    editForm,
    showManageDialog,
    manageTabActive,
    manageLoading,
    currentManageCollection,
    permissionForm,
    authForm,
    authorizations,
    authPage,
    authPageSize,
    authTotal,
    authSearchKeyword,
    showEditAuthDialog,
    editAuthForm,
    categorySortOrder,
    pointSortOrder,
    
    // 方法
    fetchCollections,
    selectCollection,
    fetchCollectionPoints,
    selectPoint,
    handleBack,
    goToWordbook,
    handleCreateCollection,
    confirmCreate,
    handleStartPractice,
    getDifficultyText,
    getDifficultyClass,
    handleCollectionMenu,
    handleManageCollection,
    handleManageDialogClosed,
    confirmEdit,
    handleDeleteCollection,
    movePointToTop,
    movePointUp,
    movePointDown,
    removePoint,
    handleAuthSearch,
    loadAuthorizations,
    savePermission,
    addAuthorization,
    editAuthorization,
    saveEditAuthorization,
    deleteAuthorization,
    getPointsByDifficulty,
    toggleMode,
    refreshRandomOrder,
    reloadPoints,
    toggleCategorySort,
    togglePointSort,
    getSubjectColor,
    getCategoryColor,
    navigateToPointFromBinding,
    goBackToPreviousPoint,
    navigationStack
  };
}
