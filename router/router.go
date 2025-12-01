package router

import (
	"practice_problems/api"

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

	// ---------------------------------------------------------
	// 【新增 1】静态资源映射
	// 这样访问 http://localhost:8080/uploads/xxx 就能看到图片了
	// ---------------------------------------------------------
	r.Static("/uploads", "./uploads")

	v1 := r.Group("/api/v1")
	{
		// -----------------------------------------------------
		// 【新增 2】图片上传接口
		// -----------------------------------------------------
		v1.POST("/upload", api.UploadImage)

		// --- 科目相关 ---
		v1.GET("/subjects", api.GetSubjectList)
		v1.GET("/subjects/:id", api.GetSubjectDetail)
		v1.POST("/subjects", api.CreateSubject)
		v1.PUT("/subjects/:id", api.UpdateSubject)
		v1.DELETE("/subjects/:id", api.DeleteSubject)

		// --- 分类相关 ---
		v1.GET("/categories", api.GetCategoryList)
		v1.POST("/categories", api.CreateCategory)
		v1.PUT("/categories/:id", api.UpdateCategory)
		v1.DELETE("/categories/:id", api.DeleteCategory)
		v1.PUT("/:id/sort", api.UpdateCategorySort)

		// --- 知识点相关 ---
		v1.GET("/points", api.GetPointList)
		v1.GET("/points/:id", api.GetPointDetail)
		v1.POST("/points", api.CreatePoint)
		v1.PUT("/points/:id", api.UpdatePoint)
		v1.DELETE("/points/:id", api.DeletePoint)
		v1.DELETE("/points/:id/image", api.DeletePointImage)

		// 题目相关
		v1.GET("/questions", api.GetQuestionList)
		v1.POST("/questions", api.CreateQuestion)
		v1.PUT("/questions/:id", api.UpdateQuestion)
		v1.DELETE("/questions/:id", api.DeleteQuestion)
	}

	return r
}
