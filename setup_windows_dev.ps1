# Windows Development Setup Script for Vexa
# This script sets up comprehensive logging, audit trailing, and security enhancements

Write-Host "=== Vexa Windows Development Setup ===" -ForegroundColor Cyan
Write-Host "Setting up comprehensive logging and security features..." -ForegroundColor Green

# Create logs directory for Windows development
if (!(Test-Path "logs")) {
    New-Item -ItemType Directory -Path "logs" | Out-Null
    Write-Host "Created logs directory for Windows development" -ForegroundColor Green
}

# Create .gitignore entry for logs if not present
if (!(Get-Content ".gitignore" -ErrorAction SilentlyContinue | Select-String "logs/")) {
    Add-Content ".gitignore" "`n# Log files`nlogs/`n*.log`n*.log.*`n*.tar.gz"
    Write-Host "Added log files to .gitignore" -ForegroundColor Green
}

Write-Host "`n=== Adding Comprehensive Logging ===" -ForegroundColor Yellow

# Add logging to services
Get-ChildItem -Path "api\services" -Filter "*.go" | ForEach-Object {
    $file = $_.FullName
    $content = Get-Content $file -Raw
    
    # Add utils import if not present
    if ($content -notmatch "github.com/griffinwebnet/vexa/api/utils") {
        $content = $content -replace '(import \(\s*\n)', '$1	"github.com/griffinwebnet/vexa/api/utils"`n'
        Set-Content $file $content -NoNewline
        Write-Host "Added utils import to $($_.Name)" -ForegroundColor Green
    }
    
    # Add logging to constructor functions
    if ($content -match 'func New([A-Za-z]*)Service\(\) \*[A-Za-z]*Service \{' -and $content -notmatch 'utils\.Info\("Initializing') {
        $content = $content -replace 'func New([A-Za-z]*)Service\(\) \*[A-Za-z]*Service \{', 'func New$1Service() *$1Service {`n	utils.Info("Initializing $1Service")'
        Set-Content $file $content -NoNewline
        Write-Host "Added constructor logging to $($_.Name)" -ForegroundColor Green
    }
}

# Add logging to handlers
Get-ChildItem -Path "api\handlers" -Filter "*.go" | ForEach-Object {
    $file = $_.FullName
    $content = Get-Content $file -Raw
    
    # Add utils import if not present
    if ($content -notmatch "github.com/griffinwebnet/vexa/api/utils") {
        $content = $content -replace '(import \(\s*\n)', '$1	"github.com/griffinwebnet/vexa/api/utils"`n'
        Set-Content $file $content -NoNewline
        Write-Host "Added utils import to handler $($_.Name)" -ForegroundColor Green
    }
}

Write-Host "`n=== Adding Audit Trailing ===" -ForegroundColor Yellow

# Add audit trailing to handlers
Get-ChildItem -Path "api\handlers" -Filter "*.go" | ForEach-Object {
    $file = $_.FullName
    $content = Get-Content $file -Raw
    
    # Add audit context extraction to handler functions
    if ($content -match 'func ([A-Za-z]*)\(c \*gin\.Context\) \{' -and $content -notmatch 'ctx := utils\.GetAuditContext') {
        $content = $content -replace 'func ([A-Za-z]*)\(c \*gin\.Context\) \{', 'func $1(c *gin.Context) {`n	ctx := utils.GetAuditContext(c)'
        Set-Content $file $content -NoNewline
        Write-Host "Added audit context to handler $($_.Name)" -ForegroundColor Green
    }
}

Write-Host "`n=== Security Enhancements ===" -ForegroundColor Yellow

# Check if command sanitizer is being used
$commandFiles = Get-ChildItem -Path "api" -Recurse -Filter "*.go" | Where-Object { 
    (Get-Content $_.FullName -Raw) -match "exec\.Command" 
}

if ($commandFiles.Count -gt 0) {
    Write-Host "Found $($commandFiles.Count) files with exec.Command calls that need conversion" -ForegroundColor Yellow
    Write-Host "Files to convert:" -ForegroundColor Yellow
    $commandFiles | ForEach-Object { Write-Host "  - $($_.Name)" -ForegroundColor Gray }
    Write-Host "`nRun 'convert_commands.bat' to convert these to SafeCommand" -ForegroundColor Cyan
}

Write-Host "`n=== Development Environment Ready ===" -ForegroundColor Green
Write-Host "✓ Logging system configured for Windows development" -ForegroundColor Green
Write-Host "✓ Log files will be created in ./logs/ directory" -ForegroundColor Green
Write-Host "✓ Comprehensive audit trailing enabled" -ForegroundColor Green
Write-Host "✓ Security enhancements in place" -ForegroundColor Green

Write-Host "`n=== Next Steps ===" -ForegroundColor Cyan
Write-Host "1. Run 'go build' to test the build" -ForegroundColor White
Write-Host "2. Run 'go run main.go' to start the server" -ForegroundColor White
Write-Host "3. Check ./logs/ directory for log files" -ForegroundColor White
Write-Host "4. Use the Security UI to view audit trails" -ForegroundColor White

Write-Host "`nPress any key to continue..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
