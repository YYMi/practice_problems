package api

import (
	"database/sql"
	"log"
	"net/http"
	"practice_problems/global"
	"practice_problems/model"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// =================================================================================
// GetSubjectList 获取科目列表
// =================================================================================
func GetSubjectList(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// ★★★ 修改：只查 subjects 表，不需要 JOIN users ★★★
	sqlStr := `
		SELECT s.id, s.name, s.status, s.creator_code, s.create_time, s.update_time 
		FROM subjects s 
		JOIN user_subjects us ON s.id = us.subject_id 
		WHERE us.user_id = ? 
		  AND s.status = 1
		  AND us.status = 1
		  AND (us.expire_time IS NULL OR us.expire_time > datetime('now', 'localtime'))
		ORDER BY s.create_time DESC
	`
	// 注意：MySQL 请用 NOW() 替换 datetime(...)

	rows, err := global.DB.Query(sqlStr, userID)
	if err != nil {
		log.Println("查询数据库失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "服务器内部错误"})
		return
	}
	defer rows.Close()

	subjects := make([]model.Subject, 0)

	for rows.Next() {
		var s model.Subject
		// ★★★ 只需要扫描基础字段，CreatorCode 本身就在 subjects 表里
		err := rows.Scan(&s.ID, &s.Name, &s.Status, &s.CreatorCode, &s.CreateTime, &s.UpdateTime)
		if err != nil {
			log.Println("数据扫描失败:", err)
			continue
		}
		subjects = append(subjects, s)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": subjects})
}

// =================================================================================
// GetSubjectDetail 获取单条详情
// =================================================================================
func GetSubjectDetail(c *gin.Context) {
	idStr := c.Param("id")
	subjectID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ID参数错误"})
		return
	}

	userID, _ := c.Get("userID")

	var s model.Subject
	// ★★★ 修改：只查 subjects 表，不需要 JOIN users ★★★
	sqlStr := `
		SELECT s.id, s.name, s.status, s.creator_code, s.create_time, s.update_time 
		FROM subjects s 
		JOIN user_subjects us ON s.id = us.subject_id 
		WHERE s.id = ? 
		  AND us.user_id = ?
		  AND us.status = 1
		  AND (us.expire_time IS NULL OR us.expire_time > datetime('now', 'localtime'))
	`

	err = global.DB.QueryRow(sqlStr, subjectID, userID).Scan(
		&s.ID, &s.Name, &s.Status, &s.CreatorCode, &s.CreateTime, &s.UpdateTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "未找到该科目或您的授权已过期"})
		} else {
			log.Println("查询详情失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": s})
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
