package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	backendLogger.Info("Backend client started")

	go bc.connectWebSocket(ctx)
	go bc.tokenRefreshLoop(ctx)
	go bc.processCommands(ctx)
}

// Stop stops the backend client
func (bc *BackendClient) Stop() {
	bc.running = false
	if bc.ws != nil {
		bc.ws.Close()
	}
	close(bc.commandsCh)
	backendLogger.Info("Backend client stopped")
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
	backendLogger.Info("Token fetched successfully, expires at: %v", tokenResp.ExpiresAt)

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
	const reconnectDelay = 10 * time.Second

	for bc.running {
		wsURL := convertHTTPToWS(bc.baseURL) + "/api/commands"
		backendLogger.Info("Connecting to WebSocket: %s", wsURL)

		ws, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
		if err != nil {
			backendLogger.Warn("Failed to connect to WebSocket: %v, retrying in %v", err, reconnectDelay)
			time.Sleep(reconnectDelay)
			continue
		}

		bc.ws = ws
		backendLogger.Info("WebSocket connected")

		for bc.running {
			var cmd Command
			if err := ws.ReadJSON(&cmd); err != nil {
				backendLogger.Error("WebSocket read error: %v", err)
				break
			}

			backendLogger.Debug("Received command: %s", cmd.Type)
			bc.commandsCh <- cmd
		}

		ws.Close()
		backendLogger.Warn("WebSocket disconnected, reconnecting in %v...", reconnectDelay)
		time.Sleep(reconnectDelay)
	}
}

// tokenRefreshLoop periodically refreshes the token
func (bc *BackendClient) tokenRefreshLoop(ctx context.Context) {
	const refreshInterval = 5 * time.Minute
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := bc.FetchToken(); err != nil {
				backendLogger.Error("Failed to refresh token: %v", err)
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
			bc.handleCommand(cmd)
		}
	}
}

// handleCommand handles a single command
func (bc *BackendClient) handleCommand(cmd Command) {
	switch cmd.Type {
	case "update":
		backendLogger.Info("Received update command (not implemented)")
	case "restart":
		backendLogger.Info("Received restart command (not implemented)")
	case "patch":
		backendLogger.Info("Received patch command (not implemented)")
	default:
		backendLogger.Warn("Unknown command type: %s", cmd.Type)
	}
}
