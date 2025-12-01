package model

// Subject 对应数据库中的 subjects 表
type Subject struct {
	ID          int    `json:"id"`          // 对应 id
	Name        string `json:"name"`        // 对应 name
	Status      int    `json:"status"`      // 对应 status
	CreateTime  string `json:"createTime"`  // 对应 create_time
	UpdateTime  string `json:"updateTime"`  // 对应 update_time
	CreatorCode string `json:"creatorCode"` // 对应创建者的code
}

// CreateSubjectRequest 用于接收创建请求的参数
type CreateSubjectRequest struct {
	Name   string `json:"name" binding:"required"` // 必填
	Status int    `json:"status"`                  // 选填，默认为0
}

// UpdateSubjectRequest 用于接收更新请求的参数
type UpdateSubjectRequest struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
}
