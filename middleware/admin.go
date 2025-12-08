package middleware

import (
	"net/http"
	"practice_problems/global"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware 管理员权限中间件
// 必须在JWTAuthMiddleware之后使用
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从context获取用户ID
		userID, exists := c.Get("userID")
		if !exists {
			global.GetLog(c).Warnf("权限校验失败: 未找到用户ID")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "未授权访问",
			})
			c.Abort()
			return
		}

		// 查询用户是否是管理员
		var isAdmin int
		err := global.DB.QueryRow("SELECT is_admin FROM users WHERE id = ?", userID).Scan(&isAdmin)
		if err != nil {
			global.GetLog(c).Errorf("查询用户权限失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "权限查询失败",
			})
			c.Abort()
			return
		}

		// 检查是否是管理员
		if isAdmin != 1 {
			global.GetLog(c).Warnf("非管理员尝试访问管理接口: UserID=%v", userID)
			c.JSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "需要管理员权限",
			})
			c.Abort()
			return
		}

		// 管理员权限验证通过
		c.Next()
	}
}
