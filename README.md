# Sistem Absensi Kantor 🏢

Sistem absensi berbasis web yang secure dan robust, hanya dapat diakses dari jaringan WiFi kantor menggunakan IP restriction. Dirancang untuk memastikan karyawan benar-benar hadir di kantor saat melakukan absensi.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Proprietary-red.svg)](LICENSE)
[![Security](https://img.shields.io/badge/Security-Audited-green.svg)](docs_and_backup/security_audit_report.md)

---

## 📋 Table of Contents

- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Security](#-security)
- [Quick Start](#-quick-start)
- [Project Structure](#-project-structure)
- [API Documentation](#-api-documentation)
- [Deployment](#-deployment)
- [Configuration](#-configuration)
- [Development](#-development)

---

## ✨ Features

### Core Features (Phase 1 & 2)
- ✅ **IP Restriction** - Hanya bisa diakses dari WiFi kantor (CIDR support)
- ✅ **User Authentication** - JWT + bcrypt password hashing
- ✅ **Absen Masuk/Pulang** - Clock in/out dengan timestamp
- ✅ **Riwayat Absensi** - View attendance history
- ✅ **Admin Dashboard** - Monitor semua karyawan
- ✅ **User Management** - CRUD users (admin only)
- ✅ **Activity Logging** - Comprehensive audit trail
- ✅ **Modern UI** - Responsive design dengan Tailwind CSS

### Security Features
- 🔒 Multi-layer security (IP → Auth → Role)
- 🔒 Password hashing dengan bcrypt
- 🔒 JWT authentication dengan expiration
- 🔒 SQL injection prevention
- 🔒 Role-based access control
- 🔒 Comprehensive audit logging
- 🔒 Soft delete (data preservation)

### Coming Soon (Phase 2.4-2.7)
- ⏳ Export to Excel (formatted reports)
- ⏳ Export to Word (formal documents)
- ⏳ Monthly reports dengan statistics
- ⏳ Print-friendly views

---

## 🛠️ Tech Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| **Backend** | Go + Gin | 1.25+ |
| **Database** | SQLite | 3.x |
| **Authentication** | JWT + bcrypt | - |
| **Frontend** | HTML + Tailwind CSS | 3.x |
| **Deployment** | Single binary | - |

### Why This Stack?
- **Go**: Fast, compiled, single binary deployment
- **SQLite**: Zero-configuration, embedded database
- **Gin**: High-performance HTTP framework
- **JWT**: Stateless authentication
- **Tailwind**: Rapid UI development

---

## 🔒 Security

### Security Layers
1. **IP Restriction** - First line of defense
2. **Authentication** - JWT with 24h expiration
3. **Authorization** - Role-based access control

### Security Score: 9/10 ✅
See [Security Audit Report](../docs_and_backup/security_audit_report.md) for details.

### Key Security Features
- ✅ Bcrypt password hashing (cost 10)
- ✅ Parameterized SQL queries (SQL injection prevention)
- ✅ Input validation (XSS prevention)
- ✅ JWT signature verification
- ✅ Role-based access control
- ✅ Comprehensive audit logging
- ✅ Soft delete (data preservation)

---

## 🚀 Quick Start

### Prerequisites
- Go 1.25 or higher
- Git

### 1. Clone Repository

```bash
git clone https://github.com/gerrymoeis/sistem_absensi_kantor.git
cd sistem_absensi_kantor
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configuration

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Edit `.env`:
```env
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_MODE=debug

# Security
JWT_SECRET=change-this-secret-key-in-production
ALLOWED_IPS=127.0.0.1/32,::1/128,192.168.1.0/24

# Database
DB_DRIVER=sqlite
DB_DSN=./data/absensi.db

# Logging
LOG_LEVEL=info
LOG_FILE=./logs/app.log
```

### 4. Create Seed Data

```bash
go run cmd/seed/main.go
```

**Default Users:**
- **Admin**: username: `admin`, password: `admin123`
- **Employee**: username: `user1`, password: `password123`

### 5. Run Server

```bash
go run cmd/server/main.go
```

Server akan berjalan di `http://localhost:8080`

### 6. Access Application

Open browser: `http://localhost:8080`

---

## 📁 Project Structure

```
main_folder/
├── cmd/
│   ├── server/              # Main application entry point
│   │   └── main.go
│   ├── seed/                # Database seeding
│   │   └── main.go
│   └── check_logs/          # Utility to check activity logs
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
│   │   └── admin_handler.go
│   ├── middleware/          # HTTP middlewares
│   │   ├── auth.go          # JWT authentication
│   │   ├── ip_restriction.go # IP whitelist
│   │   └── admin.go         # Admin authorization
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
│       └── user_service.go
├── web/
│   ├── static/              # Static assets (CSS, JS, images)
│   └── templates/           # HTML templates
│       ├── login.html
│       ├── dashboard.html
│       ├── history.html
│       └── admin_dashboard.html
├── data/                    # SQLite database (gitignored)
├── logs/                    # Application logs (gitignored)
├── .env                     # Environment variables (gitignored)
├── .env.example             # Environment variables template
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

---

## 📚 API Documentation

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

#### Logout
```http
POST /api/auth/logout
Authorization: Bearer <token>

Response: 200 OK
{
  "message": "Logged out successfully"
}
```

#### Get Current User
```http
GET /api/auth/me
Authorization: Bearer <token>

Response: 200 OK
{
  "data": { ... }
}
```

### Attendance Endpoints

#### Clock In
```http
POST /api/absensi/masuk
Authorization: Bearer <token>
Content-Type: application/json

{
  "keterangan": "Optional note"
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

Response: 200 OK
{
  "message": "Berhasil absen pulang",
  "data": { ... }
}
```

#### Get Today's Attendance
```http
GET /api/absensi/today
Authorization: Bearer <token>

Response: 200 OK
{
  "data": {
    "id": 1,
    "tanggal": "2026-04-17",
    "jam_masuk": "08:30:00",
    "jam_pulang": "17:00:00",
    "status": "hadir"
  }
}
```

#### Get Attendance History
```http
GET /api/absensi/history?limit=30&offset=0
Authorization: Bearer <token>

Response: 200 OK
{
  "data": [ ... ]
}
```

### Admin Endpoints (Admin Only)

#### Get Statistics
```http
GET /api/admin/stats
Authorization: Bearer <admin-token>

Response: 200 OK
{
  "data": {
    "total_users": 10,
    "hadir_hari_ini": 8,
    "belum_absen": 2,
    "selesai": 5
  }
}
```

#### Get All Users
```http
GET /api/admin/users
Authorization: Bearer <admin-token>

Response: 200 OK
{
  "data": [ ... ]
}
```

#### Create User
```http
POST /api/admin/users
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "username": "newuser",
  "password": "password123",
  "full_name": "New User",
  "role": "employee",
  "is_active": true
}

Response: 201 Created
{
  "message": "User created successfully",
  "data": { ... }
}
```

#### Update User
```http
PUT /api/admin/users/:id
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "full_name": "Updated Name",
  "role": "employee",
  "is_active": true
}

Response: 200 OK
{
  "message": "User updated successfully",
  "data": { ... }
}
```

#### Delete User (Soft Delete)
```http
DELETE /api/admin/users/:id
Authorization: Bearer <admin-token>

Response: 200 OK
{
  "message": "User deleted successfully"
}
```

#### Reset Password
```http
POST /api/admin/users/:id/reset-password
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "new_password": "newpassword123"
}

Response: 200 OK
{
  "message": "Password reset successfully"
}
```

#### Get Activity Logs
```http
GET /api/admin/logs?limit=100&offset=0
Authorization: Bearer <admin-token>

Response: 200 OK
{
  "data": [ ... ]
}
```

### Web Pages
- `GET /` - Redirect to login/dashboard
- `GET /login` - Login page
- `GET /dashboard` - User dashboard
- `GET /history` - Attendance history
- `GET /admin/dashboard` - Admin dashboard (admin only)

---

## 🚀 Deployment

### Build for Production

#### Linux
```bash
GOOS=linux GOARCH=amd64 go build -o absensi-server cmd/server/main.go
```

#### Windows
```bash
GOOS=windows GOARCH=amd64 go build -o absensi-server.exe cmd/server/main.go
```

### Production Configuration

Create `.env` file:
```env
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_MODE=release

# Generate strong secret: openssl rand -base64 32
JWT_SECRET=<your-strong-secret-here>

# Your office network CIDR
ALLOWED_IPS=192.168.1.0/24,10.0.0.0/8

DB_DRIVER=sqlite
DB_DSN=./data/absensi.db

LOG_LEVEL=info
LOG_FILE=./logs/app.log
```

### Systemd Service (Linux)

Create `/etc/systemd/system/absensi.service`:

```ini
[Unit]
Description=Absensi Server
After=network.target

[Service]
Type=simple
User=absensi
WorkingDirectory=/opt/absensi
ExecStart=/opt/absensi/absensi-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable absensi
sudo systemctl start absensi
sudo systemctl status absensi
```

### Nginx Reverse Proxy (HTTPS)

```nginx
server {
    listen 443 ssl http2;
    server_name absensi.yourcompany.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Docker (Optional)

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o absensi-server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/absensi-server .
COPY --from=builder /app/web ./web
EXPOSE 8080
CMD ["./absensi-server"]
```

---

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVER_HOST` | Server bind address | `0.0.0.0` | No |
| `SERVER_PORT` | Server port | `8080` | No |
| `SERVER_MODE` | Gin mode (debug/release) | `debug` | No |
| `JWT_SECRET` | JWT signing secret | - | **Yes** |
| `ALLOWED_IPS` | Comma-separated CIDR list | - | **Yes** |
| `DB_DRIVER` | Database driver | `sqlite` | No |
| `DB_DSN` | Database connection string | `./data/absensi.db` | No |
| `LOG_LEVEL` | Log level | `info` | No |
| `LOG_FILE` | Log file path | `./logs/app.log` | No |

### IP Restriction Configuration

**CIDR Notation Examples:**
```env
# Single IP (IPv4)
ALLOWED_IPS=192.168.1.100/32

# IP Range (IPv4)
ALLOWED_IPS=192.168.1.0/24

# Multiple ranges
ALLOWED_IPS=192.168.1.0/24,10.0.0.0/8,172.16.0.0/12

# IPv6 support
ALLOWED_IPS=::1/128,2001:db8::/32

# Mixed IPv4 and IPv6
ALLOWED_IPS=127.0.0.1/32,::1/128,192.168.1.0/24
```

---

## 💻 Development

### Prerequisites
- Go 1.25+
- Git
- Text editor (VS Code recommended)

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/gerrymoeis/sistem_absensi_kantor.git
cd sistem_absensi_kantor

# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Run database migrations (automatic on first run)
go run cmd/server/main.go

# Create seed data
go run cmd/seed/main.go
```

### Development Commands

```bash
# Run server (with hot reload using air)
air

# Or run directly
go run cmd/server/main.go

# Build
go build -o absensi-server cmd/server/main.go

# Run tests
go test ./...

# Format code
go fmt ./...

# Lint code
golangci-lint run

# Check logs
go run cmd/check_logs/main.go
```

### Code Style

Follow Go best practices:
- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Write clean, minimal code
- Add comments for exported functions
- Use meaningful variable names

---

## 📊 Database Schema

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
    status VARCHAR(20) DEFAULT 'hadir',
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

## 🐛 Troubleshooting

### Common Issues

**1. "Access denied: IP not allowed"**
- Check your IP address: `curl ifconfig.me`
- Update `ALLOWED_IPS` in `.env`
- Restart server after changing `.env`

**2. "Invalid or expired token"**
- Token expires after 24 hours
- Login again to get new token
- Check JWT_SECRET is consistent

**3. "Database locked"**
- SQLite doesn't support concurrent writes well
- Consider PostgreSQL for high concurrency
- Check no other process is using the database

**4. "Port already in use"**
- Change `SERVER_PORT` in `.env`
- Or kill process: `lsof -ti:8080 | xargs kill -9`

---

## 📝 License

Proprietary - Internal Use Only

© 2026 Your Company. All rights reserved.

---

## 👥 Contributors

- **Development Team** - Initial work and maintenance

---

## 📞 Support

For issues and questions:
- Create an issue on GitHub
- Contact: admin@yourcompany.com

---

## 🔄 Changelog

### Version 2.3 (Current)
- ✅ User Management (CRUD)
- ✅ Admin Dashboard
- ✅ Activity Logging
- ✅ Security Audit

### Version 1.0
- ✅ Basic authentication
- ✅ Clock in/out
- ✅ Attendance history
- ✅ IP restriction

---

**Built with ❤️ using Go**
