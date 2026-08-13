package routers

import (
	"reimburse-system/handlers"
	"reimburse-system/web"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS 中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		// 报销记录 CRUD
		api.GET("/reimbursements", handlers.ListReimbursements)
		api.GET("/reimbursements/:id", handlers.GetReimbursement)
		api.POST("/reimbursements", handlers.CreateReimbursement)
		api.PUT("/reimbursements/:id", handlers.UpdateReimbursement)
		api.DELETE("/reimbursements/:id", handlers.DeleteReimbursement)

		// 状态管理
		api.PATCH("/reimbursements/:id/status", handlers.UpdateStatus)
		api.PATCH("/reimbursements/persons/:person_id/status", handlers.UpdatePersonStatus)

		// 辅助接口
		api.GET("/persons/names", handlers.GetPersonNames)
		api.GET("/reasons/options", handlers.GetReasonOptions)
		api.GET("/statistics", handlers.GetStatistics)
	}

	// 挂载内嵌前端（需使用 -tags embed 构建；开发模式为空实现）
	web.Register(r)

	return r
}
