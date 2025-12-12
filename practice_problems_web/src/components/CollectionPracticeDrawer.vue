<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="handleVisibleChange"
    width="1600px"
    top="5vh"
    destroy-on-close
    @open="handleOpen"
    class="question-dialog"
    :show-close="false"
  >
    <!-- 1. Header -->
    <template #header="{ close, titleId, titleClass }">
      <div class="custom-header">
        <div class="header-left">
          <span :id="titleId" :class="titleClass" class="dialog-title">
            <el-icon class="mr-2"><Trophy /></el-icon>
            {{ collectionName }} - 综合刷题
            <el-tag type="danger" effect="dark" class="ml-2" v-if="isExamMode">考试模式</el-tag>
            <el-tag type="primary" effect="plain" class="ml-2" v-else>练习模式</el-tag>
          </span>

          <!-- 统计信息栏 -->
          <div class="stats-bar">
            <div class="stat-item">
              <span class="label">进度:</span>
              <span class="value">{{ answeredCount }} / {{ questionList.length }}</span>
            </div>
            <el-divider direction="vertical" v-if="isExamMode" />
            <div class="stat-item timer" v-if="isExamMode">
              <el-icon><Timer /></el-icon>
              <span class="value">{{ timerDisplay }}</span>
            </div>
            <el-divider direction="vertical" />
            <div class="stat-item">
              <span class="label">题目数:</span>
              <el-select v-model="questionLimit" size="small" style="width: 100px;" @change="loadQuestions">
                <el-option label="10题" :value="10" />
                <el-option label="20题" :value="20" />
                <el-option label="30题" :value="30" />
                <el-option label="50题" :value="50" />
                <el-option label="100题" :value="100" />
                <el-option label="200题" :value="200" />
              </el-select>
            </div>
            <el-divider direction="vertical" />
            <div class="stat-item">
              <el-button :icon="RefreshRight" size="small" @click="handleRefresh" title="刷新题目">刷新</el-button>
            </div>
          </div>
        </div>
        
        <div class="header-actions">
          <!-- 考试提交按钮 -->
          <div v-if="isExamMode" class="exam-toolbar">
             <el-button v-if="!isExamSubmitted" type="success" @click="handleSubmitExam" :disabled="answeredCount === 0">
               交卷
             </el-button>
             <el-tag v-else type="success" effect="dark" size="large">得分: {{ examScore }}</el-tag>
          </div>

          <el-divider direction="vertical" />
          
          <div class="switch-group">
            <div class="shortcut-control">
              <el-switch v-model="enableShortcuts" active-text="快捷键" inactive-text="关" inline-prompt style="--el-switch-on-color: #13ce66; --el-switch-off-color: #909399" />
              <el-button v-if="enableShortcuts" circle size="small" icon="Setting" @click="showKeyConfig = true" title="配置" class="ml-5" />
            </div>
            <!-- 模式切换 -->
            <el-switch 
              v-model="isExamMode" 
              active-text="考试模式" 
              inactive-text="练习模式" 
              inline-prompt 
              style="--el-switch-on-color: #f56c6c; --el-switch-off-color: #409eff" 
            />
          </div>
          <el-button circle :icon="Close" @click="close" class="close-btn" />
        </div>
      </div>
    </template>

    <div class="drawer-content">
      <!-- 列表区域 -->
      <div class="question-list-container">
        <div class="scroll-area">
          <el-empty v-if="questionList.length === 0" description="当前集合下暂无题目" />
          <div 
            v-for="(q, index) in questionList" 
            :key="q.id" 
            class="q-card" 
            :class="{ 
              'is-answered': q.userResult.hasAnswered,
              'is-shortcut-active': enableShortcuts && hoveredQuestionId === q.id 
            }"
            @mouseenter="hoveredQuestionId = q.id"
            @mouseleave="hoveredQuestionId = null"
          >
            <div class="q-body">
              <div class="q-text">
                <span class="index-badge">{{ index + 1 }}</span>
                <span class="text-content">{{ q.questionText }}</span>
              </div>
              
              <div class="q-options-area">
                <div v-for="(opt, optIdx) in q.shuffledOptions" :key="optIdx" class="option-item" :class="getOptionClass(q, opt)" @click="handleAnswer(q, opt)">
                  <span class="opt-char">{{ getChar(optIdx) }}.</span><span class="opt-content">{{ opt.text }}</span>
                  <span v-if="enableShortcuts && hoveredQuestionId === q.id" class="shortcut-hint">[{{ getKeyDisplay(optIdx) }}]</span>
                  <el-icon v-if="shouldShowIcon(q, opt, 'correct')" class="status-icon correct"><Select /></el-icon>
                  <el-icon v-if="shouldShowIcon(q, opt, 'wrong')" class="status-icon wrong"><CloseBold /></el-icon>
                </div>
              </div>

              <el-collapse-transition>
                <div v-if="shouldShowAnalysis(q)" class="q-analysis-box">
                  <!-- 标准视图 -->
                  <div class="standard-view">
                    <div class="standard-title"><el-icon><List /></el-icon> 标准视图 (对应解析)</div>
                    <div class="standard-options-grid">
                      <div v-for="(stdOpt, stdIdx) in q.originalOptions" :key="stdIdx" class="std-option" :class="{ 'is-std-correct': (stdIdx + 1) === q.correctAnswer }">
                        <span class="std-char">{{ getChar(stdIdx) }}.</span><span class="std-content">{{ stdOpt.text }}</span>
                        <el-icon v-if="(stdIdx + 1) === q.correctAnswer" class="std-icon"><Select /></el-icon>
                      </div>
                    </div>
                  </div>
                  
                  <div class="analysis-row"><span class="tag">解析</span><div class="text">{{ q.explanation || '暂无解析' }}</div></div>
                  
                  <!-- 笔记区域 -->
                  <div class="analysis-row note-row">
                    <span class="tag note">我的笔记</span>
                    <div v-if="editingNoteId !== q.id" class="note-display">
                      <div class="text note-text">{{ q.note || '暂无笔记，点击右侧添加...' }}</div>
                      <el-button link type="primary" :icon="Edit" @click="startEditNote(q)">编辑</el-button>
                    </div>
                    <div v-else class="note-editor">
                      <el-input v-model="tempNoteContent" type="textarea" :rows="2" placeholder="输入你的理解..." />
                      <div class="note-actions">
                        <el-button size="small" @click="cancelEditNote">取消</el-button>
                        <el-button size="small" type="primary" @click="saveNote(q)">保存</el-button>
                      </div>
                    </div>
                  </div>
                  
                </div>
              </el-collapse-transition>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 快捷键配置 -->
    <el-dialog v-model="showKeyConfig" title="配置快捷键" width="400px" append-to-body class="key-config-dialog">
      <div class="key-config-tip">点击输入框并按键绑定</div>
      <div class="key-config-list">
        <div class="key-item" v-for="(label, idx) in ['A', 'B', 'C', 'D']" :key="idx">
          <span class="key-label">选项 {{ label }}</span>
          <el-input readonly :model-value="keyBindings[idx].toUpperCase()" @keydown.prevent="(e: any) => handleBindKey(e, idx)" class="key-input"><template #append>键</template></el-input>
        </div>
      </div>
      <template #footer><el-button @click="resetKeys">恢复默认</el-button><el-button type="primary" @click="saveKeys">保存</el-button></template>
    </el-dialog>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onUnmounted } from "vue";
import { Close, Select, CloseBold, List, Setting, Trophy, Edit, Timer, RefreshRight } from "@element-plus/icons-vue"; 
import { ElMessage, ElMessageBox } from "element-plus";
import { getCollectionQuestions, type CollectionQuestionItem } from "../api/collection";
import { updateUserNote } from "../api/question";

const props = defineProps<{ visible: boolean; collectionId: number; collectionName: string; }>();
const emit = defineEmits(["update:visible"]);
const handleVisibleChange = (val: boolean) => emit("update:visible", val);

interface FrontendQuestion extends CollectionQuestionItem {
  shuffledOptions: Array<{ text: string; originalIndex: number }>;
  originalOptions: Array<{ text: string; originalIndex: number }>;
  userResult: { hasAnswered: boolean; selectedOriginalIndex: number | null; isCorrect: boolean; };
}

const questionList = ref<FrontendQuestion[]>([]);
const questionLimit = ref(20); // 题目数量
const isExamMode = ref(false);
const isExamSubmitted = ref(false);
const examScore = ref("");

// 笔记编辑状态
const editingNoteId = ref<number | null>(null);
const tempNoteContent = ref("");

// --- 计时器逻辑 ---
const timerDisplay = ref("00:00:00");
let timerInterval: any = null;
let secondsElapsed = 0;

const startTimer = () => {
  stopTimer();
  secondsElapsed = 0;
  timerDisplay.value = "00:00:00";
  timerInterval = setInterval(() => {
    secondsElapsed++;
    const h = Math.floor(secondsElapsed / 3600).toString().padStart(2, '0');
    const m = Math.floor((secondsElapsed % 3600) / 60).toString().padStart(2, '0');
    const s = (secondsElapsed % 60).toString().padStart(2, '0');
    timerDisplay.value = `${h}:${m}:${s}`;
  }, 1000);
};

const stopTimer = () => {
  if (timerInterval) {
    clearInterval(timerInterval);
    timerInterval = null;
  }
};

// --- 快捷键逻辑 ---
const enableShortcuts = ref(false);
const hoveredQuestionId = ref<number | null>(null);
const showKeyConfig = ref(false);
const defaultKeys = ['a', 'b', 'c', 'd'];
const keyBindings = reactive<string[]>([...defaultKeys]);

onMounted(() => {
  const saved = localStorage.getItem('question_shortcuts');
  if (saved) { try { const p = JSON.parse(saved); if (Array.isArray(p)) p.forEach((k, i) => keyBindings[i] = k); } catch (e) {} }
  window.addEventListener('keydown', handleKeydown);
});
onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown);
  stopTimer();
});

const handleBindKey = (e: KeyboardEvent, index: number) => {
  const key = e.key.toLowerCase();
  if (key.length > 1 && !['enter', 'space', 'arrowup', 'arrowdown', 'arrowleft', 'arrowright'].includes(key)) return;
  keyBindings[index] = key; (e.target as HTMLElement).blur();
};
const resetKeys = () => defaultKeys.forEach((k, i) => keyBindings[i] = k);
const saveKeys = () => { localStorage.setItem('question_shortcuts', JSON.stringify(keyBindings)); showKeyConfig.value = false; };
const getKeyDisplay = (idx: number) => keyBindings[idx].toUpperCase();
const handleKeydown = (e: KeyboardEvent) => {
  if (!enableShortcuts.value) return;
  const activeTag = document.activeElement?.tagName;
  if (activeTag === 'INPUT' || activeTag === 'TEXTAREA') return;
  if (!hoveredQuestionId.value) return;
  const q = questionList.value.find(item => item.id === hoveredQuestionId.value);
  if (!q) return;
  const index = keyBindings.findIndex(k => k === e.key.toLowerCase());
  if (index !== -1 && q.shuffledOptions[index]) handleAnswer(q, q.shuffledOptions[index]);
};

// --- 核心逻辑 ---
const answeredCount = computed(() => questionList.value.filter(q => q.userResult.hasAnswered).length);

// 监听模式切换
watch(isExamMode, (newVal) => {
  if (newVal) {
    resetAllAnswers();
    isExamSubmitted.value = false;
    startTimer();
    ElMessage.info("考试开始，计时已启动");
  } else {
    stopTimer();
    isExamSubmitted.value = false;
  }
});

const handleOpen = () => {
  isExamMode.value = false;
  isExamSubmitted.value = false;
  stopTimer();
  loadQuestions();
}

// 刷新题目（二级确认）
const handleRefresh = () => {
  ElMessageBox.confirm(
    '刷新后将重新随机获取题目，当前答题进度将丢失，确认刷新？',
    '确认刷新',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    loadQuestions();
    ElMessage.success('题目已刷新');
  }).catch(() => {
    // 取消操作
  });
}

const resetAllAnswers = () => { questionList.value.forEach(q => { q.userResult.hasAnswered = false; q.userResult.selectedOriginalIndex = null; q.userResult.isCorrect = false; }); }

const loadQuestions = async () => {
  if (!props.collectionId) return;
  try {
    const res = await getCollectionQuestions(props.collectionId, questionLimit.value);
    if (res.data && (res.data as any).code === 200) {
      // 后端返回的题目已经是打乱过的，前端再打乱一次
      const questions = res.data.data;
      // 第二次打乱：前端再洗牌
      for (let i = questions.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        [questions[i], questions[j]] = [questions[j], questions[i]];
      }
      
      questionList.value = questions.map((item: CollectionQuestionItem) => {
        const rawOptions = [
          { text: item.option1, originalIndex: 1 }, { text: item.option2, originalIndex: 2 },
          { text: item.option3, originalIndex: 3 }, { text: item.option4, originalIndex: 4 },
        ];
        const shuffled = [...rawOptions];
        for (let i = shuffled.length - 1; i > 0; i--) {
          const j = Math.floor(Math.random() * (i + 1));
          [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
        }
        return { 
          ...item, shuffledOptions: shuffled, originalOptions: rawOptions,
          userResult: { hasAnswered: false, selectedOriginalIndex: null, isCorrect: false } 
        };
      });
    }
  } catch (error) { 
    ElMessage.error("加载失败"); 
  }
};

const handleAnswer = (q: FrontendQuestion, opt: { originalIndex: number }) => {
  if (q.userResult.hasAnswered) return;
  q.userResult.hasAnswered = true;
  q.userResult.selectedOriginalIndex = opt.originalIndex;
  q.userResult.isCorrect = (opt.originalIndex === q.correctAnswer);
};

const handleSubmitExam = () => {
  ElMessageBox.confirm(`确认交卷？`, '交卷', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }).then(() => {
    stopTimer();
    isExamSubmitted.value = true;
    const correctCount = questionList.value.filter(q => q.userResult.isCorrect).length;
    const total = questionList.value.length;
    examScore.value = `${correctCount} / ${total}`;
    
    ElMessage.success({
      message: `考试结束！用时：${timerDisplay.value}，得分：${correctCount} / ${total}`,
      duration: 5000
    });
  });
};

// --- 笔记编辑逻辑 ---
const startEditNote = (q: FrontendQuestion) => {
  editingNoteId.value = q.id;
  tempNoteContent.value = q.note || "";
};

const cancelEditNote = () => {
  editingNoteId.value = null;
  tempNoteContent.value = "";
};

const saveNote = async (q: FrontendQuestion) => {
  try {
    const res = await updateUserNote({ question_id: q.id, note: tempNoteContent.value });
    if ((res.data as any).code === 200) {
      ElMessage.success("笔记保存成功");
      q.note = tempNoteContent.value;
      editingNoteId.value = null;
    }
  } catch (e) {
    ElMessage.error("保存失败");
  }
};

// --- 样式辅助 ---
const getOptionClass = (q: FrontendQuestion, opt: any) => {
  if (!q.userResult.hasAnswered) return 'is-pending';
  if (isExamMode.value && !isExamSubmitted.value) {
    return q.userResult.selectedOriginalIndex === opt.originalIndex ? 'is-selected-exam' : 'is-disabled';
  }
  if (opt.originalIndex === q.correctAnswer) return 'is-correct-opt';
  if (q.userResult.selectedOriginalIndex === opt.originalIndex && !q.userResult.isCorrect) return 'is-wrong-opt';
  return 'is-disabled';
};
const shouldShowIcon = (q: FrontendQuestion, opt: any, type: string) => {
  if (!q.userResult.hasAnswered) return false;
  if (isExamMode.value && !isExamSubmitted.value) return false;
  if (type === 'correct') return opt.originalIndex === q.correctAnswer;
  if (type === 'wrong') return q.userResult.selectedOriginalIndex === opt.originalIndex && opt.originalIndex !== q.correctAnswer;
  return false;
};
const shouldShowAnalysis = (q: FrontendQuestion) => {
  if (!q.userResult.hasAnswered) return false;
  return isExamMode.value ? isExamSubmitted.value : true;
};
const getChar = (i: number) => String.fromCharCode(65 + i);
</script>

<style scoped>
/* 复用样式 */
.drawer-content { height: 75vh; display: flex; flex-direction: column; background: #f5f7fa; }
.custom-header { display: flex; justify-content: space-between; align-items: center; padding-right: 20px; }
.header-left { display: flex; align-items: center; gap: 20px; }
.dialog-title { font-size: 20px; font-weight: bold; display: flex; align-items: center; }

/* 统计栏样式 */
.stats-bar {
  display: flex;
  align-items: center;
  gap: 15px;
  background: #f0f2f5;
  padding: 6px 15px;
  border-radius: 20px;
  font-size: 14px;
}
.stat-item { display: flex; align-items: center; gap: 5px; }
.stat-item .label { color: #909399; }
.stat-item .value { font-weight: bold; color: #303133; font-family: monospace; font-size: 15px; }
.stat-item.timer .value { color: #e6a23c; }

.header-actions { display: flex; align-items: center; gap: 15px; }
.exam-toolbar { display: flex; align-items: center; gap: 15px; margin-right: 10px; }
.switch-group { display: flex; align-items: center; gap: 15px; }
.shortcut-control { display: flex; align-items: center; gap: 5px; margin-right: 10px; }
.ml-5 { margin-left: 5px; }
.ml-2 { margin-left: 10px; }

.question-list-container { flex: 1; overflow-y: auto; padding: 20px; }
.q-card { display: flex; background: #fff; border-radius: 8px; box-shadow: 0 2px 12px 0 rgba(0,0,0,0.05); margin-bottom: 20px; border: 1px solid #ebeef5; overflow: hidden; transition: all 0.2s; }
.q-body { flex: 1; padding: 20px 30px; }
.q-text { font-size: 18px; font-weight: 500; margin-bottom: 20px; }
.text-content { white-space: pre-wrap; line-height: 1.6; }
.index-badge { background: #409eff; color: #fff; padding: 2px 8px; border-radius: 4px; margin-right: 8px; font-size: 14px; }
.q-card.is-shortcut-active { border-color: #13ce66; box-shadow: 0 0 8px rgba(19, 206, 102, 0.2); }

.q-options-area { display: grid; grid-template-columns: 1fr 1fr; gap: 15px; margin-bottom: 20px; }
.option-item { border: 1px solid #dcdfe6; border-radius: 6px; padding: 12px 20px; cursor: pointer; display: flex; align-items: center; background: #fff; position: relative; }
.opt-char { font-weight: bold; margin-right: 12px; color: #909399; }
.status-icon { font-size: 18px; margin-left: 10px; }
.status-icon.correct { color: #67c23a; }
.status-icon.wrong { color: #f56c6c; }
.option-item.is-pending:hover { background-color: #ecf5ff; border-color: #409eff; color: #409eff; }
.shortcut-hint { position: absolute; right: 10px; font-size: 12px; color: #c0c4cc; font-weight: bold; background: #f5f7fa; padding: 2px 6px; border-radius: 4px; }

.option-item.is-correct-opt { background-color: #f0f9eb; border-color: #67c23a; color: #67c23a; font-weight: bold; }
.option-item.is-wrong-opt { background-color: #fef0f0; border-color: #f56c6c; color: #f56c6c; }
.option-item.is-disabled { opacity: 0.6; cursor: not-allowed; background: #f5f7fa; }
.option-item.is-selected-exam { background-color: #ecf5ff; border-color: #409eff; color: #409eff; font-weight: bold; }

.q-analysis-box { background: #fffbf0; border: 1px dashed #e6a23c; border-radius: 6px; padding: 15px; margin-top: 10px; }
.standard-view { margin-bottom: 15px; background: #fff; padding: 10px; border-radius: 4px; border: 1px solid #e4e7ed; }
.standard-title { font-weight: bold; font-size: 14px; color: #303133; margin-bottom: 8px; display: flex; align-items: center; gap: 5px; }
.standard-options-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.std-option { font-size: 13px; color: #606266; padding: 4px 8px; border-radius: 4px; display: flex; align-items: center; }
.std-option.is-std-correct { color: #67c23a; font-weight: bold; background: #f0f9eb; }
.std-char { margin-right: 5px; font-weight: bold; }
.std-icon { margin-left: 5px; }

.analysis-row { display: flex; margin-bottom: 10px; }
.tag { background: #e6a23c; color: #fff; font-size: 12px; padding: 2px 6px; border-radius: 4px; height: fit-content; margin-right: 10px; white-space: nowrap; }
.tag.note { background: #409eff; }
.text { font-size: 14px; color: #606266; line-height: 1.6; flex: 1; white-space: pre-wrap; word-break: break-all; }

.note-row { align-items: flex-start; }
.note-display { flex: 1; display: flex; justify-content: space-between; align-items: flex-start; }
.note-text { white-space: pre-wrap; }
.note-editor { flex: 1; display: flex; flex-direction: column; gap: 8px; }
.note-actions { display: flex; justify-content: flex-end; gap: 8px; }

.key-config-tip { text-align: center; color: #909399; margin-bottom: 15px; font-size: 13px; }
.key-config-list { display: flex; flex-direction: column; gap: 10px; }
.key-item { display: flex; align-items: center; justify-content: space-between; }
.key-label { font-weight: bold; color: #606266; }
.key-input { width: 200px; }
</style>
