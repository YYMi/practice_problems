package model

// KnowledgePoint 知识点详情
type KnowledgePoint struct {
	ID              int    `json:"id"`
	CategoryID      int    `json:"categoryId"`      // 对应 categorie_id
	Title           string `json:"title"`           // 建议新增的标题字段
	Content         string `json:"content"`         // 详细内容
	ReferenceLinks  string `json:"referenceLinks"`  // JSON 字符串
	LocalImageNames string `json:"localImageNames"` // JSON 字符串
	CreateTime      string `json:"createTime"`
	UpdateTime      string `json:"updateTime"`
	// 新增字段
	SortOrder  int `json:"sortOrder"`
	Difficulty int `json:"difficulty"`
}

// CreatePointRequest 创建请求（只传分类ID和标题）
type CreatePointRequest struct {
	CategoryID int    `json:"categoryId" binding:"required"`
	Title      string `json:"title" binding:"required"`
}

// UpdatePointRequest 更新请求（修改内容、图片、链接等）
type UpdatePointRequest struct {
	Title           string `json:"title"`           // ✅ 大写 T
	Content         string `json:"content"`         // ✅ 大写 C
	ReferenceLinks  string `json:"referenceLinks"`  // ✅ 大写 R
	LocalImageNames string `json:"localImageNames"` // ✅ 大写 L
	// 新增：允许修改难度 (使用指针以区分 0)
	Difficulty *int `json:"difficulty"`
}
