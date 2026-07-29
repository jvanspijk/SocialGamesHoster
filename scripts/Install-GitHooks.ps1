[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"
$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path

git -C $projectRoot config core.hooksPath .githooks
if ($LASTEXITCODE -ne 0) {
    throw "Git hook configuration failed with exit code $LASTEXITCODE."
}

Write-Host "Git hooks configured."
