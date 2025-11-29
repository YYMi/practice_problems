package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"practice_problems/global"
	"practice_problems/model"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetPointList 获取知识点列表
// 优化：SQL 中只查 id 和 title，绝对不查 content (为了列表加载速度)
func GetPointList(c *gin.Context) {
	catID := c.Query("category_id")
	if catID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "必须指定分类ID"})
		return
	}

	// SQLite 语法: SELECT id, title, create_time ...
	sqlStr := "SELECT id, title, create_time FROM knowledge_points WHERE categorie_id = ? ORDER BY id"

	rows, err := global.DB.Query(sqlStr, catID)
	if err != nil {
		log.Println("查询知识点列表失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	list := make([]gin.H, 0)
	for rows.Next() {
		var id int
		var title string
		var createTime string // SQLite 存的是字符串，直接用 string 接

		err := rows.Scan(&id, &title, &createTime)
		if err != nil {
			log.Println("Scan error:", err)
			continue
		}

		list = append(list, gin.H{
			"id":         id,
			"title":      title,
			"createTime": createTime,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": list,
	})
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

// CreatePoint 创建知识点
func CreatePoint(c *gin.Context) {
	var req model.CreatePointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 插入数据
	// 注意：SQLite 没有 NOW()，通常让数据库默认值处理 create_time，这里只插业务字段
	// content 默认为空字符串，避免 NULL
	sqlStr := "INSERT INTO knowledge_points (categorie_id, title, content) VALUES (?, ?, '')"

	res, err := global.DB.Exec(sqlStr, req.CategoryID, req.Title)

	if err != nil {
		log.Println("插入数据库失败:", err)
		// 检查外键约束 (分类ID是否存在)
		if strings.Contains(err.Error(), "FOREIGN KEY") {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建失败：所属分类不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建失败，数据库错误"})
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取ID失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{"id": id}})
}

// UpdatePoint 更新知识点
func UpdatePoint(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdatePointRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 动态构建 SQL
	// SQLite 也可以用 UPDATE ... SET update_time = CURRENT_TIMESTAMP
	// 但我们之前加了 Trigger (触发器)，所以这里其实【不用手动更新时间】，Trigger 会帮我们做
	// 不过为了双重保险，手动更新一下也没坏处

	// ⚠️ 注意：MySQL 是 NOW()，SQLite 是 CURRENT_TIMESTAMP 或 datetime('now', 'localtime')
	// 为了兼容最简单写法，我们这里利用 Trigger，或者直接传入 Go 的当前时间字符串

	query := "UPDATE knowledge_points SET update_time = ?"
	args := []interface{}{time.Now().Format("2006-01-02 15:04:05")} // 手动传时间字符串，最稳

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

	query += " WHERE id = ?"
	args = append(args, id)

	_, err := global.DB.Exec(query, args...)
	if err != nil {
		log.Println("更新数据库失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新数据库失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
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
