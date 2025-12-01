package api

import (
	"database/sql"
	"log"
	_ "net/http"
	"practice_problems/global"
	"practice_problems/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

// =================================================================================
// GetQuestionList 获取某知识点下的题目列表
// 权限逻辑：只要用户绑定了该题目所属的科目 (user_subjects)，就有权查看
// =================================================================================
func GetQuestionList(c *gin.Context) {
	pointID := c.Query("point_id")
	if pointID == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "必须指定知识点ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// --- 1. 权限检查 ---
	// 路径：point -> category -> subject -> user_subjects
	var hasPerm int
	checkPermSQL := `
		SELECT 1 
		FROM knowledge_points p
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN user_subjects us ON c.subject_id = us.subject_id
		WHERE p.id = ? AND us.user_id = ?
	`
	err := global.DB.QueryRow(checkPermSQL, pointID, userID).Scan(&hasPerm)
	if err != nil || hasPerm != 1 {
		c.JSON(403, gin.H{"code": 403, "msg": "无权访问该内容"})
		return
	}

	// --- 2. 查询列表 ---
	// 注意：包含 optionX_img 字段
	sqlStr := `
		SELECT id, knowledge_point_id, question_text, 
		       option1, option1_img, option2, option2_img, 
		       option3, option3_img, option4, option4_img, 
		       correct_answer, explanation, note, create_time 
		FROM questions 
		WHERE knowledge_point_id = ? 
		ORDER BY create_time ASC
	`

	rows, err := global.DB.Query(sqlStr, pointID)
	if err != nil {
		log.Println("Query Questions Error:", err)
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	list := make([]model.Question, 0)
	for rows.Next() {
		var q model.Question
		// 必须严格对应 SQL SELECT 的顺序
		err := rows.Scan(
			&q.ID, &q.KnowledgePointID, &q.QuestionText,
			&q.Option1, &q.Option1Img, &q.Option2, &q.Option2Img,
			&q.Option3, &q.Option3Img, &q.Option4, &q.Option4Img,
			&q.CorrectAnswer, &q.Explanation, &q.Note, &q.CreateTime,
		)
		if err != nil {
			log.Println("Scan Question Error:", err)
			continue
		}
		list = append(list, q)
	}

	c.JSON(200, gin.H{"code": 200, "msg": "success", "data": list})
}

// =================================================================================
// CreateQuestion 创建题目
// 权限逻辑：必须是该【科目】的创建者
// =================================================================================
func CreateQuestion(c *gin.Context) {
	var req model.CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	currentUserCode, _ := c.Get("userCode")

	// --- 权限检查 ---
	// 通过 KnowledgePointID 向上查找科目创建者
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
	err := global.DB.QueryRow(checkSQL, req.KnowledgePointID).Scan(&subjectCreatorCode, &creatorName, &creatorEmail)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "msg": "所属知识点不存在"})
		return
	}

	currentUserCodeStr, _ := currentUserCode.(string)
	if subjectCreatorCode != currentUserCodeStr {
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(403, gin.H{"code": 403, "msg": "创建失败：您不是该科目的作者，请联系 " + contactInfo})
		return
	}

	// --- 插入数据 ---
	insertSQL := `
		INSERT INTO questions (
			knowledge_point_id, question_text, 
			option1, option1_img, option2, option2_img, 
			option3, option3_img, option4, option4_img, 
			correct_answer, explanation, note
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := global.DB.Exec(insertSQL,
		req.KnowledgePointID, req.QuestionText,
		req.Option1, req.Option1Img, req.Option2, req.Option2Img,
		req.Option3, req.Option3Img, req.Option4, req.Option4Img,
		req.CorrectAnswer, req.Explanation, req.Note,
	)

	if err != nil {
		log.Println("CreateQuestion DB Error:", err)
		c.JSON(500, gin.H{"code": 500, "msg": "创建失败"})
		return
	}

	id, _ := res.LastInsertId()
	c.JSON(200, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{"id": id}})
}

// =================================================================================
// UpdateQuestion 更新题目
// 权限逻辑：必须是该【科目】的创建者
// =================================================================================
func UpdateQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req model.UpdateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	currentUserCode, _ := c.Get("userCode")

	// --- 权限检查 ---
	// 通过 QuestionID 向上查找科目创建者
	var subjectCreatorCode string
	var creatorName string
	var creatorEmail sql.NullString

	checkSQL := `
		SELECT s.creator_code, IFNULL(u.nickname, u.username), u.email
		FROM questions q
		JOIN knowledge_points p ON q.knowledge_point_id = p.id
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN subjects s ON c.subject_id = s.id
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE q.id = ?
	`
	err := global.DB.QueryRow(checkSQL, id).Scan(&subjectCreatorCode, &creatorName, &creatorEmail)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "msg": "题目不存在"})
		return
	}

	currentUserCodeStr, _ := currentUserCode.(string)
	if subjectCreatorCode != currentUserCodeStr {
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(403, gin.H{"code": 403, "msg": "修改失败：请联系科目作者 " + contactInfo})
		return
	}

	// --- 执行更新 ---
	updateSQL := `
		UPDATE questions SET 
		question_text=?, 
		option1=?, option1_img=?, 
		option2=?, option2_img=?, 
		option3=?, option3_img=?, 
		option4=?, option4_img=?, 
		correct_answer=?, explanation=?, note=?,
		update_time=CURRENT_TIMESTAMP
		WHERE id=?
	`
	_, err = global.DB.Exec(updateSQL,
		req.QuestionText,
		req.Option1, req.Option1Img,
		req.Option2, req.Option2Img,
		req.Option3, req.Option3Img,
		req.Option4, req.Option4Img,
		req.CorrectAnswer, req.Explanation, req.Note,
		id,
	)

	if err != nil {
		log.Println("UpdateQuestion Error:", err)
		c.JSON(500, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	c.JSON(200, gin.H{"code": 200, "msg": "更新成功"})
}

// =================================================================================
// DeleteQuestion 删除题目
// 权限逻辑：必须是该【科目】的创建者
// =================================================================================
func DeleteQuestion(c *gin.Context) {
	id := c.Param("id")
	currentUserCode, _ := c.Get("userCode")

	// --- 权限检查 ---
	var subjectCreatorCode string
	var creatorName string
	var creatorEmail sql.NullString

	checkSQL := `
		SELECT s.creator_code, IFNULL(u.nickname, u.username), u.email
		FROM questions q
		JOIN knowledge_points p ON q.knowledge_point_id = p.id
		JOIN knowledge_categories c ON p.categorie_id = c.id
		JOIN subjects s ON c.subject_id = s.id
		LEFT JOIN users u ON s.creator_code = u.user_code
		WHERE q.id = ?
	`
	err := global.DB.QueryRow(checkSQL, id).Scan(&subjectCreatorCode, &creatorName, &creatorEmail)
	if err != nil {
		c.JSON(404, gin.H{"code": 404, "msg": "题目不存在"})
		return
	}

	currentUserCodeStr, _ := currentUserCode.(string)
	if subjectCreatorCode != currentUserCodeStr {
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(403, gin.H{"code": 403, "msg": "删除失败：请联系科目作者 " + contactInfo})
		return
	}

	// --- 执行删除 ---
	_, err = global.DB.Exec("DELETE FROM questions WHERE id = ?", id)
	if err != nil {
		log.Println("Delete Question Error:", err)
		c.JSON(500, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	c.JSON(200, gin.H{"code": 200, "msg": "删除成功"})
}
