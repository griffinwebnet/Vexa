# Vexa - Modern Directory Services Platform

A modern, open-source directory services platform built on Samba AD DC with secure mesh networking and a beautiful web-based management interface.

![Vexa Dashboard](vexa.png)
*Vexa's modern, responsive web interface with dark mode support*

## Features

- **Samba-based AD DC**: Full AD-compatible Domain Controller functionality
- **Secure Mesh Networking**: Built-in Headscale/Tailscale integration for secure remote access
- **Modern Web Interface**: Beautiful, responsive React-based admin interface
- **PAM Authentication**: Authenticate with Linux PAM or directory credentials
- **User & Group Management**: Easy-to-use interface for managing AD-compatible users and groups
- **Computer Management**: Deploy and manage domain-joined computers with offline scripts
- **DNS Management**: Integrated DNS management with split DNS for mesh networking
- **Mobile Responsive**: Works on desktop, tablet, and mobile devices
- **Light & Dark Mode**: Comfortable interface for any environment
- **LXC Compatible**: Can run in full Linux or LXC containers

## Architecture

- **Backend API**: Go-based REST API with PAM authentication
- **Frontend**: React + Vite + Tailwind CSS + TypeScript
- **Domain Controller**: Samba AD DC
- **Mesh Networking**: Headscale (self-hosted Tailscale control plane)
- **DNS**: Samba Internal DNS with split DNS for mesh domains
- **Authentication**: Kerberos + LDAP

## Prerequisites

- Ubuntu 24.04 LTS
- Root or sudo access
- Internet connection for initial setup

## Quick Start

### One-Command Installation

```bash
curl -fsSL https://raw.githubusercontent.com/griffinwebnet/bootstrap.sh | sudo bash
```

This will install all dependencies and start the services automatically.

### Set Up Your Domain

1. **Go to the URL**: `http://localhost:8080` (or your server's IP address)
2. **Login** with your Linux user credentials (must be a sudoer)
3. **Navigate to "Domain Setup"** in the sidebar
4. **Configure your domain**:
   - Domain name (e.g., `company.local`)
   - Domain controller hostname
   - Administrator password
5. **Click "Provision Domain"** and wait for completion

### Optional: Enable Overlay Networking

1. **Go to "Overlay Networking"** in the sidebar
2. **Enter your FQDN** (e.g., `vexa.company.com`)
3. **Click "Set Up Overlay Networking"** to enable secure mesh VPN

### Deploy Computers

1. **Go to "Computers"** in the sidebar
2. **Click "Add Computer"**
3. **Choose deployment option**:
   - **Domain Join with Tailscale**: Full domain join + secure mesh access
   - **Domain Join Only**: Local domain join only
   - **Add to Tailnet**: Add existing domain computer to mesh
4. **Download the script** and run it on the target computer

## Project Structure

```
Vexa/
├── bootstrap.sh              # One-click setup script
├── api/                      # Go backend API
│   ├── main.go
│   ├── handlers/             # API route handlers
│   ├── services/             # Business logic services
│   ├── middleware/           # Auth & CORS middleware
│   └── scripts/              # Deployment scripts
├── web/                      # React frontend
│   ├── src/
│   │   ├── components/       # Reusable UI components
│   │   ├── pages/            # Page components
│   │   ├── layouts/          # Layout components
│   │   ├── stores/           # Zustand state management
│   │   └── lib/              # API client & utilities
│   └── package.json
└── README.md
```

## Key Features

### Domain Management
- **One-click domain provisioning** with Samba AD DC
- **Automatic DNS configuration** with split DNS for mesh networking
- **Kerberos and LDAP integration** for full AD compatibility

### Secure Mesh Networking
- **Headscale integration** for self-hosted Tailscale control plane
- **Split DNS configuration** for seamless domain resolution
- **Offline deployment scripts** for remote computer setup
- **Automatic key management** with reusable infrastructure keys

### Computer Deployment
- **Offline PowerShell scripts** for Windows deployment
- **Automatic Tailscale installation** and configuration
- **Domain join automation** with unattended setup
- **Remote access via mesh network** after deployment

### User & Group Management
- **Web-based user management** with AD compatibility
- **Group membership management** with nested groups support
- **Password policy enforcement** and account management

## Firewall Configuration

Make sure these ports are open:

- **Web Interface**: 8080/tcp
- **Headscale API**: 50443/tcp (for mesh networking)

## Development

### Backend
```bash
cd api
go run .
```

### Frontend
```bash
cd web
npm run dev
```

### Build for Production

**Backend:**
```bash
cd api
go build -o vexa-api
```

**Frontend:**
```bash
cd web
npm run build
# Output will be in web/dist/
```

## Deployment Scenarios

### Local Network Only
- Deploy computers using "Domain Join Only" option
- All computers must be on the same local network
- No external access to domain resources

### Remote Access with Mesh Networking
- Enable "Overlay Networking" during setup
- Deploy computers using "Domain Join with Tailscale"
- Access domain resources from anywhere via secure mesh VPN
- Automatic DNS resolution for both local and mesh domains

### Hybrid Deployment
- Mix of local and remote computers
- Some computers domain-joined only, others with mesh access
- Flexible deployment based on security requirements

## Troubleshooting

### Domain Provisioning Issues
- Check system logs: `journalctl -u samba-ad-dc`
- Verify DNS resolution: `nslookup yourdomain.local`
- Test Kerberos: `kinit administrator@YOURDOMAIN.LOCAL`

### Mesh Networking Issues
- Check Headscale status: `systemctl status headscale`
- Verify Tailscale connection: `tailscale status`
- Test mesh DNS: `nslookup computer.domain.mesh`

### Computer Deployment Issues
- Ensure target computer has internet access for Tailscale download
- Run PowerShell as Administrator
- Check Windows Event Logs for domain join errors

## Contributing

This is an open-source project. Contributions are welcome!

## License

[Add your license here]

## Support

For issues and questions, please open an issue on GitHub.