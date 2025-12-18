package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"practice_problems/global"
	"practice_problems/model"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// =================================================================================
// GetPointList 获取知识点列表（支持分页）
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
	countErr := global.DB.QueryRow("SELECT COUNT(*) FROM knowledge_points WHERE categorie_id = ?", catID).Scan(&total)
	if countErr != nil {
		global.GetLog(c).Errorf("查询知识点总数失败: %v", countErr)
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	// 分页查询
	sqlStr := "SELECT id, title, create_time, sort_order, difficulty FROM knowledge_points WHERE categorie_id = ? ORDER BY sort_order ASC, id DESC LIMIT ? OFFSET ?"

	rows, err := global.DB.Query(sqlStr, catID, pageSize, offset)
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

	c.JSON(200, gin.H{
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
// getPointDetailData 统一的知识点详情获取逻辑（内部函数）
// 用于主页和集合页面共享数据获取逻辑
// =================================================================================
func getPointDetailData(pointID int) (gin.H, error) {
	// 1. 获取知识点详情
	var point struct {
		ID              int    `json:"id"`
		CategoryID      int    `json:"categoryId"`
		Title           string `json:"title"`
		Content         string `json:"content"`
		ReferenceLinks  string `json:"referenceLinks"`
		LocalImageNames string `json:"localImageNames"`
		UpdateTime      string `json:"updateTime"`
		VideoUrl        string `json:"videoUrl"`
		Difficulty      int    `json:"difficulty"`
	}

	sqlStr := `
		SELECT 
			id, categorie_id, title, content, 
			COALESCE(reference_links, '[]') as reference_links,
			COALESCE(local_image_names, '[]') as local_image_names,
			update_time,
			COALESCE(video_url, '[]') as video_url,
			COALESCE(difficulty, 0) as difficulty
		FROM knowledge_points 
		WHERE id = ?
	`

	err := global.DB.QueryRow(sqlStr, pointID).Scan(
		&point.ID,
		&point.CategoryID,
		&point.Title,
		&point.Content,
		&point.ReferenceLinks,
		&point.LocalImageNames,
		&point.UpdateTime,
		&point.VideoUrl,
		&point.Difficulty,
	)

	if err != nil {
		return nil, err
	}

	// 2. 获取知识点的绑定关系
	bindingsSql := `
		SELECT 
			pb.id,
			pb.bind_text,
			pb.target_point_id,
			pb.target_subject_id,
			c.id as target_category_id,
			p.title as target_point_title,
			c.categorie_name as target_category_name
		FROM point_bindings pb
		JOIN knowledge_points p ON pb.target_point_id = p.id
		JOIN knowledge_categories c ON p.categorie_id = c.id
		WHERE pb.source_point_id = ?
		ORDER BY pb.create_time DESC
	`

	bindingRows, err := global.DB.Query(bindingsSql, pointID)
	if err != nil {
		// 绑定关系查询失败不影响主要功能
		return gin.H{
			"point":    point,
			"bindings": []gin.H{},
		}, nil
	}
	defer bindingRows.Close()

	var bindings []gin.H
	for bindingRows.Next() {
		var id, targetPointID, targetSubjectID, targetCategoryID int
		var bindText, targetPointTitle, targetCategoryName string
		err := bindingRows.Scan(&id, &bindText, &targetPointID, &targetSubjectID, &targetCategoryID, &targetPointTitle, &targetCategoryName)
		if err == nil {
			bindings = append(bindings, gin.H{
				"id":                 id,
				"bindText":           bindText,
				"targetPointId":      targetPointID,
				"targetSubjectId":    targetSubjectID,
				"targetCategoryId":   targetCategoryID,
				"targetPointTitle":   targetPointTitle,
				"targetCategoryName": targetCategoryName,
			})
		}
	}

	if bindings == nil {
		bindings = []gin.H{}
	}

	return gin.H{
		"point":    point,
		"bindings": bindings,
	}, nil
}

// =================================================================================
// GetPointDetail 获取知识点详情
// =================================================================================
func GetPointDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ID参数错误"})
		return
	}

	userID, _ := c.Get("userID")

	// 1. 权限校验：验证用户是否有权访问该知识点所属的科目
	var hasPerm int
	checkPermSQL := `
		SELECT 1
		FROM knowledge_points p
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN user_subjects us ON c.subject_id = us.subject_id
		WHERE p.id = ? AND us.user_id = ?
	`
	err = global.DB.QueryRow(checkPermSQL, id, userID).Scan(&hasPerm)
	if err != nil || hasPerm != 1 {
		c.JSON(403, gin.H{"code": 403, "msg": "无权查看该知识点"})
		return
	}

	// 2. 权限验证通过，调用统一的详情获取函数
	data, err := getPointDetailData(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该知识点"})
		} else {
			global.GetLog(c).Errorf("查询知识点详情失败 (ID: %d): %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询详情失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": data,
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

	// 1. 获取当前分类下已有的知识点数量，用于生成序号
	var count int
	err = global.DB.QueryRow("SELECT COUNT(*) FROM knowledge_points WHERE categorie_id = ?", req.CategoryID).Scan(&count)
	if err != nil {
		global.GetLog(c).Errorf("统计知识点数量失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "系统错误"})
		return
	}

	// 2. 生成带序号的新标题 (例如: "5. 我的新知识点")
	// count+1 代表当前是第几个
	newTitle := fmt.Sprintf("%d. %s", count+1, req.Title)

	// 3. 计算排序值 (保留你原有的逻辑，如果你希望按序号正序排，这里可能需要调整，暂保持原样)
	var currentMin int
	row := global.DB.QueryRow("SELECT COALESCE(MIN(sort_order), 0) FROM knowledge_points WHERE categorie_id = ?", req.CategoryID)
	row.Scan(&currentMin)
	newSortOrder := currentMin - 1

	// 4. 插入数据库 (使用 newTitle 和 req.Difficulty)
	res, err := global.DB.Exec(
		"INSERT INTO knowledge_points (categorie_id, title, content, sort_order, difficulty) VALUES (?, ?, '', ?, ?)",
		req.CategoryID, newTitle, newSortOrder, req.Difficulty,
	)

	if err != nil {
		// ★★★ Error ★★★
		global.GetLog(c).Errorf("创建知识点DB错误: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "创建失败"})
		return
	}

	id, _ := res.LastInsertId()
	// ★★★ Info ★★★
	global.GetLog(c).Infof("用户[%s] 创建知识点成功: ID=%d, Title=%s", currentUserCodeStr, id, newTitle)
	c.JSON(200, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{"id": id, "title": newTitle}})
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

	// --- 权限检查 & 获取当前数据 ---
	var subjectCreatorCode string
	var currentSubjectId int
	var currentTitle string   // 新增：当前数据库中的标题
	var currentCategoryId int // 新增：当前数据库中的分类ID
	var creatorName string
	var creatorEmail sql.NullString

	// 修改 SQL：多查了 p.title 和 p.categorie_id
	checkSQL := `
		SELECT s.creator_code, s.id, p.title, p.categorie_id, IFNULL(u.nickname, u.username), u.email
		FROM knowledge_points p
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN subjects s ON c.subject_id = s.id
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE p.id = ?
	`
	// Scan 增加变量接收
	err := global.DB.QueryRow(checkSQL, id).Scan(&subjectCreatorCode, &currentSubjectId, &currentTitle, &currentCategoryId, &creatorName, &creatorEmail)
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

	// 1. 处理分类移动逻辑
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

		// 同步更新集合中该知识点的分类ID
		_, err = global.DB.Exec("UPDATE collection_items SET category_id = ? WHERE point_id = ?", *req.CategoryID, id)
		if err != nil {
			global.GetLog(c).Warnf("更新集合中知识点分类ID失败 (PointID: %s): %v", id, err)
			// 不影响主流程，只记录警告
		}
	}

	// 2. 处理常规字段 - 重点修改 Title 逻辑
	if req.Title != "" {
		// 定义正则：匹配开头的 "数字." (例如 "5." 或 "12.")
		// ^(\d+\.) 匹配开头的一个或多个数字加点，\s* 匹配可能的空格
		re := regexp.MustCompile(`^(\d+\.)\s*`)

		// [第一步] 清洗用户输入
		// 无论用户输入 "新的标题" 还是 "6. 新的标题"，都只保留 "新的标题"
		cleanUserTitle := re.ReplaceAllString(req.Title, "")
		// 防止用户把标题删光了只留了数字，导致空串
		if cleanUserTitle == "" {
			cleanUserTitle = req.Title // 回退
		}

		// [第二步] 分析数据库里的旧标题
		oldMatches := re.FindStringSubmatch(currentTitle)
		var finalTitle string

		if len(oldMatches) > 1 {
			// --- 情况 A: 旧标题有序号 (例如 "5.") ---
			// oldMatches[1] 就是 "5."
			// 强制使用旧序号 + 清洗后的新标题
			finalTitle = fmt.Sprintf("%s %s", oldMatches[1], cleanUserTitle)
		} else {
			// --- 情况 B: 旧标题没有序号 ---
			// 需要自动生成序号。
			// 逻辑：统计当前分类下有多少个知识点（包含自己），作为序号。
			// 注意：因为自己已经在数据库里了，所以 COUNT(*) 是包含自己的。
			var count int
			global.DB.QueryRow("SELECT COUNT(*) FROM knowledge_points WHERE categorie_id = ?", currentCategoryId).Scan(&count)

			// 如果 count 是 5，即生成 "5. 新标题"
			finalTitle = fmt.Sprintf("%d. %s", count, cleanUserTitle)
		}

		query += ", title = ?"
		args = append(args, finalTitle)
	}

	if req.Content != "" {
		query += ", content = ?"
		args = append(args, req.Content)
	}
	if req.ReferenceLinks != "" {
		query += ", reference_links = ?"
		args = append(args, req.ReferenceLinks)
	}

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
// =================================================================================
// SearchPoints 知识点模糊搜索
// 只返回用户有阅读权限或是创建者的知识点
// 性能优化版本
// =================================================================================
func SearchPoints(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))

	// 1. 关键词长度限制
	if keyword == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "请输入搜索关键词"})
		return
	}
	if len(keyword) > 50 {
		keyword = keyword[:50] // 截断过长关键词
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 2. 模糊搜索参数
	searchPattern := "%" + keyword + "%"

	// 3. 优化SQL：使用子查询先过滤用户有权限的科目，减少JOIN范围
	sqlStr := `
		SELECT 
			p.id,
			p.title,
			c.id,
			c.categorie_name,
			s.id,
			s.name
		FROM knowledge_points p
		INNER JOIN knowledge_categories c ON p.categorie_id = c.id
		INNER JOIN subjects s ON c.subject_id = s.id
		WHERE p.title LIKE ?
		  AND s.id IN (SELECT subject_id FROM user_subjects WHERE user_id = ?)
		ORDER BY s.name, c.categorie_name, p.id DESC
		LIMIT 50
	`

	rows, err := global.DB.Query(sqlStr, searchPattern, userID)
	if err != nil {
		global.GetLog(c).Errorf("搜索知识点失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "搜索失败"})
		return
	}
	defer rows.Close()

	// 4. 使用结构体代替 gin.H，提高序列化性能
	type SearchResult struct {
		PointId      int    `json:"pointId"`
		PointTitle   string `json:"pointTitle"`
		CategoryId   int    `json:"categoryId"`
		CategoryName string `json:"categoryName"`
		SubjectId    int    `json:"subjectId"`
		SubjectName  string `json:"subjectName"`
	}

	// 5. 预分配切片容量，减少内存分配
	list := make([]SearchResult, 0, 50)

	for rows.Next() {
		var item SearchResult
		err := rows.Scan(
			&item.PointId,
			&item.PointTitle,
			&item.CategoryId,
			&item.CategoryName,
			&item.SubjectId,
			&item.SubjectName,
		)
		if err != nil {
			continue
		}
		list = append(list, item)
	}

	// 6. 检查迭代错误
	if err = rows.Err(); err != nil {
		global.GetLog(c).Errorf("搜索知识点迭代错误: %v", err)
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": list,
	})
}

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
