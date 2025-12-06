package api

import (
	"database/sql"
	"net/http"
	"practice_problems/global" // 替换为你实际的项目路径
	"practice_problems/model"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

// =================================================================================
// CreateShareAnnouncement 发布公告
// =================================================================================
func CreateShareAnnouncement(c *gin.Context) {
	userCodeInterface, exists := c.Get("userCode")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}
	userCode := userCodeInterface.(string)

	var req struct {
		ShareCode  string `json:"shareCode" binding:"required"`
		Note       string `json:"note"`
		ExpireTime string `json:"expireTime"` // 格式: "2006-01-02 15:04:05"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	if utf8.RuneCountInString(req.Note) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "备注最多200字"})
		return
	}

	if req.ExpireTime != "" {
		_, err := time.ParseInLocation("2006-01-02 15:04:05", req.ExpireTime, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "过期时间格式错误，应为 YYYY-MM-DD HH:mm:ss"})
			return
		}
	}

	// ★★★ 检查分享码是否已过期 ★★★
	var shareCodeExpireTime sql.NullString
	err := global.DB.QueryRow(
		"SELECT expire_time FROM share_codes WHERE code = ? AND status = 1",
		req.ShareCode,
	).Scan(&shareCodeExpireTime)

	if err == sql.ErrNoRows {
		global.GetLog(c).Warnf("发布公告被拒: 分享码不存在 (User: %s, ShareCode: %s)", userCode, req.ShareCode)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "分享码不存在或已失效"})
		return
	}

	if err != nil {
		global.GetLog(c).Errorf("查询分享码失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	// 判断分享码是否已过期
	if shareCodeExpireTime.Valid && shareCodeExpireTime.String != "" {
		expireTime, parseErr := time.ParseInLocation("2006-01-02 15:04:05", shareCodeExpireTime.String, time.Local)
		if parseErr == nil && time.Now().After(expireTime) {
			global.GetLog(c).Warnf("发布公告被拒: 分享码已过期 (User: %s, ShareCode: %s, ExpireTime: %s)", userCode, req.ShareCode, shareCodeExpireTime.String)
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "发布失败：分享码已过期，不允许发布公告"})
			return
		}
	}

	createTimeStr := time.Now().Format("2006-01-02 15:04:05")

	res, err := global.DB.Exec(
		"INSERT INTO share_announcements (creator_code, share_code, note, expire_time, create_time, status) VALUES (?, ?, ?, ?, ?, 1)",
		userCode, req.ShareCode, req.Note, req.ExpireTime, createTimeStr,
	)

	if err != nil {
		global.GetLog(c).Errorf("发布公告DB错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "发布失败: " + err.Error()})
		return
	}

	id, _ := res.LastInsertId()

	data := model.ShareAnnouncement{
		ID:          int(id),
		CreatorCode: userCode,
		ShareCode:   req.ShareCode,
		Note:        req.Note,
		CreateTime:  createTimeStr,
		ExpireTime:  req.ExpireTime,
		Status:      1,
	}

	global.GetLog(c).Infof("用户[%s] 发布公告成功: ID=%d", userCode, id)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "发布成功", "data": data})
}

// =================================================================================
// GetShareAnnouncementList 获取公告列表 (按创建时间倒序，且过滤已过期的)
// =================================================================================
func GetShareAnnouncementList(c *gin.Context) {
	nowStr := time.Now().Format("2006-01-02 15:04:05")

	rows, err := global.DB.Query(`
		SELECT id, creator_code, share_code, note, create_time, expire_time, status 
		FROM share_announcements 
		WHERE status = 1 AND expire_time > ?
		ORDER BY create_time DESC
	`, nowStr)

	if err != nil {
		global.GetLog(c).Errorf("查询公告列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	var list []model.ShareAnnouncement
	for rows.Next() {
		var item model.ShareAnnouncement
		var note sql.NullString
		var expireTime sql.NullString

		err := rows.Scan(
			&item.ID,
			&item.CreatorCode,
			&item.ShareCode,
			&note,
			&item.CreateTime,
			&expireTime,
			&item.Status,
		)
		if err != nil {
			continue
		}

		item.Note = note.String
		item.ExpireTime = expireTime.String

		item.CreateTime = formatTimeStr(item.CreateTime)
		item.ExpireTime = formatTimeStr(item.ExpireTime)

		list = append(list, item)
	}

	if list == nil {
		list = []model.ShareAnnouncement{}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": list})
}

// =================================================================================
// DeleteShareAnnouncement 删除公告 (软删除，仅限创建者)
// =================================================================================
func DeleteShareAnnouncement(c *gin.Context) {
	id := c.Param("id")
	userCodeInterface, _ := c.Get("userCode")
	currentUserCode := userCodeInterface.(string)

	// ★★★ 检查公告对应的分享码是否已过期 ★★★
	var shareCode string
	err := global.DB.QueryRow(
		"SELECT share_code FROM share_announcements WHERE id = ? AND creator_code = ? AND status = 1",
		id, currentUserCode,
	).Scan(&shareCode)

	if err == sql.ErrNoRows {
		global.GetLog(c).Warnf("删除公告被拒: 无权操作或不存在 (User: %s, ID: %s)", currentUserCode, id)
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "删除失败：公告不存在或您无权删除"})
		return
	}

	if err != nil {
		global.GetLog(c).Errorf("查询公告失败 (ID: %s): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	// 检查分享码是否过期
	var shareCodeExpireTime sql.NullString
	err = global.DB.QueryRow(
		"SELECT expire_time FROM share_codes WHERE code = ? AND status = 1",
		shareCode,
	).Scan(&shareCodeExpireTime)

	if err != nil {
		global.GetLog(c).Errorf("查询分享码失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	if shareCodeExpireTime.Valid && shareCodeExpireTime.String != "" {
		expireTime, parseErr := time.ParseInLocation("2006-01-02 15:04:05", shareCodeExpireTime.String, time.Local)
		if parseErr == nil && time.Now().After(expireTime) {
			global.GetLog(c).Warnf("删除公告被拒: 分享码已过期 (User: %s, ID: %s, ShareCode: %s)", currentUserCode, id, shareCode)
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "删除失败：分享码已过期，不允许删除公告"})
			return
		}
	}

	res, err := global.DB.Exec(
		"UPDATE share_announcements SET status = 0 WHERE id = ? AND creator_code = ?",
		id, currentUserCode,
	)

	if err != nil {
		global.GetLog(c).Errorf("删除公告DB错误 (ID: %s): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "操作失败"})
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		global.GetLog(c).Warnf("删除公告被拒: 无权操作或不存在 (User: %s, ID: %s)", currentUserCode, id)
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "删除失败：公告不存在或您无权删除"})
		return
	}

	global.GetLog(c).Infof("用户[%s] 删除公告成功 (ID: %s)", currentUserCode, id)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// =================================================================================
// UpdateShareAnnouncement 修改公告 (仅限创建者)
// =================================================================================
func UpdateShareAnnouncement(c *gin.Context) {
	id := c.Param("id")
	userCodeInterface, _ := c.Get("userCode")
	currentUserCode := userCodeInterface.(string)

	var req struct {
		Note       string `json:"note"`
		ExpireTime string `json:"expireTime"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	if utf8.RuneCountInString(req.Note) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "备注最多200字"})
		return
	}

	if req.ExpireTime != "" {
		_, err := time.ParseInLocation("2006-01-02 15:04:05", req.ExpireTime, time.Local)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "过期时间格式错误"})
			return
		}
	}

	// ★★★ 检查公告是否已过期 ★★★
	var currentExpireTime sql.NullString
	var shareCode string
	err := global.DB.QueryRow(
		"SELECT expire_time, share_code FROM share_announcements WHERE id = ? AND creator_code = ? AND status = 1",
		id, currentUserCode,
	).Scan(&currentExpireTime, &shareCode)

	if err == sql.ErrNoRows {
		global.GetLog(c).Warnf("更新公告被拒: 公告不存在或无权操作 (User: %s, ID: %s)", currentUserCode, id)
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "更新失败：公告不存在或您无权修改"})
		return
	}

	if err != nil {
		global.GetLog(c).Errorf("查询公告过期时间失败 (ID: %s): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	// 判断是否已过期
	if currentExpireTime.Valid && currentExpireTime.String != "" {
		expireTime, parseErr := time.ParseInLocation("2006-01-02 15:04:05", currentExpireTime.String, time.Local)
		if parseErr == nil && time.Now().After(expireTime) {
			global.GetLog(c).Warnf("更新公告被拒: 公告已过期 (User: %s, ID: %s, ExpireTime: %s)", currentUserCode, id, currentExpireTime.String)
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "更新失败：公告已过期，不允许修改"})
			return
		}
	}

	// ★★★ 检查分享码是否已过期 ★★★
	var shareCodeExpireTime sql.NullString
	err = global.DB.QueryRow(
		"SELECT expire_time FROM share_codes WHERE code = ? AND status = 1",
		shareCode,
	).Scan(&shareCodeExpireTime)

	if err != nil {
		global.GetLog(c).Errorf("查询分享码失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	if shareCodeExpireTime.Valid && shareCodeExpireTime.String != "" {
		expireTime, parseErr := time.ParseInLocation("2006-01-02 15:04:05", shareCodeExpireTime.String, time.Local)
		if parseErr == nil && time.Now().After(expireTime) {
			global.GetLog(c).Warnf("更新公告被拒: 分享码已过期 (User: %s, ID: %s, ShareCode: %s)", currentUserCode, id, shareCode)
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "更新失败：分享码已过期，不允许修改公告"})
			return
		}
	}

	res, err := global.DB.Exec(
		"UPDATE share_announcements SET note = ?, expire_time = ? WHERE id = ? AND creator_code = ? AND status = 1",
		req.Note, req.ExpireTime, id, currentUserCode,
	)

	if err != nil {
		global.GetLog(c).Errorf("更新公告DB错误 (ID: %s): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		global.GetLog(c).Warnf("更新公告被拒: 无权操作或不存在 (User: %s, ID: %s)", currentUserCode, id)
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "更新失败：公告不存在或您无权修改"})
		return
	}

	global.GetLog(c).Infof("用户[%s] 更新公告成功 (ID: %s)", currentUserCode, id)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// 辅助函数：简单的字符串时间清洗
func formatTimeStr(t string) string {
	if t == "" {
		return ""
	}
	if len(t) >= 19 {
		return t[:10] + " " + t[11:19]
	}
	return t
}
