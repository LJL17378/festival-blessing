package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

// 说说模型
type Post struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int       `gorm:"not null" json:"userId"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	LikeCount int       `gorm:"default:0" json:"likeCount"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

type PostView struct {
	Post
	NickName string `json:"nickName"`
	IsLiked  bool   `json:"isLiked"`
}

// 点赞模型
type Like struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID    int       `gorm:"not null" json:"postId"`
	UserID    int       `gorm:"not null" json:"userId"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// 发布说说请求结构体
type CreatePostRequest struct {
	Content string `json:"content" binding:"required"`
}

// 发布说说
func createPostHandler(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)

	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseFAIL(c, http.StatusBadRequest, err.Error())
		return
	}

	post := Post{
		UserID:  userID,
		Content: req.Content,
	}

	if err := db.Create(&post).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseOK(c, gin.H{
		"id":        post.ID,
		"content":   post.Content,
		"createdAt": post.CreatedAt,
	}, "发布成功")
}

// 查询说说
func getPostsHandler(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)

	var posts []PostView
	if err := db.Table("posts").
		Select("posts.*, users.nick_name").
		Joins("LEFT JOIN users ON posts.user_id = users.id").
		Order("posts.created_at DESC").
		Find(&posts).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 检查当前用户是否已经点赞
	for i := range posts {
		var like Like
		if err := db.Where("post_id = ? AND user_id = ?", posts[i].ID, userID).First(&like).Error; err == nil {
			posts[i].IsLiked = true
		} else {
			posts[i].IsLiked = false
		}
	}

	ResponseOK(c, posts, "查询成功")
}

// 点赞说说
func likePostHandler(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)
	postIDStr := c.Param("id")

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "post id is invalid",
		})
		return
	}

	// 检查是否已经点赞
	var like Like
	if err := db.Where("post_id = ? AND user_id = ?", postID, userID).First(&like).Error; err == nil {
		ResponseFAIL(c, 500, "已经点赞")
		return
	}

	// 创建点赞记录
	like = Like{
		PostID: postID,
		UserID: userID,
	}
	if err := db.Create(&like).Error; err != nil {
		ResponseFAIL(c, 500, err.Error())
		return
	}

	// 更新点赞数
	if err := db.Model(&Post{}).Where("id = ?", postID).Update("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
		ResponseFAIL(c, 500, err.Error())
		return
	}

	ResponseOK(c, like, "点赞成功")
}

// 取消点赞
func unlikePostHandler(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)
	postIDStr := c.Param("id")

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		ResponseFAIL(c, http.StatusBadRequest, err.Error())
		return
	}
	// 查询是否存在点赞
	var like Like
	if err := db.Where("post_id = ? AND user_id = ?", postID, userID).First(&like).Error; err != nil {
		ResponseFAIL(c, 500, "不存在点赞记录，无法取消点赞")
		return
	}
	// 删除点赞记录
	if err := db.Where("post_id = ? AND user_id = ?", postID, userID).Delete(&Like{}).Error; err != nil {
		ResponseFAIL(c, 500, "取消点赞失败")
		return
	}

	// 更新点赞数
	if err := db.Model(&Post{}).Where("id = ?", postID).Update("like_count", gorm.Expr("like_count - 1")).Error; err != nil {
		ResponseFAIL(c, 500, "更新点赞数失败")
		return
	}

	ResponseOK(c, nil, "点赞成功")
}

// 查询某人已点赞的帖子
func getLikedPostsHandler(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)

	var likes []Like
	if err := db.Where("user_id = ?", userID).Find(&likes).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 提取帖子ID
	var postIDs []int
	for _, like := range likes {
		postIDs = append(postIDs, like.PostID)
	}

	// 查询帖子详情
	var posts []PostView
	if err := db.Table("posts").
		Select("posts.*, users.nick_name").
		Where("posts.id IN ?", postIDs).
		Joins("LEFT JOIN users ON posts.user_id = users.id").
		Order("posts.created_at DESC").
		Find(&posts).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 设置 IsLiked 字段为 true（因为这些帖子是用户已经点赞的）
	for i := range posts {
		posts[i].IsLiked = true
	}

	ResponseOK(c, posts, "查询成功")
}
