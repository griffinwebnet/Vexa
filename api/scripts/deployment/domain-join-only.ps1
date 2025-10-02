# Vexa Domain Join Only
# This script joins the computer to the domain without Tailscale

param(
    [Parameter(Mandatory=$true)]
    [string]$DomainName,
    
    [Parameter(Mandatory=$false)]
    [string]$ComputerName = $env:COMPUTERNAME
)

Write-Host "🚀 Starting Vexa Domain Join..." -ForegroundColor Green

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
    # Step 1: Join Domain
    Write-Host "🏢 Joining domain $DomainName..." -ForegroundColor Yellow
    
    # Add computer to domain
    $securePassword = Read-Host "Enter domain administrator password" -AsSecureString
    $domainCredential = New-Object System.Management.Automation.PSCredential("administrator@$DomainName", $securePassword)
    
    Add-Computer -DomainName $DomainName -Credential $domainCredential -ComputerName $ComputerName -Restart:$false
    
    Write-Host "✅ Successfully joined domain!" -ForegroundColor Green
    
    # Step 2: Verify connection
    Write-Host "🔍 Verifying domain membership..." -ForegroundColor Yellow
    
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
}
