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

# Build command - use current directory (portable for cloned repositories)
$currentDir = (Get-Location).Path.Replace('\', '/')
$buildCmd = @"
cd '$currentDir' && go build -ldflags='-s -w' -o absensi-server.exe ./cmd/server
"@

Write-Host "Building server..." -ForegroundColor Yellow

# Build server binary
& $msys2Path -mingw64 -defterm -no-start -c $buildCmd

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "❌ Build failed!" -ForegroundColor Red
    Write-Host "Check error messages above" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "✅ Server build successful!" -ForegroundColor Green

if (Test-Path "absensi-server.exe") {
    $fileSize = (Get-Item absensi-server.exe).Length / 1MB
    Write-Host "   Output: absensi-server.exe" -ForegroundColor Cyan
    Write-Host "   Size: $([math]::Round($fileSize, 2)) MB" -ForegroundColor Cyan
}

# Also build seed binary (needs MSYS2 for CGO/face dependencies)
Write-Host ""
Write-Host "Building seed tool..." -ForegroundColor Yellow

$seedBuildCmd = @"
cd '$currentDir' && go build -ldflags='-s -w' -o seed.exe ./cmd/seed
"@

& $msys2Path -mingw64 -defterm -no-start -c $seedBuildCmd

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Seed build successful!" -ForegroundColor Green
    
    if (Test-Path "seed.exe") {
        $seedSize = (Get-Item seed.exe).Length / 1MB
        Write-Host "   Output: seed.exe" -ForegroundColor Cyan
        Write-Host "   Size: $([math]::Round($seedSize, 2)) MB" -ForegroundColor Cyan
    }
    
    # Copy required MSYS2 runtime DLLs so binaries run outside MSYS2
    Write-Host ""
    Write-Host "Copying MSYS2 runtime DLLs..." -ForegroundColor Yellow
    $msys2Bin = "C:\msys64\mingw64\bin"
    $requiredDlls = @(
        "libgcc_s_seh-1.dll",
        "libstdc++-6.dll",
        "libwinpthread-1.dll",
        "libdlib.dll",
        "libopenblas.dll",
        "libjpeg-8.dll"
    )
    $copiedCount = 0
    foreach ($dll in $requiredDlls) {
        $src = Join-Path $msys2Bin $dll
        if (Test-Path $src) {
            Copy-Item -Path $src -Destination ".\$dll" -Force
            $copiedCount++
        }
    }
    Write-Host "   Copied $copiedCount DLLs to current directory" -ForegroundColor Cyan
    
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "All builds completed!" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "To run server:" -ForegroundColor Yellow
    Write-Host "  .\absensi-server.exe" -ForegroundColor White
    Write-Host ""
    Write-Host "To setup database & admin user:" -ForegroundColor Yellow
    Write-Host "  .\seed.exe" -ForegroundColor White
} else {
    Write-Host ""
    Write-Host "⚠️  Seed build failed!" -ForegroundColor Red
    Write-Host "  You can still run the server, but seed.exe won't be available." -ForegroundColor Yellow
    Write-Host "  To create admin user, run from MSYS2 MINGW64 shell:" -ForegroundColor Yellow
    Write-Host "  go run cmd/seed/main.go" -ForegroundColor White
}
