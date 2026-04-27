# Sistem Absensi Kantor

Sistem absensi berbasis web yang secure dan robust, hanya dapat diakses dari jaringan WiFi kantor menggunakan IP restriction. Dirancang untuk memastikan karyawan benar-benar hadir di kantor saat melakukan absensi.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Production Ready](https://img.shields.io/badge/Status-Production%20Ready-brightgreen.svg)](docs_and_backup/PRODUCTION_READY_FINAL.md)
[![Security](https://img.shields.io/badge/Security-9%2F10-green.svg)](docs_and_backup/security_audit_report.md)
[![License](https://img.shields.io/badge/License-Proprietary-red.svg)](LICENSE)

---

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Security](#security)
- [Quick Start](#quick-start)
- [Production Deployment](#production-deployment)
- [Project Structure](#project-structure)
- [API Documentation](#api-documentation)
- [Configuration](#configuration)
- [Development](#development)
- [Troubleshooting](#troubleshooting)

---

## Features

### Core Features (Phase 1 & 2 - COMPLETE)
- **IP Restriction** - Hanya bisa diakses dari WiFi kantor (CIDR support)
- **User Authentication** - JWT + bcrypt password hashing
- **Absen Masuk/Pulang** - Clock in/out dengan timestamp
- **Status Kehadiran** - Hadir, Izin, Sakit, Cuti, Alpha
- **Keterangan Detail** - Textarea untuk alasan izin/sakit/cuti
- **Riwayat Absensi** - View attendance history
- **Admin Dashboard** - Monitor semua karyawan dengan statistics
- **User Management** - CRUD users (admin only)
- **Activity Logging** - Comprehensive audit trail
- **Excel Export** - Export laporan ke Excel (All Data & Monthly)
- **Modern UI** - Responsive design dengan Tailwind CSS
- **Rate Limiting** - Protection against brute force attacks
- **Security Headers** - XSS, Clickjacking, MIME-sniffing protection

### Face Recognition Features (Week 2 Day 5 - COMPLETE) ✨
- **Face Enrollment** - Admin can enroll employee faces
- **Face Recognition** - Recognize faces for attendance (93.61% accuracy)
- **Multiple Encodings** - Support multiple face encodings per user
- **Liveness Detection** - Anti-spoofing protection
- **Replay Attack Prevention** - SHA-256 hash-based duplicate detection
- **Quality Checks** - Minimum resolution and face detection validation
- **Statistics Dashboard** - Monitor enrollment and recognition metrics
- **Production Ready** - Tested with 1000+ images, 200+ people
- **High Performance** - 74ms average response time
- **Scalable** - Linear scaling, handles 100+ concurrent users

### Security Features
- Multi-layer security (IP → Auth → Role)
- Password hashing dengan bcrypt (cost 10)
- JWT authentication dengan 24h expiration
- SQL injection prevention (parameterized queries)
- XSS protection (security headers)
- Role-based access control (RBAC)
- Rate limiting (5 req/min login, 60 req/min API)
- Comprehensive audit logging
- Soft delete (data preservation)
- Trusted proxies configuration

---

## Tech Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| **Backend** | Go + Gin | 1.25+ |
| **Database** | SQLite | 3.x |
| **Authentication** | JWT + bcrypt | - |
| **Face Recognition** | dlib + go-face | 1.0.0 |
| **Frontend** | HTML + Tailwind CSS | 3.x |
| **Excel Export** | Excelize | v2.10.1 |
| **Deployment** | Single binary | - |

### Why This Stack?
- **Go**: Fast, compiled, single binary deployment (26.58 MB)
- **SQLite**: Zero-configuration, embedded database
- **Gin**: High-performance HTTP framework
- **JWT**: Stateless authentication
- **Tailwind**: Rapid UI development
- **Excelize**: Professional Excel reports

---

## Security

### Security Score: 9/10
See [Security Audit Report](../docs_and_backup/security_audit_report.md) for details.

### Security Layers
1. **IP Restriction** - First line of defense (CIDR-based)
2. **Rate Limiting** - Brute force protection
3. **Authentication** - JWT with 24h expiration
4. **Authorization** - Role-based access control
5. **Security Headers** - XSS, Clickjacking protection

### Key Security Features
- Bcrypt password hashing (cost 10)
- Parameterized SQL queries (SQL injection prevention)
- Security headers (XSS, Clickjacking, MIME-sniffing)
- JWT signature verification
- Role-based access control (RBAC)
- Rate limiting (login: 5/min, API: 60/min)
- Comprehensive audit logging
- Soft delete (data preservation)
- Trusted proxies configuration
- Release mode by default (no debug info leak)

---

## Quick Start

### Prerequisites
- Go 1.25 or higher
- Git

### 1. Clone Repository

```bash
git clone https://github.com/gerrymoeis/sistem_absensi_kantor.git
cd sistem_absensi_kantor/main_folder
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configuration

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

Edit `.env` for your network:
```env
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_MODE=release  # Use 'debug' for development

# Security - IMPORTANT: Change JWT_SECRET in production!
JWT_SECRET=change-this-secret-key-in-production
ALLOWED_IPS=127.0.0.1/32,::1/128,192.168.1.0/24  # Your WiFi network

# Database
DB_DRIVER=sqlite
DB_DSN=./data/absensi.db

# Logging
LOG_LEVEL=info
LOG_FILE=./logs/app.log
```

### 4. Build Application

**For Standard Build (without Face Recognition):**
```bash
go build -ldflags="-s -w" -o absensi-server.exe ./cmd/server
```

**For Face Recognition Build (requires MSYS2 + dlib):**
```bash
# Use MINGW64 environment for CGO compilation
./build_server.ps1
```

**Note**: Face recognition requires additional setup. See [Face Recognition Setup Guide](../docs_and_backup/FACE_RECOGNITION_MANUAL_INSTALLATION_GUIDE.md)

### 5. Create Admin User

```bash
go run cmd/seed/main.go
```

**Default Admin:**
- Username: `admin`
- Password: `admin123`

**IMPORTANT**: Change admin password after first login!

### 6. Run Server

```bash
./absensi-server.exe
```

Server akan berjalan di `http://localhost:8080`

### 7. Access Application

Open browser: `http://localhost:8080`

---

## Production Deployment

### Step 1: Generate JWT Secret

```bash
# Linux/Mac
openssl rand -base64 32

# Windows PowerShell
[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Maximum 256 }))
```

### Step 2: Configure Production Environment

Create `.env` file:
```env
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_MODE=release

# IMPORTANT: Use generated secret from Step 1
JWT_SECRET=<your-generated-secret-here>

# Your office WiFi network CIDR
# Example: 192.168.1.0/24 allows 192.168.1.1 - 192.168.1.254
ALLOWED_IPS=192.168.1.0/24

DB_DRIVER=sqlite
DB_DSN=./data/absensi.db

LOG_LEVEL=info
LOG_FILE=./logs/app.log
```

### Step 3: Build for Production

**Standard Build (without Face Recognition):**
```bash
# Windows
go build -ldflags="-s -w" -o absensi-server.exe ./cmd/server

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o absensi-server ./cmd/server
```

**Face Recognition Build (requires MSYS2 + dlib):**
```bash
# Windows (MINGW64 environment)
./build_server.ps1

# Linux (requires dlib installed)
CGO_ENABLED=1 go build -ldflags="-s -w" -o absensi-server ./cmd/server
```

**Build Flags Explanation:**
- `-ldflags="-s -w"` - Strip debug info (reduces size by 30%)
- Standard build: ~26 MB
- Face recognition build: ~75-95 MB (includes dlib)

### Step 4: Create Admin User

```bash
go run cmd/seed/main.go
```

### Step 5: Run Server

```bash
# Windows
./absensi-server.exe

# Linux
./absensi-server
```

### Step 6: Verify Deployment

Check server output:
```
2026/04/17 10:27:19 Server starting on 0.0.0.0:8080
```

No warnings = Production ready!

### Step 7: Change Admin Password

1. Login as admin
2. Go to Admin Dashboard → User Management
3. Reset admin password
4. Use strong password (min 8 chars, mixed case, numbers, symbols)

---

## Project Structure

```
main_folder/
├── cmd/
│   ├── server/              # Main application entry point
│   │   └── main.go
│   ├── seed/                # Database seeding
│   │   └── main.go
│   └── generate_secret/     # JWT secret generator
│       └── main.go
├── internal/
│   ├── config/              # Configuration management
│   │   └── config.go
│   ├── database/            # Database connection & migrations
│   │   ├── database.go
│   │   └── migrations.go
│   ├── handler/             # HTTP request handlers
│   │   ├── auth_handler.go
│   │   ├── absensi_handler.go
│   │   ├── admin_handler.go
│   │   └── export_handler.go
│   ├── middleware/          # HTTP middlewares
│   │   ├── auth.go          # JWT authentication
│   │   ├── admin.go         # Admin authorization
│   │   ├── ip_restriction.go # IP whitelist
│   │   └── rate_limit.go    # Rate limiting
│   ├── model/               # Data models
│   │   ├── user.go
│   │   ├── absensi.go
│   │   └── activity_log.go
│   ├── repository/          # Database operations
│   │   ├── user_repository.go
│   │   ├── absensi_repository.go
│   │   ├── activity_log_repository.go
│   │   └── admin_repository.go
│   └── service/             # Business logic
│       ├── auth_service.go
│       ├── absensi_service.go
│       ├── activity_log_service.go
│       ├── admin_service.go
│       ├── user_service.go
│       └── export_service.go
├── web/
│   ├── static/              # Static assets (CSS, JS, images)
│   └── templates/           # HTML templates
│       ├── login.html
│       ├── dashboard.html
│       ├── history.html
│       └── admin_dashboard.html
├── data/                    # SQLite database (gitignored)
│   └── absensi.db
├── logs/                    # Application logs (gitignored)
│   └── app.log
├── .env                     # Environment variables (gitignored)
├── .env.example             # Environment variables template
├── .gitignore
├── build.ps1                # Build script (optimized)
├── go.mod
├── go.sum
└── README.md
```

---

## API Documentation

### Authentication Endpoints

#### Login
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}

Response: 200 OK
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "admin",
    "full_name": "Administrator",
    "role": "admin",
    "is_active": true
  }
}
```

### Attendance Endpoints

#### Clock In
```http
POST /api/absensi/masuk
Authorization: Bearer <token>
Content-Type: application/json

{
  "status": "hadir",  # hadir, izin, sakit, cuti, alpha
  "keterangan": "Optional note (required for non-hadir)"
}

Response: 200 OK
{
  "message": "Berhasil absen masuk",
  "data": { ... }
}
```

#### Clock Out
```http
POST /api/absensi/pulang
Authorization: Bearer <token>
Content-Type: application/json

{
  "keterangan": "Optional note"
}

Response: 200 OK
{
  "message": "Berhasil absen pulang",
  "data": { ... }
}
```

### Admin Endpoints (Admin Only)

#### Export Excel - All Data
```http
GET /api/admin/export/excel
Authorization: Bearer <admin-token>

Response: 200 OK (Excel file download)
Filename: Laporan_Absensi_All_2026-04-17.xlsx
```

#### Export Excel - Monthly
```http
GET /api/admin/export/excel/monthly?year=2026&month=4
Authorization: Bearer <admin-token>

Response: 200 OK (Excel file download)
Filename: Laporan_Absensi_April_2026.xlsx
```

**Excel Report Features:**
- 8 columns: No, Nama, Tanggal, Jam Masuk, Jam Pulang, Durasi, Status, Keterangan
- Professional formatting with colors and borders
- Auto-calculated duration
- Footer with total records and timestamp
- Optimized column widths

For complete API documentation, see [API Documentation](../docs_and_backup/api_documentation.md)

---

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_HOST` | Server bind address | `0.0.0.0` | No |
| `SERVER_PORT` | Server port | `8080` | No |
| `SERVER_MODE` | Gin mode (debug/release) | `release` | No |
| `JWT_SECRET` | JWT signing secret | - | **Yes** |
| `ALLOWED_IPS` | Comma-separated CIDR list | - | **Yes** |
| `DB_DRIVER` | Database driver | `sqlite` | No |
| `DB_DSN` | Database connection string | `./data/absensi.db` | No |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | `info` | No |
| `LOG_FILE` | Log file path | `./logs/app.log` | No |

### IP Restriction Configuration

**CIDR Notation Examples:**
```env
# Single IP (IPv4)
ALLOWED_IPS=192.168.1.100/32

# IP Range (IPv4) - Recommended for WiFi network
ALLOWED_IPS=192.168.1.0/24  # Allows 192.168.1.1 - 192.168.1.254

# Multiple ranges
ALLOWED_IPS=192.168.1.0/24,10.0.0.0/8,172.16.0.0/12

# IPv6 support
ALLOWED_IPS=::1/128,2001:db8::/32

# Mixed IPv4 and IPv6
ALLOWED_IPS=127.0.0.1/32,::1/128,192.168.1.0/24
```

**How to Find Your Network CIDR:**
```bash
# Windows
ipconfig

# Linux/Mac
ifconfig
ip addr show
```

Look for your local IP (e.g., 192.168.1.100), then use `/24` for the whole network.

---

## Development

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/gerrymoeis/sistem_absensi_kantor.git
cd sistem_absensi_kantor/main_folder

# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Edit .env for development
# Set SERVER_MODE=debug for verbose logging

# Create seed data
go run cmd/seed/main.go

# Run server
go run cmd/server/main.go
```

### Development Commands

```bash
# Run server
go run cmd/server/main.go

# Build (optimized)
go build -ldflags="-s -w" -o absensi-server.exe ./cmd/server

# Build all executables
./build.ps1

# Run tests
go test ./...

# Format code
go fmt ./...

# Check dependencies
go mod tidy
go mod verify
```

### Code Style

Follow Go best practices:
- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Write clean, minimal code (less is more)
- Add comments for exported functions
- Use meaningful variable names
- Keep functions small and focused

---

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) DEFAULT 'employee',
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Absensi Table
```sql
CREATE TABLE absensi (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    tanggal DATE NOT NULL,
    jam_masuk TIME,
    jam_pulang TIME,
    status VARCHAR(20) DEFAULT 'hadir',  # hadir, izin, sakit, cuti, alpha
    keterangan TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE(user_id, tanggal)
);
```

### Activity Logs Table
```sql
CREATE TABLE activity_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    action_type VARCHAR(50) NOT NULL,
    description TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    status VARCHAR(20) DEFAULT 'success',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

---

## Troubleshooting

### Common Issues

**1. "Access denied: IP not allowed"**

**Solution:**
```bash
# Check your IP address
# Windows
ipconfig

# Linux/Mac
ifconfig
curl ifconfig.me

# Update ALLOWED_IPS in .env
ALLOWED_IPS=192.168.1.0/24  # Your network

# Restart server
```

**2. "Invalid or expired token"**

**Solution:**
- Token expires after 24 hours
- Login again to get new token
- Check JWT_SECRET is consistent across restarts

**3. "Port already in use"**

**Solution:**
```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/Mac
lsof -ti:8080 | xargs kill -9

# Or change port in .env
SERVER_PORT=8081
```

**4. "Database locked"**

**Solution:**
- SQLite doesn't support high concurrent writes
- Check no other process is using the database
- For high concurrency, consider PostgreSQL

**5. "Keterangan wajib diisi untuk status IZIN"**

**Solution:**
- This is expected behavior
- Keterangan is required for: Izin, Sakit, Cuti, Alpha
- Keterangan is optional for: Hadir

**6. "File download tidak berfungsi"**

**Solution:**
- Check browser allows downloads
- Check Authorization header is sent
- Check admin role permissions
- Check server logs for errors

---

## Performance

### Metrics (Production Mode)
- **Startup Time**: ~80ms
- **Memory Usage**: ~15 MB (idle)
- **Request Latency**: <10ms (local network)
- **Throughput**: ~6000 req/sec (single core)
- **Binary Size**: 26.58 MB (optimized)

### Optimization Applied
- Release mode by default (no debug overhead)
- Binary stripped of debug symbols (-30% size)
- Optimized middleware stack (5-7 handlers)
- Efficient rate limiter (pre-allocated memory)
- No verbose logging in production

---

## License

Proprietary - Internal Use Only

© 2026 Your Company. All rights reserved.

---

## Contributors

- **Development Team** - Initial work and maintenance

---

## Support

For issues and questions:
- Create an issue on GitHub
- Contact: admin@yourcompany.com

---

## Changelog

### Version 3.0 (Current) - Face Recognition Complete ✨
- Face Recognition System (93.61% accuracy)
- Face Enrollment & Management
- Multiple Encodings per User
- Liveness Detection & Replay Attack Prevention
- Quality Checks & Validation
- Statistics Dashboard
- Tested with 1000+ images, 200+ people
- Production Ready & Scalable
- Comprehensive Documentation

### Version 2.6 - Production Ready
- Excel Export (All Data & Monthly)
- Status Kehadiran (Hadir/Izin/Sakit/Cuti/Alpha)
- Keterangan Detail (Textarea)
- Frontend Improvements (Dropdown, Validation)
- Binary Size Optimization (-30%)
- Production Hardening (Security, Performance)
- Comprehensive Documentation

### Version 2.3
- User Management (CRUD)
- Admin Dashboard
- Activity Logging
- Security Audit

### Version 1.0
- Basic authentication
- Clock in/out
- Attendance history
- IP restriction

---

## Additional Documentation

### Face Recognition
- [Face Recognition API Documentation](../docs_and_backup/FACE_RECOGNITION_API_DOCUMENTATION.md)
- [Face Recognition Installation Guide](../docs_and_backup/FACE_RECOGNITION_MANUAL_INSTALLATION_GUIDE.md)
- [Phase 3 Progressive Testing Results](../docs_and_backup/PHASE3_PROGRESSIVE_TEST_RESULTS.md)
- [Testing Complete Summary](../docs_and_backup/TESTING_COMPLETE_SUMMARY.md)
- [Dataset Research](../docs_and_backup/FACE_RECOGNITION_DATASET_RESEARCH_2026.md)

### General Documentation
- [Health Check Report](../docs_and_backup/HEALTH_CHECK_AND_CLEANUP_REPORT.md)
- [Production Ready Guide](../docs_and_backup/PRODUCTION_READY_FINAL.md)
- [Security Audit Report](../docs_and_backup/security_audit_report.md)
- [Phase 2 Complete Summary](../docs_and_backup/PHASE_2_COMPLETE_SUMMARY.md)
- [Frontend Improvements](../docs_and_backup/FRONTEND_IMPROVEMENTS_COMPLETE.md)
- [Build Guide](../docs_and_backup/BUILD_GUIDE.md)

---

**Built with Go | Production Ready | Security Score: 9/10**
