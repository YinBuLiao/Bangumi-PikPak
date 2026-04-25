param(
    [string]$Version = "",
    [string]$OutputDir = "dist/releases",
    [string]$ImageName = "bangumi-pikpak",
    [string]$ImageTag = "",
    [switch]$SkipTests,
    [switch]$SkipDocker
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$Root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $Root

if ([string]::IsNullOrWhiteSpace($Version)) {
    $Version = Get-Date -Format "yyyyMMdd-HHmmss"
}
if ([string]::IsNullOrWhiteSpace($ImageTag)) {
    $ImageTag = $Version
}

$ReleaseDir = Join-Path $Root $OutputDir
$StageDir = Join-Path $Root "dist/stage"
New-Item -ItemType Directory -Force -Path $ReleaseDir, $StageDir | Out-Null

function Invoke-Step {
    param(
        [string]$Title,
        [scriptblock]$Body
    )
    Write-Host ""
    Write-Host "==> $Title" -ForegroundColor Cyan
    & $Body
}

function Invoke-Native {
    param(
        [Parameter(Mandatory = $true)]
        [string]$FilePath,
        [string[]]$Arguments
    )
    & $FilePath @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "$FilePath failed with exit code $LASTEXITCODE"
    }
}

function Copy-CommonFiles {
    param([string]$TargetDir)
    New-Item -ItemType Directory -Force -Path $TargetDir | Out-Null
    Copy-Item -LiteralPath (Join-Path $Root "README.md") -Destination $TargetDir -Force
    Copy-Item -LiteralPath (Join-Path $Root "LICENSE") -Destination $TargetDir -Force

    $FrontendTarget = Join-Path $TargetDir "frontend"
    New-Item -ItemType Directory -Force -Path $FrontendTarget | Out-Null
    Copy-Item -LiteralPath (Join-Path $Root "frontend/dist") -Destination $FrontendTarget -Recurse -Force
}

function Add-Checksum {
    param([string]$Artifact)
    $Hash = Get-FileHash -Algorithm SHA256 -LiteralPath $Artifact
    $Line = "$($Hash.Hash.ToLower())  $(Split-Path -Leaf $Artifact)"
    Add-Content -LiteralPath (Join-Path $ReleaseDir "SHA256SUMS.txt") -Value $Line
}

Remove-Item -LiteralPath (Join-Path $ReleaseDir "SHA256SUMS.txt") -Force -ErrorAction SilentlyContinue

Invoke-Step "Build Vue frontend" {
    Push-Location (Join-Path $Root "frontend")
    try {
        Invoke-Native "npm" @("install")
        Invoke-Native "npm" @("run", "build")
    } finally {
        Pop-Location
    }
}

if (-not $SkipTests) {
    Invoke-Step "Run Go tests" {
        Invoke-Native "go" @("test", "./...")
    }
}

$LdFlags = "-s -w"

Invoke-Step "Package Windows amd64 exe" {
    $PackageName = "bangumi-pikpak-$Version-windows-amd64"
    $PackageDir = Join-Path $StageDir $PackageName
    Remove-Item -LiteralPath $PackageDir -Recurse -Force -ErrorAction SilentlyContinue
    New-Item -ItemType Directory -Force -Path $PackageDir | Out-Null

    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"
    Invoke-Native "go" @("build", "-trimpath", "-ldflags", $LdFlags, "-o", (Join-Path $PackageDir "bangumi-pikpak.exe"), ".")
    Copy-CommonFiles -TargetDir $PackageDir

    $ZipPath = Join-Path $ReleaseDir "$PackageName.zip"
    Remove-Item -LiteralPath $ZipPath -Force -ErrorAction SilentlyContinue
    Compress-Archive -Path (Join-Path $PackageDir "*") -DestinationPath $ZipPath -Force
    Add-Checksum $ZipPath
}

Invoke-Step "Package Linux amd64" {
    $PackageName = "bangumi-pikpak-$Version-linux-amd64"
    $PackageDir = Join-Path $StageDir $PackageName
    Remove-Item -LiteralPath $PackageDir -Recurse -Force -ErrorAction SilentlyContinue
    New-Item -ItemType Directory -Force -Path $PackageDir | Out-Null

    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"
    Invoke-Native "go" @("build", "-trimpath", "-ldflags", $LdFlags, "-o", (Join-Path $PackageDir "bangumi-pikpak"), ".")
    Copy-CommonFiles -TargetDir $PackageDir

    $TarPath = Join-Path $ReleaseDir "$PackageName.tar.gz"
    Remove-Item -LiteralPath $TarPath -Force -ErrorAction SilentlyContinue
    Push-Location $StageDir
    try {
        Invoke-Native "tar" @("-czf", $TarPath, $PackageName)
    } finally {
        Pop-Location
    }
    Add-Checksum $TarPath
}

Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue

if (-not $SkipDocker) {
    Invoke-Step "Build Docker image and save tar" {
        $Docker = Get-Command docker -ErrorAction SilentlyContinue
        if (-not $Docker) {
            Write-Warning "docker was not found; skip Docker packaging."
            return
        }
        $ImageRef = "${ImageName}:${ImageTag}"
        Invoke-Native "docker" @("build", "-t", $ImageRef, ".")
        $DockerTar = Join-Path $ReleaseDir "bangumi-pikpak-docker-$ImageTag.tar"
        Remove-Item -LiteralPath $DockerTar -Force -ErrorAction SilentlyContinue
        Invoke-Native "docker" @("save", "-o", $DockerTar, $ImageRef)
        Add-Checksum $DockerTar
    }
}

Write-Host ""
Write-Host "Packaging finished. Output directory:" -ForegroundColor Green
Write-Host "  $ReleaseDir"
Write-Host ""
Write-Host "Note: this script never runs git add / git commit / git push. dist/ is ignored by git." -ForegroundColor Yellow
