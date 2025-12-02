package model

// CreateShareRequest 创建分享的请求参数
// CreateShareRequest 创建分享
type CreateShareRequest struct {
	SubjectIDs   []int    `json:"subject_ids" binding:"required"`
	Duration     string   `json:"duration" binding:"required"` // 资源有效期 (给用户看的)
	CodeDuration string   `json:"code_duration"`               // ★★★ 新增：分享码有效期 (给码用的)
	Type         int      `json:"type" binding:"required"`
	Targets      []string `json:"targets"`
}

// BindShareRequest 绑定分享的请求参数
type BindShareRequest struct {
	Code string `json:"code" binding:"required"`
}

type ShareCode struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	SubjectID   int    `json:"subject_id"`
	CreatorID   int    `json:"creator_id"`
	DurationStr string `json:"duration_str"` // 存 "7d", "30d" 等
	CreateTime  string `json:"create_time"`
}

type ShareAnnouncement struct {
	ID          int    `json:"id"`
	CreatorCode string `json:"creatorCode"`
	ShareCode   string `json:"shareCode"`
	Note        string `json:"note"`
	CreateTime  string `json:"createTime"`
	ExpireTime  string `json:"expireTime"`
	Status      int    `json:"status"`
}
