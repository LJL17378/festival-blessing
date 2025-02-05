package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type FriendRequest struct {
	ID             int  `gorm:"primary_key"`
	FromID         int  `gorm:"not null"`
	ToID           int  `gorm:"not null"`
	AcceptedStatus bool `gorm:"default:false"`
}

type FriendRelationship struct {
	ID       int `gorm:"primary_key"`
	UserID   int `gorm:"not null"`
	FriendID int `gorm:"not null"`
}

func SendFriendRequest(c *gin.Context, db *gorm.DB) {
	fromID := c.GetInt("userID")
	var request struct {
		ToID int `json:"to_id"`
	}
	if err := c.ShouldBind(&request); err != nil {
		ResponseFAIL(c, http.StatusBadRequest, "Invalid input")
		return
	}
	var existingRequest FriendRequest
	if err := db.Where("from_id = ? and to_id = ? and accepted_status = 0", fromID, request.ToID).First(&existingRequest).Error; err == nil {
		ResponseFAIL(c, http.StatusBadRequest, "你已经发送过请求了")
		return
	}
	//创建好友请求
	friendRequest := FriendRequest{FromID: fromID, ToID: request.ToID}
	db.Create(&friendRequest)
	ResponseOK(c, nil, "已经发送添加请求")
}

func AcceptFriendRequest(c *gin.Context) {
	userID := c.GetInt("userID")
	var request struct {
		RequestID int `json:"request_id"`
	}
	if err := c.ShouldBind(&request); err != nil {
		ResponseFAIL(c, http.StatusBadRequest, "Invalid input")
		return
	}
	//查找好友请求
	var friendRequest FriendRequest
	if err := db.First(&friendRequest, request.RequestID).Error; err != nil || friendRequest.ToID != userID {
		ResponseFAIL(c, http.StatusNotFound, "没有找到请求")
		return
	}
	//接受请求
	db.Model(&friendRequest).Update("AcceptedStatus", true)
	db.Create(&FriendRelationship{UserID: friendRequest.FromID, FriendID: friendRequest.ToID})
	db.Create(&FriendRelationship{UserID: friendRequest.ToID, FriendID: friendRequest.FromID})

	ResponseOK(c, nil, "已接受好友请求")
}

func DeleteFriendRequest(c *gin.Context) {
	userID := c.GetInt("userID")
	var request struct {
		FriendID int `json:"friend_id"`
	}
	if err := c.ShouldBind(&request); err != nil {
		ResponseFAIL(c, http.StatusBadRequest, "Invalid input")
		return
	}

	db.Where("(user_id = ? and friend_id = ?) OR (friend_id = ? and user_id = ?)", userID, request.FriendID, userID, request.FriendID).Delete(&FriendRelationship{})
	ResponseOK(c, nil, "好友已经删除")
}

// 查询好友
func GetAllFriends(c *gin.Context) {
	userID := c.GetInt("userID")
	var friends []FriendRelationship
	err := db.Where("user_id = ?", userID).Find(&friends).Error
	if err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "获取好友失败")
		return
	}

	var friendList []struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	}

	for _, f := range friends {
		var friend User
		if f.UserID != userID {
			if err := db.First(&friend, f.UserID).Error; err == nil {
				friendList = append(friendList, struct {
					ID       int    `json:"id"`
					Username string `json:"username"`
				}{
					ID:       friend.ID,
					Username: friend.UserName,
				})
			}
		} else {
			if err := db.First(&friend, f.FriendID).Error; err == nil {
				friendList = append(friendList, struct {
					ID       int    `json:"id"`
					Username string `json:"username"`
				}{
					ID:       friend.ID,
					Username: friend.UserName,
				})
			}
		}
	}
	ResponseOK(c, gin.H{"friends": friendList}, "获取好友列表成功")
}

func GetAllReceivedFriendRequests(c *gin.Context) {
	userID := c.GetInt("userID")
	// 查询当前用户收到的所有好友请求，且未接受的
	var requests []FriendRequest
	err := db.Where("to_id = ? AND accepted_status = ?", userID, false).Find(&requests).Error
	if err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "获取好友请求失败")
		return
	}
	// 组合好友请求信息
	var requestList []gin.H
	for _, req := range requests {
		var user User
		if err := db.First(&user, req.FromID).Error; err != nil {
			continue // 如果找不到用户，则跳过
		}
		requestList = append(requestList, gin.H{
			"request_id": req.ID,
			"user_id":    user.ID,
			"username":   user.UserName,
		})
	}
	ResponseOK(c, gin.H{"requests": requestList}, "成功获取所有未接受的好友请求")
}
