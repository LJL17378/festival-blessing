package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

// 评论模型
type Comment struct {
	ID        int            `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID    int            `gorm:"not null" json:"postId"`
	UserID    int            `gorm:"not null" json:"userId"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	LikeCount int            `gorm:"default:0" json:"likeCount"`
	ParentID  *int           `gorm:"default:null" json:"parentId"` // 父评论ID
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type CommentView struct {
	Comment
	NickName string `json:"nickName"`
	IsLiked  bool   `json:"isLiked"`
}

type CommentLike struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	CommentID int       `gorm:"not null" json:"commentId"`
	UserID    int       `gorm:"not null" json:"userId"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// 发布评论请求结构体
type CreateCommentRequest struct {
	Content  string `json:"content" binding:"required"`
	ParentID *int   `json:"parentId"` // 父评论ID（可选）
}

// 发布评论
func createCommentHandler(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)
	postIDStr := c.Param("id")

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		ResponseFAIL(c, http.StatusBadRequest, "帖子ID无效")
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseFAIL(c, http.StatusBadRequest, err.Error())
		return
	}

	comment := Comment{
		PostID:   postID,
		UserID:   userID,
		Content:  req.Content,
		ParentID: req.ParentID, // 父评论ID
	}

	if err := db.Create(&comment).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseOK(c, gin.H{
		"id":        comment.ID,
		"content":   comment.Content,
		"parentId":  comment.ParentID,
		"createdAt": comment.CreatedAt,
	}, "评论发布成功")
}

// 点赞评论
func likeCommentHandler(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)
	commentIDStr := c.Param("id")

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		ResponseFAIL(c, http.StatusBadRequest, "评论ID无效")
		return
	}

	// 检查是否已经点赞
	var like CommentLike
	if err := db.Where("comment_id = ? AND user_id = ?", commentID, userID).First(&like).Error; err == nil {
		ResponseFAIL(c, http.StatusBadRequest, "已经点赞")
		return
	}

	// 创建点赞记录
	like = CommentLike{
		CommentID: commentID,
		UserID:    userID,
	}
	if err := db.Create(&like).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "点赞失败")
		return
	}

	// 更新点赞数
	if err := db.Model(&Comment{}).Where("id = ?", commentID).Update("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "更新点赞数失败")
		return
	}

	ResponseOK(c, nil, "点赞成功")
}

// 取消点赞评论
func unlikeCommentHandler(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)
	commentIDStr := c.Param("id")

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		ResponseFAIL(c, http.StatusBadRequest, "评论ID无效")
		return
	}

	// 检查是否已经点赞
	var like CommentLike
	if err := db.Where("comment_id = ? AND user_id = ?", commentID, userID).First(&like).Error; err != nil {
		ResponseFAIL(c, http.StatusBadRequest, "不存在点赞记录，无法取消点赞")
		return
	}

	// 删除点赞记录
	if err := db.Where("comment_id = ? AND user_id = ?", commentID, userID).Delete(&CommentLike{}).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "取消点赞失败")
		return
	}

	// 更新点赞数
	if err := db.Model(&Comment{}).Where("id = ?", commentID).Update("like_count", gorm.Expr("like_count - 1")).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "更新点赞数失败")
		return
	}

	ResponseOK(c, nil, "取消点赞成功")
}

// 查询评论
func getCommentsHandler(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)
	postID := c.Param("id")

	var comments []CommentView
	if err := db.Table("comments").
		Select("comments.*, users.nick_name").
		Where("post_id = ?", postID).
		Joins("LEFT JOIN users ON comments.user_id = users.id").
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, err.Error())
		return
	}

	for i := range comments {
		var like CommentLike
		if err := db.Where("comment_id = ? AND user_id = ?", comments[i].ID, userID).First(&like).Error; err == nil {
			comments[i].IsLiked = true
		} else {
			comments[i].IsLiked = false
		}
	}

	ResponseOK(c, comments, "查询评论成功")
}
