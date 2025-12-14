package global

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	DB  *sql.DB
	Log *zap.SugaredLogger

	// OSS 配置
	OssEndpoint         string // OSS Bucket 访问地址 (外网)
	OssPrefix           string // 资源路径前缀
	OssAccessKeyID      string // AccessKey ID
	OssAccessKeySecret  string // AccessKey Secret
	OssBucket           string // Bucket 名称
	OssInternalEndpoint string // 内网 Endpoint (用于上传)
	DeepseekApiKey      string // Ai DeepseekApiKey API Key
	VoiceAppKey         string // 阿里云的语音模型 API Key
)

// IsOssUploadEnabled 判断是否启用 OSS 上传
// 需要配置 AccessKeyID 和 AccessKeySecret 才能上传
func IsOssUploadEnabled() bool {
	return OssAccessKeyID != "" && OssAccessKeySecret != "" && OssBucket != ""
}

// GetOssUploadEndpoint 获取用于上传的 Endpoint
// 优先使用内网 Endpoint，如果没有则使用外网 Endpoint
func GetOssUploadEndpoint() string {
	if OssInternalEndpoint != "" {
		return OssInternalEndpoint
	}
	// 从外网 endpoint 提取域名部分 (去掉 https://bucket-name. 前缀)
	endpoint := OssEndpoint
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	// 去掉 bucket 名称前缀
	if strings.HasPrefix(endpoint, OssBucket+".") {
		endpoint = strings.TrimPrefix(endpoint, OssBucket+".")
	}
	return endpoint
}

// GetOssUrl 获取完整的 OSS 基础地址
// 如果未配置 OSS，返回空字符串
func GetOssUrl() string {
	if OssEndpoint == "" {
		return ""
	}
	// 拼接 endpoint 和 prefix
	endpoint := strings.TrimRight(OssEndpoint, "/")
	prefix := strings.TrimLeft(OssPrefix, "/")
	if prefix != "" {
		return endpoint + "/" + prefix
	}
	return endpoint
}

// GetLog 获取带 RequestID 的 Logger
func GetLog(c *gin.Context) *zap.SugaredLogger {
	if c == nil {
		return Log
	}

	reqID, exists := c.Get("RequestID")
	if !exists {
		return Log
	}

	// ★★★ 核心修正：使用 Named ★★★
	// 不要自己加方括号 "[]"，我们在 logger.go 配置里统一加
	// 也不要用 With，用 Named 才能让它跑到日志中间去
	return Log.Named(fmt.Sprintf("%s", reqID))
}
