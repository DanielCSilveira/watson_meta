# Integração Meta WhatsApp Business API → Watson

## Resumo da Implementação

Implementação completa do fluxo de integração entre Meta WhatsApp Business API e Watson Assistant.

## Fluxo de Dados

```
Meta WhatsApp → POST /webhook/meta → MetaService → Watson → Resposta → NeoHub → WhatsApp
```

1. **Meta envia webhook** com mensagem do usuário
2. **Endpoint /webhook/meta** recebe o payload
3. **MetaService extrai**:
   - Texto da mensagem: `entry[0].changes[0].value.messages[0].text.body`
   - Client ID (wa_id): `entry[0].changes[0].value.contacts[0].wa_id`
4. **Watson processa** a mensagem e retorna resposta
5. **NeoHub envia** a resposta de volta via WhatsApp

## Arquivos Criados/Modificados

### 1. `models/message.go`
- Adicionados modelos do formato Meta:
  - `MetaWebhookPayload`
  - `MetaEntry`, `MetaChange`, `MetaValue`
  - `MetaMetadata`, `MetaContact`, `MetaProfile`
  - `MetaMessage`, `MetaTextBody`

### 2. `services/meta.go` (NOVO)
- `ProcessWebhook()`: Processa payload da Meta e envia para Watson
- `extractMessageData()`: Extrai texto e client ID do payload
- `extractResponseText()`: Extrai resposta do Watson

### 3. `handlers/webhook.go`
- Adicionado `MetaService` ao `WebhookHandler`
- `HandleMetaWebhook()`: Handler para webhook da Meta

### 4. `main.go`
- Inicialização da `MetaService`
- Rota `POST /webhook/meta` registrada

## Endpoints

### POST /webhook/meta
Recebe webhooks da Meta/WhatsApp Business API

**Formato do Payload:**
```json
{
  "object": "whatsapp_business_account",
  "entry": [{
    "id": "1608476237088134",
    "changes": [{
      "value": {
        "messaging_product": "whatsapp",
        "metadata": {
          "display_phone_number": "551150289262",
          "phone_number_id": "1050528654811352"
        },
        "contacts": [{
          "profile": {"name": "Daniel Silveira"},
          "wa_id": "5511998648847"
        }],
        "messages": [{
          "from": "5511998648847",
          "id": "wamid.xxx",
          "timestamp": "1774019272",
          "text": {"body": "Olá"},
          "type": "text"
        }]
      },
      "field": "messages"
    }]
  }]
}
```

**Resposta:**
```json
{
  "status": "ok"
}
```

## Como Testar

### Opção 1: Script PowerShell
```powershell
.\test_meta_webhook.ps1
```

### Opção 2: cURL
```bash
curl -X POST http://localhost:8080/webhook/meta \
  -H "Content-Type: application/json" \
  -d @msg_meta.txt
```

### Opção 3: Postman
Importe a collection existente e adicione request para `/webhook/meta`

## Configuração no Meta

No painel de desenvolvedor da Meta, configure:
- **Webhook URL**: `https://seu-dominio.com/webhook/meta`
- **Verify Token**: (opcional - implementar se necessário)
- **Campos de webhook**: `messages`

## Próximos Passos (Opcionais)

1. **Verificação de token**: Implementar GET /webhook/meta para verificação da Meta
2. **Suporte a outros tipos**: Imagens, áudio, documentos
3. **Status de leitura**: Marcar mensagens como listas
4. **Rate limiting**: Controle de taxa de requisições
5. **Logs estruturados**: Melhorar logging para produção
6. **Testes unitários**: Adicionar testes automatizados

## Observações

- O fluxo usa **NeoHub** para enviar respostas
- O **wa_id** é usado como identificador do usuário no Watson
- Apenas mensagens de **tipo texto** são processadas atualmente
- Erros são logados e retornam status 500 com mensagem genérica
