package routes

import (
	"api-gateway/internal/handler"
	"api-gateway/middleware"
	"github.com/gin-gonic/gin"
)

func NewRouter(service ...interface{}) *gin.Engine {
	ginRouter := gin.Default()
	ginRouter.Use(middleware.Cors(), middleware.InitMiddleware(service), middleware.ErrorMiddleware())
	v1 := ginRouter.Group("/api/v1")
	{
		v1.GET("ping", func(context *gin.Context) { // test server
			context.JSON(200, "success")
		})
		v1.POST("/user/register", handler.UserRegister)
	}
	return ginRouter
}
