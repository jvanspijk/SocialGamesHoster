[CmdletBinding()]
param(
    [string]$Version = "",
    [switch]$SkipTests,
    [switch]$SkipInstaller
)

$buildLogPath = Join-Path $PSScriptRoot "Build.log"
Start-Transcript -Path $buildLogPath -Force | Out-Null

try {
& {
$ErrorActionPreference = "Stop"
$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$versionFile = Join-Path $projectRoot "VERSION"
if ([string]::IsNullOrWhiteSpace($Version)) {
    if (-not (Test-Path -LiteralPath $versionFile -PathType Leaf)) {
        throw "Version file was not found: $versionFile"
    }
    $Version = (Get-Content -LiteralPath $versionFile -Raw).Trim()
}
$webRoot = Join-Path $projectRoot "Web"
$embeddedRoot = Join-Path $projectRoot "Host\embedded\web"
$distRoot = Join-Path $projectRoot "dist"
$frontendDependencyInstaller = Join-Path $PSScriptRoot "Install-FrontendDependencies.ps1"
. $frontendDependencyInstaller
$versionMatch = [regex]::Match($Version, "^\d+(?:\.\d+){0,3}")
if (-not $versionMatch.Success) {
    throw "Version must begin with one to four numeric components, for example 1.2.3 or 1.2.3-beta.1."
}
$windowsVersionParts = @($versionMatch.Value.Split("."))
while ($windowsVersionParts.Count -lt 4) {
    $windowsVersionParts += "0"
}
$windowsVersion = $windowsVersionParts -join "."

function Assert-NativeSuccess([string]$Step) {
    if ($LASTEXITCODE -ne 0) {
        throw "$Step failed with exit code $LASTEXITCODE."
    }
}

if (-not $embeddedRoot.StartsWith($projectRoot + [IO.Path]::DirectorySeparatorChar)) {
    throw "The embedded web target is outside the project."
}

New-Item -ItemType Directory -Path $distRoot -Force | Out-Null
Get-ChildItem -LiteralPath $distRoot -Filter "SocialGamesHoster-*-windows-x64-setup.exe" -File |
    Remove-Item -Force

Push-Location $webRoot
try {
    Install-FrontendDependencies -WebRoot $webRoot
    if (-not $SkipTests) {
        npm run check
        Assert-NativeSuccess "Frontend type checks"
        npm run test:unit
        Assert-NativeSuccess "Frontend contract tests"
        npm run lint
        Assert-NativeSuccess "Frontend lint"
    }
    npm run build
    Assert-NativeSuccess "Frontend build"
}
finally {
    Pop-Location
}

Get-ChildItem -LiteralPath $embeddedRoot -Force |
    Where-Object Name -ne ".gitkeep" |
    Remove-Item -Recurse -Force
Copy-Item -Path (Join-Path $webRoot "build\*") -Destination $embeddedRoot -Recurse -Force

Push-Location $projectRoot
try {
    if (-not $SkipTests) {
        go test ./Host/...
        Assert-NativeSuccess "Go tests"
    }
    $previousGoOs = $env:GOOS
    $previousGoArch = $env:GOARCH
    $previousCgo = $env:CGO_ENABLED
    try {
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        $env:CGO_ENABLED = "0"
        go build -trimpath -ldflags "-s -w -H=windowsgui -X main.version=$Version" `
            -o (Join-Path $distRoot "SocialGamesHoster.exe") `
            ./Host/cmd/socialgameshoster
        Assert-NativeSuccess "Windows host build"
    }
    finally {
        $env:GOOS = $previousGoOs
        $env:GOARCH = $previousGoArch
        $env:CGO_ENABLED = $previousCgo
    }
}
finally {
    Pop-Location
}

$signTool = Get-Command "signtool.exe" -ErrorAction SilentlyContinue
if ($env:SGH_SIGN_CERT_THUMBPRINT -and $signTool) {
    & $signTool.Source sign /sha1 $env:SGH_SIGN_CERT_THUMBPRINT /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 `
        (Join-Path $distRoot "SocialGamesHoster.exe")
    Assert-NativeSuccess "Application signing"
}

if (-not $SkipInstaller) {
    $iscc = Get-Command "iscc.exe" -ErrorAction SilentlyContinue
    $isccPath = if ($iscc) { $iscc.Source } else { $null }
    if (-not $isccPath) {
        $innoSetupCandidates = @(
            (Join-Path $env:ProgramFiles "Inno Setup 6\ISCC.exe"),
            (Join-Path ${env:ProgramFiles(x86)} "Inno Setup 6\ISCC.exe"),
            (Join-Path $env:LOCALAPPDATA "Programs\Inno Setup 6\ISCC.exe")
        )
        $isccPath = $innoSetupCandidates |
            Where-Object { $_ -and (Test-Path -LiteralPath $_ -PathType Leaf) } |
            Select-Object -First 1
    }
    if (-not $isccPath) {
        throw "Inno Setup 6 was not found on PATH or in a standard installation directory. Install it or use -SkipInstaller."
    }
    & $isccPath "/DAppVersion=$Version" "/DWindowsVersion=$windowsVersion" `
        (Join-Path $projectRoot "packaging\windows\installer.iss")
    Assert-NativeSuccess "Installer build"

    $installer = Get-ChildItem -LiteralPath $distRoot -Filter "*-setup.exe" |
        Sort-Object LastWriteTime -Descending |
        Select-Object -First 1
    if ($env:SGH_SIGN_CERT_THUMBPRINT -and $signTool -and $installer) {
        & $signTool.Source sign /sha1 $env:SGH_SIGN_CERT_THUMBPRINT /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 `
            $installer.FullName
        Assert-NativeSuccess "Installer signing"
    }
}

Get-ChildItem -LiteralPath $distRoot -File |
    Where-Object Name -ne "SHA256SUMS.txt" |
    Get-FileHash -Algorithm SHA256 |
    ForEach-Object { "$($_.Hash.ToLowerInvariant())  $([IO.Path]::GetFileName($_.Path))" } |
    Set-Content -LiteralPath (Join-Path $distRoot "SHA256SUMS.txt") -Encoding ascii
}
}
catch {
    Write-Host "Build failed: $($_.Exception.Message)"
    if ($_.ScriptStackTrace) {
        Write-Host $_.ScriptStackTrace
    }
    throw
}
finally {
    Stop-Transcript | Out-Null
}
