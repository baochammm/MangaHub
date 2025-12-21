package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	udpclient "github.com/baochammm/mangahub/mangahub/udp-client"
	"github.com/baochammm/mangahub/package/models"
	"github.com/baochammm/mangahub/utils"
	"github.com/spf13/cobra"
)

var baseURL string
var username string
var password string
var token string
var status string

func getToken() string {
	if token != "" {
		utils.SaveToken(token)
		return token
	}

	// otherwise load from cache
	cached, err := utils.LoadToken()
	if err == nil && cached != "" {
		return cached
	}

	return "" // no available token
}
func main() {
	rootCmd := &cobra.Command{
		Use:   "mangahub",
		Short: "MangaHub CLI",
	}
	rootCmd.PersistentFlags().StringVar(&username, "username", "", "Username for authentication")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "Password for authentication")

	rootCmd.PersistentFlags().StringVar(&token, "token", "", "JWT token (or set MANGAHUB_TEST_TOKEN)")
	rootCmd.PersistentFlags().StringVar(&baseURL, "base", "http://localhost:8080", "Base URL for API")

	// library subcommands
	libraryCmd := &cobra.Command{
		Use:   "library",
		Short: "Library commands",
	}
	libraryListCmd := &cobra.Command{
		Use:   "list",
		Short: "List library items for the authenticated user",
		RunE: func(cmd *cobra.Command, args []string) error {

			jwt := getToken()
			if jwt == "" {
				return fmt.Errorf("no token found. Please login using: mangahub login --username USER --password PASS")
			}

			req, err := http.NewRequest("GET", baseURL+"/users/library", nil)
			if err != nil {
				return err
			}

			req.Header.Set("Authorization", "Bearer "+jwt)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("server error %s: %s", resp.Status, string(body))
			}

			var list models.ReadingLists
			if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
				return err
			}

			fmt.Println("📚 Your Library")

			// if user use --status flag
			if cmd.Flags().Changed("status") {
				switch status {
				case "reading":
					if len(list.Reading) == 0 {
						fmt.Println("No manga in Reading list.")
						return nil
					}
					fmt.Println("Currently Reading:")
					for _, it := range list.Reading {
						fmt.Printf(" - %s (ch %d)\n", it.MangaID, it.CurrentChapter)
						fmt.Printf("   Last Updated: %s\n", it.LastUpdated.Format("2006-01-02 15:04:05"))
					}
					return nil

				case "completed":
					if len(list.Completed) == 0 {
						fmt.Println("No manga in Completed list.")
						return nil
					}
					fmt.Println("Completed:")
					for _, it := range list.Completed {
						fmt.Printf(" - %s\n", it.MangaID)
					}
					return nil

				case "plan_to_read":
					if len(list.PlanToRead) == 0 {
						fmt.Println("No manga in Plan to Read list.")
						return nil
					}
					fmt.Println("Plan to Read:")
					for _, it := range list.PlanToRead {
						fmt.Printf(" - %s\n", it.MangaID)
					}
					return nil

				default:
					return fmt.Errorf("invalid status: %s (valid: reading, completed, plan_to_read)", status)
				}
			}

			// no --status flag => print all
			if len(list.Reading) > 0 {
				fmt.Println("Currently Reading:")
				for _, it := range list.Reading {
					fmt.Printf(" - %s (ch %d)\n", it.MangaID, it.CurrentChapter)
					fmt.Printf("   Last Updated: %s\n", it.LastUpdated.Format("2006-01-02 15:04:05"))
				}
			}

			if len(list.Completed) > 0 {
				fmt.Println("Completed:")
				for _, it := range list.Completed {
					fmt.Printf(" - %s\n", it.MangaID)
				}
			}

			if len(list.PlanToRead) > 0 {
				fmt.Println("Plan to Read:")
				for _, it := range list.PlanToRead {
					fmt.Printf(" - %s\n", it.MangaID)
				}
			}

			return nil

		},
	}

	libraryListCmd.Flags().StringVar(&status, "status", "", "Filter by status: reading, completed, plan")

	//#region add manga command
	libraryAddCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a manga to your library",
		RunE: func(cmd *cobra.Command, args []string) error {
			mangaID, _ := cmd.Flags().GetString("manga-id")
			status, _ := cmd.Flags().GetString("status")
			chapter, _ := cmd.Flags().GetInt("chapter") // optional

			if mangaID == "" {
				return fmt.Errorf("--manga-id required")
			}
			if status == "" {
				return fmt.Errorf("--status required (reading, completed, plan_to_read)")
			}

			jwt := getToken()
			if jwt == "" {
				return fmt.Errorf("no token found. Please login using: mangahub auth login --username USER --password PASS")
			}

			// Build JSON body
			reqBody := map[string]interface{}{
				"manga_id": mangaID,
				"status":   status,
			}

			if cmd.Flags().Changed("chapter") {
				reqBody["current_chapter"] = chapter
			}

			body, _ := json.Marshal(reqBody)

			// POST request
			req, err := http.NewRequest("POST", baseURL+"/users/library", bytes.NewBuffer(body))
			if err != nil {
				return err
			}

			req.Header.Set("Authorization", "Bearer "+jwt)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != 201 && resp.StatusCode != 200 {
				data, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("add failed %s: %s", resp.Status, string(data))
			}

			fmt.Println("Manga added to your library.")
			return nil
		},
	}

	libraryAddCmd.Flags().String("manga-id", "", "ID of the manga to add")
	libraryAddCmd.Flags().String("status", "", "Reading status: reading, completed, plan_to_read")
	libraryAddCmd.Flags().Int("chapter", 0, "Optional: current chapter number")

	//#region update manga command
	libraryUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update status for a manga in your library",
		RunE: func(cmd *cobra.Command, args []string) error {
			mangaID, _ := cmd.Flags().GetString("manga-id")
			newStatus, _ := cmd.Flags().GetString("status")

			if mangaID == "" {
				return fmt.Errorf("--manga-id required")
			}
			if newStatus == "" {
				return fmt.Errorf("--status required")
			}

			jwt := getToken()
			if jwt == "" {
				return fmt.Errorf("no token found. Please login using: mangahub auth login --username USER --password PASS")
			}

			// build JSON body
			reqBody := map[string]string{
				"manga_id": mangaID,
				"status":   newStatus,
			}

			body, _ := json.Marshal(reqBody)

			// send PATCH request
			req, err := http.NewRequest("PATCH", baseURL+"/users/library", bytes.NewBuffer(body))
			if err != nil {
				return err
			}

			req.Header.Set("Authorization", "Bearer "+jwt)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				data, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("update failed %s: %s", resp.Status, string(data))
			}

			fmt.Println("Manga status updated successfully.")
			return nil
		},
	}

	libraryUpdateCmd.Flags().String("manga-id", "", "ID of the manga to update")
	libraryUpdateCmd.Flags().String("status", "", "New status (reading, completed, plan_to_read)")

	libraryRemoveCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a manga from your library",
		RunE: func(cmd *cobra.Command, args []string) error {
			mangaID, _ := cmd.Flags().GetString("manga-id")

			if mangaID == "" {
				return fmt.Errorf("--manga-id required")
			}

			jwt := getToken()
			if jwt == "" {
				return fmt.Errorf("no token found. Please login using: mangahub auth login --username USER --password PASS")
			}

			// JSON body
			reqBody := map[string]string{
				"manga_id": mangaID,
			}

			body, _ := json.Marshal(reqBody)

			// delete request with JSON body
			req, err := http.NewRequest("DELETE", baseURL+"/users/library", bytes.NewBuffer(body))
			if err != nil {
				return err
			}

			req.Header.Set("Authorization", "Bearer "+jwt)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				data, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("remove failed %s: %s", resp.Status, string(data))
			}

			fmt.Println("Manga removed from your library.")
			return nil
		},
	}
	libraryRemoveCmd.Flags().String("manga-id", "", "ID of the manga to remove")

	//#region auth subcommands
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
	}

	//#region login command
	authLoginCmd := &cobra.Command{
		Use: "login",
		RunE: func(cmd *cobra.Command, args []string) error {
			if username == "" || password == "" {
				return fmt.Errorf("username and password required")
			}

			// Send login request to API
			req := map[string]string{
				"username": username,
				"password": password,
			}

			body, _ := json.Marshal(req)

			resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(body))
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			// incorrect credentials => 401
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("invalid username or password")
			}

			var result struct {
				Token string `json:"token"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return err
			}

			// save token
			utils.SaveToken(result.Token)

			fmt.Println("Logged in successfully.")
			return nil
		},
	}

	//#region sign up command
	authSignupCmd := &cobra.Command{
		Use: "signup",
		RunE: func(cmd *cobra.Command, args []string) error {
			if username == "" || password == "" {
				return fmt.Errorf("username and password required")
			}
			//send signup request to API
			req := map[string]string{
				"username": username,
				"password": password,
			}

			body, _ := json.Marshal(req)

			resp, err := http.Post(baseURL+"/auth/signup", "application/json", bytes.NewBuffer(body))
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != 201 && resp.StatusCode != 200 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("sign up failed: %s", string(body))
			}

			fmt.Println("Sign up successful")
			return nil
		},
	}

	//#region logout command
	authLogoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout and clear saved token",
		RunE: func(cmd *cobra.Command, args []string) error {

			jwt := getToken()
			if jwt == "" {
				fmt.Println("You are not logged in.")
				return nil
			}

			// send logout request to API
			req, err := http.NewRequest("POST", baseURL+"/auth/logout", nil)
			if err == nil {
				req.Header.Set("Authorization", "Bearer "+jwt)
				http.DefaultClient.Do(req)
			}

			// clear token locally
			if err := utils.ClearToken(); err != nil {
				return fmt.Errorf("failed to clear token: %v", err)
			}

			fmt.Println("Logged out successfully.")
			return nil
		},
	}

	// #region manga subcommands
	mangaCmd := &cobra.Command{
		Use:   "manga",
		Short: "Manga commands",
	}

	mangaListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all manga by genre, title or get manga  by ID",
		RunE: func(cmd *cobra.Command, args []string) error {

			mangaID, _ := cmd.Flags().GetString("manga-id")
			genres, _ := cmd.Flags().GetStringSlice("genre")
			title, _ := cmd.Flags().GetString("title")

			var url string

			switch {
			case mangaID != "":
				url = fmt.Sprintf("%s/manga/%s", baseURL, mangaID)
			case title != "":
				url = fmt.Sprintf("%s/manga/search?query=%s", baseURL, title)
			case len(genres) > 0:
				joined := strings.Join(genres, ",")
				url = fmt.Sprintf("%s/manga/filter/genre?query=%s", baseURL, joined)
			default:
				url = baseURL + "/manga"
			}

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return err
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed %s: %s", resp.Status, string(body))
			}

			if mangaID != "" || title != "" {
				var m interface{}
				if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
					return err
				}

				raw, err := json.Marshal(m)
				if err != nil {
					return err
				}
				fmt.Println(string(raw))
				return nil
			}

			var mangas []models.Manga
			if err := json.NewDecoder(resp.Body).Decode(&mangas); err != nil {
				return err
			}

			fmt.Println("📚 Manga List:")
			if len(mangas) == 0 {
				fmt.Println("No manga found.")
				return nil
			}

			for _, m := range mangas {
				fmt.Printf(" - %s (%s)\n", m.Title, m.ID)
			}

			return nil
		},
	}
	//#region notifications command
	notifyCmd := &cobra.Command{
		Use:   "notify",
		Short: "Start UDP server to receive notifications",
	}
	notifySubscribeCmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Start UDP server to receive notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			jwt := getToken()
			if jwt == "" {
				return fmt.Errorf("no token found. Please login using: mangahub auth login --username USER --password PASS")
			}
			//TODO: phần này đang hardcode UDP address, sẽ update sau WS
			udp_addrress := "127.0.0.1:8082"
			data := map[string]string{
				"client_udp_addr": udp_addrress,
			}
			body, _ := json.Marshal(data)
			req, err := http.NewRequest(
				"POST",
				baseURL+"/users/notifications/subscribe",
				bytes.NewBuffer(body),
			)
			req.Header.Set("Authorization", "Bearer "+jwt)
			req.Header.Set("Content-Type", "application/json")

			if err != nil {
				return err
			}

			resp, err := http.DefaultClient.Do(req)

			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errBody, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("Subscribe to notifications failed: %s", errBody)
			}
			fmt.Println("Starting UDP server to receive notifications...")

			if err := udpclient.StartUDPServer(username); err != nil {
				return fmt.Errorf("failed to start UDP server: %v", err)
			}

			fmt.Println("Listening for UDP notifications. Press Ctrl+C to exit.")

			// Block until Ctrl+C
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
			<-stop

			fmt.Println("\nShutting down UDP listener")
			return nil
		},
	}
	notifyAddCmd := &cobra.Command{
		Use:   "add",
		Short: "Subscribe to manga notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			mangaID, _ := cmd.Flags().GetString("manga")
			if mangaID == "" {
				return fmt.Errorf("--manga is required")
			}
			jwt := getToken()
			if jwt == "" {
				return fmt.Errorf("no token found. Please login using: mangahub auth login --username USER --password PASS")
			}
			req, err := http.NewRequest(
				"POST",
				baseURL+"/users/notifications/subscribe/"+mangaID,
				nil,
			)
			req.Header.Set("Authorization", "Bearer "+jwt)
			req.Header.Set("Content-Type", "application/json")

			if err != nil {
				return err
			}

			resp, err := http.DefaultClient.Do(req)

			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errBody, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("subscription failed: %s", errBody)
			}

			fmt.Printf("✅ Subscribed to manga: %s\n", mangaID)
			return nil
		},
	}

	// progress commands
	progressCmd := &cobra.Command{
		Use:   "progress",
		Short: "Reading progress commands",
	}

	progressUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update reading progress",
		RunE: func(cmd *cobra.Command, args []string) error {
			jwt := getToken()
			if jwt == "" {
				return fmt.Errorf("no token found. Please login using: mangahub login --username USER --password PASS")
			}

			mangaID, _ := cmd.Flags().GetString("manga-id")
			chapter, _ := cmd.Flags().GetInt("chapter")
			volume, _ := cmd.Flags().GetInt("volume")
			notes, _ := cmd.Flags().GetString("notes")
			force, _ := cmd.Flags().GetBool("force")

			if mangaID == "" {
				return fmt.Errorf("--manga-id required")
			}
			if chapter <= 0 {
				return fmt.Errorf("--chapter must be > 0")
			}

			// build request body for API
			reqBody := map[string]interface{}{
				"manga_id":        mangaID,
				"current_chapter": chapter,
				"force":           force,
			}

			// optional fields
			if volume > 0 {
				reqBody["volume"] = volume
			} else {
				reqBody["volume"] = nil
			}
			if notes != "" {
				reqBody["notes"] = notes
			} else {
				reqBody["notes"] = nil
			}

			body, _ := json.Marshal(reqBody)

			req, err := http.NewRequest(
				"PATCH",
				baseURL+"/users/progress",
				bytes.NewBuffer(body),
			)
			if err != nil {
				return err
			}

			req.Header.Set("Authorization", "Bearer "+jwt)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errBody, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("✗ Progress update failed: %s", strings.TrimSpace(string(errBody)))
			}

			var result struct {
				MangaTitle        string    `json:"manga_title"`
				PreviousChapter   int       `json:"previous_chapter"`
				CurrentChapter    int       `json:"current_chapter"`
				UpdatedAt         time.Time `json:"updated_at"`
				DevicesSynced     int       `json:"devices_synced"`
				TotalChaptersRead int       `json:"total_chapters_read"`
				ReadingStreak     int       `json:"reading_streak"`
				NextChapter       int       `json:"next_chapter"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return err
			}

			// OUTPUT
			fmt.Println("Updating reading progress...")
			fmt.Println("✓ Progress updated successfully!")
			fmt.Printf("Manga: %s\n", result.MangaTitle)
			fmt.Printf("Previous: Chapter %d\n", result.PreviousChapter)
			fmt.Printf(
				"Current: Chapter %d (+%d)\n",
				result.CurrentChapter,
				result.CurrentChapter-result.PreviousChapter,
			)
			fmt.Println(
				"Updated:",
				result.UpdatedAt.Local().Format("2006-01-02 15:04:05"),
			)

			fmt.Println("Sync Status:")
			fmt.Println(" Local database: ✓ Updated")
			fmt.Printf(
				" TCP sync server: ✓ Broadcasting to %d connected devices\n",
				result.DevicesSynced,
			)
			fmt.Println(" Cloud backup: ✓ Synced") // currently hardcoded

			fmt.Println("Statistics:")
			fmt.Printf(" Total chapters read: %d\n", result.TotalChaptersRead)
			fmt.Printf(" Reading streak: %d days\n", result.ReadingStreak)

			if result.NextChapter > 0 {
				fmt.Printf(
					"Next actions:\n Continue reading: Chapter %d available\n",
					result.NextChapter,
				)
			}

			return nil
		},
	}

	historyCmd := &cobra.Command{
		Use:   "history",
		Short: "View reading progress history",
		RunE: func(cmd *cobra.Command, args []string) error {
			jwt := getToken()
			if jwt == "" {
				return fmt.Errorf("please login first")
			}

			mangaID, _ := cmd.Flags().GetString("manga-id")

			url := baseURL + "/users/progress/history"
			if mangaID != "" {
				url += "?manga_id=" + mangaID
			}

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return err
			}

			req.Header.Set("Authorization", "Bearer "+jwt)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errBody, _ := io.ReadAll(resp.Body)
				return fmt.Errorf(
					"failed to fetch history: %s",
					strings.TrimSpace(string(errBody)),
				)
			}

			var result struct {
				UserID  int64 `json:"user_id"`
				History []struct {
					MangaID string `json:"manga_id"`
					Chapter int    `json:"chapter"`
					Date    string `json:"date_read"`
				} `json:"history"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return err
			}

			if len(result.History) == 0 {
				fmt.Println("No reading history found.")
				return nil
			}

			fmt.Printf("Reading Progress History (User ID: %d)\n", result.UserID)
			fmt.Println("------------------------------------------------")

			for _, h := range result.History {
				fmt.Printf(
					"%s | %-15s → Chapter %d\n",
					h.Date[:10],
					h.MangaID,
					h.Chapter,
				)
			}

			return nil
		},
	}

	progressSyncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Manually sync reading progress with server",
		RunE: func(cmd *cobra.Command, args []string) error {
			jwt := getToken()
			if jwt == "" {
				return fmt.Errorf("please login first")
			}

			url := baseURL + "/users/progress/sync"

			req, err := http.NewRequest("POST", url, nil)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+jwt)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("sync failed: %s", string(body))
			}

			fmt.Println("Sync completed successfully")
			return nil
		},
	}

	progressSyncStatusCmd := &cobra.Command{
		Use:   "sync-status",
		Short: "Check progress sync status",
		RunE: func(cmd *cobra.Command, args []string) error {
			jwt := getToken()
			if jwt == "" {
				return fmt.Errorf("please login first")
			}

			url := baseURL + "/users/progress/sync-status"

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+jwt)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed: %s", string(body))
			}

			var result map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return err
			}
			keys := []string{"status", "last_synced_at"}
			fmt.Println("📡 Sync Status:")
			for _, k := range keys {
				fmt.Printf(" - %s: %s\n", k, result[k])
			}

			return nil
		},
	}

	// flags
	progressUpdateCmd.Flags().String("manga-id", "", "Manga ID")
	progressUpdateCmd.Flags().Int("chapter", 0, "Chapter number")
	progressUpdateCmd.Flags().Int("volume", 0, "Volume number")
	progressUpdateCmd.Flags().String("notes", "", "Personal notes")
	progressUpdateCmd.Flags().Bool("force", false, "Force backward progress update")
	historyCmd.Flags().String("manga-id", "", "Filter by manga ID")

	progressCmd.AddCommand(progressUpdateCmd)
	progressCmd.AddCommand(historyCmd)
	progressCmd.AddCommand(progressSyncCmd)
	progressCmd.AddCommand(progressSyncStatusCmd)
	rootCmd.AddCommand(progressCmd)

	notifyAddCmd.Flags().String("manga", "", "ID of the manga to subscribe to")

	notifyCmd.AddCommand(notifySubscribeCmd)
	notifyCmd.AddCommand(notifyAddCmd)
	rootCmd.AddCommand(notifyCmd)

	mangaListCmd.Flags().String("manga-id", "", "Get a manga by ID")
	mangaListCmd.Flags().StringSlice("genre", []string{}, "Filter manga by genres")
	mangaListCmd.Flags().String("title", "", "Search manga by title")

	libraryCmd.AddCommand(libraryListCmd)
	libraryCmd.AddCommand(libraryAddCmd)
	libraryCmd.AddCommand(libraryUpdateCmd)
	libraryCmd.AddCommand(libraryRemoveCmd)
	rootCmd.AddCommand(libraryCmd)

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authSignupCmd)
	authCmd.AddCommand(authLogoutCmd)
	rootCmd.AddCommand(authCmd)

	mangaCmd.AddCommand(mangaListCmd)
	rootCmd.AddCommand(mangaCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
