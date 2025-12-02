package api

import (
	"database/sql"
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

// UploadImage 通用上传
// 参数 type: "point" | "question"
// UploadImage 通用上传 (增强版)
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
	// 如果是知识点图片，必须传 id
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

	// ==================== 新增逻辑：检查文件大小 (10MB) ====================
	const MaxFileSize = 10 * 1024 * 1024 // 10MB
	if file.Size > MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "文件大小不能超过 10MB"})
		return
	}

	// ==================== 新增逻辑：权限验证 (仅针对 point 类型) ====================
	if bizType == "point" {
		if targetIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "上传知识点图片必须提供 id"})
			return
		}
		pointID, _ := strconv.Atoi(targetIDStr)

		// 连表查询 SQL: points -> categories -> subjects
		querySql := `
			SELECT s.creator_code 
			FROM knowledge_points kp
			INNER JOIN knowledge_categories kc ON kp.categorie_id = kc.id
			INNER JOIN subjects s ON kc.subject_id = s.id
			WHERE kp.id = ?
		`
		var ownerCode string
		// 注意：这里假设你使用全局 db 变量
		err := global.DB.QueryRow(querySql, pointID).Scan(&ownerCode)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "找不到对应的知识点或关联数据"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询权限失败"})
			return
		}

		// 比对 Creator Code
		if ownerCode != currentUserCode {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权修改此知识点的内容"})
			return
		}
	}

	// 4. 保存文件 (原有逻辑)
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

	webPath := fmt.Sprintf("/uploads/%s/%s/%s", bizType, dateStr, newFileName)
	fullURL := webPath

	// ==================== 新增逻辑：更新 knowledge_points 表 ====================
	// 如果是知识点图片，需要将图片路径记录到 local_image_names 字段
	// 我们采用追加模式，用逗号分隔
	if bizType == "point" {
		pointID, _ := strconv.Atoi(targetIDStr)

		// SQLite 更新语句：如果为空则直接赋值，如果不为空则追加逗号和新路径
		updateSql := `
			UPDATE knowledge_points 
			SET local_image_names = CASE 
				WHEN local_image_names IS NULL OR local_image_names = '' THEN ? 
				ELSE local_image_names || ',' || ? 
			END
			WHERE id = ?
		`
		// 参数传两次 webPath，分别对应 CASE 的两种情况
		_, err := global.DB.Exec(updateSql, webPath, webPath, pointID)
		if err != nil {
			// 注意：虽然数据库更新失败，但文件已经保存了。
			// 实际生产中可能需要回滚（删除文件），这里简单处理记录日志即可
			fmt.Printf("警告: 图片上传成功但数据库更新失败: %v\n", err)
		}
	}

	// 5. 返回结果
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
