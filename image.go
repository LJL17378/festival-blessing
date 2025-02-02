package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"path/filepath"
)

type Avatar struct {
	gorm.Model
	UserId   int    `gorm:"not null"`
	Filename string `gorm:"unique"`
	URL      string
}

func UploadAvatar(c *gin.Context, db *gorm.DB) {
	userID := c.MustGet("userID").(int)
	uploadDir := "avatar"
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", userID, ext)

	// 保存文件
	dst := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 保存到数据库
	var image Avatar

	if err := db.Where("user_id = ?", userID).First(&image).Error; err != nil {
		// 如果没有找到记录, 创建一个新的记录
		image = Avatar{
			UserId:   userID,
			Filename: filename,
			URL:      "/avatar/" + filename,
		}
		db.Create(&image)
	} else {
		image = Avatar{
			UserId:   userID,
			Filename: filename,
			URL:      "/avatar/" + filename,
		}
		db.Save(&image)
	}

	ResponseOK(c, gin.H{
		"url": image.URL,
	}, "null")
}
