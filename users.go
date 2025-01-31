package main

import (
	"net/http"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码处理失败"})
		return
	}

	user := User{
		UserName: req.UserName,
		Password: hashedPassword,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       user.ID,
		"userName": user.UserName,
	})
}

// 登录处理函数
func loginHandler(c *gin.Context, db *gorm.DB) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User
	if err := db.Where("user_name = ?", req.UserName).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的凭证"})
		return
	}

	if !checkPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的凭证"})
		return
	}

	token, err := generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "令牌生成失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"userName": user.UserName,
		},
	})
}

func updateProfileHandler(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "个人信息更新成功"})
}

func getProfileHandler(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)

	var user User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
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

	c.JSON(http.StatusOK, response)
}

// 删除账户处理函数
func deleteAccountHandler(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)

	if err := db.Delete(&User{}, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "账户删除成功"})
}
