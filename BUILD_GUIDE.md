# Build Guide - Absensi Kantor Lokal

## Build Scripts

Project ini memiliki 2 build script:

### 1. `build.ps1` - Build Normal (Tanpa Face Recognition)

Build standard untuk production tanpa fitur face recognition.

**Cara pakai:**
```powershell
.\build.ps1
```

**Output:**
- `absensi-server.exe` - Server utama
- `seed.exe` - Database seeder
- `generate-secret.exe` - JWT secret generator

**Keuntungan:**
- ✅ Binary lebih kecil (~26 MB)
- ✅ Tidak butuh external dependencies
- ✅ Lebih cepat compile
- ✅ Cocok untuk deployment tanpa face recognition

---

### 2. `build_face.ps1` - Build dengan Face Recognition

Build dengan dukungan face recognition menggunakan dlib.

**Cara pakai:**
```powershell
.\build_face.ps1
```

**Output:**
- `absensi-server.exe` - Server dengan face recognition support

**Requirements:**
- MSYS2 terinstall di `C:\msys64`
- dlib terinstall via pacman
- MinGW64 di PATH

**Keuntungan:**
- ✅ Fitur face recognition aktif
- ✅ Tidak mengubah system environment
- ✅ Portable (developer lain tinggal jalankan script)
- ✅ Bisa switch antara build biasa dan face recognition

**Catatan:**
- Binary akan lebih besar (~75-95 MB)
- Butuh DLL dependencies saat runtime

---

## Environment Variables

### ❌ JANGAN Set di System

**JANGAN lakukan ini:**
```powershell
# JANGAN! Ini mempengaruhi semua project Go
[System.Environment]::SetEnvironmentVariable("CGO_CPPFLAGS", "...", [System.EnvironmentVariableTarget]::User)
```

### ✅ Gunakan Build Script

**Lakukan ini:**
```powershell
# BAIK! Hanya mempengaruhi build ini
.\build_face.ps1
```

Build script akan set environment variables **temporary** hanya untuk proses build tersebut.

---

## Perbandingan

| Aspek | build.ps1 | build_face.ps1 |
|-------|-----------|----------------|
| **Binary Size** | ~26 MB | ~75-95 MB |
| **Dependencies** | Tidak ada | dlib, OpenBLAS, LAPACK |
| **Compile Time** | ~10 detik | ~30-60 detik |
| **Face Recognition** | ❌ Tidak | ✅ Ya |
| **System Impact** | Tidak ada | Tidak ada (local only) |
| **Portable** | ✅ Ya | ✅ Ya (dengan DLLs) |

---

## Development Workflow

### Scenario 1: Development Tanpa Face Recognition

```powershell
# Build
.\build.ps1

# Run
.\absensi-server.exe
```

### Scenario 2: Development Dengan Face Recognition

```powershell
# Build
.\build_face.ps1

# Run (pastikan DLLs accessible)
.\absensi-server.exe
```

### Scenario 3: Testing Both Versions

```powershell
# Build normal
.\build.ps1
Rename-Item absensi-server.exe absensi-server-normal.exe

# Build face recognition
.\build_face.ps1
Rename-Item absensi-server.exe absensi-server-face.exe

# Test normal version
.\absensi-server-normal.exe

# Test face version
.\absensi-server-face.exe
```

---

## Deployment

### Production Tanpa Face Recognition

```powershell
# Build
.\build.ps1

# Copy ke production server
Copy-Item absensi-server.exe \\production-server\app\
Copy-Item .env \\production-server\app\
Copy-Item -Recurse web \\production-server\app\
```

### Production Dengan Face Recognition

```powershell
# Build
.\build_face.ps1

# Copy binary + dependencies
Copy-Item absensi-server.exe \\production-server\app\
Copy-Item .env \\production-server\app\
Copy-Item -Recurse web \\production-server\app\
Copy-Item -Recurse models \\production-server\app\

# Copy DLLs (dari MSYS2)
Copy-Item C:\msys64\mingw64\bin\libdlib.dll \\production-server\app\
Copy-Item C:\msys64\mingw64\bin\libopenblas.dll \\production-server\app\
Copy-Item C:\msys64\mingw64\bin\libgfortran-*.dll \\production-server\app\
Copy-Item C:\msys64\mingw64\bin\libquadmath-*.dll \\production-server\app\
Copy-Item C:\msys64\mingw64\bin\libgcc_s_seh-*.dll \\production-server\app\
Copy-Item C:\msys64\mingw64\bin\libwinpthread-*.dll \\production-server\app\
```

---

## Troubleshooting

### Error: "CGO not enabled"

**Solusi:** Gunakan `build_face.ps1` yang sudah set `CGO_ENABLED=1`

### Error: "dlib not found"

**Solusi:** 
1. Pastikan MSYS2 terinstall di `C:\msys64`
2. Pastikan dlib terinstall: `pacman -S mingw-w64-x86_64-dlib`
3. Pastikan MinGW64 di PATH

### Error: "cannot find -ldlib"

**Solusi:** Edit `build_face.ps1`, sesuaikan path:
```powershell
$env:CGO_LDFLAGS = "-LC:\msys64\mingw64\lib -ldlib -lopenblas -llapack"
```

### Binary terlalu besar

**Solusi:** Build script sudah menggunakan `-ldflags="-s -w"` untuk strip debug symbols. Jika masih terlalu besar, gunakan UPX:
```powershell
upx --best absensi-server.exe
```

---

## Best Practices

### ✅ DO:
- Gunakan build script untuk compile
- Commit build script ke Git
- Dokumentasikan dependencies
- Test di clean environment sebelum deploy

### ❌ DON'T:
- Jangan set CGO flags di system environment
- Jangan commit binary ke Git
- Jangan hardcode absolute paths
- Jangan lupa copy DLLs saat deploy

---

## FAQ

**Q: Kenapa ada 2 build script?**  
A: Agar bisa build dengan/tanpa face recognition tanpa mengubah system environment.

**Q: Apakah harus install MSYS2 untuk build normal?**  
A: Tidak. `build.ps1` tidak butuh MSYS2. Hanya `build_face.ps1` yang butuh.

**Q: Apakah environment variables di build script mempengaruhi sistem?**  
A: Tidak. Variables hanya berlaku untuk proses build tersebut, tidak permanent.

**Q: Bagaimana cara switch antara build biasa dan face recognition?**  
A: Tinggal jalankan script yang sesuai. Tidak perlu cleanup atau konfigurasi tambahan.

**Q: Apakah bisa build face recognition di CI/CD?**  
A: Ya, tapi CI/CD environment harus punya MSYS2 + dlib terinstall.

---

**Last Updated:** 24 April 2026
