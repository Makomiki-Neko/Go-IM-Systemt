// internal/hub/hub.go 保存和管理各用户的WS长连接
package hub

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Client 代表一个活跃的WebSocket连接
type Client struct {
	UserId uint64
	Conn   *websocket.Conn
	Send   chan []byte // 待发送消息的缓冲通道
}

// Hub 管理所有在线客户端
type Hub struct {
	// 保护 clients 的读写锁
	Mu sync.RWMutex
	// 存储在线用户: map[userId]*Client
	Clients map[uint64]*Client
	// 注册和注销通道
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uint64]*Client),
		Register:   make(chan *Client, 64),
		Unregister: make(chan *Client, 64),
	}
}

// Run 是核心循环，处理注册/注销
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mu.Lock()
			// 如果已存在，先关闭旧连接
			if old, ok := h.Clients[client.UserId]; ok {
				old.Conn.Close()
			}
			h.Clients[client.UserId] = client
			h.Mu.Unlock()
		case client := <-h.Unregister:
			h.Mu.Lock()
			if _, ok := h.Clients[client.UserId]; ok {
				delete(h.Clients, client.UserId)
				close(client.Send)
			}
			h.Mu.Unlock()
		}
	}
}

// SendToUser 向指定用户推送消息，线程安全
func (h *Hub) SendToUser(userId uint64, message []byte) bool {
	h.Mu.RLock()
	client, ok := h.Clients[userId]
	h.Mu.RUnlock()
	if !ok {
		return false
	}
	select {
	case client.Send <- message:
		return true
	default:
		// 如果缓冲区满了，可能连接异常，关闭连接
		close(client.Send)
		h.Unregister <- client
		return false
	}
}
