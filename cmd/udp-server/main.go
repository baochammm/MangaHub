package udpserver

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/baochammm/mangahub/internal/udp"
)

type NotificationSubscribeResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func StartUDPListener(port int, h *udp.UDPHandler) error {

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

	fmt.Printf("UDP Listener running on port %d...\n", port)

	buffer := make([]byte, 2048)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Read error:", err)
			continue // do NOT exit — keep listening
		}

		raw := buffer[:n]

		// Try parsing JSON
		var req udp.UDPClientRequest
		fmt.Printf("[RAW RECEIVED from %s] %s\n", clientAddr, string(raw))
		if err := json.Unmarshal(raw, &req); err == nil {
			fmt.Printf("[JSON RECEIVED from %s]\n Action: %v\n",
				clientAddr, req.Action,
			)

			resp := h.ProcessUDPRequest(req.Action, req.Token, clientAddr.String(), req.Payload)
			respBytes, err := json.Marshal(resp)
			if err != nil {
				fmt.Println("Error marshaling response:", err)
				continue
			}
			conn.WriteToUDP(respBytes, clientAddr)

		} else {
			fmt.Printf("[RAW RECEIVED from %s] %s\n", clientAddr, string(raw))
			fmt.Println("Error unmarshaling UDP request:", err)
		}

	}

}

func StartUDPServer(h *udp.UDPHandler) error {
	go func() {
		if err := StartUDPListener(9091, h); err != nil {
			fmt.Println("UDP listener error:", err)
		}
	}()
	return nil
}
