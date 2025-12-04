package model

// Question 对应新的 questions 表
type Question struct {
	ID               int    `json:"id"`
	KnowledgePointID int    `json:"knowledgePointId"` // 对应 knowledge_point_id
	QuestionText     string `json:"questionText"`     // 对应 question_text

	Option1    string `json:"option1"`
	Option1Img string `json:"option1Img"`
	Option2    string `json:"option2"`
	Option2Img string `json:"option2Img"`
	Option3    string `json:"option3"`
	Option3Img string `json:"option3Img"`
	Option4    string `json:"option4"`
	Option4Img string `json:"option4Img"`

	CorrectAnswer int    `json:"correctAnswer"` // 1, 2, 3, 4
	Explanation   string `json:"explanation"`
	Note          string `json:"note"`

	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
}

// CreateQuestionRequest 创建请求
type CreateQuestionRequest struct {
	KnowledgePointID int    `json:"knowledgePointId" binding:"required"`
	QuestionText     string `json:"questionText" binding:"required"`

	Option1    string `json:"option1"`
	Option1Img string `json:"option1Img"`
	Option2    string `json:"option2"`
	Option2Img string `json:"option2Img"`
	Option3    string `json:"option3"`
	Option3Img string `json:"option3Img"`
	Option4    string `json:"option4"`
	Option4Img string `json:"option4Img"`

	CorrectAnswer int    `json:"correctAnswer" binding:"required"` // 必填
	Explanation   string `json:"explanation"`
}

// UpdateQuestionRequest 更新请求
type UpdateQuestionRequest struct {
	QuestionText string `json:"questionText"`

	Option1    string `json:"option1"`
	Option1Img string `json:"option1Img"`
	Option2    string `json:"option2"`
	Option2Img string `json:"option2Img"`
	Option3    string `json:"option3"`
	Option3Img string `json:"option3Img"`
	Option4    string `json:"option4"`
	Option4Img string `json:"option4Img"`

	CorrectAnswer int    `json:"correctAnswer"`
	Explanation   string `json:"explanation"`
}

type UpdateNoteRequest struct {
	QuestionID int    `json:"question_id" binding:"required"`
	Note       string `json:"note"` // 允许为空，为空可能意味着清空备注
}
