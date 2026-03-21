@echo off
setlocal EnableExtensions EnableDelayedExpansion

cd /d "%~dp0"

echo Checking for winget...
where winget >nul 2>nul
if errorlevel 1 (
    echo ERROR: winget is not installed or not in PATH.
    echo Install App Installer from Microsoft Store, then rerun this script.
    exit /b 1
)

echo.
echo Installing PostgreSQL...
winget install --id=PostgreSQL.PostgreSQL.18 --exact --silent --accept-package-agreements --accept-source-agreements
if errorlevel 1 (
    echo WARNING: PostgreSQL install may have failed or is already installed.
)

set "PG_BIN="
for %%V in (18 17 16 15 14 13) do (
    if exist "C:\Program Files\PostgreSQL\%%V\bin\psql.exe" (
        set "PG_BIN=C:\Program Files\PostgreSQL\%%V\bin"
        goto :found_pg
    )
)

:found_pg
if not defined PG_BIN (
    echo ERROR: Could not find PostgreSQL bin folder automatically.
    echo Check under C:\Program Files\PostgreSQL\{version}\bin
    exit /b 1
)

set "PATH=%PATH%;%PG_BIN%"

for /f "tokens=2 delims=:" %%S in ('sc query state^= all ^| findstr /i "postgresql-x64"') do (
    set "PG_SERVICE=%%S"
    set "PG_SERVICE=!PG_SERVICE: =!"
    goto :service_found
)

:service_found
if defined PG_SERVICE (
    echo.
    echo Ensuring PostgreSQL service is running: %PG_SERVICE%
    sc start "%PG_SERVICE%" >nul 2>nul
) else (
    echo.
    echo WARNING: PostgreSQL Windows service not found automatically.
)

set "PGPASSWORD=postgres"
echo.
echo Trying to create database social_games (best effort)...
psql -U postgres -h localhost -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='social_games'" | find "1" >nul 2>nul
if errorlevel 1 (
    psql -U postgres -h localhost -d postgres -c "CREATE DATABASE social_games;" >nul 2>nul
    if errorlevel 1 (
        echo WARNING: Could not create database automatically.
        echo If needed, run this manually:
        echo   psql -U postgres -h localhost -d postgres -c "CREATE DATABASE social_games;"
    ) else (
        echo Database social_games created.
    )
) else (
    echo Database social_games already exists.
)

echo.
echo Installing .NET 10 SDK...
winget install --id Microsoft.DotNet.SDK.10 --exact --silent --accept-package-agreements --accept-source-agreements
if errorlevel 1 (
    echo WARNING: .NET 10 SDK install may have failed or is already installed.
)

echo.
echo Installing Node.js LTS...
winget install --id OpenJS.NodeJS.LTS --exact --silent --accept-package-agreements --accept-source-agreements
if errorlevel 1 (
    echo WARNING: Node.js install may have failed or is already installed.
)

echo.
echo Installed dependencies:
echo - .NET 10 SDK
echo - Node.js LTS
echo - PostgreSQL
echo.
echo PostgreSQL binaries found at:
echo %PG_BIN%
echo.
echo Use run.bat to start API + web.
exit /b 0
