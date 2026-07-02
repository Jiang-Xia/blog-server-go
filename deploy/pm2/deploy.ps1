# =============================================================================
# blog-server-go Windows 本地部署入口
# 用法：powershell -ExecutionPolicy Bypass -File deploy/pm2/deploy.ps1
#       make deploy
#
# 流程：[0] 从本仓库 env.production 生成 configs → [1] 交叉编译 → [2] 打 tar → [3] 上传 → [4] 远程 PM2 → [5] 验证
# =============================================================================

param(
  [string]$EnvFileName = 'deploy.local.env'
)

$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'ssh-lib.ps1')
Initialize-DeployConsoleEncoding

$Root = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$EnvFile = Join-Path $PSScriptRoot $EnvFileName
$PackDir = Join-Path $PSScriptRoot '.pack'

if (-not (Test-Path $EnvFile)) {
  throw "Missing $EnvFile — copy deploy.local.env.example to deploy.local.env"
}

$cfg = @{}
Get-Content $EnvFile | ForEach-Object {
  if ($_ -match '^\s*#' -or $_ -match '^\s*$') { return }
  $parts = $_ -split '=', 2
  if ($parts.Length -eq 2) {
    $cfg[$parts[0].Trim()] = $parts[1].Trim()
  }
}

$DeployHost = $cfg['DEPLOY_HOST']
$DeployUser = $cfg['DEPLOY_USER']
$DeployPort = if ($cfg['DEPLOY_PORT']) { $cfg['DEPLOY_PORT'] } else { '22' }
$DeployPassword = $cfg['DEPLOY_PASSWORD']
$RemoteDir = $cfg['DEPLOY_REMOTE_DIR']
$HostKey = $cfg['DEPLOY_HOSTKEY']
$PublicDir = $cfg['DEPLOY_PUBLIC_DIR']
$Pm2Apps = if ($cfg['DEPLOY_PM2_APPS']) { $cfg['DEPLOY_PM2_APPS'] } else { 'gateway,user,blog,rpg' }
$EcosystemFile = if ($cfg['DEPLOY_ECOSYSTEM_FILE']) { $cfg['DEPLOY_ECOSYSTEM_FILE'] } else { 'ecosystem.config.js' }
$TarName = if ($cfg['DEPLOY_TAR_NAME']) { $cfg['DEPLOY_TAR_NAME'] } else { 'blog-server-go.tar.gz' }
$DeployEnvFile = $cfg['DEPLOY_ENV_FILE']
$DeployConfigDir = if ($cfg['DEPLOY_CONFIG_DIR']) { $cfg['DEPLOY_CONFIG_DIR'] } else { 'deploy/pm2/configs' }

Assert-SafeTarName $TarName

if (-not $DeployHost -or -not $DeployUser -or -not $RemoteDir) {
  throw 'deploy.local.env needs DEPLOY_HOST, DEPLOY_USER, DEPLOY_REMOTE_DIR'
}

$EcosystemPath = Join-Path $Root $EcosystemFile
if (-not (Test-Path $EcosystemPath)) {
  throw "Missing ecosystem file: $EcosystemPath"
}

function Find-Executable {
  param([string]$Name)
  $paths = @(
    $Name,
    "$env:ProgramFiles\PuTTY\$Name.exe",
    "${env:ProgramFiles(x86)}\PuTTY\$Name.exe"
  )
  foreach ($p in $paths) {
    if (Test-Path $p) { return $p }
    $cmd = Get-Command $p -ErrorAction SilentlyContinue
    if ($cmd) { return $cmd.Source }
  }
  return $null
}

$Plink = Find-Executable -Name 'plink'
$Pscp = Find-Executable -Name 'pscp'
$usePlink = $DeployPassword -and $Plink -and $Pscp

$plinkArgs = @('-batch', '-P', $DeployPort)
if ($HostKey) { $plinkArgs += @('-hostkey', $HostKey) }
if ($usePlink) { $plinkArgs += @('-pw', $DeployPassword) }

$pscpArgs = @('-batch', '-P', $DeployPort)
if ($HostKey) { $pscpArgs += @('-hostkey', $HostKey) }
if ($usePlink) { $pscpArgs += @('-pw', $DeployPassword) }

function Invoke-Remote {
  param([string]$Command)
  if ($usePlink) {
    & $Plink @plinkArgs "$DeployUser@$DeployHost" $Command
    if ($LASTEXITCODE -ne 0) { throw 'plink failed' }
  } else {
    ssh -p $DeployPort -o BatchMode=yes -o StrictHostKeyChecking=accept-new "$DeployUser@$DeployHost" $Command
    if ($LASTEXITCODE -ne 0) { throw 'ssh failed' }
  }
}

function Copy-ToRemote {
  param([string]$LocalPath, [string]$RemotePath)
  if (-not (Test-Path -LiteralPath $LocalPath)) {
    throw "Local file not found: $LocalPath"
  }
  if ($usePlink) {
    & $Pscp @pscpArgs -scp "$LocalPath" "${DeployUser}@${DeployHost}:${RemotePath}"
    if ($LASTEXITCODE -ne 0) { throw "pscp failed: $LocalPath" }
  } else {
    scp -P $DeployPort -o BatchMode=yes -o StrictHostKeyChecking=accept-new "$LocalPath" "${DeployUser}@${DeployHost}:${RemotePath}"
    if ($LASTEXITCODE -ne 0) { throw 'scp failed' }
  }
}

function Resolve-DeployEnvPath {
  if (-not $DeployEnvFile) {
    return Join-Path $PSScriptRoot 'env.production'
  }
  if ([System.IO.Path]::IsPathRooted($DeployEnvFile)) {
    return $DeployEnvFile
  }
  $relPath = $DeployEnvFile -replace '/', '\'
  return Join-Path $Root $relPath
}

function Resolve-DeployConfigDir {
  if ([System.IO.Path]::IsPathRooted($DeployConfigDir)) {
    return $DeployConfigDir
  }
  return Join-Path $Root ($DeployConfigDir -replace '/', '\')
}

$prodEnvPath = Resolve-DeployEnvPath
$configDir = Resolve-DeployConfigDir
if (-not (Test-Path $prodEnvPath)) {
  throw "Missing $prodEnvPath — copy deploy/pm2/env.production.example to deploy/pm2/env.production and fill secrets (or once: cp ../blog-server/deploy/pm2/env.production deploy/pm2/env.production)"
}

Write-Host "==> Remote: $RemoteDir | PM2 apps: $Pm2Apps"
Write-Host "==> [0/5] Sync configs from env.production: $prodEnvPath"
Push-Location $Root
go run ./scripts/sync_pm2_config_from_nest.go --env $prodEnvPath --out $configDir
if ($LASTEXITCODE -ne 0) { throw 'sync_pm2_config_from_nest failed' }
Pop-Location

$requiredConfigs = @('gateway.yaml', 'user.yaml', 'blog.yaml', 'rpg.yaml')
foreach ($name in $requiredConfigs) {
  $path = Join-Path $configDir $name
  if (-not (Test-Path $path)) {
    throw "Missing $path after Nest env sync"
  }
}

# ---------- [1/5] 交叉编译 Linux amd64 ----------
Write-Host '==> [1/5] Cross-compile linux/amd64'
$env:GOOS = 'linux'
$env:GOARCH = 'amd64'
$env:CGO_ENABLED = '0'
Push-Location $Root
$services = @(
  @{ Name = 'gateway'; Path = './services/gateway/cmd' },
  @{ Name = 'user'; Path = './services/user/cmd' },
  @{ Name = 'blog'; Path = './services/blog/cmd' },
  @{ Name = 'rpg'; Path = './services/rpg/cmd' }
)
$buildOut = Join-Path $PackDir 'build-bin'
if (Test-Path $buildOut) { Remove-Item -Recurse -Force $buildOut }
New-Item -ItemType Directory -Path $buildOut | Out-Null
foreach ($svc in $services) {
  $out = Join-Path $buildOut $svc.Name
  Write-Host "    go build $($svc.Name)"
  go build -ldflags='-s -w' -o $out $svc.Path
  if ($LASTEXITCODE -ne 0) { throw "go build failed: $($svc.Name)" }
}
Pop-Location

# ---------- [2/5] 组装 staging 并打 tar ----------
Write-Host '==> [2/5] Pack tar'
$staging = Join-Path $PackDir 'staging'
if (Test-Path $staging) { Remove-Item -Recurse -Force $staging }
New-Item -ItemType Directory -Path (Join-Path $staging 'bin') | Out-Null
New-Item -ItemType Directory -Path (Join-Path $staging 'configs') | Out-Null

foreach ($svc in $services) {
  Copy-Item (Join-Path $buildOut $svc.Name) (Join-Path $staging "bin/$($svc.Name)")
}
foreach ($name in $requiredConfigs) {
  Copy-Item (Join-Path $configDir $name) (Join-Path $staging "configs/$name")
}
Copy-Item $EcosystemPath (Join-Path $staging $EcosystemFile)
Write-Host "==> Pack configs (from env.production): $configDir"

if (-not (Test-Path $PackDir)) { New-Item -ItemType Directory -Path $PackDir | Out-Null }
$tarLocal = Join-Path $PackDir $TarName
if (Test-Path $tarLocal) { Remove-Item -Force $tarLocal }

Push-Location $staging
tar -czf $tarLocal .
Pop-Location

$tarSize = [math]::Round((Get-Item $tarLocal).Length / 1MB, 2)
Write-Host "==> Tar: $tarLocal ($tarSize MB)"

# ---------- [3/5] 上传 ----------
Write-Host '==> [3/5] Upload'
$remoteTar = "/tmp/$TarName"
Invoke-Remote "mkdir -p $RemoteDir"
Copy-ToRemote -LocalPath $tarLocal -RemotePath $remoteTar
Copy-ToRemote -LocalPath (Join-Path $PSScriptRoot 'release-lib.sh') -RemotePath '/tmp/release-lib.sh'
Copy-ToRemote -LocalPath (Join-Path $PSScriptRoot 'remote-deploy.sh') -RemotePath '/tmp/remote-deploy.sh'

# ---------- [4/5] 远程部署 ----------
Write-Host '==> [4/5] Remote: release -> switch -> pm2 reload'
$eRemoteDir = Escape-ShellSingleQuoted $RemoteDir
$ePm2Apps = Escape-ShellSingleQuoted $Pm2Apps
$eEcosystemFile = Escape-ShellSingleQuoted $EcosystemFile
$eRemoteTar = Escape-ShellSingleQuoted $remoteTar
$ePublicDir = if ($PublicDir) { Escape-ShellSingleQuoted $PublicDir } else { '' }
$publicEnv = if ($ePublicDir) { " DEPLOY_PUBLIC_DIR='$ePublicDir'" } else { '' }
$remoteCmd = "chmod +x /tmp/remote-deploy.sh && DEPLOY_REMOTE_DIR='$eRemoteDir' DEPLOY_PM2_APPS='$ePm2Apps' DEPLOY_ECOSYSTEM_FILE='$eEcosystemFile' DEPLOY_TAR_PATH='$eRemoteTar'${publicEnv} bash /tmp/remote-deploy.sh"
Invoke-Remote $remoteCmd

# ---------- [5/5] 验证 ----------
Write-Host '==> [5/5] Verify'
$eAppsForVerify = Escape-ShellSingleQuoted $Pm2Apps
Invoke-Remote "source ~/.nvm/nvm.sh && nvm use default && DEPLOY_REMOTE_DIR='$eRemoteDir' source /tmp/release-lib.sh && release_pm2_verify_all '$eAppsForVerify'"

Write-Host '==> Deploy finished'
