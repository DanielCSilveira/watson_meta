# Quick Start - Watson WhatsApp API 🚀

## 1️⃣ Desenvolvimento Local (Go)

```bash
# 1. Entre na pasta
cd api

# 2. Configure .env
cp .env.example .env
# Edite .env com suas credenciais

# 3. Execute
go run main.go
```

Pronto! API rodando em `http://localhost:8080` 🎉

---

## 2️⃣ Docker (Modo Rápido)

```bash
# 1. Configure .env
cp .env.example .env
# Edite com suas credenciais

# 2. Build e Run
cd api
docker build -t watson-api .
docker run -p 8080:8080 --env-file .env watson-api
```

Acesse: `http://localhost:8080` 🐳

---

## 3️⃣ Docker Compose (Recomendado)

```bash
# 1. Configure .env
cp .env.example .env
# Edite com suas credenciais

# 2. Inicie tudo
docker-compose up -d

# 3. Ver logs
docker-compose logs -f
```

Acesse: `http://localhost:8080` 🎊

---

## 4️⃣ Deploy em Servidor Linux

```bash
# 1. Conectar ao servidor
ssh user@seu-servidor.com

# 2. Instalar Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 3. Transferir arquivos
git clone https://github.com/seu-repo/watson-api.git
cd watson-api

# 4. Configurar
nano .env
# Cole suas credenciais

# 5. Deploy
docker-compose up -d

# 6. Configurar domínio
sudo apt install nginx certbot python3-certbot-nginx
sudo nano /etc/nginx/sites-available/watson-api
# Configure proxy reverso (veja DEPLOY.md)
sudo certbot --nginx -d seu-dominio.com
```

Acesse: `https://seu-dominio.com` 🌐

---

## 5️⃣ Verificar Instalação

```bash
# Health check
curl http://localhost:8080/health
# Deve retornar: {"text":"OK"}

# Swagger
curl http://localhost:8080/swagger/
# Abre documentação interativa

# Estatísticas de sessões
curl http://localhost:8080/watsonx/sessions/stats
```

---

## 📋 Variáveis Obrigatórias (.env)

```env
# NeoHub
NEOHUB_BASE_URL=https://broker-api.neobpo.com.br
NEOHUB_API_KEY=seu_token_neohub
NEOHUB_WABA_ID=seu_waba_id

# Watson
WATSONX_BASE_URL=https://api.us-south.assistant.watson.cloud.ibm.com
WATSONX_API_KEY=seu_token_watson
WATSONX_ENVIRONMENT_ID=seu_env_id
WATSONX_ASSISTANT_ID=seu_assistant_id

# Redis
REDIS_ADDR=147.93.176.99:8079
REDIS_PASSWORD=r3disPa5s
```

---

## 🔧 Comandos Úteis

```bash
# Ver logs
docker-compose logs -f

# Reiniciar
docker-compose restart

# Parar
docker-compose down

# Atualizar
git pull
docker-compose down
docker-compose up -d --build

# Limpar tudo
docker-compose down -v
docker system prune -a
```

---

## 🌐 Configurar Webhook Meta

No painel Meta Developer:

1. Acesse: https://developers.facebook.com/apps
2. Vá em WhatsApp > Configuration
3. Configure:
   - **Webhook URL**: `https://seu-dominio.com/webhook/meta`
   - **Verify Token**: (opcional)
   - **Subscribe**: `messages`

---

## 🎯 Próximos Passos

✅ API rodando  
→ Configure webhook Meta  
→ Teste enviando mensagem WhatsApp  
→ Veja logs: `docker-compose logs -f`  
→ Acesse Swagger: `http://seu-dominio.com/swagger/`  

---

## 🆘 Problemas Comuns

### Container não inicia
```bash
docker logs watson-api
# Verifique erros nas credenciais
```

### Porta em uso
```bash
sudo lsof -i :8080
# Mate o processo ou use outra porta
```

### Redis não conecta
```bash
docker exec watson-api ping redis
# Verifique REDIS_ADDR no .env
```

---

## 📚 Documentação Completa

- [README.md](README.md) - Documentação geral
- [DEPLOY.md](DEPLOY.md) - Deploy detalhado
- [CONTINUE_TAG.md](CONTINUE_TAG.md) - Sistema [[CONTINUE]]
- [REDIS_INTEGRATION.md](REDIS_INTEGRATION.md) - Cache Redis

---

## 💡 Dica Rápida

Use DigitalOcean ou Linode para deploy fácil:
- Crie um droplet Ubuntu 22.04
- Rode o script acima
- Configure domínio
- Pronto! 🎉

**Custo**: ~$6-12/mês
