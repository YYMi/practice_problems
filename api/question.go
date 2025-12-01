package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"practice_problems/global"
	"practice_problems/model"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetQuestionList 获取题目列表
// 支持参数:
// 1. point_id: 获取指定知识点下的题目
// 2. category_id: 获取指定分类下的所有题目 (跨表查询)
func GetQuestionList(c *gin.Context) {
	pointID := c.Query("point_id")
	categoryID := c.Query("category_id")

	if pointID == "" && categoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "必须传递 point_id 或 category_id"})
		return
	}

	var rows *sql.Rows
	var err error
	var sqlStr string

	// 这里的 COALESCE 是标准 SQL，SQLite 和 MySQL 都支持
	// 作用是如果字段为 NULL，就返回空字符串 ''，这样 Scan 的时候就不用处理 sql.NullString 了
	selectFields := `
               SELECT q.id, q.knowledge_point_id, q.question_text, 
		       COALESCE(q.option1, ''), COALESCE(q.option1_img, ''), 
		       COALESCE(q.option2, ''), COALESCE(q.option2_img, ''), 
		       COALESCE(q.option3, ''), COALESCE(q.option3_img, ''), 
		       COALESCE(q.option4, ''), COALESCE(q.option4_img, ''), 
		       q.correct_answer, COALESCE(q.explanation, ''), COALESCE(q.note, ''), 
		       q.create_time, q.update_time 
	`

	if pointID != "" {
		// 逻辑 A: 查单个知识点
		sqlStr = selectFields + ` FROM questions q WHERE q.knowledge_point_id = ? ORDER BY q.create_time DESC`
		rows, err = global.DB.Query(sqlStr, pointID)
	} else {
		// 逻辑 B: 查全分类 (连表查询)
		// 你的表结构里知识点表关联分类的字段是 categorie_id (注意拼写)
		sqlStr = selectFields + ` 
			FROM questions q
			JOIN knowledge_points kp ON q.knowledge_point_id = kp.id
			WHERE kp.categorie_id = ? 
			ORDER BY q.create_time DESC`
		rows, err = global.DB.Query(sqlStr, categoryID)
	}

	if err != nil {
		log.Println("查询题目失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	// 初始化为空数组，避免返回 null
	list := make([]model.Question, 0)

	for rows.Next() {
		var q model.Question
		// 因为 SQL 里用了 COALESCE，这里可以直接 scan 到 string，不需要 sql.NullString
		// 且时间字段在 model 里已经是 string 了
		err := rows.Scan(
			&q.ID, &q.KnowledgePointID, &q.QuestionText,
			&q.Option1, &q.Option1Img, &q.Option2, &q.Option2Img,
			&q.Option3, &q.Option3Img, &q.Option4, &q.Option4Img,
			&q.CorrectAnswer, &q.Explanation, &q.Note,
			&q.CreateTime, &q.UpdateTime,
		)
		if err != nil {
			log.Println("Scan error:", err)
			continue
		}
		list = append(list, q)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": list})
}

// CreateQuestion 创建题目
func CreateQuestion(c *gin.Context) {
	var req model.CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	sqlStr := `
		INSERT INTO questions (
			knowledge_point_id, question_text, 
			option1, option1_img, option2, option2_img, 
			option3, option3_img, option4, option4_img, 
			correct_answer, explanation, note
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := global.DB.Exec(sqlStr,
		req.KnowledgePointID, req.QuestionText,
		req.Option1, req.Option1Img, req.Option2, req.Option2Img,
		req.Option3, req.Option3Img, req.Option4, req.Option4Img,
		req.CorrectAnswer, req.Explanation, req.Note,
	)

	if err != nil {
		log.Println("创建题目失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建失败"})
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{"id": newID}})
}

// UpdateQuestion 更新题目
func UpdateQuestion(c *gin.Context) {
	id := c.Param("id")
	var req model.UpdateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 动态构建 SQL
	// 适配 SQLite: 使用 ? 占位符手动传入时间，或者依赖 Trigger (如果有)
	// 这里为了最稳妥，手动传入 Go 格式化的当前时间
	query := "UPDATE questions SET update_time = ?"
	var args []interface{}
	args = append(args, time.Now().Format("2006-01-02 15:04:05"))

	if req.QuestionText != "" {
		query += ", question_text = ?"
		args = append(args, req.QuestionText)
	}

	// 选项 1
	if req.Option1 != "" {
		query += ", option1 = ?"
		args = append(args, req.Option1)
	}
	if req.Option1Img != "" {
		query += ", option1_img = ?"
		args = append(args, req.Option1Img)
	}

	// 选项 2
	if req.Option2 != "" {
		query += ", option2 = ?"
		args = append(args, req.Option2)
	}
	if req.Option2Img != "" {
		query += ", option2_img = ?"
		args = append(args, req.Option2Img)
	}

	// 选项 3
	if req.Option3 != "" {
		query += ", option3 = ?"
		args = append(args, req.Option3)
	}
	if req.Option3Img != "" {
		query += ", option3_img = ?"
		args = append(args, req.Option3Img)
	}

	// 选项 4
	if req.Option4 != "" {
		query += ", option4 = ?"
		args = append(args, req.Option4)
	}
	if req.Option4Img != "" {
		query += ", option4_img = ?"
		args = append(args, req.Option4Img)
	}

	// 正确答案 (大于0才更新)
	if req.CorrectAnswer > 0 {
		query += ", correct_answer = ?"
		args = append(args, req.CorrectAnswer)
	}

	// 解析和笔记
	if req.Explanation != "" {
		query += ", explanation = ?"
		args = append(args, req.Explanation)
	}
	if req.Note != "" {
		query += ", note = ?"
		args = append(args, req.Note)
	}

	query += " WHERE id = ?"
	args = append(args, id)

	_, err := global.DB.Exec(query, args...)
	if err != nil {
		fmt.Println("更新题目失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// DeleteQuestion 删除题目
func DeleteQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// 执行删除
	result, err := global.DB.Exec("DELETE FROM questions WHERE id = ?", id)
	if err != nil {
		log.Println("删除题目失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该题目"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}
