package model

// PointBinding 知识点绑定关系
type PointBinding struct {
	ID              int    `json:"id"`
	SourceSubjectID int    `json:"sourceSubjectId"` // 源科目ID
	SourcePointID   int    `json:"sourcePointId"`   // 源知识点ID
	TargetSubjectID int    `json:"targetSubjectId"` // 目标科目ID
	TargetPointID   int    `json:"targetPointId"`   // 目标知识点ID
	BindText        string `json:"bindText"`        // 绑定的文字（选中的文字）
	UserID          int    `json:"userId"`          // 创建者ID
	CreateTime      string `json:"createTime"`
}

// CreateBindingRequest 创建绑定请求
type CreateBindingRequest struct {
	SourceSubjectID int    `json:"sourceSubjectId" binding:"required"`
	SourcePointID   int    `json:"sourcePointId" binding:"required"`
	TargetSubjectID int    `json:"targetSubjectId" binding:"required"`
	TargetPointID   int    `json:"targetPointId" binding:"required"`
	BindText        string `json:"bindText" binding:"required"`
}

// BindingWithDetails 带详情的绑定信息（用于查询返回）
type BindingWithDetails struct {
	PointBinding
	SourceSubjectName string `json:"sourceSubjectName"`
	SourcePointTitle  string `json:"sourcePointTitle"`
	TargetSubjectName string `json:"targetSubjectName"`
	TargetPointTitle  string `json:"targetPointTitle"`
}
