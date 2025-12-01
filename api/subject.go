package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"practice_problems/global"
	"practice_problems/model"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// =================================================================================
// GetSubjectList 获取科目列表 (带作者信息版)
// =================================================================================
func GetSubjectList(c *gin.Context) {
	userID, exists := c.Get("userID")
	userCode, _ := c.Get("userCode")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// ★★★ 修改 SQL：LEFT JOIN users 表获取作者信息 ★★★
	sqlStr := `
		SELECT 
			s.id, s.name, s.status, s.creator_code, s.create_time, s.update_time,
			u.email, u.nickname
		FROM subjects s 
		JOIN user_subjects us ON s.id = us.subject_id 
		LEFT JOIN users u ON s.creator_code = u.user_code  -- 关联查作者
		WHERE us.user_id = ? 
		  AND s.status = 1
		  AND us.status = 1
		  AND (
		      s.creator_code = ? OR
		      (us.expire_time IS NULL OR us.expire_time > datetime('now', 'localtime'))
		  )
		ORDER BY s.create_time DESC
	`

	rows, err := global.DB.Query(sqlStr, userID, userCode)
	if err != nil {
		log.Println("查询失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "服务器内部错误"})
		return
	}
	defer rows.Close()

	var list []gin.H // 使用 map 切片更灵活

	for rows.Next() {
		var id int
		var name, statusStr, creatorCode, createTime, updateTime string
		var creatorEmail, creatorNick sql.NullString // 可能为空

		err := rows.Scan(&id, &name, &statusStr, &creatorCode, &createTime, &updateTime, &creatorEmail, &creatorNick)
		if err != nil {
			continue
		}

		// 构造返回数据
		list = append(list, gin.H{
			"id":          id,
			"name":        name,
			"status":      1, // 肯定是1
			"creatorCode": creatorCode,
			"createTime":  createTime,
			"updateTime":  updateTime,
			// ★★★ 把作者信息塞进去 ★★★
			"creatorEmail": creatorEmail.String,
			"creatorName":  creatorNick.String,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": list})
}

// =================================================================================
// GetSubjectDetail 获取单条详情 (带作者信息版)
// =================================================================================
func GetSubjectDetail(c *gin.Context) {
	idStr := c.Param("id")
	subjectID, _ := strconv.Atoi(idStr)
	userID, _ := c.Get("userID")
	userCode, _ := c.Get("userCode")

	// ★★★ 修改 SQL：同样加上 LEFT JOIN users ★★★
	sqlStr := `
		SELECT 
			s.id, s.name, s.status, s.creator_code, s.create_time, s.update_time,
			u.email, u.nickname
		FROM subjects s 
		JOIN user_subjects us ON s.id = us.subject_id 
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE s.id = ? 
		  AND us.user_id = ?
		  AND us.status = 1
		  AND (
		      s.creator_code = ? OR
		      (us.expire_time IS NULL OR us.expire_time > datetime('now', 'localtime'))
		  )
	`

	var id int
	var name, statusStr, creatorCode, createTime, updateTime string
	var creatorEmail, creatorNick sql.NullString

	err := global.DB.QueryRow(sqlStr, subjectID, userID, userCode).Scan(
		&id, &name, &statusStr, &creatorCode, &createTime, &updateTime, &creatorEmail, &creatorNick,
	)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "未找到该科目或无权查看"})
		return
	}

	data := gin.H{
		"id":           id,
		"name":         name,
		"status":       1,
		"creatorCode":  creatorCode,
		"createTime":   createTime,
		"updateTime":   updateTime,
		"creatorEmail": creatorEmail.String,
		"creatorName":  creatorNick.String,
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": data})
}

// =================================================================================
// CreateSubject 创建科目
// 逻辑：事务 -> 1. 插入subjects表  2. 插入user_subjects表(绑定自己，且永久有效)
// ★★★ 修改：插入 user_subjects 时显式设置 status=1, expire_time=NULL ★★★
// =================================================================================
func CreateSubject(c *gin.Context) {
	var req model.CreateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userCode, _ := c.Get("userCode")

	tx, err := global.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "开启事务失败"})
		return
	}

	// 1. 插入 subjects 表
	insertSubjectSQL := "INSERT INTO subjects (name, status, creator_code) VALUES (?, ?, ?)"
	result, err := tx.Exec(insertSubjectSQL, req.Name, req.Status, userCode)
	if err != nil {
		tx.Rollback()
		log.Println("创建科目失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建科目失败"})
		return
	}

	newSubjectID, _ := result.LastInsertId()

	// 2. 插入 user_subjects 表
	// ★★★ 关键点：status = 1 (有效), expire_time = NULL (永久) ★★★
	insertRelationSQL := "INSERT INTO user_subjects (user_id, subject_id, status, expire_time) VALUES (?, ?, 1, NULL)"
	_, err = tx.Exec(insertRelationSQL, userID, newSubjectID)
	if err != nil {
		tx.Rollback()
		log.Println("绑定用户失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误：绑定失败"})
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println("事务提交失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{"id": newSubjectID}})
}

// UpdateSubject 更新科目 (保持不变)
func UpdateSubject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ID参数错误"})
		return
	}

	currentUserCode, _ := c.Get("userCode")

	var req model.UpdateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	var creatorCode string
	var creatorEmail sql.NullString
	var creatorName string

	checkSQL := `
		SELECT s.creator_code, u.email, IFNULL(u.nickname, u.username)
		FROM subjects s
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE s.id = ?
	`
	err = global.DB.QueryRow(checkSQL, id).Scan(&creatorCode, &creatorEmail, &creatorName)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该科目"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统繁忙"})
		return
	}

	currentUserCodeStr, ok := currentUserCode.(string)
	if !ok || creatorCode != currentUserCodeStr {
		contactInfo := creatorName
		if creatorEmail.Valid && creatorEmail.String != "" {
			contactInfo = creatorEmail.String
		} else {
			contactInfo = creatorName + " (未设置邮箱)"
		}
		msg := "修改失败：您不是创建者，请联系作者邮箱: " + contactInfo + " 进行修改"
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": msg})
		return
	}

	updateSQL := "UPDATE subjects SET name = ?, status = ?, update_time = ? WHERE id = ?"
	nowTime := time.Now().Format("2006-01-02 15:04:05")

	_, err = global.DB.Exec(updateSQL, req.Name, req.Status, nowTime, id)
	if err != nil {
		log.Println("更新失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// DeleteSubject 删除科目 (保持不变)
func DeleteSubject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ID参数错误"})
		return
	}

	currentUserCode, _ := c.Get("userCode")

	var creatorCode string
	var creatorEmail sql.NullString
	var creatorName string

	checkSQL := `
		SELECT s.creator_code, u.email, IFNULL(u.nickname, u.username)
		FROM subjects s
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE s.id = ?
	`
	err = global.DB.QueryRow(checkSQL, id).Scan(&creatorCode, &creatorEmail, &creatorName)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该科目"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统繁忙"})
		return
	}

	currentUserCodeStr, ok := currentUserCode.(string)
	if !ok || creatorCode != currentUserCodeStr {
		contactInfo := creatorName
		if creatorEmail.Valid && creatorEmail.String != "" {
			contactInfo = creatorEmail.String
		} else {
			contactInfo = creatorName + " (未设置邮箱)"
		}
		msg := "删除失败：请联系作者邮箱: " + contactInfo + " 进行操作"
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": msg})
		return
	}

	updateSQL := "UPDATE subjects SET status = 0, update_time = ? WHERE id = ?"
	nowTime := time.Now().Format("2006-01-02 15:04:05")

	_, err = global.DB.Exec(updateSQL, nowTime, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// =================================================================================
// UpdateSubjectAuth 更新某用户的授权信息 (修改有效期)
// =================================================================================
func UpdateSubjectAuth(c *gin.Context) {
	// 这里的 id 是 user_subjects 表的主键 ID
	idStr := c.Param("id")
	// ... (鉴权逻辑：你需要查一下这个记录对应的 subject_id 是不是当前登录用户创建的，略) ...

	var req struct {
		NewExpireDate string `json:"new_expire_date"` // "2025-xx-xx" or "forever"
	}
	c.ShouldBindJSON(&req)

	var expireVal interface{}
	if req.NewExpireDate == "forever" || req.NewExpireDate == "" {
		expireVal = nil
	} else {
		// 校验格式...
		expireVal = req.NewExpireDate
	}

	_, err := global.DB.Exec("UPDATE user_subjects SET expire_time = ? WHERE id = ?", expireVal, idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// =================================================================================
// RemoveSubjectAuth 解除授权 (踢人)
// =================================================================================
func RemoveSubjectAuth(c *gin.Context) {
	idStr := c.Param("id")
	// ... (鉴权逻辑，略) ...

	// 软删除 or 硬删除
	_, err := global.DB.Exec("UPDATE user_subjects SET status = 0 WHERE id = ?", idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已解除授权"})
}

// =================================================================================
// GetSubjectAuthorizedUsers 获取列表 (排除作者自己)
// =================================================================================
func GetSubjectAuthorizedUsers(c *gin.Context) {
	subjectID, _ := strconv.Atoi(c.Param("id"))
	userID, _ := c.Get("userID") // 当前登录用户 ID (即作者 ID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	searchCode := c.DefaultQuery("user_code", "")
	offset := (page - 1) * pageSize

	// 1. 鉴权 (保持不变)
	var count int
	global.DB.QueryRow("SELECT count(*) FROM subjects WHERE id = ? AND creator_code = (SELECT user_code FROM users WHERE id = ?)", subjectID, userID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作"})
		return
	}

	// 2. 构建动态 SQL
	// ★★★ 核心修改：增加 AND us.user_id != ? 排除自己 ★★★
	baseSQL := `
		FROM user_subjects us 
		LEFT JOIN users u ON us.user_id = u.id 
		WHERE us.subject_id = ? 
		  AND us.status = 1 
		  AND us.user_id != ? 
	`

	// 参数顺序：subjectID, userID
	var args []interface{}
	args = append(args, subjectID, userID)

	// 搜索逻辑
	if searchCode != "" {
		baseSQL += " AND u.user_code LIKE ?"
		args = append(args, "%"+searchCode+"%")
	}

	// 3. 查询总数
	var total int
	countQuery := "SELECT count(*)" + baseSQL
	global.DB.QueryRow(countQuery, args...).Scan(&total)

	// 4. 查询列表
	listQuery := `
		SELECT us.id, us.user_id, us.expire_time, us.create_time,
			   u.user_code, u.username, u.nickname, u.email
	` + baseSQL + " ORDER BY us.create_time DESC LIMIT ? OFFSET ?"

	// 追加分页参数
	args = append(args, pageSize, offset)

	rows, err := global.DB.Query(listQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	var list []gin.H
	for rows.Next() {
		// ... (内部扫描和格式化逻辑保持不变) ...
		var id, uid int
		var expireTimeStr sql.NullString
		var bindTimeStr, uCode, uName, uNick, uEmail string
		rows.Scan(&id, &uid, &expireTimeStr, &bindTimeStr, &uCode, &uName, &uNick, &uEmail)

		expireDisplay := "永久"
		if expireTimeStr.Valid && expireTimeStr.String != "" {
			tStr := strings.Replace(expireTimeStr.String, "T", " ", 1)
			expireDisplay = strings.Split(tStr, "+")[0]
			expireDisplay = strings.TrimSuffix(expireDisplay, "Z")
		}

		bindTimeStr = strings.Replace(bindTimeStr, "T", " ", 1)
		bindTimeStr = strings.Split(bindTimeStr, "+")[0]
		bindTimeStr = strings.TrimSuffix(bindTimeStr, "Z")

		list = append(list, gin.H{
			"id":          id,
			"user_code":   uCode,
			"nickname":    uNick,
			"email":       uEmail,
			"bind_time":   bindTimeStr,
			"expire_time": expireDisplay,
			"raw_expire":  expireTimeStr.String,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": gin.H{"list": list, "total": total}})
}

// =================================================================================
// BatchUpdateAuth 批量更新有效期
// =================================================================================
func BatchUpdateAuth(c *gin.Context) {
	var req struct {
		Ids           []int  `json:"ids"`
		NewExpireDate string `json:"new_expire_date"` // "2025-xx-xx" or "forever"
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	var expireVal interface{}
	if req.NewExpireDate == "forever" || req.NewExpireDate == "" {
		expireVal = nil
	} else {
		expireVal = req.NewExpireDate
	}

	// 构建 IN (?,?,?)
	query := fmt.Sprintf("UPDATE user_subjects SET expire_time = ? WHERE id IN (%s)",
		strings.Trim(strings.Repeat("?,", len(req.Ids)), ","))

	args := []interface{}{expireVal}
	for _, id := range req.Ids {
		args = append(args, id)
	}

	_, err := global.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "批量更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": fmt.Sprintf("成功更新 %d 条记录", len(req.Ids))})
}

// =================================================================================
// BatchRemoveAuth 批量移除授权
// =================================================================================
func BatchRemoveAuth(c *gin.Context) {
	var req struct {
		Ids []int `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	query := fmt.Sprintf("UPDATE user_subjects SET status = 0 WHERE id IN (%s)",
		strings.Trim(strings.Repeat("?,", len(req.Ids)), ","))

	args := []interface{}{}
	for _, id := range req.Ids {
		args = append(args, id)
	}

	_, err := global.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "批量删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": fmt.Sprintf("成功移除 %d 位用户", len(req.Ids))})
}
