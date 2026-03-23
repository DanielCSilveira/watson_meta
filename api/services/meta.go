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

	// Extract text, client ID, and message ID from payload
	text, clientID, messageID, err := s.extractMessageData(payload)
	if err != nil {
		return fmt.Errorf("failed to extract message data: %w", err)
	}

	log.Printf("\n📤 Sending to Watson Assistant...")
	log.Printf("   Client ID: %s", clientID)
	log.Printf("   Message: %s", text)

	// Send message to Watson
	watsonResp, sessionID, err := s.watsonx.SendMessage(text, "", clientID)
	if err != nil {
		return fmt.Errorf("failed to send message to Watson: %w", err)
	}

	// Mark message as read on WhatsApp
	log.Printf("\n👁️  Marking message as read on WhatsApp...")
	if err := s.neohub.MarkAsRead(messageID); err != nil {
		log.Printf("⚠️  Warning: Failed to mark message as read: %v", err)
		// Don't fail the whole flow if marking as read fails
	}

	log.Printf("\n✅ Watson Response received")
	log.Printf("   Session ID: %s", sessionID)

	// Build messages from Watson response (can be multiple text messages and/or interactive)
	msgs, shouldContinue := s.buildMessagesFromWatson(watsonResp, clientID)

	log.Printf("📝 Watson returned %d message(s)", len(msgs))

	// Send each message back via NeoHub
	for i, msg := range msgs {
		log.Printf("\n📨 Sending message %d/%d to client %s via NeoHub...", i+1, len(msgs), clientID)
		log.Printf("   Type: %s", msg.Type)

		if err := s.neohub.SendStructuredMessage(msg); err != nil {
			return fmt.Errorf("failed to send message %d via NeoHub to %s: %w", i+1, clientID, err)
		}

		log.Printf("✅ Message %d/%d sent successfully", i+1, len(msgs))

		// Small delay between multiple messages for better UX
		if i < len(msgs)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	log.Printf("✅ All messages sent to client %s", clientID)

	// If [[CONTINUE]] tag was present, fetch next response asynchronously
	if shouldContinue {
		log.Printf("⏳ Scheduling continuation call in 3 seconds...")
		go s.processContinuation(clientID, sessionID)
	}

	log.Printf("========================================\n")

	return nil
}

// processContinuation handles continuation of Watson responses
func (s *MetaService) processContinuation(clientID, sessionID string) {
	log.Printf("\n========================================")
	log.Printf("🔄 Processing Continuation for client %s", clientID)
	log.Printf("========================================")

	// Wait 3 seconds before calling Watson again
	log.Printf("⏳ Waiting 3 seconds before continuation...")
	time.Sleep(3 * time.Second)

	log.Printf("📤 Sending continuation request to Watson (empty message)...")
	log.Printf("   Client ID: %s", clientID)
	log.Printf("   Session ID: %s", sessionID)
	log.Printf("   Message: \"\" (empty - continuation)")

	// Send empty message to Watson to get next response using the same session
	watsonResp, _, err := s.watsonx.SendMessage("", sessionID, clientID)
	if err != nil {
		log.Printf("❌ Error in continuation call to Watson: %v", err)
		return
	}

	log.Printf("✅ Watson Continuation Response received")

	// Build messages from Watson response
	msgs, shouldContinue := s.buildMessagesFromWatson(watsonResp, clientID)

	log.Printf("📝 Continuation returned %d message(s)", len(msgs))

	// Send each continuation message to client
	for i, msg := range msgs {
		log.Printf("📨 Sending continuation message %d/%d to client %s...", i+1, len(msgs), clientID)
		log.Printf("   Type: %s", msg.Type)

		if err := s.neohub.SendStructuredMessage(msg); err != nil {
			log.Printf("❌ Error sending continuation message %d: %v", i+1, err)
			return
		}

		log.Printf("✅ Continuation message %d/%d sent successfully", i+1, len(msgs))

		// Small delay between multiple messages
		if i < len(msgs)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	log.Printf("✅ All continuation messages sent to client %s", clientID)

	// If another [[CONTINUE]] tag was present, continue the chain
	if shouldContinue {
		log.Printf("🔄 Chaining another continuation...")
		go s.processContinuation(clientID, sessionID)
	}

	log.Printf("========================================\n")
}

// extractMessageData extracts the text, client ID, and message ID from Meta webhook payload
func (s *MetaService) extractMessageData(payload *models.MetaWebhookPayload) (text string, clientID string, messageID string, err error) {
	log.Printf("=== Extracting data from Meta webhook ===")
	log.Printf("Payload Object: %s", payload.Object)
	log.Printf("Number of entries: %d", len(payload.Entry))

	if len(payload.Entry) == 0 {
		return "", "", "", fmt.Errorf("no entries in payload")
	}

	entry := payload.Entry[0]
	log.Printf("Entry ID: %s", entry.ID)
	log.Printf("Number of changes: %d", len(entry.Changes))

	if len(entry.Changes) == 0 {
		return "", "", "", fmt.Errorf("no changes in entry")
	}

	change := entry.Changes[0]
	value := change.Value
	log.Printf("Change field: %s", change.Field)
	log.Printf("Messaging product: %s", value.MessagingProduct)

	// Extract client ID from contacts
	log.Printf("Number of contacts: %d", len(value.Contacts))
	if len(value.Contacts) == 0 {
		return "", "", "", fmt.Errorf("no contacts in payload")
	}

	contact := value.Contacts[0]
	clientID = contact.WaID
	log.Printf("📱 CLIENTE/DESTINATÁRIO: %s (Nome: %s)", clientID, contact.Profile.Name)

	// Check if this is a status update (not a message)
	if len(value.Statuses) > 0 && len(value.Messages) == 0 {
		log.Printf("⏭️  Status update detected (read/delivered/sent) - ignoring")
		return "", "", "", fmt.Errorf("IGNORE_STATUS_UPDATE")
	}

	// Extract text from messages
	log.Printf("Number of messages: %d", len(value.Messages))
	if len(value.Messages) == 0 {
		return "", "", "", fmt.Errorf("no messages in payload")
	}

	message := value.Messages[0]
	messageID = message.ID
	log.Printf("Message ID: %s", messageID)
	log.Printf("Message from: %s", message.From)
	log.Printf("Message type: %s", message.Type)
	log.Printf("Message timestamp: %s", message.Timestamp)

	// Extract text based on message type
	switch message.Type {
	case "text":
		if message.Text != nil {
			text = message.Text.Body
			log.Printf("💬 TEXT MESSAGE: '%s'", text)
		}

	case "interactive":
		// User clicked on button or selected from list
		if message.Interactive != nil {
			if message.Interactive.Type == "button_reply" && message.Interactive.ButtonReply != nil {
				text = message.Interactive.ButtonReply.Title
				log.Printf("🔘 BUTTON CLICKED: '%s' (ID: %s)", text, message.Interactive.ButtonReply.ID)
			} else if message.Interactive.Type == "list_reply" && message.Interactive.ListReply != nil {
				text = message.Interactive.ListReply.Title
				log.Printf("📋 LIST ITEM SELECTED: '%s' (ID: %s)", text, message.Interactive.ListReply.ID)
			}
		}

	case "button":
		// Legacy button format (deprecated but still possible)
		if message.Button != nil {
			text = message.Button.Text
			log.Printf("🔘 LEGACY BUTTON CLICKED: '%s'", text)
		}

	default:
		log.Printf("⚠️  Unsupported message type: %s", message.Type)
		return "", "", "", fmt.Errorf("message type not supported: %s", message.Type)
	}

	if text == "" {
		return "", "", "", fmt.Errorf("empty message text")
	}

	log.Printf("=== Extraction complete ===")

	return text, clientID, messageID, nil
}

// buildMessagesFromWatson constructs WhatsApp messages from Watson response
// Returns array of messages and whether continuation is needed
// Processes all text responses in the generic array as separate messages
func (s *MetaService) buildMessagesFromWatson(resp *models.WatsonMessageResponse, clientID string) ([]*models.OutgoingMessage, bool) {
	var messages []*models.OutgoingMessage
	var optionResponse *models.WatsonGeneric
	shouldContinue := false

	// Process each item in generic array
	for i := range resp.Output.Generic {
		g := &resp.Output.Generic[i]

		if g.ResponseType == "text" && g.Text != "" {
			textResponse := g.Text

			// Check for [[CONTINUE]] tag in text
			if strings.HasSuffix(strings.TrimSpace(g.Text), "[[CONTINUE]]") {
				shouldContinue = true
				textResponse = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(g.Text), "[[CONTINUE]]"))
			}

			// Create a text message for each text response
			msg := &models.OutgoingMessage{
				MessagingProduct: "whatsapp",
				RecipientType:    "individual",
				To:               clientID,
				Type:             "text",
				Text: &models.MessageText{
					Body: textResponse,
				},
			}
			messages = append(messages, msg)
			log.Printf("💬 Text message %d: %s", len(messages), truncateText(textResponse, 50))

		} else if g.ResponseType == "option" && len(g.Options) > 0 {
			optionResponse = g
		}
	}

	// If there are options, create interactive message
	// Options are added as a final message after all text messages
	if optionResponse != nil && len(optionResponse.Options) > 0 {
		log.Printf("🔘 Found %d options from Watson", len(optionResponse.Options))

		// Get the last text message to use as body for interactive message
		bodyText := "Escolha uma opção:"
		if len(messages) > 0 && messages[len(messages)-1].Text != nil {
			// Use last text message as body and remove it from separate messages
			bodyText = messages[len(messages)-1].Text.Body
			messages = messages[:len(messages)-1]
		}

		msg := &models.OutgoingMessage{
			MessagingProduct: "whatsapp",
			RecipientType:    "individual",
			To:               clientID,
			Type:             "interactive",
		}

		// Use button type for <= 3 options, list for more
		if len(optionResponse.Options) <= 3 {
			msg.Interactive = s.buildButtonMessage(bodyText, optionResponse)
		} else {
			msg.Interactive = s.buildListMessage(bodyText, optionResponse)
		}

		messages = append(messages, msg)
	}

	// If no messages were created, return a default message
	if len(messages) == 0 {
		msg := &models.OutgoingMessage{
			MessagingProduct: "whatsapp",
			RecipientType:    "individual",
			To:               clientID,
			Type:             "text",
			Text: &models.MessageText{
				Body: "Desculpe, não consegui processar sua mensagem.",
			},
		}
		messages = append(messages, msg)
	}

	return messages, shouldContinue
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
			Title:       truncateText(opt.Label, 24),              // WhatsApp row title max 24 chars
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
