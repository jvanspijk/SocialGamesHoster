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

$hostProcess = Start-Process -FilePath "go" -ArgumentList $arguments -WorkingDirectory $projectRoot `
    -PassThru -WindowStyle Hidden
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
    if (-not $hostProcess.HasExited) {
        Stop-Process -Id $hostProcess.Id
    }
}
