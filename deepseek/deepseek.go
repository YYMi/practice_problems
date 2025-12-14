package deepseek

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"practice_problems/global"
	"sync"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// Global state variables
var (
	client  *openai.Client
	isReady bool
	initErr error
	mu      sync.RWMutex
)

const (
	// DeepSeek 的 BaseURL (建议带上 /v1)
	DeepSeekBaseURL = "https://api.deepseek.com/v1"

	// ==========================================
	// 核心设置：使用 DeepSeek R1 推理模型
	// ==========================================
	//DeepSeekModelName = "deepseek-reasoner"
	DeepSeekModelName = "deepseek-chat"
)

// Init 初始化 DeepSeek 客户端
func Init(apiKey string) error {
	mu.Lock()
	defer mu.Unlock()

	if apiKey == "" {
		initErr = errors.New("deepseek api key is empty")
		isReady = false
		return initErr
	}

	// 1. 配置 Client
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = DeepSeekBaseURL

	// 2. 设置 HTTP 超时时间
	// R1 推理模型思考时间较长，必须设置足够长的超时 (这里设为 3 分钟)
	config.HTTPClient = &http.Client{
		Timeout: 3 * time.Minute,
	}

	// 3. 创建实例
	c := openai.NewClientWithConfig(config)

	// 4. 验证连接 (尝试列出模型)
	// 给验证请求一个短一点的超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := c.ListModels(ctx)
	if err != nil {
		initErr = fmt.Errorf("failed to connect to DeepSeek API: %v", err)
		isReady = false
		client = nil
		global.GetLog(nil).Error("[DeepSeek] Initialization failed: %v", initErr)
		return initErr
	}

	client = c
	isReady = true
	initErr = nil
	log.Println("[DeepSeek] Client initialized (Model: deepseek-reasoner, Timeout: 3min).")
	return nil
}

// GetClient 获取客户端实例 (线程安全)
func GetClient() *openai.Client {
	mu.RLock()
	defer mu.RUnlock()
	if !isReady {
		return nil
	}
	return client
}

// IsReady 返回当前 AI 服务是否可用
func IsReady() (bool, error) {
	mu.RLock()
	defer mu.RUnlock()
	return isReady, initErr
}

// Chat 核心对话方法
func Chat(ctx context.Context, messages []openai.ChatCompletionMessage) (string, error) {
	cli := GetClient()
	if cli == nil {
		return "", errors.New("deepseek client is not ready")
	}

	// 调用 API
	resp, err := cli.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    DeepSeekModelName, // deepseek-reasoner
			Messages: messages,

			// 推理模型参数建议：
			// Temperature: 0.6 (0.5-0.7 适合推理，不要太高)
			Temperature: 0.6,

			// MaxTokens: 10000 (防止回答被截断，特别是包含思维链的时候)
			MaxTokens: 8192,
		},
	)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("empty response from deepseek")
	}

	// 获取回答内容
	// 注意：deepseek-reasoner 的返回中，Content 是最终结论
	// 如果你需要思维链内容(ReasoningContent)，需要 SDK 支持或自己解析，通常 Content 就够用了
	return resp.Choices[0].Message.Content, nil
}
