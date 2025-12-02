package router

import (
	"practice_problems/api"
	"practice_problems/middleware"

	"github.com/gin-gonic/gin"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 允许所有来源
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func InitRouter() *gin.Engine {
	r := gin.Default()
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

			// 图片上传 (根据业务需求，通常建议放权鉴里，这里保持你原样)
			auth.POST("/upload", api.UploadImage)

			// ============================
			// 新增：分享与绑定接口
			// ============================
			auth.POST("/share/create", api.CreateShare) // 创建分享 (授权或生成码)
			auth.POST("/share/bind", api.BindSubject)   // 绑定资源 (输入码)
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
			auth.DELETE("/questions/:id", api.DeleteQuestion)
		}
	}

	return r
}
