@echo off
setlocal EnableExtensions EnableDelayedExpansion

cd /d "%~dp0"

set "ASPNETCORE_ENVIRONMENT=Production"
set "DOTNET_ENVIRONMENT=Production"

if "%ConnectionStrings__DefaultConnection%"=="" (
    set "ConnectionStrings__DefaultConnection=Host=localhost;Database=social_games;Username=postgres;Password=postgres"
)

set "PG_BIN="
for %%V in (18 17 16 15 14 13) do (
    if exist "C:\Program Files\PostgreSQL\%%V\bin\psql.exe" (
        set "PG_BIN=C:\Program Files\PostgreSQL\%%V\bin"
        goto :found_pg
    )
)

:found_pg
if defined PG_BIN (
    set "PATH=%PATH%;%PG_BIN%"
)

for /f "tokens=2 delims=:" %%S in ('sc query state^= all ^| findstr /i "postgresql-x64"') do (
    set "PG_SERVICE=%%S"
    set "PG_SERVICE=!PG_SERVICE: =!"
    goto :service_found
)

:service_found
if defined PG_SERVICE (
    sc start "%PG_SERVICE%" >nul 2>nul
)

if not exist "API\API.csproj" (
    echo ERROR: API\API.csproj not found.
    exit /b 1
)

if not exist "Web\package.json" (
    echo ERROR: Web\package.json not found.
    exit /b 1
)

if not exist "Web\node_modules" (
    echo Installing web dependencies...
    pushd "Web"
    call npm install
    if errorlevel 1 (
        popd
        echo ERROR: npm install failed.
        exit /b 1
    )
    popd
)

echo Starting API and Web...
start "SocialGamesHoster API" cmd /k "cd /d ""%~dp0API"" && set ASPNETCORE_ENVIRONMENT=%ASPNETCORE_ENVIRONMENT% && set DOTNET_ENVIRONMENT=%DOTNET_ENVIRONMENT% && set ConnectionStrings__DefaultConnection=%ConnectionStrings__DefaultConnection% && dotnet run"
start "SocialGamesHoster Web" cmd /k "cd /d ""%~dp0Web"" && npm run dev"

echo.
echo API: http://localhost:9090
echo Web: http://localhost:9091
echo.
echo Note: API is started with ASPNETCORE_ENVIRONMENT=Production to force PostgreSQL usage.
exit /b 0
