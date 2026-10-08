# Builds the browser version of the game into dist/web and zips it for upload.
# To use, run this command from the project's root directory: .\build_web.ps1
# To try it locally afterwards: cd src; go run ./cmd/webserve

$ErrorActionPreference = "Stop"

$root = $PSScriptRoot
$src  = "$root\src"
$out  = "$root\dist\web"
$zip  = "$root\dungeoneer-web.zip"

New-Item -ItemType Directory -Force $out | Out-Null

# wasm_exec.js must come from the same Go that builds the wasm. It moved from
# misc/wasm to lib/wasm in Go 1.24.
$goroot   = (go env GOROOT)
$wasmExec = "$goroot\lib\wasm\wasm_exec.js"
if (-not (Test-Path $wasmExec)) { $wasmExec = "$goroot\misc\wasm\wasm_exec.js" }
if (-not (Test-Path $wasmExec)) { throw "wasm_exec.js not found under $goroot" }

$prevOS   = $env:GOOS
$prevArch = $env:GOARCH
Push-Location $src
try {
  $env:GOOS   = "js"
  $env:GOARCH = "wasm"
  go build -o "$out\dungeoneer.wasm" .
  if ($LASTEXITCODE -ne 0) { throw "wasm build failed" }
} finally {
  $env:GOOS   = $prevOS
  $env:GOARCH = $prevArch
  Pop-Location
}

Copy-Item -Force "$src\web\index.html" "$out\index.html"
Copy-Item -Force $wasmExec "$out\wasm_exec.js"

Remove-Item -Force $zip -ErrorAction SilentlyContinue
Compress-Archive -Path "$out\*" -DestinationPath $zip

Write-Host "Built $out"
Write-Host "Created $zip"
