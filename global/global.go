package global

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	DB  *sql.DB
	Log *zap.SugaredLogger
)

// GetLog 获取带 RequestID 的 Logger
func GetLog(c *gin.Context) *zap.SugaredLogger {
	if c == nil {
		return Log
	}

	reqID, exists := c.Get("RequestID")
	if !exists {
		return Log
	}

	// ★★★ 核心修正：使用 Named ★★★
	// 不要自己加方括号 "[]"，我们在 logger.go 配置里统一加
	// 也不要用 With，用 Named 才能让它跑到日志中间去
	return Log.Named(fmt.Sprintf("%s", reqID))
}
