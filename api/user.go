package api

import (
	"database/sql"
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

// 生成 8 位随机数字字符串 (00000000 - 99999999)
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

	// 1. 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
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

	// 3. 插入数据库
	_, err = global.DB.Exec(
		"INSERT INTO users (username, password, user_code, nickname, email) VALUES (?, ?, ?, ?, ?)",
		req.Username, string(hash), userCode, req.Nickname, req.Email,
	)

	if err != nil {
		// 这里的错误通常是 Username 重复
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
		"SELECT id, username, password, user_code, nickname, email FROM users WHERE id = ?",
		claims.UserID,
	).Scan(&user.Id, &user.Username, &user.Password, &user.UserCode, &user.Nickname, &user.Email)

	if err != nil {
		return false
	}

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
			"need_change_pwd": false,
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
		"SELECT id, username, password, user_code, nickname, email FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.Id, &user.Username, &user.Password, &user.UserCode, &user.Nickname, &user.Email)

	if err == sql.ErrNoRows {
		global.GetLog(c).Warnf("登录失败: 用户不存在 (%s)", req.Username)
		c.JSON(404, gin.H{"code": 404, "msg": "用户不存在"})
		return
	} else if err != nil {
		global.GetLog(c).Errorf("登录查询DB失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "数据库错误"})
		return
	}

	// 密码逻辑
	forceChangePwd := false
	if user.Password == "" {
		forceChangePwd = true
		global.GetLog(c).Warnf("用户[%s] 密码为空，触发强制改密", req.Username)
	} else {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
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
			"need_change_pwd": forceChangePwd,
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

	// 从后端内存中删除 Token
	global.RemoveToken(tokenString)

	// 尝试获取用户信息打个日志，取不到也无所谓
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

		if dbPwd != "" {
			if req.OldPassword == "" {
				c.JSON(400, gin.H{"code": 400, "msg": "请输入旧密码"})
				return
			}
			if err := bcrypt.CompareHashAndPassword([]byte(dbPwd), []byte(req.OldPassword)); err != nil {
				global.GetLog(c).Warnf("修改密码失败: 旧密码错误 (UserID: %v)", userID)
				c.JSON(400, gin.H{"code": 400, "msg": "旧密码错误"})
				return
			}
		}

		hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
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
