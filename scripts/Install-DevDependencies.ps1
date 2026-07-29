[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"
$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$frontendDependencyInstaller = Join-Path $PSScriptRoot "Install-FrontendDependencies.ps1"
$gitHooksInstaller = Join-Path $PSScriptRoot "Install-GitHooks.ps1"
. $frontendDependencyInstaller

function Assert-NativeSuccess([string]$Step) {
    if ($LASTEXITCODE -ne 0) {
        throw "$Step failed with exit code $LASTEXITCODE."
    }
}

Push-Location $projectRoot
try {
    go mod download
    Assert-NativeSuccess "Go dependency installation"
}
finally {
    Pop-Location
}

Push-Location (Join-Path $projectRoot "Web")
try {
    Install-FrontendDependencies -WebRoot (Get-Location).Path
    $previousBrowserPath = $env:PLAYWRIGHT_BROWSERS_PATH
    try {
        $env:PLAYWRIGHT_BROWSERS_PATH = Join-Path (Get-Location) ".playwright-browsers"
        npx playwright install chromium
        Assert-NativeSuccess "Chromium installation"
    }
    finally {
        $env:PLAYWRIGHT_BROWSERS_PATH = $previousBrowserPath
    }
}
finally {
    Pop-Location
}

& $gitHooksInstaller

Write-Host "Developer dependencies are ready."
Write-Host "Inno Setup 6 is additionally required only when building the installer."
