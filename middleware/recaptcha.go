package middleware

import (
	"encoding/json"
	"net/http"
	"practice_problems/global"

	"github.com/gin-gonic/gin"
)

// RecaptchaRequest 请求结构
type RecaptchaRequest struct {
	RecaptchaToken string `json:"recaptcha_token" binding:"required"`
}

// RecaptchaResponse Google reCAPTCHA响应结构
type RecaptchaResponse struct {
	Success     bool     `json:"success"`
	Score       float64  `json:"score"`
	Action      string   `json:"action"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

// RecaptchaMiddleware Google reCAPTCHA验证中间件
// 用于验证前端提交的reCAPTCHA token
func RecaptchaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求体中提取 recaptcha_token
		var req struct {
			RecaptchaToken string `json:"recaptcha_token"`
		}

		// 注意：由于我们需要读取body，需要使用ShouldBindJSON
		// 但这会消耗body，所以业务逻辑需要从context获取数据
		if err := c.ShouldBindJSON(&req); err != nil {
			global.GetLog(c).Warnf("reCAPTCHA验证失败: 缺少验证码参数")
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 400,
				"msg":  "缺少Google验证码",
			})
			c.Abort()
			return
		}

		// 验证recaptcha token
		if !verifyRecaptcha(req.RecaptchaToken) {
			global.GetLog(c).Warnf("reCAPTCHA验证失败: token无效")
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Google验证码校验失败",
			})
			c.Abort()
			return
		}

		// 验证通过，继续处理
		c.Next()
	}
}

// verifyRecaptcha 验证reCAPTCHA token
func verifyRecaptcha(token string) bool {
	// TODO: 这里需要配置你的reCAPTCHA密钥
	// 从配置文件或环境变量读取
	secretKey := "YOUR_RECAPTCHA_SECRET_KEY" // 替换为实际的密钥

	// 如果密钥未配置，临时放行（开发模式）
	if secretKey == "YOUR_RECAPTCHA_SECRET_KEY" {
		// 开发模式：跳过验证
		return true
	}

	// 调用Google reCAPTCHA API验证
	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", map[string][]string{
		"secret":   {secretKey},
		"response": {token},
	})
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var result RecaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	// 验证成功条件：success为true且分数大于0.5（可调整）
	return result.Success && result.Score > 0.5
}
