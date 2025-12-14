package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"practice_problems/deepseek"
	"practice_problems/global"
	"practice_problems/middleware"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	openai "github.com/sashabaranov/go-openai"
)

// ==========================================
// 1. 全局配置与工具函数
// ==========================================

// websocket 升级器配置
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // 允许跨域
}

// UserInputObj 前端发送的 JSON 结构
// 必须包含 topic (题目) 和 content (回答内容)
type UserInputObj struct {
	Topic   string `json:"topic"`   // 题目：Java内存模型
	Content string `json:"content"` // 回答：我觉得是...
}

// WSMessage WebSocket 消息通用载荷
type WSMessage struct {
	Type    string      `json:"type"`    // 消息类型: init, chat, error, quota_exhausted
	Content interface{} `json:"content"` // 消息内容
}

// AIInterviewSession 每一个 Socket 连接对应一个 Session
type AIInterviewSession struct {
	UserID      int
	Username    string
	StartTime   time.Time
	UsedSeconds int64
	Quota       int64
	Conn        *websocket.Conn
	mu          sync.Mutex
	stopTimer   chan struct{}
	closed      bool

	// 核心：使用 Map 存储不同题目的聊天记录
	// Key: 题目名称 (Topic)
	// Value: 该题目的聊天上下文 (System + Assistant + User...)
	TopicHistories map[string][]openai.ChatCompletionMessage
}

// LoadPromptTemplate 读取 prompt.txt 文件
// 如果文件不存在，返回默认的保底 Prompt
func LoadPromptTemplate() string {
	content, err := os.ReadFile("uploads/prompt.txt")
	if err != nil {
		// 保底提示词
		return `你是一位专业的面试官。当前面试题目是：「%s」。
请注意：
1. 我会发送用户的【回答】给你。
2. 请评估回答是否正确。若正确，请进行深挖追问；若错误，请指出。`
	}
	return string(content)
}

// ==========================================
// 2. Controller 入口
// ==========================================

// AIInterviewWebSocket AI 面试官 WebSocket 接口
func AIInterviewWebSocket(c *gin.Context) {
	// 1. 获取参数
	token := c.Query("token")
	initTopic := c.Query("point_title")

	// ==========================================
	// 🔥 核心鉴权逻辑 (升级前检查)
	// ==========================================

	// A. 基础空值检查
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未携带 Token"})
		return
	}

	// B. 检查内存白名单 (防止已登出的 Token 连接)
	exists, _ := global.VerifyToken(token)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token 已失效或已登出"})
		return
	}

	// C. 解析 Token 获取用户信息
	claims, err := middleware.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token 解析失败"})
		return
	}

	// ==========================================
	// 🚀 鉴权通过，升级 WebSocket
	// ==========================================
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.GetLog(nil).Error("WebSocket Upgrade Failed: %v", err)
		return
	}

	// 辅助关闭函数 (连接建立后的错误处理)
	sendErrorAndClose := func(errType string, code int, message string) {
		msg, _ := json.Marshal(WSMessage{
			Type:    errType,
			Content: map[string]interface{}{"code": code, "message": message},
		})
		conn.WriteMessage(websocket.TextMessage, msg)
		conn.Close()
	}

	// 2. 检查 AI 服务状态
	if ready, err := deepseek.IsReady(); !ready {
		sendErrorAndClose("error", 503, fmt.Sprintf("AI服务不可用: %v", err))
		return
	}

	// 3. 检查用户配额 (从数据库查)
	var aiQuota int64
	err = global.DB.QueryRow("SELECT COALESCE(ai_quota, 0) FROM users WHERE id = ?", claims.UserID).Scan(&aiQuota)
	if err != nil {
		sendErrorAndClose("error", 500, "数据库查询失败")
		return
	}
	if aiQuota <= 0 {
		sendErrorAndClose("quota_error", 403, "您的 AI 面试时长已用尽")
		return
	}

	// 4. 初始化 Session
	session := &AIInterviewSession{
		UserID:         claims.UserID,
		Username:       claims.Username,
		StartTime:      time.Now(),
		Quota:          aiQuota,
		Conn:           conn,
		stopTimer:      make(chan struct{}),
		closed:         false,
		TopicHistories: make(map[string][]openai.ChatCompletionMessage),
	}

	// 5. 发送初始化成功消息
	session.sendRawMessage(WSMessage{Type: "init", Content: map[string]interface{}{"quota": aiQuota}})

	// 6. 发送静态欢迎语 (回显题目)
	if initTopic != "" {
		welcomeMsg := fmt.Sprintf("同学你好，我是你的 AI 面试官。\n\n基于题目 **「%s」**，请简要介绍一下你的理解。", initTopic)
		session.sendRawMessage(WSMessage{Type: "chat", Content: welcomeMsg})
	}

	// 7. 启动
	go session.startTimer()
	session.handleMessages()
}

// ==========================================
// 3. Session 逻辑实现
// ==========================================

// sendRawMessage 发送 WebSocket 消息 (线程安全)
func (s *AIInterviewSession) sendRawMessage(msg WSMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	data, _ := json.Marshal(msg)
	s.Conn.WriteMessage(websocket.TextMessage, data)
}

// startTimer 扣费计时器
func (s *AIInterviewSession) startTimer() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			if s.closed {
				s.mu.Unlock()
				return
			}

			s.UsedSeconds++
			remaining := s.Quota - s.UsedSeconds

			// 配额耗尽处理
			if remaining <= 0 {
				// 发送耗尽通知
				exhaustedMsg, _ := json.Marshal(WSMessage{Type: "quota_exhausted", Content: "时长已耗尽"})
				s.Conn.WriteMessage(websocket.TextMessage, exhaustedMsg)
				s.mu.Unlock()
				s.close() // 强制关闭
				return
			}
			s.mu.Unlock()

		case <-s.stopTimer:
			return
		}
	}
}

// handleMessages 循环读取前端消息
func (s *AIInterviewSession) handleMessages() {
	defer s.close() // 循环结束（连接断开）时自动触发清理

	for {
		_, msgData, err := s.Conn.ReadMessage()
		if err != nil {
			// 读不到消息通常意味着连接断开
			break
		}

		var msg WSMessage
		// 解析外层结构
		if err := json.Unmarshal(msgData, &msg); err != nil {
			continue
		}

		// 只处理 chat 类型的消息
		if msg.Type == "chat" {
			// 将 Content 转为 Map 来获取 topic 和 answer
			// 前端传来的数据结构: { "type": "chat", "content": { "topic": "...", "content": "..." } }
			contentMap, ok := msg.Content.(map[string]interface{})
			if !ok {
				// 容错：防止前端发来的不是 JSON 对象
				continue
			}

			topic, _ := contentMap["topic"].(string)
			answer, _ := contentMap["content"].(string)

			// 必须要有题目和回答才处理
			if topic != "" && answer != "" {
				s.handleChatLogic(topic, answer)
			}
		}
	}
}

// handleChatLogic 核心业务逻辑：组装上下文 -> 裁剪(防爆) -> 调用 AI
func (s *AIInterviewSession) handleChatLogic(topic string, userAnswer string) {
	s.mu.Lock()

	// 1. 获取或创建该题目的聊天历史
	history, exists := s.TopicHistories[topic]

	if !exists {
		// --- 情况 A: 新题目，初始化上下文 ---
		// 动态读取 Prompt
		tpl := LoadPromptTemplate()
		systemPrompt := fmt.Sprintf(tpl, topic)

		// 伪造 AI 的上一句提问 (为了让 AI 知道它问了什么)
		fakeAiQuestion := fmt.Sprintf("同学你好，我是你的 AI 面试官。基于题目「%s」，请简要介绍一下你的理解。", topic)

		history = []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleAssistant, Content: fakeAiQuestion},
			{Role: openai.ChatMessageRoleUser, Content: userAnswer}, // 追加用户当前的回答
		}
	} else {
		// --- 情况 B: 老题目，追加回答 ---
		history = append(history, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: userAnswer,
		})
	}

	// 2. 更新 Map (防止数据丢失)
	s.TopicHistories[topic] = history

	// =====================================================
	// 【防爆逻辑】裁剪历史记录，防止 Token 爆炸
	// 策略：保留 SystemPrompt + 最近的 N 轮
	// =====================================================
	const MaxHistoryRounds = 20 // 保留最近 20 条消息 (约 10 轮对话)

	// 用于发送给 AI 的临时切片
	var inputHistory []openai.ChatCompletionMessage

	// history[0] 是 System Prompt，我们要永久保留
	// 如果总长度超过了限制
	if len(history) > MaxHistoryRounds {
		inputHistory = make([]openai.ChatCompletionMessage, 0, MaxHistoryRounds+1)

		// 1. 必须保留 System Prompt (他是面试官的身份设定)
		inputHistory = append(inputHistory, history[0])

		// 2. 计算截断点，保留后半截
		// 比如总共 100 条，Max=20。我们取后 19 条拼在 System 后面
		cutoffIndex := len(history) - (MaxHistoryRounds - 1)
		if cutoffIndex < 1 {
			cutoffIndex = 1
		}

		inputHistory = append(inputHistory, history[cutoffIndex:]...)
	} else {
		// 没超过限制，全发
		inputHistory = make([]openai.ChatCompletionMessage, len(history))
		copy(inputHistory, history)
	}

	s.mu.Unlock() // 解锁，让 AI 慢慢思考

	// 3. 调用 DeepSeek (设置 3分钟 超时，匹配推理模型)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	reply, err := deepseek.Chat(ctx, inputHistory)
	if err != nil {
		global.GetLog(nil).Error("[AI Interview] Chat Error: %v", err)
		s.sendRawMessage(WSMessage{Type: "error", Content: "AI 思考超时或服务繁忙，请重试"})
		return
	}

	// 4. 收到 AI 回复，存入历史记录并发送给前端
	s.mu.Lock()
	if !s.closed {
		// 重新取出最新的 History (防止期间有并发写入)
		currentHist := s.TopicHistories[topic]
		currentHist = append(currentHist, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: reply,
		})
		s.TopicHistories[topic] = currentHist
	}
	s.mu.Unlock()

	// 发送给前端
	s.sendRawMessage(WSMessage{Type: "chat", Content: reply})
}

// close 清理资源并保存数据
func (s *AIInterviewSession) close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	s.mu.Unlock()

	// 停止计时器
	close(s.stopTimer)
	// 关闭连接
	s.Conn.Close()

	// 结算扣费
	if s.UsedSeconds > 0 {
		newQuota := s.Quota - s.UsedSeconds
		if newQuota < 0 {
			newQuota = 0
		}
		// 更新数据库
		_, err := global.DB.Exec("UPDATE users SET ai_quota = ? WHERE id = ?", newQuota, s.UserID)
		if err != nil {
			global.GetLog(nil).Error("Failed to update user quota: %v", err)
		} else {
			fmt.Printf("[AI Session End] User: %s, Used: %ds, Remaining: %ds\n", s.Username, s.UsedSeconds, newQuota)
		}
	}
}
