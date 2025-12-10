<template>
  <div 
    class="wordbook-float-icon" 
    :class="{ expanded: isExpanded }"
    :style="{ top: position.top + 'px', left: position.left + 'px', display: props.visible ? 'block' : 'none' }"
  >
    <div class="wordbook-icon" v-if="!isExpanded" @mousedown="handleMouseDown">
      <el-icon :size="24"><Reading /></el-icon>
    </div>
    
    <div class="wordbook-window" v-else @click.stop>
      <div class="wordbook-header" @mousedown="handleHeaderMouseDown">
        <span>单词本</span>
        <div class="header-actions">
          <el-button link @click="addWord" title="添加单词" class="io-button">
            ADD
          </el-button>
          <el-button link @click="exportWords" title="导出单词" class="io-button">
            OUT
          </el-button>
          <el-button link @click="importWords" title="导入单词" class="io-button">
            IN
          </el-button>
          <el-button link @click="minimize">
            <el-icon><Minus /></el-icon>
          </el-button>
          <el-button link @click="closeWordBook">
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
      </div>
      
      <div class="wordbook-content">
        <!-- 分类标签页 -->
        <el-tabs v-model="currentCategory" class="category-tabs">
          <el-tab-pane label="所有" name="all">
            <template #label>
              <span>所有 ({{ words.length }})</span>
            </template>
          </el-tab-pane>
          <el-tab-pane label="生词" name="new">
            <template #label>
              <span>生词 ({{ categoryCount.new }})</span>
            </template>
          </el-tab-pane>
          <el-tab-pane label="巩固" name="review">
            <template #label>
              <span>巩固 ({{ categoryCount.review }})</span>
            </template>
          </el-tab-pane>
          <el-tab-pane label="熟词" name="mastered">
            <template #label>
              <span>熟词 ({{ categoryCount.mastered }})</span>
            </template>
          </el-tab-pane>
        </el-tabs>
        
        <div class="search-bar">
          <el-input 
            v-model="searchKeyword" 
            placeholder="搜索单词..." 
            size="small"
            clearable
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
        
        <div class="word-list">
          <div 
            v-for="word in filteredWords" 
            :key="word.id"
            class="word-item"
          >
            <div class="word-main">
              <strong @dblclick="toggleSyllables(word)" :class="{ 'syllable-mode': word.showSyllables }">{{ getDisplayText(word) }}</strong>
              <span class="word-phonetic">{{ word.phonetic }}</span>
              <el-button link size="small" @click.stop="editWord(word)" title="编辑单词">
                <el-icon><Edit /></el-icon>
              </el-button>
              <el-dropdown 
                trigger="click" 
                @command="handleMoveCommand(word, $event)" 
                :teleported="false"
              >
                <span class="el-dropdown-link" @click.stop>
                  <el-button link size="small" title="移动单词">
                    <el-icon><Switch /></el-icon>
                  </el-button>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item v-if="word.category !== 'new'" command="new">
                      移至生词本
                    </el-dropdown-item>
                    <el-dropdown-item v-if="word.category !== 'review'" command="review">
                      移至巩固本
                    </el-dropdown-item>
                    <el-dropdown-item v-if="word.category !== 'mastered'" command="mastered">
                      移至熟词本
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <span class="category-tag" :class="'tag-' + word.category">{{ getCategoryTag(word.category) }}</span>
              <el-button link size="small" type="danger" @click.stop="deleteWord(word.id)" title="删除单词">
                <el-icon><Delete /></el-icon>
              </el-button>
              <el-button link size="small" @click="speakWord(word)">
                <el-icon><Promotion /></el-icon>
              </el-button>
            </div>
            <div class="word-meaning">{{ word.meaning }}</div>
          </div>
          
          <div v-if="filteredWords.length === 0" class="empty-state">
            <el-icon><Document /></el-icon>
            <p>暂无单词</p>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 添加单词对话框 -->
    <el-dialog v-model="addWordDialogVisible" title="添加单词" width="400px" @close="resetNewWord">
      <el-form :model="newWord" label-width="80px">
        <el-form-item label="单词">
          <div style="display: flex; gap: 10px;">
            <el-input 
              v-model="newWord.text" 
              placeholder="请输入单词" 
              style="flex: 1" 
              clearable 
              @keyup.enter="lookupWord"
            />
            <el-button type="primary" @click="lookupWord">查询</el-button>
          </div>
        </el-form-item>
        <el-form-item label="音标">
          <el-input v-model="newWord.phonetic" placeholder="点击查询或手动输入" />
        </el-form-item>
        <el-form-item label="释义">
          <el-input v-model="newWord.meaning" type="textarea" placeholder="点击查询或手动输入" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="addWordDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmAddWord">确定</el-button>
        </span>
      </template>
    </el-dialog>
    
    <!-- 编辑单词对话框 -->
    <el-dialog v-model="editWordDialogVisible" title="编辑单词" width="400px" @close="cancelEdit">
      <el-form v-if="editingWord" :model="editingWord" label-width="80px">
        <el-form-item label="单词">
          <el-input v-model="editingWord.text" placeholder="请输入单词" />
        </el-form-item>
        <el-form-item label="音标">
          <el-input v-model="editingWord.phonetic" placeholder="请输入音标" />
        </el-form-item>
        <el-form-item label="释义">
          <el-input v-model="editingWord.meaning" type="textarea" placeholder="请输入释义" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="cancelEdit">取消</el-button>
          <el-button type="primary" @click="confirmEditWord">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Reading, Minus, Close, Search, Document, Plus, Promotion, Delete, Edit, Switch, Download, Upload } from '@element-plus/icons-vue';
import { extractSyllables } from '../utils/syllableExtractor';

const props = defineProps({
  visible: { type: Boolean, default: false }
});

const emit = defineEmits(['update:visible']);

const isExpanded = ref(false);
const searchKeyword = ref('');
const currentCategory = ref('all'); // 当前分类: all-所有, new-生词, review-巩固, mastered-熟词
const addWordDialogVisible = ref(false); // 添加单词对话框可见性
const editWordDialogVisible = ref(false); // 编辑单词对话框可见性
const newWord = ref({ text: '', phonetic: '', meaning: '' }); // 新单词数据
const editingWord = ref<any>(null); // 当前编辑的单词

// 位置状态
const position = reactive({
  top: 50,
  left: window.innerWidth - 100
});

// 翻译URL
const translateUrl = computed(() => {
  return 'https://translate.google.com/?sl=auto&tl=zh-CN&op=translate';
});

// 拖拽相关
let isDragging = false;
let dragOffset = { x: 0, y: 0 };
let currentX = 0;
let currentY = 0;
let animationFrameId: number | null = null;
let startX = 0;
let startY = 0;
const DRAG_THRESHOLD = 5; // 拖拽阈值

// 默认示例单词数据（仅用于参考，不自动加载）
const defaultWords = [
  { id: 1, text: 'repository', phonetic: '/rɪˈpɑːzətɔːri/', meaning: '仓库，存储库', showSyllables: false, category: 'new' },
  { id: 2, text: 'component', phonetic: '/kəmˈpoʊnənt/', meaning: '组件，部件', showSyllables: false, category: 'new' },
  { id: 3, text: 'interface', phonetic: '/ˈɪntərfeɪs/', meaning: '接口，界面', showSyllables: false, category: 'new' },
  { id: 4, text: 'differ', phonetic: '/ˈdɪfə/', meaning: '不同', showSyllables: false, category: 'review' },
  { id: 5, text: 'better', phonetic: '/ˈbetə/', meaning: '更好的', showSyllables: false, category: 'review' },
  { id: 6, text: 'letter', phonetic: '/ˈletə/', meaning: '字母，信', showSyllables: false, category: 'mastered' },
  { id: 7, text: 'happy', phonetic: '/ˈhæpi/', meaning: '快乐的', showSyllables: false, category: 'mastered' },
  { id: 8, text: 'summer', phonetic: '/ˈsʌmə/', meaning: '夏天', showSyllables: false, category: 'new' },
  { id: 9, text: 'apple', phonetic: '/ˈæpl/', meaning: '苹果', showSyllables: false, category: 'new' },
  { id: 10, text: 'understand', phonetic: '/ˌʌndəˈstænd/', meaning: '理解', showSyllables: false, category: 'new' }
];

// 单词列表（从本地存储初始化，默认为空）
const words = ref<any[]>([]);

// 从本地存储加载单词
const loadWordsFromLocalStorage = () => {
  try {
    const savedWords = localStorage.getItem('wordbook_words');
    if (savedWords) {
      // 如果有保存的单词，直接使用（已经是倒序，最新的在前面）
      words.value = JSON.parse(savedWords);
    } else {
      // 如果没有保存的单词，保持为空
      words.value = [];
    }
  } catch (error) {
    console.error('加载单词失败:', error);
    // 如果加载失败，保持为空
    words.value = [];
  }
};

// 保存单词到本地存储
const saveWordsToLocalStorage = () => {
  try {
    localStorage.setItem('wordbook_words', JSON.stringify(words.value));
  } catch (error) {
    console.error('保存单词失败:', error);
  }
};

// 过滤后的单词（根据分类和搜索关键词）
const filteredWords = computed(() => {
  let filtered = currentCategory.value === 'all' 
    ? words.value 
    : words.value.filter(word => word.category === currentCategory.value);
  
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase();
    filtered = filtered.filter(word => 
      word.text.toLowerCase().includes(keyword) || 
      word.meaning.toLowerCase().includes(keyword)
    );
  }
  
  return filtered;
});

// 统计每个分类的单词数量
const categoryCount = computed(() => {
  return {
    new: words.value.filter(w => w.category === 'new').length,
    review: words.value.filter(w => w.category === 'review').length,
    mastered: words.value.filter(w => w.category === 'mastered').length
  };
});

// 获取显示文本（切换音节显示）
const getDisplayText = (word: any) => {
  if (!word.showSyllables || !word.phonetic) {
    return word.text;
  }
  
  // 每次都实时计算音节，不使用预设
  return extractSyllables(word.text, word.phonetic);
};

// 音节提取相关函数已移至 src/utils/syllableExtractor.ts

// 切换音节显示
const toggleSyllables = (word: any) => {
  if (!word.phonetic) {
    ElMessage.info('该单词没有音标');
    return;
  }
  
  if (word.showSyllables) {
    // 关闭音节模式
    word.showSyllables = false;
  } else {
    // 开启音节模式，使用工具函数提取音节
    word.syllableText = extractSyllables(word.text, word.phonetic);
    word.showSyllables = true;
  }
};

// 移动单词到不同分类
const moveWordToCategory = (word: any, targetCategory: string) => {
  const index = words.value.findIndex(w => w.id === word.id);
  if (index !== -1) {
    words.value[index].category = targetCategory;
    saveWordsToLocalStorage();
    ElMessage.success(`已移动到${getCategoryName(targetCategory)}`);
  }
};

// 处理移动命令
const handleMoveCommand = (word: any, category: string) => {
  moveWordToCategory(word, category);
};

// 获取分类名称
const getCategoryName = (category: string) => {
  const names: any = {
    new: '生词本',
    review: '巩固本',
    mastered: '熟词本'
  };
  return names[category] || '生词本';
};

// 获取分类标签
const getCategoryTag = (category: string) => {
  const tags: any = {
    new: '生',
    review: '巩',
    mastered: '熟'
  };
  return tags[category] || '生';
};

// 处理鼠标按下事件
const handleMouseDown = (e: MouseEvent) => {
  // 记录起始位置
  startX = e.clientX;
  startY = e.clientY;
  
  // 设置为未拖拽状态
  isDragging = false;
  
  // 添加鼠标移动和释放事件监听器
  document.addEventListener('mousemove', handleMouseMove);
  document.addEventListener('mouseup', handleMouseUp);
};

// 处理标题栏鼠标按下事件
const handleHeaderMouseDown = (e: MouseEvent) => {
  // 防止点击按钮时触发拖拽
  if ((e.target as HTMLElement).closest('button')) {
    return;
  }
  
  // 记录起始位置
  startX = e.clientX;
  startY = e.clientY;
  
  // 设置为未拖拽状态
  isDragging = false;
  
  // 添加鼠标移动和释放事件监听器
  document.addEventListener('mousemove', handleMouseMove);
  document.addEventListener('mouseup', handleMouseUp);
};

// 处理鼠标移动事件
const handleMouseMove = (e: MouseEvent) => {
  // 计算移动距离
  const deltaX = Math.abs(e.clientX - startX);
  const deltaY = Math.abs(e.clientY - startY);
  
  // 如果移动距离超过阈值，则认为是拖拽操作
  if (deltaX > DRAG_THRESHOLD || deltaY > DRAG_THRESHOLD) {
    isDragging = true;
    // 直接开始拖拽，而不是调用startDrag
    dragOffset.x = e.clientX - position.left;
    dragOffset.y = e.clientY - position.top;
    
    // 添加拖拽时的高性能样式
    const element = document.querySelector('.wordbook-float-icon') as HTMLElement;
    if (element) {
      element.style.willChange = 'transform';
    }
    
    document.addEventListener('mousemove', onDrag);
    document.addEventListener('mouseup', stopDrag);
    
    // 移除临时事件监听器
    document.removeEventListener('mousemove', handleMouseMove);
    document.removeEventListener('mouseup', handleMouseUp);
  }
};

// 处理鼠标释放事件
const handleMouseUp = (e: MouseEvent) => {
  // 移除事件监听器
  document.removeEventListener('mousemove', handleMouseMove);
  document.removeEventListener('mouseup', handleMouseUp);
  
  // 如果没有拖拽，则认为是点击操作
  if (!isDragging) {
    isExpanded.value = true;
  }
  
  // 重置拖拽状态
  isDragging = false;
};

// 拖拽中
const onDrag = (e: MouseEvent) => {
  if (!isDragging) return;
  
  currentX = e.clientX - dragOffset.x;
  currentY = e.clientY - dragOffset.y;
  
  // 使用 requestAnimationFrame 优化性能
  if (animationFrameId) {
    cancelAnimationFrame(animationFrameId);
  }
  animationFrameId = requestAnimationFrame(() => {
    position.left = currentX;
    position.top = currentY;
    
    // 边界检查 - 允许拖拽到屏幕外
    // position.left = Math.max(-250, Math.min(position.left, window.innerWidth - 50));
    // position.top = Math.max(-350, Math.min(position.top, window.innerHeight - 50));
    // 移除边界限制，允许自由拖拽
    position.left = currentX;
    position.top = currentY;
    animationFrameId = null;
  });
};

// 停止拖拽
const stopDrag = () => {
  isDragging = false;
  if (animationFrameId) {
    cancelAnimationFrame(animationFrameId);
    animationFrameId = null;
  }
  
  // 移除拖拽时的高性能样式
  const element = document.querySelector('.wordbook-float-icon') as HTMLElement;
  if (element) {
    element.style.willChange = 'auto';
  }
  
  document.removeEventListener('mousemove', onDrag);
  document.removeEventListener('mouseup', stopDrag);
};

// 最小化
const minimize = () => {
  isExpanded.value = false; // 还原为图标状态
};

// 关闭单词本
const closeWordBook = () => {
  isExpanded.value = false; // 还原为图标状态
};

// 导出单词为JSON文件
const exportWords = () => {
  try {
    if (words.value.length === 0) {
      ElMessage.warning('没有单词可导出');
      return;
    }
    
    // 创建导出数据对象
    const exportData = {
      exportTime: new Date().toISOString(),
      totalWords: words.value.length,
      categories: {
        new: words.value.filter(w => w.category === 'new'),
        review: words.value.filter(w => w.category === 'review'),
        mastered: words.value.filter(w => w.category === 'mastered')
      },
      allWords: words.value
    };
    
    // 转换为JSON字符串
    const jsonStr = JSON.stringify(exportData, null, 2);
    
    // 创建Blob对象
    const blob = new Blob([jsonStr], { type: 'application/json' });
    
    // 创建下载链接
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `wordbook_${new Date().getTime()}.json`;
    
    // 触发下载
    document.body.appendChild(link);
    link.click();
    
    // 清理
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
    
    ElMessage.success(`成功导出 ${words.value.length} 个单词`);
  } catch (error) {
    console.error('导出失败:', error);
    ElMessage.error('导出失败，请重试');
  }
};

// 导入单词从JSON文件
const importWords = () => {
  try {
    // 创建文件输入元素
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.json';
    
    input.onchange = async (e: any) => {
      const file = e.target.files[0];
      if (!file) return;
      
      try {
        // 读取文件内容
        const text = await file.text();
        const importData = JSON.parse(text);
        
        // 验证数据格式
        if (!importData.allWords || !Array.isArray(importData.allWords)) {
          ElMessage.error('文件格式不正确');
          return;
        }
        
        // 直接采用合并模式（添加新单词）
        const existingTexts = words.value.map(w => w.text.toLowerCase());
        const newWords = importData.allWords.filter(
          (w: any) => !existingTexts.includes(w.text.toLowerCase())
        );
        
        if (newWords.length === 0) {
          ElMessage.info('没有新单词需要导入');
          return;
        }
        
        // 更新ID避免冲突
        const maxId = Math.max(...words.value.map(w => w.id), 0);
        newWords.forEach((w: any, index: number) => {
          w.id = maxId + index + 1;
        });
        
        // 新导入的单词放在最上面
        words.value = [...newWords, ...words.value];
        saveWordsToLocalStorage();
        ElMessage.success(`导入成功，新增 ${newWords.length} 个单词`);
      } catch (error) {
        console.error('导入失败:', error);
        ElMessage.error('导入失败，请检查文件格式');
      }
    };
    
    // 触发文件选择
    input.click();
  } catch (error) {
    console.error('导入失败:', error);
    ElMessage.error('导入失败，请重试');
  }
};

// 添加单词
const addWord = () => {
  // 重置新单词数据
  newWord.value = { text: '', phonetic: '', meaning: '' };
  // 显示添加单词对话框
  addWordDialogVisible.value = true;
};

// 重置新单词数据
const resetNewWord = () => {
  newWord.value = { text: '', phonetic: '', meaning: '' };
};

// 查询单词音标和释义
const lookupWord = async () => {
  if (!newWord.value.text.trim()) {
    ElMessage.error('请输入单词');
    return;
  }
  
  try {
    // 使用多个API查询单词信息
    const word = newWord.value.text.trim();
    
    // 1. 尝试使用Google翻译API获取翻译
    const googleResponse = await fetch(
      `https://translate.googleapis.com/translate_a/single?client=gtx&sl=en&tl=zh-CN&dt=t&q=${encodeURIComponent(word)}`
    );
    
    if (googleResponse.ok) {
      const googleData = await googleResponse.json();
      if (googleData && googleData[0] && googleData[0][0]) {
        newWord.value.meaning = googleData[0][0][0];
      }
    }
    
    // 2. 尝试使用字典API获取音标
    // 这里使用一个免费的字典API示例
    try {
      const dictResponse = await fetch(`https://api.dictionaryapi.dev/api/v2/entries/en/${encodeURIComponent(word)}`);
      if (dictResponse.ok) {
        const dictData = await dictResponse.json();
        if (dictData && dictData[0] && dictData[0].phonetic) {
          newWord.value.phonetic = dictData[0].phonetic;
        } else if (dictData && dictData[0] && dictData[0].phonetics && dictData[0].phonetics[0]) {
          // 有些API返回的是phonetics数组
          const phoneticObj = dictData[0].phonetics.find((p: any) => p.text);
          if (phoneticObj && phoneticObj.text) {
            newWord.value.phonetic = phoneticObj.text;
          }
        }
      }
    } catch (dictError) {
      // 字典API查询失败，不影响主流程
      console.log('字典API查询失败:', dictError);
    }
    
    ElMessage.success('查询完成');
  } catch (error) {
    console.error('查询失败:', error);
    ElMessage.error('查询失败，请手动输入音标和释义');
  }
};

// 清除音标
const clearPhonetic = () => {
  newWord.value.phonetic = '';
};

// 清除释义
const clearMeaning = () => {
  newWord.value.meaning = '';
};

// 格式化音标，自动添加 / 符号
const formatPhonetic = (phonetic: string): string => {
  if (!phonetic) return '';
  
  const trimmed = phonetic.trim();
  if (!trimmed) return '';
  
  // 如果已经有 / 开头和结尾，直接返回
  if (trimmed.startsWith('/') && trimmed.endsWith('/')) {
    return trimmed;
  }
  
  // 移除可能存在的 / 符号
  let cleaned = trimmed.replace(/^\/*/, '').replace(/\/*$/, '');
  
  // 添加 / 符号
  return `/${cleaned}/`;
};

// 确认添加单词
const confirmAddWord = () => {
  // 简单验证
  if (!newWord.value.text.trim()) {
    ElMessage.error('请输入单词');
    return;
  }
  
  if (!newWord.value.meaning.trim()) {
    ElMessage.error('请输入释义');
    return;
  }
  
  // 检查单词是否已存在
  const wordText = newWord.value.text.trim().toLowerCase();
  const exists = words.value.some(w => w.text.toLowerCase() === wordText);
  if (exists) {
    ElMessage.warning('该单词已存在，无法重复添加');
    return;
  }
  
  // 添加到单词列表（新单词添加到最前面）
  const word = {
    id: Date.now(), // 使用时间戳作为唯一ID
    text: newWord.value.text.trim(),
    phonetic: formatPhonetic(newWord.value.phonetic), // 格式化音标
    meaning: newWord.value.meaning.trim(),
    showSyllables: false,
    category: 'new' // 新添加的单词默认分类为生词
  };
  
  // 将新单词添加到数组开头
  words.value.unshift(word);
  
  // 保存到本地存储
  saveWordsToLocalStorage();
  
  // 关闭对话框
  addWordDialogVisible.value = false;
  
  ElMessage.success('单词添加成功');
};

// 朗读单词
const speakWord = (word: any) => {
  // 检查浏览器是否支持语音合成
  if ('speechSynthesis' in window) {
    // 创建语音合成实例
    const utterance = new SpeechSynthesisUtterance(word.text);
    
    // 设置语音参数（可选）
    utterance.lang = 'en-US'; // 设置语言为英语
    utterance.rate = 1; // 语速
    utterance.pitch = 1; // 音调
    utterance.volume = 1; // 音量
    
    // 播放语音
    window.speechSynthesis.speak(utterance);
    
    ElMessage.success(`正在朗读: ${word.text}`);
  } else {
    ElMessage.error('您的浏览器不支持语音合成功能');
  }
};

// 保存API密钥
const saveApiKeys = () => {
  ElMessage.success('API密钥保存成功');
};

// 打开设置对话框
const openSettings = () => {
  ElMessage.info('API设置功能已移除');
};

// 编辑单词
const editWord = (word: any) => {
  editingWord.value = { ...word }; // 复制单词数据
  editWordDialogVisible.value = true;
};

// 确认编辑单词
const confirmEditWord = () => {
  if (!editingWord.value.text.trim()) {
    ElMessage.error('请输入单词');
    return;
  }
  
  if (!editingWord.value.meaning.trim()) {
    ElMessage.error('请输入释义');
    return;
  }
  
  // 查找并更新单词
  const index = words.value.findIndex(w => w.id === editingWord.value.id);
  if (index !== -1) {
    words.value[index] = {
      ...editingWord.value,
      text: editingWord.value.text.trim(),
      phonetic: formatPhonetic(editingWord.value.phonetic), // 格式化音标
      meaning: editingWord.value.meaning.trim()
    };
    
    // 保存到本地存储
    saveWordsToLocalStorage();
    
    ElMessage.success('单词修改成功');
  }
  
  editWordDialogVisible.value = false;
  editingWord.value = null;
};

// 取消编辑
const cancelEdit = () => {
  editWordDialogVisible.value = false;
  editingWord.value = null;
};

// 删除单词
const deleteWord = (id: number) => {
  // 找到要删除的单词
  const word = words.value.find(w => w.id === id);
  if (!word) return;
  
  ElMessageBox.confirm(
    `确定要删除单词 "${word.text}" 吗？`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  )
    .then(() => {
      console.log('Deleting word with id:', id);
      words.value = words.value.filter(word => word.id !== id);
      
      // 保存到本地存储
      saveWordsToLocalStorage();
      
      ElMessage.success('单词删除成功');
      console.log('Remaining words:', words.value);
    })
    .catch(() => {
      // 用户取消删除
      ElMessage.info('已取消删除');
    });
};

// 点击外部关闭
const handleClickOutside = (e: MouseEvent) => {
  const element = document.querySelector('.wordbook-float-icon');
  if (element && !element.contains(e.target as Node)) {
    if (isExpanded.value) {
      isExpanded.value = false;
    }
  }
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
  // 从本地存储加载单词
  loadWordsFromLocalStorage();
});

// 监听 visible 属性变化
watch(() => props.visible, (newVal) => {
  if (newVal) {
    // 显示时展开窗口
    isExpanded.value = true;
  }
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside);
  document.removeEventListener('mousemove', onDrag);
  document.removeEventListener('mouseup', stopDrag);
  if (animationFrameId) {
    cancelAnimationFrame(animationFrameId);
  }
  
  // 移除临时事件监听器
  document.removeEventListener('mousemove', handleMouseMove);
  document.removeEventListener('mouseup', handleMouseUp);
});
</script>

<style scoped>
.wordbook-float-icon {
  position: fixed;
  z-index: 2000;
  user-select: none;
  transition: none; /* 移除过渡动画以提升拖拽性能 */
}

.wordbook-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  transition: transform 0.2s, box-shadow 0.2s; /* 仅对变换和阴影使用过渡 */
  cursor: move;
}

.wordbook-icon:hover {
  transform: scale(1.1);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.3);
}

.wordbook-window {
  width: 600px;
  min-width: 500px;
  max-width: 800px;
  height: 500px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid #e4e7ed;
}

.wordbook-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-weight: bold;
  cursor: move;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.header-actions .el-button {
  color: white;
  padding: 4px;
}

.header-actions .io-button {
  font-size: 12px;
  font-weight: bold;
  padding: 4px 8px;
}

.wordbook-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 16px;
  overflow: hidden;
  user-select: text;
}

.category-stats {
  display: flex;
  gap: 20px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
  margin-bottom: 12px;
}

.stat-item {
  font-size: 13px;
  color: #606266;
  font-weight: 500;
}

.search-bar {
  margin-bottom: 16px;
}

.word-list {
  flex: 1;
  overflow-y: auto;
  margin-bottom: 16px;
}

.word-item {
  padding: 12px;
  border-bottom: 1px solid #eee;
  user-select: text;
}

.word-item:last-child {
  border-bottom: none;
}

.word-main {
  display: flex;
  align-items: center;
  gap: 0px;
  margin-bottom: 4px;
}

.word-main strong {
  margin-right: 4px;
  transition: font-size 0.2s;
}

.word-main strong.syllable-mode {
  font-size: 1.5em;
}

.word-phonetic {
  font-size: 12px;
  color: #909399;
  margin-right: 4px;
}

.word-main .el-button {
  padding: 0 4px;
  margin: 0;
}

.category-tag {
  display: inline-block;
  padding: 2px 6px;
  font-size: 12px;
  border-radius: 3px;
  margin: 0 4px;
  font-weight: 500;
}

.tag-new {
  background: #e1f3d8;
  color: #67c23a;
}

.tag-review {
  background: #fdf6ec;
  color: #e6a23c;
}

.tag-mastered {
  background: #f0f9ff;
  color: #409eff;
}

.word-main .el-button:last-child {
  margin-left: auto;
}

.el-dropdown-link {
  display: inline-block;
  cursor: pointer;
}

/* 修复下拉菜单定位问题 */
.word-item :deep(.el-dropdown) {
  display: inline-block;
}

.word-item :deep(.el-popper) {
  z-index: 9999 !important;
}

.word-meaning {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
}

.word-actions {
  display: flex;
  justify-content: flex-start;
  gap: 8px;
  margin-top: 8px;
}

.category-tabs {
  margin-bottom: 12px;
}

.category-tabs :deep(.el-tabs__header) {
  margin: 0;
}

.empty-state {
  text-align: center;
  padding: 40px 0;
  color: #909399;
}

.empty-state .el-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.wordbook-footer {
  display: flex;
  justify-content: center;
}
</style>