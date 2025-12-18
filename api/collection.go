package api

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"practice_problems/global"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// =================================================================================
// 通用权限校验函数
// =================================================================================

// CollectionPermissionResult 权限校验结果
type CollectionPermissionResult struct {
	HasPermission bool   // 是否有权限
	IsOwner       bool   // 是否是所有者
	IsPublic      bool   // 是否是公有集合
	OwnerUserID   int    // 集合所有者ID
	OwnerUserCode string // 集合所有者Code
}

// CheckCollectionPermission 检查集合访问权限（通用方法）
// 权限规则：
// 1. 公有集合：所有人可访问
// 2. 私有集合：
//   - 所有者：永久有权限
//   - 被授权者：在授权时间内有权限
func CheckCollectionPermission(c *gin.Context, collectionID int) (*CollectionPermissionResult, error) {
	result := &CollectionPermissionResult{}

	// 获取当前用户信息
	userID, exists := c.Get("userID")
	if !exists {
		return nil, fmt.Errorf("未授权")
	}
	userCode, _ := c.Get("userCode")
	userCodeStr, _ := userCode.(string)

	// 查询集合信息
	var ownerUserID int
	var ownerUserCode string
	var isPublic int
	err := global.DB.QueryRow(`
		SELECT c.user_id, u.user_code, c.is_public
		FROM collections c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = ?
	`, collectionID).Scan(&ownerUserID, &ownerUserCode, &isPublic)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("集合不存在")
		}
		return nil, err
	}

	result.OwnerUserID = ownerUserID
	result.OwnerUserCode = ownerUserCode
	result.IsPublic = (isPublic == 1)

	// 判断是否是所有者
	result.IsOwner = (userID == ownerUserID)

	// 权限判断逻辑
	if result.IsPublic {
		// 公有集合：所有人都有权限
		result.HasPermission = true
		return result, nil
	}

	// 私有集合
	if result.IsOwner {
		// 所有者：有权限
		result.HasPermission = true
		return result, nil
	}

	// 非所有者：检查是否有授权
	var expireTime sql.NullString
	err = global.DB.QueryRow(`
		SELECT expire_time 
		FROM collection_permissions 
		WHERE collection_id = ? AND user_code = ?
	`, collectionID, userCodeStr).Scan(&expireTime)

	if err != nil {
		if err == sql.ErrNoRows {
			// 没有授权记录
			result.HasPermission = false
			return result, nil
		}
		return nil, err
	}

	// 检查授权是否过期
	if !expireTime.Valid {
		// NULL 表示永久有效
		result.HasPermission = true
	} else {
		// 解析过期时间，尝试多种格式
		var expireT time.Time
		var parseErr error

		// 尝试常见的日期时间格式
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			time.RFC3339,
			time.RFC3339Nano,
		}

		for _, format := range formats {
			expireT, parseErr = time.Parse(format, expireTime.String)
			if parseErr == nil {
				break
			}
		}

		// 如果所有格式都失败了，返回错误
		if parseErr != nil {
			return nil, fmt.Errorf("时间解析错误: 无法解析时间格式 '%s'", expireTime.String)
		}

		// 判断是否过期
		result.HasPermission = time.Now().Before(expireT)
	}
	return result, nil
}

// CreateCollectionRequest 创建集合请求
type CreateCollectionRequest struct {
	Name string `json:"name" binding:"required"`
}

// =================================================================================
// CreateCollection 创建集合
// =================================================================================
func CreateCollection(c *gin.Context) {
	var req CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 验证集合名称长度
	if len(req.Name) == 0 || len(req.Name) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "集合名称长度应在1-50个字符之间"})
		return
	}

	// 统计当前用户已有多少个集合，用于生成序号
	var count int
	err := global.DB.QueryRow("SELECT COUNT(*) FROM collections WHERE user_id = ?", userID).Scan(&count)
	if err != nil {
		global.GetLog(c).Errorf("统计集合数量失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}

	// 拼接序号和名称（例如: "1. 我的集合"）
	finalName := fmt.Sprintf("%d. %s", count+1, req.Name)

	// 检查完整名称是否重复
	var duplicateCount int
	err = global.DB.QueryRow("SELECT COUNT(*) FROM collections WHERE user_id = ? AND name = ?", userID, finalName).Scan(&duplicateCount)
	if err != nil {
		global.GetLog(c).Errorf("检查集合名称重复DB错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}
	if duplicateCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "集合名称已存在，请使用其他名称"})
		return
	}

	// 插入集合
	sqlStr := "INSERT INTO collections (name, user_id) VALUES (?, ?)"
	result, err := global.DB.Exec(sqlStr, finalName, userID)
	if err != nil {
		global.GetLog(c).Errorf("创建集合DB错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建失败"})
		return
	}

	newID, _ := result.LastInsertId()

	global.GetLog(c).Infof("用户[%v] 创建集合成功: ID=%d, Name=%s", userID, newID, finalName)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{
		"id":   newID,
		"name": finalName,
	}})
}

// =================================================================================
// GetCollections 获取用户的所有集合（包括自己的+公有的+授权的）
// =================================================================================
func GetCollections(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	userCode, _ := c.Get("userCode")
	userCodeStr, _ := userCode.(string)

	// 查询：自己的集合 + 公有集合 + 授权的私有集合
	sqlStr := `
		SELECT DISTINCT c.id, c.name, c.is_public, c.user_id, u.user_code, c.create_time, c.update_time
		FROM collections c
		JOIN users u ON c.user_id = u.id
		LEFT JOIN collection_permissions cp ON c.id = cp.collection_id AND cp.user_code = ?
		WHERE c.user_id = ?  -- 自己的集合
		   OR c.is_public = 1  -- 公有集合
		   OR (cp.user_code = ? AND (cp.expire_time IS NULL OR cp.expire_time > datetime('now', 'localtime')))  -- 授权且未过期
		ORDER BY c.create_time DESC
	`
	rows, err := global.DB.Query(sqlStr, userCodeStr, userID, userCodeStr)
	if err != nil {
		global.GetLog(c).Errorf("查询集合列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	var list []gin.H
	for rows.Next() {
		var id int64
		var name string
		var isPublic int
		var ownerUserID int
		var ownerUserCode string
		var createTime, updateTime string
		err := rows.Scan(&id, &name, &isPublic, &ownerUserID, &ownerUserCode, &createTime, &updateTime)
		if err != nil {
			continue
		}

		// 判断是否是所有者
		isOwner := (userID == ownerUserID)

		list = append(list, gin.H{
			"id":            id,
			"name":          name,
			"isPublic":      isPublic == 1,
			"isOwner":       isOwner,
			"ownerUserCode": ownerUserCode,
			"createTime":    createTime,
			"updateTime":    updateTime,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": list})
}

// =================================================================================
// DeleteCollection 删除集合
// =================================================================================
func DeleteCollection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ID参数错误"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 权限校验：只有所有者才能删除
	permResult, err := CheckCollectionPermission(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	if !permResult.IsOwner {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作：只有集合创建者才能删除"})
		return
	}

	// 开启事务
	tx, err := global.DB.Begin()
	if err != nil {
		global.GetLog(c).Errorf("删除集合开启事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}

	// 删除集合中的所有项
	_, err = tx.Exec("DELETE FROM collection_items WHERE collection_id = ?", id)
	if err != nil {
		tx.Rollback()
		global.GetLog(c).Errorf("删除集合项失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	// 删除集合
	result, err := tx.Exec("DELETE FROM collections WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		global.GetLog(c).Errorf("删除集合失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	if err := tx.Commit(); err != nil {
		global.GetLog(c).Errorf("删除集合事务提交失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	global.GetLog(c).Infof("用户[%v] 删除集合成功: ID=%d", userID, id)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// =================================================================================
// UpdateCollection 更新集合名称
// =================================================================================
func UpdateCollection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ID参数错误"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 权限校验：只有所有者才能修改
	permResult, err := CheckCollectionPermission(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	if !permResult.IsOwner {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作：只有集合创建者才能修改"})
		return
	}

	var req CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	if len(req.Name) == 0 || len(req.Name) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "集合名称长度应在1-50个字符之间"})
		return
	}

	// 获取当前集合的原名称，提取序号
	var oldName string
	err = global.DB.QueryRow("SELECT name FROM collections WHERE id = ?", id).Scan(&oldName)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		} else {
			global.GetLog(c).Errorf("查询集合失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		}
		return
	}

	// 使用正则提取序号（如 "1. 名称" 提取 "1. "）
	var prefix string
	for i, ch := range oldName {
		if ch == '.' && i > 0 && i < len(oldName)-1 && oldName[i+1] == ' ' {
			prefix = oldName[:i+2] // 包括序号、点和空格
			break
		}
	}

	// 拼接新名称（保留原序号）
	finalName := prefix + req.Name

	// 检查名称是否与其他集合重复（排除自己）
	var count int
	err = global.DB.QueryRow("SELECT COUNT(*) FROM collections WHERE user_id = ? AND name = ? AND id != ?",
		userID, finalName, id).Scan(&count)
	if err != nil {
		global.GetLog(c).Errorf("检查集合名称重复DB错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "集合名称已存在，请使用其他名称"})
		return
	}

	sqlStr := "UPDATE collections SET name = ? WHERE id = ?"
	result, err := global.DB.Exec(sqlStr, finalName, id)
	if err != nil {
		global.GetLog(c).Errorf("更新集合失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	global.GetLog(c).Infof("用户[%v] 更新集合成功: ID=%d, Name=%s", userID, id, finalName)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功", "data": gin.H{
		"id":   id,
		"name": finalName,
	}})
}

// =================================================================================
// AddPointToCollection 添加知识点到集合
// =================================================================================
func AddPointToCollection(c *gin.Context) {
	var req struct {
		CollectionID int `json:"collection_id" binding:"required"`
		PointID      int `json:"point_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 验证知识点是否属于当前用户创建的科目，同时获取 subject_id 和 category_id
	var creatorCode string
	var subjectID, categoryID int
	err := global.DB.QueryRow(`
		SELECT s.creator_code, s.id, c.id
		FROM knowledge_points p
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN subjects s ON c.subject_id = s.id
		WHERE p.id = ?
	`, req.PointID).Scan(&creatorCode, &subjectID, &categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "知识点不存在"})
		} else {
			global.GetLog(c).Errorf("查询知识点失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		}
		return
	}

	// 获取当前用户的user_code
	var userCode string
	err = global.DB.QueryRow("SELECT user_code FROM users WHERE id = ?", userID).Scan(&userCode)
	if err != nil {
		global.GetLog(c).Errorf("查询用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}

	// 检查知识点是否属于当前用户创建的科目
	if creatorCode != userCode {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "只能分享自己创建的知识点到集合"})
		return
	}

	// 验证集合是否属于当前用户
	var ownerID int
	err = global.DB.QueryRow("SELECT user_id FROM collections WHERE id = ?", req.CollectionID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		} else {
			global.GetLog(c).Errorf("查询集合失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		}
		return
	}

	if ownerID != userID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作该集合"})
		return
	}

	// 检查知识点是否已存在于集合中
	var count int
	err = global.DB.QueryRow("SELECT COUNT(*) FROM collection_items WHERE collection_id = ? AND point_id = ?",
		req.CollectionID, req.PointID).Scan(&count)
	if err != nil {
		global.GetLog(c).Errorf("查询集合项失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "该知识点已在集合中"})
		return
	}

	// 添加到集合，同时保存 subject_id 和 category_id
	// 获取当前最大的 sort_order
	var maxSortOrder int
	err = global.DB.QueryRow("SELECT IFNULL(MAX(sort_order), -1) FROM collection_items WHERE collection_id = ?", req.CollectionID).Scan(&maxSortOrder)
	if err != nil {
		global.GetLog(c).Errorf("查询最大排序值失败: %v", err)
		maxSortOrder = -1
	}

	// 新知识点排在最前面（sort_order 最大）
	sqlStr := "INSERT INTO collection_items (collection_id, point_id, subject_id, category_id, sort_order) VALUES (?, ?, ?, ?, ?)"
	_, err = global.DB.Exec(sqlStr, req.CollectionID, req.PointID, subjectID, categoryID, maxSortOrder+1)
	if err != nil {
		global.GetLog(c).Errorf("添加知识点到集合失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "添加失败"})
		return
	}

	global.GetLog(c).Infof("添加知识点到集合成功: CollectionID=%d, PointID=%d", req.CollectionID, req.PointID)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "分享成功"})
}

// =================================================================================
// GetCollectionPoints 获取集合中的知识点列表（支持分页）
// =================================================================================
func GetCollectionPoints(c *gin.Context) {
	idStr := c.Param("id")
	collectionID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ID参数错误"})
		return
	}

	// 获取用户ID
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 权限校验：检查是否有访问权限
	permResult, err := CheckCollectionPermission(c, collectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	if !permResult.HasPermission {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权查看该集合"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	// 允许更大的 pageSize 以支持随机模式获取全部数据
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 10000 {
		pageSize = 10000 // 最大限制 10000 条
	}
	offset := (page - 1) * pageSize

	// 查询总数
	var total int
	err = global.DB.QueryRow("SELECT COUNT(*) FROM collection_items WHERE collection_id = ?", collectionID).Scan(&total)
	if err != nil {
		global.GetLog(c).Errorf("查询集合项总数失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}

	// 查询知识点列表（直接使用 subject_id 和 category_id，不用JOIN）
	// 按 sort_order 降序排列（sort_order 大的在前）
	sqlStr := `
		SELECT 
			ci.id, ci.point_id, ci.subject_id, ci.category_id, ci.create_time, ci.sort_order,
			p.title, p.difficulty as point_difficulty,
			s.name as subject_name,
			c.categorie_name as category_name, c.difficulty as category_difficulty
		FROM collection_items ci
		JOIN knowledge_points p ON ci.point_id = p.id
		JOIN subjects s ON ci.subject_id = s.id
		JOIN knowledge_categories c ON ci.category_id = c.id
		WHERE ci.collection_id = ?
		ORDER BY ci.sort_order DESC, ci.create_time DESC
		LIMIT ? OFFSET ?
	`

	rows, err := global.DB.Query(sqlStr, collectionID, pageSize, offset)
	if err != nil {
		global.GetLog(c).Errorf("查询集合项列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}
	defer rows.Close()

	var list []gin.H
	for rows.Next() {
		var id, pointID, subjectID, categoryID, pointDifficulty, categoryDifficulty, sortOrder int
		var createTime, title, subjectName, categoryName string

		err := rows.Scan(&id, &pointID, &subjectID, &categoryID, &createTime, &sortOrder, &title, &pointDifficulty, &subjectName, &categoryName, &categoryDifficulty)
		if err != nil {
			continue
		}

		list = append(list, gin.H{
			"id":                 id,
			"pointId":            pointID,
			"subjectId":          subjectID,
			"categoryId":         categoryID,
			"title":              title,
			"subjectName":        subjectName,
			"categoryName":       categoryName,
			"pointDifficulty":    pointDifficulty,
			"categoryDifficulty": categoryDifficulty,
			"sortOrder":          sortOrder,
			"createTime":         createTime,
		})
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
// GetCollectionPointDetail 获取集合中知识点的详情
// 权限判断：1. 知识点必须在集合中 2. 集合必须有访问权限（所有者/公有/授权用户）
// =================================================================================
func GetCollectionPointDetail(c *gin.Context) {
	collectionIDStr := c.Param("id")
	collectionID, err := strconv.Atoi(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "集合ID参数错误"})
		return
	}

	pointIDStr := c.Param("pointId")
	pointID, err := strconv.Atoi(pointIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "知识点ID参数错误"})
		return
	}

	// 获取用户ID
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 1. 权限校验：检查是否有访问权限
	permResult, err := CheckCollectionPermission(c, collectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	if !permResult.HasPermission {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权访问该集合"})
		return
	}

	// 2. 验证知识点是否在集合中
	var itemID int
	err = global.DB.QueryRow(
		"SELECT id FROM collection_items WHERE collection_id = ? AND point_id = ?",
		collectionID, pointID,
	).Scan(&itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "该知识点不在当前集合中"})
		} else {
			global.GetLog(c).Errorf("验证知识点是否在集合中失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		}
		return
	}

	// 3. 权限验证通过，调用统一的详情获取函数
	data, err := getPointDetailData(pointID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "知识点不存在"})
		} else {
			global.GetLog(c).Errorf("查询知识点详情失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		}
		return
	}

	global.GetLog(c).Infof("用户 从集合[%d]获取知识点[%d]详情成功", collectionID, pointID)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": data,
	})
}

// =================================================================================
// RemovePointFromCollection 从集合中移除知识点
// =================================================================================
func RemovePointFromCollection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ID参数错误"})
		return
	}

	// 获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 权限校验：只有集合所有者才能移除知识点
	// 先获取集合项对应的集合ID
	var collectionID int
	err = global.DB.QueryRow("SELECT collection_id FROM collection_items WHERE id = ?", id).Scan(&collectionID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合项不存在"})
		} else {
			global.GetLog(c).Errorf("查询集合项失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		}
		return
	}

	// 检查权限
	permResult, err := CheckCollectionPermission(c, collectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	if !permResult.IsOwner {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作：只有集合创建者才能移除知识点"})
		return
	}

	// 删除集合项
	result, err := global.DB.Exec("DELETE FROM collection_items WHERE id = ?", id)
	if err != nil {
		global.GetLog(c).Errorf("删除集合项失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合项不存在"})
		return
	}

	global.GetLog(c).Infof("用户[%v] 从集合中移除知识点成功: ItemID=%d", userID, id)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "移除成功"})
}

// =================================================================================
// RemovePointByIds 通过collectionId和pointId从集合中移除知识点
// =================================================================================
func RemovePointByIds(c *gin.Context) {
	collectionIdStr := c.Param("id")
	pointIdStr := c.Param("pointId")

	collectionId, err := strconv.Atoi(collectionIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "集合ID参数错误"})
		return
	}

	pointId, err := strconv.Atoi(pointIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "知识点ID参数错误"})
		return
	}

	// 获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 权限校验：只有集合所有者才能移除知识点
	permResult, err := CheckCollectionPermission(c, collectionId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	if !permResult.IsOwner {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作：只有集合创建者才能移除知识点"})
		return
	}

	// 删除集合项
	result, err := global.DB.Exec("DELETE FROM collection_items WHERE collection_id = ? AND point_id = ?", collectionId, pointId)
	if err != nil {
		global.GetLog(c).Errorf("删除集合项失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合项不存在"})
		return
	}

	global.GetLog(c).Infof("用户[%v] 从集合[%d]中移除知识点[%d]成功", userID, collectionId, pointId)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "移除成功"})
}

// =================================================================================
// UpdateCollectionItemsOrder 批量更新集合项的排序
// =================================================================================
func UpdateCollectionItemsOrder(c *gin.Context) {
	var req struct {
		CollectionID int `json:"collection_id" binding:"required"`
		Items        []struct {
			ID        int `json:"id" binding:"required"`
			SortOrder int `json:"sort_order" binding:"required"`
		} `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 权限校验：只有集合所有者才能更新排序
	permResult, err := CheckCollectionPermission(c, req.CollectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	if !permResult.IsOwner {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作：只有集合创建者才能更新排序"})
		return
	}

	// 开启事务
	tx, err := global.DB.Begin()
	if err != nil {
		global.GetLog(c).Errorf("开启事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}

	// 批量更新排序
	for _, item := range req.Items {
		// 验证该项是否属于该集合
		var count int
		err := tx.QueryRow("SELECT COUNT(*) FROM collection_items WHERE id = ? AND collection_id = ?", item.ID, req.CollectionID).Scan(&count)
		if err != nil || count == 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的集合项ID"})
			return
		}

		// 更新排序
		_, err = tx.Exec("UPDATE collection_items SET sort_order = ? WHERE id = ?", item.SortOrder, item.ID)
		if err != nil {
			tx.Rollback()
			global.GetLog(c).Errorf("更新排序失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		global.GetLog(c).Errorf("提交事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	global.GetLog(c).Infof("用户[%v] 更新集合[%d]排序成功，共%d项", userID, req.CollectionID, len(req.Items))
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// =================================================================================
// GetCollectionQuestions 获取集合中所有知识点的题目（用于综合刷题）
// 支持随机获取指定数量的题目
// =================================================================================
func GetCollectionQuestions(c *gin.Context) {
	idStr := c.Param("id")
	collectionID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ID参数错误"})
		return
	}

	// 获取数量参数，默认20
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 200 {
		limit = 200 // 最大限制200题（性能考虑）
	}

	// 获取用户ID
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 转换 userID 为 int
	var userID int
	switch v := userIDVal.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	default:
		userID = 0
	}

	// 权限校验：检查是否有访问权限
	permResult, err := CheckCollectionPermission(c, collectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	if !permResult.HasPermission {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权访问该集合"})
		return
	}

	// 查询集合中所有知识点的ID
	pointIDsSQL := `
		SELECT DISTINCT point_id 
		FROM collection_items 
		WHERE collection_id = ?
		ORDER BY sort_order DESC
	`
	pointRows, err := global.DB.Query(pointIDsSQL, collectionID)
	if err != nil {
		global.GetLog(c).Errorf("查询集合知识点失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer pointRows.Close()

	var pointIDs []int
	for pointRows.Next() {
		var pointID int
		if err := pointRows.Scan(&pointID); err == nil {
			pointIDs = append(pointIDs, pointID)
		}
	}

	if len(pointIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": []interface{}{}})
		return
	}

	// 构建 IN 查询条件
	placeholders := ""
	args := []interface{}{}
	for i, pid := range pointIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, pid)
	}

	// 查询所有题目ID（轻量级查询）
	questionsIDSQL := fmt.Sprintf(`
		SELECT q.id
		FROM questions q
		WHERE q.knowledge_point_id IN (%s)
	`, placeholders)

	questionIDRows, err := global.DB.Query(questionsIDSQL, args...)
	if err != nil {
		global.GetLog(c).Errorf("查询题目ID失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer questionIDRows.Close()

	var allQuestionIDs []int
	for questionIDRows.Next() {
		var qid int
		if err := questionIDRows.Scan(&qid); err == nil {
			allQuestionIDs = append(allQuestionIDs, qid)
		}
	}

	if len(allQuestionIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": []interface{}{}})
		return
	}

	// 第一次打乱：洗牌算法打乱所有ID
	for i := len(allQuestionIDs) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		allQuestionIDs[i], allQuestionIDs[j] = allQuestionIDs[j], allQuestionIDs[i]
	}

	// 随机选取指定数量的ID
	selectedIDs := allQuestionIDs
	if limit < len(allQuestionIDs) {
		selectedIDs = allQuestionIDs[:limit]
	}

	// 构建选中题目的查询条件
	selectedPlaceholders := ""
	selectedArgs := []interface{}{userID} // 第一个参数是 userID
	for i, qid := range selectedIDs {
		if i > 0 {
			selectedPlaceholders += ","
		}
		selectedPlaceholders += "?"
		selectedArgs = append(selectedArgs, qid)
	}

	// 查询选中的题目详情，并 LEFT JOIN 用户备注表
	questionsSQL := fmt.Sprintf(`
		SELECT 
			q.id, q.knowledge_point_id, q.question_text,
			q.option1, q.option1_img, q.option2, q.option2_img,
			q.option3, q.option3_img, q.option4, q.option4_img,
			q.correct_answer, q.explanation,
			IFNULL(un.note, '') as user_note,
			q.create_time
		FROM questions q
		LEFT JOIN question_user_notes un ON q.id = un.question_id AND un.user_id = ?
		WHERE q.id IN (%s)
	`, selectedPlaceholders)

	questionRows, err := global.DB.Query(questionsSQL, selectedArgs...)
	if err != nil {
		global.GetLog(c).Errorf("查询题目列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer questionRows.Close()

	list := make([]gin.H, 0)
	for questionRows.Next() {
		var id, knowledgePointID, correctAnswer int
		var questionText, option1, option1Img, option2, option2Img string
		var option3, option3Img, option4, option4Img string
		var explanation, userNote, createTime string

		err := questionRows.Scan(
			&id, &knowledgePointID, &questionText,
			&option1, &option1Img, &option2, &option2Img,
			&option3, &option3Img, &option4, &option4Img,
			&correctAnswer, &explanation, &userNote, &createTime,
		)
		if err != nil {
			global.GetLog(c).Errorf("Scan error: %v", err)
			continue
		}

		list = append(list, gin.H{
			"id":               id,
			"knowledgePointId": knowledgePointID,
			"questionText":     questionText,
			"option1":          option1,
			"option1Img":       option1Img,
			"option2":          option2,
			"option2Img":       option2Img,
			"option3":          option3,
			"option3Img":       option3Img,
			"option4":          option4,
			"option4Img":       option4Img,
			"correctAnswer":    correctAnswer,
			"explanation":      explanation,
			"note":             userNote,
			"createTime":       createTime,
		})
	}

	global.GetLog(c).Infof("用户[%v] 获取集合[%d]题目成功，总共%d题，随机返回%d题", userID, collectionID, len(allQuestionIDs), len(list))
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": list})
}

// =================================================================================
// BatchAddPointsToCollection 批量添加知识点到集合（支持科目/分类级别）
// =================================================================================
func BatchAddPointsToCollection(c *gin.Context) {
	var req struct {
		CollectionID int `json:"collection_id" binding:"required"`
		SubjectID    int `json:"subject_id"`  // 科目ID，传这个则分享整个科目
		CategoryID   int `json:"category_id"` // 分类ID，传这个则分享整个分类
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 必须传入科目ID或分类ID
	if req.SubjectID == 0 && req.CategoryID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请传入科目ID或分类ID"})
		return
	}

	// 获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 获取当前用户的user_code
	var userCode string
	err := global.DB.QueryRow("SELECT user_code FROM users WHERE id = ?", userID).Scan(&userCode)
	if err != nil {
		global.GetLog(c).Errorf("查询用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}

	// 验证集合是否属于当前用户
	var ownerID int
	err = global.DB.QueryRow("SELECT user_id FROM collections WHERE id = ?", req.CollectionID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		} else {
			global.GetLog(c).Errorf("查询集合失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		}
		return
	}

	if ownerID != userID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作该集合"})
		return
	}

	// 根据科目ID或分类ID查询知识点
	var pointsSQL string
	var args []interface{}

	if req.CategoryID > 0 {
		// 按分类分享
		// 验证分类是否属于当前用户的科目
		var creatorCode string
		var subjectID int
		err = global.DB.QueryRow(`
			SELECT s.creator_code, s.id
			FROM knowledge_categories c
			JOIN subjects s ON c.subject_id = s.id
			WHERE c.id = ?
		`, req.CategoryID).Scan(&creatorCode, &subjectID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "分类不存在"})
			} else {
				global.GetLog(c).Errorf("查询分类失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
			}
			return
		}

		if creatorCode != userCode {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "只能分享自己创建的知识点到集合"})
			return
		}

		// 查询该分类下所有知识点（排除已在集合中的）
		pointsSQL = `
			SELECT p.id, c.subject_id, p.categorie_id
			FROM knowledge_points p
			JOIN knowledge_categories c ON p.categorie_id = c.id
			WHERE p.categorie_id = ?
			AND p.id NOT IN (
				SELECT point_id FROM collection_items WHERE collection_id = ?
			)
		`
		args = []interface{}{req.CategoryID, req.CollectionID}
	} else {
		// 按科目分享
		// 验证科目是否属于当前用户
		var creatorCode string
		err = global.DB.QueryRow("SELECT creator_code FROM subjects WHERE id = ?", req.SubjectID).Scan(&creatorCode)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "科目不存在"})
			} else {
				global.GetLog(c).Errorf("查询科目失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
			}
			return
		}

		if creatorCode != userCode {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "只能分享自己创建的知识点到集合"})
			return
		}

		// 查询该科目下所有知识点（排除已在集合中的）
		pointsSQL = `
			SELECT p.id, c.subject_id, p.categorie_id
			FROM knowledge_points p
			JOIN knowledge_categories c ON p.categorie_id = c.id
			WHERE c.subject_id = ?
			AND p.id NOT IN (
				SELECT point_id FROM collection_items WHERE collection_id = ?
			)
		`
		args = []interface{}{req.SubjectID, req.CollectionID}
	}

	// 查询要添加的知识点
	rows, err := global.DB.Query(pointsSQL, args...)
	if err != nil {
		global.GetLog(c).Errorf("查询知识点失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}
	defer rows.Close()

	type pointInfo struct {
		PointID    int
		SubjectID  int
		CategoryID int
	}

	var points []pointInfo
	for rows.Next() {
		var p pointInfo
		if err := rows.Scan(&p.PointID, &p.SubjectID, &p.CategoryID); err != nil {
			continue
		}
		points = append(points, p)
	}

	if len(points) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "没有新的知识点需要添加（可能已全部在集合中）", "data": gin.H{"added": 0}})
		return
	}

	// 获取当前最大的 sort_order
	var maxSortOrder int
	err = global.DB.QueryRow("SELECT IFNULL(MAX(sort_order), -1) FROM collection_items WHERE collection_id = ?", req.CollectionID).Scan(&maxSortOrder)
	if err != nil {
		global.GetLog(c).Errorf("查询最大排序值失败: %v", err)
		maxSortOrder = -1
	}

	// 开启事务批量插入
	tx, err := global.DB.Begin()
	if err != nil {
		global.GetLog(c).Errorf("开启事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}

	addedCount := 0
	for _, p := range points {
		maxSortOrder++
		_, err := tx.Exec(
			"INSERT INTO collection_items (collection_id, point_id, subject_id, category_id, sort_order) VALUES (?, ?, ?, ?, ?)",
			req.CollectionID, p.PointID, p.SubjectID, p.CategoryID, maxSortOrder,
		)
		if err != nil {
			global.GetLog(c).Errorf("插入集合项失败: %v", err)
			continue
		}
		addedCount++
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		global.GetLog(c).Errorf("提交事务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "添加失败"})
		return
	}

	global.GetLog(c).Infof("用户[%v] 批量添加知识点到集合[%d]成功: 添加%d个", userID, req.CollectionID, addedCount)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": fmt.Sprintf("成功添加 %d 个知识点到集合", addedCount), "data": gin.H{"added": addedCount}})
}

// =================================================================================
// SetCollectionPermission 设置集合权限（公有/私有）
// =================================================================================
func SetCollectionPermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ID参数错误"})
		return
	}

	// 权限校验：只有所有者才能设置权限
	permResult, err := CheckCollectionPermission(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	if !permResult.IsOwner {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作：只有集合创建者才能设置权限"})
		return
	}

	var req struct {
		IsPublic bool `json:"isPublic"` // 不用 required，因为 false 会被认为零值导致验证失败
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	isPublicInt := 0
	if req.IsPublic {
		isPublicInt = 1
	}

	sqlStr := "UPDATE collections SET is_public = ? WHERE id = ?"
	_, err = global.DB.Exec(sqlStr, isPublicInt, id)
	if err != nil {
		global.GetLog(c).Errorf("设置集合权限失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "设置失败"})
		return
	}

	global.GetLog(c).Infof("用户[%v] 设置集合[%d]权限成功: isPublic=%v", permResult.OwnerUserID, id, req.IsPublic)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "设置成功"})
}

// =================================================================================
// AddCollectionPermission 添加集合授权
// =================================================================================
func AddCollectionPermission(c *gin.Context) {
	collectionIDStr := c.Param("id")
	collectionID, err := strconv.Atoi(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "集合ID参数错误"})
		return
	}

	// 权限校验：只有所有者才能添加授权
	permResult, err := CheckCollectionPermission(c, collectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	if !permResult.IsOwner {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作：只有集合创建者才能添加授权"})
		return
	}

	var req struct {
		UserCode   string `json:"userCode" binding:"required"`
		ExpireTime string `json:"expireTime"` // 可选，格式: "2006-01-02 15:04:05" 或空表示永久
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 验证用户code是否存在
	var targetUserID int
	err = global.DB.QueryRow("SELECT id FROM users WHERE user_code = ?", req.UserCode).Scan(&targetUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "用户code不存在"})
		} else {
			global.GetLog(c).Errorf("查询用户失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		}
		return
	}

	// 检查是否试图给自己授权（无意义）
	if req.UserCode == permResult.OwnerUserCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "不能给自己授权"})
		return
	}

	// 解析过期时间
	var expireTime sql.NullString
	if req.ExpireTime != "" {
		_, parseErr := time.Parse("2006-01-02 15:04:05", req.ExpireTime)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "过期时间格式错误，应为 YYYY-MM-DD HH:MM:SS"})
			return
		}
		expireTime = sql.NullString{String: req.ExpireTime, Valid: true}
	}

	// 插入或更新授权记录
	sqlStr := `
		INSERT INTO collection_permissions (collection_id, user_code, expire_time)
		VALUES (?, ?, ?)
		ON CONFLICT(collection_id, user_code) DO UPDATE SET
		expire_time = excluded.expire_time,
		update_time = CURRENT_TIMESTAMP
	`

	_, err = global.DB.Exec(sqlStr, collectionID, req.UserCode, expireTime)
	if err != nil {
		global.GetLog(c).Errorf("添加集合授权失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "授权失败"})
		return
	}

	global.GetLog(c).Infof("用户[%v] 为集合[%d]添加授权成功: UserCode=%s, ExpireTime=%v",
		permResult.OwnerUserID, collectionID, req.UserCode, req.ExpireTime)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "授权成功"})
}

// =================================================================================
// GetCollectionPermissions 获取集合授权列表（支持分页）
// =================================================================================
func GetCollectionPermissions(c *gin.Context) {
	collectionIDStr := c.Param("id")
	collectionID, err := strconv.Atoi(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "集合ID参数错误"})
		return
	}

	// 权限校验：只有所有者才能查看授权列表
	permResult, err := CheckCollectionPermission(c, collectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	if !permResult.IsOwner {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作：只有集合创建者才能查看授权列表"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.DefaultQuery("search", "")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// 查询总数（带搜索条件）
	var total int
	var countErr error
	if search != "" {
		searchPattern := "%" + search + "%"
		countErr = global.DB.QueryRow("SELECT COUNT(*) FROM collection_permissions WHERE collection_id = ? AND user_code LIKE ?", collectionID, searchPattern).Scan(&total)
	} else {
		countErr = global.DB.QueryRow("SELECT COUNT(*) FROM collection_permissions WHERE collection_id = ?", collectionID).Scan(&total)
	}
	if countErr != nil {
		global.GetLog(c).Errorf("查询授权总数失败: %v", countErr)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}

	// 查询授权列表（分页+搜索）
	var sqlStr string
	var rows *sql.Rows
	var queryErr error
	if search != "" {
		searchPattern := "%" + search + "%"
		sqlStr = `
			SELECT cp.user_code, u.nickname, u.email, cp.expire_time, cp.create_time
			FROM collection_permissions cp
			LEFT JOIN users u ON cp.user_code = u.user_code
			WHERE cp.collection_id = ? AND cp.user_code LIKE ?
			ORDER BY cp.create_time DESC
			LIMIT ? OFFSET ?
		`
		rows, queryErr = global.DB.Query(sqlStr, collectionID, searchPattern, pageSize, offset)
	} else {
		sqlStr = `
			SELECT cp.user_code, u.nickname, u.email, cp.expire_time, cp.create_time
			FROM collection_permissions cp
			LEFT JOIN users u ON cp.user_code = u.user_code
			WHERE cp.collection_id = ?
			ORDER BY cp.create_time DESC
			LIMIT ? OFFSET ?
		`
		rows, queryErr = global.DB.Query(sqlStr, collectionID, pageSize, offset)
	}
	if queryErr != nil {
		global.GetLog(c).Errorf("查询集合授权列表失败: %v", queryErr)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	var list []gin.H
	for rows.Next() {
		var userCode, createTime string
		var nickname, email, expireTime sql.NullString

		err := rows.Scan(&userCode, &nickname, &email, &expireTime, &createTime)
		if err != nil {
			global.GetLog(c).Errorf("Scan授权记录失败: %v", err)
			continue
		}

		// 处理NULL值
		nicknameStr := ""
		if nickname.Valid {
			nicknameStr = nickname.String
		}
		emailStr := ""
		if email.Valid {
			emailStr = email.String
		}
		expireTimeStr := ""
		if expireTime.Valid {
			expireTimeStr = expireTime.String
		}

		list = append(list, gin.H{
			"userCode":   userCode,
			"nickname":   nicknameStr,
			"email":      emailStr,
			"expireTime": expireTimeStr,
			"createTime": createTime,
		})
	}

	// 确保 list 不为 null
	if list == nil {
		list = []gin.H{}
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
// UpdateCollectionPermission 更新集合授权时间
// =================================================================================
func UpdateCollectionPermission(c *gin.Context) {
	collectionIDStr := c.Param("id")
	collectionID, err := strconv.Atoi(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "集合ID参数错误"})
		return
	}

	// 权限校验：只有所有者才能更新授权
	permResult, err := CheckCollectionPermission(c, collectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	if !permResult.IsOwner {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作：只有集合创建者才能更新授权"})
		return
	}

	var req struct {
		UserCode   string `json:"userCode" binding:"required"`
		ExpireTime string `json:"expireTime"` // 可选，格式: "2006-01-02 15:04:05" 或空表示永久
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 检查授权记录是否存在
	var count int
	err = global.DB.QueryRow("SELECT COUNT(*) FROM collection_permissions WHERE collection_id = ? AND user_code = ?",
		collectionID, req.UserCode).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "授权记录不存在"})
		return
	}

	// 解析过期时间
	var expireTime sql.NullString
	if req.ExpireTime != "" {
		_, parseErr := time.Parse("2006-01-02 15:04:05", req.ExpireTime)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "过期时间格式错误，应为 YYYY-MM-DD HH:MM:SS"})
			return
		}
		expireTime = sql.NullString{String: req.ExpireTime, Valid: true}
	}

	// 更新授权时间
	sqlStr := "UPDATE collection_permissions SET expire_time = ?, update_time = CURRENT_TIMESTAMP WHERE collection_id = ? AND user_code = ?"
	_, err = global.DB.Exec(sqlStr, expireTime, collectionID, req.UserCode)
	if err != nil {
		global.GetLog(c).Errorf("更新集合授权时间失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	global.GetLog(c).Infof("用户[%v] 更新集合[%d]授权时间成功: UserCode=%s, ExpireTime=%v",
		permResult.OwnerUserID, collectionID, req.UserCode, req.ExpireTime)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// =================================================================================
// DeleteCollectionPermission 删除集合授权
// =================================================================================
func DeleteCollectionPermission(c *gin.Context) {
	collectionIDStr := c.Param("id")
	collectionID, err := strconv.Atoi(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "集合ID参数错误"})
		return
	}

	// 权限校验：只有所有者才能删除授权
	permResult, err := CheckCollectionPermission(c, collectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "集合不存在"})
		return
	}

	if !permResult.IsOwner {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作：只有集合创建者才能删除授权"})
		return
	}

	userCode := c.Query("userCode")
	if userCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少userCode参数"})
		return
	}

	// 检查是否试图删除自己的授权（不应该存在这种情况，但做个防护）
	if userCode == permResult.OwnerUserCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "不能删除自己的授权"})
		return
	}

	// 删除授权记录
	sqlStr := "DELETE FROM collection_permissions WHERE collection_id = ? AND user_code = ?"
	result, err := global.DB.Exec(sqlStr, collectionID, userCode)
	if err != nil {
		global.GetLog(c).Errorf("删除集合授权失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "授权记录不存在"})
		return
	}

	global.GetLog(c).Infof("用户[%v] 删除集合[%d]授权成功: UserCode=%s",
		permResult.OwnerUserID, collectionID, userCode)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// =================================================================================
// GetPointCollections 获取知识点已绑定的集合列表
// =================================================================================
func GetPointCollections(c *gin.Context) {
	pointIDStr := c.Query("point_id")
	if pointIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少point_id参数"})
		return
	}

	pointID, err := strconv.Atoi(pointIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "point_id参数错误"})
		return
	}

	// 获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// 查询该知识点已绑定的集合列表（只返回当前用户的集合）
	sqlStr := `
		SELECT c.id, c.name
		FROM collection_items ci
		JOIN collections c ON ci.collection_id = c.id
		WHERE ci.point_id = ? AND c.user_id = ?
		ORDER BY c.create_time DESC
	`

	rows, err := global.DB.Query(sqlStr, pointID, userID)
	if err != nil {
		global.GetLog(c).Errorf("查询知识点已绑定集合失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}
	defer rows.Close()

	var collections []gin.H
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		collections = append(collections, gin.H{
			"id":   id,
			"name": name,
		})
	}

	if collections == nil {
		collections = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": collections})
}

// =================================================================================
// FindPointInCollections 查找知识点在哪个集合中（用于绑定跳转）
// 优先级：1. 当前集合 2. 用户其他集合
// 返回：集合ID、知识点所在页码、该页的知识点列表
// =================================================================================
func FindPointInCollections(c *gin.Context) {
	pointIDStr := c.Query("point_id")
	currentCollectionIDStr := c.Query("current_collection_id")

	if pointIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少point_id参数"})
		return
	}

	pointID, err := strconv.Atoi(pointIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "point_id参数错误"})
		return
	}

	currentCollectionID := 0
	if currentCollectionIDStr != "" {
		currentCollectionID, _ = strconv.Atoi(currentCollectionIDStr)
	}

	// 获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	pageSize := 15 // 与前端保持一致

	// 1. 先查当前集合（如果有）
	if currentCollectionID > 0 {
		// 检查权限
		permResult, err := CheckCollectionPermission(c, currentCollectionID)
		if err == nil && permResult.HasPermission {
			// 查询知识点是否在该集合中
			var itemID int
			var sortOrder int
			err := global.DB.QueryRow(
				"SELECT id, sort_order FROM collection_items WHERE collection_id = ? AND point_id = ?",
				currentCollectionID, pointID,
			).Scan(&itemID, &sortOrder)

			if err == nil {
				// 找到了，计算分页信息
				page := findPageForPoint(currentCollectionID, pointID, pageSize)
				pointsData := getCollectionPointsPage(c, currentCollectionID, page, pageSize)

				if pointsData != nil {
					c.JSON(http.StatusOK, gin.H{
						"code": 200,
						"data": gin.H{
							"found":        true,
							"collectionId": currentCollectionID,
							"page":         page,
							"points":       pointsData["list"],
							"total":        pointsData["total"],
						},
					})
					return
				}
			}
		}
	}

	// 2. 当前集合没找到，查询用户其他集合
	sqlStr := `
		SELECT c.id
		FROM collection_items ci
		JOIN collections c ON ci.collection_id = c.id
		LEFT JOIN collection_permissions cp ON c.id = cp.collection_id AND cp.user_code = (
			SELECT user_code FROM users WHERE id = ?
		)
		WHERE ci.point_id = ? 
		  AND (c.user_id = ? OR c.is_public = 1 OR (
			  c.is_public = 0 AND cp.expire_time IS NOT NULL AND 
			  (cp.expire_time IS NULL OR cp.expire_time > datetime('now'))
		  ))
		ORDER BY 
			CASE WHEN c.user_id = ? THEN 0 ELSE 1 END,  -- 自己的集合优先
			c.create_time DESC
		LIMIT 1
	`

	var foundCollectionID int
	err = global.DB.QueryRow(sqlStr, userID, pointID, userID, userID).Scan(&foundCollectionID)

	if err != nil {
		if err == sql.ErrNoRows {
			// 没找到
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"data": gin.H{
					"found":   false,
					"message": "知识点不存在或作者未分享",
				},
			})
			return
		}

		global.GetLog(c).Errorf("查找知识点所在集合失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}

	// 找到了，计算分页信息
	page := findPageForPoint(foundCollectionID, pointID, pageSize)
	pointsData := getCollectionPointsPage(c, foundCollectionID, page, pageSize)

	if pointsData != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": gin.H{
				"found":        true,
				"collectionId": foundCollectionID,
				"page":         page,
				"points":       pointsData["list"],
				"total":        pointsData["total"],
			},
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
	}
}

// findPageForPoint 计算知识点在集合中的页码（按 sort_order 排序）
func findPageForPoint(collectionID int, pointID int, pageSize int) int {
	// 统计该知识点之前有多少个知识点（按 sort_order 升序）
	var count int
	err := global.DB.QueryRow(`
		SELECT COUNT(*)
		FROM collection_items ci1
		WHERE ci1.collection_id = ?
		  AND ci1.sort_order < (
			  SELECT ci2.sort_order 
			  FROM collection_items ci2 
			  WHERE ci2.collection_id = ? AND ci2.point_id = ?
		  )
	`, collectionID, collectionID, pointID).Scan(&count)

	if err != nil {
		return 1 // 错误情况返回第一页
	}

	// 计算页码（count 是从 0 开始，页码从 1 开始）
	page := (count / pageSize) + 1
	return page
}

// getCollectionPointsPage 获取集合的指定分页数据
func getCollectionPointsPage(c *gin.Context, collectionID int, page int, pageSize int) gin.H {
	offset := (page - 1) * pageSize

	// 查询总数
	var total int
	err := global.DB.QueryRow(
		"SELECT COUNT(*) FROM collection_items WHERE collection_id = ?",
		collectionID,
	).Scan(&total)

	if err != nil {
		global.GetLog(c).Errorf("查询集合知识点总数失败: %v", err)
		return nil
	}

	// 查询列表
	sqlStr := `
		SELECT 
			ci.id, ci.point_id, ci.sort_order, ci.create_time,
			p.title,
			p.difficulty as point_difficulty,
			c.difficulty as category_difficulty,
			s.id as subject_id, s.name as subject_name,
			c.id as category_id, c.categorie_name as category_name
		FROM collection_items ci
		JOIN knowledge_points p ON ci.point_id = p.id
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN subjects s ON c.subject_id = s.id
		WHERE ci.collection_id = ?
		ORDER BY ci.sort_order ASC
		LIMIT ? OFFSET ?
	`

	rows, err := global.DB.Query(sqlStr, collectionID, pageSize, offset)
	if err != nil {
		global.GetLog(c).Errorf("查询集合知识点列表失败: %v", err)
		return nil
	}
	defer rows.Close()

	var list []gin.H
	for rows.Next() {
		var id, pointID, subjectID, categoryID int
		var sortOrder int
		var createTime, title, subjectName, categoryName string
		var pointDifficulty, categoryDifficulty int

		err := rows.Scan(
			&id, &pointID, &sortOrder, &createTime,
			&title, &pointDifficulty, &categoryDifficulty,
			&subjectID, &subjectName, &categoryID, &categoryName,
		)

		if err == nil {
			list = append(list, gin.H{
				"id":                 id,
				"pointId":            pointID,
				"subjectId":          subjectID,
				"categoryId":         categoryID,
				"title":              title,
				"subjectName":        subjectName,
				"categoryName":       categoryName,
				"pointDifficulty":    pointDifficulty,
				"categoryDifficulty": categoryDifficulty,
				"sortOrder":          sortOrder,
				"createTime":         createTime,
			})
		}
	}

	if list == nil {
		list = []gin.H{}
	}

	return gin.H{
		"list":  list,
		"total": total,
	}
}
