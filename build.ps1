# herdrweb Windows build: the `make build` equivalent for machines without make.
#
#   pwsh -File build.ps1            # web UI + bin\herdrweb.exe
#   pwsh -File build.ps1 -SkipWeb   # Go only (reuses internal\webui\dist)
param([switch]$SkipWeb)
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

$dist = 'internal\webui\dist'
if (-not $SkipWeb) {
    Push-Location web
    try {
        # ci, not install: builds from the lockfile without rewriting it.
        npm ci --no-audit --no-fund
        if ($LASTEXITCODE) { throw 'npm ci failed' }
        npm run build
        if ($LASTEXITCODE) { throw 'npm run build failed' }
    } finally { Pop-Location }
    Remove-Item -Recurse -Force $dist -ErrorAction SilentlyContinue
    Copy-Item -Recurse web\build $dist
    # The dist dir stays tracked through its .gitkeep (see AGENTS.md).
    git checkout -- "$dist/.gitkeep"
}

$version = git describe --tags --always --dirty 2>$null
if (-not $version) { $version = 'dev' }
$env:CGO_ENABLED = '0'
go build -trimpath -ldflags "-s -w -X main.version=$version" -o bin\herdrweb.exe .\cmd\herdr-bridge
if ($LASTEXITCODE) { throw 'go build failed' }
Write-Host "built bin\herdrweb.exe ($version)"
