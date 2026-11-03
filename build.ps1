# build.ps1 - Packaging script for go-ffmpeg-tools (Wails v2 + pnpm)
#
# Usage:
#   .\build.ps1                    # build for current platform (production)
#   .\build.ps1 -DebugBuild        # build debug binary
#   .\build.ps1 -Platform windows/amd64
#   .\build.ps1 -Platform "windows/amd64 linux/amd64"
#
# Requirements: Go, Node.js, pnpm, Wails CLI (auto-installed if missing).

[CmdletBinding()]
param(
    # Wails platform list, e.g. "windows/amd64" or "windows/amd64 linux/amd64".
    [string]$Platform = "",

    # Build a debug binary instead of a production one. (Named DebugBuild to
    # avoid clashing with PowerShell's automatic common parameter -Debug.)
    [switch]$DebugBuild
)

$ErrorActionPreference = "Stop"

function Assert-Command {
    param([string]$Name, [string]$InstallHint)
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        Write-Error "Required command '$Name' not found. $InstallHint"
    }
}

Write-Host "==> Checking prerequisites..." -ForegroundColor Cyan
Assert-Command "go" "Install Go from https://go.dev/dl/."
Assert-Command "node" "Install Node.js from https://nodejs.org/."
Assert-Command "pnpm" "Install pnpm with: npm install -g pnpm"

# Install Wails CLI if it is not available yet (pinned to v2.16.0).
if (-not (Get-Command "wails" -ErrorAction SilentlyContinue)) {
    Write-Host "==> Installing Wails CLI v2.16.0..." -ForegroundColor Cyan
    go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0
}

$repoRoot = $PSScriptRoot
if (-not $repoRoot) { $repoRoot = (Get-Location).Path }
Set-Location $repoRoot

Write-Host "==> Installing frontend dependencies (pnpm)..." -ForegroundColor Cyan
Push-Location "$repoRoot/frontend"
pnpm install --frozen-lockfile
if ($LASTEXITCODE -ne 0) { Pop-Location; Write-Error "pnpm install failed." }
Pop-Location

Write-Host "==> Running Wails build..." -ForegroundColor Cyan
# 编译时间（本地时区），通过 -ldflags -X 注入 main.buildTime，标题栏展示。
# 时间含空格，值内层加单引号，go build 解析 ldflags 时按引号保留完整值
$buildTime = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
Write-Host "    build time: $buildTime"
$wailsArgs = @("build", "-clean", "-ldflags", "-X 'main.buildTime=$buildTime'")
if ($DebugBuild) { $wailsArgs += "-debug" }
if ($Platform -ne "") { $wailsArgs += @("-platform", $Platform) }
& wails @wailsArgs
if ($LASTEXITCODE -ne 0) { Write-Error "wails build failed." }

# Copy bundled ffmpeg binaries next to the output executable so the app can
# find them at runtime: <exe dir>/build/ffmpeg/<goos>/<binary>.
Write-Host "==> Copying ffmpeg binaries into output..." -ForegroundColor Cyan
$binDir = Join-Path $repoRoot "build/bin"
if (Test-Path $binDir) {
    Get-ChildItem $binDir -Directory | ForEach-Object {
        $goos = $_.Name
        $src = Join-Path $repoRoot "build/ffmpeg/$goos"
        if (Test-Path $src) {
            $dst = Join-Path $_.FullName "build/ffmpeg/$goos"
            New-Item -ItemType Directory -Force -Path $dst | Out-Null
            Copy-Item -Path (Join-Path $src "*") -Destination $dst -Force
            Write-Host "    copied ffmpeg -> $dst"
        }
    }
}

Write-Host "==> Build finished. Output in build/bin/" -ForegroundColor Green
