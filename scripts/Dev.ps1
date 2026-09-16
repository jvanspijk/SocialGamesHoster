[CmdletBinding()]
param(
    [switch]$Diagnostics
)

$ErrorActionPreference = "Stop"
$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$arguments = @("run", "./Host/cmd/socialgameshoster", "--", "--no-tray")
if ($Diagnostics) {
    $arguments += "--diagnostics"
}

$hostJob = Start-Job -ArgumentList $projectRoot, $arguments -ScriptBlock {
    param($root, [string[]]$goArguments)
    Set-Location -LiteralPath $root
    & go @goArguments
}
try {
    Push-Location (Join-Path $projectRoot "Web")
    try {
        npm run dev
    }
    finally {
        Pop-Location
    }
}
finally {
    Stop-Job -Job $hostJob -ErrorAction SilentlyContinue
    Remove-Job -Job $hostJob -Force -ErrorAction SilentlyContinue
}
