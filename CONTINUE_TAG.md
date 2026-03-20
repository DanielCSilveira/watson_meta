# Sistema de Continuação Automática Watson - [[CONTINUE]]

## Resumo

Sistema que permite ao Watson enviar múltiplas mensagens sequenciais ao cliente sem aguardar resposta, usando a tag especial `[[CONTINUE]]`.

## Como Funciona

### Fluxo Normal (sem continuação)
```
1. Cliente envia: "Olá"
2. Watson responde: "Olá! Como posso ajudar?"
3. Cliente envia: "Quero fazer uma compra"
4. Watson responde: "Ótimo! Qual produto?"
```

### Fluxo com [[CONTINUE]]
```
1. Cliente envia: "Olá"
2. Watson responde: "Vou pesquisar isso para você...[[CONTINUE]]"
   ↓
   Sistema detecta [[CONTINUE]]
   - Envia: "Vou pesquisar isso para você..."
   - Aguarda 3 segundos
   - Chama Watson com ""
   ↓
3. Watson responde: "Encontrei! Aqui estão os resultados:..."
   ↓
   Sistema envia segunda mensagem
```

## Implementação

### Detecção da Tag

```go
shouldContinue := strings.HasSuffix(strings.TrimSpace(replyText), "[[CONTINUE]]")
```

A tag deve estar **exatamente no final** da resposta do Watson.

### Processo

#### 1. Resposta Inicial
- Watson responde com `[[CONTINUE]]` no final
- Sistema remove a tag da mensagem
- Envia mensagem limpa ao cliente
- Inicia processo assíncrono de continuação

#### 2. Continuação (após 3s)
- Chama Watson com mensagem vazia `""`
- Watson retorna próxima resposta
- Verifica se tem outra tag `[[CONTINUE]]`
- Envia resposta ao cliente
- Se tiver outra tag, continua o processo (recursivo)

### Características

✅ **Assíncrono**: Não bloqueia resposta HTTP  
✅ **Recursivo**: Suporta múltiplas continuações em cadeia  
✅ **Mantém Contexto**: Usa mesma sessão do Redis  
✅ **Delay Configurado**: 3 segundos entre mensagens  
✅ **Logs Detalhados**: Rastreamento completo do processo  

## Configuração no Watson

### Exemplo de Resposta Watson

**Resposta 1 (com continuação):**
```json
{
  "output": {
    "generic": [
      {
        "response_type": "text",
        "text": "Vou pesquisar os produtos disponíveis para você...[[CONTINUE]]"
      }
    ]
  }
}
```

**Resposta 2 (após chamada vazia):**
```json
{
  "output": {
    "generic": [
      {
        "response_type": "text",
        "text": "Encontrei 5 produtos que podem te interessar:\n1. Produto A\n2. Produto B\n..."
      }
    ]
  }
}
```

### Múltiplas Continuações

```
Watson Resposta 1: "Etapa 1 concluída...[[CONTINUE]]"
  ↓ (3s)
Watson Resposta 2: "Etapa 2 concluída...[[CONTINUE]]"
  ↓ (3s)
Watson Resposta 3: "Processo finalizado!"
```

## Logs

### Quando [[CONTINUE]] é Detectado

```
✅ Watson Response: Vou pesquisar isso para você...[[CONTINUE]]
🔄 [[CONTINUE]] tag detected - will fetch next message after sending this one
📨 Sending reply to client 5511998648847 via NeoHub...
✅ Successfully sent reply to client 5511998648847
⏳ Scheduling continuation call in 3 seconds...
========================================

========================================
🔄 Processing Continuation for client 5511998648847
========================================
⏳ Waiting 3 seconds before continuation...
📤 Sending continuation request to Watson (empty message)...
   Client ID: 5511998648847
   Message: "" (empty - continuation)
✅ Watson Continuation Response: Encontrei os resultados!
📨 Sending continuation reply to client 5511998648847...
✅ Successfully sent continuation reply to client 5511998648847
========================================
```

### Múltiplas Continuações

```
🔄 [[CONTINUE]] tag detected - will fetch next message after sending this one
⏳ Scheduling continuation call in 3 seconds...

...

🔄 Another [[CONTINUE]] tag detected in continuation response
🔄 Chaining another continuation...
```

## Casos de Uso

### 1. Busca/Pesquisa Externa
```
Cliente: "Busque produtos de tecnologia"
Watson: "Estou buscando os produtos...[[CONTINUE]]"
Watson: "Encontrei 10 produtos!"
```

### 2. Processamento em Etapas
```
Cliente: "Faça meu pedido"
Watson: "Validando estoque...[[CONTINUE]]"
Watson: "Calculando frete...[[CONTINUE]]"
Watson: "Pedido confirmado!"
```

### 3. Integração com APIs
```
Cliente: "Qual a previsão do tempo?"
Watson: "Consultando API...[[CONTINUE]]"
Watson: "Hoje: 25°C, ensolarado!"
```

## Erro e Tratamento

### Erro na Continuação

```go
if err != nil {
    log.Printf("❌ Error in continuation call to Watson: %v", err)
    return // Não envia nada, apenas loga erro
}
```

**Comportamento**: Se a continuação falhar, o cliente recebe apenas a primeira mensagem. O erro é logado mas não interfere no fluxo principal.

### Sessão Expirada

O sistema usa a mesma sessão Redis, então se a sessão expirar durante a continuação, o Watson automaticamente cria uma nova (comportamento já existente).

## Limitações

### Delay Fixo
- 3 segundos hard-coded
- Pode ser parametrizado se necessário

### Sem Confirmação
- Cliente não pode cancelar continuação
- Processo é automático e irreversível

### Goroutine
- Continuação roda em background
- Não há limite de continuações simultâneas

## Melhorias Futuras (Opcional)

1. **Delay Configurável**: Via env ou parâmetro
2. **Limite de Continuações**: Evitar loops infinitos
3. **Cancelamento**: Cliente pode interromper
4. **Timeout**: Limite de tempo para continuação
5. **Retry**: Tentar novamente em caso de erro
6. **Métricas**: Quantidade de continuações por dia

## Teste Manual

### 1. Configure Watson

No Watson Assistant, crie uma resposta que termine com `[[CONTINUE]]`:

```
Vou pesquisar isso para você, aguarde um momento...[[CONTINUE]]
```

E configure o próximo step/action para responder quando receber mensagem vazia.

### 2. Execute o Servidor

```powershell
go run main.go
```

### 3. Envie Mensagem

```powershell
# Use o teste normal
..\test_meta_webhook.ps1
```

### 4. Observe os Logs

Você verá:
- Primeira resposta sendo enviada
- Detecção do [[CONTINUE]]
- Espera de 3 segundos
- Segunda chamada ao Watson
- Segunda resposta sendo enviada

## Exemplo Completo

### Timeline

```
t=0s    Cliente envia: "Busque produtos"
t=0.5s  Watson responde: "Buscando...[[CONTINUE]]"
t=0.6s  Sistema envia: "Buscando..."
t=0.7s  Sistema agenda continuação
t=3.7s  Sistema chama Watson com ""
t=4.2s  Watson responde: "Encontrei 5 produtos!"
t=4.3s  Sistema envia: "Encontrei 5 produtos!"
```

### Perspectiva do Cliente (WhatsApp)

```
14:30:15 Cliente: Busque produtos
14:30:16 Bot: Buscando...
14:30:19 Bot: Encontrei 5 produtos!
             1. Produto A
             2. Produto B
             ...
```

Cliente recebe duas mensagens seguidas sem precisar responder!

## Debug

### Verificar se Tag está Correta

```go
log.Printf("Raw response: '%s'", replyText)
log.Printf("Has suffix: %v", strings.HasSuffix(strings.TrimSpace(replyText), "[[CONTINUE]]"))
```

### Testar Continuação Manualmente

Você pode testar a continuação diretamente:

```go
go metaService.processContinuation("5511998648847")
```

## Segurança

### Prevenção de Loops

Se o Watson sempre responder com `[[CONTINUE]]`, haverá loop infinito. Recomendações:

1. **Watson deve ter lógica de parada**: Não retornar `[[CONTINUE]]` eternamente
2. **Implementar contador** (opcional): Limitar a 5 continuações
3. **Timeout** (opcional): Limitar tempo total de continuação

### Exemplo de Proteção (Opcional)

```go
// Adicionar ao MetaService
maxContinuations := 5
currentCount := 0

func (s *MetaService) processContinuation(clientID string, depth int) {
    if depth >= maxContinuations {
        log.Printf("⚠️  Max continuations reached for %s", clientID)
        return
    }
    
    // ... resto do código ...
    
    if shouldContinue {
        go s.processContinuation(clientID, depth+1)
    }
}
```

## Performance

### Impacto

- **HTTP Response**: Não impactado (continuação é async)
- **Goroutines**: Uma por continuação
- **Memória**: Mínimo (goroutine é leve)
- **Redis**: Mesma sessão reutilizada

### Monitoramento

```
# Métricas recomendadas
- Quantidade de continuações por minuto
- Tempo médio de continuação
- Taxa de erro em continuações
- Clientes com mais continuações
```

## Conclusão

Sistema simples mas poderoso que permite ao Watson:
- ✅ Informar que está processando
- ✅ Enviar resultado quando pronto
- ✅ Múltiplas etapas de feedback
- ✅ Melhor experiência do usuário

Tudo acontece automaticamente ao detectar `[[CONTINUE]]` no final da resposta!
