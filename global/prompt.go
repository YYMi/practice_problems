package global

import "os"

// LoadPromptTemplate 读取 prompt.txt 文件
// 你可以在 main.go 启动时调用一次，或者每次建立连接时调用（支持热更）
func LoadPromptTemplate() string {
	// 这里为了简单，每次都读文件，方便你随时改 txt 生效
	// 生产环境建议读一次存内存，或者用 fsnotify 监听文件变化
	content, err := os.ReadFile("uploads/prompt.txt")
	if err != nil {
		// 如果读不到文件，返回一个默认的保底
		return "你是一位面试官，当前题目是：%s"
	}
	return string(content)
}
