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

	log.Printf("\n✅ Watson Response received")

	// Build message from Watson response (can be text or interactive)
	msg, replyText, shouldContinue := s.buildMessageFromWatson(watsonResp, clientID)

	log.Printf("📝 Reply type: %s", msg.Type)
	if msg.Type == "text" {
		log.Printf("💬 Text: %s", replyText)
	} else {
		log.Printf("🔘 Interactive message with options")
	}

	// Send reply back via NeoHub
	log.Printf("\n📨 Sending reply to client %s via NeoHub...", clientID)
	if err := s.neohub.SendStructuredMessage(msg); err != nil {
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

	log.Printf("✅ Watson Continuation Response received")

	// Build message from Watson response
	msg, replyText, shouldContinue := s.buildMessageFromWatson(watsonResp, clientID)

	log.Printf("📝 Reply type: %s", msg.Type)
	if msg.Type == "text" {
		log.Printf("💬 Text: %s", replyText)
	} else {
		log.Printf("🔘 Interactive message with options")
	}

	// Send continuation reply to client
	log.Printf("📨 Sending continuation reply to client %s...", clientID)
	if err := s.neohub.SendStructuredMessage(msg); err != nil {
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

	// Check if this is a status update (not a message)
	if len(value.Statuses) > 0 && len(value.Messages) == 0 {
		log.Printf("⏭️  Status update detected (read/delivered/sent) - ignoring")
		return "", "", fmt.Errorf("IGNORE_STATUS_UPDATE")
	}

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

// buildMessageFromWatson constructs a WhatsApp message from Watson response
// Returns the message, reply text, and whether continuation is needed
func (s *MetaService) buildMessageFromWatson(resp *models.WatsonMessageResponse, clientID string) (*models.OutgoingMessage, string, bool) {
	var textResponse string
	var optionResponse *models.WatsonGeneric
	shouldContinue := false

	// First pass: find text and option responses
	for i := range resp.Output.Generic {
		g := &resp.Output.Generic[i]
		
		if g.ResponseType == "text" && g.Text != "" {
			textResponse = g.Text
			
			// Check for [[CONTINUE]] tag in text
			if strings.HasSuffix(strings.TrimSpace(g.Text), "[[CONTINUE]]") {
				shouldContinue = true
				textResponse = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(g.Text), "[[CONTINUE]]"))
			}
		} else if g.ResponseType == "option" && len(g.Options) > 0 {
			optionResponse = g
		}
	}

	// If no text response found, use default
	if textResponse == "" {
		textResponse = "Desculpe, não consegui processar sua mensagem."
	}

	// Build base message structure
	msg := &models.OutgoingMessage{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               clientID,
	}

	// If there are options, create interactive message
	if optionResponse != nil && len(optionResponse.Options) > 0 {
		log.Printf("🔘 Found %d options from Watson", len(optionResponse.Options))
		
		// Use button type for <= 3 options, list for more
		if len(optionResponse.Options) <= 3 {
			msg.Type = "interactive"
			msg.Interactive = s.buildButtonMessage(textResponse, optionResponse)
		} else {
			msg.Type = "interactive"
			msg.Interactive = s.buildListMessage(textResponse, optionResponse)
		}
	} else {
		// Simple text message
		msg.Type = "text"
		msg.Text = &models.MessageText{
			Body: textResponse,
		}
	}

	return msg, textResponse, shouldContinue
}

// buildButtonMessage creates a button-type interactive message (max 3 buttons)
func (s *MetaService) buildButtonMessage(bodyText string, optionResp *models.WatsonGeneric) *models.InteractiveMessage {
	buttons := make([]models.InteractiveButton, 0, len(optionResp.Options))
	
	for i, opt := range optionResp.Options {
		if i >= 3 {
			break // WhatsApp allows max 3 buttons
		}
		
		buttons = append(buttons, models.InteractiveButton{
			Type: "reply",
			Reply: models.InteractiveButtonReply{
				ID:    fmt.Sprintf("opt_%d", i),
				Title: truncateText(opt.Label, 20), // WhatsApp button title max 20 chars
			},
		})
	}
	
	header := optionResp.Title
	if header == "" {
		header = "Escolha uma opção"
	}
	
	return &models.InteractiveMessage{
		Type: "button",
		Body: models.InteractiveBody{
			Text: bodyText,
		},
		Action: models.InteractiveAction{
			Buttons: buttons,
		},
	}
}

// buildListMessage creates a list-type interactive message (4-10 options)
func (s *MetaService) buildListMessage(bodyText string, optionResp *models.WatsonGeneric) *models.InteractiveMessage {
	rows := make([]models.InteractiveRow, 0, len(optionResp.Options))
	
	for i, opt := range optionResp.Options {
		if i >= 10 {
			break // WhatsApp allows max 10 list items
		}
		
		rows = append(rows, models.InteractiveRow{
			ID:          fmt.Sprintf("opt_%d", i),
			Title:       truncateText(opt.Label, 24), // WhatsApp row title max 24 chars
			Description: truncateText(optionResp.Description, 72), // Max 72 chars
		})
	}
	
	header := optionResp.Title
	if header == "" {
		header = "Escolha uma opção"
	}
	
	buttonText := "Ver opções"
	
	return &models.InteractiveMessage{
		Type: "list",
		Header: &models.InteractiveHeader{
			Type: "text",
			Text: header,
		},
		Body: models.InteractiveBody{
			Text: bodyText,
		},
		Action: models.InteractiveAction{
			Button: buttonText,
			Sections: []models.InteractiveSection{
				{
					Title: "Opções",
					Rows:  rows,
				},
			},
		},
	}
}

// truncateText truncates text to max length
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
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
