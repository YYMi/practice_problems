package model

type DeletePointImageRequest struct {
	// 前端传过来的图片路径，必须和数据库里存的一模一样
	// 例如: "/uploads/point/20231128/abc.jpg"
	FilePath string `json:"filePath" binding:"required"`
}

// 定义结构体来匹配数据库里的 JSON 对象
type ImageItem struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}
