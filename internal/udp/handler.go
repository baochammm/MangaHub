package udp

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/baochammm/mangahub/utils"
	"github.com/gin-gonic/gin"
)

//	type PingResult struct {
//		ServerAddr string
//		RTT        time.Duration
//		Success    bool
//	}
type NotificationSubscribeResponse struct {
	Message string
	Success bool
}
type Notification struct {
	MangaID   string
	Chapter   int64
	Timestamp time.Time
}
type UDPClientMessage struct {
	ClientUDPAddr string `json:"client_udp_addr"`
	// MangaID       string `json:"manga_id"`
}
type UDPHandler struct {
	repo *UDPRepository
}

func NewUDPHandler(repo *UDPRepository) *UDPHandler {
	return &UDPHandler{repo: repo}
}

func (h *UDPHandler) ProcessUDPAddress(c *gin.Context) {
	userID, err := utils.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	var entry UDPClientMessage
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if entry.ClientUDPAddr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_udp_addr is required"})
		return
	}
	err = SendSuccessSubscription(entry.ClientUDPAddr, 60*time.Second)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = h.repo.CreateNotificationEntry(userID, entry.ClientUDPAddr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Subscription request processed for user " + fmt.Sprint(userID)})
}

func (h *UDPHandler) SubscribeToManga(c *gin.Context) {
	userID, err := utils.GetUserIdFromContext(c)
	mangaID := c.Param("manga")

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	clientAddr, err := h.repo.GetUserUDPAddress(userID)
	if err != nil || clientAddr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please register your UDP address first by running mangahub notify subscribe"})
		return
	}

	//add manga subscription
	isSubscriptionExist, err := h.repo.SubscriptionExists(userID, mangaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if isSubscriptionExist {
		c.JSON(http.StatusConflict, gin.H{"error": "subscription already exists for this manga"})
		return
	}
	err = h.repo.CreateSubscriptionEntry(userID, mangaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Subscribed to manga " + mangaID})
}
func (h *UDPHandler) NotifyNewChapter(mangaID string, chapter int64) (int64, error) {
	subscribers, err := h.repo.GetMangaSubscribers(mangaID)
	if err != nil {
		return 0, err
	}
	successCount := int64(0)

	for _, userID := range subscribers {
		clientAddr, err := h.repo.GetUserUDPAddress(userID)
		if err != nil {
			// log and continue
			fmt.Printf("[UDP] no address for user %d: %v\n", userID, err)
			continue
		}

		if clientAddr == "" {
			continue
		}

		err = SendNewChapterNotification(
			clientAddr,
			mangaID,
			chapter,
			60*time.Second,
		)
		if err != nil {
			fmt.Printf("[UDP] send failed to %s: %v\n", clientAddr, err)
			continue
		}
		successCount++
	}
	return successCount, nil
}

// helper function to send UDP message
func SendSuccessSubscription(clientUDPAddr string, timeout time.Duration) error {
	var message NotificationSubscribeResponse
	serverAddress, err := net.ResolveUDPAddr("udp", clientUDPAddr)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return fmt.Errorf("error resolving address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddress)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return fmt.Errorf("error connecting: %v", err)
	}
	defer conn.Close()
	// Measure time
	start := time.Now()
	message = NotificationSubscribeResponse{Message: "Subscription successful", Success: true}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	// Send to server
	_, err = conn.Write([]byte(data))
	if err != nil {
		fmt.Println("Error sending UDP Message", err)
		return fmt.Errorf("error sending UDP Message: %v", err)
	}
	totalRTT := time.Since(start)
	//Return result
	fmt.Printf("Sent subscription success to %s in %v\n", clientUDPAddr, totalRTT)
	return nil

}
func SendNewChapterNotification(clientUDPAddr string, manga_id string, chapter int64, timeout time.Duration) error {
	var message Notification
	serverAddress, err := net.ResolveUDPAddr("udp", clientUDPAddr)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return fmt.Errorf("error resolving address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddress)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return fmt.Errorf("error connecting: %v", err)
	}
	defer conn.Close()
	// Measure time
	start := time.Now()
	message = Notification{MangaID: manga_id, Chapter: chapter, Timestamp: time.Now()}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	// Send to server
	_, err = conn.Write([]byte(data))
	if err != nil {
		fmt.Println("Error sending UDP Message", err)
		return fmt.Errorf("error sending UDP Message: %v", err)
	}
	totalRTT := time.Since(start)
	//Return result
	fmt.Printf("Sent new chapter notification to %s in %v\n", clientUDPAddr, totalRTT)
	return nil

}
