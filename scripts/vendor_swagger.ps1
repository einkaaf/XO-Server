$ErrorActionPreference = "Stop"

$dest = Join-Path $PSScriptRoot "..\docs\swagger-ui"
New-Item -ItemType Directory -Force $dest | Out-Null

$base = "https://unpkg.com/swagger-ui-dist@5"
$files = @(
  "swagger-ui.css",
  "swagger-ui-bundle.js",
  "swagger-ui-standalone-preset.js",
  "favicon-16x16.png",
  "favicon-32x32.png"
)

foreach ($f in $files) {
  $url = "$base/$f"
  $out = Join-Path $dest $f
  Write-Host "Downloading $url -> $out"
  Invoke-WebRequest -Uri $url -OutFile $out
}

Write-Host "Swagger UI assets downloaded to $dest"
