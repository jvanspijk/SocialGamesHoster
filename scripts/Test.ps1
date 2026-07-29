[CmdletBinding()]
param(
    [switch]$SkipBrowserJourney
)

$ErrorActionPreference = "Stop"
$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$embeddedRoot = Join-Path $projectRoot "Host\embedded\web"
$frontendDependencyInstaller = Join-Path $PSScriptRoot "Install-FrontendDependencies.ps1"
. $frontendDependencyInstaller

function Assert-NativeSuccess([string]$Step, [string]$Remediation = "") {
    if ($LASTEXITCODE -ne 0) {
        $message = "$Step failed with exit code $LASTEXITCODE."
        if ($Remediation) {
            $message += "`n$Remediation"
        }
        throw $message
    }
}

function Invoke-Check([string]$Step, [scriptblock]$Action, [string]$Remediation = "") {
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    & $Action
    Assert-NativeSuccess $Step $Remediation
    $stopwatch.Stop()
    Write-Host "$Step completed in $($stopwatch.Elapsed.ToString('m\\:ss'))."
}

Push-Location $projectRoot
try {
    $previousGoCache = $env:GOCACHE
    $env:GOCACHE = Join-Path $projectRoot ".tmp\go-build-cache"
    try {
        Invoke-Check "Go tests" { go test ./Host/... }
        Invoke-Check "Go vet" { go vet ./Host/... }
    }
    finally {
        $env:GOCACHE = $previousGoCache
    }
}
finally {
    Pop-Location
}

Push-Location (Join-Path $projectRoot "Web")
try {
    Install-FrontendDependencies -WebRoot (Get-Location).Path
    Invoke-Check "Frontend type checks" { npm run check }
    Invoke-Check "Frontend contract tests" { npm run test:unit }
    Invoke-Check "Frontend formatting check" { npm run format:check } 'Run `npm run format` from the Web directory, commit the resulting files, then rerun the check.'
    Invoke-Check "Frontend ESLint check" { npm run lint:eslint } 'Run `npm run lint:eslint` from the Web directory and fix the reported lint errors.'
    Invoke-Check "Frontend build" { npm run build }
    Invoke-Check "Production dependency audit" { npm audit --omit=dev --audit-level=high }

    if (-not $embeddedRoot.StartsWith($projectRoot + [IO.Path]::DirectorySeparatorChar)) {
        throw "The embedded web target is outside the project."
    }
    Get-ChildItem -LiteralPath $embeddedRoot -Force |
        Where-Object Name -ne ".gitkeep" |
        Remove-Item -Recurse -Force
    Copy-Item -Path (Join-Path (Get-Location) "build\*") -Destination $embeddedRoot -Recurse -Force

    if ($SkipBrowserJourney) {
        Write-Host "Browser journey skipped."
    }
    else {
        Invoke-Check "Browser journey" { npm run test:e2e }
    }
}
finally {
    Pop-Location
}
