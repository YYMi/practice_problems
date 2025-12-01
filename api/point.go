package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"practice_problems/global"
	"practice_problems/model"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetPointList 获取知识点列表 (精简版，不查 content)
func GetPointList(c *gin.Context) {
	catID := c.Query("category_id")
	if catID == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "必须指定分类ID"})
		return
	}

	// ✅ 修改：查询 sort_order 和 difficulty，并按 sort_order 排序
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
		var sortOrder int  // 新增
		var difficulty int // 新增

		// ✅ 修改：Scan 增加接收变量
		err := rows.Scan(&id, &title, &createTime, &sortOrder, &difficulty)
		if err != nil {
			continue
		}

		list = append(list, gin.H{
			"id":         id,
			"title":      title,
			"createTime": createTime,
			"sortOrder":  sortOrder,  // 返回给前端
			"difficulty": difficulty, // 返回给前端
		})
	}

	c.JSON(200, gin.H{"code": 200, "msg": "success", "data": list})
}

// GetPointDetail 获取知识点详情
// 只有在这里，我们才查询 content
func GetPointDetail(c *gin.Context) {
	id := c.Param("id")

	var p model.KnowledgePoint

	// SQLite 语法适配:
	// 1. 使用 COALESCE 替代 IFNULL (虽然 SQLite 也支持 IFNULL，但 COALESCE 更通用)
	// 2. 确保查出来的 JSON 字段如果是 NULL，就变成空数组 '[]'
	sqlStr := `SELECT id, categorie_id, title, content, 
	           COALESCE(reference_links, '[]'), COALESCE(local_image_names, '[]'), 
	           create_time, update_time 
	           FROM knowledge_points WHERE id = ?`

	// 注意：p.CreateTime 和 p.UpdateTime 在 model 里必须是 string 类型
	err := global.DB.QueryRow(sqlStr, id).Scan(
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

func CreatePoint(c *gin.Context) {
	var req model.CreatePointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 1. 计算新排序值 (当前最小值 - 1)，实现自动置顶
	var currentMin int
	row := global.DB.QueryRow("SELECT COALESCE(MIN(sort_order), 0) FROM knowledge_points WHERE categorie_id = ?", req.CategoryID)
	row.Scan(&currentMin)
	newSortOrder := currentMin - 1

	// 2. 插入数据 (包含 sort_order, difficulty 默认为 0)
	res, err := global.DB.Exec(
		"INSERT INTO knowledge_points (categorie_id, title, content, sort_order, difficulty) VALUES (?, ?, '', ?, 0)",
		req.CategoryID, req.Title, newSortOrder,
	)

	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "创建失败"})
		return
	}

	id, _ := res.LastInsertId()
	c.JSON(200, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{"id": id}})
}

func UpdatePoint(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdatePointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

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

	// ✅ 新增：修改难度
	if req.Difficulty != nil {
		if *req.Difficulty < 0 || *req.Difficulty > 3 {
			c.JSON(400, gin.H{"code": 400, "msg": "难度无效"})
			return
		}
		query += ", difficulty = ?"
		args = append(args, *req.Difficulty)
	}

	query += " WHERE id = ?"
	args = append(args, id)

	_, err := global.DB.Exec(query, args...)
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "更新失败"})
		return
	}
	c.JSON(200, gin.H{"code": 200, "msg": "更新成功"})
}

type UpdateSortRequest struct {
	Action string `json:"action"` // "top", "up", "down"
}

func UpdatePointSort(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req UpdateSortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	tx, err := global.DB.Begin()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "服务器错误"})
		return
	}

	var currentSort, catID int
	err = tx.QueryRow("SELECT sort_order, categorie_id FROM knowledge_points WHERE id = ?", id).Scan(&currentSort, &catID)
	if err != nil {
		tx.Rollback()
		c.JSON(404, gin.H{"code": 404, "msg": "知识点不存在"})
		return
	}

	if req.Action == "top" {
		var minSort int
		tx.QueryRow("SELECT MIN(sort_order) FROM knowledge_points WHERE categorie_id = ?", catID).Scan(&minSort)
		newSort := minSort - 1
		tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", newSort, id)
	} else if req.Action == "up" {
		var prevID, prevSort int
		err = tx.QueryRow("SELECT id, sort_order FROM knowledge_points WHERE categorie_id = ? AND sort_order < ? ORDER BY sort_order DESC LIMIT 1", catID, currentSort).Scan(&prevID, &prevSort)
		if err == nil {
			tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", prevSort, id)
			tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", currentSort, prevID)
		}
	} else if req.Action == "down" {
		var nextID, nextSort int
		err = tx.QueryRow("SELECT id, sort_order FROM knowledge_points WHERE categorie_id = ? AND sort_order > ? ORDER BY sort_order ASC LIMIT 1", catID, currentSort).Scan(&nextID, &nextSort)
		if err == nil {
			tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", nextSort, id)
			tx.Exec("UPDATE knowledge_points SET sort_order = ? WHERE id = ?", currentSort, nextID)
		}
	}

	tx.Commit()
	c.JSON(200, gin.H{"code": 200, "msg": "排序成功"})
}

// DeletePointImage 删除知识点图片
func DeletePointImage(c *gin.Context) {
	id := c.Param("id")

	var req model.DeletePointImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	var localImageNamesStr string
	// COALESCE 防止 NULL
	err := global.DB.QueryRow("SELECT COALESCE(local_image_names, '[]') FROM knowledge_points WHERE id = ?", id).Scan(&localImageNamesStr)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该知识点"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询数据失败"})
		}
		return
	}

	var images []string
	if err := json.Unmarshal([]byte(localImageNamesStr), &images); err != nil {
		images = []string{}
	}

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
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "该知识点下未找到指定图片"})
		return
	}

	newJsonBytes, _ := json.Marshal(newImages)
	newJsonStr := string(newJsonBytes)

	// 执行更新
	_, err = global.DB.Exec("UPDATE knowledge_points SET local_image_names = ? WHERE id = ?", newJsonStr, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新数据库失败"})
		return
	}

	// 物理删除文件
	err = RemoveFileFromDisk(req.FilePath)
	if err != nil {
		fmt.Println("警告: 数据库已更新，但本地文件删除失败:", err)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "图片删除成功"})
}

// DeletePoint 删除知识点
func DeletePoint(c *gin.Context) {
	id := c.Param("id")

	// 执行删除
	_, err := global.DB.Exec("DELETE FROM knowledge_points WHERE id = ?", id)

	if err != nil {
		log.Println("删除知识点失败:", err)
		// SQLite 外键约束检查
		// 如果这个知识点下面挂了题目 (questions 表)，删除会失败
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败：该知识点下仍有题目，请先删除题目"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}
