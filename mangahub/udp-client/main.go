package udp_listener

import (
	"encoding/json"
	"fmt"
	"net"
)

type NotificationSubscribeResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
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
		var resp NotificationSubscribeResponse
		if err := json.Unmarshal(raw, &resp); err == nil {
			fmt.Printf("[JSON RECEIVED from %s]\nMessage: %s | Success: %v\n",
				serverAddr, resp.Message, resp.Success,
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
		if err := StartUDPListener(8082, username); err != nil {
			fmt.Println("UDP listener error:", err)
		}
	}()
	return nil
}
