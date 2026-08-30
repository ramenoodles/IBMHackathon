# Deploy Onbober to Fly.io (backend + frontend).
# Prerequisites: fly CLI installed, `fly auth login`, Watsonx secrets on API app.
#
# Usage (from repo root):
#   fly auth login
#   fly secrets set WATSONX_API_KEY=... WATSONX_PROJECT_ID=... -a onbober-api
#   .\scripts\deploy-fly.ps1

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$ApiApp = "onbober-api"
$WebApp = "onbober-web"

function Require-Fly {
    if (-not (Get-Command fly -ErrorAction SilentlyContinue)) {
        Write-Host "Install flyctl: https://fly.io/docs/hands-on/install-flyctl/" -ForegroundColor Red
        exit 1
    }
    $whoami = fly auth whoami 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Not logged in. Run: fly auth login" -ForegroundColor Yellow
        exit 1
    }
    Write-Host "Fly.io account: $whoami"
}

function Ensure-App($Name) {
    fly status -a $Name 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Creating app $Name..." -ForegroundColor Yellow
        fly apps create $Name --org personal
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    }
}

function Deploy-Backend {
    Write-Host "`n=== Deploying API ($ApiApp) ===" -ForegroundColor Cyan
    Push-Location (Join-Path $Root "backend")
    try {
        fly deploy --ha=false
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    } finally {
        Pop-Location
    }
}

function Deploy-Frontend {
    Write-Host "`n=== Deploying web ($WebApp) ===" -ForegroundColor Cyan
    Push-Location (Join-Path $Root "frontend")
    try {
        fly deploy --ha=false
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    } finally {
        Pop-Location
    }
}

function Show-SecretsHint {
    $secrets = fly secrets list -a $ApiApp 2>&1
    if ($LASTEXITCODE -ne 0 -or $secrets -notmatch "WATSONX_API_KEY") {
        Write-Host "`nWarning: WATSONX_API_KEY may not be set on $ApiApp." -ForegroundColor Yellow
        Write-Host "AI labels/explain will be disabled until you run:" -ForegroundColor Yellow
        Write-Host "  fly secrets set WATSONX_API_KEY=... WATSONX_PROJECT_ID=... WATSONX_MODEL=ibm/granite-4-h-small -a $ApiApp" -ForegroundColor White
    }
}

Require-Fly
Ensure-App $ApiApp
Ensure-App $WebApp
Show-SecretsHint
Deploy-Backend
Deploy-Frontend

Write-Host "`n=== Done ===" -ForegroundColor Green
Write-Host "Demo URL: https://$WebApp.fly.dev"
Write-Host "API:      https://$ApiApp.fly.dev"
Write-Host "Logs:     fly logs -a $WebApp"
