param(
    [string]$ComposeFile = "docker-compose.full.yml",
    [switch]$Build,
    [switch]$Pull,
    [switch]$Down
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$Root = Resolve-Path (Join-Path $PSScriptRoot "..")
$ComposePath = Join-Path $Root $ComposeFile
$SecretsDir = Join-Path $Root "docker-data\secrets"
$MysqlPasswordPath = Join-Path $SecretsDir "mysql_password.txt"
$MysqlRootPasswordPath = Join-Path $SecretsDir "mysql_root_password.txt"
$AdminPasswordPath = Join-Path $SecretsDir "admin_password.txt"

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

function New-RandomSecret {
    param([int]$Length = 32)
    $Chars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
    $Bytes = [byte[]]::new($Length)
    [System.Security.Cryptography.RandomNumberGenerator]::Fill($Bytes)
    $Result = New-Object System.Text.StringBuilder
    foreach ($Byte in $Bytes) {
        [void]$Result.Append($Chars[$Byte % $Chars.Length])
    }
    return $Result.ToString()
}

function Ensure-Secrets {
    New-Item -ItemType Directory -Force -Path $SecretsDir | Out-Null
    if (-not (Test-Path -LiteralPath $MysqlPasswordPath) -or [string]::IsNullOrWhiteSpace((Get-Content -Raw -LiteralPath $MysqlPasswordPath -ErrorAction SilentlyContinue))) {
        Set-Content -LiteralPath $MysqlPasswordPath -Value (New-RandomSecret 24) -NoNewline -Encoding ASCII
    }
    if (-not (Test-Path -LiteralPath $MysqlRootPasswordPath) -or [string]::IsNullOrWhiteSpace((Get-Content -Raw -LiteralPath $MysqlRootPasswordPath -ErrorAction SilentlyContinue))) {
        Set-Content -LiteralPath $MysqlRootPasswordPath -Value (New-RandomSecret 32) -NoNewline -Encoding ASCII
    }
    if (-not (Test-Path -LiteralPath $AdminPasswordPath) -or [string]::IsNullOrWhiteSpace((Get-Content -Raw -LiteralPath $AdminPasswordPath -ErrorAction SilentlyContinue))) {
        Set-Content -LiteralPath $AdminPasswordPath -Value (New-RandomSecret 16) -NoNewline -Encoding ASCII
    }
}

function Show-Credentials {
    $MysqlPassword = (Get-Content -Raw -LiteralPath $MysqlPasswordPath).Trim()
    $MysqlRootPassword = (Get-Content -Raw -LiteralPath $MysqlRootPasswordPath).Trim()
    $AdminPassword = (Get-Content -Raw -LiteralPath $AdminPasswordPath).Trim()
    Write-Host ""
    Write-Host "AnimeX Docker 一体化环境已启动。" -ForegroundColor Green
    Write-Host "Web 地址:        http://127.0.0.1:8080"
    Write-Host ""
    Write-Host "系统已自动安装，直接登录：" -ForegroundColor Cyan
    Write-Host "管理员账号:      admin"
    Write-Host "管理员密码:      $AdminPassword"
    Write-Host ""
    Write-Host "安装向导中填写以下数据库信息：" -ForegroundColor Cyan
    Write-Host "MySQL 主机:      mysql"
    Write-Host "MySQL 端口:      3306"
    Write-Host "MySQL 数据库:    animex"
    Write-Host "MySQL 用户名:    animex"
    Write-Host "MySQL 密码:      $MysqlPassword"
    Write-Host "Redis 地址:      redis:6379"
    Write-Host "Redis 密码:      留空"
    Write-Host "Redis DB:        0"
    Write-Host ""
    Write-Host "凭据文件：" -ForegroundColor Yellow
    Write-Host "  $MysqlPasswordPath"
    Write-Host "  $MysqlRootPasswordPath"
    Write-Host "  $AdminPasswordPath"
    Write-Host ""
    Write-Host "MySQL root 密码: $MysqlRootPassword"
}

Push-Location $Root
try {
    if (-not (Test-Path -LiteralPath $ComposePath)) {
        throw "Compose file not found: $ComposePath"
    }

    $Docker = Get-Command docker -ErrorAction SilentlyContinue
    if (-not $Docker) {
        throw "docker was not found in PATH"
    }

    if ($Down) {
        Invoke-Native "docker" @("compose", "-f", $ComposePath, "down")
        return
    }

    Ensure-Secrets

    if ($Build) {
        Invoke-Native "docker" @("build", "-t", "yinbuliao/bangumi-pikpak:latest", ".")
    } elseif ($Pull) {
        Invoke-Native "docker" @("compose", "-f", $ComposePath, "pull", "animex", "mysql", "redis", "mysql-secret-init")
    } else {
        Invoke-Native "docker" @("compose", "-f", $ComposePath, "pull", "animex", "mysql", "redis", "mysql-secret-init")
    }

    Invoke-Native "docker" @("compose", "-f", $ComposePath, "up", "-d")
    Show-Credentials
} finally {
    Pop-Location
}
