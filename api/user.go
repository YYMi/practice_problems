package api

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/rand"
	"practice_problems/global"
	"practice_problems/middleware"
	"practice_problems/model"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// =================================================================
// 辅助函数：MD5 加密
// =================================================================
func md5V(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// 生成 8 位随机数字字符串
func generateRandomCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%08d", r.Intn(100000000))
}

// 获取全表唯一的 UserCode
func getUniqueUserCode() (string, error) {
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		code := generateRandomCode()
		var exists int
		err := global.DB.QueryRow("SELECT 1 FROM users WHERE user_code = ?", code).Scan(&exists)

		if err == sql.ErrNoRows {
			return code, nil
		} else if err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("生成唯一编码失败，请重试")
}

// =======================
// 创建用户 (注册)
// =======================
func CreateUser(c *gin.Context) {
	var req model.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 检查用户表是否为空，如果为空则第一个用户设置为管理员
	var userCount int
	_ = global.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	isFirstUser := userCount == 0

	// ★★★ 修改点：先进行一次后端 MD5，再进行 Bcrypt ★★★
	// 流程：前端MD5 -> 后端MD5 -> Bcrypt -> 数据库
	doubleMd5Pwd := md5V(req.Password)

	hash, err := bcrypt.GenerateFromPassword([]byte(doubleMd5Pwd), bcrypt.DefaultCost)
	if err != nil {
		global.GetLog(c).Errorf("注册失败(密码加密): %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "密码加密失败"})
		return
	}

	// 2. 生成唯一的 8 位 UserCode
	userCode, err := getUniqueUserCode()
	if err != nil {
		global.GetLog(c).Errorf("注册失败(生成UserCode): %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "系统繁忙，生成用户编码失败"})
		return
	}

	// 3. 插入数据库，第一个用户自动设置为管理员
	isAdmin := 0
	if isFirstUser {
		isAdmin = 1
		global.GetLog(c).Infof("第一个用户注册，自动设置为管理员: %s", req.Username)
	}

	_, err = global.DB.Exec(
		"INSERT INTO users (username, password, user_code, nickname, email, is_admin) VALUES (?, ?, ?, ?, ?, ?)",
		req.Username, string(hash), userCode, req.Nickname, req.Email, isAdmin,
	)

	if err != nil {
		global.GetLog(c).Warnf("注册失败(DB插入): %v, Username: %s", err, req.Username)
		c.JSON(500, gin.H{"code": 500, "msg": "注册失败，用户名可能已存在"})
		return
	}

	global.GetLog(c).Infof("新用户注册成功: %s (Code: %s)", req.Username, userCode)
	c.JSON(200, gin.H{"code": 200, "msg": "注册成功"})
}

// =======================
// 主入口：用户登录
// =======================
func UserLogin(c *gin.Context) {
	// 1. 尝试 Token 自动登录
	if tryTokenLogin(c) {
		return
	}

	// 2. 尝试 账号密码 登录
	tryPasswordLogin(c)
}

// ---------------------------------------------------------
// 逻辑拆分 A：处理 Token 登录
// ---------------------------------------------------------
func tryTokenLogin(c *gin.Context) bool {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false
	}
	tokenString := parts[1]

	// 1. 查白名单
	exists, _ := global.VerifyToken(tokenString)
	if !exists {
		return false
	}

	// 2. 解析 Token
	token, err := jwt.ParseWithClaims(tokenString, &middleware.MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return middleware.JwtSecret, nil
	})
	if err != nil || !token.Valid {
		return false
	}

	// 3. 查库获取最新信息
	claims, ok := token.Claims.(*middleware.MyClaims)
	if !ok {
		return false
	}

	var user model.DbUser
	err = global.DB.QueryRow(
		"SELECT id, username, password, user_code, nickname, email, is_admin, status FROM users WHERE id = ?",
		claims.UserID,
	).Scan(&user.Id, &user.Username, &user.Password, &user.UserCode, &user.Nickname, &user.Email, &user.IsAdmin, &user.Status)

	if err != nil {
		return false
	}

	// 如果用户ID=1且不是管理员，自动升级为管理员
	if user.Id == 1 && user.IsAdmin != 1 {
		_, _ = global.DB.Exec("UPDATE users SET is_admin = 1 WHERE id = 1")
		user.IsAdmin = 1
		global.GetLog(c).Infof("用户ID=1自动升级为管理员: %s", user.Username)
	}

	// 检查用户状态：0-正常，1-禁用
	if user.Id != 1 && user.Status == 1 {

		global.GetLog(c).Warnf("Token登录失败: 用户已被禁用 (%s)", user.Username)
		c.JSON(403, gin.H{"code": 403, "msg": "该账号已被禁用，请联系管理员"})
		return true // 返回true表示已处理，不继续尝试密码登录
	}

	// 更新最后登录时间
	_, _ = global.DB.Exec("UPDATE users SET last_login_time = CURRENT_TIMESTAMP WHERE id = ?", user.Id)

	global.GetLog(c).Infof("用户[%s] Token自动登录成功", user.Username)

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "自动登录成功",
		"data": gin.H{
			"token":           tokenString,
			"user_code":       user.UserCode,
			"username":        user.Username,
			"nickname":        user.Nickname.String,
			"email":           user.Email.String,
			"is_admin":        user.IsAdmin,
			"need_change_pwd": false,
			"oss_url":         global.GetOssUrl(), // OSS 地址（未配置时为空）
		},
	})
	return true
}

// ---------------------------------------------------------
// 逻辑拆分 B：处理 账号密码 登录
// ---------------------------------------------------------
func tryPasswordLogin(c *gin.Context) {
	var req model.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	var user model.DbUser
	err := global.DB.QueryRow(
		"SELECT id, username, password, user_code, nickname, email, is_admin, status FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.Id, &user.Username, &user.Password, &user.UserCode, &user.Nickname, &user.Email, &user.IsAdmin, &user.Status)

	if err == sql.ErrNoRows {
		global.GetLog(c).Warnf("登录失败: 用户不存在 (%s)", req.Username)
		c.JSON(404, gin.H{"code": 404, "msg": "用户不存在"})
		return
	} else if err != nil {
		global.GetLog(c).Errorf("登录查询DB失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "数据库错误"})
		return
	}

	// 检查用户状态：0-正常，1-禁用
	if user.Status == 1 {
		global.GetLog(c).Warnf("登录失败: 用户已被禁用 (%s)", req.Username)
		c.JSON(403, gin.H{"code": 403, "msg": "该账号已被禁用，请联系管理员"})
		return
	}

	// 如果用户ID=1且不是管理员，自动升级为管理员
	if user.Id == 1 && user.IsAdmin != 1 {
		_, _ = global.DB.Exec("UPDATE users SET is_admin = 1 WHERE id = 1")
		user.IsAdmin = 1
		global.GetLog(c).Infof("用户ID=1自动升级为管理员: %s", user.Username)
	}

	// 密码逻辑
	forceChangePwd := false
	if user.Password == "" {
		// 密码为空，允许登录但强制改密
		forceChangePwd = true
		global.GetLog(c).Warnf("用户[%s] 密码为空，触发强制改密", req.Username)
	} else {
		// ★★★ 修改点：先进行后端 MD5，再和数据库的 Bcrypt Hash 比较 ★★★
		doubleMd5Pwd := md5V(req.Password)

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(doubleMd5Pwd)); err != nil {
			global.GetLog(c).Warnf("登录失败: 密码错误 (%s)", req.Username)
			c.JSON(402, gin.H{"code": 402, "msg": "密码错误"})
			return
		}
	}

	// 生成新 Token
	newToken, err := middleware.GenerateToken(user.Id, user.Username, user.UserCode)
	if err != nil {
		global.GetLog(c).Errorf("Token生成失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "Token 生成失败"})
		return
	}

	// 存入白名单
	global.SaveToken(newToken, user.UserCode)

	// 更新最后登录时间
	_, err = global.DB.Exec("UPDATE users SET last_login_time = CURRENT_TIMESTAMP WHERE id = ?", user.Id)
	if err != nil {
		global.GetLog(c).Warnf("更新登录时间失败: %v", err)
	}

	global.GetLog(c).Infof("用户[%s] 密码登录成功", req.Username)

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "登录成功",
		"data": gin.H{
			"token":           newToken,
			"user_code":       user.UserCode,
			"username":        user.Username,
			"nickname":        user.Nickname.String,
			"email":           user.Email.String,
			"is_admin":        user.IsAdmin,
			"need_change_pwd": forceChangePwd,
			"oss_url":         global.GetOssUrl(), // OSS 地址（未配置时为空）
		},
	})
}

// =======================
// 用户退出登录
// =======================
func UserLogout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(200, gin.H{"code": 200, "msg": "退出成功"})
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	global.RemoveToken(tokenString)

	userCode, _ := c.Get("userCode")
	global.GetLog(c).Infof("用户[%v] 退出登录", userCode)

	c.JSON(200, gin.H{"code": 200, "msg": "退出成功"})
}

// =======================
// 修改用户信息 / 修改密码
// =======================
func UpdateUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	var req model.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 3. 处理修改密码逻辑
	if req.NewPassword != "" {
		var dbPwd string
		err := global.DB.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&dbPwd)
		if err != nil {
			global.GetLog(c).Errorf("修改密码查询失败: %v", err)
			c.JSON(500, gin.H{"code": 500, "msg": "查询用户失败"})
			return
		}

		// 判断数据库密码是否为空
		if dbPwd == "" {
			// 情况 1: 数据库密码为空 (强制改密流程) -> 直接允许设置新密码
			global.GetLog(c).Infof("用户[%v] 初始设置密码 (原密码为空)", userID)
		} else {
			// 情况 2: 数据库密码存在 -> 必须校验旧密码
			if req.OldPassword == "" {
				c.JSON(400, gin.H{"code": 400, "msg": "请输入旧密码"})
				return
			}

			// ★★★ 修改点：旧密码验证也需要先 MD5 ★★★
			oldDoubleMd5 := md5V(req.OldPassword)

			if err := bcrypt.CompareHashAndPassword([]byte(dbPwd), []byte(oldDoubleMd5)); err != nil {
				global.GetLog(c).Warnf("修改密码失败: 旧密码错误 (UserID: %v)", userID)
				c.JSON(400, gin.H{"code": 400, "msg": "旧密码错误"})
				return
			}
		}

		// ★★★ 修改点：新密码保存前先 MD5，再 Bcrypt ★★★
		newDoubleMd5 := md5V(req.NewPassword)
		hash, _ := bcrypt.GenerateFromPassword([]byte(newDoubleMd5), bcrypt.DefaultCost)

		_, err = global.DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(hash), userID)
		if err != nil {
			global.GetLog(c).Errorf("密码更新DB失败: %v", err)
			c.JSON(500, gin.H{"code": 500, "msg": "密码更新失败"})
			return
		}
		global.GetLog(c).Infof("用户[%v] 修改密码成功", userID)
	}

	// 4. 处理修改基本信息逻辑
	if req.Nickname != "" || req.Email != "" {
		var err error
		if req.Nickname != "" && req.Email == "" {
			_, err = global.DB.Exec("UPDATE users SET nickname = ? WHERE id = ?", req.Nickname, userID)
		} else if req.Email != "" && req.Nickname == "" {
			_, err = global.DB.Exec("UPDATE users SET email = ? WHERE id = ?", req.Email, userID)
		} else if req.Nickname != "" && req.Email != "" {
			_, err = global.DB.Exec("UPDATE users SET nickname = ?, email = ? WHERE id = ?", req.Nickname, req.Email, userID)
		}

		if err != nil {
			global.GetLog(c).Errorf("用户信息更新DB失败: %v", err)
			c.JSON(500, gin.H{"code": 500, "msg": "信息更新失败"})
			return
		}
		global.GetLog(c).Infof("用户[%v] 更新资料成功", userID)
	}

	c.JSON(200, gin.H{"code": 200, "msg": "更新成功"})
}
