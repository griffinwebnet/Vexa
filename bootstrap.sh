#!/bin/bash
set -e

# Vexa Bootstrap Script
echo "======================================"
echo "  Vexa Bootstrap Installer"
echo "======================================"
echo ""

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}ERROR: This script must be run as root${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

# Detect OS family
detect_os() {
    if [ ! -f /etc/os-release ]; then
        echo -e "${RED}Cannot detect OS. /etc/os-release not found.${NC}"
        exit 1
    fi
    
    . /etc/os-release
    
    case "$ID" in
        # Debian family
        ubuntu|debian|linuxmint|pop)
            OS_FAMILY="debian"
            ;;
        
        # RHEL family
        rhel|centos|rocky|almalinux|fedora)
            OS_FAMILY="rhel"
            ;;
        
        # SUSE family
        opensuse*|sles|suse)
            OS_FAMILY="suse"
            ;;
        
        *)
            echo -e "${RED}Unsupported OS: $PRETTY_NAME${NC}"
            echo "Supported distributions:"
            echo "- Debian/Ubuntu and derivatives"
            echo "- RHEL/Rocky/CentOS and derivatives"
            echo "- SUSE Linux Enterprise/OpenSUSE"
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}Detected OS: $PRETTY_NAME${NC}"
}

# Install dependencies based on OS family
install_deps() {
    echo -e "${YELLOW}Installing dependencies...${NC}"
    
    case "$OS_FAMILY" in
        debian)
            # Update package lists
            apt-get update
            
            # Install dependencies
            DEBIAN_FRONTEND=noninteractive apt-get install -y \
                samba \
                samba-dsdb-modules \
                winbind \
                krb5-user \
                krb5-config \
                ldb-tools \
                attr \
                acl \
                build-essential \
                pkg-config \
                git \
                curl \
                nginx \
                ddclient \
                jq \
                pamtester
            ;;
            
        rhel)
            # Enable EPEL
            if command -v dnf &> /dev/null; then
                dnf install -y epel-release
                
                # Install dependencies
                dnf install -y \
                    samba \
                    samba-dsdb-modules \
                    samba-winbind \
                    krb5-workstation \
                    krb5-server \
                    ldb-tools \
                    attr \
                    acl \
                    gcc \
                    make \
                    pkg-config \
                    git \
                    curl \
                    nginx \
                    ddclient \
                    jq
            else
                # Fallback for older RHEL systems using yum
                yum install -y epel-release
                
                yum install -y \
                    samba \
                    samba-dsdb-modules \
                    samba-winbind \
                    krb5-workstation \
                    krb5-server \
                    ldb-tools \
                    attr \
                    acl \
                    gcc \
                    make \
                    pkg-config \
                    git \
                    curl \
                    nginx
            fi
            ;;
            
        suse)
            # Enable development tools repo
            zypper addrepo -f https://download.opensuse.org/repositories/devel:/tools/15.4/devel:tools.repo
            
            # Install dependencies
            zypper --non-interactive install \
                samba \
                samba-dsdb-modules \
                samba-winbind \
                krb5 \
                krb5-server \
                ldb-tools \
                attr \
                acl \
                gcc \
                make \
                pkg-config \
                git \
                curl \
                nginx
            ;;
    esac
}

# Install Go
install_golang() {
    echo -e "${YELLOW}Installing Go...${NC}"
    
    # Download and install Go 1.21
    curl -fsSL https://golang.org/dl/go1.21.3.linux-amd64.tar.gz -o go.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go.tar.gz
    rm go.tar.gz
    
    # Add Go to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
    source /etc/profile.d/go.sh
    
    echo -e "${GREEN}Go installed: $(go version)${NC}"
}

# Install Node.js
install_nodejs() {
    echo -e "${YELLOW}Installing Node.js...${NC}"
    
    case "$OS" in
        ubuntu|debian)
            curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
            apt-get install -y nodejs
            ;;
        centos|rhel|rocky)
            curl -fsSL https://rpm.nodesource.com/setup_20.x | bash -
            dnf install -y nodejs
            ;;
        arch)
            pacman -S --noconfirm nodejs npm
            ;;
    esac
    
    echo -e "${GREEN}Node.js: $(node --version)${NC}"
    echo -e "${GREEN}npm: $(npm --version)${NC}"
}

# Install Dart SDK
install_dart() {
    echo -e "${YELLOW}Installing Dart SDK...${NC}"
    
    case "$OS" in
        ubuntu|debian)
            curl -fsSL https://dl-ssl.google.com/linux/linux_signing_key.pub | gpg --dearmor -o /usr/share/keyrings/dart.gpg
            echo 'deb [signed-by=/usr/share/keyrings/dart.gpg arch=amd64] https://storage.googleapis.com/download.dartlang.org/linux/debian stable main' > /etc/apt/sources.list.d/dart_stable.list
            apt-get update
            apt-get install -y dart
            ;;
        centos|rhel|rocky)
            # For RHEL/CentOS, we'll use the tarball
            curl -fsSL https://storage.googleapis.com/dart-archive/channels/stable/release/latest/sdk/dartsdk-linux-x64-release.zip -o dart.zip
            unzip dart.zip -d /usr/local
            rm dart.zip
            echo 'export PATH=$PATH:/usr/local/dart-sdk/bin' > /etc/profile.d/dart.sh
            source /etc/profile.d/dart.sh
            ;;
        arch)
            pacman -S --noconfirm dart
            ;;
    esac
    
    echo -e "${GREEN}Dart: $(dart --version)${NC}"
}

# Prepare system for Samba AD DC
prepare_system() {
    echo -e "${YELLOW}Preparing system for Samba AD DC...${NC}"
    
    # Stop and disable conflicting services
    services=(
        "smbd"
        "nmbd"
        "winbind"
        "systemd-resolved"
    )
    
    for service in "${services[@]}"; do
        if systemctl is-active --quiet "$service"; then
            echo "Stopping $service..."
            systemctl stop "$service" || true
            systemctl disable "$service" || true
        fi
    done
    
    # Clear existing Samba config
    echo "Clearing existing Samba configuration..."
    rm -f /etc/samba/smb.conf
    
    # Create base smb.conf to bind to all interfaces
    cat > /etc/samba/smb.conf << 'EOF'
[global]
    # Listen on all interfaces
    interfaces = 0.0.0.0/0
    bind interfaces only = no
EOF
    
    # Prepare Samba directories
    echo "Setting up Samba directories..."
    mkdir -p /var/lib/samba/sysvol
    chmod 750 /var/lib/samba/sysvol
    
    # Handle DNS configuration
    echo "Configuring DNS..."
    if [ -L /etc/resolv.conf ]; then
        rm -f /etc/resolv.conf
        touch /etc/resolv.conf
    fi
    
    # Create required Samba user/group
    echo "Setting up Samba system accounts..."
    if ! getent group samba >/dev/null; then
        groupadd -r samba
    fi
    if ! getent passwd samba >/dev/null; then
        useradd -r -g samba -d /var/lib/samba -s /sbin/nologin samba
    fi
    
    echo -e "${GREEN}System prepared for Samba AD DC${NC}"
}

# Setup Vexa
setup_vexa() {
    echo -e "${YELLOW}Setting up Vexa...${NC}"
    
    # Create directories
    mkdir -p /opt/vexa
    mkdir -p /var/log/vexa
    mkdir -p /var/www/vexa
    mkdir -p /etc/vexa
    mkdir -p /var/lib/vexa
    
    # Set permissions
    chmod 750 /etc/vexa
    chmod 750 /var/lib/vexa
    chmod 755 /var/log/vexa
    chmod 755 /var/www/vexa
    
    cd /opt/vexa
    
    # Get latest release version
    echo -e "${YELLOW}Fetching latest release...${NC}"
    local latest_version
    latest_version=$(curl -s https://api.github.com/repos/griffinwebnet/Vexa/releases/latest | grep -Po '"tag_name": "\K[^"]*')
    
    if [ -z "$latest_version" ]; then
        echo -e "${RED}Failed to get latest version${NC}"
        exit 1
    fi
    
    echo "Latest release: $latest_version"
    
    # Clone just that release
    git clone --depth 1 --branch "$latest_version" https://github.com/griffinwebnet/Vexa.git source
    cd source
    
    # Build from release source
    build_components
}

# Build all components
build_components() {
    # Build Dart CLI
    echo "Building Vexa CLI..."
    cd vexa-cli/vexa
    dart pub get
    dart compile exe bin/vexa.dart -o /usr/local/bin/vexa
    chmod +x /usr/local/bin/vexa
    
    # Build Go API
    echo "Building API server..."
    cd ../../api
    go build -o /usr/local/bin/vexa-api
    chmod +x /usr/local/bin/vexa-api
    
    # Build React frontend
    echo "Building web interface..."
    cd ../web
    npm ci
    npm run build
    
    # Install web files
    rm -rf /var/www/vexa/*
    mv dist/* /var/www/vexa/
    
    echo -e "${GREEN}Vexa components built successfully${NC}"
}

# Configure nginx
setup_nginx() {
    echo -e "${YELLOW}Configuring nginx...${NC}"
    
    # Create nginx config
    cat > /etc/nginx/conf.d/vexa.conf << 'EOF'
server {
    listen 80;
    server_name _;
    
    root /var/www/vexa;
    index index.html;
    
    # API proxy
    location /api/ {
        proxy_pass http://localhost:8080/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
    
    # Static files
    location / {
        try_files $uri $uri/ /index.html;
    }
}
EOF
    
    # Test and reload nginx
    nginx -t
    systemctl reload nginx
    
    echo -e "${GREEN}Nginx configured successfully${NC}"
}

# Create systemd services
setup_services() {
    echo -e "${YELLOW}Creating systemd services...${NC}"
    
    # API service
    cat > /etc/systemd/system/vexa-api.service << 'EOF'
[Unit]
Description=Vexa API Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/vexa-api
Restart=always
RestartSec=5
Environment=ENV=production

[Install]
WantedBy=multi-user.target
EOF
    
    # Enable and start services
    systemctl daemon-reload
    systemctl enable nginx vexa-api
    systemctl restart nginx vexa-api
    
    echo -e "${GREEN}Services configured and started${NC}"
}

# Verify installations
verify_installations() {
    echo -e "${YELLOW}Verifying installations...${NC}"
    
    commands=(
        "samba"
        "samba-tool"
        "kinit"
        "go"
        "node"
        "npm"
        "dart"
        "nginx"
        "vexa"
        "vexa-api"
    )
    
    all_ok=true
    for cmd in "${commands[@]}"; do
        if command -v "$cmd" &> /dev/null; then
            echo -e "${GREEN}✓${NC} $cmd"
        else
            echo -e "${RED}✗${NC} $cmd (not found)"
            all_ok=false
        fi
    done
    
    if [ "$all_ok" = true ]; then
        echo -e "${GREEN}All required commands are available${NC}"
    else
        echo -e "${YELLOW}Some commands are missing - check output above${NC}"
    fi
}

# Main installation flow
main() {
    detect_os
    install_deps
    install_golang
    install_nodejs
    install_dart
    prepare_system
    setup_vexa
    setup_nginx
    setup_services
    verify_installations
    
    echo ""
    echo -e "${GREEN}======================================"
    echo "  Vexa Installation Complete!"
    echo "======================================${NC}"
    echo ""
    echo "You can now access Vexa at: http://$(hostname -I | awk '{print $1}')"
    echo ""
    echo "Next steps:"
    echo "1. Open the web interface"
    echo "2. Complete the domain setup wizard"
    echo "3. Start managing your domain"
    echo ""
    echo -e "${YELLOW}Note: Make sure ports are open in your firewall:"
    echo "  - HTTP (80/tcp) - Web interface"
    echo "  - DNS (53/tcp, 53/udp)"
    echo "  - Kerberos (88/tcp, 88/udp)"
    echo "  - LDAP (389/tcp)"
    echo "  - SMB (445/tcp)"
    echo "  - Kerberos Password (464/tcp, 464/udp)"
    echo "  - LDAPS (636/tcp)"
    echo "  - Global Catalog (3268/tcp, 3269/tcp)"
    echo -e "${NC}"
    
    echo ""
    echo -e "${GREEN}======================================"
    echo "  Vexa Installation Complete!"
    echo "======================================${NC}"
    echo ""
    echo "Services installed:"
    echo "1. vexa-api - API server (port 8080)"
    echo "2. nginx - Web interface (port 80)"
    echo "3. samba-ad-dc - Active Directory"
    echo "4. bind9 - DNS Server"
    echo ""
    echo "You can now access Vexa at: http://$(hostname -I | awk '{print $1}')"
    echo ""
    echo -e "${YELLOW}Note: Make sure to configure your firewall to allow:"
    echo "  - HTTP (80/tcp) - Web interface"
    echo "  - API (8080/tcp) - API server"
    echo "  - DNS (53/tcp, 53/udp)"
    echo "  - Kerberos (88/tcp, 88/udp)"
    echo "  - LDAP (389/tcp)"
    echo "  - SMB (445/tcp)"
    echo "  - Kerberos Password (464/tcp, 464/udp)"
    echo "  - LDAPS (636/tcp)"
    echo "  - Global Catalog (3268/tcp, 3269/tcp)"
    echo -e "${NC}"
}

# Run main function
main