package sv_grpc

import (
	"fmt"
)

func BroadcastProgressUpdate(
	userID int64,
	mangaID string,
	chapter int64,
) {
	// Example payload
	message := fmt.Sprintf(
		"USER %d PROGRESS %s CHAPTER %d",
		userID,
		mangaID,
		chapter,
	)

	// TODO:
	// - send to TCP hub
	// - or SignalR
	// - or websocket gateway

	fmt.Println("[TCP BROADCAST]", message)
}
