package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	db := InitDb()

	r := gin.Default()

	// 公共路由
	r.POST("/register", func(c *gin.Context) { registerHandler(c, db) })
	r.POST("/login", func(c *gin.Context) { loginHandler(c, db) })

	// 需要认证的路由
	authGroup := r.Group("/")
	authGroup.Use(func(c *gin.Context) { authMiddleware(c, db) }) // 传入 db
	{
		authGroup.PUT("/profile", func(c *gin.Context) { updateProfileHandler(c, db) })
		authGroup.GET("/profile", func(c *gin.Context) { getProfileHandler(c, db) })
		authGroup.DELETE("/account", func(c *gin.Context) { deleteAccountHandler(c, db) })
	}
	r.Run(":8080")
}
