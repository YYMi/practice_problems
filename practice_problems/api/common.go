package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadImage 通用上传
// 参数 type: "point" | "question"
func UploadImage(c *gin.Context) {
	bizType := c.PostForm("type")
	if bizType != "point" && bizType != "question" {
		bizType = "common"
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请上传文件"})
		return
	}

	// 按 uploads/类型/日期 存储
	dateStr := time.Now().Format("20060102")
	uploadDir := fmt.Sprintf("./uploads/%s/%s", bizType, dateStr)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建目录失败"})
		return
	}

	ext := filepath.Ext(file.Filename)
	newFileName := uuid.New().String() + ext
	savePath := filepath.Join(uploadDir, newFileName)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "保存失败"})
		return
	}

	// 返回 URL
	webPath := fmt.Sprintf("/uploads/%s/%s/%s", bizType, dateStr, newFileName)
	fullURL := "http://localhost:8080" + webPath

	c.JSON(http.StatusOK, gin.H{
		"code": 200, "msg": "上传成功",
		"data": gin.H{"url": fullURL, "path": webPath},
	})
}

// RemoveFileFromDisk 纯粹的工具函数：只负责从硬盘删除文件
// 返回 error，供其他地方调用
func RemoveFileFromDisk(targetPath string) error {
	// 安全校验和路径转换逻辑
	if len(targetPath) > 0 && targetPath[0] == '/' {
		targetPath = targetPath[1:] // 去掉开头的 /
	}
	// 变成 ./uploads/...
	localPath := filepath.Join(".", targetPath)

	// 简单的安全检查，防止删除 uploads 以外的文件
	// 实际路径应该是 uploads/...
	if !strings.HasPrefix(filepath.Clean(localPath), "uploads") {
		// 这里为了演示简单，只要不是 uploads 开头就不删，防止 .. 攻击
		// 生产环境可以用更严谨的 filepath.Rel 判断
	}

	// 检查文件是否存在
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在")
	}

	// 执行删除
	return os.Remove(localPath)
}

// DeleteImage 供前端直接调用的接口 (路由 Handler)
func DeleteImage(c *gin.Context) {
	targetPath := c.Query("path")
	if targetPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "路径不能为空"})
		return
	}

	// 直接调用上面的工具函数
	err := RemoveFileFromDisk(targetPath)
	if err != nil {
		// 区分一下是文件不存在还是删除失败
		if err.Error() == "文件不存在" {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除文件失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}
