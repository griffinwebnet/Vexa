# Vexa Domain Join with Tailscale
# This script downloads and installs Tailscale, joins the domain, and connects to Tailnet

param(
    [Parameter(Mandatory=$true)]
    [string]$DomainController,
    
    [Parameter(Mandatory=$true)]
    [string]$DomainName,
    
    [Parameter(Mandatory=$false)]
    [string]$TailscaleAuthKey,
    
    [Parameter(Mandatory=$false)]
    [string]$ComputerName = $env:COMPUTERNAME
)

Write-Host "🚀 Starting Vexa Domain Join with Tailscale..." -ForegroundColor Green

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
    
    # Step 2: Connect to Tailnet (if auth key provided)
    if ($TailscaleAuthKey) {
        Write-Host "🔗 Connecting to Tailnet..." -ForegroundColor Yellow
        & "C:\Program Files\Tailscale\tailscale.exe" up --authkey $TailscaleAuthKey --login-server "{{LOGIN_SERVER}}" --accept-routes --accept-dns=false --hostname $ComputerName
    } else {
        Write-Host "🔗 Please complete Tailscale login manually..." -ForegroundColor Yellow
        & "C:\Program Files\Tailscale\tailscale.exe" login
    }
    
    # Step 3: Join Domain
    Write-Host "🏢 Joining domain $DomainName..." -ForegroundColor Yellow
    
    # Add computer to domain
    $securePassword = Read-Host "Enter domain administrator password" -AsSecureString
    $domainCredential = New-Object System.Management.Automation.PSCredential("administrator@$DomainName", $securePassword)
    
    Add-Computer -DomainName $DomainName -Credential $domainCredential -ComputerName $ComputerName -Restart:$false
    
    Write-Host "✅ Successfully joined domain!" -ForegroundColor Green
    
    # Step 4: Verify connection
    Write-Host "🔍 Verifying connections..." -ForegroundColor Yellow
    
    # Check Tailscale status
    $tailscaleStatus = & "C:\Program Files\Tailscale\tailscale.exe" status --json | ConvertFrom-Json
    if ($tailscaleStatus.BackendState -eq "Running") {
        Write-Host "✅ Tailscale connected successfully" -ForegroundColor Green
    } else {
        Write-Warning "⚠️ Tailscale connection may need attention"
    }
    
    # Check domain membership
    $domainInfo = Get-ComputerInfo | Select-Object -ExpandProperty CsDomain
    if ($domainInfo -eq $DomainName) {
        Write-Host "✅ Domain membership verified" -ForegroundColor Green
    } else {
        Write-Warning "⚠️ Domain membership verification failed"
    }
    
    Write-Host "🎉 Setup complete! Restart required to finalize domain join." -ForegroundColor Green
    Write-Host "💡 The computer will restart in 30 seconds..." -ForegroundColor Cyan
    
    # Optional restart
    $restart = Read-Host "Restart now? (Y/N)"
    if ($restart -eq "Y" -or $restart -eq "y") {
        Restart-Computer -Force
    }
    
} catch {
    Write-Error "❌ Setup failed: $($_.Exception.Message)"
    exit 1
} finally {
    # Cleanup
    if (Test-Path $tailscaleInstaller) {
        Remove-Item $tailscaleInstaller -Force
    }
}
