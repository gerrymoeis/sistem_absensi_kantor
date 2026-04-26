# Build Script - Face Recognition Support
# Hanya mempengaruhi build ini, tidak mengubah system environment

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Building Absensi Server with Face Recognition" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Set CGO flags (temporary, hanya untuk proses ini)
$env:CGO_ENABLED = "1"
$env:CGO_CPPFLAGS = "-IC:\msys64\mingw64\include"
$env:CGO_LDFLAGS = "-LC:\msys64\mingw64\lib -ldlib -lopenblas -llapack"

Write-Host "CGO Configuration:" -ForegroundColor Yellow
Write-Host "  CGO_ENABLED    : $env:CGO_ENABLED"
Write-Host "  CGO_CPPFLAGS   : $env:CGO_CPPFLAGS"
Write-Host "  CGO_LDFLAGS    : $env:CGO_LDFLAGS"
Write-Host ""

# Build
Write-Host "Building..." -ForegroundColor Green
go build -ldflags="-s -w" -o absensi-server.exe ./cmd/server

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Build successful!" -ForegroundColor Green
    Write-Host "   Output: absensi-server.exe"
    
    # Show file size
    $fileSize = (Get-Item absensi-server.exe).Length / 1MB
    Write-Host "   Size: $([math]::Round($fileSize, 2)) MB"
} else {
    Write-Host ""
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}
