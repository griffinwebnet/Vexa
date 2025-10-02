# Vexa Tailnet Add
# This script adds an existing domain-joined computer to Tailnet

param(
    [Parameter(Mandatory=$false)]
    [string]$TailscaleAuthKey,
    
    [Parameter(Mandatory=$false)]
    [string]$ComputerName = $env:COMPUTERNAME
)

Write-Host "🚀 Adding computer to Tailnet..." -ForegroundColor Green

# Function to check if running as administrator
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Check if running as administrator
if (-not (Test-Administrator)) {
    Write-Error "❌ This script must be run as Administrator"
    exit 1
}

try {
    # Step 1: Download and Install Tailscale
    Write-Host "📥 Downloading Tailscale..." -ForegroundColor Yellow
    $tailscaleUrl = "https://pkgs.tailscale.com/stable/windows/tailscale-setup-latest-amd64.msi"
    $tailscaleInstaller = "$env:TEMP\tailscale-setup.msi"
    
    Invoke-WebRequest -Uri $tailscaleUrl -OutFile $tailscaleInstaller -UseBasicParsing
    
    Write-Host "🔧 Installing Tailscale..." -ForegroundColor Yellow
    Start-Process msiexec.exe -Wait -ArgumentList "/i $tailscaleInstaller /quiet /norestart"
    
    # Wait for Tailscale service to start
    Start-Sleep -Seconds 10
    
    # Step 2: Connect to Tailnet
    if ($TailscaleAuthKey) {
        Write-Host "🔗 Connecting to Tailnet with auth key..." -ForegroundColor Yellow
        & "C:\Program Files\Tailscale\tailscale.exe" login --authkey $TailscaleAuthKey
    } else {
        Write-Host "🔗 Please complete Tailscale login manually..." -ForegroundColor Yellow
        & "C:\Program Files\Tailscale\tailscale.exe" login
    }
    
    # Step 3: Verify connection
    Write-Host "🔍 Verifying Tailnet connection..." -ForegroundColor Yellow
    
    # Check Tailscale status
    $tailscaleStatus = & "C:\Program Files\Tailscale\tailscale.exe" status --json | ConvertFrom-Json
    if ($tailscaleStatus.BackendState -eq "Running") {
        Write-Host "✅ Successfully connected to Tailnet!" -ForegroundColor Green
        Write-Host "📍 Computer IP: $($tailscaleStatus.Self.TailscaleIPs[0])" -ForegroundColor Cyan
        Write-Host "🏷️  Computer name: $($tailscaleStatus.Self.HostName)" -ForegroundColor Cyan
    } else {
        Write-Warning "⚠️ Tailnet connection may need attention"
    }
    
    # Check domain membership
    $domainInfo = Get-ComputerInfo | Select-Object -ExpandProperty CsDomain
    if ($domainInfo -and $domainInfo -ne "WORKGROUP") {
        Write-Host "✅ Domain membership verified: $domainInfo" -ForegroundColor Green
    } else {
        Write-Warning "⚠️ Computer is not domain-joined"
    }
    
    Write-Host "🎉 Tailnet setup complete!" -ForegroundColor Green
    
} catch {
    Write-Error "❌ Setup failed: $($_.Exception.Message)"
    exit 1
} finally {
    # Cleanup
    if (Test-Path $tailscaleInstaller) {
        Remove-Item $tailscaleInstaller -Force
    }
}
