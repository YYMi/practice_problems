package router

import (
	"practice_problems/api"
	"practice_problems/middleware"

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

			// --- 题目 ---
			auth.GET("/questions", api.GetQuestionList)
			auth.POST("/questions", api.CreateQuestion)
			auth.PUT("/questions/:id", api.UpdateQuestion)
			// ★★★ 新增：修改用户题目备注 ★★★
			auth.POST("/questions/note", api.UpdateUserNote)
			auth.DELETE("/questions/:id", api.DeleteQuestion)
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
