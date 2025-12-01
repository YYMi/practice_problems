package model

// KnowledgeCategory 对应数据库 knowledge_categories 表
type KnowledgeCategory struct {
	ID           int    `json:"id"`
	SubjectID    int    `json:"subjectId"`    // 关联的科目ID
	CategoryName string `json:"categoryName"` // 对应数据库 categorie_name
	CreateTime   string `json:"createTime"`
	UpdateTime   string `json:"updateTime"`
	SortOrder    int    `json:"sortOrder"`  // 新增：排序字段
	Difficulty   int    `json:"difficulty"` // 新增：难度 (0-简单, 1-普通, 2-困难, 3-地狱)
	CreatorCode  string `json:"creatorCode"`
}

// CreateCategoryRequest 创建分类时的参数
type CreateCategoryRequest struct {
	SubjectID    int    `json:"subjectId" binding:"required"`    // 必须指定属于哪个科目
	CategoryName string `json:"categoryName" binding:"required"` // 分类名称必填
}

// UpdateCategoryRequest 更新分类时的参数
type UpdateCategoryRequest struct {
	CategoryName string `json:"categoryName" binding:"required"`
	Difficulty   *int   `json:"difficulty"`
}
