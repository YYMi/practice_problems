package api

import (
	"database/sql"
	"log"
	"net/http"
	"practice_problems/global" // 保持原样
	"practice_problems/model"  // 保持原样
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetCategoryList 获取分类列表
func GetCategoryList(c *gin.Context) {
	subjectIDStr := c.Query("subject_id")

	var rows *sql.Rows
	var err error

	// ✅ 修改点：在这里加上 difficulty
	fields := "id, subject_id, categorie_name, create_time, update_time, sort_order, difficulty"

	if subjectIDStr != "" {
		// 查指定科目
		// 建议加上 id DESC 作为第二排序，保证同样排序值的最新创建在上面
		sqlStr := "SELECT " + fields + " FROM knowledge_categories WHERE subject_id = ? ORDER BY sort_order ASC, id DESC"
		rows, err = global.DB.Query(sqlStr, subjectIDStr)
	} else {
		// 查所有
		sqlStr := "SELECT " + fields + " FROM knowledge_categories ORDER BY sort_order ASC, id DESC"
		rows, err = global.DB.Query(sqlStr)
	}

	if err != nil {
		log.Println("查询分类失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询失败"})
		return
	}
	defer rows.Close()

	list := make([]model.KnowledgeCategory, 0)

	for rows.Next() {
		var item model.KnowledgeCategory
		// 这里的 Scan 顺序必须和上面的 fields 一模一样
		err := rows.Scan(
			&item.ID,
			&item.SubjectID,
			&item.CategoryName,
			&item.CreateTime,
			&item.UpdateTime,
			&item.SortOrder,
			&item.Difficulty, // ✅ 现在查了7个，接7个，就对应上了
		)

		if err != nil {
			log.Println("扫描数据失败:", err)
			continue
		}
		list = append(list, item)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": list})
}

// CreateCategory 创建分类
func CreateCategory(c *gin.Context) {
	var req model.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// --- 步骤 1: 计算新数据的 sort_order ---
	// 逻辑：找出当前科目下最小的 sort_order，然后 -1。
	// 如果没有任何数据，默认给 0。
	var currentMinSort int
	// COALESCE 是 SQL 函数：如果 MIN 返回 NULL，就用默认值 0
	sqlQueryMin := "SELECT COALESCE(MIN(sort_order), 0) FROM knowledge_categories WHERE subject_id = ?"

	err := global.DB.QueryRow(sqlQueryMin, req.SubjectID).Scan(&currentMinSort)
	if err != nil {
		// 理论上不太会报错，除非数据库连接挂了
		log.Println("获取最小排序失败:", err)
		currentMinSort = 0
	}

	newSortOrder := currentMinSort - 1

	// --- 步骤 2: 插入数据 ---
	// 注意：这里 SQL 语句里多加了 sort_order 字段
	sqlStr := "INSERT INTO knowledge_categories (subject_id, categorie_name, sort_order,difficulty) VALUES (?, ?, ?,?)"

	result, err := global.DB.Exec(sqlStr, req.SubjectID, req.CategoryName, newSortOrder, 0)

	if err != nil {
		log.Println("创建分类失败:", err)
		if strings.Contains(err.Error(), "FOREIGN KEY") {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建失败：所属科目不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建失败"})
		return
	}

	newID, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功", "data": gin.H{"id": newID}})
}

// UpdateCategory 修改分类名称
func UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req model.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 1. 动态构建 SQL
	// 只要有更新，update_time 就应该刷新
	query := "UPDATE knowledge_categories SET update_time = CURRENT_TIMESTAMP"
	var args []interface{}

	// 只有当 CategoryName 不为空时，才更新它
	if req.CategoryName != "" {
		query += ", categorie_name = ?"
		args = append(args, req.CategoryName)
	}

	// 只有当 Difficulty 指针不为 nil 时，才更新它 (这样可以安全地更新为 0)
	if req.Difficulty != nil {
		// 做一个简单的合法性校验 (0-3)
		if *req.Difficulty < 0 || *req.Difficulty > 3 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "难度值必须在 0-3 之间"})
			return
		}
		query += ", difficulty = ?"
		args = append(args, *req.Difficulty)
	}

	// 如果没有字段需要更新 (参数都没传)，直接返回成功或报错
	if len(args) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "无变更"})
		return
	}

	// 加上 WHERE 条件
	query += " WHERE id = ?"
	args = append(args, id)

	// 2. 执行 SQL
	_, err := global.DB.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// DeleteCategory 删除分类
func DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	sqlStr := "DELETE FROM knowledge_categories WHERE id = ?"
	result, err := global.DB.Exec(sqlStr, id)

	if err != nil {
		log.Println("删除分类失败:", err)
		// 【优化点】：SQLite 外键约束拦截时的特殊处理
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败：该分类下仍有知识点，请先清理知识点"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该分类"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// UpdateCategorySortRequest 排序请求参数
type UpdateCategorySortRequest struct {
	Action string `json:"action" binding:"required,oneof=top up down"` // 限制只能传 top, up, down
}

// UpdateCategorySort 更新分类排序
// 路由建议: PUT /api/v1/categories/:id/sort
func UpdateCategorySort(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req UpdateCategorySortRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: action必须为 top, up, down"})
		return
	}

	// 1. 查询当前数据的信息 (我们需要知道它属于哪个科目，以及当前的排序值)
	var currentSubjectID int
	var currentSortOrder int
	err := global.DB.QueryRow("SELECT subject_id, sort_order FROM knowledge_categories WHERE id = ?", id).Scan(&currentSubjectID, &currentSortOrder)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "未找到该分类"})
		return
	}

	// 开启事务 (因为交换涉及两条数据的修改，必须保证原子性)
	tx, err := global.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "系统错误"})
		return
	}
	// 发生 panic 或 return error 时回滚，成功则提交
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 2. 根据动作执行逻辑
	switch req.Action {
	case "top":
		// === 置顶 ===
		// 找到当前最小的 sort_order
		var minSort int
		_ = tx.QueryRow("SELECT COALESCE(MIN(sort_order), 0) FROM knowledge_categories WHERE subject_id = ?", currentSubjectID).Scan(&minSort)

		// 设置为 最小 - 1
		newSort := minSort - 1
		_, err = tx.Exec("UPDATE knowledge_categories SET sort_order = ? WHERE id = ?", newSort, id)

	case "up":
		// === 上移 ===
		// 寻找“排在它前面”的最近一条数据
		// 逻辑：同科目，sort_order 比它小，且是其中最大的那个
		var targetID, targetSort int
		err = tx.QueryRow(`
			SELECT id, sort_order FROM knowledge_categories 
			WHERE subject_id = ? AND sort_order < ? 
			ORDER BY sort_order DESC LIMIT 1`, currentSubjectID, currentSortOrder).Scan(&targetID, &targetSort)

		if err == sql.ErrNoRows {
			// 前面没有数据了，说明已经是第一名，无需移动
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已经是第一位了"})
			return
		}

		// 交换排序值
		// 1. 把上面的变成我的排序值
		_, _ = tx.Exec("UPDATE knowledge_categories SET sort_order = ? WHERE id = ?", currentSortOrder, targetID)
		// 2. 把我的变成上面的排序值
		_, err = tx.Exec("UPDATE knowledge_categories SET sort_order = ? WHERE id = ?", targetSort, id)

	case "down":
		// === 下移 ===
		// 寻找“排在它后面”的最近一条数据
		// 逻辑：同科目，sort_order 比它大，且是其中最小的那个
		var targetID, targetSort int
		err = tx.QueryRow(`
			SELECT id, sort_order FROM knowledge_categories 
			WHERE subject_id = ? AND sort_order > ? 
			ORDER BY sort_order ASC LIMIT 1`, currentSubjectID, currentSortOrder).Scan(&targetID, &targetSort)

		if err == sql.ErrNoRows {
			// 后面没有数据了，说明已经是最后一名
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已经是最后一位了"})
			return
		}

		// 交换排序值
		_, _ = tx.Exec("UPDATE knowledge_categories SET sort_order = ? WHERE id = ?", currentSortOrder, targetID)
		_, err = tx.Exec("UPDATE knowledge_categories SET sort_order = ? WHERE id = ?", targetSort, id)
	}

	if err != nil {
		log.Println("排序更新失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "排序更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "排序成功"})
}
