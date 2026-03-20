# Integração Redis - Cache de Sessões Watson

## Resumo

Sistema de cache de sessões do Watson Assistant usando Redis como backend persistente.

## Arquitetura

```
Cliente → Meta Webhook → MetaService → WatsonXService → Redis
                                              ↓
                                         Watson API
```

## Vantagens do Redis

### ✅ Antes (Memória)
- ❌ Perdia todas as sessões ao reiniciar o servidor
- ❌ Não escalável (cada instância tem seu próprio cache)
- ❌ Sem persistência

### ✅ Agora (Redis)
- ✅ Sessões persistidas entre reinícios
- ✅ Escalável (múltiplas instâncias compartilham o mesmo cache)
- ✅ TTL automático (sessões expiram após 24h)
- ✅ Gerenciamento de memória automático

## Configuração

### Variáveis de Ambiente (.env)
```env
REDIS_ADDR=147.93.176.99:8079
REDIS_PASSWORD=r3disPa5s
```

### Estrutura das Chaves Redis
```
watson:session:{clientID}  →  {
    "session_id": "abc123...",
    "created_at": 1710950000,
    "last_used": 1710950100
}
```

### TTL
- **Padrão**: 24 horas
- **Renovação**: Automática a cada acesso (last_used atualizado)
- **Expiração**: Gerenciada pelo Redis

## Implementação

### 1. RedisService (services/redis.go)

```go
// Métodos principais
SetSession(clientID, sessionID string, ttl time.Duration) error
GetSession(clientID string) (string, error)
DeleteSession(clientID string) error
GetAllSessions() (map[string]interface{}, error)
```

**Recursos:**
- Conexão automática com ping test
- Fallback gracioso se Redis não estiver disponível
- Serialização JSON automática
- Renovação de TTL em cada acesso

### 2. WatsonXService (services/watsonx.go)

**Modificações:**
- ❌ Removido: `map[string]*SessionInfo` em memória
- ❌ Removido: `sync.RWMutex` 
- ❌ Removido: `CleanupOldSessions()`
- ✅ Adicionado: `redis *RedisService`
- ✅ Modificado: `GetOrCreateSession()` usa Redis
- ✅ Modificado: `RemoveSession()` usa Redis
- ✅ Modificado: `GetSessionStats()` usa Redis

### 3. Main (main.go)

```go
// Ordem de inicialização
redisService := services.NewRedisService(cfg)
watsonxService := services.NewWatsonXService(cfg, redisService)
```

## Fluxo de Dados

### Primeira Mensagem
```
1. Cliente envia mensagem
2. WatsonXService.GetOrCreateSession(clientID)
3. Redis.GetSession(clientID) → não encontrado
4. Cria nova sessão no Watson
5. Redis.SetSession(clientID, sessionID, 24h)
6. Retorna sessionID
```

### Mensagens Subsequentes
```
1. Cliente envia mensagem
2. WatsonXService.GetOrCreateSession(clientID)
3. Redis.GetSession(clientID) → encontrado!
4. Atualiza last_used no Redis
5. Renova TTL para +24h
6. Retorna sessionID existente
```

### Sessão Expirada (404 do Watson)
```
1. Envio falha com 404
2. Redis.DeleteSession(clientID)
3. Recursive call → cria nova sessão
4. Retry automático com nova sessão
```

## Logs

### Conexão Redis
```
✅ Connected to Redis at 147.93.176.99:8079
❌ Failed to connect to Redis: connection refused
⚠️  Redis not configured, sessions will not be persisted
```

### Operações de Cache
```
💾 Session saved to Redis for client 5511998648847 (TTL: 24h0m0s)
📥 Session retrieved from Redis for client 5511998648847: abc123
♻️  Reusing existing session from Redis for client 5511998648847: abc123
🗑️  Session deleted from Redis for client 5511998648847
```

### Sessões
```
🆕 Creating new session for client 5511998648847
✅ Session cached in Redis for client 5511998648847: abc123
⚠️  Error saving session to Redis: connection timeout
```

## Endpoints

### GET /watsonx/sessions/stats
Retorna estatísticas de sessões armazenadas no Redis

**Resposta:**
```json
{
  "total_sessions": 2,
  "sessions": {
    "5511998648847": {
      "session_id": "abc123...",
      "created_at": 1710950000,
      "last_used": 1710950100
    },
    "5511999999999": {
      "session_id": "def456...",
      "created_at": 1710951000,
      "last_used": 1710951200
    }
  }
}
```

## Testes

### Teste Básico
```powershell
# Iniciar servidor
go run main.go

# Executar testes
..\test_redis.ps1
```

### Verificar Sessões
```powershell
# Via API
Invoke-RestMethod http://localhost:8080/watsonx/sessions/stats

# Via Redis CLI (se disponível)
redis-cli -h 147.93.176.99 -p 8079 -a r3disPa5s KEYS "watson:session:*"
redis-cli -h 147.93.176.99 -p 8079 -a r3disPa5s GET "watson:session:5511998648847"
```

## Monitoramento

### Métricas Importantes
- **total_sessions**: Número de sessões ativas
- **created_at**: Timestamp de criação
- **last_used**: Timestamp do último acesso

### Alertas Recomendados
- Redis connection failures
- TTL muito curto (múltiplas criações da mesma sessão)
- Crescimento anormal do número de sessões

## Troubleshooting

### Redis não conecta
```
❌ Failed to connect to Redis: dial tcp 147.93.176.99:8079: i/o timeout
```
**Solução**: Verificar firewall, credenciais e disponibilidade do Redis

### Sessões não persistem
```
⚠️  Redis not configured, sessions will not be persisted
```
**Solução**: Verificar .env com REDIS_ADDR e REDIS_PASSWORD

### Overhead de latência
**Sintoma**: Requisições mais lentas
**Solução**: Redis local ou otimizar TTL/refresh

## Próximos Passos (Opcionais)

1. **Redis Cluster**: Para alta disponibilidade
2. **Métricas**: Exportar para Prometheus
3. **TTL Configurável**: Via variável de ambiente
4. **Cache Warming**: Pre-criar sessões
5. **Compressão**: Reduzir tamanho dos dados armazenados
