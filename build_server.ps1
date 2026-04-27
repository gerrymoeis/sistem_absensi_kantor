# Build Script for Absensi Server with Face Recognition
# Uses MINGW64 for CGO compilation

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Building Absensi Server" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if MSYS2 is installed
$msys2Path = "C:\msys64\msys2_shell.cmd"
if (-not (Test-Path $msys2Path)) {
    Write-Host "ERROR: MSYS2 not found at $msys2Path" -ForegroundColor Red
    Write-Host "Please install MSYS2 first" -ForegroundColor Yellow
    exit 1
}

Write-Host "Using MSYS2 MINGW64 for CGO compilation..." -ForegroundColor Green
Write-Host ""

# Build command
$buildCmd = @"
cd '/d/Gerry/Programmer/Best Terbaik(2026)/Experiments/absensi_kantor_lokal/main_folder' && go build -ldflags='-s -w' -o absensi-server.exe ./cmd/server
"@

Write-Host "Building server..." -ForegroundColor Yellow

# Execute in MINGW64 environment
& $msys2Path -mingw64 -defterm -no-start -c $buildCmd

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Build successful!" -ForegroundColor Green
    
    if (Test-Path "absensi-server.exe") {
        $fileSize = (Get-Item absensi-server.exe).Length / 1MB
        Write-Host "   Output: absensi-server.exe" -ForegroundColor Cyan
        Write-Host "   Size: $([math]::Round($fileSize, 2)) MB" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "To run server:" -ForegroundColor Yellow
        Write-Host "  .\absensi-server.exe" -ForegroundColor White
    }
} else {
    Write-Host ""
    Write-Host "❌ Build failed!" -ForegroundColor Red
    Write-Host "Check error messages above" -ForegroundColor Yellow
    exit 1
}
