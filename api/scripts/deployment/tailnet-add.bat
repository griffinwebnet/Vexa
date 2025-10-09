@echo off
setlocal enabledelayedexpansion

:: ========================================
:: Vexa Tailnet Add
:: ========================================
:: This script adds an existing computer to Tailnet

:: INJECTED VALUES - These will be replaced by the API
set LOGIN_SERVER={{LOGIN_SERVER}}
set AUTH_KEY={{AUTH_KEY}}

:: Check for admin privileges and elevate if needed
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo This script requires Administrator privileges.
    echo Elevating with UAC prompt...
    powershell -Command "Start-Process '%~f0' -Verb RunAs"
    exit /b 0
)

echo ========================================
echo Vexa Tailnet Add
echo ========================================
echo.

:: Get current hostname
set currentHostname=%COMPUTERNAME%
echo Current hostname: %currentHostname%
echo.

:: Step 1: Download and Install Tailscale
echo ========================================
echo Step 1: Installing Tailscale
echo ========================================
echo.

set TAILSCALE_URL=https://pkgs.tailscale.com/stable/tailscale-setup-latest-amd64.msi
set TAILSCALE_INSTALLER=%TEMP%\tailscale-setup.msi

echo Downloading Tailscale from %TAILSCALE_URL%...
powershell -Command "(New-Object Net.WebClient).DownloadFile('%TAILSCALE_URL%', '%TAILSCALE_INSTALLER%')"
if %errorLevel% neq 0 (
    echo ERROR: Failed to download Tailscale.
    pause
    exit /b 1
)

echo Installing Tailscale...
msiexec /i "%TAILSCALE_INSTALLER%" /quiet /norestart
if %errorLevel% neq 0 (
    echo ERROR: Failed to install Tailscale.
    pause
    exit /b 1
)

:: Wait for Tailscale service to start
echo Waiting for Tailscale service to start...
timeout /t 15 /nobreak >nul

echo Tailscale installed successfully.
echo.

:: Step 2: Connect to Tailnet
echo ========================================
echo Step 2: Connecting to Tailnet
echo ========================================
echo.

echo Connecting to Tailnet with auth key...
"C:\Program Files\Tailscale\tailscale.exe" up --authkey=%AUTH_KEY% --login-server=%LOGIN_SERVER% --accept-routes --accept-dns=false --hostname=%currentHostname% --unattended
if %errorLevel% neq 0 (
    echo ERROR: Failed to connect to Tailnet.
    pause
    exit /b 1
)

echo Tailnet connection successful.
echo.

:: Step 3: Verify Connection
echo ========================================
echo Step 3: Verifying Connection
echo ========================================
echo.

echo Checking Tailscale status...
"C:\Program Files\Tailscale\tailscale.exe" status
if %errorLevel% equ 0 (
    echo Tailscale is running successfully.
) else (
    echo WARNING: Tailscale status check returned an error.
)
echo.

:: Check domain membership
echo Checking domain membership...
for /f "tokens=2 delims=:" %%i in ('systeminfo ^| findstr /C:"Domain"') do set domain=%%i
set domain=%domain:~1%
if "%domain%"=="WORKGROUP" (
    echo INFO: Computer is not domain-joined.
) else (
    echo Domain: %domain%
)
echo.

:: Cleanup
if exist "%TAILSCALE_INSTALLER%" (
    del /f /q "%TAILSCALE_INSTALLER%" >nul 2>&1
)

:: Success message
echo ========================================
echo SUCCESS!
echo ========================================
echo.
echo The computer has been connected to Tailnet.
echo You can now access this computer through the Tailscale network.
echo.
echo No reboot is required for Tailscale-only setup.
pause

