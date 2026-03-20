package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"api/config"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisService(cfg *config.Config) *RedisService {
	if cfg.RedisAddr == "" {
		log.Println("⚠️  Redis not configured, sessions will not be persisted")
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("❌ Failed to connect to Redis: %v", err)
		return nil
	}

	log.Printf("✅ Connected to Redis at %s", cfg.RedisAddr)
	return &RedisService{
		client: client,
		ctx:    ctx,
	}
}

// SetSession stores a session in Redis with TTL
func (r *RedisService) SetSession(clientID, sessionID string, ttl time.Duration) error {
	if r == nil || r.client == nil {
		return fmt.Errorf("redis not available")
	}

	sessionData := map[string]interface{}{
		"session_id": sessionID,
		"created_at": time.Now().Unix(),
		"last_used":  time.Now().Unix(),
	}

	data, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	key := fmt.Sprintf("watson:session:%s", clientID)
	if err := r.client.Set(r.ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set session in redis: %w", err)
	}

	log.Printf("💾 Session saved to Redis for client %s (TTL: %v)", clientID, ttl)
	return nil
}

// GetSession retrieves a session from Redis
func (r *RedisService) GetSession(clientID string) (string, error) {
	if r == nil || r.client == nil {
		return "", fmt.Errorf("redis not available")
	}

	key := fmt.Sprintf("watson:session:%s", clientID)
	data, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Session not found
	}
	if err != nil {
		return "", fmt.Errorf("failed to get session from redis: %w", err)
	}

	var sessionData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
		return "", fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	// Update last_used timestamp
	sessionData["last_used"] = time.Now().Unix()
	updatedData, _ := json.Marshal(sessionData)
	r.client.Set(r.ctx, key, updatedData, 24*time.Hour) // Refresh TTL

	sessionID := sessionData["session_id"].(string)
	log.Printf("📥 Session retrieved from Redis for client %s: %s", clientID, sessionID)
	return sessionID, nil
}

// DeleteSession removes a session from Redis
func (r *RedisService) DeleteSession(clientID string) error {
	if r == nil || r.client == nil {
		return fmt.Errorf("redis not available")
	}

	key := fmt.Sprintf("watson:session:%s", clientID)
	if err := r.client.Del(r.ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete session from redis: %w", err)
	}

	log.Printf("🗑️  Session deleted from Redis for client %s", clientID)
	return nil
}

// GetAllSessions returns all session keys and their data
func (r *RedisService) GetAllSessions() (map[string]interface{}, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("redis not available")
	}

	keys, err := r.client.Keys(r.ctx, "watson:session:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session keys: %w", err)
	}

	sessions := make(map[string]interface{})
	for _, key := range keys {
		data, err := r.client.Get(r.ctx, key).Result()
		if err != nil {
			continue
		}

		var sessionData map[string]interface{}
		if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
			continue
		}

		// Extract client ID from key
		clientID := key[len("watson:session:"):]
		sessions[clientID] = sessionData
	}

	return sessions, nil
}

// DeleteAllSessions removes all Watson sessions from Redis
func (r *RedisService) DeleteAllSessions() (int, error) {
	if r == nil || r.client == nil {
		return 0, fmt.Errorf("redis not available")
	}

	keys, err := r.client.Keys(r.ctx, "watson:session:*").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get session keys: %w", err)
	}

	if len(keys) == 0 {
		log.Printf("ℹ️  No sessions to delete")
		return 0, nil
	}

	deleted, err := r.client.Del(r.ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to delete sessions: %w", err)
	}

	log.Printf("🗑️  Deleted %d session(s) from Redis", deleted)
	return int(deleted), nil
}

// Close closes the Redis connection
func (r *RedisService) Close() error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Close()
}
