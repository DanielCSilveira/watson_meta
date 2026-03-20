package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"api/models"
)

type MetaService struct {
	watsonx *WatsonXService
	neohub  *NeoHubService
}

func NewMetaService(watsonx *WatsonXService, neohub *NeoHubService) *MetaService {
	return &MetaService{
		watsonx: watsonx,
		neohub:  neohub,
	}
}

// ProcessAndReply processes incoming webhook from Meta/WhatsApp Business API
// Extracts text and client ID, sends to Watson, and sends reply back via NeoHub
func (s *MetaService) ProcessAndReply(payload *models.MetaWebhookPayload) error {
	log.Printf("\n========================================")
	log.Printf("🔔 Meta Webhook Received")
	log.Printf("========================================")

	// Extract text and client ID from payload
	text, clientID, err := s.extractMessageData(payload)
	if err != nil {
		return fmt.Errorf("failed to extract message data: %w", err)
	}

	log.Printf("\n📤 Sending to Watson Assistant...")
	log.Printf("   Client ID: %s", clientID)
	log.Printf("   Message: %s", text)

	// Send message to Watson
	watsonResp, _, err := s.watsonx.SendMessage(text, "", clientID)
	if err != nil {
		return fmt.Errorf("failed to send message to Watson: %w", err)
	}

	// Extract response text from Watson
	replyText := extractResponseText(watsonResp)

	log.Printf("\n✅ Watson Response: %s", replyText)

	// Check if response contains [[CONTINUE]] tag
	shouldContinue := strings.HasSuffix(strings.TrimSpace(replyText), "[[CONTINUE]]")

	// Remove [[CONTINUE]] tag from message before sending
	if shouldContinue {
		replyText = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(replyText), "[[CONTINUE]]"))
		log.Printf("🔄 [[CONTINUE]] tag detected - will fetch next message after sending this one")
	}

	// Send reply back via NeoHub
	log.Printf("\n📨 Sending reply to client %s via NeoHub...", clientID)
	if err := s.neohub.SendMessage(clientID, replyText); err != nil {
		return fmt.Errorf("failed to send reply via NeoHub to %s: %w", clientID, err)
	}

	log.Printf("✅ Successfully sent reply to client %s", clientID)

	// If [[CONTINUE]] tag was present, fetch next response asynchronously
	if shouldContinue {
		log.Printf("⏳ Scheduling continuation call in 3 seconds...")
		go s.processContinuation(clientID)
	}

	log.Printf("========================================\n")

	return nil
}

// processContinuation handles continuation of Watson responses
func (s *MetaService) processContinuation(clientID string) {
	log.Printf("\n========================================")
	log.Printf("🔄 Processing Continuation for client %s", clientID)
	log.Printf("========================================")

	// Wait 3 seconds before calling Watson again
	log.Printf("⏳ Waiting 3 seconds before continuation...")
	time.Sleep(3 * time.Second)

	log.Printf("📤 Sending continuation request to Watson (empty message)...")
	log.Printf("   Client ID: %s", clientID)
	log.Printf("   Message: \"\" (empty - continuation)")

	// Send empty message to Watson to get next response
	watsonResp, _, err := s.watsonx.SendMessage("", "", clientID)
	if err != nil {
		log.Printf("❌ Error in continuation call to Watson: %v", err)
		return
	}

	// Extract response text from Watson
	replyText := extractResponseText(watsonResp)
	log.Printf("✅ Watson Continuation Response: %s", replyText)

	// Check if this response also has [[CONTINUE]] tag (recursive continuation)
	shouldContinue := strings.HasSuffix(strings.TrimSpace(replyText), "[[CONTINUE]]")

	// Remove [[CONTINUE]] tag from message before sending
	if shouldContinue {
		replyText = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(replyText), "[[CONTINUE]]"))
		log.Printf("🔄 Another [[CONTINUE]] tag detected in continuation response")
	}

	// Send continuation reply to client
	log.Printf("📨 Sending continuation reply to client %s...", clientID)
	if err := s.neohub.SendMessage(clientID, replyText); err != nil {
		log.Printf("❌ Error sending continuation reply: %v", err)
		return
	}

	log.Printf("✅ Successfully sent continuation reply to client %s", clientID)

	// If another [[CONTINUE]] tag was present, continue the chain
	if shouldContinue {
		log.Printf("🔄 Chaining another continuation...")
		go s.processContinuation(clientID)
	}

	log.Printf("========================================\n")
}

// extractMessageData extracts the text and client ID from Meta webhook payload
func (s *MetaService) extractMessageData(payload *models.MetaWebhookPayload) (text string, clientID string, err error) {
	log.Printf("=== Extracting data from Meta webhook ===")
	log.Printf("Payload Object: %s", payload.Object)
	log.Printf("Number of entries: %d", len(payload.Entry))

	if len(payload.Entry) == 0 {
		return "", "", fmt.Errorf("no entries in payload")
	}

	entry := payload.Entry[0]
	log.Printf("Entry ID: %s", entry.ID)
	log.Printf("Number of changes: %d", len(entry.Changes))

	if len(entry.Changes) == 0 {
		return "", "", fmt.Errorf("no changes in entry")
	}

	change := entry.Changes[0]
	value := change.Value
	log.Printf("Change field: %s", change.Field)
	log.Printf("Messaging product: %s", value.MessagingProduct)

	// Extract client ID from contacts
	log.Printf("Number of contacts: %d", len(value.Contacts))
	if len(value.Contacts) == 0 {
		return "", "", fmt.Errorf("no contacts in payload")
	}

	contact := value.Contacts[0]
	clientID = contact.WaID
	log.Printf("📱 CLIENTE/DESTINATÁRIO: %s (Nome: %s)", clientID, contact.Profile.Name)

	// Extract text from messages
	log.Printf("Number of messages: %d", len(value.Messages))
	if len(value.Messages) == 0 {
		return "", "", fmt.Errorf("no messages in payload")
	}

	message := value.Messages[0]
	log.Printf("Message ID: %s", message.ID)
	log.Printf("Message from: %s", message.From)
	log.Printf("Message type: %s", message.Type)
	log.Printf("Message timestamp: %s", message.Timestamp)

	if message.Type == "text" && message.Text != nil {
		text = message.Text.Body
	} else {
		return "", "", fmt.Errorf("message type not supported: %s", message.Type)
	}

	if text == "" {
		return "", "", fmt.Errorf("empty message text")
	}

	log.Printf("💬 MENSAGEM RECEBIDA: '%s'", text)
	log.Printf("=== Extraction complete ===")

	return text, clientID, nil
}

// extractResponseText extracts text from Watson response
func extractResponseText(resp *models.WatsonMessageResponse) string {
	for _, g := range resp.Output.Generic {
		if g.ResponseType == "text" && g.Text != "" {
			return g.Text
		}
	}
	return "Desculpe, não consegui processar sua mensagem."
}
