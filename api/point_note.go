package api

import (
	"database/sql"
	"fmt"
	"strconv"

	"practice_problems/global"

	"github.com/gin-gonic/gin"
)

// GetPointNote 获取知识点笔记
func GetPointNote(c *gin.Context) {
	pointIDStr := c.Param("id")
	if pointIDStr == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "知识点ID不能为空"})
		return
	}

	pointID, err := strconv.Atoi(pointIDStr)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "知识点ID格式错误"})
		return
	}

	// 从上下文中获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	var note string
	var createTime, updateTime string
	err = global.DB.QueryRow(`
		SELECT note, create_time, update_time 
		FROM point_user_notes 
		WHERE point_id = ? AND user_id = ?
	`, pointID, userID).Scan(&note, &createTime, &updateTime)

	if err != nil {
		if err == sql.ErrNoRows {
			// 没有找到笔记，返回空笔记
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "success",
				"data": map[string]interface{}{
					"note":        "",
					"create_time": "",
					"update_time": "",
				},
			})
			return
		}
		global.GetLog(c).Errorf("查询知识点笔记失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": map[string]interface{}{
			"note":        note,
			"create_time": createTime,
			"update_time": updateTime,
		},
	})
}

// SavePointNote 保存知识点笔记
func SavePointNote(c *gin.Context) {
	pointIDStr := c.Param("id")
	if pointIDStr == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "知识点ID不能为空"})
		return
	}

	pointID, err := strconv.Atoi(pointIDStr)
	if err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "知识点ID格式错误"})
		return
	}

	// 从上下文中获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	// 获取请求参数
	var req struct {
		Note string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": fmt.Sprintf("参数错误: %v", err)})
		return
	}

	// 检查知识点是否存在
	var pointExists bool
	err = global.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM knowledge_points WHERE id = ?)", pointID).Scan(&pointExists)
	if err != nil {
		global.GetLog(c).Errorf("检查知识点是否存在失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "检查知识点失败"})
		return
	}
	if !pointExists {
		c.JSON(400, gin.H{"code": 400, "msg": "知识点不存在"})
		return
	}

	// 检查是否已有笔记记录
	var noteExists bool
	err = global.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM point_user_notes WHERE point_id = ? AND user_id = ?)", pointID, userID).Scan(&noteExists)
	if err != nil {
		global.GetLog(c).Errorf("检查笔记是否存在失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "检查笔记失败"})
		return
	}

	// 插入或更新笔记
	if noteExists {
		// 更新笔记
		_, err = global.DB.Exec(`
			UPDATE point_user_notes 
			SET note = ?, update_time = CURRENT_TIMESTAMP 
			WHERE point_id = ? AND user_id = ?
		`, req.Note, pointID, userID)
	} else {
		// 插入新笔记
		_, err = global.DB.Exec(`
			INSERT INTO point_user_notes (point_id, user_id, note) 
			VALUES (?, ?, ?)
		`, pointID, userID, req.Note)
	}

	if err != nil {
		global.GetLog(c).Errorf("保存知识点笔记失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "保存失败"})
		return
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "保存成功",
	})
}
