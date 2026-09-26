[CmdletBinding()]
param()

# Dev.ps1 runs the Vite UI on 9091 and the Go host on 8090.
$ErrorActionPreference = "Stop"
$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path.TrimEnd('\')
$webRoots = @((Join-Path $projectRoot "Web"))
if (Get-Command git -ErrorAction SilentlyContinue) {
    $worktreeList = & git -C $projectRoot worktree list --porcelain 2>$null
    if ($LASTEXITCODE -eq 0) {
        foreach ($line in $worktreeList) {
            if ($line -match '^worktree (.+)$') {
                $webRoots += Join-Path $Matches[1] "Web"
            }
        }
    }
}
$webRoots = @($webRoots | ForEach-Object { [IO.Path]::GetFullPath($_) } | Select-Object -Unique)

function Test-DevViteProcess {
    param(
        [Parameter(Mandatory = $true)]
        $Process
    )

    if (($Process.Name -ine "node.exe") -or
        ($Process.CommandLine -notmatch '(?i)[\\/]vite[\\/]bin[\\/]vite\.js"?\s+dev\b') -or
        ($Process.CommandLine -notmatch '(?i)--port\s+9091\b')) {
        return $false
    }

    foreach ($webRoot in $webRoots) {
        if ($Process.CommandLine.IndexOf($webRoot + '\', [StringComparison]::OrdinalIgnoreCase) -ge 0) {
            return $true
        }
    }
    return $false
}

function Test-DevHostProcess {
    param(
        [Parameter(Mandatory = $true)]
        $Process
    )

    if (($Process.Name -ine "socialgameshoster.exe") -or
        ($Process.ExecutablePath -notmatch "(?i)[\\/]go-build[^\\/]*[\\/]")) {
        return $false
    }

    $parent = Get-CimInstance -ClassName Win32_Process `
        -Filter "ProcessId = $($Process.ParentProcessId)" `
        -ErrorAction SilentlyContinue
    return ($null -ne $parent) -and
        ($parent.Name -ieq "go.exe") -and
        ($parent.CommandLine -match "(?i)\brun\b.*Host[\\/]cmd[\\/]socialgameshoster") -and
        ($Process.CommandLine -match "(?i)--no-tray\b")
}

$targets = @{}
foreach ($port in @(8090, 9091)) {
    $listeners = @(Get-NetTCPConnection -State Listen -LocalPort $port -ErrorAction SilentlyContinue)
    foreach ($listener in $listeners) {
        $processId = [int]$listener.OwningProcess
        $process = Get-CimInstance -ClassName Win32_Process `
            -Filter "ProcessId = $processId" `
            -ErrorAction SilentlyContinue
        if ($null -eq $process) {
            Write-Warning "Could not inspect the process holding port $port (PID $processId). Leaving it running."
            continue
        }

        $isExpectedProcess = if ($port -eq 9091) {
            Test-DevViteProcess -Process $process
        } else {
            Test-DevHostProcess -Process $process
        }

        if ($isExpectedProcess) {
            $targets[$processId] = $process
        } else {
            Write-Warning "Port $port is held by $($process.Name) (PID $processId), which does not look like this repository's dev server. Leaving it running."
        }
    }
}

if ($targets.Count -eq 0) {
    Write-Host "No matching SocialGamesHoster dev server is listening on ports 8090 or 9091."
    return
}

foreach ($process in $targets.Values) {
    Write-Host "Stopping $($process.Name) (PID $($process.ProcessId)): $($process.CommandLine)"
    Stop-Process -Id ([int]$process.ProcessId) -Force
}

Start-Sleep -Seconds 1
$stillListening = @(Get-NetTCPConnection -State Listen -ErrorAction SilentlyContinue |
    Where-Object { ($_.LocalPort -in @(8090, 9091)) -and $targets.ContainsKey([int]$_.OwningProcess) })
if ($stillListening.Count -gt 0) {
    $remaining = ($stillListening | ForEach-Object { "port $($_.LocalPort), PID $($_.OwningProcess)" }) -join "; "
    throw "Some dev server listeners are still present: $remaining"
}

Write-Host "SocialGamesHoster dev server stopped."
