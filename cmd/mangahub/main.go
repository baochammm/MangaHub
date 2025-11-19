package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/baochammm/mangahub/internal/auth"
	"github.com/baochammm/mangahub/package/models"
	"github.com/spf13/cobra"
)

var baseURL string
var username string
var password string
var token string

func getToken() string {
	if token != "" {
		auth.SaveToken(token)
		return token
	}

	// otherwise load from cache
	cached, err := auth.LoadToken()
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

			if len(list.Reading) > 0 {
				fmt.Println("\nCurrently Reading:")
				for _, it := range list.Reading {
					fmt.Printf(" - %s (ch %d)", it.MangaID, it.CurrentChapter)
					fmt.Printf(" - Last Updated : %s\n", it.LastUpdated.Format("2006-01-02 15:04:05"))

				}
			}

			if len(list.Completed) > 0 {
				fmt.Println("\nCompleted:")
				for _, it := range list.Completed {
					fmt.Printf(" - %s\n", it.MangaID)
				}
			}

			if len(list.PlanToRead) > 0 {
				fmt.Println("\nPlan to Read:")
				for _, it := range list.PlanToRead {
					fmt.Printf(" - %s\n", it.MangaID)
				}
			}

			return nil
		},
	}

	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
	}
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

			var result struct {
				Token string `json:"token"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return err
			}

			// save token
			auth.SaveToken(result.Token)

			fmt.Println("Logged in successfully.")
			return nil
		},
	}

	libraryCmd.AddCommand(libraryListCmd)
	rootCmd.AddCommand(libraryCmd)
	authCmd.AddCommand(authLoginCmd)
	rootCmd.AddCommand(authCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
