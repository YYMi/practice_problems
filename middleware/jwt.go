package middleware

import (
	"net/http"
	"practice_problems/global" // 确保路径正确
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JwtSecret 密钥
var JwtSecret = []byte("YOUR_SUPER_SECRET_KEY_CHANGE_ME")

// MyClaims 自定义载荷
type MyClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	UserCode string `json:"user_code"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 Token
func GenerateToken(userID int, username string, userCode string) (string, error) {
	claims := MyClaims{
		UserID:   userID,
		Username: username,
		UserCode: userCode,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 30 * time.Hour)), // 24小时有效期
			Issuer:    "practice_system",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtSecret)
}

// JWTAuthMiddleware 鉴权中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "请求未携带 Token"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token 格式错误"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// ==========================================
		// 🔥 核心逻辑：检查内存白名单
		// ==========================================
		exists, storedUserCode := global.VerifyToken(tokenString)
		if !exists {
			// 哪怕 Token 签名是对的，只要内存里没有（重启了/退出了），就视为无效
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "登录已失效，请重新登录"})
			c.Abort()
			return
		}
		// ==========================================

		// 虽然内存验证通过了，还是解析一下拿到 UserID
		token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
			return JwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token 解析失败"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*MyClaims); ok {
			c.Set("userID", claims.UserID)
			c.Set("username", claims.Username)
			// 这里用内存里查出来的 userCode，双重保险
			c.Set("userCode", storedUserCode)
		}

		c.Next()
	}
}
