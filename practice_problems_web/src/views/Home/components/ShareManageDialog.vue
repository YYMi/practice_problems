<template>
  <!-- 主弹窗：我的分享码管理 -->
  <el-dialog 
    v-model="visible" 
    width="1000px" 
    destroy-on-close 
    class="custom-dialog" 
    :show-close="false"
  >
    <!-- ★★★ 1. 统一的紫色头部 ★★★ -->
    <template #header>
      <div class="dialog-header-purple">
        <div class="dh-icon-box"><el-icon><List /></el-icon></div>
        <span class="dh-title-center">我的分享码管理</span>
        <el-icon class="dh-close" @click="emit('update:visible', false)"><Close /></el-icon>
      </div>
    </template>

    <div class="dialog-table-content">
      <!-- ★★★ 2. 表格区域美化 ★★★ -->
      <el-table 
        :data="list" 
        style="width: 100%" 
        v-loading="loading" 
        :header-cell-style="{ background: '#f8f9fb', color: '#606266', height: '50px' }"
      >
        
        <el-table-column label="分享码" width="180">
          <template #default="{ row }">
            <div class="share-code-badge">
              <span class="code-font">{{ row.code }}</span>
              <el-tooltip content="点击复制" placement="top">
                <el-icon class="copy-btn" @click="copyCode(row.code)"><CopyDocument /></el-icon>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="已领取" width="100" align="center">
          <template #default="{ row }">
            <div class="used-count-box">
              <span class="num">{{ row.used_count }}</span>
              <span class="label">人</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="资源有效期" width="120" align="center">
          <template #default="{ row }">
            <el-tag effect="plain" type="info" round>{{ formatDuration(row.resource_time) }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="码截止时间" min-width="180">
          <template #default="{ row }">
            <div class="time-display" :class="{ 'is-expired': row.status === 'expired' }">
              <el-icon><Clock /></el-icon>
              <span>{{ formatTime(row.expire_time) }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="公告发布" width="110" align="center">
          <template #default="{ row }">
            <el-button 
              v-if="row.status === 'active'" 
              class="announce-btn"
              size="small" 
              icon="Bell"
              @click="handleAnnouncement(row)"
            >
              发布
            </el-button>
            <span v-else class="disabled-text">不可用</span>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="90" align="center">
          <template #default="{ row }">
            <div v-if="row.status === 'active'" class="status-dot success">
              <span></span>有效
            </div>
            <div v-else class="status-dot danger">
              <span></span>过期
            </div>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="140" fixed="right" align="center">
          <template #default="{ row }">
            <div class="action-group">
              <!-- ★★★ 分享码过期时禁用编辑 ★★★ -->
              <el-tooltip v-if="row.status === 'expired'" content="分享码已过期，不允许编辑" placement="top">
                <el-button link type="info" size="small" disabled>编辑</el-button>
              </el-tooltip>
              <el-button v-else link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              
              <div class="divider-v"></div>
              
              <el-popconfirm title="确定删除？" @confirm="handleDelete(row.id)">
                <template #reference>
                  <el-button link type="danger" size="small">删除</el-button>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- ★★★ 1. 深度美化：编辑分享码弹窗 ★★★ -->
    <el-dialog 
      v-model="editDialogVisible" 
      width="460px" 
      append-to-body 
      class="custom-dialog"
      :show-close="false"
    >
      <template #header>
        <!-- 新设计的矮头部 -->
        <div class="dialog-header-purple">
          <div class="dh-icon-box"><el-icon><Edit /></el-icon></div>
          <span class="dh-title-center">修改设置</span>
          <el-icon class="dh-close" @click="editDialogVisible = false"><Close /></el-icon>
        </div>
      </template>

      <div class="dialog-body-content">
        <el-form label-position="top" class="stylish-form">
          
          <!-- 资源有效期 -->
          <el-form-item label="资源有效期">
            <!-- 组合输入框：使用 Flex 布局保证高度一致 -->
            <div class="combined-input-wrapper" :class="{ disabled: presetDuration === 'forever' }">
              <el-input 
                v-model="resCustomVal" 
                type="number" 
                min="1" 
                class="flex-input"
                placeholder="输入数值"
                @input="handleInputChanged"
                :disabled="presetDuration === 'forever'"
              />
              <div class="divider-vertical"></div>
              <el-select 
                v-model="resCustomUnit" 
                class="flex-select"
                @change="handleInputChanged"
                :disabled="presetDuration === 'forever'"
                style="width: 80px"
              >
                <el-option label="天" value="d" />
                <el-option label="周" value="w" />
                <el-option label="月" value="m" />
                <el-option label="年" value="y" />
              </el-select>
            </div>
            
            <!-- 快捷选择胶囊 -->
            <div class="preset-pills mt-3">
              <span 
                v-for="opt in [{l:'1天',v:'1d'},{l:'1周',v:'7d'},{l:'1月',v:'30d'},{l:'1年',v:'365d'},{l:'永久',v:'forever'}]"
                :key="opt.v"
                class="pill-item"
                :class="{ active: presetDuration === opt.v }"
                @click="handlePresetChange(opt.v)"
              >
                {{ opt.l }}
              </span>
            </div>
          </el-form-item>

          <div class="divider-line"></div>

          <!-- 截止时间 -->
          <el-form-item label="分享码失效时间">
            <el-date-picker
              v-model="editForm.newExpireDate"
              type="datetime"
              placeholder="请选择失效日期"
              value-format="YYYY-MM-DD HH:mm:ss"
              class="full-width-picker"
              :disabled-date="disabledDate"
              :prefix-icon="Calendar"
            />
            <div class="form-tip">
              <el-icon><Warning /></el-icon> 最晚延期至创建后 1 年内
            </div>
          </el-form-item>
        </el-form>
      </div>

      <template #footer>
        <div class="dialog-footer-simple">
          <el-button class="btn-cancel" @click="editDialogVisible = false">取消</el-button>
          <el-button class="btn-submit" type="primary" @click="submitEdit">保存修改</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- ★★★ 2. 深度美化：发布公告弹窗 ★★★ -->
    <el-dialog 
      v-model="annoDialogVisible" 
      width="460px" 
      append-to-body
      class="custom-dialog"
      :show-close="false"
    >
      <template #header>
        <!-- 新设计的矮头部 -->
        <div class="dialog-header-purple">
          <div class="dh-icon-box"><el-icon><BellFilled /></el-icon></div>
          <span class="dh-title-center">发布公告</span>
          <el-icon class="dh-close" @click="annoDialogVisible = false"><Close /></el-icon>
        </div>
      </template>

      <div class="dialog-body-content">
        <el-form :model="annoForm" label-position="top" class="stylish-form">
          
          <el-form-item label="关联分享码">
            <div class="read-only-box">
              <el-icon><Link /></el-icon>
              <span>{{ annoForm.shareCode }}</span>
            </div>
          </el-form-item>

          <el-form-item label="公告失效时间">
            <el-date-picker
              v-model="annoForm.expireTime"
              type="datetime"
              placeholder="选择失效时间"
              value-format="YYYY-MM-DD HH:mm:ss"
              class="full-width-picker"
              :disabled-date="disabledAnnoDate" 
              :prefix-icon="Calendar"
            />
            <div class="form-tip warning">
              <el-icon><WarningFilled /></el-icon> 不能晚于分享码失效时间
            </div>
          </el-form-item>

          <el-form-item label="公告内容">
            <el-input
              v-model="annoForm.note"
              type="textarea"
              :rows="4"
              placeholder="请输入公告内容，例如：Java全栈面试资料，包含..."
              maxlength="200"
              show-word-limit
              class="custom-textarea"
            />
          </el-form-item>
        </el-form>
      </div>

      <template #footer>
        <div class="dialog-footer-simple">
          <el-button class="btn-cancel" @click="annoDialogVisible = false">取消</el-button>
          <el-button class="btn-submit purple-btn" type="primary" @click="submitAnnouncement" :loading="annoSubmitting">确认发布</el-button>
        </div>
      </template>
    </el-dialog>

  </el-dialog>
</template>


<script setup lang="ts">
// ... Script 逻辑部分完全不需要动，保持你原来的即可 ...
// 记得引入一下图标
import { ref, reactive, computed, watch } from 'vue';
import { getMyShareCodes, deleteShareCode, updateShareCode, createShareAnnouncement } from '../../../api/share';
import { ElMessage } from 'element-plus';
import { CopyDocument, Bell, Edit, BellFilled, Link, Calendar, Warning, WarningFilled, Close, List, Clock } from '@element-plus/icons-vue';


// ... (此处省略 Script 内容，与上一版代码一致) ...
// ... 请直接复制上一版 Script 代码 ...

const props = defineProps(['visible']);
const emit = defineEmits(['update:visible']);

const visible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
});

const list = ref<any[]>([]);
const loading = ref(false);

// --- 编辑相关 ---
const editDialogVisible = ref(false);
const currentEditId = ref(0);
const currentCreateTime = ref('');
const editForm = reactive({ newExpireDate: '', newDuration: '' });

// 编辑弹窗的辅助变量
const resCustomVal = ref('');
const resCustomUnit = ref('d');
const presetDuration = ref(''); 

// --- 公告相关 ---
const annoDialogVisible = ref(false);
const annoSubmitting = ref(false);
const currentShareCodeExpire = ref('');
const annoForm = reactive({ shareCode: '', note: '', expireTime: '' });

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

// ============ 编辑逻辑 (核心修改) ============

// ★★★ 修复 3：回显逻辑优化 + 过期检查 ★★★
const handleEdit = (row: any) => {
  // 检查分享码是否已过期
  if (row.status === 'expired') {
    ElMessage.warning('分享码已过期，不允许编辑');
    return;
  }

  currentEditId.value = row.id;
  currentCreateTime.value = row.create_time;
  
  editForm.newExpireDate = formatTime(row.expire_time);
  
  const val = row.resource_time || '3d';
  
  // 先重置
  resCustomVal.value = '';
  resCustomUnit.value = 'd';
  presetDuration.value = '';

  if (val === 'forever') {
    presetDuration.value = 'forever';
  } else {
    // 解析
    const match = val.match(/^(\d+)([dwmy])$/);
    if (match) {
      resCustomVal.value = match[1];
      resCustomUnit.value = match[2];
      
      // 只有完全匹配快捷值时，才高亮快捷键
      if (['1d', '7d', '30d', '365d'].includes(val)) {
        presetDuration.value = val;
      }
    } else {
      // 默认兜底
      resCustomVal.value = '3';
    }
  }
  
  editDialogVisible.value = true;
};

// ★★★ 修复 1：输入框逻辑 ★★★
// 只要用户在输入框打字，就取消所有快捷键的高亮，除非是永久
const handleInputChanged = () => {
  // 如果当前是永久，不允许输入（因为输入框被 disabled 了，这里其实进不来，但为了保险）
  if (presetDuration.value === 'forever') return;

  // 只要用户手动输入了，就取消下方的快捷胶囊选中状态
  // 这样用户就知道现在是“自定义模式”
  presetDuration.value = ''; 
};

// ★★★ 修复 2：快捷键逻辑 ★★★
const handlePresetChange = (val: string) => {
  // 1. 立即更新高亮状态
  presetDuration.value = val;

  // 2. 处理数值填充
  if (val === 'forever') {
    // 如果是永久，清空输入框，输入框变灰
    resCustomVal.value = '';
    // 单位可以随便给一个，或者保持不变，因为输入框 disabled 了看不到
  } else {
    // 如果是 1年、1月等
    const match = val.match(/^(\d+)([a-z]+)$/);
    if (match) {
      resCustomVal.value = match[1]; // 填入数字，比如 365
      resCustomUnit.value = match[2]; // 填入单位，比如 d
    }
    // 注意：此时输入框是【可以编辑】的
    // 用户点完“1年”后，可以把 365 改成 366，这时候 handleInputChanged 会触发，取消“1年”的高亮
  }
};
// 提交编辑
const submitEdit = async () => {
  if (!editForm.newExpireDate) return ElMessage.warning('请选择截止时间');

  // 计算最终的 duration 字符串
  let finalDuration = '';
  if (presetDuration.value === 'forever') {
    finalDuration = 'forever';
  } else {
    if (!resCustomVal.value) return ElMessage.warning('请输入资源有效时长');
    finalDuration = `${resCustomVal.value}${resCustomUnit.value}`;
  }

  try {
    const res = await updateShareCode(currentEditId.value, editForm.newExpireDate, finalDuration);
    if (res.data && res.data.code === 200) {
      ElMessage.success('更新成功');
      editDialogVisible.value = false;
      fetchList();
    } else {
      ElMessage.error(res.data?.msg || '更新失败');
    }
  } catch (e) { console.error(e); }
};

const disabledDate = (time: Date) => {
  const now = new Date(); now.setHours(0, 0, 0, 0);
  if (time.getTime() < now.getTime()) return true;
  if (currentCreateTime.value) {
    const createDate = new Date(currentCreateTime.value.replace(/-/g, '/').replace('T', ' ').split('.')[0]);
    if (!isNaN(createDate.getTime())) {
      const limitDate = new Date(createDate);
      limitDate.setFullYear(limitDate.getFullYear() + 1);
      return time.getTime() > limitDate.getTime();
    }
  }
  return false;
};

// ============ 公告逻辑 ============

const handleAnnouncement = (row: any) => {
  // ★★★ 检查分享码是否已过期 ★★★
  if (row.status === 'expired') {
    ElMessage.warning('分享码已过期，不允许发布公告');
    return;
  }

  annoForm.shareCode = row.code;
  annoForm.note = ''; 
  currentShareCodeExpire.value = row.expire_time;
  annoForm.expireTime = formatTime(row.expire_time); // 默认一致
  annoDialogVisible.value = true;
};

const disabledAnnoDate = (time: Date) => {
  const now = new Date(); now.setHours(0,0,0,0);
  if (time.getTime() < now.getTime()) return true;
  if (currentShareCodeExpire.value) {
    const limitStr = currentShareCodeExpire.value.replace(/-/g, '/').replace('T', ' ').split('.')[0];
    const limitDate = new Date(limitStr);
    if (!isNaN(limitDate.getTime())) {
      return time.getTime() > limitDate.getTime();
    }
  }
  return false;
};

const submitAnnouncement = async () => {
  if (!annoForm.note) return ElMessage.warning('请输入备注说明');
  const selectedTime = new Date(annoForm.expireTime).getTime();
  const limitTime = new Date(currentShareCodeExpire.value.replace('T', ' ')).getTime();
  if (selectedTime > limitTime) return ElMessage.error('公告截止时间不能晚于分享码失效时间');

  annoSubmitting.value = true;
  try {
    await createShareAnnouncement(annoForm);
    ElMessage.success('公告发布成功');
    annoDialogVisible.value = false;
  } catch (e) {
    ElMessage.error('发布失败');
  } finally {
    annoSubmitting.value = false;
  }
};

// ============ 通用逻辑 ============

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
<style>
/* ============================================================
   1. 弹窗容器基础重置
   ============================================================ */
.custom-dialog {
  border-radius: 12px !important;
  overflow: hidden !important; /* 关键：切掉头部多余的直角 */
  box-shadow: 0 20px 50px rgba(0,0,0,0.2) !important;
  padding: 0 !important; /* 确保外壳没有内边距 */
}

/* 强制去掉 Element 默认的各种 padding */
.custom-dialog .el-dialog__header {
  padding: 0 !important; margin: 0 !important; width: 100% !important; display: block !important;
}
.custom-dialog .el-dialog__headerbtn { display: none !important; }
.custom-dialog .el-dialog__body { padding: 0 !important; }
.custom-dialog .el-dialog__footer { padding: 0 !important; }

/* ============================================================
   2. 统一的紫色头部样式 (精致横条)
   ============================================================ */
.dialog-header-purple {
  width: 100%; 
  height: 54px;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  padding: 0 20px; /* 内部内容的间距 */
  box-sizing: border-box;
  color: #fff;
  position: relative;
}

.dh-icon-box {
  width: 28px; height: 28px;
  background: rgba(255,255,255,0.2);
  border-radius: 6px;
  display: flex; align-items: center; justify-content: center;
  font-size: 16px;
  margin-right: 10px;
  flex-shrink: 0;
}

.dh-title-center {
  font-size: 16px;
  font-weight: 600;
  letter-spacing: 0.5px;
  flex: 1; /* 占据中间空间 */
}

.dh-close {
  font-size: 18px;
  cursor: pointer;
  opacity: 0.8;
  transition: opacity 0.2s;
  padding: 5px;
  border-radius: 4px;
}
.dh-close:hover { opacity: 1; background: rgba(255,255,255,0.15); }

/* ============================================================
   3. 主表格区域美化
   ============================================================ */
.dialog-table-content {
  padding: 20px;
  background: #fff;
  min-height: 400px; /* 给个最小高度，防止空数据时太丑 */
}

/* 分享码胶囊样式 */
.share-code-badge {
  display: inline-flex;
  align-items: center;
  background: #f2f4f7;
  padding: 4px 10px;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
  transition: all 0.2s;
}
.share-code-badge:hover {
  background: #fff;
  border-color: #764ba2;
  box-shadow: 0 2px 8px rgba(118, 75, 162, 0.1);
}
.code-font {
  font-family: 'Monaco', monospace;
  font-weight: 600;
  color: #303133;
  margin-right: 8px;
  font-size: 13px;
}
.copy-btn {
  cursor: pointer;
  color: #909399;
  font-size: 14px;
  transition: transform 0.2s;
}
.copy-btn:hover { color: #764ba2; transform: scale(1.1); }

/* 已领取人数 */
.used-count-box .num { font-size: 16px; font-weight: bold; color: #409eff; margin-right: 2px; }
.used-count-box .label { font-size: 12px; color: #909399; }

/* 时间显示 */
.time-display {
  display: flex; align-items: center; gap: 6px; color: #606266; font-size: 13px;
}
.time-display.is-expired { color: #c0c4cc; text-decoration: line-through; }

/* 发布按钮 */
.announce-btn {
  color: #e6a23c; border-color: #fbe6c8; background: #fdf6ec;
}
.announce-btn:hover {
  color: #fff; background: #e6a23c; border-color: #e6a23c;
}
.disabled-text { font-size: 12px; color: #dcdfe6; }

/* 状态小圆点 */
.status-dot { display: flex; align-items: center; justify-content: center; gap: 6px; font-size: 12px; }
.status-dot span { width: 6px; height: 6px; border-radius: 50%; }
.status-dot.success { color: #67c23a; }
.status-dot.success span { background: #67c23a; box-shadow: 0 0 0 2px rgba(103, 194, 58, 0.2); }
.status-dot.danger { color: #f56c6c; }
.status-dot.danger span { background: #f56c6c; }

/* 操作组 */
.action-group { display: flex; align-items: center; justify-content: center; }
.divider-v { width: 1px; height: 12px; background: #e4e7ed; margin: 0 10px; }

/* ============================================================
   4. 编辑/发布弹窗的内容样式
   ============================================================ */
.dialog-body-content {
  padding: 25px 30px;
  background: #fff;
}

.stylish-form .el-form-item__label {
  font-weight: 600; color: #303133; padding-bottom: 8px; font-size: 14px;
}

/* 组合输入框 */
.combined-input-wrapper {
  display: flex; align-items: center;
  border: 1px solid #dcdfe6; border-radius: 6px;
  overflow: hidden; transition: all 0.3s; background: #fff;
  width: 100%;
}
.combined-input-wrapper:hover { border-color: #c0c4cc; }
.combined-input-wrapper:focus-within { border-color: #764ba2; box-shadow: 0 0 0 3px rgba(118, 75, 162, 0.1); }
.combined-input-wrapper.disabled { background: #f5f7fa; cursor: not-allowed; opacity: 0.8; }

.flex-input .el-input__wrapper, .flex-select .el-input__wrapper {
  box-shadow: none !important; background: transparent !important; padding: 8px 12px !important;
}
.flex-input { flex: 1; }
.divider-vertical { width: 1px; height: 20px; background: #eee; }

/* 快捷胶囊 */
.mt-3 { margin-top: 12px; }
.preset-pills { display: flex; gap: 8px; flex-wrap: wrap; }
.pill-item {
  padding: 5px 14px;
  background: #f5f7fa;
  border-radius: 4px;
  font-size: 12px;
  color: #606266;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s;
}
.pill-item:hover { background: #e6e8eb; color: #333; }
.pill-item.active {
  background: #764ba2;
  color: #fff;
  font-weight: bold;
  box-shadow: 0 2px 6px rgba(118, 75, 162, 0.4);
}

/* 只读框样式 */
.read-only-box {
  width: 100%; padding: 8px 12px; background: #f5f7fa; border-radius: 6px;
  color: #606266; font-family: monospace; display: flex; align-items: center; gap: 8px; font-size: 14px;
  box-sizing: border-box;
}

/* 其他通用 */
.full-width-picker { width: 100% !important; }
.full-width-picker .el-input__wrapper { border-radius: 6px !important; box-shadow: 0 0 0 1px #dcdfe6 inset !important; padding: 8px 12px !important; }
.custom-textarea .el-textarea__inner { border-radius: 6px !important; padding: 10px !important; box-shadow: 0 0 0 1px #dcdfe6 inset !important; }
.custom-textarea .el-textarea__inner:focus { border-color: #764ba2; box-shadow: 0 0 0 2px rgba(118, 75, 162, 0.2) inset !important; }

.form-tip { font-size: 12px; color: #909399; margin-top: 6px; display: flex; align-items: center; gap: 4px; }
.form-tip.warning { color: #e6a23c; }
.divider-line { height: 1px; background: #f0f0f0; margin: 20px 0; }

/* 底部按钮 */
.dialog-footer-simple { padding: 0 30px 25px 30px; display: flex; justify-content: flex-end; gap: 12px; }
.btn-cancel { border: 1px solid #dcdfe6; background: #fff; color: #606266; height: 36px; border-radius: 6px; padding: 0 20px; }
.btn-cancel:hover { background: #f5f7fa; color: #303133; }
.btn-submit { height: 36px; border-radius: 6px; font-weight: bold; border: none; background: linear-gradient(90deg, #409eff, #337ecc); padding: 0 24px; }
.btn-submit:hover { opacity: 0.95; transform: translateY(-1px); }
.btn-submit.purple-btn { background: linear-gradient(90deg, #667eea, #764ba2); }
</style>
