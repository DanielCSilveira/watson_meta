package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"api/config"
	"api/models"
)

type NeoHubService struct {
	baseURL    string
	apiKey     string
	wabaID     string
	httpClient *http.Client
}

func NewNeoHubService(cfg *config.Config) *NeoHubService {
	return &NeoHubService{
		baseURL:    cfg.NeoHubBaseURL,
		apiKey:     cfg.NeoHubAPIKey,
		wabaID:     cfg.NeoHubWabaID,
		httpClient: &http.Client{},
	}
}

// SendMessage sends a message to a WhatsApp number via NeoHub
func (s *NeoHubService) SendMessage(to, body string) error {
	msg := models.OutgoingMessage{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               to,
		Type:             "text",
		Text: &models.MessageText{
			Body: body,
		},
	}
	
	return s.SendStructuredMessage(&msg)
}

// SendStructuredMessage sends a structured message (can be text or interactive) via NeoHub
func (s *NeoHubService) SendStructuredMessage(msg *models.OutgoingMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Log payload being sent
	var prettyPayload bytes.Buffer
	json.Indent(&prettyPayload, payload, "", "  ")
	fmt.Printf("Sending to NeoHub: %s\n", prettyPayload.String())

	sendURL := fmt.Sprintf("%s/v1/%s/messages", s.baseURL, s.wabaID)
	fmt.Printf("URL: %s\n", sendURL)

	req, err := http.NewRequest("POST", sendURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body := new(bytes.Buffer)
		body.ReadFrom(resp.Body)
		return fmt.Errorf("neohub returned status %d: %s", resp.StatusCode, body.String())
	}

	fmt.Printf("NeoHub response status: %d\n", resp.StatusCode)
	return nil
}
