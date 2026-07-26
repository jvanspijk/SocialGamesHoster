[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"
$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$embeddedRoot = Join-Path $projectRoot "Host\embedded\web"

function Assert-NativeSuccess([string]$Step) {
    if ($LASTEXITCODE -ne 0) {
        throw "$Step failed with exit code $LASTEXITCODE."
    }
}

Push-Location $projectRoot
try {
    go test ./Host/...
    Assert-NativeSuccess "Go tests"
    go vet ./Host/...
    Assert-NativeSuccess "Go vet"
}
finally {
    Pop-Location
}

Push-Location (Join-Path $projectRoot "Web")
try {
    npm ci
    Assert-NativeSuccess "Frontend dependency installation"
    npm run check
    Assert-NativeSuccess "Frontend type checks"
    npm run test:unit
    Assert-NativeSuccess "Frontend contract tests"
    npm run lint
    Assert-NativeSuccess "Frontend lint"
    npm run build
    Assert-NativeSuccess "Frontend build"
    npm audit --omit=dev --audit-level=high
    Assert-NativeSuccess "Production dependency audit"

    if (-not $embeddedRoot.StartsWith($projectRoot + [IO.Path]::DirectorySeparatorChar)) {
        throw "The embedded web target is outside the project."
    }
    Get-ChildItem -LiteralPath $embeddedRoot -Force |
        Where-Object Name -ne ".gitkeep" |
        Remove-Item -Recurse -Force
    Copy-Item -Path (Join-Path (Get-Location) "build\*") -Destination $embeddedRoot -Recurse -Force

    npm run test:e2e
    Assert-NativeSuccess "Browser journey"
}
finally {
    Pop-Location
}
