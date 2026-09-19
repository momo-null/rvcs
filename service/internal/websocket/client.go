package websocket

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client WebSocket客户端
type Client struct {
	ID         string
	UserID     string
	DeviceID   string
	Connection *websocket.Conn
	Send       chan Message
	Hub        *Hub
	mu         sync.Mutex
}

// NewClient 创建新客户端
func NewClient(hub *Hub, deviceID string, conn *websocket.Conn) *Client {
	return &Client{
		ID:         uuid.New().String(),
		DeviceID:   deviceID,
		Connection: conn,
		Send:       make(chan Message, 256),
		Hub:        hub,
	}
}

// ReadPump 读取消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Connection.Close()
	}()

	c.Connection.SetReadLimit(512)

	for {
		var msg Message
		if err := c.Connection.ReadJSON(&msg); err != nil {
			break
		}

		// 处理消息
		c.Hub.MessageHandler(c, &msg)
	}
}

// WritePump 写入消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Connection.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Connection.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			// 发送ping
			c.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMsg 发送消息
func (c *Client) SendMsg(msg Message) {
	// 添加nil检查
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查连接是否有效
	if c.Connection == nil {
		return
	}

	// 发送消息，如果channel已满则丢弃（不关闭channel）
	select {
	case c.Send <- msg:
		// 发送成功
	default:
		// channel已满，记录日志但不关闭channel
		// 关闭channel会导致后续发送失败
	}
}
