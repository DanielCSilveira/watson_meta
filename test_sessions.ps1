# Script para testar cache de sessões do Watson

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "🧪 TESTE DE CACHE DE SESSÕES WATSON" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$statsUrl = "http://localhost:8080/watsonx/sessions/stats"

# Função para ver estatísticas
function Get-SessionStats {
    Write-Host "📊 Estatísticas de Sessões:" -ForegroundColor Yellow
    try {
        $stats = Invoke-RestMethod -Uri $statsUrl -Method Get
        $stats | ConvertTo-Json -Depth 5
        Write-Host ""
    } catch {
        Write-Host "❌ Erro ao buscar estatísticas: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Ver estatísticas iniciais
Write-Host "1️⃣ Estatísticas ANTES das mensagens:" -ForegroundColor Green
Get-SessionStats

# Enviar mensagem via Meta webhook (cliente 1)
Write-Host "2️⃣ Enviando mensagem do Cliente 1 (5511998648847)..." -ForegroundColor Green
$body1 = @{
    object = "whatsapp_business_account"
    entry = @(
        @{
            id = "1608476237088134"
            changes = @(
                @{
                    value = @{
                        messaging_product = "whatsapp"
                        metadata = @{
                            display_phone_number = "551150289262"
                            phone_number_id = "1050528654811352"
                        }
                        contacts = @(
                            @{
                                profile = @{ name = "Cliente 1" }
                                wa_id = "5511998648847"
                            }
                        )
                        messages = @(
                            @{
                                from = "5511998648847"
                                id = "msg1"
                                timestamp = "1774019272"
                                text = @{ body = "Olá, primeira mensagem" }
                                type = "text"
                            }
                        )
                    }
                    field = "messages"
                }
            )
        }
    )
} | ConvertTo-Json -Depth 10

try {
    Invoke-RestMethod -Uri "http://localhost:8080/webhook/meta" -Method Post -Body $body1 -ContentType "application/json" | Out-Null
    Write-Host "✅ Mensagem enviada" -ForegroundColor Green
} catch {
    Write-Host "❌ Erro: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Ver estatísticas após primeira mensagem
Write-Host "3️⃣ Estatísticas APÓS primeira mensagem:" -ForegroundColor Green
Get-SessionStats

# Enviar segunda mensagem do mesmo cliente
Write-Host "4️⃣ Enviando SEGUNDA mensagem do Cliente 1 (mesma sessão)..." -ForegroundColor Green
$body2 = @{
    object = "whatsapp_business_account"
    entry = @(
        @{
            id = "1608476237088134"
            changes = @(
                @{
                    value = @{
                        messaging_product = "whatsapp"
                        metadata = @{
                            display_phone_number = "551150289262"
                            phone_number_id = "1050528654811352"
                        }
                        contacts = @(
                            @{
                                profile = @{ name = "Cliente 1" }
                                wa_id = "5511998648847"
                            }
                        )
                        messages = @(
                            @{
                                from = "5511998648847"
                                id = "msg2"
                                timestamp = "1774019273"
                                text = @{ body = "Olá, segunda mensagem" }
                                type = "text"
                            }
                        )
                    }
                    field = "messages"
                }
            )
        }
    )
} | ConvertTo-Json -Depth 10

try {
    Invoke-RestMethod -Uri "http://localhost:8080/webhook/meta" -Method Post -Body $body2 -ContentType "application/json" | Out-Null
    Write-Host "✅ Mensagem enviada" -ForegroundColor Green
} catch {
    Write-Host "❌ Erro: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Ver estatísticas após segunda mensagem
Write-Host "5️⃣ Estatísticas APÓS segunda mensagem (deve reusar sessão):" -ForegroundColor Green
Get-SessionStats

# Enviar mensagem de um novo cliente
Write-Host "6️⃣ Enviando mensagem do Cliente 2 (nova sessão)..." -ForegroundColor Green
$body3 = @{
    object = "whatsapp_business_account"
    entry = @(
        @{
            id = "1608476237088134"
            changes = @(
                @{
                    value = @{
                        messaging_product = "whatsapp"
                        metadata = @{
                            display_phone_number = "551150289262"
                            phone_number_id = "1050528654811352"
                        }
                        contacts = @(
                            @{
                                profile = @{ name = "Cliente 2" }
                                wa_id = "5511999999999"
                            }
                        )
                        messages = @(
                            @{
                                from = "5511999999999"
                                id = "msg3"
                                timestamp = "1774019274"
                                text = @{ body = "Oi, sou outro cliente" }
                                type = "text"
                            }
                        )
                    }
                    field = "messages"
                }
            )
        }
    )
} | ConvertTo-Json -Depth 10

try {
    Invoke-RestMethod -Uri "http://localhost:8080/webhook/meta" -Method Post -Body $body3 -ContentType "application/json" | Out-Null
    Write-Host "✅ Mensagem enviada" -ForegroundColor Green
} catch {
    Write-Host "❌ Erro: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Ver estatísticas finais
Write-Host "7️⃣ Estatísticas FINAIS (deve ter 2 sessões):" -ForegroundColor Green
Get-SessionStats

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "✅ Teste Concluído!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
