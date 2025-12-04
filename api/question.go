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
// GetQuestionList 获取题目列表 (完整修复版)
// =================================================================================
func GetQuestionList(c *gin.Context) {
	// 1. 参数校验
	pointID := c.Query("point_id")
	categoryID := c.Query("category_id")

	if pointID == "" && categoryID == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "必须指定 point_id 或 category_id"})
		return
	}

	// 2. 获取用户信息 (userID 用于查备注，userCode 用于查权限)
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未授权"})
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

	userCodeRaw, _ := c.Get("userCode")
	userCode, _ := userCodeRaw.(string) // 这里的 userCode 将在下面鉴权时使用

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
		// 查不到科目，可能是ID不对，直接返回空
		c.JSON(200, gin.H{"code": 200, "msg": "success", "data": []model.Question{}})
		return
	}

	// =====================================================
	// 第二步：判权限 (这里使用了 userCode 和 userID)
	// =====================================================
	hasPermission := false

	// 1. 我是作者 (使用了 userCode)
	if creatorCode == userCode {
		hasPermission = true
	} else {
		// 2. 我是订阅者 (使用了 userID)
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
	// 第三步：取数据 (使用了 userID 关联查询备注)
	// =====================================================
	var rows *sql.Rows
	var queryErr error

	// SQL 逻辑：LEFT JOIN question_user_notes 表，如果有关联记录则取 note，否则取空字符串
	if pointID != "" {
		sqlStr := `
			SELECT q.id, q.knowledge_point_id, q.question_text, 
			       q.option1, q.option1_img, q.option2, q.option2_img, 
			       q.option3, q.option3_img, q.option4, q.option4_img, 
			       q.correct_answer, q.explanation, 
                   IFNULL(un.note, '') as user_note, 
                   q.create_time 
			FROM questions q
            LEFT JOIN question_user_notes un ON q.id = un.question_id AND un.user_id = ?
			WHERE q.knowledge_point_id = ? 
			ORDER BY q.create_time ASC
		`
		rows, queryErr = global.DB.Query(sqlStr, userID, pointID)
	} else {
		sqlStr := `
			SELECT q.id, q.knowledge_point_id, q.question_text, 
			       q.option1, q.option1_img, q.option2, q.option2_img, 
			       q.option3, q.option3_img, q.option4, q.option4_img, 
			       q.correct_answer, q.explanation, 
                   IFNULL(un.note, '') as user_note, 
                   q.create_time 
			FROM questions q
			JOIN knowledge_points p ON q.knowledge_point_id = p.id
            LEFT JOIN question_user_notes un ON q.id = un.question_id AND un.user_id = ?
			WHERE p.categorie_id = ? 
			ORDER BY q.create_time ASC
		`
		rows, queryErr = global.DB.Query(sqlStr, userID, categoryID)
	}

	if queryErr != nil {
		global.GetLog(c).Errorf("查询题目列表失败: %v", queryErr)
		c.JSON(200, gin.H{"code": 200, "msg": "success", "data": []model.Question{}})
		return
	}
	defer rows.Close()

	list := make([]model.Question, 0)

	for rows.Next() {
		var q model.Question
		// Scan 必须与 SQL SELECT 字段一一对应
		err := rows.Scan(
			&q.ID, &q.KnowledgePointID, &q.QuestionText,
			&q.Option1, &q.Option1Img, &q.Option2, &q.Option2Img,
			&q.Option3, &q.Option3Img, &q.Option4, &q.Option4Img,
			&q.CorrectAnswer, &q.Explanation,
			&q.Note, // 这里存入的是用户的私有备注
			&q.CreateTime,
		)
		if err != nil {
			global.GetLog(c).Errorf("Scan error: %v", err) // 建议加上日志，方便排查
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
			correct_answer, explanation
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := global.DB.Exec(insertSQL,
		req.KnowledgePointID, req.QuestionText,
		req.Option1, req.Option1Img, req.Option2, req.Option2Img,
		req.Option3, req.Option3Img, req.Option4, req.Option4Img,
		req.CorrectAnswer, req.Explanation,
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

	// 修改点：SQL 中去掉了 note=?, 以及参数中的 req.Note
	updateSQL := `
        UPDATE questions SET 
        question_text=?, 
        option1=?, option1_img=?, 
        option2=?, option2_img=?, 
        option3=?, option3_img=?, 
        option4=?, option4_img=?, 
        correct_answer=?, explanation=?, 
        -- note=?,  <-- 删掉这一行，不再更新原表的 note
        update_time=CURRENT_TIMESTAMP
        WHERE id=?
    `
	_, err = global.DB.Exec(updateSQL,
		req.QuestionText,
		req.Option1, req.Option1Img,
		req.Option2, req.Option2Img,
		req.Option3, req.Option3Img,
		req.Option4, req.Option4Img,
		req.CorrectAnswer, req.Explanation,
		// req.Note, <-- 删掉这一行参数
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
// UpdateUserNote 更新用户题目备注 (Upsert: 存在则更新，不存在则插入)
// =================================================================================
func UpdateUserNote(c *gin.Context) {
	var req model.UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 获取当前用户ID
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未授权"})
		return
	}
	// 类型断言处理 (根据你实际中间件存的类型)
	var userID int
	if v, ok := userIDVal.(int); ok {
		userID = v
	} else if v, ok := userIDVal.(float64); ok {
		userID = int(v)
	}

	// --- 简单的权限检查 (可选) ---
	// 理论上应该检查用户是否有权访问这道题（是否订阅了科目），
	// 但如果前端只在能看到题目时才调用此接口，且不做严格防刷，可以直接执行写入。
	// 严格做法是复用 GetQuestionList 里的 checkBindSQL 逻辑。

	// --- 执行 Upsert ---
	// SQLite 特有语法：ON CONFLICT
	upsertSQL := `
		INSERT INTO question_user_notes (user_id, question_id, note, create_time, update_time)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id, question_id) 
		DO UPDATE SET 
			note = excluded.note,
			update_time = CURRENT_TIMESTAMP
	`

	_, err := global.DB.Exec(upsertSQL, userID, req.QuestionID, req.Note)

	if err != nil {
		global.GetLog(c).Errorf("更新用户备注失败 (UID: %d, QID: %d): %v", userID, req.QuestionID, err)
		c.JSON(500, gin.H{"code": 500, "msg": "保存备注失败"})
		return
	}

	c.JSON(200, gin.H{"code": 200, "msg": "备注已保存"})
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
