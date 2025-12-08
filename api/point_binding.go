package api

import (
	"net/http"
	"practice_problems/global"
	"practice_problems/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateBinding 创建知识点绑定
func CreateBinding(c *gin.Context) {
	var req model.CreateBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 获取当前用户信息
	userID, _ := c.Get("userID")
	currentUserCode, _ := c.Get("userCode")
	currentUserCodeStr, _ := currentUserCode.(string)

	// --- 权限校验：检查当前用户是否是源知识点的作者 ---
	var subjectCreatorCode string
	checkSQL := `
		SELECT s.creator_code
		FROM knowledge_points p
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN subjects s ON c.subject_id = s.id
		WHERE p.id = ?
	`
	err := global.DB.QueryRow(checkSQL, req.SourcePointID).Scan(&subjectCreatorCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "知识点不存在"})
		return
	}

	if subjectCreatorCode != currentUserCodeStr {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "绑定失败：您不是该科目的作者，无权绑定"})
		return
	}

	// 检查是否已存在相同绑定
	var count int
	err = global.DB.QueryRow(`
		SELECT COUNT(*) FROM point_bindings 
		WHERE source_point_id = ? AND target_point_id = ? AND bind_text = ?
	`, req.SourcePointID, req.TargetPointID, req.BindText).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "该绑定已存在"})
		return
	}

	// 插入绑定
	result, err := global.DB.Exec(`
		INSERT INTO point_bindings (source_subject_id, source_point_id, target_subject_id, target_point_id, bind_text, user_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`, req.SourceSubjectID, req.SourcePointID, req.TargetSubjectID, req.TargetPointID, req.BindText, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建绑定失败: " + err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "绑定成功", "data": gin.H{"id": id}})
}

// GetBindingsByPoint 获取知识点的所有绑定
func GetBindingsByPoint(c *gin.Context) {
	pointID := c.Param("pointId")

	rows, err := global.DB.Query(`
		SELECT 
			pb.id, pb.source_subject_id, pb.source_point_id, pb.target_subject_id, pb.target_point_id, 
			pb.bind_text, pb.user_id, pb.create_time,
			COALESCE(ss.name, '') as source_subject_name,
			COALESCE(sp.title, '') as source_point_title,
			COALESCE(ts.name, '') as target_subject_name,
			COALESCE(tp.title, '') as target_point_title
		FROM point_bindings pb
		LEFT JOIN subjects ss ON pb.source_subject_id = ss.id
		LEFT JOIN knowledge_points sp ON pb.source_point_id = sp.id
		LEFT JOIN subjects ts ON pb.target_subject_id = ts.id
		LEFT JOIN knowledge_points tp ON pb.target_point_id = tp.id
		WHERE pb.source_point_id = ?
		ORDER BY pb.create_time DESC
	`, pointID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败: " + err.Error()})
		return
	}
	defer rows.Close()

	var bindings []model.BindingWithDetails
	for rows.Next() {
		var b model.BindingWithDetails
		err := rows.Scan(
			&b.ID, &b.SourceSubjectID, &b.SourcePointID, &b.TargetSubjectID, &b.TargetPointID,
			&b.BindText, &b.UserID, &b.CreateTime,
			&b.SourceSubjectName, &b.SourcePointTitle, &b.TargetSubjectName, &b.TargetPointTitle,
		)
		if err != nil {
			continue
		}
		bindings = append(bindings, b)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": bindings})
}

// DeleteBinding 删除绑定
func DeleteBinding(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的ID"})
		return
	}

	currentUserCode, _ := c.Get("userCode")
	currentUserCodeStr, _ := currentUserCode.(string)

	// --- 权限校验：检查当前用户是否是源知识点的作者 ---
	var subjectCreatorCode string
	checkSQL := `
		SELECT s.creator_code
		FROM point_bindings pb
		JOIN knowledge_points p ON pb.source_point_id = p.id
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN subjects s ON c.subject_id = s.id
		WHERE pb.id = ?
	`
	err = global.DB.QueryRow(checkSQL, id).Scan(&subjectCreatorCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "绑定不存在"})
		return
	}

	if subjectCreatorCode != currentUserCodeStr {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "删除失败：您不是该科目的作者，无权删除"})
		return
	}

	_, err = global.DB.Exec("DELETE FROM point_bindings WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// GetCategoriesBySubjectForBinding 获取科目下的所有分类（用于绑定选择）
func GetCategoriesBySubjectForBinding(c *gin.Context) {
	subjectID := c.Param("subjectId")

	rows, err := global.DB.Query("SELECT id, categorie_name FROM knowledge_categories WHERE subject_id = ? ORDER BY sort_order", subjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	var categories []gin.H
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err == nil {
			categories = append(categories, gin.H{"id": id, "name": name})
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": categories})
}

// GetPointsByCategoryForBinding 获取分类下的所有知识点（用于绑定选择）
func GetPointsByCategoryForBinding(c *gin.Context) {
	categoryID := c.Param("categoryId")

	rows, err := global.DB.Query(`
		SELECT id, title FROM knowledge_points 
		WHERE categorie_id = ?
		ORDER BY sort_order
	`, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	var points []gin.H
	for rows.Next() {
		var id int
		var title string
		if err := rows.Scan(&id, &title); err == nil {
			points = append(points, gin.H{"id": id, "title": title})
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": points})
}
