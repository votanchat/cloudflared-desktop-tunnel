package services

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

// BackendService handles communication with the backend API
type BackendService struct {
	baseURL    string
	httpClient *http.Client
	ws         *websocket.Conn
	token      string
	running    bool
	commandsCh chan Command
	ctx        context.Context
	cancel     context.CancelFunc
}

// Command represents a command from the backend
type Command struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// TokenResponse represents the response from token endpoint
type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// NewBackendService creates a new backend service
func NewBackendService(baseURL string) *BackendService {
	ctx, cancel := context.WithCancel(context.Background())
	return &BackendService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		commandsCh: make(chan Command, 10),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start starts the backend client
func (s *BackendService) Start() {
	s.running = true
	go s.connectWebSocket()
	go s.tokenRefreshLoop()
	go s.processCommands()
}

// Stop stops the backend client
func (s *BackendService) Stop() {
	s.running = false
	s.cancel()
	if s.ws != nil {
		s.ws.Close()
	}
	close(s.commandsCh)
}

// FetchToken fetches a tunnel token from the backend
func (s *BackendService) FetchToken() (string, error) {
	resp, err := s.httpClient.Get(s.baseURL + "/api/token")
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

	s.token = tokenResp.Token
	return tokenResp.Token, nil
}

// ReportStatus reports tunnel status to the backend
func (s *BackendService) ReportStatus(status map[string]interface{}) error {
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Post(
		s.baseURL+"/api/status",
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

// convertHTTPToWS converts HTTP(S) URL to WS(S)
func convertHTTPToWS(baseURL string) string {
	if len(baseURL) > 4 && baseURL[:4] == "http" {
		return "ws" + baseURL[4:]
	}
	return baseURL
}

// connectWebSocket connects to the backend WebSocket for real-time commands
func (s *BackendService) connectWebSocket() {
	const reconnectDelay = 10 * time.Second

	for s.running {
		wsURL := convertHTTPToWS(s.baseURL) + "/api/commands"
		ws, _, err := websocket.DefaultDialer.DialContext(s.ctx, wsURL, nil)
		if err != nil {
			time.Sleep(reconnectDelay)
			continue
		}

		s.ws = ws

		for s.running {
			var cmd Command
			if err := ws.ReadJSON(&cmd); err != nil {
				break
			}
			s.commandsCh <- cmd
		}

		ws.Close()
		time.Sleep(reconnectDelay)
	}
}

// tokenRefreshLoop periodically refreshes the token
func (s *BackendService) tokenRefreshLoop() {
	const refreshInterval = 5 * time.Minute
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if _, err := s.FetchToken(); err != nil {
				// Log error but continue
			}
		}
	}
}

// processCommands processes commands received from the backend
func (s *BackendService) processCommands() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case cmd, ok := <-s.commandsCh:
			if !ok {
				return
			}
			s.handleCommand(cmd)
		}
	}
}

// handleCommand handles a single command
func (s *BackendService) handleCommand(cmd Command) {
	switch cmd.Type {
	case "update", "restart", "patch":
		// Commands not implemented yet
	default:
		// Unknown command
	}
}

