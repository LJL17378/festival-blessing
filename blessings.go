package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Blessing struct {
	ID         int       `gorm:"primary_key"`
	SenderID   int       `gorm:"not null"`
	ReceiverID *int      `gorm:"default:null"`
	Content    string    `gorm:"not null"`
	Font       string    `gorm:"not null"`
	PaperStyle string    `gorm:"not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func SendBlessings(c *gin.Context) {
	var req struct {
		ReceiverID *int   `json:"receiver_id"`
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
	if count == 0 && req.ReceiverID != nil {
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

	if blessing.ReceiverID != nil {
		ResponseOK(c, gin.H{"blessings_sent": blessing}, "祝福发送成功")
	} else {
		ResponseOK(c, gin.H{"blessing_sent": blessing}, "已将此祝福存储至草稿箱")
	}
}

// 查询自己发送的祝福
func GetSentBlessings(c *gin.Context) {
	userID := c.GetInt("userID")

	var blessings []struct {
		ID           int    `json:"id"`
		ReceiverID   int    `json:"receiver_id"`
		ReceiverName string `json:"receiver_name"`
		Content      string `json:"content"`
		Font         string `json:"font"`
		PaperStyle   string `json:"paper_style"`
		Timestamp    string `json:"timestamp"`
	}
	err := db.Table("blessings").
		Select("blessings.id, blessings.receiver_id, users.user_name as receiver_name, blessings.content, blessings.font, blessings.paper_style, blessings.created_at as timestamp").
		Joins("JOIN users ON blessings.receiver_id = users.id").
		Where("blessings.sender_id = ?", userID).
		Order("blessings.created_at DESC").
		Scan(&blessings).Error

	if err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "无法获取发出的祝福")
		return
	}

	var drafts []struct {
		ID         int    `json:"id"`
		ReceiverID *int   `json:"receiver_id"`
		Content    string `json:"content"`
		Font       string `json:"font"`
		PaperStyle string `json:"paper_style"`
		Timestamp  string `json:"timestamp"`
	}

	err = db.Table("blessings").
		Select("blessings.id, blessings.receiver_id, blessings.content, blessings.font, blessings.paper_style, blessings.created_at as timestamp").
		Where("(blessings.sender_id = ?) and (blessings.receiver_id is null)", userID).
		Order("blessings.created_at DESC").
		Scan(&drafts).Error

	if err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "无法获取草稿箱")
		return
	}

	ResponseOK(c, gin.H{
		"sent_blessings": blessings,
		"草稿箱":            drafts,
	}, "查询祝福成功")
}

// 查询自己收到的祝福
func GetReceivedBlessings(c *gin.Context) {
	userID := c.GetInt("userID")

	var blessings []struct {
		ID         int    `json:"id"`
		SenderID   uint   `json:"sender_id"`
		SenderName string `json:"sender_name"`
		Content    string `json:"content"`
		Font       string `json:"font"`
		PaperStyle string `json:"paper_style"`
		Timestamp  string `json:"timestamp"`
	}

	err := db.Table("blessings").
		Select("blessings.id, blessings.sender_id, users.user_name as sender_name, blessings.content, blessings.font, blessings.paper_style, blessings.created_at as timestamp").
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

func ShareBlessings(c *gin.Context) {
	var req struct {
		ReceiverID *int   `json:"receiver_id"`
		Content    string `json:"content"`
		Font       string `json:"font"`
		PaperStyle string `json:"paper_style"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseFAIL(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	senderID := c.GetInt("userID")

	blessing := Blessing{
		SenderID:   senderID,
		ReceiverID: nil,
		Content:    req.Content,
		Font:       req.Font,
		PaperStyle: req.PaperStyle,
	}

	if err := db.Create(&blessing).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "分享祝福失败")
		return
	}

	ResponseOK(c, gin.H{
		"blessing_id": blessing.ID,
	}, "您可以以此id分享祝福")
}

func ReceiveByLink(c *gin.Context) {
	blessingID := c.Query("id")
	userID := c.GetInt("userID")

	var blessing Blessing
	if err := db.First(&blessing, blessingID).Error; err != nil {
		ResponseFAIL(c, http.StatusNotFound, "没有查询到这条祝福")
	}

	if blessing.ReceiverID == nil {
		blessing.ReceiverID = &userID
		if err := db.Save(&blessing).Error; err != nil {
			ResponseFAIL(c, http.StatusInternalServerError, "接收祝福失败")
			return
		}
	} else if *blessing.ReceiverID != userID {
		ResponseFAIL(c, http.StatusForbidden, "您无权查看已发送给别人的祝福")
		return
	}

	ResponseOK(c, gin.H{
		"blessing_id": blessing.ID,
		"sender_id":   blessing.SenderID,
		"receiver_id": blessing.ReceiverID,
		"content":     blessing.Content,
		"font":        blessing.Font,
		"paper_style": blessing.PaperStyle,
		"created_at":  blessing.CreatedAt,
	}, "成功获取祝福")
}
