[CmdletBinding()]
param(
    [switch]$Run
)

function Install-FrontendDependencies {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string]$WebRoot
    )

    $packageJsonPath = Join-Path $WebRoot "package.json"
    $packageLockPath = Join-Path $WebRoot "package-lock.json"
    $nodeModulesPath = Join-Path $WebRoot "node_modules"
    $stampPath = Join-Path $nodeModulesPath ".sgh-package-fingerprint"

    if (-not (Test-Path -LiteralPath $packageJsonPath -PathType Leaf)) {
        throw "Frontend package manifest was not found: $packageJsonPath"
    }
    if (-not (Test-Path -LiteralPath $packageLockPath -PathType Leaf)) {
        throw "Frontend lockfile was not found: $packageLockPath"
    }

    $packageFingerprint = (Get-FileHash -LiteralPath $packageJsonPath -Algorithm SHA256).Hash.ToLowerInvariant()
    $lockFingerprint = (Get-FileHash -LiteralPath $packageLockPath -Algorithm SHA256).Hash.ToLowerInvariant()
    $fingerprint = "$packageFingerprint`n$lockFingerprint"

    if ((Test-Path -LiteralPath $nodeModulesPath -PathType Container) -and
        (Test-Path -LiteralPath $stampPath -PathType Leaf) -and
        ((Get-Content -LiteralPath $stampPath -Raw).Trim() -eq $fingerprint.Trim())) {
        Write-Host "Frontend dependencies already match the package manifests; skipping npm ci."
        return
    }

    Push-Location $WebRoot
    try {
        npm ci --no-audit --no-fund
        if ($LASTEXITCODE -ne 0) {
            throw "Frontend dependency installation failed with exit code $LASTEXITCODE."
        }
        Set-Content -LiteralPath $stampPath -Value $fingerprint -Encoding ascii
    }
    finally {
        Pop-Location
    }
}

if ($Run) {
    $projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
    Install-FrontendDependencies -WebRoot (Join-Path $projectRoot "Web")
}
