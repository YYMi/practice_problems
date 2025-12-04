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
            {{ title }}
            <el-tag v-if="pointId" size="small" type="info" class="ml-2">知识点特训</el-tag>
            <el-tag v-else size="small" type="warning" class="ml-2">分类综合练习</el-tag>
          </span>

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
          </div>
        </div>
        
        <div class="header-actions">
          <div v-if="isExamMode" class="exam-toolbar">
            <el-button v-if="!isExamSubmitted" type="success" @click="handleSubmitExam" :disabled="answeredCount === 0">交卷查看结果</el-button>
            <!-- ★★★ 修复：得分显示更直观 ★★★ -->
            <el-tag v-else type="danger" effect="dark" size="large" class="score-tag">
              得分: {{ examScore }}
            </el-tag>
          </div>

          <div v-if="isBatchMode && hasPermission" class="batch-toolbar">
            <el-checkbox v-model="isAllSelected" :indeterminate="isIndeterminate" @change="handleSelectAll" class="select-all-checkbox">全选</el-checkbox>
            <el-button type="warning" size="small" :icon="Edit" :disabled="selectedIds.length !== 1" @click="handleEditQuestion">修改</el-button>
            <el-button type="danger" size="small" :icon="Delete" :disabled="selectedIds.length === 0" @click="handleBatchDelete">删除</el-button>
          </div>

          <el-divider direction="vertical" />
          
          <div class="switch-group">
            <div class="shortcut-control">
              <el-switch v-model="enableShortcuts" active-text="快捷键" inactive-text="关" inline-prompt style="--el-switch-on-color: #13ce66; --el-switch-off-color: #909399;" />
              <el-button v-if="enableShortcuts" circle size="small" :icon="Setting" @click="showKeyConfig = true" title="配置" class="ml-2" />
            </div>
            
            <el-switch v-model="isExamMode" active-text="考试" inactive-text="练习" inline-prompt style="--el-switch-on-color: #f56c6c; --el-switch-off-color: #409eff;" />
            
            <el-switch v-if="hasPermission" v-model="isBatchMode" :disabled="isExamMode" active-text="管理" inactive-text="视图" inline-prompt style="--el-switch-on-color: #e6a23c; --el-switch-off-color: #909399;" />
            
            <el-tooltip :disabled="!!pointId" content="请进入具体知识点进行录入" placement="bottom">
              <div class="inline-block">
                <el-button 
                  v-if="hasPermission" 
                  :type="showAddForm ? 'info' : 'primary'" 
                  :icon="showAddForm ? Close : Plus" 
                  @click="toggleAddForm" 
                  :disabled="isExamMode || (!pointId && !isEditMode)"
                >
                  {{ showAddForm ? "取消" : "录入" }}
                </el-button>
              </div>
            </el-tooltip>
          </div>
          <el-button circle :icon="Close" @click="close" class="close-btn" />
        </div>
      </div>
    </template>

    <div class="drawer-content">
      <!-- Form -->
      <el-collapse-transition>
        <div v-show="showAddForm" class="form-wrapper">
          <div class="form-box" :class="{ 'is-edit': isEditMode }">
            <div class="form-title-bar">
              <div class="form-title-text">
                {{ isEditMode ? "修改题目" : "录入新题" }} 
                <span v-if="isEditMode" class="edit-tip">(ID: {{ editingId }})</span>
              </div>
              <el-button v-if="!isEditMode" size="small" type="warning" plain @click="showJsonImport = !showJsonImport">{{ showJsonImport ? "关闭导入" : "JSON 导入" }}</el-button>
            </div>

            <el-collapse-transition>
              <div v-if="showJsonImport" class="json-import-area">
                <div class="json-header-tool">
                  <div class="json-tip">支持单个对象或数组对象。</div>
                  <el-button size="small" type="primary" link :icon="CopyDocument" @click="copyJsonTemplate">复制模板</el-button>
                </div>
                <el-input v-model="jsonContent" type="textarea" :rows="6" placeholder="粘贴 JSON..." />
                <div class="json-actions">
                  <el-button size="small" @click="jsonContent = ''">清空</el-button>
                  <el-button size="small" type="primary" :loading="isImporting" @click="handleParseJson">导入</el-button>
                </div>
              </div>
            </el-collapse-transition>

            <div class="form-grid">
              <div class="form-left">
                <div class="input-group"><div class="label">题目描述</div><el-input v-model="form.questionText" type="textarea" :rows="5" /></div>
                <div class="input-group"><div class="label">答案解析</div><el-input v-model="form.explanation" type="textarea" :rows="5" /></div>
              </div>
              <div class="form-right">
                <div class="label">选项设置</div>
                <div class="option-inputs">
                  <el-input v-model="form.option1"><template #prepend>A</template></el-input>
                  <el-input v-model="form.option2"><template #prepend>B</template></el-input>
                  <el-input v-model="form.option3"><template #prepend>C</template></el-input>
                  <el-input v-model="form.option4"><template #prepend>D</template></el-input>
                </div>
                <div class="correct-select">
                  <span class="label-inline">正确答案：</span>
                  <el-radio-group v-model="form.correctAnswer">
                    <el-radio-button :label="1">A</el-radio-button><el-radio-button :label="2">B</el-radio-button>
                    <el-radio-button :label="3">C</el-radio-button><el-radio-button :label="4">D</el-radio-button>
                  </el-radio-group>
                  <span v-if="!form.correctAnswer" style="color: #f56c6c; font-size: 12px; margin-left: 10px; font-weight: bold;">* 必选</span>
                </div>
                <div class="form-btns">
                  <el-button @click="closeAddForm">取消</el-button>
                  <el-button type="primary" @click="handleSubmit">保存</el-button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-collapse-transition>

      <!-- List -->
      <div class="question-list-container">
        <div class="scroll-area">
          <el-empty v-if="questionList.length === 0" description="暂无题目" />
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
            <div v-if="isBatchMode && hasPermission" class="q-check">
              <el-checkbox v-model="q.isChecked" size="large" />
            </div>
            
            <div class="q-body">
              <div class="q-text">
                <span class="index-badge">{{ index + 1 }}</span>
                <span class="text-content">{{ q.questionText }}</span>
              </div>
              
              <div class="q-options-area">
                <div 
                  v-for="(opt, optIdx) in q.shuffledOptions" 
                  :key="optIdx" 
                  class="option-item" 
                  :class="getOptionClass(q, opt)" 
                  @click="handleAnswer(q, opt)"
                >
                  <span class="opt-char">{{ getChar(optIdx) }}.</span>
                  <span class="opt-content">{{ opt.text }}</span>
                  
                  <span v-if="enableShortcuts && hoveredQuestionId === q.id" class="shortcut-hint">
                    [{{ getKeyDisplay(optIdx) }}]
                  </span>
                  
                  <!-- 状态图标 -->
                  <el-icon v-if="shouldShowIcon(q, opt, 'correct')" class="status-icon correct"><Select /></el-icon>
                  <el-icon v-if="shouldShowIcon(q, opt, 'wrong')" class="status-icon wrong"><CloseBold /></el-icon>
                </div>
              </div>

              <el-collapse-transition>
                <div v-if="shouldShowAnalysis(q)" class="q-analysis-box">
                  <div class="standard-view">
                    <div class="standard-title"><el-icon><List /></el-icon> 标准视图</div>
                    <div class="standard-options-grid">
                      <div v-for="(stdOpt, stdIdx) in q.originalOptions" :key="stdIdx" class="std-option" :class="{ 'is-std-correct': stdIdx + 1 === q.correctAnswer }">
                        <span class="std-char">{{ getChar(stdIdx) }}.</span><span class="std-content">{{ stdOpt.text }}</span>
                        <el-icon v-if="stdIdx + 1 === q.correctAnswer" class="std-icon"><Select /></el-icon>
                      </div>
                    </div>
                  </div>
                  
                  <div class="analysis-row"><span class="tag">解析</span><div class="text">{{ q.explanation || "暂无解析" }}</div></div>
                  
                  <div class="analysis-row note-row">
                    <span class="tag note">笔记</span>
                    <div v-if="editingNoteId !== q.id" class="note-display">
                      <div class="text note-text">{{ q.note || "暂无笔记..." }}</div>
                      <el-button link type="primary" :icon="Edit" @click="startEditNote(q)">编辑</el-button>
                    </div>
                    <div v-else class="note-editor">
                      <el-input v-model="tempNoteContent" type="textarea" :rows="2" />
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

    <!-- 快捷键配置 (修复样式) -->
    <el-dialog v-model="showKeyConfig" title="配置答题快捷键" width="400px" append-to-body class="key-config-dialog">
      <div class="key-config-tip">点击输入框并按下键盘上的任意键进行绑定</div>
      <div class="key-config-list">
        <div class="key-item" v-for="(label, idx) in ['A', 'B', 'C', 'D']" :key="idx">
          <span class="key-label">选项 {{ label }}</span>
          <div class="key-input-wrapper">
            <input 
              class="key-input-inner" 
              readonly 
              :value="keyBindings[idx].toUpperCase()" 
              @keydown.prevent.stop="(e) => handleBindKey(e, idx)"
              placeholder="按下按键"
            />
          </div>
        </div>
      </div>
      <template #footer><el-button @click="resetKeys">恢复默认</el-button><el-button type="primary" @click="saveKeys">保存</el-button></template>
    </el-dialog>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onUnmounted } from "vue";
import { Plus, Close, Edit, Delete, Select, CloseBold, Reading, List, Setting, CopyDocument, Timer, Trophy } from "@element-plus/icons-vue";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  getQuestions,
  getQuestionsByCategory,
  createQuestion,
  updateQuestion,
  deleteQuestion,
  updateUserNote, // <--- 新增它！修改备注要用
  type QuestionItem,
} from "../api/question";

const props = defineProps<{
  visible: boolean;
  pointId?: number;    
  categoryId?: number;
  title: string;
  viewMode: string; 
  isOwner: boolean; 
}>();

const emit = defineEmits(["update:visible"]);
const handleVisibleChange = (val: boolean) => emit("update:visible", val);

interface FrontendQuestion extends QuestionItem {
  isChecked: boolean;
  shuffledOptions: Array<{ text: string; originalIndex: number }>;
  originalOptions: Array<{ text: string; originalIndex: number }>;
  userResult: { hasAnswered: boolean; selectedOriginalIndex: number | null; isCorrect: boolean; };
}

const questionList = ref<FrontendQuestion[]>([]);
const showAddForm = ref(false);
const isBatchMode = ref(false);
const isExamMode = ref(false);
const isExamSubmitted = ref(false);
const isEditMode = ref(false);
const editingId = ref<number | null>(null);
const editingNoteId = ref<number | null>(null);
const tempNoteContent = ref("");
const examScore = ref("");

const hasPermission = computed(() => {
  if (props.viewMode === 'read') return false;
  if (props.viewMode === 'dev') return true;
  return props.isOwner;
});

const timerDisplay = ref("00:00:00");
let timerInterval: any = null;
let secondsElapsed = 0;
const startTimer = () => {
  stopTimer(); secondsElapsed = 0; timerDisplay.value = "00:00:00";
  timerInterval = setInterval(() => {
    secondsElapsed++;
    const h = Math.floor(secondsElapsed / 3600).toString().padStart(2, '0');
    const m = Math.floor((secondsElapsed % 3600) / 60).toString().padStart(2, '0');
    const s = (secondsElapsed % 60).toString().padStart(2, '0');
    timerDisplay.value = `${h}:${m}:${s}`;
  }, 1000);
};
const stopTimer = () => { if (timerInterval) { clearInterval(timerInterval); timerInterval = null; } };

const form = reactive({
  questionText: "", option1: "", option2: "", option3: "", option4: "", 
  correctAnswer: undefined as number | undefined, 
  explanation: "",
});

const showJsonImport = ref(false);
const jsonContent = ref("");
const isImporting = ref(false);
const copyJsonTemplate = () => { 
  const t = `[{"topic":"题目","answer1":"A","answer2":"B","answer3":"C","answer4":"D","proper":"1","resolve":"解析"}]`;
  navigator.clipboard.writeText(t).then(() => ElMessage.success("复制成功"));
};
const handleParseJson = async () => {
  if (!jsonContent.value) return ElMessage.warning("请先粘贴 JSON 内容");
  try {
    const data = JSON.parse(jsonContent.value);
    if (Array.isArray(data)) {
      isImporting.value = true;
      let successCount = 0;
      for (const item of data) {
        let importCorrectAnswer = 1; 
        if (item.proper) {
          const n = Number(item.proper);
          if (!isNaN(n) && n >= 1 && n <= 4) importCorrectAnswer = n;
        }
        const payload = {
          knowledgePointId: props.pointId || 0,
          questionText: item.topic || "",
          option1: item.answer1 || "", option2: item.answer2 || "", option3: item.answer3 || "", option4: item.answer4 || "",
          correctAnswer: importCorrectAnswer,
          explanation: item.resolve || "",
          note: "",
        };
        if (payload.questionText && payload.option1) {
          try { await createQuestion(payload); successCount++; } catch (err) {}
        }
      }
      isImporting.value = false;
      ElMessage.success(`导入成功 ${successCount} 条`);
      jsonContent.value = ""; showJsonImport.value = false; await loadQuestions();
    } else {
      if (data.topic) form.questionText = data.topic;
      if (data.answer1) form.option1 = data.answer1;
      if (data.answer2) form.option2 = data.answer2;
      if (data.answer3) form.option3 = data.answer3;
      if (data.answer4) form.option4 = data.answer4;
      if (data.resolve) form.explanation = data.resolve;
      if (data.proper) form.correctAnswer = Number(data.proper);
      ElMessage.success("已填充"); showJsonImport.value = false;
    }
  } catch (e) {
    isImporting.value = false; ElMessage.error("JSON 格式错误");
  }
};

const enableShortcuts = ref(false);
const hoveredQuestionId = ref<number | null>(null);
const showKeyConfig = ref(false);
const defaultKeys = ["a", "b", "c", "d"];
const keyBindings = reactive<string[]>([...defaultKeys]);
onMounted(() => {
  const saved = localStorage.getItem("question_shortcuts");
  if (saved) { try { const p = JSON.parse(saved); if (Array.isArray(p)) p.forEach((k, i) => keyBindings[i] = k); } catch (e) {} }
  window.addEventListener("keydown", handleKeydown);
});
onUnmounted(() => { window.removeEventListener("keydown", handleKeydown); stopTimer(); });

const handleBindKey = (e: any, index: number) => {
  const key = e.key.toLowerCase();
  if (key.length > 1 && !['enter', 'space', 'arrowup', 'arrowdown', 'arrowleft', 'arrowright'].includes(key)) return;
  keyBindings[index] = key;
  if (e.target) e.target.blur();
};
const resetKeys = () => { defaultKeys.forEach((k, i) => keyBindings[i] = k); };
const saveKeys = () => { localStorage.setItem("question_shortcuts", JSON.stringify(keyBindings)); showKeyConfig.value = false; ElMessage.success("设置已保存"); };
const getKeyDisplay = (idx: number) => keyBindings[idx].toUpperCase();
const handleKeydown = (e: KeyboardEvent) => {
  if (!enableShortcuts.value) return;
  const activeTag = document.activeElement?.tagName;
  if (activeTag === "INPUT" || activeTag === "TEXTAREA") return;
  if (!hoveredQuestionId.value) return;
  const q = questionList.value.find(item => item.id === hoveredQuestionId.value);
  if (!q) return;
  const index = keyBindings.findIndex(k => k === e.key.toLowerCase());
  if (index !== -1 && q.shuffledOptions[index]) handleAnswer(q, q.shuffledOptions[index]);
};

const selectedIds = computed(() => questionList.value.filter((q) => q.isChecked).map((q) => q.id));
const isAllSelected = computed(() => questionList.value.length > 0 && questionList.value.every((q) => q.isChecked));
const isIndeterminate = computed(() => selectedIds.value.length > 0 && selectedIds.value.length < questionList.value.length);
const answeredCount = computed(() => questionList.value.filter((q) => q.userResult.hasAnswered).length);

watch(isExamMode, (newVal) => {
  if (newVal) {
    isBatchMode.value = false; closeAddForm(); resetAllAnswers(); isExamSubmitted.value = false; startTimer(); ElMessage.info("考试开始");
  } else {
    stopTimer(); isExamSubmitted.value = false;
  }
});
watch(isBatchMode, (newVal) => { if (newVal && showAddForm.value) closeAddForm(); });

const handleOpen = () => {
  isBatchMode.value = false; isExamMode.value = false; isExamSubmitted.value = false;
  closeAddForm(); stopTimer(); loadQuestions();
};
const closeAddForm = () => { showAddForm.value = false; showJsonImport.value = false; resetForm(); };
const resetForm = () => {
  isEditMode.value = false; editingId.value = null;
  form.questionText = ""; form.option1 = ""; form.option2 = ""; form.option3 = ""; form.option4 = ""; form.correctAnswer = undefined; form.explanation = "";
};
const resetAllAnswers = () => {
  questionList.value.forEach((q) => {
    q.userResult.hasAnswered = false; q.userResult.selectedOriginalIndex = null; q.userResult.isCorrect = false;
  });
};

const loadQuestions = async () => {
  questionList.value = [];
  let res;
  try {
    if (props.pointId) res = await getQuestions(props.pointId);
    else if (props.categoryId) res = await getQuestionsByCategory(props.categoryId);
    else return;

    if (res.data && (res.data as any).code === 200) {
      questionList.value = res.data.data.map((item: QuestionItem) => {
        const rawOptions = [{ text: item.option1, originalIndex: 1 }, { text: item.option2, originalIndex: 2 }, { text: item.option3, originalIndex: 3 }, { text: item.option4, originalIndex: 4 }];
        const shuffled = [...rawOptions];
        for (let i = shuffled.length - 1; i > 0; i--) {
          const j = Math.floor(Math.random() * (i + 1));
          [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
        }
        return { ...item, isChecked: false, shuffledOptions: shuffled, originalOptions: rawOptions, userResult: { hasAnswered: false, selectedOriginalIndex: null, isCorrect: false } };
      });
    }
  } catch (error) { ElMessage.error("加载失败"); }
};

// ★★★ 修复：考试模式逻辑 ★★★
const handleAnswer = (q: FrontendQuestion, opt: { originalIndex: number }) => {
  // 考试模式且未交卷：允许修改答案，且不标记为“已出结果”
  if (isExamMode.value && !isExamSubmitted.value) {
    q.userResult.hasAnswered = true;
    q.userResult.selectedOriginalIndex = opt.originalIndex;
    return;
  }

  // 练习模式：选了就定死，显示对错
  if (q.userResult.hasAnswered) return;
  q.userResult.hasAnswered = true;
  q.userResult.selectedOriginalIndex = opt.originalIndex;
  q.userResult.isCorrect = opt.originalIndex === q.correctAnswer;
};

// ★★★ 修复：得分计算 ★★★
const handleSubmitExam = () => {
  ElMessageBox.confirm(`确认交卷？`, "交卷", { confirmButtonText: "确定", cancelButtonText: "取消", type: "warning" }).then(() => {
    stopTimer(); isExamSubmitted.value = true;
    // 交卷时统一计算对错
    questionList.value.forEach(q => {
      q.userResult.isCorrect = q.userResult.selectedOriginalIndex === q.correctAnswer;
    });
    
    const correctCount = questionList.value.filter(q => q.userResult.isCorrect).length;
    const total = questionList.value.length;
    const percent = total > 0 ? Math.round((correctCount / total) * 100) : 0;
    
    examScore.value = `${correctCount} / ${total} ( ${percent}% )`;
    ElMessage.success(`考试结束！`);
  });
};

// ★★★ 修复：样式逻辑，增强区分度 ★★★
const getOptionClass = (q: FrontendQuestion, opt: any) => {
  // 1. 考试模式未交卷：只显示选中状态 (蓝色)，不显示对错
  if (isExamMode.value && !isExamSubmitted.value) {
    return q.userResult.selectedOriginalIndex === opt.originalIndex ? 'is-selected-exam' : '';
  }

  // 2. 练习模式 或 考试已交卷：显示对错
  if (!q.userResult.hasAnswered) return 'is-pending'; // 没答题

  // 正确答案：永远绿色
  if (opt.originalIndex === q.correctAnswer) return 'is-correct-opt';
  
  // 我选错了：显示红色
  if (q.userResult.selectedOriginalIndex === opt.originalIndex && !q.userResult.isCorrect) return 'is-wrong-opt';
  
  // 其他没选的干扰项：变灰
  return 'is-disabled';
};

const shouldShowIcon = (q: FrontendQuestion, opt: any, type: string) => { if (!q.userResult.hasAnswered) return false; if (isExamMode.value && !isExamSubmitted.value) return false; if (type === 'correct') return opt.originalIndex === q.correctAnswer; if (type === 'wrong') return q.userResult.selectedOriginalIndex === opt.originalIndex && opt.originalIndex !== q.correctAnswer; return false; };
const shouldShowAnalysis = (q: FrontendQuestion) => { if (!q.userResult.hasAnswered) return false; return isExamMode.value ? isExamSubmitted.value : true; };
const startEditNote = (q: FrontendQuestion) => { editingNoteId.value = q.id; tempNoteContent.value = q.note || ""; };
const cancelEditNote = () => { editingNoteId.value = null; };
const saveNote = async (q: FrontendQuestion) => {
  try {
     // ★★★ 以前是调用 updateQuestion，现在改成调用 updateUserNote ★★★
    const res = await updateUserNote({
      question_id: q.id,
      note: tempNoteContent.value
    });

    // 判断返回结果 (根据你的 request 封装，可能是 res.data.code 或 res.code)
    if ((res as any).code === 200 || (res.data as any).code === 200) {
      ElMessage.success("笔记保存成功");
      
      // 更新前端视图
      q.note = tempNoteContent.value;
      editingNoteId.value = null;
    }
  } catch (e) {
    ElMessage.error("保存失败");
  }
};
const toggleAddForm = () => { showAddForm.value ? closeAddForm() : (showAddForm.value = true); };
const handleEditQuestion = () => {
  if (selectedIds.value.length !== 1) return;
  const target = questionList.value.find((q) => q.id === selectedIds.value[0]);
  if (target) {
    form.questionText = target.questionText;
    form.option1 = target.option1; form.option2 = target.option2; form.option3 = target.option3; form.option4 = target.option4;
    form.correctAnswer = target.correctAnswer;
    form.explanation = target.explanation;
    isEditMode.value = true; editingId.value = target.id; showAddForm.value = true; isBatchMode.value = false;
  }
};

const handleSubmit = async () => {
  if (!form.questionText) return ElMessage.warning("请完善题目描述");
  if (!form.correctAnswer) return ElMessage.warning("请选择正确答案！");

  const payload = { 
    knowledgePointId: props.pointId || 0, 
    ...form, 
    correctAnswer: form.correctAnswer, 
  };
  
  if (isEditMode.value && editingId.value) {
    const t = questionList.value.find((q) => q.id === editingId.value);
    if (t) payload.knowledgePointId = t.knowledgePointId;
  }
  try {
    if (isEditMode.value && editingId.value) await updateQuestion(editingId.value, payload);
    else await createQuestion(payload);
    ElMessage.success("操作成功"); closeAddForm(); loadQuestions();
  } catch (e) { ElMessage.error("操作失败"); }
};

const handleBatchDelete = () => {
  ElMessageBox.confirm(`确认删除?`, "警告", { type: "warning" }).then(async () => {
    try {
      for (const id of selectedIds.value) await deleteQuestion(id);
      ElMessage.success("删除成功"); loadQuestions();
    } catch (e) { ElMessage.error("删除失败"); }
  });
};
const handleSelectAll = (val: boolean) => { questionList.value.forEach((q) => (q.isChecked = val)); };
const getChar = (i: number) => String.fromCharCode(65 + i);
</script>

<style scoped>
/* ... Layout ... */
.drawer-content { height: 75vh; display: flex; flex-direction: column; background: #f5f7fa; }
.custom-header { display: flex; justify-content: space-between; align-items: center; padding-right: 20px; }
.header-left { display: flex; align-items: center; gap: 20px; }
.dialog-title { font-size: 20px; font-weight: bold; display: flex; align-items: center; }
.stats-bar { display: flex; align-items: center; gap: 15px; background: #f0f2f5; padding: 6px 15px; border-radius: 20px; font-size: 14px; }
.stat-item { display: flex; align-items: center; gap: 5px; }
.stat-item .label { color: #909399; }
.stat-item .value { font-weight: bold; color: #303133; font-family: monospace; font-size: 15px; }
.stat-item.timer .value { color: #e6a23c; }
.header-actions { display: flex; align-items: center; gap: 15px; }
.exam-toolbar { display: flex; align-items: center; gap: 15px; margin-right: 10px; }
.batch-toolbar { display: flex; align-items: center; gap: 10px; background: #fff; padding: 4px 12px; border-radius: 4px; border: 1px solid #dcdfe6; }
.switch-group { display: flex; align-items: center; gap: 15px; }
.shortcut-control { display: flex; align-items: center; gap: 5px; margin-right: 10px; }
.ml-2 { margin-left: 10px; }
.ml-5 { margin-left: 5px; }
.inline-block { display: inline-block; }

/* ... Form ... */
.form-wrapper { padding: 15px 20px 0 20px; background: #fff; border-bottom: 1px solid #e4e7ed; }
.form-box { background: #f0f9eb; border: 1px solid #e1f3d8; border-radius: 8px; padding: 20px; margin-bottom: 15px; }
.form-box.is-edit { background: #fdf6ec; border-color: #faecd8; }
.form-title-bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px; border-bottom: 1px dashed #ccc; padding-bottom: 10px; }
.form-title-text { font-weight: bold; font-size: 16px; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 30px; }
.input-group { margin-bottom: 12px; }
.label { font-weight: bold; margin-bottom: 5px; color: #606266; }
.option-inputs { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin-bottom: 15px; }
.correct-select { display: flex; align-items: center; margin-bottom: 15px; }
.label-inline { font-weight: bold; margin-right: 10px; color: #67c23a; }
.form-btns { display: flex; justify-content: flex-end; gap: 10px; }
.json-import-area { background: #fff; padding: 15px; border: 1px dashed #e6a23c; border-radius: 6px; margin-bottom: 20px; }
.json-header-tool { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.json-tip { font-size: 12px; color: #909399; }
.json-actions { display: flex; justify-content: flex-end; margin-top: 10px; gap: 10px; }

/* ... List & Options (增强样式) ... */
.question-list-container { flex: 1; overflow-y: auto; padding: 20px; }
.q-card { display: flex; background: #fff; border-radius: 8px; box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05); margin-bottom: 20px; border: 1px solid #ebeef5; overflow: hidden; transition: all 0.2s; }
.q-check { width: 50px; background: #fafafa; display: flex; justify-content: center; padding-top: 25px; border-right: 1px solid #ebeef5; }
.q-body { flex: 1; padding: 20px 30px; }
.q-text { font-size: 18px; font-weight: 500; margin-bottom: 20px; }
.q-options-area { display: grid; grid-template-columns: 1fr 1fr; gap: 15px; margin-bottom: 20px; }
.option-item { border: 1px solid #dcdfe6; border-radius: 6px; padding: 12px 20px; cursor: pointer; display: flex; align-items: center; background: #fff; position: relative; transition: all 0.2s; }
.opt-char { font-weight: bold; margin-right: 12px; color: #909399; }
.status-icon { font-size: 18px; margin-left: 10px; }
.status-icon.correct { color: #67c23a; }
.status-icon.wrong { color: #f56c6c; }
.shortcut-hint { position: absolute; right: 10px; font-size: 12px; color: #c0c4cc; font-weight: bold; background: #f5f7fa; padding: 2px 6px; border-radius: 4px; }

/* ★★★ 核心样式区分 ★★★ */
/* 1. 练习模式/未选状态 */
.option-item.is-pending:hover { background-color: #ecf5ff; border-color: #409eff; }

/* 2. 考试模式：选中但未交卷 (强高亮) */
.option-item.is-selected-exam {
  background-color: #ecf5ff;
  border: 2px solid #409eff; /* 加粗边框 */
  color: #409eff;
  font-weight: bold;
}

/* 3. 结果模式：正确答案 (绿色) */
.option-item.is-correct-opt {
  background-color: #f0f9eb;
  border-color: #67c23a;
  color: #67c23a;
  font-weight: bold;
}

/* 4. 结果模式：我选错的 (红色) */
.option-item.is-wrong-opt {
  background-color: #fef0f0;
  border-color: #f56c6c;
  color: #f56c6c;
}

/* 5. 结果模式：未选的其他项 (变灰) */
.option-item.is-disabled {
  opacity: 0.5;
  cursor: default;
  background: #f9fafe;
}

/* Analysis */
.q-analysis-box { background: #fffbf0; border: 1px dashed #e6a23c; border-radius: 6px; padding: 15px; margin-top: 10px; }
.standard-view { margin-bottom: 15px; background: #fff; padding: 10px; border-radius: 4px; border: 1px solid #e4e7ed; }
.standard-title { font-weight: bold; font-size: 14px; color: #303133; margin-bottom: 8px; display: flex; align-items: center; gap: 5px; }
.standard-options-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.std-option { font-size: 13px; color: #606266; padding: 4px 8px; border-radius: 4px; display: flex; align-items: center; }
.std-option.is-std-correct { color: #67c23a; font-weight: bold; background: #f0f9eb; }
.analysis-row { display: flex; margin-bottom: 10px; }
.tag { background: #e6a23c; color: #fff; font-size: 12px; padding: 2px 6px; border-radius: 4px; height: fit-content; margin-right: 10px; white-space: nowrap; }
.tag.note { background: #409eff; }
.text { font-size: 14px; color: #606266; line-height: 1.6; flex: 1; white-space: pre-wrap; word-break: break-all; }
.note-row { align-items: flex-start; }
.note-display { flex: 1; display: flex; justify-content: space-between; align-items: flex-start; }
.note-editor { flex: 1; display: flex; flex-direction: column; gap: 8px; }
.note-actions { display: flex; justify-content: flex-end; gap: 8px; }

/* ★★★ 修复：快捷键配置样式 ★★★ */
.key-config-tip { text-align: center; color: #909399; margin-bottom: 15px; font-size: 13px; }
.key-config-list { display: flex; flex-direction: column; gap: 10px; }
.key-item { display: flex; align-items: center; justify-content: space-between; }
.key-label { font-weight: bold; color: #606266; }
.key-input-wrapper {
  width: 200px;
  height: 32px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  display: flex;
  align-items: center;
  padding: 0 10px;
  background-color: #fff;
  transition: border-color 0.2s;
}
.key-input-wrapper:focus-within { border-color: #409eff; }
.key-input-inner {
  border: none;
  outline: none;
  width: 100%;
  text-align: center;
  color: #606266;
  font-size: 14px;
  text-transform: uppercase;
  cursor: pointer;
  background: transparent;
}
</style>