# Script para testar integração Redis

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "🧪 TESTE DE INTEGRAÇÃO REDIS" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$statsUrl = "http://localhost:8080/watsonx/sessions/stats"
$metaUrl = "http://localhost:8080/webhook/meta"

# Função para ver estatísticas
function Get-RedisStats {
    param([string]$label)
    
    Write-Host "📊 $label" -ForegroundColor Yellow
    try {
        $stats = Invoke-RestMethod -Uri $statsUrl -Method Get
        $stats | ConvertTo-Json -Depth 5
        Write-Host ""
    } catch {
        Write-Host "❌ Erro ao buscar estatísticas: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Ver estatísticas iniciais do Redis
Get-RedisStats "Estatísticas INICIAIS (Redis deve estar vazio):"

# Enviar primeira mensagem
Write-Host "1️⃣ Enviando PRIMEIRA mensagem (deve criar nova sessão no Redis)..." -ForegroundColor Green
$msg1 = @{
    object = "whatsapp_business_account"
    entry = @(@{
        id = "123"
        changes = @(@{
            value = @{
                messaging_product = "whatsapp"
                metadata = @{
                    display_phone_number = "551150289262"
                    phone_number_id = "1050528654811352"
                }
                contacts = @(@{
                    profile = @{ name = "Teste Redis" }
                    wa_id = "5511998648847"
                })
                messages = @(@{
                    from = "5511998648847"
                    id = "msg1"
                    timestamp = "1774019272"
                    text = @{ body = "Primeira mensagem - teste Redis" }
                    type = "text"
                })
            }
            field = "messages"
        })
    })
} | ConvertTo-Json -Depth 10

try {
    Invoke-RestMethod -Uri $metaUrl -Method Post -Body $msg1 -ContentType "application/json" | Out-Null
    Write-Host "✅ Mensagem enviada - sessão criada no Redis" -ForegroundColor Green
} catch {
    Write-Host "❌ Erro: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 2
Get-RedisStats "Estatísticas APÓS primeira mensagem (deve ter 1 sessão):"

# Enviar segunda mensagem do mesmo cliente
Write-Host "2️⃣ Enviando SEGUNDA mensagem do MESMO cliente (deve REUSAR sessão do Redis)..." -ForegroundColor Green
$msg2 = @{
    object = "whatsapp_business_account"
    entry = @(@{
        id = "123"
        changes = @(@{
            value = @{
                messaging_product = "whatsapp"
                metadata = @{
                    display_phone_number = "551150289262"
                    phone_number_id = "1050528654811352"
                }
                contacts = @(@{
                    profile = @{ name = "Teste Redis" }
                    wa_id = "5511998648847"
                })
                messages = @(@{
                    from = "5511998648847"
                    id = "msg2"
                    timestamp = "1774019273"
                    text = @{ body = "Segunda mensagem - mesma sessão" }
                    type = "text"
                })
            }
            field = "messages"
        })
    })
} | ConvertTo-Json -Depth 10

try {
    Invoke-RestMethod -Uri $metaUrl -Method Post -Body $msg2 -ContentType "application/json" | Out-Null
    Write-Host "✅ Mensagem enviada - sessão reusada do Redis" -ForegroundColor Green
} catch {
    Write-Host "❌ Erro: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 2
Get-RedisStats "Estatísticas APÓS segunda mensagem (ainda 1 sessão, last_used atualizado):"

# Enviar mensagem de outro cliente
Write-Host "3️⃣ Enviando mensagem de OUTRO cliente (deve criar NOVA sessão no Redis)..." -ForegroundColor Green
$msg3 = @{
    object = "whatsapp_business_account"
    entry = @(@{
        id = "123"
        changes = @(@{
            value = @{
                messaging_product = "whatsapp"
                metadata = @{
                    display_phone_number = "551150289262"
                    phone_number_id = "1050528654811352"
                }
                contacts = @(@{
                    profile = @{ name = "Cliente 2" }
                    wa_id = "5511999999999"
                })
                messages = @(@{
                    from = "5511999999999"
                    id = "msg3"
                    timestamp = "1774019274"
                    text = @{ body = "Olá, sou outro cliente" }
                    type = "text"
                })
            }
            field = "messages"
        })
    })
} | ConvertTo-Json -Depth 10

try {
    Invoke-RestMethod -Uri $metaUrl -Method Post -Body $msg3 -ContentType "application/json" | Out-Null
    Write-Host "✅ Mensagem enviada - nova sessão criada no Redis" -ForegroundColor Green
} catch {
    Write-Host "❌ Erro: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 2
Get-RedisStats "Estatísticas FINAIS (deve ter 2 sessões no Redis):"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "✅ Teste Concluído!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "ℹ️  Verifique os logs do servidor para ver:" -ForegroundColor Cyan
Write-Host "   - ✅ Connected to Redis" -ForegroundColor White
Write-Host "   - 💾 Session saved to Redis" -ForegroundColor White
Write-Host "   - 📥 Session retrieved from Redis" -ForegroundColor White
Write-Host "   - ♻️  Reusing existing session from Redis" -ForegroundColor White
