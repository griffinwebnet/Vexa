# Vexa API - Development Guide

## 🚀 Quick Start

The Vexa API supports **both development and production modes**:

- **Development Mode**: Uses dummy data for UI testing (no Samba required)
- **Production Mode**: Uses real Samba commands (requires Samba environment)

### Development Mode (UI Testing)

**Perfect for frontend development - no Samba setup required!**

1. **Start the API server in development mode:**
   ```bash
   cd api
   ENV=development go run main.go
   ```

2. **Access the web interface:**
   - Frontend: http://localhost:3000
   - API: http://localhost:8080

3. **Test with dummy data:**
   - Users: 6 sample users with different roles
   - Groups: 7 sample groups with memberships
   - Computers: 5 sample computers with different connection types
   - Domain: Pre-configured example.local domain

4. **Login credentials:**
   - **Admin:** `admin` / `admin` (full access to all features)
   - **User:** `user` / `user` (dashboard + self-service only)

### Production Mode (Real Samba)

**For testing actual Samba integration:**

1. **Prerequisites:**
   - **Linux/WSL2**: Install Samba tools
     ```bash
     sudo apt install samba samba-tools  # Ubuntu/Debian
     sudo yum install samba samba-client # CentOS/RHEL
     sudo pacman -S samba                # Arch
     ```

2. **Setup:**
   ```bash
   # Run the development setup script
   ./api/scripts/dev-setup.sh
   
   # Start in production mode
   cd api
   ENV=production go run main.go
   ```

3. **Provision a domain:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/domain/provision \
     -H "Content-Type: application/json" \
     -d '{
       "domain": "DEV",
       "realm": "dev.local", 
       "admin_password": "DevPass123!"
     }'
   ```

## 🔧 Development Features

### Dual Mode Support
- ✅ **Development Mode** - Rich dummy data for UI testing
- ✅ **Production Mode** - Real Samba commands and data
- ✅ **Easy switching** - Just change the ENV variable
- ✅ **Same API interface** - No code changes needed

### Architecture
- ✅ **Clean separation of concerns** - Handlers, Services, Exec layers
- ✅ **Testable code** - Services can be unit tested
- ✅ **Maintainable** - Each layer has one responsibility

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ENV` | `development` | Environment mode |
| `PORT` | `8080` | API server port |

## 🐛 Troubleshooting

### "samba-tool not found"
- **Linux**: Install Samba packages
- **Windows**: Use WSL2 or Docker

### "Domain not provisioned"
- Run the domain provision API call (see setup step 3)
- Check Samba service: `sudo systemctl status samba-ad-dc`

### "Permission denied"
- Make sure you're not running as root
- Check file permissions on Samba directories

## 📝 API Testing

Test the API with real commands:

```bash
# List users (will be empty initially)
curl http://localhost:8080/api/v1/users

# Create a user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "username": "testuser",
    "password": "TestPass123!",
    "full_name": "Test User"
  }'

# List groups
curl http://localhost:8080/api/v1/groups
```

## 🎯 Benefits

### Development Mode
- **🚀 Fast UI development** - No Samba setup required
- **📊 Rich test data** - Multiple users, groups, computers to test with
- **🎨 UI testing** - Perfect for frontend development and design
- **⚡ Instant feedback** - No waiting for real commands

### Production Mode  
- **🔍 Real development** - Work with actual Samba commands
- **🐛 Better testing** - Catch real errors and edge cases  
- **🚀 Production-like** - Same behavior as production environment
- **✅ No surprises** - What you develop is what you deploy

## 🔄 Switching Modes

```bash
# Development mode (dummy data)
ENV=development go run main.go

# Production mode (real Samba)
ENV=production go run main.go

# Default is development mode
go run main.go  # Same as ENV=development
```

---

**Perfect for your workflow**: Use development mode for UI work, then switch to production mode when you're ready to test real Samba integration on your Linux box!
