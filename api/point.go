package api

import (
	"database/sql"
	"encoding/json"
	_ "fmt"
	"net/http"
	"practice_problems/global"
	"practice_problems/model"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// =================================================================================
// GetPointList 获取知识点列表
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

	sqlStr := "SELECT id, title, create_time, sort_order, difficulty FROM knowledge_points WHERE categorie_id = ? ORDER BY sort_order ASC, id DESC"

	rows, err := global.DB.Query(sqlStr, catID)
	if err != nil {
		// ★★★ Error ★★★
		global.GetLog(c).Errorf("查询知识点列表失败 (CatID: %s): %v", catID, err)
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
// =================================================================================
func GetPointDetail(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	// 1. 权限校验
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

	var p model.KnowledgePoint

	// 2. 查询详情 (核心修改)
	// ★★★ 这里的 SQL 加入了 COALESCE(video_url, '[]') ★★★
	sqlStr := `SELECT id, categorie_id, title, content, 
	           COALESCE(video_url, '[]'),      -- <--- 新增：查出视频字段
	           COALESCE(reference_links, '[]'), COALESCE(local_image_names, '[]'), 
	           create_time, update_time 
	           FROM knowledge_points WHERE id = ?`

	// 3. 扫描赋值
	err = global.DB.QueryRow(sqlStr, id).Scan(
		&p.ID, &p.CategoryID, &p.Title, &p.Content,
		&p.VideoUrl, // <--- 新增：赋值给结构体 (注意顺序要和 SQL 对应)
		&p.ReferenceLinks, &p.LocalImageNames,
		&p.CreateTime, &p.UpdateTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该知识点"})
		} else {
			global.GetLog(c).Errorf("查询知识点详情失败 (ID: %s): %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询详情失败"})
		}
		return
	}

	// 4. 查询该知识点的所有绑定（包含目标知识点所属分类ID，便于前端按需加载缓存）
	bindings := make([]gin.H, 0)
	bindingsSQL := `
		SELECT pb.id, pb.bind_text, pb.target_point_id,
		       COALESCE(tp.categorie_id, 0) as target_category_id
		FROM point_bindings pb
		LEFT JOIN knowledge_points tp ON pb.target_point_id = tp.id
		WHERE pb.source_point_id = ?
		ORDER BY pb.create_time DESC
	`
	bindingRows, err := global.DB.Query(bindingsSQL, id)
	if err == nil {
		defer bindingRows.Close()
		for bindingRows.Next() {
			var bindID, targetPointID, targetCategoryID int
			var bindText string
			if bindingRows.Scan(&bindID, &bindText, &targetPointID, &targetCategoryID) == nil {
				bindings = append(bindings, gin.H{
					"id":               bindID,
					"bindText":         bindText,
					"targetPointId":    targetPointID,
					"targetCategoryId": targetCategoryID,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"point":    p,
			"bindings": bindings,
		},
	})
}

// =================================================================================
// CreatePoint 创建知识点
// =================================================================================
func CreatePoint(c *gin.Context) {
	var req model.CreatePointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
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

	if subjectCreatorCode != currentUserCodeStr {
		// ★★★ Warn ★★★
		global.GetLog(c).Warnf("创建知识点被拒: 无权操作 (User: %s, CatID: %d)", currentUserCodeStr, req.CategoryID)
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(403, gin.H{"code": 403, "msg": "创建失败：您不是该科目的作者，请联系 " + contactInfo})
		return
	}

	// --- 逻辑执行 ---
	var currentMin int
	row := global.DB.QueryRow("SELECT COALESCE(MIN(sort_order), 0) FROM knowledge_points WHERE categorie_id = ?", req.CategoryID)
	row.Scan(&currentMin)
	newSortOrder := currentMin - 1

	res, err := global.DB.Exec(
		"INSERT INTO knowledge_points (categorie_id, title, content, sort_order, difficulty) VALUES (?, ?, '', ?, 0)",
		req.CategoryID, req.Title, newSortOrder,
	)

	if err != nil {
		// ★★★ Error ★★★
		global.GetLog(c).Errorf("创建知识点DB错误: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "创建失败"})
		return
	}

	id, _ := res.LastInsertId()
	// ★★★ Info ★★★
	global.GetLog(c).Infof("用户[%s] 创建知识点成功: ID=%d, Title=%s", currentUserCodeStr, id, req.Title)
	c.JSON(200, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{"id": id}})
}

// =================================================================================
// cleanupOrphanedBindings 清理内容中不再存在的绑定文本
// =================================================================================
func cleanupOrphanedBindings(c *gin.Context, pointID string, newContent string) {
	// 查询该知识点的所有绑定
	rows, err := global.DB.Query("SELECT id, bind_text FROM point_bindings WHERE source_point_id = ?", pointID)
	if err != nil {
		global.GetLog(c).Warnf("查询绑定列表失败: %v", err)
		return
	}
	defer rows.Close()

	var toDelete []int
	for rows.Next() {
		var bindID int
		var bindText string
		if rows.Scan(&bindID, &bindText) == nil {
			// 检查绑定文本是否还存在于新内容中
			if !strings.Contains(newContent, bindText) {
				toDelete = append(toDelete, bindID)
			}
		}
	}

	// 删除不再匹配的绑定
	for _, bindID := range toDelete {
		_, err := global.DB.Exec("DELETE FROM point_bindings WHERE id = ?", bindID)
		if err != nil {
			global.GetLog(c).Warnf("删除孤立绑定失败 (ID: %d): %v", bindID, err)
		} else {
			global.GetLog(c).Infof("自动清理不匹配的绑定 (ID: %d, PointID: %s)", bindID, pointID)
		}
	}
}

// =================================================================================
// UpdatePoint 修改知识点
// =================================================================================
func UpdatePoint(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdatePointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	currentUserCode, _ := c.Get("userCode")
	currentUserCodeStr, _ := currentUserCode.(string)

	// --- 权限检查 (保持原样) ---
	var subjectCreatorCode string
	var currentSubjectId int
	var creatorName string
	var creatorEmail sql.NullString

	checkSQL := `
		SELECT s.creator_code, s.id, IFNULL(u.nickname, u.username), u.email
		FROM knowledge_points p
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN subjects s ON c.subject_id = s.id
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE p.id = ?
	`
	err := global.DB.QueryRow(checkSQL, id).Scan(&subjectCreatorCode, &currentSubjectId, &creatorName, &creatorEmail)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "msg": "知识点不存在"})
		return
	}

	if subjectCreatorCode != currentUserCodeStr {
		global.GetLog(c).Warnf("修改知识点被拒: 无权操作 (User: %s, PointID: %s)", currentUserCodeStr, id)
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(403, gin.H{"code": 403, "msg": "修改失败：请联系科目作者 " + contactInfo})
		return
	}

	// --- 核心修改：检查并准备 SQL 更新 ---
	query := "UPDATE knowledge_points SET update_time = CURRENT_TIMESTAMP"
	var args []interface{}

	// 1. 处理分类移动逻辑 (保持原样)
	if req.CategoryID != nil && *req.CategoryID > 0 {
		var targetSubjectId int
		err := global.DB.QueryRow("SELECT subject_id FROM knowledge_categories WHERE id = ?", *req.CategoryID).Scan(&targetSubjectId)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(400, gin.H{"code": 400, "msg": "目标分类不存在"})
				return
			}
			global.GetLog(c).Errorf("检查目标分类失败: %v", err)
			c.JSON(500, gin.H{"code": 500, "msg": "系统错误"})
			return
		}

		if targetSubjectId != currentSubjectId {
			c.JSON(400, gin.H{"code": 400, "msg": "不能跨科目移动知识点"})
			return
		}
		query += ", categorie_id = ?"
		args = append(args, *req.CategoryID)
	}

	// 2. 处理常规字段
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

	// ★★★ 新增：处理视频链接更新 ★★★
	// 使用指针判断：如果前端传了这个字段(不为nil)，就更新它
	if req.VideoUrl != nil {
		query += ", video_url = ?"
		args = append(args, *req.VideoUrl)
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

	// --- 执行更新 ---
	if len(args) == 0 {
		c.JSON(200, gin.H{"code": 200, "msg": "无变更"})
		return
	}

	query += " WHERE id = ?"
	args = append(args, id)

	_, err = global.DB.Exec(query, args...)
	if err != nil {
		global.GetLog(c).Errorf("更新知识点DB错误 (ID: %s): %v", id, err)
		c.JSON(500, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	// ★★★ 如果更新了content，清理不再匹配的绑定 ★★★
	if req.Content != "" {
		cleanupOrphanedBindings(c, id, req.Content)
	}

	global.GetLog(c).Infof("用户[%s] 更新知识点成功 (ID: %s)", currentUserCodeStr, id)
	c.JSON(200, gin.H{"code": 200, "msg": "更新成功"})
}

// =================================================================================
// DeletePoint 删除知识点
// =================================================================================
func DeletePoint(c *gin.Context) {
	id := c.Param("id")
	currentUserCode, _ := c.Get("userCode")
	currentUserCodeStr, _ := currentUserCode.(string)

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

	if subjectCreatorCode != currentUserCodeStr {
		global.GetLog(c).Warnf("删除知识点被拒: 无权操作 (User: %s, PointID: %s)", currentUserCodeStr, id)
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(403, gin.H{"code": 403, "msg": "删除失败：请联系科目作者 " + contactInfo})
		return
	}

	// --- 执行删除 ---
	_, err = global.DB.Exec("DELETE FROM knowledge_points WHERE id = ?", id)

	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			global.GetLog(c).Infof("删除知识点失败(外键约束): ID=%s", id)
			c.JSON(500, gin.H{"code": 500, "msg": "删除失败：该知识点下仍有题目，请先删除题目"})
			return
		}
		global.GetLog(c).Errorf("删除知识点DB错误 (ID: %s): %v", id, err)
		c.JSON(500, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	global.GetLog(c).Infof("用户[%s] 删除知识点成功 (ID: %s)", currentUserCodeStr, id)
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
	currentUserCodeStr, _ := currentUserCode.(string)

	// --- 权限检查 ---
	var subjectCreatorCode string
	var creatorName string
	var creatorEmail sql.NullString
	var currentCategoryID int
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

	if subjectCreatorCode != currentUserCodeStr {
		global.GetLog(c).Warnf("知识点排序被拒: 无权操作 (User: %s, PointID: %d)", currentUserCodeStr, id)
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
		var minSort int
		tx.QueryRow("SELECT COALESCE(MIN(sort_order), 0) FROM knowledge_points WHERE categorie_id = ?", currentCategoryID).Scan(&minSort)
		tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", minSort-1, id)

	} else if req.Action == "up" {
		var prevID, prevSort int
		err = tx.QueryRow("SELECT id, sort_order FROM knowledge_points WHERE categorie_id = ? AND sort_order < ? ORDER BY sort_order DESC LIMIT 1", currentCategoryID, currentSortOrder).Scan(&prevID, &prevSort)
		if err == nil {
			tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", prevSort, id)
			tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", currentSortOrder, prevID)
		}

	} else if req.Action == "down" {
		var nextID, nextSort int
		err = tx.QueryRow("SELECT id, sort_order FROM knowledge_points WHERE categorie_id = ? AND sort_order > ? ORDER BY sort_order ASC LIMIT 1", currentCategoryID, currentSortOrder).Scan(&nextID, &nextSort)
		if err == nil {
			tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", nextSort, id)
			tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", currentSortOrder, nextID)
		}
	}

	if err := tx.Commit(); err != nil {
		global.GetLog(c).Errorf("知识点排序事务提交失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "排序失败"})
		return
	}

	global.GetLog(c).Infof("用户[%s] 排序知识点成功 (ID: %d, Action: %s)", currentUserCodeStr, id, req.Action)
	c.JSON(200, gin.H{"code": 200, "msg": "排序成功"})
}

// DeletePointImage 删除知识点图片
func DeletePointImage(c *gin.Context) {
	id := c.Param("id")
	currentUserCode, _ := c.Get("userCode")
	currentUserCodeStr, _ := currentUserCode.(string)

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

	if subjectCreatorCode != currentUserCodeStr {
		global.GetLog(c).Warnf("删除图片被拒: 无权操作 (User: %s, PointID: %s)", currentUserCodeStr, id)
		c.JSON(403, gin.H{"code": 403, "msg": "无权删除图片"})
		return
	}

	// --- 逻辑执行 ---
	var localImageNamesStr string
	err = global.DB.QueryRow("SELECT COALESCE(local_image_names, '[]') FROM knowledge_points WHERE id = ?", id).Scan(&localImageNamesStr)
	if err != nil {
		global.GetLog(c).Errorf("删除图片查询DB失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "数据库查询失败"})
		return
	}

	var images []model.ImageItem
	if err := json.Unmarshal([]byte(localImageNamesStr), &images); err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "图片数据解析异常"})
		return
	}

	newImages := make([]model.ImageItem, 0)
	found := false

	for _, item := range images {
		cleanDbUrl := strings.TrimPrefix(item.Url, "/")
		cleanReqUrl := strings.TrimPrefix(req.FilePath, "/")

		if cleanDbUrl != cleanReqUrl {
			newImages = append(newImages, item)
		} else {
			found = true
		}
	}

	if !found {
		c.JSON(404, gin.H{"code": 404, "msg": "未找到指定图片"})
		return
	}

	newJsonBytes, _ := json.Marshal(newImages)

	_, err = global.DB.Exec("UPDATE knowledge_points SET local_image_names = ? WHERE id = ?", string(newJsonBytes), id)
	if err != nil {
		global.GetLog(c).Errorf("删除图片更新DB失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "数据库更新失败"})
		return
	}

	diskPath := strings.TrimPrefix(req.FilePath, "/")
	RemoveFileFromDisk(diskPath)

	global.GetLog(c).Infof("用户[%s] 删除图片成功: %s", currentUserCodeStr, req.FilePath)
	c.JSON(200, gin.H{"code": 200, "msg": "图片删除成功"})
}
