<template>
  <el-dialog v-model="visible" title="我的分享码管理" width="900px" destroy-on-close>
    <el-table :data="list" style="width: 100%" v-loading="loading">
      
      <el-table-column label="分享码" width="140">
        <template #default="{ row }">
          <span class="code-text">{{ row.code }}</span>
          <el-icon class="copy-icon" @click="copyCode(row.code)"><CopyDocument /></el-icon>
        </template>
      </el-table-column>

      <el-table-column label="已领取" width="80" align="center">
        <template #default="{ row }">
          <span style="color: #409eff; font-weight: bold;">{{ row.used_count }}</span> 人
        </template>
      </el-table-column>

      <el-table-column label="资源有效期" width="120">
        <template #default="{ row }">
          <el-tag type="info">{{ formatDuration(row.resource_time) }}</el-tag>
        </template>
      </el-table-column>

      <el-table-column label="码截止时间" min-width="180">
        <template #default="{ row }">
          <div :class="{ 'expired-text': row.status === 'expired' }">
            {{ formatTime(row.expire_time) }}
          </div>
        </template>
      </el-table-column>

      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag v-if="row.status === 'active'" type="success" size="small">有效</el-tag>
          <el-tag v-else type="danger" size="small">已过期</el-tag>
        </template>
      </el-table-column>

      <el-table-column label="操作" width="140" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
          <el-popconfirm title="确定删除？" @confirm="handleDelete(row.id)">
            <template #reference>
              <el-button link type="danger" size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <!-- ★★★ 编辑弹窗 ★★★ -->
    <el-dialog v-model="editDialogVisible" title="修改分享码设置" width="450px" append-to-body>
      <el-form label-position="top">
        
        <!-- 1. 修改截止时间 (基于创建时间限制) -->
        <el-form-item label="分享码截止时间 (何时失效)">
          <el-date-picker
            v-model="editForm.newExpireDate"
            type="datetime"
            placeholder="选择日期时间"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%"
            :disabled-date="disabledDate"
          />
          <div class="form-tip">
            限制：该码创建于 {{ formatTime(currentCreateTime) }}，最晚只能延期至创建后 1 年。
          </div>
        </el-form-item>

        <!-- 2. 修改资源有效期 (全套快捷键) -->
        <el-form-item label="资源有效期 (用户领取后能看多久)">
          <div style="display: flex; gap: 10px; margin-bottom: 10px;">
            <el-input 
              v-model="resCustomVal" 
              type="number" 
              min="1" 
              style="flex: 1" 
              placeholder="自定义时长"
              @input="handleInputChanged"
            >
              <template #append>
                <el-select v-model="resCustomUnit" style="width: 70px" @change="handleInputChanged">
                  <el-option label="天" value="d" /><el-option label="周" value="w" />
                  <el-option label="月" value="m" /><el-option label="年" value="y" />
                </el-select>
              </template>
            </el-input>
          </div>
          
          <!-- 快捷按钮：日 周 月 年 永久 -->
          <el-radio-group v-model="presetDuration" size="small" @change="handlePresetChange">
            <el-radio-button label="1d">1天</el-radio-button>
            <el-radio-button label="7d">1周</el-radio-button>
            <el-radio-button label="30d">1月</el-radio-button>
            <el-radio-button label="365d">1年</el-radio-button>
            <el-radio-button label="forever">永久</el-radio-button>
          </el-radio-group>
        </el-form-item>

      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitEdit">保存修改</el-button>
      </template>
    </el-dialog>

  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
import { getMyShareCodes, deleteShareCode, updateShareCode } from '../../../api/share';
import { ElMessage } from 'element-plus';
import { CopyDocument } from '@element-plus/icons-vue';

const props = defineProps(['visible']);
const emit = defineEmits(['update:visible']);

const visible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
});

const list = ref([]);
const loading = ref(false);

// 编辑相关状态
const editDialogVisible = ref(false);
const currentEditId = ref(0);
const currentCreateTime = ref(''); // ★★★ 新增：记录当前码的创建时间

const editForm = reactive({
  newExpireDate: '',
  newDuration: ''
});

const resCustomVal = ref('');
const resCustomUnit = ref('d');
const presetDuration = ref(''); 

const fetchList = async () => {
  loading.value = true;
  try {
    const res = await getMyShareCodes();
    if (res.data && res.data.code === 200) {
      list.value = res.data.data;
    }
  } finally {
    loading.value = false;
  }
};

watch(() => props.visible, (val) => {
  if (val) fetchList();
});

// ============ 互斥逻辑 ============
const handleInputChanged = () => {
  if (resCustomVal.value) {
    presetDuration.value = ''; 
    editForm.newDuration = `${resCustomVal.value}${resCustomUnit.value}`;
  }
};

const handlePresetChange = (val: string) => {
  editForm.newDuration = val; 
  if (val === 'forever') {
    resCustomVal.value = ''; 
  } else {
    const match = val.match(/^(\d+)([a-z]+)$/);
    if (match) {
      resCustomVal.value = match[1];
      resCustomUnit.value = match[2];
    }
  }
};

// ============ 打开编辑 ============
const handleEdit = (row: any) => {
  currentEditId.value = row.id;
  currentCreateTime.value = row.create_time; // ★★★ 保存创建时间
  
  editForm.newExpireDate = formatTime(row.expire_time);
  editForm.newDuration = row.resource_time;
  
  // 回填资源有效期UI
  if (row.resource_time === 'forever') {
    presetDuration.value = 'forever';
    resCustomVal.value = '';
  } else {
    if (['1d', '7d', '30d', '365d'].includes(row.resource_time)) {
      presetDuration.value = row.resource_time;
    } else {
      presetDuration.value = '';
    }
    const match = row.resource_time.match(/^(\d+)([a-z]+)$/);
    if (match) {
      resCustomVal.value = match[1];
      resCustomUnit.value = match[2];
    }
  }
  
  editDialogVisible.value = true;
};

// ============ 日期限制逻辑 (基于创建时间) ============
const disabledDate = (time: Date) => {
  // 1. 最小时间：不能早于今天 (已经过去的日期没意义)
  const now = new Date();
  now.setHours(0, 0, 0, 0);
  if (time.getTime() < now.getTime()) return true;

  // 2. 最大时间：不能晚于 创建时间 + 1年
  if (currentCreateTime.value) {
    // 需要处理一下格式，保证 Date 能解析 (Safari兼容性)
    const createDateStr = currentCreateTime.value.replace(/-/g, '/').replace('T', ' ').split('.')[0];
    const createDate = new Date(createDateStr);
    
    if (!isNaN(createDate.getTime())) {
      const limitDate = new Date(createDate);
      limitDate.setFullYear(limitDate.getFullYear() + 1); // 往后推1年
      
      return time.getTime() > limitDate.getTime();
    }
  }
  return false;
};

const submitEdit = async () => {
  if (!editForm.newExpireDate) return;
  try {
    const res = await updateShareCode(currentEditId.value, editForm.newExpireDate, editForm.newDuration);
    if (res.data && res.data.code === 200) {
      ElMessage.success('更新成功');
      editDialogVisible.value = false;
      fetchList();
    } else {
      ElMessage.error(res.data?.msg || '更新失败');
    }
  } catch (e) { console.error(e); }
};

const handleDelete = async (id: number) => {
  try {
    const res = await deleteShareCode(id);
    if (res.data && res.data.code === 200) {
      ElMessage.success('删除成功');
      fetchList();
    } else {
      ElMessage.error(res.data?.msg || '删除失败');
    }
  } catch (e) { console.error(e); }
};

const copyCode = (code: string) => {
  navigator.clipboard.writeText(code);
  ElMessage.success('已复制');
};

const formatDuration = (str: string) => {
  if (str === 'forever') return '永久';
  return str.replace('d', '天').replace('w', '周').replace('m', '月').replace('y', '年');
};

const formatTime = (str: string) => {
  if (!str) return '';
  return str.replace('T', ' ').replace('Z', '').split('+')[0];
};
</script>

<style scoped>
.code-text { font-family: monospace; font-weight: bold; color: #303133; margin-right: 5px; }
.copy-icon { cursor: pointer; color: #909399; vertical-align: middle; }
.copy-icon:hover { color: #409eff; }
.expired-text { color: #909399; text-decoration: line-through; }
.form-tip { font-size: 12px; color: #e6a23c; margin-top: 5px; }
</style>