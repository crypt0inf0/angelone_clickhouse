package ws

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	HeartbeatInterval = 10 * time.Second
	ReconnectDelay    = 5 * time.Second
)

type WebSocketClient struct {
	conn           *websocket.Conn
	url            string
	OnMessage      func([]byte)
	isConnected    bool
	reconnectDelay time.Duration
	Headers        map[string]string
}

func NewWebSocketClient(url string, headers map[string]string) *WebSocketClient {
	return &WebSocketClient{
		url:            url,
		Headers:        headers,
		reconnectDelay: 5 * time.Second,
	}
}

func (c *WebSocketClient) Connect() error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	// Add headers to connection request
	conn, _, err := dialer.Dial(c.url, c.getHttpHeaders())
	if err != nil {
		return err
	}

	c.conn = conn
	c.isConnected = true

	// Start heartbeat
	go c.heartbeat()

	return nil
}

func (c *WebSocketClient) getHttpHeaders() http.Header {
	headers := http.Header{}
	for key, value := range c.Headers {
		headers.Set(key, value)
	}
	return headers
}

func (c *WebSocketClient) heartbeat() {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		<-ticker.C
		if err := c.conn.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
			log.Printf("Failed to send heartbeat: %v", err)
			c.reconnect()
			return
		}
	}
}

func (c *WebSocketClient) reconnect() {
	c.isConnected = false
	c.conn.Close()
	time.Sleep(c.reconnectDelay)

	for !c.isConnected {
		if err := c.Connect(); err != nil {
			log.Printf("Reconnection failed: %v", err)
			time.Sleep(c.reconnectDelay)
			continue
		}
	}
}

func (c *WebSocketClient) Listen() {
	for {
		if !c.isConnected {
			if err := c.Connect(); err != nil {
				log.Printf("Connection failed: %v, retrying in %v", err, c.reconnectDelay)
				time.Sleep(c.reconnectDelay)
				continue
			}
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			c.isConnected = false
			c.conn.Close()
			continue
		}

		if c.OnMessage != nil {
			c.OnMessage(message)
		}
	}
}

func (c *WebSocketClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// Add SendJSON method
func (c *WebSocketClient) SendJSON(v interface{}) error {
	if !c.isConnected {
		return fmt.Errorf("websocket is not connected")
	}
	return c.conn.WriteJSON(v)
}
