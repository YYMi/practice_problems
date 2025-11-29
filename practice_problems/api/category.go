package api

import (
	"database/sql"
	"log"
	"net/http"
	"practice_problems/global" // 保持原样
	"practice_problems/model"  // 保持原样
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetCategoryList 获取分类列表
func GetCategoryList(c *gin.Context) {
	// 1. 尝试获取 subject_id 参数
	subjectIDStr := c.Query("subject_id")

	var rows *sql.Rows
	var err error

	// 2. 根据是否有 subject_id 决定执行哪条 SQL
	if subjectIDStr != "" {
		// 查询指定科目下的分类
		sqlStr := "SELECT id, subject_id, categorie_name, create_time, update_time FROM knowledge_categories WHERE subject_id = ?"
		rows, err = global.DB.Query(sqlStr, subjectIDStr)
	} else {
		// 查所有
		sqlStr := "SELECT id, subject_id, categorie_name, create_time, update_time FROM knowledge_categories"
		rows, err = global.DB.Query(sqlStr)
	}

	if err != nil {
		log.Println("查询分类失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	// 3. 组装数据
	// 建议初始化为非 nil 的切片，这样如果没有数据，返回的是 [] 而不是 null
	list := make([]model.KnowledgeCategory, 0)

	for rows.Next() {
		var item model.KnowledgeCategory
		// 因为 model 里的 CreateTime/UpdateTime 已经是 string 了，这里直接 Scan 不会报错
		err := rows.Scan(&item.ID, &item.SubjectID, &item.CategoryName, &item.CreateTime, &item.UpdateTime)
		if err != nil {
			log.Println("扫描数据失败:", err)
			continue
		}
		list = append(list, item)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": list})
}

// CreateCategory 创建分类
func CreateCategory(c *gin.Context) {
	var req model.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 插入数据
	sqlStr := "INSERT INTO knowledge_categories (subject_id, categorie_name) VALUES (?, ?)"
	result, err := global.DB.Exec(sqlStr, req.SubjectID, req.CategoryName)

	if err != nil {
		log.Println("创建分类失败:", err)
		// SQLite 如果开启了外键，SubjectID 不存在时也会报错
		if strings.Contains(err.Error(), "FOREIGN KEY") {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建失败：所属科目不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建失败"})
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{"id": newID}})
}

// UpdateCategory 修改分类名称
func UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req model.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	sqlStr := "UPDATE knowledge_categories SET categorie_name = ? WHERE id = ?"
	_, err := global.DB.Exec(sqlStr, req.CategoryName, id)
	if err != nil {
		log.Println("更新分类失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// DeleteCategory 删除分类
func DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	sqlStr := "DELETE FROM knowledge_categories WHERE id = ?"
	result, err := global.DB.Exec(sqlStr, id)

	if err != nil {
		log.Println("删除分类失败:", err)
		// 【优化点】：SQLite 外键约束拦截时的特殊处理
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败：该分类下仍有知识点，请先清理知识点"})
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
