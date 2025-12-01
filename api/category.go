package api

import (
	"database/sql"
	"log"
	"net/http"
	"practice_problems/global"
	"practice_problems/model"
	"strconv"
	"strings"
	_ "time"

	"github.com/gin-gonic/gin"
)

// =================================================================================
// GetCategoryList 获取分类列表
// 逻辑：只要用户绑定了该科目(user_subjects)，就有权查看
// =================================================================================
func GetCategoryList(c *gin.Context) {
	subjectIDStr := c.Query("subject_id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 不需要查 creator_code 了
	fields := "c.id, c.subject_id, c.categorie_name, c.create_time, c.update_time, c.sort_order, c.difficulty"

	var rows *sql.Rows
	var err error

	if subjectIDStr != "" {
		// --- 场景 A：查指定科目 ---
		// 1. 权限检查：用户必须在 user_subjects 表中绑定了该科目
		var hasPerm int
		checkPermSQL := "SELECT 1 FROM user_subjects WHERE user_id = ? AND subject_id = ?"
		err := global.DB.QueryRow(checkPermSQL, userID, subjectIDStr).Scan(&hasPerm)
		if err != nil || hasPerm != 1 {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权访问该科目"})
			return
		}

		// 2. 查询分类
		sqlStr := "SELECT " + fields + " FROM knowledge_categories c WHERE c.subject_id = ? ORDER BY c.sort_order ASC, c.id DESC"
		rows, err = global.DB.Query(sqlStr, subjectIDStr)

	} else {
		// --- 场景 B：查所有可见科目的分类 ---
		sqlStr := `
			SELECT ` + fields + ` 
			FROM knowledge_categories c
			JOIN user_subjects us ON c.subject_id = us.subject_id
			WHERE us.user_id = ?
			ORDER BY c.sort_order ASC, c.id DESC
		`
		rows, err = global.DB.Query(sqlStr, userID)
	}

	if err != nil {
		log.Println("查询分类失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	list := make([]model.KnowledgeCategory, 0)
	for rows.Next() {
		var item model.KnowledgeCategory
		err := rows.Scan(
			&item.ID,
			&item.SubjectID,
			&item.CategoryName,
			&item.CreateTime,
			&item.UpdateTime,
			&item.SortOrder,
			&item.Difficulty,
		)
		if err != nil {
			continue
		}
		list = append(list, item)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": list})
}

// =================================================================================
// CreateCategory 创建分类
// 逻辑：必须是【所属科目】的创建者
// =================================================================================
func CreateCategory(c *gin.Context) {
	var req model.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	currentUserCode, _ := c.Get("userCode")

	// --- 权限检查 ---
	// 查询 subjects 表的 creator_code
	var subjectCreatorCode string
	var creatorName string
	var creatorEmail sql.NullString

	checkSQL := `
		SELECT s.creator_code, IFNULL(u.nickname, u.username), u.email
		FROM subjects s
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE s.id = ?
	`
	err := global.DB.QueryRow(checkSQL, req.SubjectID).Scan(&subjectCreatorCode, &creatorName, &creatorEmail)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "所属科目不存在"})
		return
	}

	currentUserCodeStr, _ := currentUserCode.(string)
	if subjectCreatorCode != currentUserCodeStr {
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "创建失败：您不是该科目的作者，请联系 " + contactInfo})
		return
	}

	// --- 计算排序并插入 ---
	var currentMinSort int
	sqlQueryMin := "SELECT COALESCE(MIN(sort_order), 0) FROM knowledge_categories WHERE subject_id = ?"
	_ = global.DB.QueryRow(sqlQueryMin, req.SubjectID).Scan(&currentMinSort)
	newSortOrder := currentMinSort - 1

	sqlStr := "INSERT INTO knowledge_categories (subject_id, categorie_name, sort_order, difficulty) VALUES (?, ?, ?, ?)"
	result, err := global.DB.Exec(sqlStr, req.SubjectID, req.CategoryName, newSortOrder, 0)

	if err != nil {
		log.Println("创建分类失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建失败"})
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{"id": newID}})
}

// =================================================================================
// UpdateCategory 修改分类
// 逻辑：必须是【所属科目】的创建者
// =================================================================================
func UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req model.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	currentUserCode, _ := c.Get("userCode")

	// --- 权限检查 ---
	// 通过分类找科目，再查科目的创建者
	var subjectCreatorCode string
	var creatorName string
	var creatorEmail sql.NullString

	checkSQL := `
		SELECT s.creator_code, IFNULL(u.nickname, u.username), u.email
		FROM knowledge_categories c
		JOIN subjects s ON c.subject_id = s.id
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE c.id = ?
	`
	err := global.DB.QueryRow(checkSQL, id).Scan(&subjectCreatorCode, &creatorName, &creatorEmail)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "分类不存在"})
		return
	}

	currentUserCodeStr, _ := currentUserCode.(string)
	if subjectCreatorCode != currentUserCodeStr {
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "修改失败：请联系科目作者 " + contactInfo})
		return
	}

	// --- 执行更新 ---
	query := "UPDATE knowledge_categories SET update_time = CURRENT_TIMESTAMP"
	var args []interface{}

	if req.CategoryName != "" {
		query += ", categorie_name = ?"
		args = append(args, req.CategoryName)
	}
	if req.Difficulty != nil {
		if *req.Difficulty < 0 || *req.Difficulty > 3 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "难度值非法"})
			return
		}
		query += ", difficulty = ?"
		args = append(args, *req.Difficulty)
	}

	if len(args) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "无变更"})
		return
	}

	query += " WHERE id = ?"
	args = append(args, id)

	_, err = global.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// =================================================================================
// DeleteCategory 删除分类
// 逻辑：必须是【所属科目】的创建者
// =================================================================================
func DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	currentUserCode, _ := c.Get("userCode")

	// --- 权限检查 ---
	var subjectCreatorCode string
	var creatorName string
	var creatorEmail sql.NullString

	checkSQL := `
		SELECT s.creator_code, IFNULL(u.nickname, u.username), u.email
		FROM knowledge_categories c
		JOIN subjects s ON c.subject_id = s.id
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE c.id = ?
	`
	err := global.DB.QueryRow(checkSQL, id).Scan(&subjectCreatorCode, &creatorName, &creatorEmail)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "分类不存在"})
		return
	}

	currentUserCodeStr, _ := currentUserCode.(string)
	if subjectCreatorCode != currentUserCodeStr {
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "删除失败：请联系科目作者 " + contactInfo})
		return
	}

	// --- 执行删除 ---
	sqlStr := "DELETE FROM knowledge_categories WHERE id = ?"
	result, err := global.DB.Exec(sqlStr, id)
	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY") {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败：该分类下仍有知识点"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该分类"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// UpdateCategorySortRequest 排序请求
type UpdateCategorySortRequest struct {
	Action string `json:"action" binding:"required,oneof=top up down"`
}

// =================================================================================
// UpdateCategorySort 排序
// 逻辑：必须是【所属科目】的创建者
// =================================================================================
func UpdateCategorySort(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req UpdateCategorySortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	currentUserCode, _ := c.Get("userCode")

	// --- 权限检查 ---
	var subjectCreatorCode string
	var creatorName string
	var creatorEmail sql.NullString
	var currentSubjectID int
	var currentSortOrder int

	checkSQL := `
		SELECT s.creator_code, IFNULL(u.nickname, u.username), u.email, c.subject_id, c.sort_order
		FROM knowledge_categories c
		JOIN subjects s ON c.subject_id = s.id
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE c.id = ?
	`
	err := global.DB.QueryRow(checkSQL, id).Scan(&subjectCreatorCode, &creatorName, &creatorEmail, &currentSubjectID, &currentSortOrder)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "分类不存在"})
		return
	}

	currentUserCodeStr, _ := currentUserCode.(string)
	if subjectCreatorCode != currentUserCodeStr {
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "排序失败：请联系科目作者 " + contactInfo})
		return
	}

	// --- 开启事务执行排序 ---
	tx, _ := global.DB.Begin()
	defer tx.Rollback()

	switch req.Action {
	case "top":
		var minSort int
		_ = tx.QueryRow("SELECT COALESCE(MIN(sort_order), 0) FROM knowledge_categories WHERE subject_id = ?", currentSubjectID).Scan(&minSort)
		_, _ = tx.Exec("UPDATE knowledge_categories SET sort_order = ? WHERE id = ?", minSort-1, id)

	case "up":
		var targetID, targetSort int
		err = tx.QueryRow(`
			SELECT id, sort_order FROM knowledge_categories 
			WHERE subject_id = ? AND sort_order < ? 
			ORDER BY sort_order DESC LIMIT 1`, currentSubjectID, currentSortOrder).Scan(&targetID, &targetSort)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已是第一位"})
			return
		}
		_, _ = tx.Exec("UPDATE knowledge_categories SET sort_order = ? WHERE id = ?", currentSortOrder, targetID)
		_, _ = tx.Exec("UPDATE knowledge_categories SET sort_order = ? WHERE id = ?", targetSort, id)

	case "down":
		var targetID, targetSort int
		err = tx.QueryRow(`
			SELECT id, sort_order FROM knowledge_categories 
			WHERE subject_id = ? AND sort_order > ? 
			ORDER BY sort_order ASC LIMIT 1`, currentSubjectID, currentSortOrder).Scan(&targetID, &targetSort)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已是最后一位"})
			return
		}
		_, _ = tx.Exec("UPDATE knowledge_categories SET sort_order = ? WHERE id = ?", currentSortOrder, targetID)
		_, _ = tx.Exec("UPDATE knowledge_categories SET sort_order = ? WHERE id = ?", targetSort, id)
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "排序成功"})
}

// 辅助函数
func getContactInfo(name string, email sql.NullString) string {
	if email.Valid && email.String != "" {
		return email.String
	}
	return name + " (未设置邮箱)"
}
