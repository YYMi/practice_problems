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
// GetQuestionList 获取题目列表
// =================================================================================
func GetQuestionList(c *gin.Context) {
	pointID := c.Query("point_id")
	categoryID := c.Query("category_id")

	if pointID == "" && categoryID == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "必须指定 point_id 或 category_id"})
		return
	}

	userID, exists := c.Get("userID")
	userCode, _ := c.Get("userCode")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未授权"})
		return
	}

	// =====================================================
	// 第一步：查归属 (找到所属科目，用于鉴权)
	// =====================================================
	var subjectID int
	var creatorCode string
	var err error

	if pointID != "" {
		// A. 按知识点查科目
		findSubjectSQL := `
			SELECT s.id, s.creator_code
			FROM knowledge_points p
			JOIN knowledge_categories c ON p.categorie_id = c.id 
			JOIN subjects s ON c.subject_id = s.id
			WHERE p.id = ?
		`
		err = global.DB.QueryRow(findSubjectSQL, pointID).Scan(&subjectID, &creatorCode)
	} else {
		// B. 按分类查科目
		findSubjectSQL := `
			SELECT s.id, s.creator_code
			FROM knowledge_categories c
			JOIN subjects s ON c.subject_id = s.id
			WHERE c.id = ?
		`
		err = global.DB.QueryRow(findSubjectSQL, categoryID).Scan(&subjectID, &creatorCode)
	}

	if err != nil {
		// 查不到科目，可能是ID不对，不报错，直接返回空
		c.JSON(200, gin.H{"code": 200, "msg": "success", "data": []model.Question{}})
		return
	}

	// =====================================================
	// 第二步：判权限
	// =====================================================
	hasPermission := false

	// 1. 我是作者
	if creatorCode == userCode.(string) {
		hasPermission = true
	} else {
		// 2. 我是订阅者 (检查绑定)
		checkBindSQL := `
			SELECT count(*) 
			FROM user_subjects 
			WHERE user_id = ? 
			  AND subject_id = ? 
			  AND status = 1 
			  AND (expire_time IS NULL OR expire_time > datetime('now', 'localtime'))
		`
		var count int
		err := global.DB.QueryRow(checkBindSQL, userID, subjectID).Scan(&count)
		if err == nil && count > 0 {
			hasPermission = true
		}
	}

	if !hasPermission {
		c.JSON(403, gin.H{"code": 403, "msg": "您无权访问该内容，请先获取授权"})
		return
	}

	// =====================================================
	// 第三步：取数据
	// =====================================================
	var rows *sql.Rows
	var queryErr error

	if pointID != "" {
		sqlStr := `
			SELECT id, knowledge_point_id, question_text, 
			       option1, option1_img, option2, option2_img, 
			       option3, option3_img, option4, option4_img, 
			       correct_answer, explanation, note, create_time 
			FROM questions 
			WHERE knowledge_point_id = ? 
			ORDER BY create_time ASC
		`
		rows, queryErr = global.DB.Query(sqlStr, pointID)
	} else {
		sqlStr := `
			SELECT q.id, q.knowledge_point_id, q.question_text, 
			       q.option1, q.option1_img, q.option2, q.option2_img, 
			       q.option3, q.option3_img, q.option4, q.option4_img, 
			       q.correct_answer, q.explanation, q.note, q.create_time 
			FROM questions q
			JOIN knowledge_points p ON q.knowledge_point_id = p.id
			WHERE p.categorie_id = ? 
			ORDER BY q.create_time ASC
		`
		rows, queryErr = global.DB.Query(sqlStr, categoryID)
	}

	if queryErr != nil {
		// ★★★ Error ★★★
		global.GetLog(c).Errorf("查询题目列表失败 (PointID: %s, CatID: %s): %v", pointID, categoryID, queryErr)
		c.JSON(200, gin.H{"code": 200, "msg": "success", "data": []model.Question{}})
		return
	}
	defer rows.Close()

	list := make([]model.Question, 0)

	for rows.Next() {
		var q model.Question
		err := rows.Scan(
			&q.ID, &q.KnowledgePointID, &q.QuestionText,
			&q.Option1, &q.Option1Img, &q.Option2, &q.Option2Img,
			&q.Option3, &q.Option3Img, &q.Option4, &q.Option4Img,
			&q.CorrectAnswer, &q.Explanation, &q.Note, &q.CreateTime,
		)
		if err != nil {
			continue
		}
		list = append(list, q)
	}

	c.JSON(200, gin.H{"code": 200, "msg": "success", "data": list})
}

// =================================================================================
// CreateQuestion 创建题目
// =================================================================================
func CreateQuestion(c *gin.Context) {
	var req model.CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
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

	if subjectCreatorCode != currentUserCodeStr {
		// ★★★ Warn ★★★
		global.GetLog(c).Warnf("创建题目被拒: 无权操作 (User: %s, PointID: %d)", currentUserCodeStr, req.KnowledgePointID)
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
		// ★★★ Error ★★★
		global.GetLog(c).Errorf("创建题目DB错误: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "创建失败"})
		return
	}

	id, _ := res.LastInsertId()
	// ★★★ Info ★★★
	global.GetLog(c).Infof("用户[%s] 创建题目成功: ID=%d", currentUserCodeStr, id)
	c.JSON(200, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{"id": id}})
}

// =================================================================================
// UpdateQuestion 更新题目
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
	currentUserCodeStr, _ := currentUserCode.(string)

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

	if subjectCreatorCode != currentUserCodeStr {
		global.GetLog(c).Warnf("修改题目被拒: 无权操作 (User: %s, QuestionID: %d)", currentUserCodeStr, id)
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
		global.GetLog(c).Errorf("更新题目DB错误 (ID: %d): %v", id, err)
		c.JSON(500, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	global.GetLog(c).Infof("用户[%s] 更新题目成功 (ID: %d)", currentUserCodeStr, id)
	c.JSON(200, gin.H{"code": 200, "msg": "更新成功"})
}

// =================================================================================
// DeleteQuestion 删除题目
// =================================================================================
func DeleteQuestion(c *gin.Context) {
	id := c.Param("id")
	currentUserCode, _ := c.Get("userCode")
	currentUserCodeStr, _ := currentUserCode.(string)

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

	if subjectCreatorCode != currentUserCodeStr {
		global.GetLog(c).Warnf("删除题目被拒: 无权操作 (User: %s, QuestionID: %s)", currentUserCodeStr, id)
		contactInfo := getContactInfo(creatorName, creatorEmail)
		c.JSON(403, gin.H{"code": 403, "msg": "删除失败：请联系科目作者 " + contactInfo})
		return
	}

	// --- 执行删除 ---
	_, err = global.DB.Exec("DELETE FROM questions WHERE id = ?", id)
	if err != nil {
		log.Println("Delete Question Error:", err)
		global.GetLog(c).Errorf("删除题目DB错误 (ID: %s): %v", id, err)
		c.JSON(500, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	global.GetLog(c).Infof("用户[%s] 删除题目成功 (ID: %s)", currentUserCodeStr, id)
	c.JSON(200, gin.H{"code": 200, "msg": "删除成功"})
}
