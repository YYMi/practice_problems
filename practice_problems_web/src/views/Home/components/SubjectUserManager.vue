<template>
  <el-dialog v-model="visible" :title="`管理授权用户 - ${subjectName}`" width="950px" destroy-on-close>
    
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <div class="left-tools">
        <el-input
          v-model="searchCode"
          placeholder="输入用户 Code 搜索"
          style="width: 240px"
          clearable
          @clear="fetchData"
          @keyup.enter="fetchData"
        >
          <template #append>
            <el-button icon="Search" @click="fetchData" />
          </template>
        </el-input>
      </div>

      <!-- 批量操作区 -->
      <div class="batch-actions" v-if="selectedIds.length > 0">
        <span class="selected-tip">已选 {{ selectedIds.length }} 人</span>
        
        <el-button type="primary" plain icon="Timer" @click="openBatchEditTime">批量修改时间</el-button>
        
        <el-popconfirm :title="`确定移除选中的 ${selectedIds.length} 位用户吗？`" @confirm="handleBatchRemove">
          <template #reference>
            <el-button type="danger" plain icon="Delete">批量移除</el-button>
          </template>
        </el-popconfirm>
      </div>
    </div>

    <el-table 
      :data="list" 
      style="width: 100%" 
      v-loading="loading" 
      border
      @selection-change="handleSelectionChange"
    >
      <!-- 多选框 -->
      <el-table-column type="selection" width="50" align="center" />

      <!-- 用户信息列 -->
      <el-table-column label="用户" min-width="240">
        <template #default="{ row }">
          <div class="user-info-cell">
            <el-avatar :size="36" style="background:#409eff; flex-shrink:0;">{{ row.nickname?.charAt(0).toUpperCase() || 'U' }}</el-avatar>
            <div class="texts">
              <div class="u-name">
                {{ row.nickname }} 
                <el-tag size="small" type="info" class="copy-tag" @click.stop="copyText(row.user_code)">
                  @{{ row.user_code }} <el-icon><CopyDocument /></el-icon>
                </el-tag>
              </div>
              <div class="u-email">
                <span v-if="row.email">{{ row.email }}</span>
                <span v-else style="color:#ccc">无邮箱</span>
                <!-- 修复：邮箱复制按钮 -->
                <el-icon v-if="row.email" class="copy-btn-mini" title="复制邮箱" @click.stop="copyText(row.email)">
                  <CopyDocument />
                </el-icon>
              </div>
            </div>
          </div>
        </template>
      </el-table-column>

      <el-table-column label="绑定时间" width="170" prop="bind_time" align="center" />

      <!-- 截止日期 (点击编辑图标打开弹窗) -->
      <el-table-column label="截止日期" width="200">
        <template #default="{ row }">
          <div class="display-area">
            <span :class="{'forever-tag': row.expire_time === '永久', 'expired-tag': isExpired(row.expire_time)}">
              {{ row.expire_time }}
            </span>
            <el-icon class="edit-icon" @click="openSingleEditDialog(row)"><EditPen /></el-icon>
          </div>
        </template>
      </el-table-column>

      <el-table-column label="操作" width="100" align="center" fixed="right">
        <template #default="{ row }">
          <el-popconfirm title="确定移除该用户的权限吗？" @confirm="handleSingleRemove(row.id)">
            <template #reference>
              <el-button type="danger" link size="small">移除权限</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-area">
      <el-pagination 
        background 
        layout="total, prev, pager, next" 
        :total="total" 
        :page-size="pageSize"
        @current-change="handlePageChange"
      />
    </div>

    <!-- ★★★ 单个编辑弹窗 ★★★ -->
    <el-dialog v-model="singleEditDialogVisible" title="修改截止时间" width="450px" append-to-body>
      <el-form label-position="top">
        <el-form-item :label="`用户: ${currentEditUser.nickname} (@${currentEditUser.user_code})`">
          <div style="display: flex; gap: 10px; align-items: center;">
            <el-date-picker
              v-model="singleNewDate"
              type="datetime"
              placeholder="选择截止时间"
              value-format="YYYY-MM-DD HH:mm:ss"
              style="flex: 1;"
            />
            <el-button type="warning" plain @click="singleNewDate = ''">永久</el-button>
          </div>
          <div class="form-hint" style="margin-top: 8px; font-size: 12px; color: #909399;">
            <el-icon><InfoFilled /></el-icon>
            点击“永久”按钮即可设置为永久有效
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="singleEditDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmSingleUpdate">确定修改</el-button>
      </template>
    </el-dialog>

    <!-- ★★★ 批量修改时间弹窗 ★★★ -->
    <el-dialog v-model="batchTimeDialogVisible" title="批量修改截止时间" width="450px" append-to-body>
      <el-form label-position="top">
        <el-form-item label="统一设置为">
          <div style="display: flex; gap: 10px; align-items: center;">
            <el-date-picker
              v-model="batchNewDate"
              type="datetime"
              placeholder="选择截止时间"
              value-format="YYYY-MM-DD HH:mm:ss"
              style="flex: 1;"
            />
            <!-- ★★★ 添加永久按钮 ★★★ -->
            <el-button type="warning" plain @click="batchNewDate = ''">永久</el-button>
          </div>
          <div class="form-hint" style="margin-top: 8px; font-size: 12px; color: #909399;">
            <el-icon><InfoFilled /></el-icon>
            点击“永久”按钮即可设置为永久有效
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchTimeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmBatchUpdate">确定修改</el-button>
      </template>
    </el-dialog>

  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { getSubjectUsers, updateAuth, removeAuth, batchUpdateAuth, batchRemoveAuth } from '../../../api/share';
import { ElMessage } from 'element-plus';
import { CopyDocument, EditPen, Search, Delete, Timer, InfoFilled } from '@element-plus/icons-vue';

const props = defineProps(['visible', 'subjectId', 'subjectName']);
const emit = defineEmits(['update:visible']);

const visible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
});

const list = ref<any[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const loading = ref(false);
const searchCode = ref(''); 
const selectedIds = ref<number[]>([]); 

// 单个编辑弹窗状态
const singleEditDialogVisible = ref(false);
const singleNewDate = ref('');
const currentEditUser = ref<any>({ id: 0, nickname: '', user_code: '' });

// 批量修改弹窗状态
const batchTimeDialogVisible = ref(false);
const batchNewDate = ref('');

const fetchData = async () => {
  if(!props.subjectId) return;
  loading.value = true;
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value,
      user_code: searchCode.value // 传递搜索参数
    };
    const res = await getSubjectUsers(props.subjectId, params);
    if (res.data && res.data.code === 200) {
      list.value = res.data.data.list;
      total.value = res.data.data.total;
    }
  } finally {
    loading.value = false;
  }
};

watch(() => props.visible, (val) => {
  if (val) {
    page.value = 1;
    searchCode.value = '';
    selectedIds.value = [];
    fetchData();
  }
});

const handlePageChange = (p: number) => {
  page.value = p;
  fetchData();
};

const handleSelectionChange = (selection: any[]) => {
  selectedIds.value = selection.map(item => item.id);
};

// --- 批量操作 ---

// 1. 批量移除
const handleBatchRemove = async () => {
  if (selectedIds.value.length === 0) return;
  try {
    loading.value = true;
    // 调用后端批量接口
    const res = await batchRemoveAuth(selectedIds.value);
    if (res.data && res.data.code === 200) {
      ElMessage.success(res.data.msg);
      selectedIds.value = []; // 清空选择
      fetchData();
    } else {
      ElMessage.error('操作失败');
    }
  } finally {
    loading.value = false;
  }
};

// 2. 批量修改时间
const openBatchEditTime = () => {
  batchNewDate.value = ''; // 默认空（永久）
  batchTimeDialogVisible.value = true;
};

const confirmBatchUpdate = async () => {
  if (selectedIds.value.length === 0) return;
  try {
    loading.value = true;
    const finalDate = batchNewDate.value || 'forever';
    const res = await batchUpdateAuth(selectedIds.value, finalDate);
    
    if (res.data && res.data.code === 200) {
      ElMessage.success(res.data.msg);
      batchTimeDialogVisible.value = false;
      selectedIds.value = [];
      fetchData();
    } else {
      ElMessage.error('操作失败');
    }
  } finally {
    loading.value = false;
  }
};

// --- 单行操作 ---

// 打开单个编辑弹窗
const openSingleEditDialog = (row: any) => {
  currentEditUser.value = { 
    id: row.id, 
    nickname: row.nickname, 
    user_code: row.user_code 
  };
  singleNewDate.value = row.raw_expire || ''; 
  singleEditDialogVisible.value = true;
};

// 确认单个修改
const confirmSingleUpdate = async () => {
  const finalDate = singleNewDate.value || 'forever'; 
  try {
    loading.value = true;
    const res = await updateAuth(currentEditUser.value.id, finalDate);
    if (res.data && res.data.code === 200) {
      ElMessage.success('修改成功');
      singleEditDialogVisible.value = false;
      fetchData(); 
    } else {
      ElMessage.error('修改失败');
    }
  } finally {
    loading.value = false;
  }
};

const handleSingleRemove = async (id: number) => {
  const res = await removeAuth(id);
  if (res.data && res.data.code === 200) {
    ElMessage.success('已移除');
    fetchData();
  }
};

// --- 工具函数 ---

const copyText = (text: string) => {
  if (!text) return;
  navigator.clipboard.writeText(text);
  ElMessage.success('已复制: ' + text);
};

const isExpired = (expireStr: string) => {
  if (expireStr === '永久') return false;
  const expireDate = new Date(expireStr);
  return expireDate < new Date();
};
</script>

<style scoped>
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px; }
.batch-actions { display: flex; align-items: center; gap: 10px; }
.selected-tip { font-size: 13px; color: #909399; margin-right: 5px; }

.user-info-cell { display: flex; align-items: center; gap: 12px; }
.texts { display: flex; flex-direction: column; line-height: 1.5; }
.u-name { font-size: 14px; font-weight: 500; color: #303133; display: flex; align-items: center; }
.copy-tag { margin-left: 5px; cursor: pointer; display: inline-flex; align-items: center; gap: 3px; }
.copy-tag:hover { color: #409eff; border-color: #c6e2ff; background-color: #ecf5ff; }

.u-email { font-size: 12px; color: #606266; display: flex; align-items: center; gap: 6px; }
.copy-btn-mini { cursor: pointer; color: #909399; font-size: 13px; transition: color 0.2s; }
.copy-btn-mini:hover { color: #409eff; }

.edit-area { display: flex; align-items: center; gap: 5px; }
.display-area { display: flex; align-items: center; justify-content: space-between; width: 100%; }
.edit-icon { cursor: pointer; color: #409eff; opacity: 0.6; transition: opacity 0.2s; }
.display-area:hover .edit-icon { opacity: 1; }

.forever-tag { color: #67c23a; font-weight: bold; }
.expired-tag { color: #f56c6c; text-decoration: line-through; }
.pagination-area { margin-top: 15px; display: flex; justify-content: flex-end; }
</style>