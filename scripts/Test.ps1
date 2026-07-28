[CmdletBinding()]
param()

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
    Install-FrontendDependencies -WebRoot (Get-Location).Path
    npm run check
    Assert-NativeSuccess "Frontend type checks"
    npm run test:unit
    Assert-NativeSuccess "Frontend contract tests"
    npm run format:check
    Assert-NativeSuccess "Frontend formatting check" 'Run `npm run format` from the Web directory, commit the resulting files, then rerun the check.'
    npm run lint:eslint
    Assert-NativeSuccess "Frontend ESLint check" 'Run `npm run lint:eslint` from the Web directory and fix the reported lint errors.'
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
