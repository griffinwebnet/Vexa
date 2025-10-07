# PowerShell script to fix all import paths to use correct GitHub username
# This script updates all import statements to use github.com/griffinwebnet/vexa/api

Write-Host "Fixing import paths to use correct GitHub username..." -ForegroundColor Green

# Function to fix import paths in a file
function Fix-ImportPaths {
    param($file)
    
    $content = Get-Content $file -Raw
    
    # Fix the import paths
    $content = $content -replace 'github\.com/vexa/api', 'github.com/griffinwebnet/vexa/api'
    
    Set-Content $file $content -NoNewline
    Write-Host "Fixed import paths in $file" -ForegroundColor Green
}

# Fix all Go files in the api directory
Get-ChildItem -Path "api" -Recurse -Filter "*.go" | ForEach-Object {
    Fix-ImportPaths $_.FullName
}

# Fix the PowerShell scripts
Get-ChildItem -Path "." -Filter "*.ps1" | ForEach-Object {
    Fix-ImportPaths $_.FullName
}

# Fix the batch files
Get-ChildItem -Path "." -Filter "*.bat" | ForEach-Object {
    Fix-ImportPaths $_.FullName
}

Write-Host "All import paths have been fixed!" -ForegroundColor Green
Write-Host "Updated module path: github.com/griffinwebnet/vexa/api" -ForegroundColor Cyan
