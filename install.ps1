# herdrweb installer for Windows - downloads a prebuilt release binary.
#
#   irm https://raw.githubusercontent.com/sarathsp06/herdrweb/main/install.ps1 | iex
#
# Environment overrides:
#   HERDR_VERSION   version to install, e.g. v0.1.0 or 0.1.0 (default: latest release)
#   BINDIR          install directory (default: %LOCALAPPDATA%\Programs\herdrweb)
#   HERDRWEB_REPO   GitHub owner/repo to install from (default: sarathsp06/herdrweb)
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$repo = if ($env:HERDRWEB_REPO) { $env:HERDRWEB_REPO } else { 'sarathsp06/herdrweb' }
$project = 'herdrweb'
$bin = 'herdrweb.exe'

function Info($m) { Write-Host "==> $m" -ForegroundColor Cyan }
function Warn($m) { Write-Host "warn: $m" -ForegroundColor Yellow }

# ---- platform ----
$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'amd64' }
    'ARM64' { 'arm64' }
    default { throw "unsupported arch: $env:PROCESSOR_ARCHITECTURE (amd64 and arm64 only)" }
}

# ---- resolve version ----
$ver = $env:HERDR_VERSION
if (-not $ver) {
    Info 'resolving latest release'
    $ver = (Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest").tag_name
    if (-not $ver) { throw 'could not resolve latest release - set HERDR_VERSION=vX.Y.Z' }
}
$tag = if ($ver.StartsWith('v')) { $ver } else { "v$ver" }
$num = $tag.Substring(1)
$asset = "${project}_${num}_windows_${arch}.zip"
$base = "https://github.com/$repo/releases/download/$tag"

# ---- install dir ----
$bindir = if ($env:BINDIR) { $env:BINDIR } else { Join-Path $env:LOCALAPPDATA "Programs\$project" }
New-Item -ItemType Directory -Force -Path $bindir | Out-Null

# ---- download, verify, install ----
$tmp = Join-Path ([IO.Path]::GetTempPath()) ([IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
    Info "downloading $asset ($tag)"
    Invoke-WebRequest "$base/$asset" -OutFile "$tmp\$asset"

    $sums = "$tmp\checksums.txt"
    try { Invoke-WebRequest "$base/checksums.txt" -OutFile $sums } catch { $sums = $null }
    if (-not $sums) {
        Warn 'checksums.txt not available; skipping verification'
    } elseif ($line = Select-String -Path $sums -Pattern " $([regex]::Escape($asset))$" | Select-Object -First 1) {
        $want = ($line.Line -split '\s+')[0]
        # .NET directly: Get-FileHash is missing when powershell.exe inherits pwsh's PSModulePath.
        $sha = [Security.Cryptography.SHA256]::Create()
        try { $got = -join ($sha.ComputeHash([IO.File]::ReadAllBytes("$tmp\$asset")) | ForEach-Object { $_.ToString('x2') }) }
        finally { $sha.Dispose() }
        if ($got -ne $want) { throw "checksum mismatch for $asset" }
        Info 'checksum ok'
    } else {
        Warn "no checksum entry for $asset; skipping verification"
    }

    Expand-Archive "$tmp\$asset" -DestinationPath "$tmp\x"
    if (-not (Test-Path "$tmp\x\$bin")) { throw "archive did not contain $bin" }

    $target = Join-Path $bindir $bin
    $replaced = Test-Path $target
    if ($replaced) {
        # Through cmd so Windows PowerShell 5.1 does not turn the stderr line into an error.
        $oldver = cmd /c "`"$target`" -version 2>&1"
        Info "replacing existing $target ($oldver)"
        # A running exe cannot be overwritten but can be renamed out of the way.
        Remove-Item "$target.old" -Force -ErrorAction SilentlyContinue
        Move-Item $target "$target.old" -Force
    }
    Move-Item "$tmp\x\$bin" $target -Force
    Remove-Item "$target.old" -Force -ErrorAction SilentlyContinue
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}

if ($replaced) {
    Warn "if $project is already running (standalone or as a service), restart it to pick up this update"
}
Info "installed $project $num -> $target"

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$entries = @($userPath -split ';' | Where-Object { $_ })
if ($entries -notcontains $bindir) {
    [Environment]::SetEnvironmentVariable('Path', (($entries + $bindir) -join ';'), 'User')
    Warn "added $bindir to your user PATH - open a new terminal to use '$project'"
}

@"

Run the bridge (loopback only by default):
  $project

Run it in the background, started at every logon (a scheduled task):
  $project -service install
  $project -service start

Reach it from your phone over Tailscale (HTTPS = installable PWA + push):
  tailscale serve --bg --https=443 127.0.0.1:7331
  # then open https://<machine>.<tailnet>.ts.net on the phone
"@
