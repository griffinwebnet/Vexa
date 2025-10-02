@echo off
REM Development Setup Script for Vexa API
REM This script sets up a minimal Samba AD DC environment for development

echo 🚀 Setting up Vexa API development environment...

REM Check if samba-tool is available
where samba-tool >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ samba-tool not found. Please install Samba first:
    echo    Windows: Install WSL2 with Ubuntu and run: sudo apt install samba samba-tools
    echo    Or use Docker with a Linux container
    pause
    exit /b 1
)

REM Check if domain is already provisioned
samba-tool domain info 127.0.0.1 >nul 2>nul
if %errorlevel% equ 0 (
    echo ✅ Domain already provisioned
) else (
    echo 📋 No domain found. You'll need to provision one using the API:
    echo    POST /api/v1/domain/provision
    echo    {
    echo      "domain": "DEV",
    echo      "realm": "dev.local",
    echo      "admin_password": "DevPass123!"
    echo    }
)

REM Set development environment
set ENV=development

echo ✅ Development environment configured
echo.
echo 🎯 Next steps:
echo    1. Start the API server: go run main.go
echo    2. Access the web interface at http://localhost:3000
echo    3. Login with admin/DevPass123! (after domain is provisioned)
echo.
echo 💡 The API will now use real Samba commands instead of dummy data
pause
