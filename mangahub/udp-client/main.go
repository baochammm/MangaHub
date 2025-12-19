package udp_client

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Notification struct {
	MangaID   string
	Chapter   int64
	Timestamp time.Time
}

func StartUDPListener(port int, serverID string) error {

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error listening:", err)
		return err
	}
	defer conn.Close()

	fmt.Printf("UDP Listener (%s) running on port %d...\n", serverID, port)

	buffer := make([]byte, 2048)

	for {
		n, serverAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Read error:", err)
			continue // do NOT exit — keep listening
		}

		raw := buffer[:n]

		// Try parsing JSON
		var resp Notification
		if err := json.Unmarshal(raw, &resp); err == nil {
			fmt.Printf("[JSON RECEIVED from %s]\nManga: %s | Chapter: %d | Timestamp: %s\n",
				serverAddr, resp.MangaID, resp.Chapter, resp.Timestamp,
			)
		} else {
			fmt.Printf("[RAW RECEIVED from %s] %s\n", serverAddr, string(raw))
		}

		// Optional: reply PONG if you ever need RTT
		// conn.WriteToUDP([]byte("PONG"), serverAddr)
	}

}

func StartUDPServer(username string) error {
	go func() {
		if err := StartUDPListener(3002, username); err != nil {
			fmt.Println("UDP listener error:", err)
		}
	}()
	return nil
}
