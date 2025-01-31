package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
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
