package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware 为每个请求生成唯一 ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 生成 UUID
		fullID := uuid.New().String()

		// 2. ★★★ 截取前 8 位，并去掉横杠 (以防万一) ★★★
		// strings.ReplaceAll 确保去掉 UUID 里的 "-"
		cleanID := strings.ReplaceAll(fullID, "-", "")
		shortID := cleanID[:8]

		// 3. 塞入上下文
		c.Set("RequestID", shortID)
		c.Writer.Header().Set("X-Request-ID", shortID)

		c.Next()
	}
}
