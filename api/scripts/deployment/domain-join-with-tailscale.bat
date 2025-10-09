@echo off
setlocal enabledelayedexpansion

:: ========================================
:: Vexa Domain Join with Tailscale
:: ========================================
:: This script will:
:: 1. Install Tailscale
:: 2. Connect to the Tailnet
:: 3. Join the domain
:: 4. Reboot

:: INJECTED VALUES - These will be replaced by the API
set DOMAIN_NAME={{DOMAIN_NAME}}
set DOMAIN_REALM={{DOMAIN_REALM}}
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
echo Vexa Domain Join with Tailscale
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

:: Wait for network stabilization
echo Waiting for network to stabilize...
timeout /t 5 /nobreak >nul

:: Step 3: Test Domain Resolution
echo ========================================
echo Step 3: Testing Domain Resolution
echo ========================================
echo.

echo Flushing DNS cache...
ipconfig /flushdns >nul

echo Testing domain resolution for %DOMAIN_REALM%...
nslookup %DOMAIN_REALM% >nul 2>&1
if %errorLevel% neq 0 (
    echo WARNING: Domain resolution failed for %DOMAIN_REALM%
    echo This may cause domain join to fail.
    echo Continuing anyway...
)

:: Check if it resolves to the expected IP
for /f "tokens=2" %%i in ('nslookup %DOMAIN_REALM% 2^>nul ^| findstr "Address:"') do set resolvedIP=%%i
if "%resolvedIP%"=="100.64.0.1" (
    echo Domain resolution successful: %DOMAIN_REALM% -^> %resolvedIP%
) else (
    echo Domain resolved to: %resolvedIP%
)
echo.

:: Step 4: Join Domain
echo ========================================
echo Step 4: Joining Domain
echo ========================================
echo.

echo Joining domain %DOMAIN_REALM%...
echo You will be prompted for domain administrator credentials.
set /p adminUser="Enter domain admin username (e.g., administrator): "
set /p adminPassword="Enter domain admin password: "

powershell -Command "$cred = New-Object System.Management.Automation.PSCredential('%adminUser%', (ConvertTo-SecureString '%adminPassword%' -AsPlainText -Force)); Add-Computer -DomainName '%DOMAIN_REALM%' -Credential $cred -Force"
if %errorLevel% neq 0 (
    echo ERROR: Failed to join domain.
    echo Please check credentials and network connectivity.
    pause
    exit /b 1
)

echo Successfully joined domain %DOMAIN_REALM%.
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
echo The computer has been:
echo   [X] Connected to Tailnet
echo   [X] Joined to domain %DOMAIN_NAME%
echo.
echo The system will now reboot to complete the setup.
echo.
pause

echo Rebooting in 5 seconds...
timeout /t 5 /nobreak >nul
shutdown /r /f /t 0

