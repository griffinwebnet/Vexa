# PowerShell script to add comprehensive logging to all Go services
# This script adds utils import and logging statements to service functions

Write-Host "Adding comprehensive logging to Go services..." -ForegroundColor Green

# Function to add logging to a service file
function Add-LoggingToService {
    param($file)
    Write-Host "Adding logging to $file..." -ForegroundColor Yellow
    
    $content = Get-Content $file -Raw
    
    # Add utils import if not present
    if ($content -notmatch "github.com/griffinwebnet/vexa/api/utils") {
        # Find the import block and add utils import
        $content = $content -replace '(import \(\s*\n)', '$1	"github.com/griffinwebnet/vexa/api/utils"`n'
    }
    
    # Add logging to service constructor functions
    $content = $content -replace 'func New([A-Za-z]*)Service\(\) \*[A-Za-z]*Service \{', 'func New$1Service() *$1Service {`n	utils.Info("Initializing $1Service")'
    
    # Add logging to main service functions
    $content = $content -replace 'func \([^)]*\) ([A-Za-z]*)\([^)]*\) error \{', 'func ($1) $2($3) error {`n	utils.Info("Executing $2")'
    
    # Add error logging to error returns
    $content = $content -replace 'return fmt\.Errorf\("([^"]*)", ([^)]*)\)', 'utils.Error("Failed: $1", $2)`n	return fmt.Errorf("$1", $2)'
    
    Set-Content $file $content -NoNewline
    Write-Host "Added logging to $file" -ForegroundColor Green
}

# Add logging to all service files
Get-ChildItem -Path "api\services" -Filter "*.go" | ForEach-Object {
    $file = $_.FullName
    # Skip files that already have extensive logging
    if ($file -notmatch "auth_service.go" -and 
        $file -notmatch "domain_service.go" -and 
        $file -notmatch "overlay_service.go") {
        Add-LoggingToService $file
    }
}

# Add logging to handlers
Get-ChildItem -Path "api\handlers" -Filter "*.go" | ForEach-Object {
    $file = $_.FullName
    Write-Host "Adding logging to handler $file..." -ForegroundColor Yellow
    
    $content = Get-Content $file -Raw
    
    # Add utils import if not present
    if ($content -notmatch "github.com/griffinwebnet/vexa/api/utils") {
        $content = $content -replace '(import \(\s*\n)', '$1	"github.com/griffinwebnet/vexa/api/utils"`n'
    }
    
    # Add logging to handler functions
    $content = $content -replace 'func ([A-Za-z]*)\(c \*gin\.Context\) \{', 'func $1(c *gin.Context) {`n	utils.Info("Handler called: $1")'
    
    Set-Content $file $content -NoNewline
}

Write-Host "Comprehensive logging enhancement complete!" -ForegroundColor Green
Write-Host "Note: You may need to review and adjust the generated logging statements." -ForegroundColor Yellow
