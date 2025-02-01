package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func DeserializeJSON(c *gin.Context) (map[string]interface{}, error) {
	b, err := c.GetRawData()
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func ResponseOK(c *gin.Context, data interface{}, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": data,
		"msg":  msg,
	})
}

func ResponseFAIL(c *gin.Context, status int, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code": status,
		"data": nil,
		"msg":  msg,
	})
}
