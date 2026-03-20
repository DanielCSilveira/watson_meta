# Watson WhatsApp API 🤖💬

API que integra WhatsApp Business (via Meta/NeoHub) com IBM Watson Assistant, permitindo conversas automatizadas com clientes.

## 🚀 Features

- ✅ Integração com Meta WhatsApp Business API
- ✅ IBM Watson Assistant v2
- ✅ Cache de sessões com Redis
- ✅ Sistema de continuação automática `[[CONTINUE]]`
- ✅ Logs detalhados e rastreamento
- ✅ Documentação Swagger/OpenAPI
- ✅ Health checks
- ✅ Docker e Docker Compose prontos

## 📋 Pré-requisitos

- Go 1.21+ (desenvolvimento)
- Docker (produção)
- Redis (pode usar externo ou Docker)
- Credenciais da Meta WhatsApp Business API
- Credenciais do IBM Watson Assistant
- Credenciais do NeoHub

## 🛠️ Instalação Local

### 1. Clone o repositório

```bash
git clone https://github.com/seu-usuario/watson-whatsapp-api.git
cd watson-whatsapp-api
```

### 2. Configure variáveis de ambiente

```bash
cd api
cp .env.example .env
# Edite .env com suas credenciais
```

### 3. Instale dependências

```bash
go mod download
```

### 4. Execute

```bash
go run main.go
```

A API estará disponível em `http://localhost:8080`

## 🐳 Deploy com Docker

### Build e Run (modo rápido)

```bash
cd api
docker build -t watson-whatsapp-api .
docker run -p 8080:8080 --env-file .env watson-whatsapp-api
```

### Docker Compose (recomendado)

```bash
# Configure o .env primeiro
docker-compose up -d

# Ver logs
docker-compose logs -f

# Parar
docker-compose down
```

Veja [DEPLOY.md](DEPLOY.md) para instruções completas de deploy em produção.

## 📚 Documentação API

Acesse a documentação interativa Swagger:
```
http://localhost:8080/swagger/
```

## 🔌 Endpoints Principais

### Webhook

- `POST /webhook/meta` - Recebe mensagens da Meta/WhatsApp

### Watson Assistant

- `POST /watsonx/session` - Cria nova sessão
- `POST /watsonx/message` - Envia mensagem direta ao Watson
- `GET /watsonx/sessions/stats` - Estatísticas de sessões
- `DELETE /watsonx/sessions/reset` - Reseta todas as sessões

### WhatsApp

- `POST /whatsapp/send` - Envia mensagem direta ao WhatsApp

### Utilitários

- `GET /health` - Health check

## 🎯 Fluxo de Funcionamento

```
Cliente WhatsApp
       ↓
Meta Webhook → POST /webhook/meta
       ↓
MetaService (extrai texto + clientID)
       ↓
Watson Assistant (processa mensagem)
       ↓
       → Resposta com [[CONTINUE]]? 
       ↓ Sim: Aguarda 3s → nova chamada
       ↓ Não: Envia direto
       ↓
NeoHub → WhatsApp → Cliente
```

## 🔄 Sistema de Continuação

O Watson pode enviar múltiplas mensagens sequenciais usando a tag `[[CONTINUE]]`:

**Exemplo:**

Resposta 1: `"Buscando produtos...[[CONTINUE]]"`
- Sistema envia: "Buscando produtos..."
- Aguarda 3 segundos
- Chama Watson novamente com ""

Resposta 2: `"Encontrei 5 produtos!"`
- Sistema envia segunda mensagem

Veja [CONTINUE_TAG.md](CONTINUE_TAG.md) para detalhes.

## 💾 Cache de Sessões (Redis)

Sessões do Watson são armazenadas no Redis para:
- Manter contexto da conversa
- Reutilizar sessões (reduz calls ao Watson)
- Persistir entre reinícios
- Compartilhar entre múltiplas instâncias

TTL padrão: 24 horas

Veja [REDIS_INTEGRATION.md](REDIS_INTEGRATION.md) para detalhes.

## ⚙️ Variáveis de Ambiente

```env
# Servidor
PORT=8080

# NeoHub (WhatsApp)
NEOHUB_BASE_URL=https://broker-api.neobpo.com.br
NEOHUB_API_KEY=your_key
NEOHUB_WABA_ID=your_waba_id

# Watson Assistant
WATSONX_BASE_URL=https://api.us-south.assistant.watson.cloud.ibm.com
WATSONX_API_KEY=your_watson_key
WATSONX_ENVIRONMENT_ID=your_env_id
WATSONX_ASSISTANT_ID=your_assistant_id
WATSONX_VERSION=2021-11-27

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=your_password
```

## 🧪 Testes

```bash
# Testar webhook Meta
.\test_meta_webhook.ps1

# Testar sessões Redis
.\test_redis.ps1

# Gerenciar sessões
.\manage_sessions.ps1
```

## 📊 Monitoramento

### Ver estatísticas de sessões

```bash
curl http://localhost:8080/watsonx/sessions/stats
```

### Health check

```bash
curl http://localhost:8080/health
```

### Logs

```bash
# Docker Compose
docker-compose logs -f app

# Docker
docker logs -f watson-api

# Local
# Logs aparecem no console
```

## 🔒 Segurança

- ✅ Nunca commitar `.env` com credenciais
- ✅ Usar HTTPS em produção
- ✅ Configurar rate limiting
- ✅ Validar webhooks da Meta (verify token)
- ✅ Usar secrets manager em produção

## 📁 Estrutura do Projeto

```
.
├── api/
│   ├── config/          # Configurações
│   ├── docs/            # Swagger docs (gerado)
│   ├── handlers/        # HTTP handlers
│   ├── models/          # Data models
│   ├── services/        # Business logic
│   ├── main.go          # Entry point
│   ├── Dockerfile       # Container config
│   └── .env             # Environment vars
├── docker-compose.yml   # Multi-container
├── DEPLOY.md            # Deploy guide
├── CONTINUE_TAG.md      # Continuation docs
├── REDIS_INTEGRATION.md # Redis docs
└── README.md            # This file
```

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## 📝 TODO

- [ ] Autenticação de webhooks Meta
- [ ] Suporte a mais tipos de mensagem (imagem, áudio, etc)
- [ ] Métricas com Prometheus
- [ ] Testes unitários
- [ ] CI/CD pipeline
- [ ] Rate limiting
- [ ] Retry logic melhorado

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo LICENSE para mais detalhes.

## 👥 Autores

- Seu Nome - [seu-usuario](https://github.com/seu-usuario)

## 🆘 Suporte

Para problemas ou dúvidas:
1. Verifique a [documentação](DEPLOY.md)
2. Abra uma [issue](https://github.com/seu-usuario/watson-whatsapp-api/issues)
3. Entre em contato: seu-email@example.com

---

Feito com ❤️ usando Go, Watson e WhatsApp
