package api

import (
	"database/sql"
	"encoding/json" // ★★★ 新增：用于处理 JSON
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"practice_problems/global"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 定义结构体 (为了和 DeletePointImage 保持一致)
type ImageItem struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

// UploadImage 通用上传 (包含 10 张上限限制 + JSON 格式修复 + 日志 + 权限修复)
func UploadImage(c *gin.Context) {
	// 1. 获取用户信息
	_, exists := c.Get("userID")
	userCodeInterface, _ := c.Get("userCode")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}
	currentUserCode := userCodeInterface.(string)

	// 2. 获取基本参数
	bizType := c.PostForm("type")
	targetIDStr := c.PostForm("pointId")

	if bizType != "point" && bizType != "question" {
		bizType = "common"
	}

	// 3. 接收文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请上传文件"})
		return
	}

	// 检查文件大小 (10MB)
	const MaxFileSize = 10 * 1024 * 1024
	if file.Size > MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "文件大小不能超过 10MB"})
		return
	}

	// ==================== 核心修改：权限验证 + JSON数量检查 ====================
	var currentImages []ImageItem // 用于暂存当前的图片列表

	if bizType == "point" {
		if targetIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "上传知识点图片必须提供 id"})
			return
		}
		pointID, _ := strconv.Atoi(targetIDStr)

		// 查询 creator_code 和 local_image_names
		querySql := `
			SELECT s.creator_code, COALESCE(kp.local_image_names, '[]')
			FROM knowledge_points kp
			INNER JOIN knowledge_categories kc ON kp.categorie_id = kc.id
			INNER JOIN subjects s ON kc.subject_id = s.id
			WHERE kp.id = ?
		`

		var ownerCode string
		var localImageNamesStr string

		err := global.DB.QueryRow(querySql, pointID).Scan(&ownerCode, &localImageNamesStr)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "找不到对应的知识点或关联数据"})
				return
			}
			global.GetLog(c).Errorf("上传图片查询失败 (PointID: %d): %v", pointID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询数据失败"})
			return
		}

		// 1. 验证权限
		if ownerCode != currentUserCode {
			global.GetLog(c).Warnf("上传图片被拒: 越权操作 (User: %s, PointID: %d)", currentUserCode, pointID)
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权修改此知识点的内容"})
			return
		}

		// 2. 验证图片数量 (解析 JSON)
		if err := json.Unmarshal([]byte(localImageNamesStr), &currentImages); err != nil {
			currentImages = make([]ImageItem, 0)
		}

		if len(currentImages) >= 10 {
			global.GetLog(c).Warnf("上传图片被拒: 数量超限 (User: %s, PointID: %d, Count: %d)", currentUserCode, pointID, len(currentImages))
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "图片数量已达上限 (最多 10 张)"})
			return
		}
	}
	// ==================== 结束修改 ====================

	// 4. 保存文件 (★ 核心修复点 ★)
	dateStr := time.Now().Format("20060102")
	uploadDir := fmt.Sprintf("./uploads/%s/%s", bizType, dateStr)

	// A. 创建目录，显式指定 0755
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		global.GetLog(c).Errorf("创建上传目录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建目录失败"})
		return
	}

	// B. 双重保险：强制修改权限为 755 (防止 umask 干扰)
	// 这样确保 Nginx (www-data) 绝对能进入这个目录读取图片
	if err := os.Chmod(uploadDir, 0755); err != nil {
		// 这个错误不致命，只记录警告
		global.GetLog(c).Warnf("强制修改目录权限失败 (Dir: %s): %v", uploadDir, err)
	}

	ext := filepath.Ext(file.Filename)
	newFileName := uuid.New().String() + ext
	savePath := filepath.Join(uploadDir, newFileName)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		global.GetLog(c).Errorf("保存文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "保存失败"})
		return
	}

	// C. 文件本身的权限也建议保险一下 (设为 644: rw-r--r--)
	// 这样任何人都能读，但只有拥有者能改
	os.Chmod(savePath, 0644)

	webPath := fmt.Sprintf("/uploads/%s/%s/%s", bizType, dateStr, newFileName)
	fullURL := webPath

	// 5. 更新数据库 (使用 JSON 格式更新)
	if bizType == "point" {
		pointID, _ := strconv.Atoi(targetIDStr)

		newImg := ImageItem{
			Name: "新图片",
			Url:  webPath,
		}
		currentImages = append(currentImages, newImg)

		newJsonBytes, _ := json.Marshal(currentImages)

		updateSql := `UPDATE knowledge_points SET local_image_names = ? WHERE id = ?`
		_, err := global.DB.Exec(updateSql, string(newJsonBytes), pointID)
		if err != nil {
			global.GetLog(c).Errorf("图片上传成功但DB更新失败 (PointID: %d): %v", pointID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "数据库更新失败"})
			return
		}
	}

	global.GetLog(c).Infof("New版本 用户[%s] 上传图片成功: %s", currentUserCode, webPath)

	c.JSON(http.StatusOK, gin.H{
		"code": 200, "msg": "上传成功",
		"data": gin.H{"url": fullURL, "path": webPath},
	})
}

// RemoveFileFromDisk 纯粹的工具函数：只负责从硬盘删除文件
func RemoveFileFromDisk(targetPath string) error {
	// 安全校验和路径转换逻辑
	if len(targetPath) > 0 && targetPath[0] == '/' {
		targetPath = targetPath[1:] // 去掉开头的 /
	}
	// 变成 ./uploads/...
	localPath := filepath.Join(".", targetPath)

	// 简单的安全检查
	if !strings.HasPrefix(filepath.Clean(localPath), "uploads") {
		return fmt.Errorf("非法路径")
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
		if err.Error() == "文件不存在" {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": err.Error()})
		} else {
			global.GetLog(c).Errorf("删除独立文件失败: %s, error: %v", targetPath, err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除文件失败"})
		}
		return
	}

	global.GetLog(c).Infof("独立文件删除成功: %s", targetPath)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}
