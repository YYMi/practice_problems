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
	SortOrder  int    `json:"sortOrder"`
	Difficulty int    `json:"difficulty"`
	VideoUrl   string `json:"videoUrl"`
}

// CreatePointRequest 创建请求（传分类ID、标题和难度）
type CreatePointRequest struct {
	CategoryID int    `json:"categoryId" binding:"required"`
	Title      string `json:"title" binding:"required"`
	Difficulty int    `json:"difficulty"` // 难度：0-简单，1-中等，2-困难，3-重点
}

// UpdatePointRequest 更新请求（修改内容、图片、链接等）
type UpdatePointRequest struct {
	Title           string  `json:"title"`           // ✅ 大写 T
	Content         string  `json:"content"`         // ✅ 大写 C
	ReferenceLinks  string  `json:"referenceLinks"`  // ✅ 大写 R
	LocalImageNames string  `json:"localImageNames"` // ✅ 大写 L
	VideoUrl        *string `json:"videoUrl"`
	// 新增：允许修改难度 (使用指针以区分 0)
	Difficulty *int `json:"difficulty"`
	// ★★★ 新增 ★★★
	CategoryID *int `json:"categoryId"`
}
