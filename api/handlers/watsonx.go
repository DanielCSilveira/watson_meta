package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"api/models"
	"api/services"
)

type WatsonXHandler struct {
	watsonx *services.WatsonXService
}

func NewWatsonXHandler(watsonx *services.WatsonXService) *WatsonXHandler {
	return &WatsonXHandler{watsonx: watsonx}
}

// HandleDirect sends a message directly to WatsonX and returns the response
// @Summary      Send message directly to WatsonX
// @Description  Sends a text message to WatsonX Assistant and returns the response (bypasses WhatsApp flow)
// @Tags         watsonx
// @Accept       json
// @Produce      json
// @Param        body  body      models.DirectWatsonXRequest  true  "Message payload"
// @Success      200   {object}  models.WatsonXDirectResponse
// @Failure      400   {object}  models.APIResponse
// @Failure      500   {object}  models.APIResponse
// @Router       /watsonx/message [post]
func (h *WatsonXHandler) HandleDirect(w http.ResponseWriter, r *http.Request) {
	var req models.DirectWatsonXRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.APIResponse{Status: "error", Message: "invalid request body"})
		return
	}

	watsonResp, sessionID, err := h.watsonx.SendMessage(req.Text, req.SessionID, req.UserID)
	if err != nil {
		log.Printf("Error communicating with WatsonX: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{Status: "error", Message: "failed to communicate with WatsonX"})
		return
	}

	replyText := extractResponseText(watsonResp)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.WatsonXDirectResponse{
		Status:    "ok",
		Reply:     replyText,
		SessionID: sessionID,
	})
}

// HandleCreateSession creates a new Watson Assistant session
// @Summary      Create a new Watson session
// @Description  Creates a new session with Watson Assistant and returns the session ID
// @Tags         watsonx
// @Accept       json
// @Produce      json
// @Success      200   {object}  models.CreateSessionResponse
// @Failure      500   {object}  models.APIResponse
// @Router       /watsonx/session [post]
func (h *WatsonXHandler) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := h.watsonx.CreateSession()
	if err != nil {
		log.Printf("Error creating Watson session: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{Status: "error", Message: "failed to create session"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.CreateSessionResponse{
		Status:    "ok",
		SessionID: sessionID,
	})
}

// HandleSessionStats returns session cache statistics
// @Summary      Get session cache statistics
// @Description  Returns information about cached sessions
// @Tags         watsonx
// @Produce      json
// @Success      200   {object}  map[string]interface{}
// @Router       /watsonx/sessions/stats [get]
func (h *WatsonXHandler) HandleSessionStats(w http.ResponseWriter, r *http.Request) {
	stats := h.watsonx.GetSessionStats()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// HandleResetSessions deletes all cached sessions
// @Summary      Reset all sessions
// @Description  Deletes all cached Watson sessions from Redis
// @Tags         watsonx
// @Produce      json
// @Success      200   {object}  map[string]interface{}
// @Failure      500   {object}  models.APIResponse
// @Router       /watsonx/sessions/reset [delete]
func (h *WatsonXHandler) HandleResetSessions(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔄 Reset sessions request received")

	deleted, err := h.watsonx.ResetAllSessions()
	if err != nil {
		log.Printf("❌ Error resetting sessions: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.APIResponse{Status: "error", Message: err.Error()})
		return
	}

	log.Printf("✅ Successfully reset %d session(s)", deleted)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "sessions reset successfully",
		"deleted": deleted,
	})
}

func extractResponseText(resp *models.WatsonMessageResponse) string {
	for _, g := range resp.Output.Generic {
		if g.ResponseType == "text" && g.Text != "" {
			return g.Text
		}
	}
	return "Desculpe, não consegui processar sua mensagem."
}
