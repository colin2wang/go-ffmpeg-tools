# generate-icon.ps1 - Generate Wails app icons from frontend/public/favicon.svg
#
# Outputs:
#   build/appicon.png        1024x1024 PNG (Wails uses this during build; on
#                            Windows `wails build` also regenerates icon.ico
#                            from this file automatically)
#   build/windows/icon.ico   multi-size Windows icon (16..256)
#
# Usage: .\generate-icon.ps1
#        .\generate-icon.ps1 -Source path\to\other.svg
#
# Rasterization strategy (first available wins):
#   1. ImageMagick (magick)      - best quality, also builds the ICO directly
#   2. Inkscape                  - SVG rasterization only
#   3. Headless Edge/Chrome      - no extra installs needed on Windows; the
#                                  ICO is then assembled from the PNG with
#                                  System.Drawing (PNG-compressed entries)

[CmdletBinding()]
param(
    # Source SVG. Defaults to frontend/public/favicon.svg.
    [string]$Source = "",

    # Size of the generated appicon.png.
    [int]$Size = 1024
)

$ErrorActionPreference = "Stop"

$repoRoot = $PSScriptRoot
if (-not $repoRoot) { $repoRoot = (Get-Location).Path }
Set-Location $repoRoot

if ($Source -eq "") { $Source = Join-Path $repoRoot "frontend/public/favicon.svg" }
if (-not (Test-Path $Source)) { Write-Error "Source SVG not found: $Source" }

$appIconPath = Join-Path $repoRoot "build/appicon.png"
$icoPath = Join-Path $repoRoot "build/windows/icon.ico"
New-Item -ItemType Directory -Force -Path (Split-Path $appIconPath) | Out-Null
New-Item -ItemType Directory -Force -Path (Split-Path $icoPath) | Out-Null

function Get-CommandPath {
    param([string]$Name, [string[]]$Locations = @())
    if (Get-Command $Name -ErrorAction SilentlyContinue) {
        return (Get-Command $Name).Source
    }
    foreach ($loc in $Locations) {
        if (Test-Path $loc) { return $loc }
    }
    return $null
}

function Convert-SvgToPngWithMagick {
    param([string]$Magick, [string]$Svg, [string]$Png, [int]$Px)
    & $Magick -background none $Svg -resize "${Px}x${Px}" $Png
    if ($LASTEXITCODE -ne 0) { Write-Error "magick failed to rasterize the SVG." }
}

function Convert-SvgToPngWithInkscape {
    param([string]$Inkscape, [string]$Svg, [string]$Png, [int]$Px)
    & $Inkscape $Svg -o $Png -w $Px -h $Px
    if ($LASTEXITCODE -ne 0) { Write-Error "inkscape failed to rasterize the SVG." }
}

function Convert-SvgToPngWithBrowser {
    param([string]$Browser, [string]$Svg, [string]$Png, [int]$Px)
    $svgUri = ([System.Uri]$Svg).AbsoluteUri
    $args = @(
        "--headless=new", "--disable-gpu",
        "--default-background-color=00000000",
        "--window-size=$Px,$Px",
        "--screenshot=$Png",
        $svgUri
    )
    # Edge/Chrome writes harmless warnings to stderr (e.g. QQBrowser probe).
    # Under $ErrorActionPreference='Stop' PS 5.1 turns native stderr into a
    # NativeCommandError, so relax it just for this call.
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        & $Browser @args 2>&1 | Out-Null
    } finally {
        $ErrorActionPreference = $prevEap
    }
    if (-not (Test-Path $Png)) { Write-Error "Headless browser did not produce $Png." }
}

function Convert-PngToIco {
    # Build a multi-size .ico from a PNG using System.Drawing.
    # Every entry embeds a PNG blob (valid on Windows Vista and later).
    param([string]$Png, [string]$Ico, [int[]]$Sizes = @(16, 24, 32, 48, 64, 128, 256))

    Add-Type -AssemblyName System.Drawing

    $src = [System.Drawing.Image]::FromFile($Png)
    try {
        $entries = @()
        foreach ($px in $Sizes) {
            $bmp = New-Object System.Drawing.Bitmap $px, $px
            try {
                $g = [System.Drawing.Graphics]::FromImage($bmp)
                $g.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
                $g.DrawImage($src, 0, 0, $px, $px)
                $g.Dispose()
                $ms = New-Object System.IO.MemoryStream
                $bmp.Save($ms, [System.Drawing.Imaging.ImageFormat]::Png)
                $entries += ,@($px, $ms.ToArray())
                $ms.Dispose()
            } finally {
                $bmp.Dispose()
            }
        }
    } finally {
        $src.Dispose()
    }

    $icoStream = New-Object System.IO.MemoryStream
    $bw = New-Object System.IO.BinaryWriter $icoStream
    $bw.Write([UInt16]0)                    # reserved
    $bw.Write([UInt16]1)                    # type: icon
    $bw.Write([UInt16]$entries.Count)       # image count
    $offset = 6 + 16 * $entries.Count
    foreach ($e in $entries) {
        $px = $e[0]; $data = [byte[]]$e[1]
        $dim = if ($px -ge 256) { [byte]0 } else { [byte]$px }
        $bw.Write($dim)                     # width  (0 = 256)
        $bw.Write($dim)                     # height (0 = 256)
        $bw.Write([byte]0)                  # palette
        $bw.Write([byte]0)                  # reserved
        $bw.Write([UInt16]1)                # color planes
        $bw.Write([UInt16]32)               # bits per pixel
        $bw.Write([UInt32]$data.Length)
        $bw.Write([UInt32]$offset)
        $offset += $data.Length
    }
    foreach ($e in $entries) { $bw.Write([byte[]]$e[1]) }
    $bw.Flush()
    [System.IO.File]::WriteAllBytes($Ico, $icoStream.ToArray())
    $bw.Dispose()
    $icoStream.Dispose()
}

Write-Host "==> Rasterizing SVG ($Size px): $Source" -ForegroundColor Cyan

$magick = Get-CommandPath "magick" @()
$inkscape = Get-CommandPath "inkscape" @()
$browser = Get-CommandPath "msedge" @(
    "C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe",
    "C:\Program Files\Microsoft\Edge\Application\msedge.exe"
)
if (-not $browser) {
    $browser = Get-CommandPath "chrome" @(
        "C:\Program Files\Google\Chrome\Application\chrome.exe"
    )
}

if ($magick) {
    Convert-SvgToPngWithMagick -Magick $magick -Svg $Source -Png $appIconPath -Px $Size
    Write-Host "    appicon.png <- ImageMagick"
    & $magick $appIconPath -define icon:auto-resize=256,128,64,48,32,24,16 $icoPath
    if ($LASTEXITCODE -ne 0) { Write-Error "magick failed to build the ICO." }
    Write-Host "    icon.ico    <- ImageMagick"
} else {
    if ($inkscape) {
        Convert-SvgToPngWithInkscape -Inkscape $inkscape -Svg $Source -Png $appIconPath -Px $Size
        Write-Host "    appicon.png <- Inkscape"
    } elseif ($browser) {
        Convert-SvgToPngWithBrowser -Browser $browser -Svg $Source -Png $appIconPath -Px $Size
        Write-Host "    appicon.png <- headless browser"
    } else {
        Write-Error "No rasterizer found. Install ImageMagick or Inkscape, or ensure Edge/Chrome is present."
    }
    Write-Host "==> Building ICO (System.Drawing)" -ForegroundColor Cyan
    Convert-PngToIco -Png $appIconPath -Ico $icoPath
    Write-Host "    icon.ico    <- System.Drawing"
}

Write-Host "==> Icons generated:" -ForegroundColor Green
Write-Host "    $appIconPath"
Write-Host "    $icoPath"
Write-Host "Note: wails build regenerates build/windows/icon.ico from appicon.png automatically."
