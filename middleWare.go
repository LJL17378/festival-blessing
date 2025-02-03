package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecret = []byte("your_secret_key")

// 生成Token
func generateToken(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString(jwtSecret)
}

// 密码哈希
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// 验证密码
func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// 认证中间件
func authMiddleware(c *gin.Context, db *gorm.DB) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		ResponseFAIL(c, http.StatusUnauthorized, "缺少认证令牌")
		c.Abort()
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("无效的签名方法")
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		ResponseFAIL(c, http.StatusUnauthorized, "无效的令牌")
		c.Abort()
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		ResponseFAIL(c, http.StatusUnauthorized, "无效的令牌声明")
		c.Abort()
		return
	}

	userID := int(claims["user_id"].(float64))
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		ResponseFAIL(c, http.StatusUnauthorized, "用户不存在")
		c.Abort()
		return
	}

	c.Set("userID", userID)
	c.Next()
}
