package main

import (
	"encoding/json"
	"log"
	"net/http"

	"api/config"
	_ "api/docs"
	"api/handlers"
	"api/services"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Message struct {
	Text string `json:"text"`
}

// @title           WhatsApp WatsonX API
// @version         1.0
// @description     API que intermedia a comunicação entre WhatsApp (via NeoHub) e WatsonX Assistant.

// @host      localhost:8080
// @BasePath  /

func main() {
	cfg := config.Load()

	// Initialize services
	redisService := services.NewRedisService(cfg)
	neohubService := services.NewNeoHubService(cfg)
	watsonxService := services.NewWatsonXService(cfg, redisService)
	metaService := services.NewMetaService(watsonxService, neohubService)

	webhookHandler := handlers.NewWebhookHandler(metaService)
	watsonxHandler := handlers.NewWatsonXHandler(watsonxService)
	whatsappHandler := handlers.NewWhatsAppHandler(neohubService)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Message{Text: "OK"})
	})

	// Webhook Meta - receives WhatsApp messages from Meta/WhatsApp Business API
	mux.HandleFunc("POST /webhook/meta", webhookHandler.HandleMetaWebhook)

	// Direct endpoints (auxiliary)
	mux.HandleFunc("POST /watsonx/session", watsonxHandler.HandleCreateSession)
	mux.HandleFunc("POST /watsonx/message", watsonxHandler.HandleDirect)
	mux.HandleFunc("GET /watsonx/sessions/stats", watsonxHandler.HandleSessionStats)
	mux.HandleFunc("DELETE /watsonx/sessions/reset", watsonxHandler.HandleResetSessions)
	mux.HandleFunc("POST /whatsapp/send", whatsappHandler.HandleDirect)

	// Swagger
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	log.Printf("Server running on :%s", cfg.Port)
	log.Printf("Swagger available at http://localhost:%s/swagger/", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
