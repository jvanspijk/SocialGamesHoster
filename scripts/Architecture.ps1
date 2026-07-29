[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"
$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$hostRoot = Join-Path $projectRoot "Host"
$goDepcheckVersion = "v0.0.2"
$modulePath = "github.com/jvanspijk/SocialGamesHoster"

function Assert-NativeSuccess([string]$Step) {
    if ($LASTEXITCODE -ne 0) {
        throw "$Step failed with exit code $LASTEXITCODE."
    }
}

Push-Location $projectRoot
try {
    $configPath = Join-Path $hostRoot "depcheck.yml"
    if (Test-Path -LiteralPath $configPath) {
        throw "depcheck.yml already exists; the architecture check will not overwrite it."
    }
    @"
ignorePatterns:
  - '.*_test\.go$'
rules:
  # Pure shared policy must remain independent of persistence, platform, and features.
  - from: '^$modulePath/Host/internal/features/gamepolicy$'
    to:
      - '^github.com/pocketbase/'
      - '^github.com/pocketbase/dbx'
      - '^$modulePath/Host/internal/platform(/.*)?$'
      - '^$modulePath/Host/internal/features(/.*)?$'
  # Platform packages provide generic mechanisms and must not own feature policy.
  - from: '^$modulePath/Host/internal/platform(/.*)?$'
    to:
      - '^$modulePath/Host/internal/features(/.*)?$'
  # Feature and platform code must not depend on the composition root.
  - from: '^$modulePath/Host/internal/(features|platform)(/.*)?$'
    to:
      - '^$modulePath/Host/cmd(/.*)?$'
"@ | Set-Content -LiteralPath $configPath

    $previousGocache = $env:GOCACHE
    try {
        Write-Host "Checking package boundaries with go-depcheck..."
        $env:GOCACHE = Join-Path $projectRoot ".tmp\go-build-cache"
        go install "github.com/v-standard/go-depcheck/cmd/depcheck@$goDepcheckVersion"
        Assert-NativeSuccess "go-depcheck installation"
        $depcheckTool = Join-Path (go env GOPATH) "bin\depcheck.exe"
        if (-not (Test-Path -LiteralPath $depcheckTool)) {
            throw "go-depcheck was installed but its executable was not found at $depcheckTool."
        }
        Push-Location $hostRoot
        try {
            go vet ("-vettool={0}" -f $depcheckTool) ./...
        }
        finally {
            Pop-Location
        }
        Assert-NativeSuccess "go-depcheck"
    }
    finally {
        $env:GOCACHE = $previousGocache
        Remove-Item -LiteralPath $configPath -Force -ErrorAction SilentlyContinue
    }
}
finally {
    Pop-Location
}

Write-Host "Architecture checks passed."
