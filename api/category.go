package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"practice_problems/global"
	"practice_problems/model"
	"regexp"
	"strconv"
	"strings"
	_ "time"

	"github.com/gin-gonic/gin"
)

// =================================================================================
// GetCategoryList 获取分类列表（支持分页）
// =================================================================================
func GetCategoryList(c *gin.Context) {
	subjectIDStr := c.Query("subject_id")

	if subjectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "必须指定科目ID"})
		return
	}

	userID, exists := c.Get("userID")
	userCode, _ := c.Get("userCode")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	var hasPerm int
	checkPermSQL := `
		SELECT 1 
		FROM subjects s
		LEFT JOIN user_subjects us ON s.id = us.subject_id AND us.user_id = ?
		WHERE s.id = ? 
		  AND (
		      s.creator_code = ?
		      OR
		      (us.id IS NOT NULL AND us.status = 1 AND (us.expire_time IS NULL OR us.expire_time > datetime('now', 'localtime')))
		  )
	`

	err := global.DB.QueryRow(checkPermSQL, userID, subjectIDStr, userCode).Scan(&hasPerm)

	if err != nil || hasPerm != 1 {
		// 这里的权限拒绝比较常见（比如过期），可以不打日志，或者打 Debug
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权访问该科目或授权已过期"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "11"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 11
	}
	offset := (page - 1) * pageSize

	// 查询总数
	var total int
	countErr := global.DB.QueryRow("SELECT COUNT(*) FROM knowledge_categories WHERE subject_id = ?", subjectIDStr).Scan(&total)
	if countErr != nil {
		global.GetLog(c).Errorf("查询分类总数失败: %v", countErr)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	// 分页查询分类列表
	fields := "id, subject_id, categorie_name, create_time, update_time, sort_order, difficulty"
	sqlStr := fmt.Sprintf("SELECT %s FROM knowledge_categories WHERE subject_id = ? ORDER BY sort_order ASC, id DESC LIMIT ? OFFSET ?", fields)

	rows, err := global.DB.Query(sqlStr, subjectIDStr, pageSize, offset)
	if err != nil {
		// ★★★ Error: 数据库查询出错需要记录 ★★★
		global.GetLog(c).Errorf("查询分类列表失败 (SubjectID: %s): %v", subjectIDStr, err)
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

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"list":     list,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// =================================================================================
// CreateCategory 创建分类
// =================================================================================
func CreateCategory(c *gin.Context) {
	var req model.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	currentUserCode, _ := c.Get("userCode")
	currentUserCodeStr, _ := currentUserCode.(string)

	// --- 权限检查 ---
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

	if subjectCreatorCode != currentUserCodeStr {
		// ★★★ Warn: 记录越权操作尝试 ★★★
		global.GetLog(c).Warnf("创建分类失败: 无权操作 (User: %s, Subject: %d)", currentUserCodeStr, req.SubjectID)
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "创建失败：您不是该科目的作者，请联系 " + contactInfo})
		return
	}

	// --- ★★★ 新增：生成带序号的名称 ★★★ ---
	// 1. 统计当前科目下已有多少个分类
	var count int
	err = global.DB.QueryRow("SELECT COUNT(*) FROM knowledge_categories WHERE subject_id = ?", req.SubjectID).Scan(&count)
	if err != nil {
		global.GetLog(c).Errorf("统计分类数量失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "系统错误"})
		return
	}

	// 2. 拼接新名称 (例如: "3. 基础语法")
	finalCategoryName := fmt.Sprintf("%d. %s", count+1, req.CategoryName)

	// --- 计算排序并插入 ---
	var currentMinSort int
	sqlQueryMin := "SELECT COALESCE(MIN(sort_order), 0) FROM knowledge_categories WHERE subject_id = ?"
	_ = global.DB.QueryRow(sqlQueryMin, req.SubjectID).Scan(&currentMinSort)
	newSortOrder := currentMinSort - 1

	// 注意：这里使用 finalCategoryName 和 req.Difficulty
	sqlStr := "INSERT INTO knowledge_categories (subject_id, categorie_name, sort_order, difficulty) VALUES (?, ?, ?, ?)"
	result, err := global.DB.Exec(sqlStr, req.SubjectID, finalCategoryName, newSortOrder, req.Difficulty)

	if err != nil {
		// ★★★ Error: 数据库插入失败 ★★★
		global.GetLog(c).Errorf("创建分类DB错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建失败"})
		return
	}

	newID, _ := result.LastInsertId()

	// ★★★ Info: 记录成功创建 ★★★
	global.GetLog(c).Infof("用户[%s] 创建分类成功: ID=%d, Name=%s", currentUserCodeStr, newID, finalCategoryName)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{"id": newID, "name": finalCategoryName}})
}

// =================================================================================
// UpdateCategory 修改分类
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
	currentUserCodeStr, _ := currentUserCode.(string)

	// --- 权限检查 & 获取旧数据 ---
	var subjectCreatorCode string
	var creatorName string
	var creatorEmail sql.NullString
	// ★★★ 新增变量接收旧数据
	var currentCategoryName string
	var currentSubjectID int

	// ★★★ 修改 SQL: 多查询了 c.categorie_name 和 c.subject_id
	checkSQL := `
		SELECT s.creator_code, c.categorie_name, c.subject_id, IFNULL(u.nickname, u.username), u.email
		FROM knowledge_categories c
		JOIN subjects s ON c.subject_id = s.id
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE c.id = ?
	`
	err := global.DB.QueryRow(checkSQL, id).Scan(&subjectCreatorCode, &currentCategoryName, &currentSubjectID, &creatorName, &creatorEmail)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "分类不存在"})
		return
	}

	if subjectCreatorCode != currentUserCodeStr {
		global.GetLog(c).Warnf("修改分类失败: 无权操作 (User: %s, CategoryID: %d)", currentUserCodeStr, id)
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "修改失败：请联系科目作者 " + contactInfo})
		return
	}

	// --- 执行更新 ---
	query := "UPDATE knowledge_categories SET update_time = CURRENT_TIMESTAMP"
	var args []interface{}

	// ★★★ 核心修改：处理分类名称更新 ★★★
	if req.CategoryName != "" {
		// 1. 正则匹配开头的 "数字."
		re := regexp.MustCompile(`^(\d+\.)\s*`)

		// 2. 清洗用户输入 (去掉用户自己瞎写的序号)
		cleanNewName := re.ReplaceAllString(req.CategoryName, "")
		if cleanNewName == "" {
			cleanNewName = req.CategoryName // 防止被洗空
		}

		// 3. 分析旧名称，决定使用什么序号
		oldMatches := re.FindStringSubmatch(currentCategoryName)
		var finalName string

		if len(oldMatches) > 1 {
			// 情况A: 旧名称本来就有序号 (oldMatches[1] 是 "1.")
			// 强制保留旧序号
			finalName = fmt.Sprintf("%s %s", oldMatches[1], cleanNewName)
		} else {
			// 情况B: 旧名称没有序号，自动生成
			// 统计该科目下有多少分类 (包含自己)
			var count int
			global.DB.QueryRow("SELECT COUNT(*) FROM knowledge_categories WHERE subject_id = ?", currentSubjectID).Scan(&count)

			// 生成新序号
			finalName = fmt.Sprintf("%d. %s", count, cleanNewName)
		}

		query += ", categorie_name = ?"
		args = append(args, finalName)
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
		global.GetLog(c).Errorf("更新分类DB错误 (ID: %d): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	global.GetLog(c).Infof("用户[%s] 更新分类成功 (ID: %d)", currentUserCodeStr, id)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// =================================================================================
// DeleteCategory 删除分类
// =================================================================================
func DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	currentUserCode, _ := c.Get("userCode")
	currentUserCodeStr, _ := currentUserCode.(string)

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

	if subjectCreatorCode != currentUserCodeStr {
		global.GetLog(c).Warnf("删除分类失败: 无权操作 (User: %s, CategoryID: %d)", currentUserCodeStr, id)
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "删除失败：请联系科目作者 " + contactInfo})
		return
	}

	// --- 执行删除 ---
	sqlStr := "DELETE FROM knowledge_categories WHERE id = ?"
	result, err := global.DB.Exec(sqlStr, id)
	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY") {
			// 这种属于业务逻辑阻止，不算系统错误，也可以记一下 Info 或 Warn
			global.GetLog(c).Infof("删除分类失败(外键约束): ID=%d", id)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败：该分类下仍有知识点"})
			return
		}
		global.GetLog(c).Errorf("删除分类DB错误 (ID: %d): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该分类"})
		return
	}

	global.GetLog(c).Infof("用户[%s] 删除分类成功 (ID: %d)", currentUserCodeStr, id)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// UpdateCategorySortRequest 排序请求
type UpdateCategorySortRequest struct {
	Action string `json:"action" binding:"required,oneof=top up down"`
}

// =================================================================================
// UpdateCategorySort 排序
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
	currentUserCodeStr, _ := currentUserCode.(string)

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

	if subjectCreatorCode != currentUserCodeStr {
		global.GetLog(c).Warnf("分类排序失败: 无权操作 (User: %s, CategoryID: %d)", currentUserCodeStr, id)
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

	if err := tx.Commit(); err != nil {
		global.GetLog(c).Errorf("分类排序事务提交失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "排序失败"})
		return
	}

	global.GetLog(c).Infof("用户[%s] 排序分类成功 (ID: %d, Action: %s)", currentUserCodeStr, id, req.Action)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "排序成功"})
}

// 辅助函数
func getContactInfo(name string, email sql.NullString) string {
	if email.Valid && email.String != "" {
		return email.String
	}
	return name + " (未设置邮箱)"
}
