package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"api/models"
	"api/services"
)

type WhatsAppHandler struct {
	neohub *services.NeoHubService
}

func NewWhatsAppHandler(neohub *services.NeoHubService) *WhatsAppHandler {
	return &WhatsAppHandler{neohub: neohub}
}

// HandleDirect sends a message directly to WhatsApp via NeoHub
// @Summary      Send message directly to WhatsApp
// @Description  Sends a text message to a WhatsApp number via NeoHub broker (bypasses WatsonX flow)
// @Tags         whatsapp
// @Accept       json
// @Produce      json
// @Param        body  body      models.DirectWhatsAppRequest  true  "Message payload"
// @Success      200   {object}  models.APIResponse
// @Failure      400   {object}  models.APIResponse
// @Failure      500   {object}  models.APIResponse
// @Router       /whatsapp/send [post]
func (h *WhatsAppHandler) HandleDirect(w http.ResponseWriter, r *http.Request) {
	var req models.DirectWhatsAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIResponse{Status: "error", Message: "invalid request body"})
		return
	}

	if err := h.neohub.SendMessage(req.To, req.Body); err != nil {
		log.Printf("Error sending message via NeoHub: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{Status: "error", Message: "failed to send message"})
		return
	}

	log.Printf("Direct message sent to %s", req.To)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{Status: "ok", Message: "message sent"})
}
