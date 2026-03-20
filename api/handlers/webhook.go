package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"api/models"
	"api/services"
)

type WebhookHandler struct {
	meta *services.MetaService
}

func NewWebhookHandler(meta *services.MetaService) *WebhookHandler {
	return &WebhookHandler{
		meta: meta,
	}
}

// HandleMetaWebhook receives webhook from Meta/WhatsApp Business API
// @Summary      Webhook for Meta WhatsApp Business API
// @Description  Receives messages from Meta WhatsApp Business API, processes through Watson, and sends reply back
// @Tags         webhook
// @Accept       json
// @Produce      json
// @Param        body  body      models.MetaWebhookPayload  true  "Meta webhook payload"
// @Success      200   {object}  models.APIResponse
// @Failure      400   {object}  models.APIResponse
// @Failure      500   {object}  models.APIResponse
// @Router       /webhook/meta [post]
func (h *WebhookHandler) HandleMetaWebhook(w http.ResponseWriter, r *http.Request) {
	log.Printf("📥 POST /webhook/meta - Meta webhook received")

	// Read body for debugging
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Error reading request body: %v", err)
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	// Log raw JSON received
	log.Printf("📦 Raw JSON received:\n%s", string(bodyBytes))

	// Restore body for decoding
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var payload models.MetaWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("❌ Error decoding Meta webhook payload: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Process webhook and send reply (all logic in service)
	if err := h.meta.ProcessAndReply(&payload); err != nil {
		// If it's a status update, just ignore it (don't return error)
		if err.Error() == "failed to extract message data: IGNORE_STATUS_UPDATE" {
			log.Printf("ℹ️  Status update webhook - returning 200 OK")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(models.APIResponse{Status: "ok", Message: "status update ignored"})
			return
		}
		log.Printf("❌ Error processing Meta webhook: %v", err)
		http.Error(w, "failed to process webhook", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{Status: "ok"})
}
