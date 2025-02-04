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
	Font       string    `gorm:"not null"`
	PaperStyle string    `gorm:"not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func SendBlessings(c *gin.Context) {
	var req struct {
		ReceiverID int    `json:"receiver_id"`
		Content    string `json:"content"`
		Font       string `json:"font"`
		PaperStyle string `json:"paper_style"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseFAIL(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	senderID := c.GetInt("userID")

	//确保只能发给好友
	var count int64
	db.Table("friend_relationships").Where("(user_id = ? and friend_id = ?) or (user_id = ? and friend_id = ?)", senderID, req.ReceiverID, req.ReceiverID, senderID).Count(&count)
	if count == 0 {
		ResponseFAIL(c, http.StatusForbidden, "对方不是你的好友")
		return
	}

	blessing := Blessing{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Content:    req.Content,
		Font:       req.Font,
		PaperStyle: req.PaperStyle,
	}

	if err := db.Create(&blessing).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "发送祝福失败")
		return
	}

	ResponseOK(c, gin.H{"blessings_sent": blessing}, "祝福发送成功")
}

// 查询自己发送的祝福
func GetSentBlessings(c *gin.Context) {
	userID := c.GetInt("userID")

	var blessings []struct {
		ReceiverID   int    `json:"receiver_id"`
		ReceiverName string `json:"receiver_name"`
		Content      string `json:"content"`
		Font         string `json:"font"`
		PaperStyle   string `json:"paper_style"`
		Timestamp    string `json:"timestamp"`
	}
	err := db.Table("blessings").
		Select("blessings.receiver_id, users.user_name as receiver_name, blessings.content, blessings.font, blessings.paper_style, blessings.created_at as timestamp").
		Joins("JOIN users ON blessings.receiver_id = users.id").
		Where("blessings.sender_id = ?", userID).
		Order("blessings.created_at DESC").
		Scan(&blessings).Error

	if err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "无法获取发出的祝福")
		return
	}

	ResponseOK(c, gin.H{
		"sent_blessings": blessings,
	}, "查询祝福成功")
}

// 查询自己收到的祝福
func GetReceivedBlessings(c *gin.Context) {
	userID := c.GetInt("userID")

	var blessings []struct {
		SenderID   uint   `json:"sender_id"`
		SenderName string `json:"sender_name"`
		Content    string `json:"content"`
		Font       string `json:"font"`
		PaperStyle string `json:"paper_style"`
		Timestamp  string `json:"timestamp"`
	}

	err := db.Table("blessings").
		Select("blessings.sender_id, users.user_name as sender_name, blessings.content, blessings.font, blessings.paper_style, blessings.created_at as timestamp").
		Joins("JOIN users ON blessings.sender_id = users.id").
		Where("blessings.receiver_id = ?", userID).
		Order("blessings.created_at DESC").
		Scan(&blessings).Error

	if err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "无法获取收到的祝福")
		return
	}

	ResponseOK(c, gin.H{
		"received_blessings": blessings,
	}, "查询成功")
}
