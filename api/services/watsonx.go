package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"api/config"
	"api/models"
)

const iamTokenURL = "https://iam.cloud.ibm.com/identity/token"

type WatsonXService struct {
	baseURL       string
	apiKey        string
	assistantID   string
	environmentID string
	version       string
	httpClient    *http.Client

	token    string
	tokenExp time.Time
	tokenMu  sync.Mutex

	// Redis for session storage
	redis *RedisService
}

func NewWatsonXService(cfg *config.Config, redis *RedisService) *WatsonXService {
	return &WatsonXService{
		baseURL:       cfg.WatsonXBaseURL,
		apiKey:        cfg.WatsonXAPIKey,
		assistantID:   cfg.WatsonXAssistantID,
		environmentID: cfg.WatsonXEnvironmentID,
		version:       cfg.WatsonXVersion,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		redis:         redis,
	}
}

// getToken returns a valid IAM Bearer token, refreshing if expired
func (s *WatsonXService) getToken() (string, error) {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()

	// Reuse token if still valid (with 60s margin)
	if s.token != "" && time.Now().Before(s.tokenExp.Add(-60*time.Second)) {
		return s.token, nil
	}

	log.Println("Refreshing IAM token...")

	data := url.Values{}
	data.Set("grant_type", "urn:ibm:params:oauth:grant-type:apikey")
	data.Set("apikey", s.apiKey)

	resp, err := s.httpClient.Post(iamTokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to request IAM token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("IAM token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp models.IAMTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode IAM token response: %w", err)
	}

	s.token = tokenResp.AccessToken
	s.tokenExp = time.Unix(tokenResp.Expiration, 0)

	log.Println("IAM token refreshed successfully")
	return s.token, nil
}

// CreateSession creates a new Watson Assistant session
func (s *WatsonXService) CreateSession() (string, error) {
	token, err := s.getToken()
	if err != nil {
		return "", err
	}

	reqURL := fmt.Sprintf("%s/v2/assistants/%s/environments/%s/sessions?version=%s",
		s.baseURL, s.assistantID, s.environmentID, s.version)
	log.Printf("Creating session at URL: %s", reqURL)

	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create session request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create session failed with status %d: %s", resp.StatusCode, string(body))
	}

	var sessionResp models.WatsonSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return "", fmt.Errorf("failed to decode session response: %w", err)
	}

	log.Printf("Watson session created: %s", sessionResp.SessionID)
	return sessionResp.SessionID, nil
}

// GetOrCreateSession gets an existing session or creates a new one for a client
func (s *WatsonXService) GetOrCreateSession(clientID string) (string, error) {
	// Try to get existing session from Redis
	if s.redis != nil {
		sessionID, err := s.redis.GetSession(clientID)
		if err != nil {
			log.Printf("⚠️  Error getting session from Redis: %v", err)
		} else if sessionID != "" {
			log.Printf("♻️  Reusing existing session from Redis for client %s: %s", clientID, sessionID)
			return sessionID, nil
		}
	}

	// Create new session
	log.Printf("🆕 Creating new session for client %s", clientID)
	sessionID, err := s.CreateSession()
	if err != nil {
		return "", err
	}

	// Save to Redis with 24 hour TTL
	if s.redis != nil {
		if err := s.redis.SetSession(clientID, sessionID, 24*time.Hour); err != nil {
			log.Printf("⚠️  Error saving session to Redis: %v", err)
		} else {
			log.Printf("✅ Session cached in Redis for client %s: %s", clientID, sessionID)
		}
	}

	return sessionID, nil
}

// RemoveSession removes a session from Redis
func (s *WatsonXService) RemoveSession(clientID string) {
	if s.redis != nil {
		if err := s.redis.DeleteSession(clientID); err != nil {
			log.Printf("⚠️  Error removing session from Redis: %v", err)
		}
	}
}

// GetSessionStats returns session cache statistics from Redis
func (s *WatsonXService) GetSessionStats() map[string]interface{} {
	if s.redis == nil {
		return map[string]interface{}{
			"error": "Redis not available",
		}
	}

	sessions, err := s.redis.GetAllSessions()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	return map[string]interface{}{
		"total_sessions": len(sessions),
		"sessions":       sessions,
	}
}

// ResetAllSessions deletes all cached sessions from Redis
func (s *WatsonXService) ResetAllSessions() (int, error) {
	if s.redis == nil {
		return 0, fmt.Errorf("redis not available")
	}

	deleted, err := s.redis.DeleteAllSessions()
	if err != nil {
		return 0, err
	}

	log.Printf("♻️  Reset complete: %d session(s) deleted from cache", deleted)
	return deleted, nil
}

// SendMessage sends a message to Watson Assistant within a session
func (s *WatsonXService) SendMessage(text, sessionID, userID string) (*models.WatsonMessageResponse, string, error) {
	token, err := s.getToken()
	if err != nil {
		return nil, "", err
	}

	// Use default user ID if not provided
	if userID == "" {
		userID = "default_user"
	}

	// If no sessionID provided but userID is, try to get/create session from cache
	if sessionID == "" && userID != "default_user" {
		log.Printf("🔍 Looking up session for user: %s", userID)
		sessionID, err = s.GetOrCreateSession(userID)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get or create session: %w", err)
		}
	} else if sessionID == "" {
		// Legacy behavior: create session if not provided
		sessionID, err = s.CreateSession()
		if err != nil {
			return nil, "", fmt.Errorf("failed to create session: %w", err)
		}
	}

	reqBody := models.WatsonMessageRequest{
		Input: models.WatsonInput{
			MessageType: "text",
			Text:        text,
		},
		UserID: userID,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Log payload being sent to Watson
	var prettyPayload bytes.Buffer
	json.Indent(&prettyPayload, payload, "", "  ")
	log.Printf("📤 Payload enviado para Watson:\n%s", prettyPayload.String())

	reqURL := fmt.Sprintf("%s/v2/assistants/%s/environments/%s/sessions/%s/message?version=%s",
		s.baseURL, s.assistantID, s.environmentID, sessionID, s.version)

	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)

		// If session not found/expired, remove from cache and retry once
		if resp.StatusCode == 404 && userID != "default_user" {
			log.Printf("⚠️  Session %s expired for user %s, creating new one...", sessionID, userID)
			s.RemoveSession(userID)

			// Retry with new session (only once to avoid infinite loop)
			newSessionID, err := s.GetOrCreateSession(userID)
			if err != nil {
				return nil, "", fmt.Errorf("failed to create new session after expiration: %w", err)
			}

			// Recursive call with new session (will not retry again due to fresh session)
			return s.SendMessage(text, newSessionID, userID)
		}

		return nil, "", fmt.Errorf("watson message failed with status %d: %s", resp.StatusCode, string(body))
	}

	var watsonResp models.WatsonMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&watsonResp); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Log Watson response
	respJSON, _ := json.MarshalIndent(watsonResp, "", "  ")
	log.Printf("Watson response: %s", string(respJSON))

	return &watsonResp, sessionID, nil
}

// DeleteSession deletes a Watson Assistant session
func (s *WatsonXService) DeleteSession(sessionID string) error {
	token, err := s.getToken()
	if err != nil {
		return err
	}

	reqURL := fmt.Sprintf("%s/v2/assistants/%s/environments/%s/sessions/%s?version=%s",
		s.baseURL, s.assistantID, s.environmentID, sessionID, s.version)

	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete session failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("Watson session deleted: %s", sessionID)
	return nil
}
