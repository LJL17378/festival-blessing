package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
)

var db *gorm.DB
var once sync.Once
var dbErr error

func InitDb() *gorm.DB {
	once.Do(func() {
		dsn := "root:123456@tcp(127.0.0.1:3306)/festival_blessing?charset=utf8mb4&parseTime=True&loc=Local"
		db, dbErr = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if dbErr != nil {
			panic("failed to connect database")
		}
	})
	return db
}
