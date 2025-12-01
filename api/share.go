package api

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"practice_problems/global"
	"practice_problems/model"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 辅助函数：计算过期时间 (保持不变)
func calculateExpireTime(durationStr string) *time.Time {
	if durationStr == "forever" {
		return nil
	}
	re := regexp.MustCompile(`^(\d+)([dwmy])$`)
	matches := re.FindStringSubmatch(durationStr)
	if len(matches) != 3 {
		t := time.Now().AddDate(0, 0, 7)
		return &t
	}
	num, _ := strconv.Atoi(matches[1])
	unit := matches[2]
	now := time.Now()
	var expireTime time.Time
	switch unit {
	case "d":
		expireTime = now.AddDate(0, 0, num)
	case "w":
		expireTime = now.AddDate(0, 0, num*7)
	case "m":
		expireTime = now.AddDate(0, num, 0)
	case "y":
		expireTime = now.AddDate(num, 0, 0)
	default:
		expireTime = now.AddDate(0, 0, 7)
	}
	return &expireTime
}

// =================================================================================
// CreateShare 创建分享 (支持多科目)
// =================================================================================
func CreateShare(c *gin.Context) {
	var req model.CreateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	if len(req.SubjectIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请至少选择一个科目"})
		return
	}

	userID, _ := c.Get("userID")     // int
	userCode, _ := c.Get("userCode") // string

	// 开启事务
	tx, err := global.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "服务器错误"})
		return
	}
	defer tx.Rollback()

	// 1. 校验所有科目归属权 (保持不变)
	for _, sid := range req.SubjectIDs {
		var count int
		err := tx.QueryRow("SELECT count(*) FROM subjects WHERE id = ? AND creator_code = ?", sid, userCode).Scan(&count)
		if err != nil || count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": fmt.Sprintf("科目ID %d 不属于您或不存在", sid)})
			return
		}
	}

	// 2. 分支处理
	if req.Type == 1 {
		// === 指定用户 (私密) ===
		if len(req.Targets) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请指定目标用户"})
			return
		}

		// =========== 【新增】校验目标用户是否存在 ===========
		// 假设 req.Targets 存的是 user_id (字符串形式)
		// 如果存的是 user_code，请把 SQL 改为 WHERE user_code = ?
		for _, target := range req.Targets {
			var userCount int
			// 这里假设前端传的是 user_id，如果是 user_code 请自行调整
			err := tx.QueryRow("SELECT count(*) FROM users WHERE user_code = ?", target).Scan(&userCount)
			if err != nil || userCount == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": fmt.Sprintf("用户 %s 不存在，请检查后重试", target)})
				return
			}
		}
		// =================================================

		count := handleDirectShareTx(tx, req, userID.(int))
		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": fmt.Sprintf("成功授权给 %d 位用户", count)})
		log.Printf("授权成功 ")

	} else {
		// === 生成分享码 (公开) ===
		code, err := handleCodeShareTx(tx, req, userID.(int))
		if err != nil {
			log.Println("生成分享码失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成失败"})
			return
		}
		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "生成成功", "data": code})
	}
}

// 事务处理：直接分享
func handleDirectShareTx(tx *sql.Tx, req model.CreateShareRequest, operatorID int) int {
	expireTimeObj := calculateExpireTime(req.Duration)

	// ★★★ 修改点：转字符串 ★★★
	var expireTimeStr interface{}
	if expireTimeObj == nil {
		expireTimeStr = nil
	} else {
		expireTimeStr = expireTimeObj.Format("2006-01-02 15:04:05")
	}

	successCount := 0

	sqlStr := `
		INSERT INTO user_subjects (user_id, subject_id, status, expire_time) 
		VALUES (?, ?, 1, ?)
		ON CONFLICT(user_id, subject_id) 
		DO UPDATE SET expire_time = excluded.expire_time, status = 1
	`

	for _, targetCode := range req.Targets {
		targetCode = strings.TrimSpace(targetCode)

		var realUserID int
		err := tx.QueryRow("SELECT id FROM users WHERE user_code = ?", targetCode).Scan(&realUserID)
		if err != nil {
			continue
		}

		for _, subID := range req.SubjectIDs {
			// 传字符串
			_, err := tx.Exec(sqlStr, realUserID, subID, expireTimeStr)
			if err != nil {
				log.Println("授权写入失败:", err)
			}
		}
		successCount++
	}
	return successCount
}

// 事务处理：生成分享码
func handleCodeShareTx(tx *sql.Tx, req model.CreateShareRequest, operatorID int) (string, error) {
	// 1. 生成随机 Code
	rand.Seed(time.Now().UnixNano())
	randomStr := fmt.Sprintf("%06X", rand.Intn(16777215))
	shareCode := "SHARE-" + randomStr

	// 2. 设定【分享码】有效期
	durationStr := req.CodeDuration
	if durationStr == "" {
		durationStr = "3d"
	}

	var num int
	var unit string
	fmt.Sscanf(durationStr, "%d%s", &num, &unit)

	var checkDuration time.Duration
	switch unit {
	case "d":
		checkDuration = time.Duration(num) * 24 * time.Hour
	case "w":
		checkDuration = time.Duration(num) * 7 * 24 * time.Hour
	case "m":
		checkDuration = time.Duration(num) * 30 * 24 * time.Hour
	case "y":
		checkDuration = time.Duration(num) * 365 * 24 * time.Hour
	case "h":
		checkDuration = time.Duration(num) * time.Hour
	default:
		checkDuration = 3 * 24 * time.Hour
	}

	if checkDuration > 366*24*time.Hour {
		return "", fmt.Errorf("非法操作：分享码有效期不能超过 1 年")
	}

	// 计算过期时间点
	codeExpireTime := time.Now().Add(checkDuration)

	// ★★★ 修改点：格式化为 "2006-01-02 15:04:05" 字符串 ★★★
	// 这样数据库里存的就是干净的年月日时分秒
	expireTimeStr := codeExpireTime.Format("2006-01-02 15:04:05")

	// 3. 插入主表
	res, err := tx.Exec(
		`INSERT INTO share_codes (code, creator_id, duration_str, expire_time) VALUES (?, ?, ?, ?)`,
		shareCode, operatorID, req.Duration, expireTimeStr, // <--- 传字符串
	)
	if err != nil {
		return "", err
	}

	shareCodeID, _ := res.LastInsertId()

	// 4. 插入关联表
	insertRelSQL := `INSERT INTO share_code_subjects (share_code_id, subject_id) VALUES (?, ?)`
	for _, subID := range req.SubjectIDs {
		_, err := tx.Exec(insertRelSQL, shareCodeID, subID)
		if err != nil {
			return "", err
		}
	}

	return shareCode, nil
}

func BindSubject(c *gin.Context) {
	var req model.BindShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	currentUserID, _ := c.Get("userID")

	// 1. 查主表
	var shareCodeID int
	var resourceDurationStr string
	var expireTimeStr string // ★★★ 修改点：这里用 string 接收

	err := global.DB.QueryRow(
		"SELECT id, duration_str, expire_time FROM share_codes WHERE code = ?",
		req.Code,
	).Scan(&shareCodeID, &resourceDurationStr, &expireTimeStr)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "分享码无效"})
		return
	}

	// =========== 【逻辑 1】校验分享码是否过期 ===========

	// 尝试解析时间（兼容多种格式）
	var codeExpireTime time.Time
	var parseErr error

	// 格式 A: 我们期望的标准格式 "2006-01-02 15:04:05"
	codeExpireTime, parseErr = time.ParseInLocation("2006-01-02 15:04:05", expireTimeStr, time.Local)

	if parseErr != nil {
		// 格式 B: 尝试解析旧数据格式 (RFC3339) "2006-01-02T15:04:05Z"
		// SQLite 默认存的时间往往是这种
		codeExpireTime, parseErr = time.Parse(time.RFC3339, expireTimeStr)

		if parseErr != nil {
			// 格式 C: 有时候 SQLite 存的是 "2006-01-02T15:04:05+08:00"
			codeExpireTime, parseErr = time.Parse("2006-01-02T15:04:05Z07:00", expireTimeStr)
		}
	}

	if parseErr != nil {
		log.Printf("时间解析彻底失败. 输入字符串: %s, 错误: %v", expireTimeStr, parseErr)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统时间数据异常"})
		return
	}

	if time.Now().After(codeExpireTime) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "该分享码已失效 (超过有效期)"})
		return
	}
	// =================================================

	// 2. 查关联表
	rows, err := global.DB.Query("SELECT subject_id FROM share_code_subjects WHERE share_code_id = ?", shareCodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询关联科目失败"})
		return
	}
	defer rows.Close()

	var subjectIDs []int
	for rows.Next() {
		var sid int
		if err := rows.Scan(&sid); err == nil {
			subjectIDs = append(subjectIDs, sid)
		}
	}

	if len(subjectIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的分享码"})
		return
	}

	// 3. 执行绑定
	tx, err := global.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "事务开启失败"})
		return
	}
	defer tx.Rollback()

	// =========== 【逻辑 2】计算用户资源的过期时间 ===========
	userResourceExpireObj := calculateExpireTime(resourceDurationStr)

	// ★★★ 优化：把资源过期时间也转成字符串存入 user_subjects (可选，建议统一) ★★★
	var userResourceExpireStr interface{} // 使用 interface{} 兼容 nil
	if userResourceExpireObj == nil {
		userResourceExpireStr = nil // 永久
	} else {
		userResourceExpireStr = userResourceExpireObj.Format("2006-01-02 15:04:05")
	}

	bindSQL := `
		INSERT INTO user_subjects (user_id, subject_id, status, expire_time) 
		VALUES (?, ?, 1, ?)
		ON CONFLICT(user_id, subject_id) 
		DO UPDATE SET expire_time = excluded.expire_time, status = 1
	`

	count := 0
	for _, sid := range subjectIDs {
		_, err := tx.Exec(bindSQL, currentUserID, sid, userResourceExpireStr) // <--- 传字符串或nil
		if err != nil {
			continue
		}
		count++
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": fmt.Sprintf("成功绑定 %d 个科目！", count)})
}
