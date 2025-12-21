package websocket_client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type ChatMessage struct {
	Room      string `json:"room"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

func ReadMessages(ctx context.Context, conn *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Connection closed.")
				return
			}

			var m ChatMessage
			if err := json.Unmarshal(msg, &m); err != nil {
				fmt.Println(string(msg))
				continue
			}

			t := time.Unix(m.Timestamp, 0).Format("15:04")
			fmt.Printf("[%s] %s: %s\n", t, m.Username, m.Message)
		}
	}
}
func WriteMessages(ctx context.Context, conn *websocket.Conn) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if !scanner.Scan() {
				return
			}
			text := scanner.Text()
			if text == "" {
				continue
			}
			conn.WriteMessage(websocket.TextMessage, []byte(text))
		}
	}
}
