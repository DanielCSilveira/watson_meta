# Guia de Deploy - Watson WhatsApp API

## Pré-requisitos

- Docker instalado
- Docker Compose instalado (opcional)
- Domínio configurado (para produção)
- SSL/TLS certificado (Let's Encrypt recomendado)

## Opção 1: Docker (Modo Simples)

### 1. Build da Imagem

```bash
cd api
docker build -t watson-whatsapp-api .
```

### 2. Executar Container

```bash
docker run -d \
  --name watson-api \
  -p 8080:8080 \
  -e PORT=8080 \
  -e NEOHUB_BASE_URL=https://broker-api.neobpo.com.br \
  -e NEOHUB_API_KEY=your_key \
  -e NEOHUB_WABA_ID=your_waba_id \
  -e WATSONX_BASE_URL=https://api.us-south.assistant.watson.cloud.ibm.com \
  -e WATSONX_API_KEY=your_watson_key \
  -e WATSONX_ENVIRONMENT_ID=your_env_id \
  -e WATSONX_ASSISTANT_ID=your_assistant_id \
  -e WATSONX_VERSION=2021-11-27 \
  -e REDIS_ADDR=147.93.176.99:8079 \
  -e REDIS_PASSWORD=r3disPa5s \
  --restart unless-stopped \
  watson-whatsapp-api
```

### 3. Verificar Logs

```bash
docker logs -f watson-api
```

### 4. Testar

```bash
curl http://localhost:8080/health
```

## Opção 2: Docker Compose (Recomendado)

### 1. Copiar Arquivo de Exemplo

```bash
cp .env.example .env
```

### 2. Editar Variáveis de Ambiente

Edite o arquivo `.env` com suas credenciais reais.

### 3. Iniciar Serviços

```bash
docker-compose up -d
```

### 4. Ver Logs

```bash
docker-compose logs -f app
```

### 5. Parar Serviços

```bash
docker-compose down
```

## Opção 3: Deploy em Servidor Linux (VPS/Cloud)

### 1. Conectar ao Servidor

```bash
ssh user@seu-servidor.com
```

### 2. Instalar Docker

```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

### 3. Clonar ou Transferir Código

```bash
# Opção A: Git
git clone https://github.com/seu-repo/watson-api.git
cd watson-api

# Opção B: SCP
scp -r ./api user@servidor:/home/user/watson-api/
```

### 4. Configurar Variáveis

```bash
nano .env
# Cole suas credenciais
```

### 5. Build e Run

```bash
docker-compose up -d
```

### 6. Configurar Nginx (Proxy Reverso)

Instalar Nginx:
```bash
sudo apt update
sudo apt install nginx
```

Criar configuração:
```bash
sudo nano /etc/nginx/sites-available/watson-api
```

Conteúdo:
```nginx
server {
    listen 80;
    server_name seu-dominio.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Ativar:
```bash
sudo ln -s /etc/nginx/sites-available/watson-api /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 7. Configurar SSL com Let's Encrypt

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d seu-dominio.com
```

O Certbot vai:
- Gerar certificado SSL
- Configurar HTTPS automaticamente
- Configurar renovação automática

### 8. Verificar Instalação

```bash
curl https://seu-dominio.com/health
```

## Opção 4: Deploy em Cloud Providers

### AWS (EC2 + Elastic Container Registry)

1. **Criar ECR Repository**
```bash
aws ecr create-repository --repository-name watson-whatsapp-api
```

2. **Build e Push**
```bash
docker build -t watson-whatsapp-api .
docker tag watson-whatsapp-api:latest {aws_account_id}.dkr.ecr.{region}.amazonaws.com/watson-whatsapp-api:latest
docker push {aws_account_id}.dkr.ecr.{region}.amazonaws.com/watson-whatsapp-api:latest
```

3. **Deploy no EC2**
- Lançar instância EC2
- Instalar Docker
- Pull e executar imagem do ECR

### Google Cloud (Cloud Run)

```bash
gcloud builds submit --tag gcr.io/{project-id}/watson-whatsapp-api
gcloud run deploy watson-api --image gcr.io/{project-id}/watson-whatsapp-api --platform managed
```

### Azure (Container Instances)

```bash
az container create \
  --resource-group myResourceGroup \
  --name watson-api \
  --image myregistry.azurecr.io/watson-whatsapp-api:latest \
  --ports 8080
```

### DigitalOcean (App Platform)

1. Conectar repositório GitHub
2. Configurar Dockerfile path
3. Adicionar variáveis de ambiente
4. Deploy automático

## Gestão e Monitoramento

### Ver Logs em Tempo Real

```bash
docker-compose logs -f
```

### Verificar Status

```bash
docker-compose ps
docker stats watson-api
```

### Acessar Container

```bash
docker exec -it watson-api /bin/sh
```

### Backup Redis (se usar local)

```bash
docker exec watson-redis redis-cli BGSAVE
docker cp watson-redis:/data/dump.rdb ./backup/
```

### Reiniciar Serviço

```bash
docker-compose restart app
```

### Atualizar Aplicação

```bash
git pull
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## Segurança

### 1. Firewall

```bash
# Permitir apenas portas necessárias
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable
```

### 2. Variáveis de Ambiente

⚠️ **NUNCA** commitar `.env` com credenciais reais!

Use:
- Secrets do Docker
- Vault (HashiCorp)
- AWS Secrets Manager
- Azure Key Vault

### 3. Rate Limiting (Nginx)

```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

server {
    location / {
        limit_req zone=api_limit burst=20 nodelay;
        # ... resto da config
    }
}
```

## Troubleshooting

### Container não inicia

```bash
docker logs watson-api
docker inspect watson-api
```

### Erros de conexão Redis

```bash
# Testar conectividade
docker exec watson-api ping redis
docker exec watson-api nc -zv redis 6379
```

### Porta em uso

```bash
# Ver o que está usando a porta
sudo lsof -i :8080
sudo netstat -tulpn | grep 8080

# Matar processo
sudo kill -9 {PID}
```

### Rebuild completo

```bash
docker-compose down -v
docker-compose build --no-cache
docker-compose up -d
```

## URLs Importantes

Após deploy, sua API estará disponível em:

- **Produção**: `https://seu-dominio.com`
- **Health Check**: `https://seu-dominio.com/health`
- **Swagger**: `https://seu-dominio.com/swagger/`
- **Stats**: `https://seu-dominio.com/watsonx/sessions/stats`

## Webhook da Meta

Configure no painel Meta:
- **Webhook URL**: `https://seu-dominio.com/webhook/meta`
- **Verify Token**: (se implementar verificação)

## Monitoramento Recomendado

1. **Uptime Monitor**: UptimeRobot, Pingdom
2. **APM**: New Relic, DataDog
3. **Logs**: ELK Stack, Graylog
4. **Alertas**: PagerDuty, Opsgenie

## Performance

### Otimizações

1. **Multi-stage build**: Imagem final ~15MB
2. **Cache de dependências**: Build mais rápido
3. **Health checks**: Auto-restart se falhar
4. **Resource limits**: Prevenir overconsumption

### Recursos Recomendados

- **CPU**: 1-2 cores
- **RAM**: 512MB - 1GB
- **Disco**: 10GB (logs + Redis)
- **Banda**: 1TB/mês

## Custos Estimados (USD/mês)

- **DigitalOcean Droplet**: $6-12
- **AWS EC2 t3.micro**: $8-15
- **Google Cloud Run**: $5-20 (pay-per-use)
- **Azure Container**: $10-20

## CI/CD (Opcional)

### GitHub Actions

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Build and push
        run: |
          docker build -t watson-api .
          docker tag watson-api registry.digitalocean.com/myregistry/watson-api
          docker push registry.digitalocean.com/myregistry/watson-api
      - name: Deploy
        run: |
          ssh user@server "docker pull registry.digitalocean.com/myregistry/watson-api && docker-compose up -d"
```

## Suporte

Para problemas, verifique:
1. Logs da aplicação
2. Logs do Nginx/proxy
3. Conectividade Redis
4. Credenciais corretas
5. Firewall/Security Groups

---

**Dica**: Comece com DigitalOcean ou Heroku para simplicidade, depois migre para AWS/GCP para escala.
