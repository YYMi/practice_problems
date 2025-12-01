package api

import (
	"database/sql"
	"encoding/json"
	_ "fmt"
	"log"
	"net/http"
	"practice_problems/global"
	"practice_problems/model"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// =================================================================================
// GetPointList 获取知识点列表
// 权限逻辑：只要用户绑定了该知识点所属的科目 (user_subjects)，就有权查看
// =================================================================================
func GetPointList(c *gin.Context) {
	catID := c.Query("category_id")
	if catID == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "必须指定分类ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// --- 1. 权限检查 ---
	// 逻辑：通过 category_id 找到 subject_id，然后检查 user_subjects 表
	var hasPerm int
	checkPermSQL := `
		SELECT 1 
		FROM knowledge_categories c
		JOIN user_subjects us ON c.subject_id = us.subject_id
		WHERE c.id = ? AND us.user_id = ?
	`
	err := global.DB.QueryRow(checkPermSQL, catID, userID).Scan(&hasPerm)
	if err != nil || hasPerm != 1 {
		c.JSON(403, gin.H{"code": 403, "msg": "无权访问该分类下的内容"})
		return
	}

	// --- 2. 查询列表 ---
	sqlStr := "SELECT id, title, create_time, sort_order, difficulty FROM knowledge_points WHERE categorie_id = ? ORDER BY sort_order ASC, id DESC"

	rows, err := global.DB.Query(sqlStr, catID)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	list := make([]gin.H, 0)
	for rows.Next() {
		var id int
		var title string
		var createTime string
		var sortOrder int
		var difficulty int

		err := rows.Scan(&id, &title, &createTime, &sortOrder, &difficulty)
		if err != nil {
			continue
		}

		list = append(list, gin.H{
			"id":         id,
			"title":      title,
			"createTime": createTime,
			"sortOrder":  sortOrder,
			"difficulty": difficulty,
		})
	}

	c.JSON(200, gin.H{"code": 200, "msg": "success", "data": list})
}

// =================================================================================
// GetPointDetail 获取知识点详情
// 权限逻辑：同上 (绑定即可看)
// =================================================================================
func GetPointDetail(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	// --- 1. 权限检查 ---
	// 通过 point -> category -> subject -> user_subjects
	var hasPerm int
	checkPermSQL := `
		SELECT 1
		FROM knowledge_points p
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN user_subjects us ON c.subject_id = us.subject_id
		WHERE p.id = ? AND us.user_id = ?
	`
	err := global.DB.QueryRow(checkPermSQL, id, userID).Scan(&hasPerm)
	if err != nil || hasPerm != 1 {
		c.JSON(403, gin.H{"code": 403, "msg": "无权查看该知识点"})
		return
	}

	// --- 2. 查询详情 ---
	var p model.KnowledgePoint
	sqlStr := `SELECT id, categorie_id, title, content, 
	           COALESCE(reference_links, '[]'), COALESCE(local_image_names, '[]'), 
	           create_time, update_time 
	           FROM knowledge_points WHERE id = ?`

	err = global.DB.QueryRow(sqlStr, id).Scan(
		&p.ID, &p.CategoryID, &p.Title, &p.Content,
		&p.ReferenceLinks, &p.LocalImageNames,
		&p.CreateTime, &p.UpdateTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该知识点"})
		} else {
			log.Println("查询详情失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询详情失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": p,
	})
}

// =================================================================================
// CreatePoint 创建知识点
// 权限逻辑：必须是该【科目】的创建者 (creator_code)
// =================================================================================
func CreatePoint(c *gin.Context) {
	var req model.CreatePointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	currentUserCode, _ := c.Get("userCode")

	// --- 权限检查 ---
	// 通过 category_id -> subject -> creator_code
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
	err := global.DB.QueryRow(checkSQL, req.CategoryID).Scan(&subjectCreatorCode, &creatorName, &creatorEmail)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "msg": "所属分类不存在"})
		return
	}

	currentUserCodeStr, _ := currentUserCode.(string)
	if subjectCreatorCode != currentUserCodeStr {
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(403, gin.H{"code": 403, "msg": "创建失败：您不是该科目的作者，请联系 " + contactInfo})
		return
	}

	// --- 逻辑执行 ---
	// 1. 计算新排序值
	var currentMin int
	row := global.DB.QueryRow("SELECT COALESCE(MIN(sort_order), 0) FROM knowledge_points WHERE categorie_id = ?", req.CategoryID)
	row.Scan(&currentMin)
	newSortOrder := currentMin - 1

	// 2. 插入数据 (不再需要 creator_code，因为权限归属科目)
	// 注意：Content 默认为空字符串
	res, err := global.DB.Exec(
		"INSERT INTO knowledge_points (categorie_id, title, content, sort_order, difficulty) VALUES (?, ?, '', ?, 0)",
		req.CategoryID, req.Title, newSortOrder,
	)

	if err != nil {
		log.Println("CreatePoint DB Error:", err)
		c.JSON(500, gin.H{"code": 500, "msg": "创建失败"})
		return
	}

	id, _ := res.LastInsertId()
	c.JSON(200, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{"id": id}})
}

// =================================================================================
// UpdatePoint 修改知识点
// 权限逻辑：必须是该【科目】的创建者
// =================================================================================
func UpdatePoint(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdatePointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	currentUserCode, _ := c.Get("userCode")

	// --- 权限检查 ---
	// point -> category -> subject -> creator_code
	var subjectCreatorCode string
	var creatorName string
	var creatorEmail sql.NullString

	checkSQL := `
		SELECT s.creator_code, IFNULL(u.nickname, u.username), u.email
		FROM knowledge_points p
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN subjects s ON c.subject_id = s.id
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE p.id = ?
	`
	err := global.DB.QueryRow(checkSQL, id).Scan(&subjectCreatorCode, &creatorName, &creatorEmail)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "msg": "知识点不存在"})
		return
	}

	currentUserCodeStr, _ := currentUserCode.(string)
	if subjectCreatorCode != currentUserCodeStr {
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(403, gin.H{"code": 403, "msg": "修改失败：请联系科目作者 " + contactInfo})
		return
	}

	// --- 执行更新 ---
	query := "UPDATE knowledge_points SET update_time = CURRENT_TIMESTAMP"
	var args []interface{}

	if req.Title != "" {
		query += ", title = ?"
		args = append(args, req.Title)
	}
	if req.Content != "" {
		query += ", content = ?"
		args = append(args, req.Content)
	}
	if req.ReferenceLinks != "" {
		query += ", reference_links = ?"
		args = append(args, req.ReferenceLinks)
	}
	if req.LocalImageNames != "" {
		query += ", local_image_names = ?"
		args = append(args, req.LocalImageNames)
	}
	if req.Difficulty != nil {
		if *req.Difficulty < 0 || *req.Difficulty > 3 {
			c.JSON(400, gin.H{"code": 400, "msg": "难度无效"})
			return
		}
		query += ", difficulty = ?"
		args = append(args, *req.Difficulty)
	}

	if len(args) == 0 {
		c.JSON(200, gin.H{"code": 200, "msg": "无变更"})
		return
	}

	query += " WHERE id = ?"
	args = append(args, id)

	_, err = global.DB.Exec(query, args...)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "更新失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "msg": "更新成功"})
}

// =================================================================================
// DeletePoint 删除知识点
// 权限逻辑：必须是该【科目】的创建者
// =================================================================================
func DeletePoint(c *gin.Context) {
	id := c.Param("id")
	currentUserCode, _ := c.Get("userCode")

	// --- 权限检查 ---
	var subjectCreatorCode string
	var creatorName string
	var creatorEmail sql.NullString

	checkSQL := `
		SELECT s.creator_code, IFNULL(u.nickname, u.username), u.email
		FROM knowledge_points p
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN subjects s ON c.subject_id = s.id
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE p.id = ?
	`
	err := global.DB.QueryRow(checkSQL, id).Scan(&subjectCreatorCode, &creatorName, &creatorEmail)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "msg": "知识点不存在"})
		return
	}

	currentUserCodeStr, _ := currentUserCode.(string)
	if subjectCreatorCode != currentUserCodeStr {
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(403, gin.H{"code": 403, "msg": "删除失败：请联系科目作者 " + contactInfo})
		return
	}

	// --- 执行删除 ---
	_, err = global.DB.Exec("DELETE FROM knowledge_points WHERE id = ?", id)

	if err != nil {
		log.Println("删除知识点失败:", err)
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			c.JSON(500, gin.H{"code": 500, "msg": "删除失败：该知识点下仍有题目，请先删除题目"})
			return
		}
		c.JSON(500, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	c.JSON(200, gin.H{"code": 200, "msg": "删除成功"})
}

type UpdateSortRequest struct {
	Action string `json:"action"` // "top", "up", "down"
}

// =================================================================================
// UpdatePointSort 知识点排序
// =================================================================================
func UpdatePointSort(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req UpdateSortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	currentUserCode, _ := c.Get("userCode")

	// --- 权限检查 ---
	var subjectCreatorCode string
	var creatorName string
	var creatorEmail sql.NullString
	var currentCategoryID int

	// ★★★ 统一变量名：这里叫 currentSortOrder ★★★
	var currentSortOrder int

	checkSQL := `
		SELECT s.creator_code, IFNULL(u.nickname, u.username), u.email, p.categorie_id, p.sort_order
		FROM knowledge_points p
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN subjects s ON c.subject_id = s.id
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE p.id = ?
	`
	err := global.DB.QueryRow(checkSQL, id).Scan(&subjectCreatorCode, &creatorName, &creatorEmail, &currentCategoryID, &currentSortOrder)

	if err != nil {
		c.JSON(404, gin.H{"code": 404, "msg": "知识点不存在"})
		return
	}

	currentUserCodeStr, _ := currentUserCode.(string)
	if subjectCreatorCode != currentUserCodeStr {
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(403, gin.H{"code": 403, "msg": "排序失败：请联系科目作者 " + contactInfo})
		return
	}

	// --- 开启事务排序 ---
	tx, err := global.DB.Begin()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "服务器错误"})
		return
	}
	defer tx.Rollback()

	if req.Action == "top" {
		// 置顶
		var minSort int
		tx.QueryRow("SELECT COALESCE(MIN(sort_order), 0) FROM knowledge_points WHERE categorie_id = ?", currentCategoryID).Scan(&minSort)
		tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", minSort-1, id)

	} else if req.Action == "up" {
		// 上移
		var prevID, prevSort int
		err = tx.QueryRow("SELECT id, sort_order FROM knowledge_points WHERE categorie_id = ? AND sort_order < ? ORDER BY sort_order DESC LIMIT 1", currentCategoryID, currentSortOrder).Scan(&prevID, &prevSort)
		if err == nil {
			// ★★★ 修复：这里使用 currentSortOrder ★★★
			tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", prevSort, id)
			tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", currentSortOrder, prevID)
		}

	} else if req.Action == "down" {
		// 下移
		var nextID, nextSort int
		err = tx.QueryRow("SELECT id, sort_order FROM knowledge_points WHERE categorie_id = ? AND sort_order > ? ORDER BY sort_order ASC LIMIT 1", currentCategoryID, currentSortOrder).Scan(&nextID, &nextSort)
		if err == nil {
			// ★★★ 修复：这里使用 currentSortOrder ★★★
			tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", nextSort, id)
			tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", currentSortOrder, nextID)
		}
	}

	tx.Commit()
	c.JSON(200, gin.H{"code": 200, "msg": "排序成功"})
}

// DeletePointImage 删除知识点图片 (通常无需严格权限，或者跟随修改权限)
// 这里为了简单，也加上权限判断吧
func DeletePointImage(c *gin.Context) {
	id := c.Param("id")
	currentUserCode, _ := c.Get("userCode")

	var req model.DeletePointImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// --- 权限检查 ---
	var subjectCreatorCode string
	checkSQL := `
		SELECT s.creator_code
		FROM knowledge_points p
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN subjects s ON c.subject_id = s.id
		WHERE p.id = ?
	`
	err := global.DB.QueryRow(checkSQL, id).Scan(&subjectCreatorCode)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "msg": "知识点不存在"})
		return
	}
	currentUserCodeStr, _ := currentUserCode.(string)
	if subjectCreatorCode != currentUserCodeStr {
		c.JSON(403, gin.H{"code": 403, "msg": "无权删除图片"})
		return
	}

	// --- 逻辑执行 ---
	var localImageNamesStr string
	err = global.DB.QueryRow("SELECT COALESCE(local_image_names, '[]') FROM knowledge_points WHERE id = ?", id).Scan(&localImageNamesStr)

	var images []string
	json.Unmarshal([]byte(localImageNamesStr), &images)

	newImages := make([]string, 0)
	found := false
	for _, img := range images {
		if img != req.FilePath {
			newImages = append(newImages, img)
		} else {
			found = true
		}
	}

	if !found {
		c.JSON(404, gin.H{"code": 404, "msg": "未找到指定图片"})
		return
	}

	newJsonBytes, _ := json.Marshal(newImages)
	global.DB.Exec("UPDATE knowledge_points SET local_image_names = ? WHERE id = ?", string(newJsonBytes), id)

	RemoveFileFromDisk(req.FilePath)

	c.JSON(200, gin.H{"code": 200, "msg": "图片删除成功"})
}
