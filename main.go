package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	db := InitDb()

	r := gin.Default()
	r.Static("/avatar", "./avatar")

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
		authGroup.POST("/posts", func(c *gin.Context) { createPostHandler(c, db) })
		authGroup.GET("/posts", func(c *gin.Context) { getPostsHandler(c, db) })
		authGroup.POST("/posts/:id/like", func(c *gin.Context) { likePostHandler(c, db) })
		authGroup.POST("/posts/:id/unlike", func(c *gin.Context) { unlikePostHandler(c, db) })
		authGroup.GET("/posts/liked", func(c *gin.Context) { getLikedPostsHandler(c, db) }) // 查询某人已点赞的帖子
		authGroup.POST("/posts/:id/comments", func(c *gin.Context) { createCommentHandler(c, db) })
		authGroup.GET("/posts/:id/comments", func(c *gin.Context) { getCommentsHandler(c, db) })
		authGroup.POST("/comments/:id/like", func(c *gin.Context) { likeCommentHandler(c, db) })     // 点赞评论
		authGroup.POST("/comments/:id/unlike", func(c *gin.Context) { unlikeCommentHandler(c, db) }) // 取消点赞评
		authGroup.POST("/friend/request", func(c *gin.Context) { SendFriendRequest(c, db) })
		authGroup.POST("/friend/accept", AcceptFriendRequest)
		authGroup.POST("/friend/delete", DeleteFriendRequest)
		authGroup.GET("/friend/list", GetAllFriends)
		authGroup.GET("/friend/getrequests", GetAllReceivedFriendRequests)
		authGroup.POST("/avatar/upload", func(c *gin.Context) { UploadAvatar(c, db) })
		authGroup.POST("/blessings", SendBlessings)                // 发送祝福
		authGroup.GET("/blessings/sent", GetSentBlessings)         // 查询自己发出的祝福
		authGroup.GET("/blessings/received", GetReceivedBlessings) // 查询自己收到的祝福
	}
	r.Run(":8080")
}
