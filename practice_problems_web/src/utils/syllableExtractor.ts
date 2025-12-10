/**
 * 音节提取工具 (V4.0 - 修复美式长音符号误判问题)
 * 完美解决 repository -> re·pos·i·to·ry
 */

/* ================= 1. 核心配置 (Critical Config) ================= */

// 定义哪些 IPA 音素表现出“短元音特性” (即倾向于吸附后面的辅音形成闭音节)
// 修改点：加入了 "ɑ", "ɑː", "a" 
// 原因：在美式音标中，hot/pot/repository 中的 o 常标为 /ɑː/，但拼读上属于短音规则
const SHORT_VOWELS = new Set([
  "æ", "e", "ɪ", "ɒ", "ʌ", "ʊ", "ɛ", 
  "ɑ", "ɑː", "ɑ:", "a", "aː" // <--- 新增这些，专门处理美式 "Short O"
]);

// 必须连在一起的辅音组合 (Digraphs)
const DIGRAPHS = [
  "th", "sh", "ch", "ph", "wh", "ck", "qu", "ng",
  "tr", "dr", "cr", "br", "fr", "gr", "pr",
  "st", "sp", "sk", "sl", "sm", "sn", "sw",
  "cl", "bl", "fl", "gl", "pl"
];

/* ================= 2. IPA 解析器 ================= */

interface IPAVowelInfo {
  phone: string;
  isStressed: boolean;
  isShort: boolean;
}

function parseIPAVowels(ipa: string): IPAVowelInfo[] {
  // 预处理：有些音标源用普通的冒号 : 代替长音符号 ː ，统一替换方便处理
  const clean = ipa.replace(/:/g, "ː").replace(/[\/\[\]\\]/g, "");
  
  const result: IPAVowelInfo[] = [];
  let isNextStressed = false;
  
  for (let i = 0; i < clean.length; i++) {
    const char = clean[i];
    
    // 标记重音
    if (char === 'ˈ') {
      isNextStressed = true;
      continue;
    }
    // 次重音 (ˌ) 有时也需要吸附辅音 (如 composition -> com·po·si·tion)
    // 这里我们把次重音也视为 stress 的一种，增强吸附力
    if (char === 'ˌ') {
      isNextStressed = true; 
      continue;
    }

    // 识别元音
    if (/[aeiouəʌæɒɔʊɪɛɜɑ]/.test(char)) {
      let vowel = char;
      
      // 向后看1位：处理长音符 (ː) 或双元音组合
      if (i + 1 < clean.length) {
        const next = clean[i+1];
        if (next === 'ː') {
          vowel += 'ː';
          i++; 
        } else if (/[aeiouəɪʊ]/.test(next)) {
          // 简单的双元音匹配 (如 aɪ, oʊ)
          vowel += next;
          i++;
          // 如果双元音后面还有长音符 (极少见但以防万一)
          if (i + 1 < clean.length && clean[i+1] === 'ː') {
            i++;
          }
        }
      }

      // 核心判断：这个元音在我们的短音白名单里吗？
      const isShort = SHORT_VOWELS.has(vowel);
      
      result.push({
        phone: vowel,
        isStressed: isNextStressed,
        isShort: isShort
      });

      isNextStressed = false; // 重置
    }
  }
  return result;
}

/* ================= 3. 拼写与切割逻辑 ================= */

interface VowelGroup {
  start: number;
  end: number;
  text: string;
}

function findSpellingVowelGroups(word: string): VowelGroup[] {
  // 包含 y 作为元音
  const matches = word.matchAll(/[aeiouy]+/gi);
  const groups: VowelGroup[] = [];
  for (const m of matches) {
    if (m.index !== undefined) {
      groups.push({
        start: m.index,
        end: m.index + m[0].length,
        text: m[0]
      });
    }
  }
  return groups;
}

function findCutPosition(bridge: string, prevVowelInfo: IPAVowelInfo | undefined): number {
  const len = bridge.length;
  if (len === 0) return 0;

  // === 规则 1: 单辅音处理 (re.po vs rep.o) ===
  if (len === 1) {
    // 只要是【重读】且是【短元音特性】，就吸附辅音
    // 现在 /ɑː/ 被视为短元音特性，所以会进入这里
    if (prevVowelInfo && prevVowelInfo.isStressed && prevVowelInfo.isShort) {
      return 1; // 切后 -> pos.i
    }
    return 0; // 切前 -> re.po
  }

  // === 规则 2: 双写辅音必切 ===
  if (len === 2 && bridge[0].toLowerCase() === bridge[1].toLowerCase()) {
    return 1; // cur.rent
  }

  // === 规则 3: 保护 Digraphs ===
  if (DIGRAPHS.includes(bridge.toLowerCase())) {
    return 0; // tea.cher
  }

  // === 规则 4: 多辅音默认切分 ===
  // 优先保护尾部的 Digraph (如 nstr -> n.str)
  for (let i = 0; i < len - 1; i++) {
    const tail = bridge.slice(i + 1).toLowerCase();
    if (DIGRAPHS.includes(tail)) {
      return i + 1;
    }
  }

  // 默认：从第一个辅音后切 (con.current)
  return 1;
}

/* ================= 4. 主流程 ================= */

export function extractSyllables(word: string, phonetic: string): string {
  const cleanWord = word.trim();
  
  // 1. 解析 IPA
  const ipaVowels = parseIPAVowels(phonetic);
  const targetCount = ipaVowels.length;
  
  if (targetCount <= 1) return cleanWord;

  // 2. 找到拼写元音
  let vowelGroups = findSpellingVowelGroups(cleanWord);

  // 3. 修正数量 (处理 Silent E)
  // 策略：如果拼写元音组比 IPA 多，通常是结尾的 e 不发音，或者是 diphthong 被当成两个字母了
  // 我们尝试从尾部修剪
  while (vowelGroups.length > targetCount) {
    const last = vowelGroups[vowelGroups.length - 1];
    
    // 如果最后是 'e' (like make, repository 没有这个问题，但为了通用性)
    if (last.text.toLowerCase().startsWith('e')) {
      vowelGroups.pop(); 
    } else {
      // 强制对齐：如果还多，可能是 io, ia 这种连在一起的元音被当成了两组
      // 这里的处理比较粗暴，但对于辅助脚本来说，去掉尾部比中间乱切要安全
      vowelGroups.pop();
    }
  }

  // 再次检查
  if (vowelGroups.length <= 1) return cleanWord;

  // 4. 切割
  const syllables: string[] = [];
  let lastCutIndex = 0;

  // 只需要遍历到倒数第二个元音组，因为我们是在两个组之间切
  // 注意：这里需要防止 ipaVowels 越界 (如果对齐失败)
  const loopLimit = Math.min(vowelGroups.length - 1, ipaVowels.length);

  for (let i = 0; i < loopLimit; i++) {
    const v1 = vowelGroups[i];
    const v2 = vowelGroups[i + 1];
    
    // 获取对应的 IPA 信息
    const v1IPA = ipaVowels[i]; 

    const bridgeStart = v1.end;
    const bridgeEnd = v2.start;
    const bridge = cleanWord.slice(bridgeStart, bridgeEnd);

    const cutOffset = findCutPosition(bridge, v1IPA);
    const absoluteCutIndex = bridgeStart + cutOffset;

    syllables.push(cleanWord.slice(lastCutIndex, absoluteCutIndex));
    lastCutIndex = absoluteCutIndex;
  }

  // 添加剩余部分
  syllables.push(cleanWord.slice(lastCutIndex));

  return syllables.join("·");
}

/* ================= 测试 ================= */
// extractSyllables("repository", "/rɪˈpɑːzətɔːri/");
// 逻辑流程:
// 1. 解析 IPA: 
//    - /rɪ/ (弱)
//    - /ˈpɑː/ (重读, 在 SHORT_VOWELS 列表中!)
//    - /zə/ (弱)
//    - ...
// 2. 处理 'pos' 之间的 'o' 和 'i':
//    - IPA 是 /ˈpɑː/。规则：Stressed + Short -> 吸附辅音。
//    - 辅音桥是 's'。
//    - 切割点：s 后面。
// 3. 结果: re·pos·i·to·ry
