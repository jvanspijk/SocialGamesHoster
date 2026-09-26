[CmdletBinding()]
param(
    [switch]$Diagnostics,
    [string]$DataDir
)

$ErrorActionPreference = "Stop"
$stopDevScript = Join-Path $PSScriptRoot "Stop-Dev.ps1"
& $stopDevScript

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$hostAddress = "127.0.0.1:8090"
$hostStatusUrl = "http://$hostAddress/api/app/v1/setup/status"
$arguments = @("run", "./Host/cmd/socialgameshoster", "--", "--no-tray", "--http=$hostAddress")
if ($Diagnostics) {
    $arguments += "--diagnostics"
}
if ($DataDir) {
    $arguments += "--dir=$DataDir"
}

$portProbe = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, 8090)
try {
    $portProbe.Start()
}
catch {
    throw "Port 8090 is already in use. Stop the other server before starting the dev environment."
}
finally {
    $portProbe.Stop()
}

$hostJob = Start-Job -ArgumentList $projectRoot, $arguments -ScriptBlock {
    param($root, [string[]]$goArguments)
    Set-Location -LiteralPath $root
    & go @goArguments
    if ($LASTEXITCODE -ne 0) {
        throw "Go host exited with code $LASTEXITCODE."
    }
}
try {
    Write-Host "Waiting for the Go host at $hostAddress..."
    $deadline = (Get-Date).AddMinutes(5)
    $hostReady = $false
    while ((Get-Date) -lt $deadline) {
        if ($hostJob.State -ne "Running") {
            $hostOutput = (Receive-Job -Job $hostJob -Keep *>&1 | Out-String).Trim()
            throw "The Go host stopped before it became ready. $hostOutput"
        }
        try {
            $response = Invoke-WebRequest -Uri $hostStatusUrl -UseBasicParsing -TimeoutSec 2
            if ($response.StatusCode -eq 200) {
                $hostReady = $true
                break
            }
        }
        catch {
            # A new build may still be compiling or applying migrations.
        }
        Start-Sleep -Milliseconds 500
    }
    if (-not $hostReady) {
        $hostOutput = (Receive-Job -Job $hostJob -Keep *>&1 | Out-String).Trim()
        throw "The Go host did not become ready at $hostStatusUrl within 5 minutes. $hostOutput"
    }

    Write-Host "Go host ready. Starting Vite on port 9091..."
    Push-Location (Join-Path $projectRoot "Web")
    try {
        npm run dev
        if ($LASTEXITCODE -ne 0) {
            throw "Vite exited with code $LASTEXITCODE."
        }
    }
    finally {
        Pop-Location
    }
}
finally {
    Stop-Job -Job $hostJob -ErrorAction SilentlyContinue
    Remove-Job -Job $hostJob -Force -ErrorAction SilentlyContinue
}
