[CmdletBinding()]
param(
    [switch]$SkipTests,
    [switch]$SkipInstaller
)

$logsRoot = Join-Path $PSScriptRoot "logs"
New-Item -ItemType Directory -Path $logsRoot -Force | Out-Null
$buildLogPath = Join-Path $logsRoot "Build.log"
Remove-Item -LiteralPath $buildLogPath -Force -ErrorAction SilentlyContinue
Start-Transcript -Path $buildLogPath -Force | Out-Null

try {
& {
$ErrorActionPreference = "Stop"
$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$git = Get-Command "git" -ErrorAction SilentlyContinue
if (-not $git) {
    throw "Git is required to build a release version. Run this script from a Git checkout with commit metadata."
}
$commitHash = & $git.Source -C $projectRoot rev-parse --short=7 HEAD
if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($commitHash)) {
    throw "Could not determine the Git commit for the release version. Run this script from a Git checkout with commit metadata."
}
$utcDate = [DateTime]::UtcNow.ToString("yyyy.MM.dd", [Globalization.CultureInfo]::InvariantCulture)
$Version = "$utcDate-$($commitHash.Trim())"
$webRoot = Join-Path $projectRoot "Web"
$embeddedRoot = Join-Path $projectRoot "Host\embedded\web"
$distRoot = Join-Path $projectRoot "dist"
$frontendDependencyInstaller = Join-Path $PSScriptRoot "Install-FrontendDependencies.ps1"
. $frontendDependencyInstaller
$windowsVersion = "$utcDate.0"

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
        npm run check:types
        Assert-NativeSuccess "Frontend type checks"
        npm run test:unit
        Assert-NativeSuccess "Frontend contract tests"
        npm run format:check
        Assert-NativeSuccess "Frontend formatting check"
        npm run lint:eslint
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
    $previousGoOs = $env:GOOS
    $previousGoArch = $env:GOARCH
    $previousCgo = $env:CGO_ENABLED
    $mainPackageRoot = Join-Path $projectRoot "Host\cmd\socialgameshoster"
    $versionResourcePath = Join-Path $mainPackageRoot "versioninfo.syso"
    $versionResourceGenerated = $false
    try {
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        $env:CGO_ENABLED = "0"
        if (-not $SkipTests) {
            go test -trimpath ./Host/...
            Assert-NativeSuccess "Go tests"
        }
        if (Test-Path -LiteralPath $versionResourcePath) {
            throw "Temporary Windows version resource already exists: $versionResourcePath"
        }
        $versionResourceGenerated = $true
        $versionParts = @($utcDate.Split("."))
        & go run "github.com/josephspurrier/goversioninfo/cmd/goversioninfo@v1.7.0" `
            "-64" "-o=$versionResourcePath" `
            "-ver-major=$($versionParts[0])" "-ver-minor=$($versionParts[1])" `
            "-ver-patch=$($versionParts[2])" "-ver-build=0" `
            "-product-ver-major=$($versionParts[0])" "-product-ver-minor=$($versionParts[1])" `
            "-product-ver-patch=$($versionParts[2])" "-product-ver-build=0" `
            "-file-version=$windowsVersion" "-product-version=$Version" `
            "-company=Social Games Hoster contributors" "-description=Local-first social games app" `
            "-internal-name=SocialGamesHoster" "-original-name=SocialGamesHoster.exe" `
            "-product-name=Social Games Hoster" `
            (Join-Path $projectRoot "packaging\windows\versioninfo.json")
        Assert-NativeSuccess "Windows executable version resource generation"
        go build -trimpath -ldflags "-s -w -H=windowsgui -X main.version=$Version" `
            -o (Join-Path $distRoot "SocialGamesHoster.exe") `
            ./Host/cmd/socialgameshoster
        Assert-NativeSuccess "Windows host build"
    }
    finally {
        if ($versionResourceGenerated) {
            Remove-Item -LiteralPath $versionResourcePath -Force -ErrorAction SilentlyContinue
        }
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
