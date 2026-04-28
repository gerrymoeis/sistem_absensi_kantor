# Sistem Absensi Kantor dengan Face Recognition

Sistem absensi berbasis web dengan face recognition, IP restriction, dan comprehensive security features. Dirancang untuk memastikan karyawan benar-benar hadir di kantor saat melakukan absensi.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Production Ready](https://img.shields.io/badge/Status-Production%20Ready-brightgreen.svg)](https://github.com/gerrymoeis/sistem_absensi_kantor)
[![Security](https://img.shields.io/badge/Security-9%2F10-green.svg)](docs_and_backup/THOROUGH_CHECKUP_FINAL_REPORT.md)
[![Accuracy](https://img.shields.io/badge/Face%20Recognition-93.61%25-blue.svg)](docs_and_backup/PHASE5_POLISH_AND_TESTING.md)

---

## 🎯 Key Features

### Core Features
- ✅ **Face Recognition** - Login & attendance dengan face recognition (93.61% accuracy)
- ✅ **IP Restriction** - Hanya bisa diakses dari WiFi kantor (CIDR support)
- ✅ **JWT Authentication** - Secure token-based authentication
- ✅ **Clock In/Out** - Absensi masuk/pulang dengan timestamp
- ✅ **Status Kehadiran** - Hadir, Izin, Sakit, Cuti, Alpha
- ✅ **Admin Dashboard** - User management & statistics
- ✅ **Activity Logging** - Comprehensive audit trail
- ✅ **Excel Export** - Laporan absensi (All Data & Monthly)
- ✅ **Rate Limiting** - Brute force protection
- ✅ **Modern UI** - Responsive design dengan Tailwind CSS

### Face Recognition Features
- ✅ Face login (alternative to password)
- ✅ Face attendance (clock in/out)
- ✅ Admin face enrollment
- ✅ Multiple encodings per user
- ✅ Replay attack prevention
- ✅ Quality validation
- ✅ 93.61% accuracy (tested with 1000+ images)
- ✅ 74ms average response time
- ✅ 83% time savings vs manual

---

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- MINGW64 (untuk face recognition)

### Installation
```bash
# Clone repository
git clone https://github.com/gerrymoeis/sistem_absensi_kantor.git
cd sistem_absensi_kantor/main_folder

# Configure
cp .env.example .env
# Edit .env: JWT_SECRET, ALLOWED_IPS

# Build (tanpa face recognition)
go build -ldflags="-s -w" -o absensi-server.exe ./cmd/server

# Build (dengan face recognition - butuh MINGW64)
./build_server.ps1

# Create admin user
go run cmd/seed/main.go

# Run
./absensi-server.exe
```

Akses: `http://localhost:8080`  
Login: `admin` / `admin123` (ganti password setelah login!)

**Deployment Guide**: Lihat [`DEPLOYMENT_GUIDE_2026.md`](../docs_and_backup/DEPLOYMENT_GUIDE_2026.md)

---

## 📋 Table of Contents

- [Tech Stack](#tech-stack)
- [Security](#security)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Project Structure](#project-structure)
- [Troubleshooting](#troubleshooting)
- [Documentation](#documentation)

---

## 🛠 Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go 1.21+ + Gin Framework |
| Database | SQLite 3.x |
| Authentication | JWT + bcrypt |
| Face Recognition | dlib + go-face |
| Frontend | HTML + Tailwind CSS 3.x |
| Excel Export | Excelize v2.10.1 |

**Why This Stack?**
- **Go**: Fast, single binary deployment (~26-95 MB)
- **SQLite**: Zero-configuration, embedded database
- **dlib**: State-of-the-art face recognition (93.61% accuracy)
- **Tailwind**: Rapid UI development

---

## 🔒 Security

**Security Score**: 9/10 ([Full Report](../docs_and_backup/THOROUGH_CHECKUP_FINAL_REPORT.md))

### Security Layers
1. **IP Restriction** - CIDR-based network filtering
2. **Rate Limiting** - Brute force protection (5/min login, 300/min API)
3. **Authentication** - JWT with 24h expiration
4. **Authorization** - Role-based access control (RBAC)
5. **Security Headers** - XSS, Clickjacking, MIME-sniffing protection

### Key Features
- Bcrypt password hashing (cost 10)
- Parameterized SQL queries (SQL injection prevention)
- Account locking (5 failed attempts = 15 min)
- Replay attack prevention (SHA-256 hash)
- Comprehensive audit logging
- Soft delete (data preservation)

---

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `0.0.0.0` | Server bind address |
| `SERVER_PORT` | `8080` | Server port |
| `SERVER_MODE` | `release` | Gin mode (debug/release) |
| `ENVIRONMENT` | `production` | Affects rate limiting |
| `JWT_SECRET` | - | **WAJIB diganti!** |
| `ALLOWED_IPS` | - | Network CIDR (comma-separated) |
| `FACE_RECOGNITION_ENABLED` | `false` | Enable face features |
| `FACE_MATCH_THRESHOLD` | `0.6` | Recognition threshold (0.0-1.0) |

### IP Restriction Examples
```env
# Single IP
ALLOWED_IPS=192.168.1.100/32

# Network range (recommended)
ALLOWED_IPS=192.168.1.0/24

# Multiple ranges
ALLOWED_IPS=192.168.1.0/24,10.0.0.0/8

# IPv6 support
ALLOWED_IPS=::1/128,2001:db8::/32
```

**Find your network:**
```bash
# Windows
ipconfig

# Linux/Mac
ifconfig
```

---

## 📡 API Documentation

### Authentication
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

### Face Login
```http
POST /api/auth/login-face
Content-Type: application/json

{
  "image": "base64-encoded-image"
}
```

### Clock In
```http
POST /api/absensi/masuk
Authorization: Bearer <token>

{
  "status": "hadir",
  "keterangan": "Optional"
}
```

### Face Attendance
```http
POST /api/face/recognize
Authorization: Bearer <token>

{
  "image": "base64-encoded-image"
}
```

**Full API Documentation**: [`FACE_RECOGNITION_API_DOCUMENTATION.md`](../docs_and_backup/FACE_RECOGNITION_API_DOCUMENTATION.md)

---

## 📁 Project Structure

```
main_folder/
├── cmd/
│   ├── server/          # Main application
│   └── seed/            # Database seeding
├── internal/
│   ├── config/          # Configuration
│   ├── database/        # Database & migrations
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # Auth, IP restriction, rate limit
│   ├── model/           # Data models
│   ├── repository/      # Database operations
│   └── service/         # Business logic
├── web/
│   ├── static/          # CSS, JS, images
│   └── templates/       # HTML templates
├── models/              # Face recognition models
├── data/                # SQLite database
├── logs/                # Application logs
└── .env                 # Configuration
```

---

## 🐛 Troubleshooting

### Camera tidak berfungsi
- ✅ Pastikan HTTPS enabled (wajib untuk getUserMedia)
- ✅ Check browser permissions (allow camera)
- ✅ Gunakan Chrome/Edge (recommended)
- ✅ Camera tidak digunakan aplikasi lain

### Face tidak dikenali
- ✅ Enrollment dengan foto berkualitas baik
- ✅ Lighting cukup terang
- ✅ Wajah terlihat jelas (tidak blur)
- ✅ Coba adjust `FACE_MATCH_THRESHOLD` (default: 0.6)

### "Access denied: IP not allowed"
```bash
# Check your IP
ipconfig  # Windows
ifconfig  # Linux/Mac

# Update ALLOWED_IPS in .env
ALLOWED_IPS=192.168.1.0/24

# Restart server
```

### Port already in use
```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux
lsof -ti:8080 | xargs kill -9
```

### Build error (CGO)
- ✅ Gunakan MINGW64 terminal (bukan PowerShell)
- ✅ Install dlib via MSYS2: `pacman -S mingw-w64-x86_64-dlib`
- ✅ Gunakan `build_server.ps1` script

---

## 📚 Documentation

### Face Recognition
- [`DEPLOYMENT_GUIDE_2026.md`](../docs_and_backup/DEPLOYMENT_GUIDE_2026.md) - Deployment guide
- [`FACE_RECOGNITION_API_DOCUMENTATION.md`](../docs_and_backup/FACE_RECOGNITION_API_DOCUMENTATION.md) - API reference
- [`FACE_RECOGNITION_MANUAL_INSTALLATION_GUIDE.md`](../docs_and_backup/FACE_RECOGNITION_MANUAL_INSTALLATION_GUIDE.md) - Installation guide
- [`PHASE5_POLISH_AND_TESTING.md`](../docs_and_backup/PHASE5_POLISH_AND_TESTING.md) - Testing results

### Project Status
- [`FINAL_STATUS.md`](../docs_and_backup/FINAL_STATUS.md) - Current status
- [`THOROUGH_CHECKUP_FINAL_REPORT.md`](../docs_and_backup/THOROUGH_CHECKUP_FINAL_REPORT.md) - Complete check up report
- [`IMPLEMENTATION_PROGRESS_SUMMARY.md`](../docs_and_backup/IMPLEMENTATION_PROGRESS_SUMMARY.md) - Implementation details
- [`INDEX.md`](../docs_and_backup/INDEX.md) - Documentation index (110+ files)

---

## 📊 Performance

| Metric | Value |
|--------|-------|
| Startup Time | ~80ms |
| Memory Usage | ~15 MB (idle) |
| Face Recognition | 74ms average |
| Recognition Accuracy | 93.61% |
| Time Savings | 83% vs manual |
| Binary Size | 26-95 MB |
| Throughput | ~6000 req/sec |

---

## 📝 Changelog

### v2.0 (Current) - Face Recognition Complete
- ✅ Face recognition login & attendance
- ✅ Admin face enrollment
- ✅ 93.61% accuracy (tested with 1000+ images)
- ✅ 83% time savings
- ✅ Production ready & scalable

### v1.0 - Core Features
- ✅ JWT authentication
- ✅ IP restriction
- ✅ Clock in/out
- ✅ Admin dashboard
- ✅ Excel export

---

## 🤝 Contributing

This is a proprietary project for internal use.

---

## 📞 Support

- **Documentation**: Check [`docs_and_backup/`](../docs_and_backup/) folder
- **Issues**: Check logs first (`logs/app.log`)
- **Deployment**: See [`DEPLOYMENT_GUIDE_2026.md`](../docs_and_backup/DEPLOYMENT_GUIDE_2026.md)

---

## 📄 License

Proprietary - Internal Use Only

---

**Built with Go | Production Ready | Security: 9/10 | Accuracy: 93.61%**

