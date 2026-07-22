// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"IMM/gateway/internal/hub"
	"IMM/gateway/internal/svc"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

const (
	// 发送通道缓冲大小
	sendBufSize = 256
	// 写超时
	writeWait = 10 * time.Second
	// 心跳间隔（服务端主动ping）
	pingPeriod = 30 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		/* 限定来自指定域名的连接可以变为长连接
				origin := r.Header.Get("Origin")
		        allowedOrigins := map[string]bool{
		            "https://your-frontend.com": true,
		        }
		        return allowedOrigins[origin]
		*/
		return true
	},
}

func WsGatewayHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取token，Head里或是query里携带
		var tokenStr string
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			tokenStr = r.URL.Query().Get("token")
		}

		if tokenStr == "" {
			httpx.Error(w, errors.New("Miss Access Token."))
			return
		}

		// 验证JWT Token并提取userId
		userId, err := validateToken(tokenStr, svcCtx.Config.JwtAuth.AccessSecret)
		if err != nil {
			httpx.Error(w, errors.New("Invalid Access Token."))
			return
		}

		// 3. 升级HTTP连接为WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		// 4. 创建客户端并注册到Hub
		client := &hub.Client{
			UserId: userId,
			Conn:   conn,
			Send:   make(chan []byte, sendBufSize),
		}
		svcCtx.ClientHub.Register <- client

		// 5. 启动该Client对应的读写协程
		go writePump(client)
		go readPump(svcCtx, client)
	}
}

// validateToken 验证JWT并返回userId
func validateToken(tokenStr, secret string) (uint64, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}
	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid claims type")
	}

	// 从claims中提取uid
	userIdVal, ok := claims["uid"]
	if !ok {
		return 0, fmt.Errorf("uid not found in token")
	}
	// JWT中的数字默认为float64
	userId, ok := userIdVal.(float64)
	if !ok {
		return 0, fmt.Errorf("uid is not a number")
	}
	return uint64(userId), nil
}

// writePump 从 client.Send 通道读取消息并写入WebSocket
func writePump(client *hub.Client) {
	// 设置Ping处理器（服务端主动发送Ping）
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			// 若通道已关闭，则发送关闭帧并退出
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// 设置写超时
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			// 写入消息（文本消息）
			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logx.Errorf("write message error: %v", err)
				return
			}
		case <-ticker.C:
			// 定时发送Ping保持连接
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logx.Errorf("ping error: %v", err)
				return
			}
		}
	}
}

// readPump 从WebSocket读取消息（如心跳、上行消息等）
func readPump(svcCtx *svc.ServiceContext, client *hub.Client) {
	defer func() {
		// 客户端断开时注销
		svcCtx.ClientHub.Unregister <- client
		client.Conn.Close()
	}()

	// 设置读超时（若长时间无消息则断开）
	client.Conn.SetReadDeadline(time.Now().Add(180 * time.Second))
	// 设置Pong处理器，收到Pong时刷新读超时，心跳保活处理
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		return nil
	})

	for {
		// 读取消息（文本或二进制，这里假设文本）
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Errorf("read error: %v", err)
			}
			break
		}

		// 处理消息（异步或同步，根据业务需求）
		go svc.HandleMessage(client, message, svcCtx)
	}
}
