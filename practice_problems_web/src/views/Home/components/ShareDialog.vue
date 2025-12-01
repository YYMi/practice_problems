<template>
  <el-dialog
    v-model="dialogVisible"
    title="资源分享与绑定"
    width="550px"
    :before-close="handleClose"
    destroy-on-close
  >
    <el-tabs v-model="activeTab" type="card" class="share-tabs">
      
      <!-- ================= 1. 我要分享 ================= -->
      <el-tab-pane label="我要分享" name="share">
        <el-form :model="shareForm" label-width="100px" label-position="top">
          
          <!-- 1. 选择科目 -->
          <el-form-item label="选择科目 (支持多选)">
            <el-select 
              v-model="shareForm.subjectIds" 
              multiple 
              collapse-tags
              collapse-tags-tooltip
              placeholder="请选择您创建的科目" 
              style="width: 100%"
            >
              <el-option
                v-for="item in mySubjects"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
            <div v-if="mySubjects.length === 0" class="form-tip">您还没有创建任何科目，无法分享。</div>
          </el-form-item>

          <!-- 2. 分享方式 -->
          <el-form-item label="分享方式">
            <el-radio-group v-model="shareForm.type">
              <el-radio :label="1">指定用户 (私密)</el-radio>
              <el-radio :label="2">生成分享码 (公开)</el-radio>
            </el-radio-group>
          </el-form-item>

          <!-- ============ 配置区域 A: 资源有效期 ============ -->
          <el-form-item label="资源有效期 (用户绑定后能看多久)">
            <div class="duration-input-row">
              <el-input 
                v-model="resCustomVal" 
                placeholder="时长" 
                type="number" 
                min="1"
                style="width: 180px"
                @input="handleResDurationChange"
              >
                <template #append>
                  <el-select v-model="resCustomUnit" style="width: 70px" @change="handleResDurationChange">
                    <el-option label="天" value="d" />
                    <el-option label="周" value="w" />
                    <el-option label="月" value="m" />
                    <el-option label="年" value="y" />
                  </el-select>
                </template>
              </el-input>
              <span class="hint-text">快捷：</span>
              <el-radio-group v-model="shareForm.duration" size="small" @change="handleResPresetChange">
                <el-radio-button label="7d">1周</el-radio-button>
                <el-radio-button label="30d">1月</el-radio-button>
                <el-radio-button label="365d">1年</el-radio-button>
                <el-radio-button label="forever">永久</el-radio-button>
              </el-radio-group>
            </div>
          </el-form-item>

          <!-- ============ 配置区域 B: 分享码有效期 (带安全校验) ============ -->
          <div v-if="shareForm.type === 2" class="sub-form-area">
            <el-form-item 
              label="分享码有效期 (码多久后失效)"
              :error="isCodeDurationInvalid ? '分享码有效期不能超过 1 年' : ''"
            >
               <div class="duration-input-row">
                <el-input 
                  v-model="codeCustomVal" 
                  placeholder="时长" 
                  type="number" 
                  min="1"
                  style="width: 180px"
                  @input="handleCodeDurationChange"
                >
                  <template #append>
                    <el-select v-model="codeCustomUnit" style="width: 70px" @change="handleCodeDurationChange">
                      <el-option label="天" value="d" />
                      <el-option label="周" value="w" />
                      <el-option label="月" value="m" />
                      <el-option label="年" value="y" />
                    </el-select>
                  </template>
                </el-input>
                <span class="hint-text">快捷：</span>
                <el-radio-group v-model="shareForm.codeDuration" size="small" @change="handleCodePresetChange">
                  <el-radio-button label="1d">1天</el-radio-button>
                  <el-radio-button label="7d">1周</el-radio-button>
                  <el-radio-button label="30d">1月</el-radio-button>
                  <el-radio-button label="365d">1年</el-radio-button>
                </el-radio-group>
              </div>
              <div class="form-tip">为了安全，分享码最长有效期为 1 年，不可设置永久。</div>
            </el-form-item>
          </div>

          <!-- 方式 A: 指定用户输入框 -->
          <div v-if="shareForm.type === 1" class="sub-form-area">
            <el-form-item label="目标用户账号 (User Code)">
              <el-input 
                v-model="shareForm.targetUsers" 
                type="textarea" 
                :rows="2"
                placeholder="例如: admin, 88291022" 
              />
            </el-form-item>
            <div class="form-tip">对方无需操作，刷新列表即可直接看到这 <span style="color:red;font-weight:bold">{{ shareForm.subjectIds.length }}</span> 个科目。</div>
          </div>

        </el-form>
        
        <div class="dialog-footer-btn">
          <!-- 禁用条件：没选科目 OR 码有效期超限 -->
          <el-button 
            type="primary" 
            @click="handleCreateShare" 
            :disabled="shareForm.subjectIds.length === 0 || isCodeDurationInvalid"
          >
            {{ shareForm.type === 1 ? '立即授权' : '生成分享码' }}
          </el-button>
        </div>

        <!-- 结果展示 -->
        <div v-if="generatedResult" class="result-box">
          <div class="result-title">
            <el-icon><SuccessFilled /></el-icon> 
            {{ shareForm.type === 1 ? '授权成功' : '分享码生成成功' }}
          </div>
          <div v-if="shareForm.type === 2" class="code-display">
            <span class="the-code">{{ generatedResult }}</span>
            <el-button type="primary" link @click="copyResult">复制</el-button>
          </div>
          <div v-else class="text-display">{{ generatedResult }}</div>
        </div>
      </el-tab-pane>

      <!-- ================= 2. 我要绑定 ================= -->
      <el-tab-pane label="绑定资源" name="bind">
        <div class="bind-wrapper">
          <div class="bind-icon-box">
            <el-icon size="40" color="#409eff"><Connection /></el-icon>
          </div>
          <el-form>
            <el-form-item label="请输入分享码">
              <el-input v-model="bindCode" placeholder="SHARE-XXXXXX" size="large" clearable prefix-icon="Key" />
            </el-form-item>
          </el-form>
          <el-button type="success" size="large" class="w-100" @click="handleBindSubject" :disabled="!bindCode">
            立即绑定
          </el-button>
        </div>
      </el-tab-pane>

    </el-tabs>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { SuccessFilled, Connection, Key } from '@element-plus/icons-vue';
import { createShare, bindShare } from '../../../api/share'; 

const props = defineProps(['visible', 'subjects', 'userInfo']);
const emit = defineEmits(['update:visible', 'refresh']);

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
});

const activeTab = ref('share');
const generatedResult = ref(''); 
const bindCode = ref('');

// ============ 状态管理 ============
const resCustomVal = ref('1'); 
const resCustomUnit = ref('w'); 

const codeCustomVal = ref('3'); 
const codeCustomUnit = ref('d');
const isCodeDurationInvalid = ref(false); // 标记是否超限

const shareForm = reactive({
  subjectIds: [] as number[],
  type: 1, 
  targetUsers: '',
  duration: '7d',       
  codeDuration: '3d'    
});

const mySubjects = computed(() => {
  if (!props.subjects || !props.userInfo) return [];
  return props.subjects.filter((s: any) => s.creatorCode === props.userInfo.user_code);
});

// 重置
watch(() => props.visible, (val) => {
  if (val) {
    generatedResult.value = '';
    bindCode.value = '';
    shareForm.targetUsers = '';
    shareForm.subjectIds = []; 
    isCodeDurationInvalid.value = false;
    
    shareForm.duration = '7d';
    resCustomVal.value = '1'; resCustomUnit.value = 'w';

    shareForm.codeDuration = '3d';
    codeCustomVal.value = '3'; codeCustomUnit.value = 'd';
  }
});

const handleClose = () => {
  emit('update:visible', false);
};

// ============ 逻辑 A: 资源有效期联动 ============
const handleResDurationChange = () => {
  if (resCustomVal.value) shareForm.duration = `${resCustomVal.value}${resCustomUnit.value}`;
};
const handleResPresetChange = (val: string) => {
  if (val === 'forever') { resCustomVal.value = ''; return; }
  smartFillInput(val, resCustomVal, resCustomUnit);
};

// ============ 逻辑 B: 分享码有效期联动 (带校验) ============
const handleCodeDurationChange = () => {
  if (!codeCustomVal.value) return;

  // 1. 计算天数
  const num = parseInt(codeCustomVal.value);
  const unit = codeCustomUnit.value;
  let days = 0;

  if (unit === 'd') days = num;
  else if (unit === 'w') days = num * 7;
  else if (unit === 'm') days = num * 30;
  else if (unit === 'y') days = num * 365;

  // 2. 校验是否超过 366 天
  if (days > 366) {
    isCodeDurationInvalid.value = true;
  } else {
    isCodeDurationInvalid.value = false;
    shareForm.codeDuration = `${codeCustomVal.value}${codeCustomUnit.value}`;
  }
};

const handleCodePresetChange = (val: string) => {
  isCodeDurationInvalid.value = false; // 快捷键肯定是合法的
  smartFillInput(val, codeCustomVal, codeCustomUnit);
  shareForm.codeDuration = val;
};

// 通用智能回填
const smartFillInput = (val: string, numRef: any, unitRef: any) => {
  if (val === '7d') {
    numRef.value = '1'; unitRef.value = 'w';
  } else if (val === '30d') {
    numRef.value = '1'; unitRef.value = 'm';
  } else if (val === '365d') {
    numRef.value = '1'; unitRef.value = 'y';
  } else {
    const match = val.match(/^(\d+)([a-z]+)$/);
    if (match) {
      numRef.value = match[1];
      unitRef.value = match[2];
    }
  }
};

// ============ 提交逻辑 ============
const handleCreateShare = async () => {
  if (shareForm.subjectIds.length === 0) return;
  if (isCodeDurationInvalid.value) return; // 双重保险

  try {
    const params = {
      subject_ids: shareForm.subjectIds,
      type: shareForm.type,
      duration: shareForm.duration,          
      code_duration: shareForm.codeDuration, 
      targets: shareForm.type === 1 ? shareForm.targetUsers.split(/,|，|\s+/).filter(Boolean) : []
    };

    const res = await createShare(params);

    if (res.data && res.data.code === 200) {
      if (shareForm.type === 1) {
        generatedResult.value = res.data.msg; 
        ElMessage.success("授权成功");
      } else {
        generatedResult.value = res.data.data || ''; 
        ElMessage.success("分享码生成成功");
      }
    } else {
      ElMessage.error(res.data?.msg || "操作失败");
    }
  } catch (e) {
    console.error(e);
  }
};

const handleBindSubject = async () => {
  if (!bindCode.value) return;
  try {
    const res = await bindShare({ code: bindCode.value });
    if (res.data && res.data.code === 200) {
      ElMessage.success(res.data.msg);
      emit('refresh'); 
      handleClose();
    } else {
      ElMessage.error(res.data?.msg || "绑定失败");
    }
  } catch (e) { console.error(e); }
};

const copyResult = () => {
  navigator.clipboard.writeText(generatedResult.value);
  ElMessage.success("已复制");
};
</script>

<style scoped>
.share-tabs { margin-top: -10px; }
.form-tip { font-size: 12px; color: #909399; margin-top: 5px; line-height: 1.4; }
.sub-form-area { background: #f8f9fa; padding: 15px; border-radius: 6px; margin-bottom: 18px; border: 1px dashed #dcdfe6; }
.dialog-footer-btn { text-align: right; margin-top: 20px; }
.duration-input-row { display: flex; align-items: center; margin-bottom: 5px; gap: 10px; }
.hint-text { font-size: 12px; color: #606266; white-space: nowrap; }
.result-box { margin-top: 20px; background: #f0f9eb; padding: 20px; border-radius: 8px; border: 1px solid #e1f3d8; text-align: center; }
.result-title { font-size: 14px; color: #67c23a; font-weight: bold; margin-bottom: 10px; display: flex; align-items: center; justify-content: center; gap: 5px; }
.code-display { display: flex; align-items: center; justify-content: center; gap: 10px; background: #fff; padding: 10px; border-radius: 4px; border: 1px dashed #67c23a; }
.the-code { font-family: monospace; font-size: 24px; font-weight: bold; color: #303133; letter-spacing: 2px; }
.text-display { font-size: 14px; color: #606266; }
.bind-wrapper { padding: 20px 30px; }
.bind-icon-box { text-align: center; margin-bottom: 20px; }
.w-100 { width: 100%; }
</style>