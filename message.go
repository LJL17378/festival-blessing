package main

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// 消息结构体
type Message struct {
	FROM    int
	To      int    `json:"to"`
	Content string `json:"content"`
}

// 数据库消息模型
type ChatMessage struct {
	ID         int `gorm:"primaryKey"`
	SenderID   int `gorm:"index"`
	ReceiverID int `gorm:"index"`
	Content    string
	CreatedAt  time.Time `gorm:"index"`
}

var (
	clients  = sync.Map{} // 在线用户连接池 map[userID]*websocket.Conn
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
)

// WebSocket处理
func wsHandler(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	// 升级WebSocket连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	// 注册连接并处理离线消息
	clients.Store(userID, conn)
	defer clients.Delete(userID)
	go pullOfflineMessages(userID, conn) // 拉取离线消息
	// 处理消息
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}
		msg.FROM = userID
		handleIncomingMessage(userID, msg)
	}
}

// 处理收到的消息
func handleIncomingMessage(senderID int, msg Message) {
	// 持久化到数据库
	db.Create(&ChatMessage{
		SenderID:   senderID,
		ReceiverID: msg.To,
		Content:    msg.Content,
		CreatedAt:  time.Now(),
	})

	// 检查接收方是否在线
	if targetConn, ok := clients.Load(msg.To); ok {
		// 在线：直接发送
		targetConn.(*websocket.Conn).WriteJSON(msg)
	} else {
		// 离线：存入Redis List
		msgData, _ := json.Marshal(msg)
		rdb.RPush(context.Background(), "offline:"+strconv.Itoa(msg.To), msgData)
		rdb.Expire(context.Background(), "offline:"+strconv.Itoa(msg.To), 7*24*time.Hour)
	}
}

// 拉取离线消息
func pullOfflineMessages(userID int, conn *websocket.Conn) {
	key := "offline:" + strconv.Itoa(userID)
	messages, err := rdb.LRange(context.Background(), key, 0, -1).Result()
	if err != nil || len(messages) == 0 {
		return
	}

	// 发送并清空
	for _, msg := range messages {
		conn.WriteMessage(websocket.TextMessage, []byte(msg))
	}
	rdb.Del(context.Background(), key)
}
