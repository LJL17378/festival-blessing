package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Blessing struct {
	ID         int       `gorm:"primary_key"`
	SenderID   int       `gorm:"not null"`
	ReceiverID int       `gorm:"not null"`
	Content    string    `gorm:"not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func SendBlessings(c *gin.Context) {
	var req struct {
		ReceiverID int    `json:"receiver_id"`
		Content    string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	senderID := c.GetInt("userID")

	//确保只能发给好友
	var count int64
	db.Table("friend_relationships").Where("(user_id = ? and friend_id = ?) or (user_id = ? and friend_id = ?)", senderID, req.ReceiverID, req.ReceiverID, senderID).Count(&count)
	if count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "对方不是你的好友"})
		return
	}

	blessing := Blessing{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
	}

	if err := db.Create(&blessing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送祝福失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "祝福发送成功"})
}

// 查询自己发送的祝福
func GetSentBlessings(c *gin.Context) {
	userID := c.GetInt("userID")

	var blessings []struct {
		ReceiverID   int    `json:"receiver_id"`
		ReceiverName string `json:"receiver_name"`
		Content      string `json:"content"`
		Timestamp    string `json:"timestamp"`
	}
	err := db.Table("blessings").
		Select("blessings.receiver_id, users.user_name as receiver_name, blessings.content, blessings.created_at as timestamp").
		Joins("JOIN users ON blessings.receiver_id = users.id").
		Where("blessings.sender_id = ?", userID).
		Order("blessings.created_at DESC").
		Scan(&blessings).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sent blessings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sent_blessings": blessings})
}

// 查询自己收到的祝福
func GetReceivedBlessings(c *gin.Context) {
	userID := c.GetInt("userID")

	var blessings []struct {
		SenderID   uint   `json:"sender_id"`
		SenderName string `json:"sender_name"`
		Content    string `json:"content"`
		Timestamp  string `json:"timestamp"`
	}

	err := db.Table("blessings").
		Select("blessings.sender_id, users.user_name as sender_name, blessings.content, blessings.created_at as timestamp").
		Joins("JOIN users ON blessings.sender_id = users.id").
		Where("blessings.receiver_id = ?", userID).
		Order("blessings.created_at DESC").
		Scan(&blessings).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve received blessings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"received_blessings": blessings})
}
