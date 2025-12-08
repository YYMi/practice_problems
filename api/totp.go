package api

import (
	"database/sql"
	"fmt"
	"practice_problems/global"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

// GenerateTotpSecret 生成TOTP密钥和二维码
func GenerateTotpSecret(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 查询用户信息
	var username string
	var totpSecret sql.NullString
	err := global.DB.QueryRow(
		"SELECT username, totp_secret FROM users WHERE id = ?",
		userID,
	).Scan(&username, &totpSecret)

	if err != nil {
		global.GetLog(c).Errorf("查询用户失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "查询用户失败"})
		return
	}

	// 如果已绑定，返回已绑定状态
	if totpSecret.Valid && totpSecret.String != "" {
		c.JSON(200, gin.H{
			"code": 200,
			"msg":  "已绑定",
			"data": gin.H{
				"bound": true,
			},
		})
		return
	}

	// 生成新的TOTP密钥
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Practice Problems",
		AccountName: username,
	})

	if err != nil {
		global.GetLog(c).Errorf("生成TOTP密钥失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "生成密钥失败"})
		return
	}

	// 返回密钥和二维码URL
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"bound":  false,
			"secret": key.Secret(),
			"qrcode": key.URL(), // otpauth://totp/...
		},
	})
}

// VerifyTotpCode 验证TOTP验证码并绑定
func VerifyTotpCode(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 解析请求参数
	var req struct {
		Secret string `json:"secret" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 验证TOTP码
	valid := totp.Validate(req.Code, req.Secret)
	if !valid {
		global.GetLog(c).Warnf("TOTP验证失败: UserID=%v", userID)
		c.JSON(400, gin.H{"code": 400, "msg": "验证码错误，请重试"})
		return
	}

	// 验证成功，保存密钥到数据库
	_, err := global.DB.Exec(
		"UPDATE users SET totp_secret = ? WHERE id = ?",
		req.Secret, userID,
	)

	if err != nil {
		global.GetLog(c).Errorf("保存TOTP密钥失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "绑定失败"})
		return
	}

	global.GetLog(c).Infof("用户绑定TOTP成功: UserID=%v", userID)
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "绑定成功",
	})
}

// CheckTotpBound 检查用户是否已绑定TOTP
func CheckTotpBound(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 查询是否已绑定
	var totpSecret sql.NullString
	var isAdmin int
	err := global.DB.QueryRow(
		"SELECT totp_secret, is_admin FROM users WHERE id = ?",
		userID,
	).Scan(&totpSecret, &isAdmin)

	if err != nil {
		global.GetLog(c).Errorf("查询用户失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	// 判断是否已绑定
	bound := totpSecret.Valid && totpSecret.String != ""

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"bound":    bound,
			"is_admin": isAdmin,
		},
	})
}

// ValidateTotpCode 验证TOTP码（用于数据库操作前验证）
func ValidateTotpCode(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 解析请求参数
	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 查询用户的TOTP密钥
	var totpSecret sql.NullString
	err := global.DB.QueryRow(
		"SELECT totp_secret FROM users WHERE id = ?",
		userID,
	).Scan(&totpSecret)

	if err != nil {
		global.GetLog(c).Errorf("查询用户失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	// 检查是否已绑定
	if !totpSecret.Valid || totpSecret.String == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "未绑定TOTP"})
		return
	}

	// 验证TOTP码
	valid := totp.Validate(req.Code, totpSecret.String)
	if !valid {
		global.GetLog(c).Warnf("TOTP验证失败: UserID=%v", userID)
		c.JSON(400, gin.H{"code": 400, "msg": "验证码错误"})
		return
	}

	// 验证成功
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "验证成功",
	})
}

// UnbindTotp 解绑TOTP（需要验证码确认）
func UnbindTotp(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 解析请求参数
	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 查询用户的TOTP密钥
	var totpSecret sql.NullString
	err := global.DB.QueryRow(
		"SELECT totp_secret FROM users WHERE id = ?",
		userID,
	).Scan(&totpSecret)

	if err != nil {
		global.GetLog(c).Errorf("查询用户失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	// 检查是否已绑定
	if !totpSecret.Valid || totpSecret.String == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "未绑定TOTP"})
		return
	}

	// 验证TOTP码
	valid := totp.Validate(req.Code, totpSecret.String)
	if !valid {
		global.GetLog(c).Warnf("TOTP验证失败: UserID=%v", userID)
		c.JSON(400, gin.H{"code": 400, "msg": "验证码错误"})
		return
	}

	// 解绑：清空TOTP密钥
	_, err = global.DB.Exec(
		"UPDATE users SET totp_secret = NULL WHERE id = ?",
		userID,
	)

	if err != nil {
		global.GetLog(c).Errorf("解绑TOTP失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "解绑失败"})
		return
	}

	global.GetLog(c).Infof("用户解绑TOTP成功: UserID=%v", userID)
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "解绑成功",
	})
}

// VerifyTotpForOperation 用于数据库操作的TOTP验证辅助函数
func VerifyTotpForOperation(userID int, code string) error {
	// 查询用户的TOTP密钥
	var totpSecret sql.NullString
	err := global.DB.QueryRow(
		"SELECT totp_secret FROM users WHERE id = ?",
		userID,
	).Scan(&totpSecret)

	if err != nil {
		return fmt.Errorf("查询用户失败: %v", err)
	}

	// 检查是否已绑定
	if !totpSecret.Valid || totpSecret.String == "" {
		return fmt.Errorf("未绑定TOTP")
	}

	// 验证TOTP码
	valid := totp.Validate(code, totpSecret.String)
	if !valid {
		return fmt.Errorf("验证码错误")
	}

	return nil
}
