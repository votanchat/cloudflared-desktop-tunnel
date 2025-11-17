package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// BackendClient handles communication with the backend API
type BackendClient struct {
	baseURL    string
	httpClient *http.Client
	ws         *websocket.Conn
	token      string
	running    bool
	commandsCh chan Command
}

// convertHTTPToWS converts HTTP(S) URL to WS(S)
func convertHTTPToWS(baseURL string) string {
	if len(baseURL) > 4 && baseURL[:4] == "http" {
		return "ws" + baseURL[4:]
	}
	return baseURL
}

// Command represents a command from the backend
type Command struct {
	Type    string                 `json:"type"`    // "update", "restart", "patch", etc.
	Payload map[string]interface{} `json:"payload"` // Command-specific data
}

// TokenResponse represents the response from token endpoint
type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// NewBackendClient creates a new backend client
func NewBackendClient(baseURL string) *BackendClient {
	return &BackendClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		commandsCh: make(chan Command, 10),
	}
}

// Start starts the backend client
func (bc *BackendClient) Start(ctx context.Context) {
	bc.running = true
	log.Println("Backend client started")

	// Connect to WebSocket for real-time commands
	go bc.connectWebSocket(ctx)

	// Periodic token refresh
	go bc.tokenRefreshLoop(ctx)

	// Process commands
	go bc.processCommands(ctx)
}

// Stop stops the backend client
func (bc *BackendClient) Stop() {
	bc.running = false
	if bc.ws != nil {
		bc.ws.Close()
	}
	close(bc.commandsCh)
	log.Println("Backend client stopped")
}

// FetchToken fetches a tunnel token from the backend
func (bc *BackendClient) FetchToken() (string, error) {
	resp, err := bc.httpClient.Get(bc.baseURL + "/api/token")
	if err != nil {
		return "", fmt.Errorf("failed to fetch token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	bc.token = tokenResp.Token
	log.Printf("Token fetched successfully, expires at: %v", tokenResp.ExpiresAt)

	return tokenResp.Token, nil
}

// ReportStatus reports tunnel status to the backend
func (bc *BackendClient) ReportStatus(status map[string]interface{}) error {
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}

	resp, err := bc.httpClient.Post(
		bc.baseURL+"/api/status",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// connectWebSocket connects to the backend WebSocket for real-time commands
func (bc *BackendClient) connectWebSocket(ctx context.Context) {
	for bc.running {
		// Convert http(s) to ws(s)
		wsURL := convertHTTPToWS(bc.baseURL) + "/api/commands"

		log.Printf("Connecting to WebSocket: %s", wsURL)

		ws, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
		if err != nil {
			log.Printf("Failed to connect to WebSocket: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		bc.ws = ws
		log.Println("WebSocket connected")

		// Read messages
		for bc.running {
			var cmd Command
			err := ws.ReadJSON(&cmd)
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				break
			}

			log.Printf("Received command: %s", cmd.Type)
			bc.commandsCh <- cmd
		}

		ws.Close()
		log.Println("WebSocket disconnected, reconnecting in 10 seconds...")
		time.Sleep(10 * time.Second)
	}
}

// tokenRefreshLoop periodically refreshes the token
func (bc *BackendClient) tokenRefreshLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := bc.FetchToken(); err != nil {
				log.Printf("Failed to refresh token: %v", err)
			}
		}
	}
}

// processCommands processes commands received from the backend
func (bc *BackendClient) processCommands(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case cmd, ok := <-bc.commandsCh:
			if !ok {
				return
			}

			switch cmd.Type {
			case "update":
				log.Println("Received update command")
				// TODO: Implement update logic
			case "restart":
				log.Println("Received restart command")
				// TODO: Implement restart logic
			case "patch":
				log.Println("Received patch command")
				// TODO: Implement patch logic
			default:
				log.Printf("Unknown command type: %s", cmd.Type)
			}
		}
	}
}
