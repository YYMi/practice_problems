package middleware

import (
	"fmt"
	"practice_problems/global"
	"time"

	"github.com/gin-gonic/gin"
)

// GinLogger 替代 Gin 默认的 Logger 中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 开始时间
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 2. 处理请求
		c.Next()

		// 3. 结束时间 & 计算耗时
		end := time.Now()
		latency := end.Sub(start)

		// 4. 获取请求信息
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + " ? " + raw
		}
		// 1. 转为毫秒浮点数 (例如 0.51 或 12.3)
		msFloat := float64(latency.Nanoseconds()) / 1e6

		// 2. 四舍五入转为整数 (0.51 -> 1, 0.49 -> 0, 12.3 -> 12)
		// 逻辑：加 0.5 然后转 int
		costInt := int64(msFloat + 0.5)

		msg := fmt.Sprintf("[%d] [%dms] [%s] %s %s",
			statusCode,
			costInt,
			clientIP,
			method,
			path,
		)

		// 6. 根据状态码决定打印级别 (可选)
		logger := global.GetLog(c) // 获取带 RequestID 的 Logger

		if statusCode >= 500 {
			logger.Errorf("[GIN] %s", msg) // 服务器错误用 Error
		} else if statusCode >= 400 {
			logger.Warnf("[GIN] %s", msg) // 客户端错误用 Warn
		} else {
			logger.Infof("[GIN] %s", msg) // 正常用 Info
		}
	}
}
