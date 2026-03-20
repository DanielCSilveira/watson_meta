package models

// --- NeoHub (WhatsApp) ---

// IncomingMessage represents a message received from NeoHub webhook
type IncomingMessage struct {
	From      string `json:"from"`
	Body      string `json:"body"`
	MessageID string `json:"message_id"`
	Timestamp string `json:"timestamp"`
}

// OutgoingMessage represents a message to be sent via NeoHub
type OutgoingMessage struct {
	MessagingProduct string      `json:"messaging_product"`
	RecipientType    string      `json:"recipient_type"`
	To               string      `json:"to"`
	Type             string      `json:"type"`
	Text             MessageText `json:"text"`
}

// MessageText represents the text content of a WhatsApp message
type MessageText struct {
	Body string `json:"body"`
}

// --- Watson Assistant v2 ---

// WatsonMessageRequest is the payload sent to Watson Assistant v2 message endpoint
type WatsonMessageRequest struct {
	Input   WatsonInput    `json:"input"`
	Context *WatsonContext `json:"context,omitempty"`
	UserID  string         `json:"user_id"`
}

type WatsonInput struct {
	MessageType string `json:"message_type"`
	Text        string `json:"text"`
}

type WatsonContext struct {
	Skills map[string]WatsonSkill `json:"skills,omitempty"`
}

type WatsonSkill struct {
	UserDefined map[string]interface{} `json:"user_defined,omitempty"`
}

// WatsonMessageResponse is the response from Watson Assistant v2 message endpoint
type WatsonMessageResponse struct {
	Output WatsonOutput `json:"output"`
}

type WatsonOutput struct {
	Generic  []WatsonGeneric `json:"generic"`
	Intents  []WatsonIntent  `json:"intents,omitempty"`
	Entities []WatsonEntity  `json:"entities,omitempty"`
}

type WatsonGeneric struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

type WatsonIntent struct {
	Intent     string  `json:"intent"`
	Confidence float64 `json:"confidence"`
}

type WatsonEntity struct {
	Entity     string  `json:"entity"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
}

// WatsonSessionResponse is the response when creating a session
type WatsonSessionResponse struct {
	SessionID string `json:"session_id"`
}

// IAMTokenRequest / IAMTokenResponse for IBM Cloud IAM authentication
type IAMTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Expiration  int64  `json:"expiration"`
}

// --- API Request/Response (endpoints auxiliares) ---

// DirectWatsonXRequest represents a direct message request to WatsonX
type DirectWatsonXRequest struct {
	Text      string `json:"text" example:"Olá, preciso de ajuda"`
	SessionID string `json:"session_id,omitempty" example:""`
	UserID    string `json:"user_id,omitempty" example:"user123"`
}

// DirectWhatsAppRequest represents a direct message request to WhatsApp via NeoHub
type DirectWhatsAppRequest struct {
	To   string `json:"to" example:"5511999999999"`
	Body string `json:"body" example:"Olá, esta é uma mensagem de teste"`
}

// APIResponse represents a generic API response
type APIResponse struct {
	Status  string `json:"status" example:"ok"`
	Message string `json:"message,omitempty" example:"message sent"`
}

// WatsonXDirectResponse represents the response for direct WatsonX calls
type WatsonXDirectResponse struct {
	Status    string `json:"status" example:"ok"`
	Reply     string `json:"reply" example:"Olá! Como posso ajudar?"`
	SessionID string `json:"session_id" example:"abc123"`
}

// CreateSessionResponse represents the response when creating a new Watson session
type CreateSessionResponse struct {
	Status    string `json:"status" example:"ok"`
	SessionID string `json:"session_id" example:"abc123"`
}

// --- Meta (WhatsApp Business API) ---

// MetaWebhookPayload represents the webhook payload from Meta/WhatsApp Business API
type MetaWebhookPayload struct {
	Object string      `json:"object"`
	Entry  []MetaEntry `json:"entry"`
}

type MetaEntry struct {
	ID      string       `json:"id"`
	Changes []MetaChange `json:"changes"`
}

type MetaChange struct {
	Value MetaValue `json:"value"`
	Field string    `json:"field"`
}

type MetaValue struct {
	MessagingProduct string        `json:"messaging_product"`
	Metadata         MetaMetadata  `json:"metadata"`
	Contacts         []MetaContact `json:"contacts"`
	Messages         []MetaMessage `json:"messages"`
}

type MetaMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type MetaContact struct {
	Profile MetaProfile `json:"profile"`
	WaID    string      `json:"wa_id"`
}

type MetaProfile struct {
	Name string `json:"name"`
}

type MetaMessage struct {
	From      string        `json:"from"`
	ID        string        `json:"id"`
	Timestamp string        `json:"timestamp"`
	Type      string        `json:"type"`
	Text      *MetaTextBody `json:"text,omitempty"`
}

type MetaTextBody struct {
	Body string `json:"body"`
}
