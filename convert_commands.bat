@echo off
REM Batch script to convert exec.Command calls to SafeCommand calls
REM This is a bulk conversion script for the Vexa codebase on Windows

echo Converting exec.Command calls to SafeCommand...

REM Function to convert a single file
:convert_file
set file=%1
echo Converting %file%...

REM Add utils import if not present
findstr /C:"github.com/griffinwebnet/vexa/api/utils" "%file%" >nul
if errorlevel 1 (
    REM Find the import block and add utils import
    powershell -Command "(Get-Content '%file%') -replace '(import \(\s*\n)', '$1	\"github.com/griffinwebnet/vexa/api/utils\"`n' | Set-Content '%file%'"
)

REM Convert exec.Command to utils.SafeCommand
powershell -Command "(Get-Content '%file%') -replace 'exec\.Command\(', 'utils.SafeCommand(' | Set-Content '%file%'"

REM Convert exec.CommandContext to utils.SafeCommandContext  
powershell -Command "(Get-Content '%file%') -replace 'exec\.CommandContext\(', 'utils.SafeCommandContext(' | Set-Content '%file%'"

echo Converted %file%
goto :eof

REM Convert all Go files in the api directory
for /r api %%f in (*.go) do (
    REM Skip the command.go file itself
    if not "%%f"=="api\utils\command.go" (
        call :convert_file "%%f"
    )
)

echo Conversion complete!
echo Note: You may need to fix variable name conflicts and error handling manually.
pause
