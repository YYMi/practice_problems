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
// 逻辑：生成 -> 查库 -> 如果存在就重试 -> 直到唯一
func getUniqueUserCode() (string, error) {
	maxRetries := 10 // 防止极端情况下的死循环
	for i := 0; i < maxRetries; i++ {
		code := generateRandomCode()

		// 查询数据库是否存在
		var exists int
		err := global.DB.QueryRow("SELECT 1 FROM users WHERE user_code = ?", code).Scan(&exists)

		if err == sql.ErrNoRows {
			// 找不到记录，说明这个 code 是唯一的，可以用！
			return code, nil
		} else if err != nil {
			// 数据库查询出错
			return "", err
		}
		// 如果 err == nil，说明查到了(exists=1)，也就是重复了，继续下一次循环
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
		c.JSON(500, gin.H{"code": 500, "msg": "密码加密失败"})
		return
	}

	// 2. 生成唯一的 8 位 UserCode
	userCode, err := getUniqueUserCode()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "系统繁忙，生成用户编码失败"})
		return
	}

	// 3. 插入数据库
	_, err = global.DB.Exec(
		"INSERT INTO users (username, password, user_code, nickname, email) VALUES (?, ?, ?, ?, ?)",
		req.Username, string(hash), userCode, req.Nickname, req.Email,
	)

	if err != nil {
		// 这里的错误通常是 Username 重复（因为 user_code 已经检查过了）
		c.JSON(500, gin.H{"code": 500, "msg": "注册失败，用户名可能已存在"})
		return
	}

	c.JSON(200, gin.H{"code": 200, "msg": "注册成功"})
}

// =======================
// 主入口：用户登录
// =======================
func UserLogin(c *gin.Context) {
	// 1. 尝试 Token 自动登录
	// 如果 Header 里有 Token，且验证通过，直接返回，不再走下面的逻辑
	if tryTokenLogin(c) {
		return
	}

	// 2. 尝试 账号密码 登录
	// 如果上面没通过（没传Token或无效），走传统的账号密码流程
	tryPasswordLogin(c)
}

// ---------------------------------------------------------
// 逻辑拆分 A：处理 Token 登录
// 返回 bool 表示是否处理成功 (true=成功响应, false=继续走密码登录)
// ---------------------------------------------------------
func tryTokenLogin(c *gin.Context) bool {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}

	// 格式校验 "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false
	}
	tokenString := parts[1]

	// 1. 查白名单
	exists, _ := global.VerifyToken(tokenString)
	if !exists {
		return false // Token 不在白名单，视为无效，转去尝试密码登录
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

	// 4. 成功！直接返回
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "自动登录成功",
		"data": gin.H{
			"token":           tokenString, // 原样返回旧 Token
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
	// 注意：这里不能用 ShouldBindJSON，因为它会消耗掉 Body 流
	// 如果 tryTokenLogin 里没读 Body 没事，但为了保险，这里是最后的兜底
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
		c.JSON(404, gin.H{"code": 404, "msg": "用户不存在"})
		return
	} else if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "数据库错误"})
		return
	}

	// 密码逻辑
	forceChangePwd := false
	if user.Password == "" {
		forceChangePwd = true
	} else {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			c.JSON(402, gin.H{"code": 402, "msg": "密码错误"})
			return
		}
	}

	// 生成新 Token
	newToken, err := middleware.GenerateToken(user.Id, user.Username, user.UserCode)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "Token 生成失败"})
		return
	}

	// 存入白名单
	global.SaveToken(newToken, user.UserCode)

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

	// ==========================================
	// 🔥 核心逻辑：从后端内存中删除 Token
	// ==========================================
	global.RemoveToken(tokenString)
	// ==========================================

	c.JSON(200, gin.H{"code": 200, "msg": "退出成功"})
}

// =======================
// 修改用户信息 / 修改密码
// =======================
func UpdateUser(c *gin.Context) {
	// 1. 从 JWT 中间件获取当前用户ID
	// (因为经过了中间件，所以 c.Get("userID") 一定有值)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 2. 绑定请求参数
	var req model.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 3. 处理修改密码逻辑
	if req.NewPassword != "" {
		// 先查询当前数据库里的旧密码
		var dbPwd string
		err := global.DB.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&dbPwd)
		if err != nil {
			c.JSON(500, gin.H{"code": 500, "msg": "查询用户失败"})
			return
		}

		// 只有当数据库里的密码不为空时，才校验旧密码
		// (如果数据库密码为空，说明是初始状态强制改密，允许直接设置新密码)
		if dbPwd != "" {
			if req.OldPassword == "" {
				c.JSON(400, gin.H{"code": 400, "msg": "请输入旧密码"})
				return
			}
			if err := bcrypt.CompareHashAndPassword([]byte(dbPwd), []byte(req.OldPassword)); err != nil {
				c.JSON(400, gin.H{"code": 400, "msg": "旧密码错误"})
				return
			}
		}

		// 加密新密码并更新
		hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		_, err = global.DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(hash), userID)
		if err != nil {
			c.JSON(500, gin.H{"code": 500, "msg": "密码更新失败"})
			return
		}
	}

	// 4. 处理修改基本信息逻辑 (昵称、邮箱)
	// 只有当前端传了这些字段且不为空时才更新
	if req.Nickname != "" || req.Email != "" {
		// 注意：这里做一个简单的处理，实际场景可能需要更灵活的动态 SQL 构建
		// 这里假设前端如果想修改，就会传值；不想修改的字段不要传空字符串覆盖

		// 如果只想改昵称
		if req.Nickname != "" && req.Email == "" {
			_, err := global.DB.Exec("UPDATE users SET nickname = ? WHERE id = ?", req.Nickname, userID)
			if err != nil {
				c.JSON(500, gin.H{"code": 500, "msg": "昵称更新失败"})
				return
			}
		}

		// 如果只想改邮箱
		if req.Email != "" && req.Nickname == "" {
			_, err := global.DB.Exec("UPDATE users SET email = ? WHERE id = ?", req.Email, userID)
			if err != nil {
				c.JSON(500, gin.H{"code": 500, "msg": "邮箱更新失败"})
				return
			}
		}

		// 如果两个都改
		if req.Nickname != "" && req.Email != "" {
			_, err := global.DB.Exec("UPDATE users SET nickname = ?, email = ? WHERE id = ?", req.Nickname, req.Email, userID)
			if err != nil {
				c.JSON(500, gin.H{"code": 500, "msg": "信息更新失败"})
				return
			}
		}
	}

	c.JSON(200, gin.H{"code": 200, "msg": "更新成功"})
}
