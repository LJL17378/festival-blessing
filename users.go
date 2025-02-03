package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 用户模型
type User struct {
	ID        int            `gorm:"primaryKey;autoIncrement"`
	UserName  string         `gorm:"type:varchar(100);not null;unique" json:"userName" binding:"required"`
	Password  string         `gorm:"type:varchar(100);not null" json:"-"`
	NickName  string         `gorm:"type:varchar(100)" json:"nickName"`
	Age       int            `gorm:"default:0" json:"age"`
	Birthday  string         `gorm:"type:varchar(50)" json:"birthday"`
	Gender    string         `gorm:"type:varchar(10)" json:"gender"`
	Interests string         `gorm:"type:text" json:"interests"`
	Status    string         `gorm:"type:varchar(50)" json:"status"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 注册请求结构体
type RegisterRequest struct {
	UserName string `json:"userName" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 登录请求结构体
type LoginRequest struct {
	UserName string `json:"userName" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 更新个人信息请求结构体
type UpdateProfileRequest struct {
	NickName  string `json:"nickName"`
	Age       int    `json:"age"`
	Birthday  string `json:"birthday"`
	Gender    string `json:"gender"`
	Interests string `json:"interests"`
	Status    string `json:"status"`
}

// 注册处理函数
func registerHandler(c *gin.Context, db *gorm.DB) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseFAIL(c, http.StatusBadRequest, err.Error())
		return
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, err.Error())
		return
	}

	user := User{
		UserName: req.UserName,
		Password: hashedPassword,
	}

	if err := db.Create(&user).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, err.Error())
		return
	}
	ResponseOK(c, gin.H{
		"id":       user.ID,
		"userName": user.UserName,
	}, "注册成功")
}

// 登录处理函数
func loginHandler(c *gin.Context, db *gorm.DB) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseFAIL(c, http.StatusBadRequest, err.Error())
		return
	}

	var user User
	if err := db.Where("user_name = ?", req.UserName).First(&user).Error; err != nil {
		ResponseFAIL(c, http.StatusUnauthorized, "无效的凭证")
		return
	}

	if !checkPassword(req.Password, user.Password) {
		ResponseFAIL(c, http.StatusUnauthorized, "无效的凭证")
		return
	}

	token, err := generateToken(user.ID)
	if err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "令牌生成失败")
		return
	}

	ResponseOK(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"userName": user.UserName,
		},
	}, "登录成功")
}

func updateProfileHandler(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseFAIL(c, http.StatusBadRequest, err.Error())
		return
	}

	updates := map[string]interface{}{
		"nick_name": req.NickName,
		"age":       req.Age,
		"gender":    req.Gender,
		"interests": req.Interests,
		"status":    req.Status,
		"birthday":  req.Birthday,
	}

	if err := db.Model(&User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "更新失败")
		return
	}
	ResponseOK(c, updates, "个人信息更新成功")
}

func getProfileHandler(c *gin.Context, db *gorm.DB) {
	userID, _ := strconv.Atoi(c.Query("userID"))
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		ResponseFAIL(c, http.StatusNotFound, "用户不存在")
		return
	}

	response := gin.H{
		"userName":  user.UserName,
		"nickName":  user.NickName,
		"age":       user.Age,
		"gender":    user.Gender,
		"interests": user.Interests,
		"status":    user.Status,
		"birthday":  user.Birthday,
	}
	ResponseOK(c, response, "获取成功")
}

// 删除账户处理函数
func deleteAccountHandler(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)

	if err := db.Delete(&User{}, userID).Error; err != nil {
		ResponseFAIL(c, http.StatusInternalServerError, "删除失败")
		return
	}
	ResponseOK(c, nil, "账户删除成功")
}
