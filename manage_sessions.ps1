# Script para gerenciar sessões do Watson

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "🔧 GERENCIAMENTO DE SESSÕES WATSON" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080"
$statsUrl = "$baseUrl/watsonx/sessions/stats"
$resetUrl = "$baseUrl/watsonx/sessions/reset"

function Show-Menu {
    Write-Host ""
    Write-Host "Escolha uma opção:" -ForegroundColor Yellow
    Write-Host "1. 📊 Ver estatísticas de sessões"
    Write-Host "2. 🗑️  Resetar todas as sessões"
    Write-Host "3. 🔄 Ver stats → Resetar → Ver stats"
    Write-Host "4. ❌ Sair"
    Write-Host ""
}

function Get-SessionStats {
    Write-Host "📊 Buscando estatísticas..." -ForegroundColor Cyan
    try {
        $stats = Invoke-RestMethod -Uri $statsUrl -Method Get
        Write-Host ""
        Write-Host "Resultado:" -ForegroundColor Green
        $stats | ConvertTo-Json -Depth 5
        Write-Host ""
    } catch {
        Write-Host "❌ Erro: $($_.Exception.Message)" -ForegroundColor Red
    }
}

function Reset-Sessions {
    Write-Host "🗑️  Resetando todas as sessões..." -ForegroundColor Yellow
    try {
        $result = Invoke-RestMethod -Uri $resetUrl -Method Delete
        Write-Host ""
        Write-Host "Resultado:" -ForegroundColor Green
        $result | ConvertTo-Json
        Write-Host ""
        
        if ($result.status -eq "ok") {
            Write-Host "✅ Sessões resetadas com sucesso!" -ForegroundColor Green
            Write-Host "   Total deletado: $($result.deleted)" -ForegroundColor White
        }
    } catch {
        Write-Host "❌ Erro: $($_.Exception.Message)" -ForegroundColor Red
    }
}

function Test-ResetFlow {
    Write-Host "🔄 Executando fluxo completo..." -ForegroundColor Cyan
    Write-Host ""
    
    Write-Host "1️⃣ ANTES do reset:" -ForegroundColor Yellow
    Get-SessionStats
    
    Start-Sleep -Seconds 2
    
    Write-Host "2️⃣ RESETANDO:" -ForegroundColor Yellow
    Reset-Sessions
    
    Start-Sleep -Seconds 2
    
    Write-Host "3️⃣ DEPOIS do reset:" -ForegroundColor Yellow
    Get-SessionStats
}

# Loop principal
while ($true) {
    Show-Menu
    $choice = Read-Host "Digite sua opção"
    
    switch ($choice) {
        "1" {
            Get-SessionStats
        }
        "2" {
            Write-Host ""
            $confirm = Read-Host "Tem certeza que deseja resetar TODAS as sessões? (S/N)"
            if ($confirm -eq "S" -or $confirm -eq "s") {
                Reset-Sessions
            } else {
                Write-Host "❌ Operação cancelada" -ForegroundColor Yellow
            }
        }
        "3" {
            Test-ResetFlow
        }
        "4" {
            Write-Host ""
            Write-Host "👋 Até logo!" -ForegroundColor Cyan
            Write-Host ""
            exit
        }
        default {
            Write-Host "❌ Opção inválida. Tente novamente." -ForegroundColor Red
        }
    }
}
