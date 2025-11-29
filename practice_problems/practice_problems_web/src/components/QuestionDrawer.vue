<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="handleVisibleChange"
    width="1600px"
    top="7vh"
    destroy-on-close
    @open="handleOpen"
    class="question-dialog"
    :show-close="false"
  >
    <!-- 1. Header -->
    <template #header="{ close, titleId, titleClass }">
      <div class="custom-header">
        <span :id="titleId" :class="titleClass" class="dialog-title">
          <el-icon class="mr-2"><Reading /></el-icon>
          {{ title }} - 专项练习
        </span>
        <div class="header-actions">
          <div v-if="isExamMode" class="exam-toolbar">
            <span class="exam-progress"
              >进度: {{ answeredCount }} / {{ questionList.length }}</span
            >
            <el-button
              v-if="!isExamSubmitted"
              type="success"
              @click="handleSubmitExam"
              :disabled="answeredCount === 0"
              >交卷查看结果</el-button
            >
            <el-tag v-else type="success" effect="dark">考试结束</el-tag>
          </div>
          <div v-if="isBatchMode" class="batch-toolbar">
            <el-checkbox
              v-model="isAllSelected"
              :indeterminate="isIndeterminate"
              @change="handleSelectAll"
              class="select-all-checkbox"
              >全选</el-checkbox
            >
            <el-button
              type="warning"
              size="small"
              :icon="Edit"
              :disabled="selectedIds.length !== 1"
              @click="handleEditQuestion"
              >修改选中</el-button
            >
            <el-button
              type="danger"
              size="small"
              :icon="Delete"
              :disabled="selectedIds.length === 0"
              @click="handleBatchDelete"
              >批量删除</el-button
            >
          </div>
          <el-divider direction="vertical" />
          <div class="switch-group">
            <div class="shortcut-control">
              <el-switch
                v-model="enableShortcuts"
                active-text="快捷答题"
                inactive-text="关闭"
                inline-prompt
                style="
                  --el-switch-on-color: #13ce66;
                  --el-switch-off-color: #909399;
                "
              />
              <el-button
                v-if="enableShortcuts"
                circle
                size="small"
                icon="Setting"
                @click="showKeyConfig = true"
                title="配置快捷键"
                class="ml-5"
              />
            </div>
            <el-switch
              v-model="isExamMode"
              active-text="考试模式"
              inactive-text="练习模式"
              inline-prompt
              style="
                --el-switch-on-color: #f56c6c;
                --el-switch-off-color: #409eff;
              "
            />
            <el-switch
              v-model="isBatchMode"
              :disabled="isExamMode"
              active-text="批量管理"
              inactive-text="普通视图"
              inline-prompt
              style="
                --el-switch-on-color: #e6a23c;
                --el-switch-off-color: #909399;
              "
            />
            <el-button
              :type="showAddForm ? 'info' : 'primary'"
              :icon="showAddForm ? Close : Plus"
              @click="toggleAddForm"
              :disabled="isExamMode"
              >{{ showAddForm ? "取消录入" : "录入新题" }}</el-button
            >
          </div>
          <el-button circle :icon="Close" @click="close" class="close-btn" />
        </div>
      </div>
    </template>

    <div class="drawer-content">
      <!-- 2. Form -->
      <el-collapse-transition>
        <div v-show="showAddForm" class="form-wrapper">
          <div class="form-box" :class="{ 'is-edit': isEditMode }">
            <div class="form-title-bar">
              <div class="form-title-text">
                {{ isEditMode ? "修改题目内容" : "录入新题" }}
                <span v-if="isEditMode" class="edit-tip"
                  >(ID: {{ editingId }})</span
                >
              </div>
              <el-button
                v-if="!isEditMode"
                size="small"
                type="warning"
                plain
                @click="showJsonImport = !showJsonImport"
                >{{ showJsonImport ? "关闭导入" : "JSON 导入" }}</el-button
              >
            </div>

            <el-collapse-transition>
              <div v-if="showJsonImport" class="json-import-area">
                <div class="json-header-tool">
                  <div class="json-tip">
                    支持 <b>单个对象</b> 或 <b>数组对象</b> (批量)。请粘贴
                    JSON：
                  </div>
                  <el-button
                    size="small"
                    type="primary"
                    link
                    icon="CopyDocument"
                    @click="copyJsonTemplate"
                    >复制标准模板</el-button
                  >
                </div>

                <el-input
                  v-model="jsonContent"
                  type="textarea"
                  :rows="6"
                  placeholder="请在此处粘贴 JSON 内容..."
                />

                <div class="json-actions">
                  <el-button size="small" @click="jsonContent = ''"
                    >清空内容</el-button
                  >
                  <el-button
                    size="small"
                    type="primary"
                    :loading="isImporting"
                    v-reclick="handleParseJson"
                  >
                    {{ isImporting ? "正在导入..." : "开始识别并导入" }}
                  </el-button>
                </div>
              </div>
            </el-collapse-transition>

            <div class="form-grid">
              <div class="form-left">
                <div class="input-group">
                  <div class="label">题目描述</div>
                  <el-input
                    v-model="form.questionText"
                    type="textarea"
                    :rows="5"
                  />
                </div>
                <div class="input-group">
                  <div class="label">答案解析</div>
                  <el-input
                    v-model="form.explanation"
                    type="textarea"
                    :rows="5"
                  />
                </div>
              </div>
              <div class="form-right">
                <div class="label">选项设置</div>
                <div class="option-inputs">
                  <el-input v-model="form.option1"
                    ><template #prepend>A</template></el-input
                  >
                  <el-input v-model="form.option2"
                    ><template #prepend>B</template></el-input
                  >
                  <el-input v-model="form.option3"
                    ><template #prepend>C</template></el-input
                  >
                  <el-input v-model="form.option4"
                    ><template #prepend>D</template></el-input
                  >
                </div>
                <div class="correct-select">
                  <span class="label-inline">正确答案：</span>
                  <el-radio-group v-model="form.correctAnswer">
                    <el-radio-button :label="1">A</el-radio-button
                    ><el-radio-button :label="2">B</el-radio-button>
                    <el-radio-button :label="3">C</el-radio-button
                    ><el-radio-button :label="4">D</el-radio-button>
                  </el-radio-group>
                  <!-- 增加一个必填提示，如果未选中时显示（可选） -->
                  <span v-if="!form.correctAnswer" style="color: #f56c6c; font-size: 12px; margin-left: 10px;">* 必选</span>
                </div>
                <div class="form-btns">
                  <el-button @click="closeAddForm">取消</el-button>
                  <el-button type="primary" v-reclick="handleSubmit">保存</el-button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-collapse-transition>

      <!-- 3. List -->
      <div class="question-list-container">
        <div class="scroll-area">
          <el-empty v-if="questionList.length === 0" description="暂无题目" />
          <div
            v-for="(q, index) in questionList"
            :key="q.id"
            class="q-card"
            :class="{
              'is-answered': q.userResult.hasAnswered,
              'is-shortcut-active':
                enableShortcuts && hoveredQuestionId === q.id,
            }"
            @mouseenter="hoveredQuestionId = q.id"
            @mouseleave="hoveredQuestionId = null"
          >
            <div v-if="isBatchMode" class="q-check">
              <el-checkbox v-model="q.isChecked" size="large" />
            </div>
            <div class="q-body">
              <div class="q-text">
                <span class="index-badge">{{ index + 1 }}</span
                ><span class="text-content">{{ q.questionText }}</span>
              </div>
              <div class="q-options-area">
                <div
                  v-for="(opt, optIdx) in q.shuffledOptions"
                  :key="optIdx"
                  class="option-item"
                  :class="getOptionClass(q, opt)"
                  @click="handleAnswer(q, opt)"
                >
                  <span class="opt-char">{{ getChar(optIdx) }}.</span
                  ><span class="opt-content">{{ opt.text }}</span>
                  <span
                    v-if="enableShortcuts && hoveredQuestionId === q.id"
                    class="shortcut-hint"
                    >[{{ getKeyDisplay(optIdx) }}]</span
                  >
                  <el-icon
                    v-if="shouldShowIcon(q, opt, 'correct')"
                    class="status-icon correct"
                    ><Select
                  /></el-icon>
                  <el-icon
                    v-if="shouldShowIcon(q, opt, 'wrong')"
                    class="status-icon wrong"
                    ><CloseBold
                  /></el-icon>
                </div>
              </div>
              <el-collapse-transition>
                <div v-if="shouldShowAnalysis(q)" class="q-analysis-box">
                  <div class="standard-view">
                    <div class="standard-title">
                      <el-icon><List /></el-icon> 标准视图 (对应下方解析文本)
                    </div>
                    <div class="standard-options-grid">
                      <div
                        v-for="(stdOpt, stdIdx) in q.originalOptions"
                        :key="stdIdx"
                        class="std-option"
                        :class="{
                          'is-std-correct': stdIdx + 1 === q.correctAnswer,
                        }"
                      >
                        <span class="std-char">{{ getChar(stdIdx) }}.</span
                        ><span class="std-content">{{ stdOpt.text }}</span>
                        <el-icon
                          v-if="stdIdx + 1 === q.correctAnswer"
                          class="std-icon"
                          ><Select
                        /></el-icon>
                      </div>
                    </div>
                  </div>
                  <div class="analysis-row">
                    <span class="tag">解析</span>
                    <div class="text">{{ q.explanation || "暂无解析" }}</div>
                  </div>
                  <div class="analysis-row note-row">
                    <span class="tag note">我的笔记</span>
                    <div v-if="editingNoteId !== q.id" class="note-display">
                      <div class="text note-text">
                        {{ q.note || "暂无笔记..." }}
                      </div>
                      <el-button
                        link
                        type="primary"
                        :icon="Edit"
                        @click="startEditNote(q)"
                        >编辑</el-button
                      >
                    </div>
                    <div v-else class="note-editor">
                      <el-input
                        v-model="tempNoteContent"
                        type="textarea"
                        :rows="2"
                      />
                      <div class="note-actions">
                        <el-button size="small" @click="cancelEditNote"
                          >取消</el-button
                        ><el-button
                          size="small"
                          type="primary"
                          @click="saveNote(q)"
                          >保存</el-button
                        >
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

    <!-- Key Config -->
    <el-dialog
      v-model="showKeyConfig"
      title="配置答题快捷键"
      width="400px"
      append-to-body
      class="key-config-dialog"
    >
      <div class="key-config-tip">点击输入框并按下键盘上的任意键进行绑定</div>
      <div class="key-config-list">
        <div
          class="key-item"
          v-for="(label, idx) in ['A', 'B', 'C', 'D']"
          :key="idx"
        >
          <span class="key-label">选项 {{ label }}</span>
          <el-input
            readonly
            :model-value="keyBindings[idx].toUpperCase()"
            @keydown.prevent="(e: any) => handleBindKey(e, idx)"
            placeholder="按下按键绑定"
            class="key-input"
          >
            <template #append>键</template>
          </el-input>
        </div>
      </div>
      <template #footer
        ><el-button @click="resetKeys">恢复默认</el-button
        >
        <el-button type="primary" v-reclick="saveKeys">确定保存</el-button>
      </template>
    </el-dialog>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onUnmounted } from "vue";
import {
  Plus,
  Close,
  Edit,
  Delete,
  Select,
  CloseBold,
  Reading,
  List,
  Setting,
  CopyDocument,
} from "@element-plus/icons-vue";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  getQuestions,
  createQuestion,
  updateQuestion,
  deleteQuestion,
  type QuestionItem,
} from "../api/question";

const props = defineProps<{
  visible: boolean;
  pointId: number;
  title: string;
}>();
const emit = defineEmits(["update:visible"]);
const handleVisibleChange = (val: boolean) => emit("update:visible", val);

interface FrontendQuestion extends QuestionItem {
  isChecked: boolean;
  shuffledOptions: Array<{ text: string; originalIndex: number }>;
  originalOptions: Array<{ text: string; originalIndex: number }>;
  userResult: {
    hasAnswered: boolean;
    selectedOriginalIndex: number | null;
    isCorrect: boolean;
  };
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

// ★★★ 修改点1：correctAnswer 默认值改为 undefined，不默认选A ★★★
const form = reactive({
  questionText: "",
  option1: "",
  option2: "",
  option3: "",
  option4: "",
  correctAnswer: undefined as number | undefined,
  explanation: "",
});

// --- JSON 导入逻辑 ---
const showJsonImport = ref(false);
const jsonContent = ref("");
const isImporting = ref(false);

// 复制模板功能
const copyJsonTemplate = () => {
  const template = `[
  {
    "topic": "这里填写题目内容",
    "answer1": "选项A内容",
    "answer2": "选项B内容",
    "answer3": "选项C内容",
    "answer4": "选项D内容",
    "proper": "1",
    "resolve": "这里填写解析"
  }
]`;
  navigator.clipboard
    .writeText(template)
    .then(() => {
      ElMessage.success("标准模板已复制到剪贴板");
    })
    .catch(() => {
      ElMessage.error("复制失败，请手动复制");
    });
};

const handleParseJson = async () => {
  if (!jsonContent.value) return ElMessage.warning("请先粘贴 JSON 内容");
  try {
    const data = JSON.parse(jsonContent.value);
    if (Array.isArray(data)) {
      isImporting.value = true;
      let successCount = 0;
      for (const item of data) {
        // JSON 导入时，如果没有指定正确答案，也暂时默认为1，或者可以在这里也做校验
        // 为了保证导入成功率，这里保留默认值1，但你可以根据需求改为 undefined
        let importCorrectAnswer = 1; 
        if (item.proper) {
          const n = Number(item.proper);
          if (!isNaN(n) && n >= 1 && n <= 4) importCorrectAnswer = n;
        }
        
        const payload = {
          knowledgePointId: props.pointId,
          questionText: item.topic || "",
          option1: item.answer1 || "",
          option2: item.answer2 || "",
          option3: item.answer3 || "",
          option4: item.answer4 || "",
          correctAnswer: importCorrectAnswer,
          explanation: item.resolve || "",
          note: "",
        };
        
        if (payload.questionText && payload.option1 && payload.option2) {
          try {
            await createQuestion(payload);
            successCount++;
          } catch (err) {}
        }
      }
      isImporting.value = false;
      ElMessage.success(`批量导入完成，成功 ${successCount} 条`);
      jsonContent.value = "";
      showJsonImport.value = false;
      await loadQuestions();
    } else {
      if (data.topic) form.questionText = data.topic;
      if (data.answer1) form.option1 = data.answer1;
      if (data.answer2) form.option2 = data.answer2;
      if (data.answer3) form.option3 = data.answer3;
      if (data.answer4) form.option4 = data.answer4;
      if (data.resolve) form.explanation = data.resolve;
      if (data.proper) {
        const n = Number(data.proper);
        if (!isNaN(n) && n >= 1 && n <= 4) form.correctAnswer = n;
      }
      ElMessage.success("解析成功，表单已填充");
      showJsonImport.value = false;
    }
  } catch (e) {
    isImporting.value = false;
    ElMessage.error("JSON 格式错误，请检查");
  }
};

// --- Shortcuts ---
const enableShortcuts = ref(false);
const hoveredQuestionId = ref<number | null>(null);
const showKeyConfig = ref(false);
const defaultKeys = ["a", "b", "c", "d"];
const keyBindings = reactive<string[]>([...defaultKeys]);

onMounted(() => {
  const saved = localStorage.getItem("question_shortcuts");
  if (saved) {
    try {
      const parsed = JSON.parse(saved);
      if (Array.isArray(parsed)) parsed.forEach((k, i) => (keyBindings[i] = k));
    } catch (e) {}
  }
  window.addEventListener("keydown", handleKeydown);
});
onUnmounted(() => {
  window.removeEventListener("keydown", handleKeydown);
});

const handleBindKey = (e: KeyboardEvent, index: number) => {
  const key = e.key.toLowerCase();
  if (
    key.length > 1 &&
    ![
      "enter",
      "space",
      "arrowup",
      "arrowdown",
      "arrowleft",
      "arrowright",
    ].includes(key)
  )
    return;
  keyBindings[index] = key;
  (e.target as HTMLElement).blur();
};
const resetKeys = () => {
  defaultKeys.forEach((k, i) => (keyBindings[i] = k));
};
const saveKeys = () => {
  localStorage.setItem("question_shortcuts", JSON.stringify(keyBindings));
  showKeyConfig.value = false;
};
const getKeyDisplay = (idx: number) => keyBindings[idx].toUpperCase();
const handleKeydown = (e: KeyboardEvent) => {
  if (!enableShortcuts.value) return;
  const activeTag = document.activeElement?.tagName;
  if (activeTag === "INPUT" || activeTag === "TEXTAREA") return;
  if (!hoveredQuestionId.value) return;
  const q = questionList.value.find(
    (item) => item.id === hoveredQuestionId.value
  );
  if (!q) return;
  const index = keyBindings.findIndex((k) => k === e.key.toLowerCase());
  if (index !== -1 && q.shuffledOptions[index])
    handleAnswer(q, q.shuffledOptions[index]);
};

// --- Watchers & Init ---
const selectedIds = computed(() =>
  questionList.value.filter((q) => q.isChecked).map((q) => q.id)
);
const isAllSelected = computed(
  () =>
    questionList.value.length > 0 &&
    questionList.value.every((q) => q.isChecked)
);
const isIndeterminate = computed(
  () =>
    selectedIds.value.length > 0 &&
    selectedIds.value.length < questionList.value.length
);
const answeredCount = computed(
  () => questionList.value.filter((q) => q.userResult.hasAnswered).length
);

watch(isExamMode, (newVal) => {
  if (newVal) {
    isBatchMode.value = false;
    closeAddForm();
    resetAllAnswers();
    isExamSubmitted.value = false;
    ElMessage.info("已进入考试模式");
  } else {
    isExamSubmitted.value = false;
  }
});
watch(isBatchMode, (newVal) => {
  if (newVal && showAddForm.value) closeAddForm();
});
watch(showAddForm, (newVal) => {
  if (newVal && isBatchMode.value) isBatchMode.value = false;
});

const handleOpen = () => {
  isBatchMode.value = false;
  isExamMode.value = false;
  isExamSubmitted.value = false;
  closeAddForm();
  loadQuestions();
};
const closeAddForm = () => {
  showAddForm.value = false;
  showJsonImport.value = false;
  resetForm();
};

// ★★★ 修改点2：重置表单时，correctAnswer 设为 undefined ★★★
const resetForm = () => {
  isEditMode.value = false;
  editingId.value = null;
  form.questionText = "";
  form.option1 = "";
  form.option2 = "";
  form.option3 = "";
  form.option4 = "";
  form.correctAnswer = undefined; 
  form.explanation = "";
};
const resetAllAnswers = () => {
  questionList.value.forEach((q) => {
    q.userResult.hasAnswered = false;
    q.userResult.selectedOriginalIndex = null;
    q.userResult.isCorrect = false;
  });
};

const loadQuestions = async () => {
  if (!props.pointId) return;
  try {
    const res = await getQuestions(props.pointId);
    if (res.data && (res.data as any).code === 200) {
      questionList.value = res.data.data.map((item: QuestionItem) => {
        const rawOptions = [
          { text: item.option1, originalIndex: 1 },
          { text: item.option2, originalIndex: 2 },
          { text: item.option3, originalIndex: 3 },
          { text: item.option4, originalIndex: 4 },
        ];
        const shuffled = [...rawOptions];
        for (let i = shuffled.length - 1; i > 0; i--) {
          const j = Math.floor(Math.random() * (i + 1));
          [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
        }
        return {
          ...item,
          isChecked: false,
          shuffledOptions: shuffled,
          originalOptions: rawOptions,
          userResult: {
            hasAnswered: false,
            selectedOriginalIndex: null,
            isCorrect: false,
          },
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
  q.userResult.isCorrect = opt.originalIndex === q.correctAnswer;
};

const handleSubmitExam = () => {
  ElMessageBox.confirm(`确认交卷？`, "交卷", {
    confirmButtonText: "确定",
    cancelButtonText: "取消",
    type: "warning",
  }).then(() => {
    isExamSubmitted.value = true;
    const correctCount = questionList.value.filter(
      (q) => q.userResult.isCorrect
    ).length;
    ElMessage.success(`得分：${correctCount} / ${questionList.value.length}`);
  });
};

// --- Render Logic ---
const getOptionClass = (q: FrontendQuestion, opt: any) => {
  if (!q.userResult.hasAnswered) return "is-pending";
  if (isExamMode.value && !isExamSubmitted.value) {
    return q.userResult.selectedOriginalIndex === opt.originalIndex
      ? "is-selected-exam"
      : "is-disabled";
  }
  if (opt.originalIndex === q.correctAnswer) return "is-correct-opt";
  if (
    q.userResult.selectedOriginalIndex === opt.originalIndex &&
    !q.userResult.isCorrect
  )
    return "is-wrong-opt";
  return "is-disabled";
};
const shouldShowIcon = (q: FrontendQuestion, opt: any, type: string) => {
  if (!q.userResult.hasAnswered) return false;
  if (isExamMode.value && !isExamSubmitted.value) return false;
  if (type === "correct") return opt.originalIndex === q.correctAnswer;
  if (type === "wrong")
    return (
      q.userResult.selectedOriginalIndex === opt.originalIndex &&
      opt.originalIndex !== q.correctAnswer
    );
  return false;
};
const shouldShowAnalysis = (q: FrontendQuestion) => {
  if (!q.userResult.hasAnswered) return false;
  return isExamMode.value ? isExamSubmitted.value : true;
};

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
    const res = await updateQuestion(q.id, {
      ...q,
      note: tempNoteContent.value,
      knowledgePointId: props.pointId,
    });
    if ((res.data as any).code === 200) {
      ElMessage.success("笔记保存成功");
      q.note = tempNoteContent.value;
      editingNoteId.value = null;
    }
  } catch (e) {
    ElMessage.error("保存失败");
  }
};
const toggleAddForm = () => {
  showAddForm.value ? closeAddForm() : (showAddForm.value = true);
};
const handleEditQuestion = () => {
  if (selectedIds.value.length !== 1) return;
  const target = questionList.value.find((q) => q.id === selectedIds.value[0]);
  if (target) {
    form.questionText = target.questionText;
    form.option1 = target.option1;
    form.option2 = target.option2;
    form.option3 = target.option3;
    form.option4 = target.option4;
    form.correctAnswer = target.correctAnswer;
    form.explanation = target.explanation;
    isEditMode.value = true;
    editingId.value = target.id;
    showAddForm.value = true;
    isBatchMode.value = false;
  }
};

// ★★★ 修改点3：提交时增加必填校验 ★★★
const handleSubmit = async () => {
  if (!form.questionText) return ElMessage.warning("请完善题目描述");
  
  // 核心修改：如果没有选择正确答案，提示错误并拦截
  if (!form.correctAnswer) {
    return ElMessage.warning("请选择正确答案！");
  }

  const payload = { 
    knowledgePointId: props.pointId, 
    ...form, 
    correctAnswer: form.correctAnswer, // 确保是 number 类型
    note: "" 
  };
  
  if (isEditMode.value && editingId.value) {
    const t = questionList.value.find((q) => q.id === editingId.value);
    if (t) payload.note = t.note;
  }
  try {
    if (isEditMode.value && editingId.value) {
      await updateQuestion(editingId.value, payload);
      ElMessage.success("修改成功");
    } else {
      await createQuestion(payload);
      ElMessage.success("录入成功");
    }
    closeAddForm();
    loadQuestions();
  } catch (e) {
    ElMessage.error("操作失败");
  }
};
const handleBatchDelete = () => {
  ElMessageBox.confirm(`确认删除?`, "警告", { type: "warning" }).then(
    async () => {
      try {
        await Promise.all(selectedIds.value.map((id) => deleteQuestion(id)));
        ElMessage.success("删除成功");
        loadQuestions();
      } catch (e) {
        ElMessage.error("删除失败");
      }
    }
  );
};
const handleSelectAll = (val: boolean) => {
  questionList.value.forEach((q) => (q.isChecked = val));
};
const getChar = (i: number) => String.fromCharCode(65 + i);
</script>

<style scoped>
/* --- Layout --- */
.drawer-content {
  height: 75vh;
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
}
.custom-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-right: 20px;
}
.dialog-title {
  font-size: 20px;
  font-weight: bold;
  display: flex;
  align-items: center;
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 15px;
}
.exam-toolbar {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-right: 10px;
}
.exam-progress {
  font-size: 14px;
  font-weight: bold;
  color: #909399;
}
.batch-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #fff;
  padding: 4px 12px;
  border-radius: 4px;
  border: 1px solid #dcdfe6;
}
.select-all-checkbox {
  margin-right: 10px;
}
.switch-group {
  display: flex;
  align-items: center;
  gap: 15px;
}
.shortcut-control {
  display: flex;
  align-items: center;
  gap: 5px;
  margin-right: 10px;
}
.ml-5 {
  margin-left: 5px;
}

/* --- Form --- */
.form-wrapper {
  padding: 15px 20px 0 20px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
}
.form-box {
  background: #f0f9eb;
  border: 1px solid #e1f3d8;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 15px;
}
.form-box.is-edit {
  background: #fdf6ec;
  border-color: #faecd8;
}

/* Form Title Bar & JSON Import */
.form-title-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  border-bottom: 1px dashed #ccc;
  padding-bottom: 10px;
}
.form-title-text {
  font-weight: bold;
  font-size: 16px;
}
.json-import-area {
  background: #fff;
  padding: 15px;
  border: 1px dashed #e6a23c;
  border-radius: 6px;
  margin-bottom: 20px;
}
.json-header-tool {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.json-tip {
  font-size: 12px;
  color: #909399;
}
.json-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
  gap: 10px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 30px;
}
.input-group {
  margin-bottom: 12px;
}
.label {
  font-weight: bold;
  margin-bottom: 5px;
  color: #606266;
}
.option-inputs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 15px;
}
.correct-select {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
}
.label-inline {
  font-weight: bold;
  margin-right: 10px;
  color: #67c23a;
}
.form-btns {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

/* --- List --- */
.question-list-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}
.q-card {
  display: flex;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
  margin-bottom: 20px;
  border: 1px solid #ebeef5;
  overflow: hidden;
  transition: all 0.2s;
}
.q-check {
  width: 50px;
  background: #fafafa;
  display: flex;
  justify-content: center;
  padding-top: 25px;
  border-right: 1px solid #ebeef5;
}
.q-body {
  flex: 1;
  padding: 20px 30px;
}
.q-text {
  font-size: 18px;
  font-weight: 500;
  margin-bottom: 20px;
}
.text-content {
  white-space: pre-wrap;
  line-height: 1.6;
}
.index-badge {
  background: #409eff;
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  margin-right: 8px;
  font-size: 14px;
}
.q-card.is-shortcut-active {
  border-color: #13ce66;
  box-shadow: 0 0 8px rgba(19, 206, 102, 0.2);
}

/* Options */
.q-options-area {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 15px;
  margin-bottom: 20px;
}
.option-item {
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  padding: 12px 20px;
  cursor: pointer;
  display: flex;
  align-items: center;
  background: #fff;
  position: relative;
}
.opt-char {
  font-weight: bold;
  margin-right: 12px;
  color: #909399;
}
.status-icon {
  font-size: 18px;
  margin-left: 10px;
}
.status-icon.correct {
  color: #67c23a;
}
.status-icon.wrong {
  color: #f56c6c;
}
.option-item.is-pending:hover {
  background-color: #ecf5ff;
  border-color: #409eff;
  color: #409eff;
}
.shortcut-hint {
  position: absolute;
  right: 10px;
  font-size: 12px;
  color: #c0c4cc;
  font-weight: bold;
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 4px;
}

/* States */
.option-item.is-correct-opt {
  background-color: #f0f9eb;
  border-color: #67c23a;
  color: #67c23a;
  font-weight: bold;
}
.option-item.is-wrong-opt {
  background-color: #fef0f0;
  border-color: #f56c6c;
  color: #f56c6c;
}
.option-item.is-disabled {
  opacity: 0.6;
  cursor: not-allowed;
  background: #f5f7fa;
}
.option-item.is-selected-exam {
  background-color: #ecf5ff;
  border-color: #409eff;
  color: #409eff;
  font-weight: bold;
}

/* Analysis */
.q-analysis-box {
  background: #fffbf0;
  border: 1px dashed #e6a23c;
  border-radius: 6px;
  padding: 15px;
  margin-top: 10px;
}
.standard-view {
  margin-bottom: 15px;
  background: #fff;
  padding: 10px;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
}
.standard-title {
  font-weight: bold;
  font-size: 14px;
  color: #303133;
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 5px;
}
.standard-options-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}
.std-option {
  font-size: 13px;
  color: #606266;
  padding: 4px 8px;
  border-radius: 4px;
  display: flex;
  align-items: center;
}
.std-option.is-std-correct {
  color: #67c23a;
  font-weight: bold;
  background: #f0f9eb;
}
.std-char {
  margin-right: 5px;
  font-weight: bold;
}
.std-icon {
  margin-left: 5px;
}

.analysis-row {
  display: flex;
  margin-bottom: 10px;
}
.tag {
  background: #e6a23c;
  color: #fff;
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 4px;
  height: fit-content;
  margin-right: 10px;
  white-space: nowrap;
}
.tag.note {
  background: #409eff;
}
.text {
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
  flex: 1;
  white-space: pre-wrap;
  word-break: break-all;
}
.note-row {
  align-items: flex-start;
}
.note-display {
  flex: 1;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}
.note-editor {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.note-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

/* Key Config Dialog */
.key-config-tip {
  text-align: center;
  color: #909399;
  margin-bottom: 15px;
  font-size: 13px;
}
.key-config-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.key-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.key-label {
  font-weight: bold;
  color: #606266;
}
.key-input {
  width: 200px;
}
</style>