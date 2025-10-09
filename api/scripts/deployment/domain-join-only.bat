@echo off
setlocal enabledelayedexpansion

:: ========================================
:: Vexa Domain Join Only
:: ========================================
:: This script will join the domain without installing Tailscale

:: INJECTED VALUES - These will be replaced by the API
set DOMAIN_NAME={{DOMAIN_NAME}}
set DOMAIN_REALM={{DOMAIN_REALM}}
set ADMIN_USER={{ADMIN_USER}}
set ADMIN_PASSWORD={{ADMIN_PASSWORD}}

:: Check for admin privileges
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: This script requires Administrator privileges.
    echo Please run Command Prompt as Administrator.
    pause
    exit /b 1
)

echo ========================================
echo Vexa Domain Join
echo ========================================
echo.

:: Get current hostname
set currentHostname=%COMPUTERNAME%
echo Current hostname: %currentHostname%
echo Target domain: %DOMAIN_REALM%
echo.

:: Test Domain Resolution
echo ========================================
echo Testing Domain Resolution
echo ========================================
echo.

echo Flushing DNS cache...
ipconfig /flushdns >nul

echo Testing domain resolution for %DOMAIN_REALM%...
nslookup %DOMAIN_REALM% >nul 2>&1
if %errorLevel% neq 0 (
    echo ERROR: Domain resolution failed for %DOMAIN_REALM%
    echo Please check your network configuration.
    pause
    exit /b 1
)

echo Domain resolution successful.
echo.

:: Join Domain
echo ========================================
echo Joining Domain
echo ========================================
echo.

echo Joining domain %DOMAIN_REALM%...
powershell -Command "$cred = New-Object System.Management.Automation.PSCredential('%ADMIN_USER%', (ConvertTo-SecureString '%ADMIN_PASSWORD%' -AsPlainText -Force)); Add-Computer -DomainName '%DOMAIN_REALM%' -Credential $cred -Force"
if %errorLevel% neq 0 (
    echo ERROR: Failed to join domain.
    echo Please check credentials and network connectivity.
    pause
    exit /b 1
)

echo Successfully joined domain %DOMAIN_REALM%.
echo.

:: Success message
echo ========================================
echo SUCCESS!
echo ========================================
echo.
echo The computer has been joined to domain %DOMAIN_NAME%.
echo The system will now reboot to complete the setup.
echo.
pause

echo Rebooting in 5 seconds...
timeout /t 5 /nobreak >nul
shutdown /r /f /t 0

