# Script para testar o webhook da Meta/WhatsApp Business API

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "🧪 TESTE DO WEBHOOK META" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# URL do endpoint
$url = "http://localhost:8080/webhook/meta"

Write-Host "Endpoint: $url" -ForegroundColor Yellow
Write-Host "" 

# Payload no formato da Meta (baseado no msg_meta.txt)
$body = @{
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
                                profile = @{
                                    name = "Daniel Silveira"
                                }
                                wa_id = "5511998648847"
                            }
                        )
                        messages = @(
                            @{
                                from = "5511998648847"
                                id = "wamid.HBgNNTUxMTk5ODY0ODg0NxUCABIYIEFDOTJEM0JBMjE3RUZDQUNCREI4N0Q4OUVGODlENjM5AA=="
                                timestamp = "1774019272"
                                text = @{
                                    body = "Olá, preciso de ajuda"
                                }
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

Write-Host "Enviando mensagem para o webhook da Meta..." -ForegroundColor Cyan
Write-Host "URL: $url" -ForegroundColor Yellow

try {
    $response = Invoke-RestMethod -Uri $url -Method Post -Body $body -ContentType "application/json"
    Write-Host "`nResposta:" -ForegroundColor Green
    $response | ConvertTo-Json
} catch {
    Write-Host "`nErro:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    if ($_.ErrorDetails.Message) {
        Write-Host $_.ErrorDetails.Message
    }
}
