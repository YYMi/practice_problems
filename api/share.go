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

// =================================================================================
// 辅助函数：计算过期时间
// 输入: "7d", "1w", "1m", "1y", "forever"
// 输出: 具体的时间对象指针 (如果是 forever，返回 nil)
// =================================================================================
func calculateExpireTime(durationStr string) *time.Time {
	if durationStr == "forever" {
		return nil
	}
	re := regexp.MustCompile(`^(\d+)([dwmy])$`)
	matches := re.FindStringSubmatch(durationStr)
	if len(matches) != 3 {
		// 默认给7天防止错误
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

	// 1. 校验所有科目归属权
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

		// 校验目标用户是否存在 (使用 user_code)
		for _, target := range req.Targets {
			target = strings.TrimSpace(target)
			var userCount int
			err := tx.QueryRow("SELECT count(*) FROM users WHERE user_code = ?", target).Scan(&userCount)
			if err != nil || userCount == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": fmt.Sprintf("用户 %s 不存在，请检查后重试", target)})
				return
			}
		}

		count := handleDirectShareTx(tx, req, userID.(int))
		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": fmt.Sprintf("成功授权给 %d 位用户", count)})
		log.Printf("授权成功")

	} else {
		// === 生成分享码 (公开) ===
		code, err := handleCodeShareTx(tx, req, userID.(int))
		if err != nil {
			log.Println("生成分享码失败:", err)
			// 如果是自定义错误（如超过1年），返回 400
			if strings.Contains(err.Error(), "非法操作") {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成失败"})
			}
			return
		}
		tx.Commit()
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "生成成功", "data": code})
	}
}

// =================================================================================
// 事务处理：直接分享
// =================================================================================
func handleDirectShareTx(tx *sql.Tx, req model.CreateShareRequest, operatorID int) int {
	expireTimeObj := calculateExpireTime(req.Duration)

	// 转字符串存入 DB，保持格式统一
	var expireTimeStr interface{}
	if expireTimeObj == nil {
		expireTimeStr = nil
	} else {
		expireTimeStr = expireTimeObj.Format("2006-01-02 15:04:05")
	}

	successCount := 0

	// SQLite Upsert 语法 (如果是 MySQL 请自行替换 ON DUPLICATE KEY UPDATE)
	sqlStr := `
		INSERT INTO user_subjects (user_id, subject_id, status, expire_time) 
		VALUES (?, ?, 1, ?)
		ON CONFLICT(user_id, subject_id) 
		DO UPDATE SET expire_time = excluded.expire_time, status = 1
	`

	for _, targetCode := range req.Targets {
		targetCode = strings.TrimSpace(targetCode)

		// 必须先查 ID
		var realUserID int
		err := tx.QueryRow("SELECT id FROM users WHERE user_code = ?", targetCode).Scan(&realUserID)
		if err != nil {
			log.Printf("绑定失败：找不到 user_code 为 %s 的用户ID", targetCode)
			continue
		}

		for _, subID := range req.SubjectIDs {
			_, err := tx.Exec(sqlStr, realUserID, subID, expireTimeStr)
			if err != nil {
				log.Println("授权写入失败:", err)
			}
		}
		successCount++
	}
	return successCount
}

// =================================================================================
// 事务处理：生成分享码
// =================================================================================
func handleCodeShareTx(tx *sql.Tx, req model.CreateShareRequest, operatorID int) (string, error) {
	// 1. 生成随机 Code
	rand.Seed(time.Now().UnixNano())
	randomStr := fmt.Sprintf("%06X", rand.Intn(16777215))
	shareCode := "SHARE-" + randomStr

	// 2. 设定【分享码】有效期 (Code Expiration)
	durationStr := req.CodeDuration
	if durationStr == "" {
		durationStr = "3d"
	}

	// 安全拦截：解析并检查是否超过 1 年
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
	expireTimeStr := codeExpireTime.Format("2006-01-02 15:04:05")

	// 3. 插入主表
	// duration_str: 资源有效期 (给用户看的)
	// expire_time: 分享码失效时间 (给系统判断用的)
	res, err := tx.Exec(
		`INSERT INTO share_codes (code, creator_id, duration_str, expire_time) VALUES (?, ?, ?, ?)`,
		shareCode, operatorID, req.Duration, expireTimeStr,
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

// =================================================================================
// BindSubject 绑定资源 (智能去重版)
// =================================================================================
func BindSubject(c *gin.Context) {
	var req model.BindShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	currentUserID, _ := c.Get("userID")

	// 1. 查主表信息
	var shareCodeID int
	var resourceDurationStr string
	var creatorID int
	var expireTimeStr string
	var currentUsedCount int

	// ★★★ 修改 SQL：多查一个 creator_id ★★★
	err := global.DB.QueryRow(
		"SELECT id, creator_id, duration_str, expire_time, used_count FROM share_codes WHERE code = ? AND status = 1",
		req.Code,
	).Scan(&shareCodeID, &creatorID, &resourceDurationStr, &expireTimeStr, &currentUsedCount)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "分享码无效或已失效"})
		return
	} else if err != nil {
		log.Println("查询分享码失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}

	// ★★★ 新增逻辑：禁止绑定自己创建的码 ★★★
	// 注意：currentUserID 从 gin.Context 获取可能是 int 或 float64，建议强转 int 比较
	if creatorID == currentUserID.(int) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "您是该分享码的创建者，无需绑定"})
		return
	}

	// 2. 校验分享码有效期
	var codeExpireTime time.Time
	var parseErr error
	codeExpireTime, parseErr = time.ParseInLocation("2006-01-02 15:04:05", expireTimeStr, time.Local)
	if parseErr != nil {
		codeExpireTime, _ = time.Parse(time.RFC3339, expireTimeStr) // 兼容旧格式
	}

	if time.Now().After(codeExpireTime) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "该分享码已失效 (超过有效期)"})
		return
	}

	// 3. 获取该码包含的科目 ID
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
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的分享码：未包含任何科目"})
		return
	}

	// 4. 开启事务
	tx, err := global.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "事务开启失败"})
		return
	}
	defer tx.Rollback()

	// =========== 统计逻辑 (保持不变) ===========
	usageSQL := `INSERT OR IGNORE INTO share_code_usage (share_code_id, user_id) VALUES (?, ?)`
	res, err := tx.Exec(usageSQL, shareCodeID, currentUserID)
	if err == nil {
		affected, _ := res.RowsAffected()
		if affected > 0 {
			tx.Exec("UPDATE share_codes SET used_count = used_count + 1 WHERE id = ?", shareCodeID)
			currentUsedCount++
		}
	}

	// =========== 核心优化：计算资源有效期 & 智能绑定 ===========

	// 计算新的过期时间
	userResourceExpireObj := calculateExpireTime(resourceDurationStr)
	var userResourceExpireStr interface{}
	if userResourceExpireObj == nil {
		userResourceExpireStr = nil
	} else {
		userResourceExpireStr = userResourceExpireObj.Format("2006-01-02 15:04:05")
	}

	// 绑定 SQL
	bindSQL := `
		INSERT INTO user_subjects (user_id, subject_id, status, expire_time, source_share_code_id) 
		VALUES (?, ?, 1, ?, ?)
		ON CONFLICT(user_id, subject_id) 
		DO UPDATE SET 
			expire_time = excluded.expire_time, 
			status = 1,
			source_share_code_id = excluded.source_share_code_id
	`

	successCount := 0
	skippedCount := 0

	for _, sid := range subjectIDs {
		// ★★★ 性能优化：先检查该用户是否已经拥有该科目的【有效】权限 ★★★
		// 条件：status=1 且 (expire_time 是 NULL 或者是 未来时间)
		// 注意：SQLite 的 datetime('now', 'localtime') 用于比较字符串时间
		checkSQL := `
			SELECT id FROM user_subjects 
			WHERE user_id = ? AND subject_id = ? AND status = 1 
			AND (expire_time IS NULL OR expire_time > datetime('now', 'localtime'))
		`
		var existingID int
		err := tx.QueryRow(checkSQL, currentUserID, sid).Scan(&existingID)

		if err == nil {
			// 查到了！说明用户已经有这个科目且没过期。
			// 按照需求：不允许重复绑定（跳过），除非已过期。
			skippedCount++
			continue
		}

		// 没查到（不存在 或 已过期 或 status=0），执行绑定/续期
		_, err = tx.Exec(bindSQL, currentUserID, sid, userResourceExpireStr, shareCodeID)
		if err != nil {
			log.Println("绑定单个科目失败:", err)
			continue
		}
		successCount++
	}

	tx.Commit()

	// 构造返回消息
	msg := ""
	if successCount > 0 {
		msg = fmt.Sprintf("成功绑定 %d 个新科目！", successCount)
		if skippedCount > 0 {
			msg += fmt.Sprintf(" (另有 %d 个科目您已拥有且未过期，已跳过)", skippedCount)
		}
	} else {
		if skippedCount > 0 {
			msg = "您已拥有该分享码包含的所有科目，且均在有效期内，无需重复绑定。"
		} else {
			msg = "绑定操作完成，但没有科目发生变更。"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  msg,
		"data": gin.H{
			"success_count": successCount,
			"skipped_count": skippedCount,
			"total_users":   currentUsedCount,
		},
	})
}

// =================================================================================
// GetMyShareCodes 获取我创建的分享码列表
// =================================================================================
func GetMyShareCodes(c *gin.Context) {
	userID, _ := c.Get("userID")

	// ★★★ 修改点：增加 AND sc.status = 1 ★★★
	sqlStr := `
		SELECT 
			sc.id, sc.code, sc.duration_str, sc.expire_time, sc.used_count, sc.create_time,
			(SELECT COUNT(*) FROM share_code_subjects WHERE share_code_id = sc.id) as subject_count
		FROM share_codes sc
		WHERE sc.creator_id = ? AND sc.status = 1
		ORDER BY sc.create_time DESC
	`

	rows, err := global.DB.Query(sqlStr, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	var list []gin.H
	for rows.Next() {
		var id, usedCount, subjectCount int
		var code, durationStr, expireTimeStr, createTimeStr string

		err = rows.Scan(&id, &code, &durationStr, &expireTimeStr, &usedCount, &createTimeStr, &subjectCount)
		if err != nil {
			continue
		}

		// 格式化时间显示，去掉 T 和 Z
		expireTimeStr = strings.Replace(expireTimeStr, "T", " ", 1)
		expireTimeStr = strings.Split(expireTimeStr, "+")[0]
		expireTimeStr = strings.TrimSuffix(expireTimeStr, "Z")

		// 判断过期状态
		status := "active"
		// 尝试解析多种格式
		expireTime, err := time.ParseInLocation("2006-01-02 15:04:05", expireTimeStr, time.Local)
		if err != nil {
			expireTime, _ = time.Parse(time.RFC3339, expireTimeStr)
		}

		if time.Now().After(expireTime) {
			status = "expired"
		}

		list = append(list, gin.H{
			"id":            id,
			"code":          code,
			"resource_time": durationStr,
			"expire_time":   expireTimeStr,
			"used_count":    usedCount,
			"subject_count": subjectCount,
			"create_time":   createTimeStr,
			"status":        status,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": list})
}

func DeleteShareCode(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	userID, _ := c.Get("userID")

	// ★★★ 修改点：改为软删除 (status = 0) ★★★
	res, err := global.DB.Exec(
		"UPDATE share_codes SET status = 0 WHERE id = ? AND creator_id = ?",
		id, userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "分享码不存在或无权删除"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// =================================================================================
// UpdateShareCode 更新分享码
// =================================================================================
func UpdateShareCode(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	userID, _ := c.Get("userID")

	var req struct {
		NewExpireDate string `json:"new_expire_date"`
		NewDuration   string `json:"new_duration"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 动态构建 SQL
	var args []interface{}
	sqlStr := "UPDATE share_codes SET "

	// 1. 校验并处理截止时间
	if req.NewExpireDate != "" {
		newTime, err := time.ParseInLocation("2006-01-02 15:04:05", req.NewExpireDate, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "截止时间格式错误"})
			return
		}

		// 查出该码的创建时间 (且必须是未删除的 status=1)
		var createTime time.Time
		err = global.DB.QueryRow("SELECT create_time FROM share_codes WHERE id = ? AND status = 1", id).Scan(&createTime)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "找不到该分享码或已删除"})
			return
		}

		// 限制：最大有效期 = 创建时间 + 1年
		limitTime := createTime.AddDate(1, 0, 0)
		if newTime.After(limitTime) {
			limitStr := limitTime.Format("2006-01-02 15:04:05")
			msg := fmt.Sprintf("非法操作：该码最晚有效期只能到 %s (创建后1年内)", limitStr)
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": msg})
			return
		}
		if newTime.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "截止时间不能早于当前时间"})
			return
		}

		sqlStr += "expire_time = ?, "
		args = append(args, req.NewExpireDate)
	}

	// 2. 处理资源有效期
	if req.NewDuration != "" {
		sqlStr += "duration_str = ?, "
		args = append(args, req.NewDuration)
	}

	if len(args) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "无变更"})
		return
	}

	sqlStr = strings.TrimSuffix(sqlStr, ", ")

	// ★★★ 核心权限控制：必须匹配 creator_id ★★★
	sqlStr += " WHERE id = ? AND creator_id = ? AND status = 1"
	args = append(args, id, userID)

	res, err := global.DB.Exec(sqlStr, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		// 走到这里说明：要么码不存在，要么 id 存在但 creator_id 不匹配（即不是你创建的）
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "分享码不存在或您无权修改"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}
