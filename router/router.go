package router

import (
	"practice_problems/api"
	"practice_problems/middleware"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	// 使用 gin.New()，跳过默认的 Logger 和 Recovery，我们需要手动配置
	r := gin.New()

	// 1. ★★★ RequestID 中间件 (必须放在第一个) ★★★
	// 它负责生成 ID，后续的 Logger 才能拿到
	r.Use(middleware.RequestIDMiddleware())

	// 2. ★★★ 自定义 Zap 日志中间件 (替代 gin.Logger()) ★★★
	// 这样请求日志格式就和业务日志完全统一了
	r.Use(middleware.GinLogger())

	// 3. 崩溃恢复中间件 (防止程序 Panic 挂掉)
	r.Use(gin.Recovery())

	// 4. 跨域中间件
	r.Use(corsMiddleware())

	// ★★★ WebSocket 路由 (不能使用 gzip，必须在 gzip 中间件之前注册) ★★★
	r.GET("/api/v1/ws/ai-interview", api.AIInterviewWebSocket)

	// 5. gzip 压缩中间件 (放在 WebSocket 路由之后)
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// 静态资源
	r.Static("/uploads", "./uploads")

	v1 := r.Group("/api/v1")
	{
		// ============================
		// 公开接口 (无需 Token)
		// ============================
		// 用户认证
		v1.POST("/auth/register", api.CreateUser) // 创建用户
		v1.POST("/auth/login", api.UserLogin)     // 用户登录 (含空密码逻辑)

		// ============================
		// 需要 JWT 认证的接口
		// ============================
		auth := v1.Group("/")
		auth.Use(middleware.JWTAuthMiddleware()) // 👈 挂载 JWT 中间件
		{
			// 用户相关
			auth.PUT("/user/profile", api.UpdateUser) // 修改用户信息/密码
			auth.POST("/auth/logout", api.UserLogout)

			// TOTP相关（谷歌验证码）
			auth.GET("/totp/check", api.CheckTotpBound)        // 检查是否已绑定
			auth.GET("/totp/generate", api.GenerateTotpSecret) // 生成密钥和二维码
			auth.POST("/totp/bind", api.VerifyTotpCode)        // 验证并绑定
			auth.POST("/totp/verify", api.ValidateTotpCode)    // 验证TOTP码
			auth.POST("/totp/unbind", api.UnbindTotp)          // 解绑TOTP

			// 图片上传
			auth.POST("/upload", api.UploadImage)

			// 公告相关接口
			auth.POST("/share/announcement", api.CreateShareAnnouncement)
			auth.GET("/share/announcements", api.GetShareAnnouncementList)
			auth.DELETE("/share/announcement/:id", api.DeleteShareAnnouncement)
			auth.PUT("/share/announcement/:id", api.UpdateShareAnnouncement)

			// ============================
			// 分享与绑定接口
			// ============================
			auth.POST("/share/create", api.CreateShare) // 创建分享
			auth.POST("/share/bind", api.BindSubject)   // 绑定资源
			auth.GET("/share/list", api.GetMyShareCodes)
			auth.DELETE("/share/:id", api.DeleteShareCode)
			auth.PUT("/share/:id", api.UpdateShareCode)

			// --- 科目 ---
			auth.GET("/subjects", api.GetSubjectList)
			auth.GET("/subjects/:id", api.GetSubjectDetail)
			auth.POST("/subjects", api.CreateSubject)
			auth.PUT("/subjects/:id", api.UpdateSubject)
			auth.DELETE("/subjects/:id", api.DeleteSubject)
			auth.GET("/subject/:id/users", api.GetSubjectAuthorizedUsers)
			auth.PUT("/auth/:id", api.UpdateSubjectAuth)
			auth.DELETE("/auth/:id", api.RemoveSubjectAuth)
			auth.PUT("/auth/batch/update", api.BatchUpdateAuth)
			auth.PUT("/auth/batch/remove", api.BatchRemoveAuth)

			// --- 分类 ---
			auth.GET("/categories", api.GetCategoryList)
			auth.POST("/categories", api.CreateCategory)
			auth.PUT("/categories/:id", api.UpdateCategory)
			auth.DELETE("/categories/:id", api.DeleteCategory)
			auth.POST("/categories/:id/sort", api.UpdateCategorySort)

			// --- 知识点 ---
			auth.GET("/points", api.GetPointList)
			auth.GET("/points/:id", api.GetPointDetail)
			auth.POST("/points", api.CreatePoint)
			auth.PUT("/points/:id", api.UpdatePoint)
			auth.DELETE("/points/:id", api.DeletePoint)
			auth.DELETE("/points/:id/image", api.DeletePointImage)
			auth.PUT("/points/:id/sort", api.UpdatePointSort)

			// --- 知识点笔记 ---
			auth.GET("/points/:id/note", api.GetPointNote)   // 获取知识点笔记
			auth.POST("/points/:id/note", api.SavePointNote) // 保存知识点笔记

			// --- 知识点绑定 ---
			auth.POST("/point-bindings", api.CreateBinding)
			auth.GET("/point-bindings/:pointId", api.GetBindingsByPoint)
			auth.DELETE("/point-bindings/:id", api.DeleteBinding)
			auth.GET("/binding/subjects/:subjectId/categories", api.GetCategoriesBySubjectForBinding)
			auth.GET("/binding/categories/:categoryId/points", api.GetPointsByCategoryForBinding)

			// --- 题目 ---
			auth.GET("/questions", api.GetQuestionList)
			auth.POST("/questions", api.CreateQuestion)
			auth.PUT("/questions/:id", api.UpdateQuestion)
			// ★★★ 新增：修改用户题目备注 ★★★
			auth.POST("/questions/note", api.UpdateUserNote)
			auth.DELETE("/questions/:id", api.DeleteQuestion)

			// --- 集合 ---
			auth.GET("/collections", api.GetCollections)                                // 获取集合列表
			auth.POST("/collections", api.CreateCollection)                             // 创建集合
			auth.PUT("/collections/:id", api.UpdateCollection)                          // 更新集合
			auth.DELETE("/collections/:id", api.DeleteCollection)                       // 删除集合
			auth.GET("/collections/:id/points", api.GetCollectionPoints)                // 获取集合的知识点列表（分页）
			auth.GET("/collections/:id/points/:pointId", api.GetCollectionPointDetail)  // 获取集合中知识点详情
			auth.GET("/collections/:id/questions", api.GetCollectionQuestions)          // 获取集合中所有题目（综合刷题）
			auth.POST("/collections/points", api.AddPointToCollection)                  // 添加知识点到集合
			auth.POST("/collections/points/batch", api.BatchAddPointsToCollection)      // 批量添加知识点到集合（科目/分类级别）
			auth.GET("/collections/point-collections", api.GetPointCollections)         // 获取知识点已绑定的集合列表
			auth.DELETE("/collections/items/:id", api.RemovePointFromCollection)        // 从集合中移除知识点
			auth.PUT("/collections/items/order", api.UpdateCollectionItemsOrder)        // 更新集合项排序
			auth.PUT("/collections/:id/permission", api.SetCollectionPermission)        // 设置集合权限（公有/私有）
			auth.POST("/collections/:id/permissions", api.AddCollectionPermission)      // 添加集合授权
			auth.GET("/collections/:id/permissions", api.GetCollectionPermissions)      // 获取集合授权列表
			auth.PUT("/collections/:id/permissions", api.UpdateCollectionPermission)    // 更新集合授权时间
			auth.DELETE("/collections/:id/permissions", api.DeleteCollectionPermission) // 删除集合授权
			auth.GET("/collections/find-point", api.FindPointInCollections)             // 查找知识点在哪个集合中

			// ============================
			// 数据库管理接口（仅管理员）
			// ============================
			admin := auth.Group("/admin")
			admin.Use(middleware.AdminMiddleware()) // 管理员权限中间件
			{
				// 查询相关（不需要reCAPTCHA）
				admin.GET("/db/tables", api.GetAllTables)                                    // 获取所有表
				admin.GET("/db/tables/:table/structure", api.GetTableStructure)              // 获取表结构
				admin.GET("/db/tables/:table/data", api.GetTableData)                        // 获取表数据
				admin.GET("/db/tables/:table/comment", api.GetTableComment)                  // 获取表备注
				admin.GET("/db/tables/:table/columns/:column/comment", api.GetColumnComment) // 获取字段备注
				admin.GET("/db/table-comments", api.GetAllTableComments)                     // 获取所有表备注
				admin.GET("/db/column-comments", api.GetAllColumnComments)                   // 获取所有字段备注

				// 修改相关（需要reCAPTCHA验证）
				// 注意：reCAPTCHA中间件会消耗request body，所以这里不使用中间件
				// 而是在各个API内部检查recaptcha_token字段
				admin.POST("/db/tables/:table/insert", api.InsertTableRow)                    // 插入数据
				admin.PUT("/db/tables/:table/update", api.UpdateTableRow)                     // 更新数据
				admin.DELETE("/db/tables/:table/delete", api.DeleteTableRows)                 // 删除数据
				admin.PUT("/db/tables/:table/batch-update", api.BatchUpdateTableRows)         // 批量更新
				admin.DELETE("/db/tables/:table/batch-delete", api.BatchDeleteTableRows)      // 批量删除
				admin.POST("/db/tables/:table/comment", api.SetTableComment)                  // 设置表备注
				admin.POST("/db/tables/:table/columns/:column/comment", api.SetColumnComment) // 设置字段备注
				admin.POST("/db/tables/:table/columns", api.AddColumn)                        // 添加字段
				admin.DELETE("/db/tables/:table/columns/:column", api.DropColumn)             // 删除字段
				admin.GET("/db/tables/:table/column-orders", api.GetColumnOrders)             // 获取字段排序
				admin.POST("/db/tables/:table/column-orders", api.SaveColumnOrders)           // 保存字段排序
			}
		}
	}

	return r
}

// corsMiddleware 跨域中间件 (保持你原有的逻辑)
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
