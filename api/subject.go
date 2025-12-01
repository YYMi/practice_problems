package api

import (
	"database/sql"
	"log"
	"net/http"
	"practice_problems/global"
	"practice_problems/model"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetSubjectList 获取科目列表接口
func GetSubjectList(c *gin.Context) {
	// 1. 编写 SQL 语句
	// 查询状态为 1 (启用) 的科目
	sqlStr := "SELECT id, name, status, create_time, update_time FROM subjects WHERE status = 1"

	// 2. 执行查询
	rows, err := global.DB.Query(sqlStr)
	if err != nil {
		log.Println("查询数据库失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "服务器内部错误",
		})
		return
	}
	defer rows.Close()

	// 3. 组装数据
	subjects := make([]model.Subject, 0)

	for rows.Next() {
		var s model.Subject
		// model 里的时间已经是 string 类型，可以直接 Scan
		err := rows.Scan(&s.ID, &s.Name, &s.Status, &s.CreateTime, &s.UpdateTime)
		if err != nil {
			log.Println("数据扫描失败:", err)
			continue
		}
		subjects = append(subjects, s)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": subjects,
	})
}

// GetSubjectDetail 获取单条详情
func GetSubjectDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ID参数错误"})
		return
	}

	var s model.Subject
	sqlStr := "SELECT id, name, status, create_time, update_time FROM subjects WHERE id = ?"
	err = global.DB.QueryRow(sqlStr, id).Scan(&s.ID, &s.Name, &s.Status, &s.CreateTime, &s.UpdateTime)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该科目"})
		} else {
			log.Println("查询详情失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": s})
}

// CreateSubject 创建科目
func CreateSubject(c *gin.Context) {
	var req model.CreateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 插入数据
	// 这里的 status (1/0) 直接插入 INTEGER 字段，没问题
	sqlStr := "INSERT INTO subjects (name, status) VALUES (?, ?)"
	result, err := global.DB.Exec(sqlStr, req.Name, req.Status)
	if err != nil {
		log.Println("创建失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建失败"})
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{"id": newID}})
}

// UpdateSubject 更新科目
func UpdateSubject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ID参数错误"})
		return
	}

	var req model.UpdateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 更新操作
	// 优化：手动更新 update_time，确保时间变动
	sqlStr := "UPDATE subjects SET name = ?, status = ?, update_time = ? WHERE id = ?"
	nowTime := time.Now().Format("2006-01-02 15:04:05")

	result, err := global.DB.Exec(sqlStr, req.Name, req.Status, nowTime, id)
	if err != nil {
		log.Println("更新失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到记录或数据未变更"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// DeleteSubject 删除科目 (软删除：仅修改状态为禁用)
func DeleteSubject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ID参数错误"})
		return
	}

	// 软删除
	// 优化：同样手动更新 update_time
	sqlStr := "UPDATE subjects SET status = 0, update_time = ? WHERE id = ?"
	nowTime := time.Now().Format("2006-01-02 15:04:05")

	result, err := global.DB.Exec(sqlStr, nowTime, id)
	if err != nil {
		log.Println("软删除失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该记录或已删除"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}
