package api

import (
	"database/sql"
	_ "database/sql"
	"encoding/json"
	"fmt"
	"practice_problems/global"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// sanitizeIdentifier 净化SQL标识符，防止SQL注入
func sanitizeIdentifier(identifier string) string {
	// 只允许字母、数字和下划线
	reg := regexp.MustCompile("[^a-zA-Z0-9_]")
	cleaned := reg.ReplaceAllString(identifier, "")

	// 不能以数字开头
	if len(cleaned) > 0 && cleaned[0] >= '0' && cleaned[0] <= '9' {
		cleaned = "_" + cleaned
	}

	// 不能为空
	if cleaned == "" {
		cleaned = "field"
	}

	return cleaned
}

// checkAdminModifyPermission 检查管理员修改权限
// 只有超级管理员（ID最小的管理员）可以修改管理员信息，其他管理员只能修改普通用户
func checkAdminModifyPermission(c *gin.Context, operatorID int, where map[string]interface{}) error {
	// 获取超级管理员（is_admin=1的用户中ID最小的那个）
	var superAdminID int
	err := global.DB.QueryRow("SELECT id FROM users WHERE is_admin = 1 ORDER BY id ASC LIMIT 1").Scan(&superAdminID)
	if err != nil {
		// 没有管理员，允许操作
		return nil
	}

	// 如果操作者是超级管理员，允许所有操作
	if operatorID == superAdminID {
		return nil
	}

	// 获取目标用户的ID
	var targetID interface{}
	if id, ok := where["id"]; ok {
		targetID = id
	} else {
		// 如果没有指定id，尝试通过username查找
		if username, ok := where["username"]; ok {
			var id int
			err := global.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&id)
			if err == nil {
				targetID = id
			}
		}
	}

	if targetID == nil {
		return nil // 无法确定目标用户，允许操作
	}

	// 检查目标用户是否是管理员
	var targetIsAdmin int
	var targetUsername string
	err = global.DB.QueryRow("SELECT username, is_admin FROM users WHERE id = ?", targetID).Scan(&targetUsername, &targetIsAdmin)
	if err != nil {
		return nil // 查询失败，允许操作
	}

	// 如果目标用户是管理员，拒绝操作
	if targetIsAdmin == 1 {
		global.GetLog(c).Warnf("权限拒绝: 用户ID[%d]尝试修改管理员[%s]的信息", operatorID, targetUsername)
		return fmt.Errorf("只有超级管理员可以修改管理员信息")
	}

	return nil
}

// ===============================
// 数据库管理API - 查询相关
// ===============================

// GetAllTables 获取所有表列表
func GetAllTables(c *gin.Context) {
	// 查询SQLite的所有表
	rows, err := global.DB.Query(`
		SELECT name FROM sqlite_master 
		WHERE type='table' AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`)
	if err != nil {
		global.GetLog(c).Errorf("查询表列表失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	// 需要过滤的系统表
	hiddenTables := map[string]bool{
		"table_comments":  true,
		"column_comments": true,
		"column_orders":   true,
	}

	tables := []map[string]interface{}{}
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}

		// 过滤系统表
		if hiddenTables[tableName] {
			continue
		}

		// 查询每个表的记录数
		var count int
		countErr := global.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
		if countErr != nil {
			count = 0
		}

		tables = append(tables, map[string]interface{}{
			"name":  tableName,
			"count": count,
		})
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": tables,
	})
}

// GetTableStructure 获取表结构
func GetTableStructure(c *gin.Context) {
	tableName := c.Param("table")
	if tableName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名不能为空"})
		return
	}

	// 使用PRAGMA获取表结构
	rows, err := global.DB.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		global.GetLog(c).Errorf("查询表结构失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	columns := []map[string]interface{}{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue interface{}

		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			continue
		}

		columns = append(columns, map[string]interface{}{
			"cid":      cid,
			"name":     name,
			"type":     ctype,
			"not_null": notnull == 1,
			"default":  dfltValue,
			"pk":       pk == 1,
		})
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": columns,
	})
}

// GetTableData 获取表数据（支持分页、字段选择、条件查询）
func GetTableData(c *gin.Context) {
	tableName := c.Param("table")
	if tableName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名不能为空"})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 获取需要查询的字段
	fields := c.DefaultQuery("fields", "*")
	if fields == "" {
		fields = "*"
	}

	// 获取查询条件（JSON格式）
	whereConditions := c.DefaultQuery("where", "")
	var whereClause string
	var args []interface{}

	if whereConditions != "" {
		var conditions map[string]map[string]interface{}
		if err := json.Unmarshal([]byte(whereConditions), &conditions); err == nil {
			whereParts := []string{}
			for key, condition := range conditions {
				operator, opExists := condition["operator"]
				value, valExists := condition["value"]

				if !opExists {
					continue
				}

				switch operator {
				case "eq": // 等于
					if valExists && value != "" {
						whereParts = append(whereParts, fmt.Sprintf("%s = ?", sanitizeIdentifier(key)))
						args = append(args, value)
					}
				case "ne": // 不等于
					if valExists && value != "" {
						whereParts = append(whereParts, fmt.Sprintf("%s != ?", sanitizeIdentifier(key)))
						args = append(args, value)
					}
				case "like": // 包含
					if valExists && value != "" {
						whereParts = append(whereParts, fmt.Sprintf("%s LIKE ?", sanitizeIdentifier(key)))
						args = append(args, "%"+fmt.Sprintf("%v", value)+"%")
					}
				case "starts": // 开头是
					if valExists && value != "" {
						whereParts = append(whereParts, fmt.Sprintf("%s LIKE ?", sanitizeIdentifier(key)))
						args = append(args, fmt.Sprintf("%v", value)+"%")
					}
				case "ends": // 结尾是
					if valExists && value != "" {
						whereParts = append(whereParts, fmt.Sprintf("%s LIKE ?", sanitizeIdentifier(key)))
						args = append(args, "%"+fmt.Sprintf("%v", value))
					}
				case "null": // 为空
					whereParts = append(whereParts, fmt.Sprintf("%s IS NULL", sanitizeIdentifier(key)))
				case "notnull": // 不为空
					whereParts = append(whereParts, fmt.Sprintf("%s IS NOT NULL", sanitizeIdentifier(key)))
				}
			}
			if len(whereParts) > 0 {
				whereClause = " WHERE " + strings.Join(whereParts, " AND ")
			}
		}
	}

	// 查询总数
	var total int
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", tableName, whereClause)
	if err := global.DB.QueryRow(countSQL, args...).Scan(&total); err != nil {
		global.GetLog(c).Errorf("查询总数失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	// 查询数据
	dataSQL := fmt.Sprintf("SELECT %s FROM %s%s LIMIT ? OFFSET ?", fields, tableName, whereClause)
	args = append(args, pageSize, offset)

	rows, err := global.DB.Query(dataSQL, args...)
	if err != nil {
		global.GetLog(c).Errorf("查询数据失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		global.GetLog(c).Errorf("获取列名失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}

	// 读取数据
	data := []map[string]interface{}{}
	for rows.Next() {
		// 创建接收数据的切片
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		// 构建map
		rowData := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			rowData[col] = v
		}
		data = append(data, rowData)
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"list":      data,
		},
	})
}

// ===============================
// 数据库管理API - 修改相关（需要reCAPTCHA）
// ===============================

// InsertTableRow 插入数据
func InsertTableRow(c *gin.Context) {
	tableName := c.Param("table")
	if tableName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名不能为空"})
		return
	}

	var req struct {
		Data           map[string]interface{} `json:"data" binding:"required"`
		RecaptchaToken string                 `json:"recaptcha_token"` // 已在中间件验证
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 验证TOTP
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	if err := VerifyTotpForOperation(userID.(int), req.RecaptchaToken); err != nil {
		global.GetLog(c).Warnf("数据插入失败，TOTP验证错误: %v", err)
		c.JSON(403, gin.H{"code": 403, "msg": "Google验证码错误，请检查后重试"})
		return
	}

	// 构建插入SQL
	fields := []string{}
	placeholders := []string{}
	values := []interface{}{}

	for key, value := range req.Data {
		fields = append(fields, key)
		placeholders = append(placeholders, "?")
		values = append(values, value)
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := global.DB.Exec(insertSQL, values...)
	if err != nil {
		global.GetLog(c).Errorf("插入数据失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "插入失败: " + err.Error()})
		return
	}

	lastID, _ := result.LastInsertId()
	global.GetLog(c).Infof("管理员插入数据: 表=%s, ID=%d", tableName, lastID)

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "插入成功",
		"data": gin.H{"id": lastID},
	})
}

// UpdateTableRow 更新数据
func UpdateTableRow(c *gin.Context) {
	tableName := c.Param("table")
	if tableName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名不能为空"})
		return
	}

	var req struct {
		Where          map[string]interface{} `json:"where" binding:"required"`
		Data           map[string]interface{} `json:"data" binding:"required"`
		RecaptchaToken string                 `json:"recaptcha_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 验证TOTP
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	if err := VerifyTotpForOperation(userID.(int), req.RecaptchaToken); err != nil {
		global.GetLog(c).Warnf("数据更新失败，TOTP验证错误: %v", err)
		c.JSON(403, gin.H{"code": 403, "msg": "Google验证码错误，请检查后重试"})
		return
	}

	// 如果是 users 表，检查权限：只有 admin 账号可以修改管理员信息
	if tableName == "users" {
		if err := checkAdminModifyPermission(c, userID.(int), req.Where); err != nil {
			c.JSON(403, gin.H{"code": 403, "msg": err.Error()})
			return
		}
	}

	// 构建SET子句
	setParts := []string{}
	setValues := []interface{}{}
	for key, value := range req.Data {
		setParts = append(setParts, fmt.Sprintf("%s = ?", key))
		setValues = append(setValues, value)
	}

	// 构建WHERE子句
	whereParts := []string{}
	whereValues := []interface{}{}
	for key, value := range req.Where {
		whereParts = append(whereParts, fmt.Sprintf("%s = ?", key))
		whereValues = append(whereValues, value)
	}

	if len(whereParts) == 0 {
		c.JSON(400, gin.H{"code": 400, "msg": "必须指定WHERE条件"})
		return
	}

	updateSQL := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		tableName,
		strings.Join(setParts, ", "),
		strings.Join(whereParts, " AND "),
	)

	values := append(setValues, whereValues...)
	result, err := global.DB.Exec(updateSQL, values...)
	if err != nil {
		global.GetLog(c).Errorf("更新数据失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "更新失败: " + err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	global.GetLog(c).Infof("管理员更新数据: 表=%s, 影响行数=%d", tableName, affected)

	// 如果是users表且status被设置为1（禁用），立即清除该用户的token
	if tableName == "users" && affected > 0 {
		if status, ok := req.Data["status"]; ok {
			// 检查status是否为1（禁用）
			isDisabled := false
			switch v := status.(type) {
			case float64:
				isDisabled = v == 1
			case int:
				isDisabled = v == 1
			case string:
				isDisabled = v == "1"
			}
			if isDisabled {
				// 获取被禁用用户的user_code并清除token
				if targetID, ok := req.Where["id"]; ok {
					var userCode string
					err := global.DB.QueryRow("SELECT user_code FROM users WHERE id = ?", targetID).Scan(&userCode)
					if err == nil && userCode != "" {
						global.ClearUserTokens(userCode)
						global.GetLog(c).Infof("用户被禁用，已清除token: userCode=%s", userCode)
					}
				}
			}
		}
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "更新成功",
		"data": gin.H{"affected": affected},
	})
}

// DeleteTableRows 删除数据
func DeleteTableRows(c *gin.Context) {
	tableName := c.Param("table")
	if tableName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名不能为空"})
		return
	}

	var req struct {
		Where          map[string]interface{} `json:"where" binding:"required"`
		RecaptchaToken string                 `json:"recaptcha_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 验证TOTP
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	if err := VerifyTotpForOperation(userID.(int), req.RecaptchaToken); err != nil {
		global.GetLog(c).Warnf("数据删除失败，TOTP验证错误: %v", err)
		c.JSON(403, gin.H{"code": 403, "msg": "Google验证码错误，请检查后重试"})
		return
	}

	// 如果是 users 表，检查权限：只有 admin 账号可以删除管理员
	if tableName == "users" {
		if err := checkAdminModifyPermission(c, userID.(int), req.Where); err != nil {
			c.JSON(403, gin.H{"code": 403, "msg": err.Error()})
			return
		}
	}

	// 构建WHERE子句
	whereParts := []string{}
	values := []interface{}{}
	for key, value := range req.Where {
		whereParts = append(whereParts, fmt.Sprintf("%s = ?", key))
		values = append(values, value)
	}

	if len(whereParts) == 0 {
		c.JSON(400, gin.H{"code": 400, "msg": "必须指定WHERE条件"})
		return
	}

	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE %s",
		tableName,
		strings.Join(whereParts, " AND "),
	)

	result, err := global.DB.Exec(deleteSQL, values...)
	if err != nil {
		global.GetLog(c).Errorf("删除数据失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "删除失败: " + err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	global.GetLog(c).Infof("管理员删除数据: 表=%s, 影响行数=%d", tableName, affected)

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "删除成功",
		"data": gin.H{"affected": affected},
	})
}

// BatchUpdateTableRows 批量更新数据
func BatchUpdateTableRows(c *gin.Context) {
	tableName := c.Param("table")
	if tableName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名不能为空"})
		return
	}

	var req struct {
		Items          []map[string]interface{} `json:"items" binding:"required"`
		PrimaryKey     string                   `json:"primary_key" binding:"required"`
		RecaptchaToken string                   `json:"recaptcha_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 验证TOTP
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	if err := VerifyTotpForOperation(userID.(int), req.RecaptchaToken); err != nil {
		global.GetLog(c).Warnf("批量更新失败，TOTP验证错误: %v", err)
		c.JSON(403, gin.H{"code": 403, "msg": "Google验证码错误，请检查后重试"})
		return
	}

	// 开启事务
	tx, err := global.DB.Begin()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "事务开启失败"})
		return
	}
	defer tx.Rollback()

	affected := int64(0)
	for _, item := range req.Items {
		pkValue, ok := item[req.PrimaryKey]
		if !ok {
			continue
		}

		delete(item, req.PrimaryKey) // 移除主键，避免更新主键

		setParts := []string{}
		values := []interface{}{}
		for key, value := range item {
			setParts = append(setParts, fmt.Sprintf("%s = ?", key))
			values = append(values, value)
		}
		values = append(values, pkValue)

		updateSQL := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?",
			tableName,
			strings.Join(setParts, ", "),
			req.PrimaryKey,
		)

		result, err := tx.Exec(updateSQL, values...)
		if err != nil {
			global.GetLog(c).Errorf("批量更新失败: %v", err)
			c.JSON(500, gin.H{"code": 500, "msg": "更新失败: " + err.Error()})
			return
		}

		rows, _ := result.RowsAffected()
		affected += rows
	}

	if err := tx.Commit(); err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "事务提交失败"})
		return
	}

	global.GetLog(c).Infof("管理员批量更新: 表=%s, 影响行数=%d", tableName, affected)

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "批量更新成功",
		"data": gin.H{"affected": affected},
	})
}

// BatchDeleteTableRows 批量删除数据
func BatchDeleteTableRows(c *gin.Context) {
	tableName := c.Param("table")
	if tableName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名不能为空"})
		return
	}

	var req struct {
		Ids            []interface{} `json:"ids" binding:"required"`
		PrimaryKey     string        `json:"primary_key" binding:"required"`
		RecaptchaToken string        `json:"recaptcha_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 验证TOTP
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	if err := VerifyTotpForOperation(userID.(int), req.RecaptchaToken); err != nil {
		global.GetLog(c).Warnf("批量删除失败，TOTP验证错误: %v", err)
		c.JSON(403, gin.H{"code": 403, "msg": "Google验证码错误，请检查后重试"})
		return
	}

	if len(req.Ids) == 0 {
		c.JSON(400, gin.H{"code": 400, "msg": "删除ID列表不能为空"})
		return
	}

	// 构建IN子句
	placeholders := strings.Repeat("?,", len(req.Ids))
	placeholders = placeholders[:len(placeholders)-1]

	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE %s IN (%s)",
		tableName,
		req.PrimaryKey,
		placeholders,
	)

	result, err := global.DB.Exec(deleteSQL, req.Ids...)
	if err != nil {
		global.GetLog(c).Errorf("批量删除失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "删除失败: " + err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	global.GetLog(c).Infof("管理员批量删除: 表=%s, 影响行数=%d", tableName, affected)

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "批量删除成功",
		"data": gin.H{"affected": affected},
	})
}

// GetTableComment 获取表备注
func GetTableComment(c *gin.Context) {
	tableName := c.Param("table")
	if tableName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名不能为空"})
		return
	}

	var comment string
	err := global.DB.QueryRow("SELECT comment FROM table_comments WHERE table_name = ?", tableName).Scan(&comment)
	if err != nil {
		if err == sql.ErrNoRows {
			comment = ""
		} else {
			global.GetLog(c).Errorf("查询表备注失败: %v", err)
			c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
			return
		}
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": comment,
	})
}

// SetTableComment 设置表备注
func SetTableComment(c *gin.Context) {
	tableName := c.Param("table")
	if tableName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名不能为空"})
		return
	}

	var req struct {
		Comment        string `json:"comment"`
		RecaptchaToken string `json:"recaptcha_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 验证TOTP
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	if err := VerifyTotpForOperation(userID.(int), req.RecaptchaToken); err != nil {
		global.GetLog(c).Warnf("设置表备注失败，TOTP验证错误: %v", err)
		c.JSON(403, gin.H{"code": 403, "msg": "Google验证码错误，请检查后重试"})
		return
	}

	// 插入或更新备注
	_, err := global.DB.Exec(`
		INSERT INTO table_comments (table_name, comment) VALUES (?, ?)
		ON CONFLICT(table_name) DO UPDATE SET comment = ?, update_time = CURRENT_TIMESTAMP
	`, tableName, req.Comment, req.Comment)
	if err != nil {
		global.GetLog(c).Errorf("设置表备注失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "设置失败: " + err.Error()})
		return
	}

	global.GetLog(c).Infof("管理员设置表备注: 表=%s", tableName)

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "设置成功",
	})
}

// GetColumnComment 获取字段备注
func GetColumnComment(c *gin.Context) {
	tableName := c.Param("table")
	columnName := c.Param("column")

	if tableName == "" || columnName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名和字段名不能为空"})
		return
	}

	var comment string
	err := global.DB.QueryRow("SELECT comment FROM column_comments WHERE table_name = ? AND column_name = ?", tableName, columnName).Scan(&comment)
	if err != nil {
		if err == sql.ErrNoRows {
			comment = ""
		} else {
			global.GetLog(c).Errorf("查询字段备注失败: %v", err)
			c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
			return
		}
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": comment,
	})
}

// SetColumnComment 设置字段备注
func SetColumnComment(c *gin.Context) {
	tableName := c.Param("table")
	columnName := c.Param("column")

	if tableName == "" || columnName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名和字段名不能为空"})
		return
	}

	var req struct {
		Comment        string `json:"comment"`
		RecaptchaToken string `json:"recaptcha_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 验证TOTP
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	if err := VerifyTotpForOperation(userID.(int), req.RecaptchaToken); err != nil {
		global.GetLog(c).Warnf("设置字段备注失败，TOTP验证错误: %v", err)
		c.JSON(403, gin.H{"code": 403, "msg": "Google验证码错误，请检查后重试"})
		return
	}

	// 插入或更新备注
	_, err := global.DB.Exec(`
		INSERT INTO column_comments (table_name, column_name, comment) VALUES (?, ?, ?)
		ON CONFLICT(table_name, column_name) DO UPDATE SET comment = ?, update_time = CURRENT_TIMESTAMP
	`, tableName, columnName, req.Comment, req.Comment)
	if err != nil {
		global.GetLog(c).Errorf("设置字段备注失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "设置失败: " + err.Error()})
		return
	}

	global.GetLog(c).Infof("管理员设置字段备注: 表=%s, 字段=%s", tableName, columnName)

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "设置成功",
	})
}

// GetAllTableComments 获取所有表备注
func GetAllTableComments(c *gin.Context) {
	rows, err := global.DB.Query("SELECT table_name, comment FROM table_comments")
	if err != nil {
		global.GetLog(c).Errorf("查询所有表备注失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	comments := make(map[string]string)
	for rows.Next() {
		var tableName, comment string
		if err := rows.Scan(&tableName, &comment); err != nil {
			continue
		}
		comments[tableName] = comment
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": comments,
	})
}

// GetAllColumnComments 获取所有字段备注
func GetAllColumnComments(c *gin.Context) {
	rows, err := global.DB.Query("SELECT table_name, column_name, comment FROM column_comments")
	if err != nil {
		global.GetLog(c).Errorf("查询所有字段备注失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	comments := make(map[string]map[string]string)
	for rows.Next() {
		var tableName, columnName, comment string
		if err := rows.Scan(&tableName, &columnName, &comment); err != nil {
			continue
		}

		if comments[tableName] == nil {
			comments[tableName] = make(map[string]string)
		}
		comments[tableName][columnName] = comment
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": comments,
	})
}

// ===============================
// 字段管理API
// ===============================

// AddColumn 添加字段
func AddColumn(c *gin.Context) {
	tableName := c.Param("table")
	if tableName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名不能为空"})
		return
	}

	var req struct {
		ColumnName     string `json:"column_name" binding:"required"`
		ColumnType     string `json:"column_type" binding:"required"`
		DefaultValue   string `json:"default_value"`
		RecaptchaToken string `json:"recaptcha_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 验证TOTP
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	if err := VerifyTotpForOperation(userID.(int), req.RecaptchaToken); err != nil {
		global.GetLog(c).Warnf("添加字段失败，TOTP验证错误: %v", err)
		c.JSON(403, gin.H{"code": 403, "msg": "Google验证码错误，请检查后重试"})
		return
	}

	// 净化标识符
	safeTableName := sanitizeIdentifier(tableName)
	safeColumnName := sanitizeIdentifier(req.ColumnName)
	safeColumnType := sanitizeIdentifier(req.ColumnType)

	// 构建SQL
	var alterSQL string
	if req.DefaultValue != "" {
		// 处理默认值
		defaultVal := req.DefaultValue
		// 如果是文本类型，需要加引号
		if strings.ToUpper(safeColumnType) == "TEXT" || strings.ToUpper(safeColumnType) == "VARCHAR" {
			defaultVal = "'" + strings.ReplaceAll(defaultVal, "'", "''") + "'"
		}
		alterSQL = fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s DEFAULT %s",
			safeTableName, safeColumnName, safeColumnType, defaultVal)
	} else {
		alterSQL = fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
			safeTableName, safeColumnName, safeColumnType)
	}

	_, err := global.DB.Exec(alterSQL)
	if err != nil {
		global.GetLog(c).Errorf("添加字段失败: %v, SQL: %s", err, alterSQL)
		c.JSON(500, gin.H{"code": 500, "msg": "添加字段失败: " + err.Error()})
		return
	}

	global.GetLog(c).Infof("管理员添加字段: 表=%s, 字段=%s, 类型=%s", safeTableName, safeColumnName, safeColumnType)

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "添加字段成功",
	})
}

// DropColumn 删除字段
func DropColumn(c *gin.Context) {
	tableName := c.Param("table")
	columnName := c.Param("column")
	if tableName == "" || columnName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名和字段名不能为空"})
		return
	}

	var req struct {
		RecaptchaToken string `json:"recaptcha_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 验证TOTP
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	if err := VerifyTotpForOperation(userID.(int), req.RecaptchaToken); err != nil {
		global.GetLog(c).Warnf("删除字段失败，TOTP验证错误: %v", err)
		c.JSON(403, gin.H{"code": 403, "msg": "Google验证码错误，请检查后重试"})
		return
	}

	// 净化标识符
	safeTableName := sanitizeIdentifier(tableName)
	safeColumnName := sanitizeIdentifier(columnName)

	// 构建SQL
	alterSQL := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", safeTableName, safeColumnName)

	_, err := global.DB.Exec(alterSQL)
	if err != nil {
		global.GetLog(c).Errorf("删除字段失败: %v, SQL: %s", err, alterSQL)
		c.JSON(500, gin.H{"code": 500, "msg": "删除字段失败: " + err.Error()})
		return
	}

	// 同时删除字段备注
	_, _ = global.DB.Exec("DELETE FROM column_comments WHERE table_name = ? AND column_name = ?", safeTableName, safeColumnName)
	// 同时删除字段排序
	_, _ = global.DB.Exec("DELETE FROM column_orders WHERE table_name = ? AND column_name = ?", safeTableName, safeColumnName)

	global.GetLog(c).Infof("管理员删除字段: 表=%s, 字段=%s", safeTableName, safeColumnName)

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "删除字段成功",
	})
}

// ===============================
// 字段排序API
// ===============================

// GetColumnOrders 获取字段排序
func GetColumnOrders(c *gin.Context) {
	tableName := c.Param("table")
	if tableName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名不能为空"})
		return
	}

	rows, err := global.DB.Query("SELECT column_name, sort_order FROM column_orders WHERE table_name = ? ORDER BY sort_order", tableName)
	if err != nil {
		global.GetLog(c).Errorf("查询字段排序失败: %v", err)
		c.JSON(500, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	orders := make(map[string]int)
	for rows.Next() {
		var columnName string
		var sortOrder int
		if err := rows.Scan(&columnName, &sortOrder); err != nil {
			continue
		}
		orders[columnName] = sortOrder
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": orders,
	})
}

// SaveColumnOrders 保存字段排序
func SaveColumnOrders(c *gin.Context) {
	tableName := c.Param("table")
	if tableName == "" {
		c.JSON(400, gin.H{"code": 400, "msg": "表名不能为空"})
		return
	}

	var req struct {
		Orders         []string `json:"orders" binding:"required"` // 字段名数组，按顺序排列
		RecaptchaToken string   `json:"recaptcha_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	// 验证TOTP
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"code": 401, "msg": "未登录"})
		return
	}

	if err := VerifyTotpForOperation(userID.(int), req.RecaptchaToken); err != nil {
		global.GetLog(c).Warnf("保存字段排序失败，TOTP验证错误: %v", err)
		c.JSON(403, gin.H{"code": 403, "msg": "Google验证码错误，请检查后重试"})
		return
	}

	// 开启事务
	tx, err := global.DB.Begin()
	if err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "事务开启失败"})
		return
	}
	defer tx.Rollback()

	// 遍历字段数组，保存排序
	for i, columnName := range req.Orders {
		_, err := tx.Exec(`
			INSERT INTO column_orders (table_name, column_name, sort_order) VALUES (?, ?, ?)
			ON CONFLICT(table_name, column_name) DO UPDATE SET sort_order = ?, update_time = CURRENT_TIMESTAMP
		`, tableName, columnName, i, i)
		if err != nil {
			global.GetLog(c).Errorf("保存字段排序失败: %v", err)
			c.JSON(500, gin.H{"code": 500, "msg": "保存失败: " + err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "事务提交失败"})
		return
	}

	global.GetLog(c).Infof("管理员保存字段排序: 表=%s, 字段数=%d", tableName, len(req.Orders))

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "保存成功",
	})
}
