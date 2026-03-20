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
	MessagingProduct string              `json:"messaging_product"`
	RecipientType    string              `json:"recipient_type"`
	To               string              `json:"to"`
	Type             string              `json:"type"`
	Text             *MessageText        `json:"text,omitempty"`
	Interactive      *InteractiveMessage `json:"interactive,omitempty"`
}

// MessageText represents the text content of a WhatsApp message
type MessageText struct {
	Body string `json:"body"`
}

// InteractiveMessage represents an interactive message (buttons or list)
type InteractiveMessage struct {
	Type   string             `json:"type"` // "button" or "list"
	Header *InteractiveHeader `json:"header,omitempty"`
	Body   InteractiveBody    `json:"body"`
	Footer *InteractiveFooter `json:"footer,omitempty"`
	Action InteractiveAction  `json:"action"`
}

type InteractiveHeader struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"`
}

type InteractiveBody struct {
	Text string `json:"text"`
}

type InteractiveFooter struct {
	Text string `json:"text"`
}

type InteractiveAction struct {
	Button   string               `json:"button,omitempty"`   // For list type
	Buttons  []InteractiveButton  `json:"buttons,omitempty"`  // For button type (max 3)
	Sections []InteractiveSection `json:"sections,omitempty"` // For list type
}

type InteractiveButton struct {
	Type  string                 `json:"type"` // "reply"
	Reply InteractiveButtonReply `json:"reply"`
}

type InteractiveButtonReply struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type InteractiveSection struct {
	Title string           `json:"title,omitempty"`
	Rows  []InteractiveRow `json:"rows"`
}

type InteractiveRow struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
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
	ResponseType string         `json:"response_type"`
	Text         string         `json:"text,omitempty"`
	Title        string         `json:"title,omitempty"`
	Description  string         `json:"description,omitempty"`
	Options      []WatsonOption `json:"options,omitempty"`
}

type WatsonOption struct {
	Label string            `json:"label"`
	Value WatsonOptionValue `json:"value"`
}

type WatsonOptionValue struct {
	Input WatsonOptionInput `json:"input,omitempty"`
}

type WatsonOptionInput struct {
	Text string `json:"text,omitempty"`
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
	Contacts         []MetaContact `json:"contacts,omitempty"`
	Messages         []MetaMessage `json:"messages,omitempty"`
	Statuses         []MetaStatus  `json:"statuses,omitempty"`
}

type MetaStatus struct {
	ID           string                 `json:"id"`
	Status       string                 `json:"status"`
	Timestamp    string                 `json:"timestamp"`
	RecipientID  string                 `json:"recipient_id"`
	Conversation map[string]interface{} `json:"conversation,omitempty"`
	Pricing      map[string]interface{} `json:"pricing,omitempty"`
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
	From        string                `json:"from"`
	ID          string                `json:"id"`
	Timestamp   string                `json:"timestamp"`
	Type        string                `json:"type"`
	Text        *MetaTextBody         `json:"text,omitempty"`
	Interactive *MetaInteractiveReply `json:"interactive,omitempty"`
	Button      *MetaButtonReply      `json:"button,omitempty"`
}

type MetaTextBody struct {
	Body string `json:"body"`
}

// MetaInteractiveReply represents when user selects from list or clicks button
type MetaInteractiveReply struct {
	Type        string               `json:"type"` // "button_reply" or "list_reply"
	ButtonReply *MetaButtonReplyData `json:"button_reply,omitempty"`
	ListReply   *MetaListReplyData   `json:"list_reply,omitempty"`
}

type MetaButtonReplyData struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type MetaListReplyData struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// MetaButtonReply represents legacy button click (deprecated but still possible)
type MetaButtonReply struct {
	Payload string `json:"payload"`
	Text    string `json:"text"`
}
