# Build Script for Absensi Kantor Lokal
# Builds optimized binaries with stripped debug symbols

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Building Absensi Kantor Lokal" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Clean old binaries
Write-Host "Cleaning old binaries..." -ForegroundColor Yellow
Remove-Item -Path "*.exe" -ErrorAction SilentlyContinue
Write-Host "✓ Old binaries removed`n" -ForegroundColor Green

# Build server (production)
Write-Host "Building absensi-server.exe..." -ForegroundColor Yellow
go build -ldflags="-s -w" -o absensi-server.exe ./cmd/server
if ($LASTEXITCODE -eq 0) {
    $size = [math]::Round((Get-Item absensi-server.exe).Length/1MB,2)
    Write-Host "✓ absensi-server.exe: $size MB" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to build absensi-server.exe" -ForegroundColor Red
    exit 1
}

# Build seed (development)
Write-Host "Building seed.exe..." -ForegroundColor Yellow
go build -ldflags="-s -w" -o seed.exe ./cmd/seed
if ($LASTEXITCODE -eq 0) {
    $size = [math]::Round((Get-Item seed.exe).Length/1MB,2)
    Write-Host "✓ seed.exe: $size MB" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to build seed.exe" -ForegroundColor Red
    exit 1
}

# Build generate-secret (utility)
Write-Host "Building generate-secret.exe..." -ForegroundColor Yellow
go build -ldflags="-s -w" -o generate-secret.exe ./cmd/generate_secret
if ($LASTEXITCODE -eq 0) {
    $size = [math]::Round((Get-Item generate-secret.exe).Length/1MB,2)
    Write-Host "✓ generate-secret.exe: $size MB" -ForegroundColor Green
} else {
    Write-Host "✗ Failed to build generate-secret.exe" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Build Complete!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Show summary
Write-Host "Binary Sizes:" -ForegroundColor Cyan
Get-ChildItem *.exe | Select-Object Name, @{Name="Size(MB)";Expression={[math]::Round($_.Length/1MB,2)}} | Format-Table -AutoSize

$totalSize = [math]::Round((Get-ChildItem *.exe | Measure-Object -Property Length -Sum).Sum/1MB,2)
Write-Host "Total Size: $totalSize MB" -ForegroundColor Cyan
Write-Host ""
Write-Host "To run the server: .\absensi-server.exe" -ForegroundColor Yellow
Write-Host "To seed database: .\seed.exe" -ForegroundColor Yellow
Write-Host "To generate JWT secret: .\generate-secret.exe" -ForegroundColor Yellow
