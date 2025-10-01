#!/bin/bash
set -e

# Vexa Bootstrap Script
# Installs all required dependencies for Samba AD DC replacement

echo "======================================"
echo "  Vexa Bootstrap Installer"
echo "  Active Directory Replacement"
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

# Detect OS
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
        echo -e "${GREEN}Detected OS: $PRETTY_NAME${NC}"
    else
        echo -e "${RED}Cannot detect OS. /etc/os-release not found.${NC}"
        exit 1
    fi
}

# Install dependencies based on OS
install_debian_ubuntu() {
    echo -e "${YELLOW}Installing dependencies for Debian/Ubuntu...${NC}"
    
    # Update package lists
    apt-get update
    
    # Install Samba AD DC and dependencies
    echo "Installing Samba AD DC..."
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        samba \
        smbclient \
        winbind \
        libpam-winbind \
        libnss-winbind \
        krb5-user \
        krb5-config \
        libpam-krb5 \
        bind9 \
        bind9-dnsutils \
        bind9-utils \
        dnsutils \
        ldb-tools \
        attr \
        acl \
        python3-setproctitle \
        python3-dnspython \
        libpam-pwquality \
        libpam-dev \
        build-essential \
        pkg-config
    
    echo -e "${GREEN}Debian/Ubuntu dependencies installed successfully${NC}"
}

install_rhel_centos() {
    echo -e "${YELLOW}Installing dependencies for RHEL/CentOS/Rocky...${NC}"
    
    # Install EPEL repository
    dnf install -y epel-release || yum install -y epel-release
    
    # Install Samba AD DC and dependencies
    echo "Installing Samba AD DC..."
    dnf install -y \
        samba \
        samba-dc \
        samba-winbind \
        samba-winbind-clients \
        krb5-workstation \
        krb5-server \
        krb5-libs \
        pam_krb5 \
        bind \
        bind-utils \
        ldb-tools \
        attr \
        acl \
        python3-dns \
        pam-devel \
        gcc \
        make \
        pkg-config \
        || yum install -y \
        samba \
        samba-dc \
        samba-winbind \
        samba-winbind-clients \
        krb5-workstation \
        krb5-server \
        krb5-libs \
        pam_krb5 \
        bind \
        bind-utils \
        ldb-tools \
        attr \
        acl \
        python3-dns \
        pam-devel \
        gcc \
        make \
        pkg-config
    
    echo -e "${GREEN}RHEL/CentOS dependencies installed successfully${NC}"
}

install_arch() {
    echo -e "${YELLOW}Installing dependencies for Arch Linux...${NC}"
    
    pacman -Sy --noconfirm \
        samba \
        krb5 \
        bind \
        bind-tools \
        pam \
        acl \
        attr \
        python-dnspython \
        base-devel \
        pkg-config
    
    echo -e "${GREEN}Arch Linux dependencies installed successfully${NC}"
}

# Stop conflicting services
stop_conflicting_services() {
    echo -e "${YELLOW}Stopping conflicting services...${NC}"
    
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
    
    echo -e "${GREEN}Conflicting services stopped${NC}"
}

# Create Vexa directories
create_directories() {
    echo -e "${YELLOW}Creating Vexa directory structure...${NC}"
    
    SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
    
    mkdir -p "$SCRIPT_DIR/api"
    mkdir -p "$SCRIPT_DIR/web"
    mkdir -p "$SCRIPT_DIR/data"
    mkdir -p "$SCRIPT_DIR/logs"
    mkdir -p "/etc/vexa"
    mkdir -p "/var/lib/vexa"
    mkdir -p "/var/log/vexa"
    
    # Set permissions
    chmod 755 "$SCRIPT_DIR/api"
    chmod 755 "$SCRIPT_DIR/web"
    chmod 750 "$SCRIPT_DIR/data"
    chmod 755 "$SCRIPT_DIR/logs"
    chmod 750 "/etc/vexa"
    chmod 750 "/var/lib/vexa"
    chmod 755 "/var/log/vexa"
    
    echo -e "${GREEN}Directory structure created${NC}"
}

# Verify installations
verify_installations() {
    echo -e "${YELLOW}Verifying installations...${NC}"
    
    commands=(
        "samba"
        "samba-tool"
        "smbclient"
        "kinit"
        "klist"
        "named"
        "dig"
        "go"
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
        return 0
    else
        echo -e "${YELLOW}Some commands are missing but this may be OK${NC}"
        return 0
    fi
}

# Check Node.js/npm for frontend
check_nodejs() {
    echo -e "${YELLOW}Checking Node.js installation...${NC}"
    
    if ! command -v node &> /dev/null; then
        echo -e "${YELLOW}Node.js not found. Installing...${NC}"
        
        case "$OS" in
            ubuntu|debian)
                curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
                apt-get install -y nodejs
                ;;
            centos|rhel|rocky)
                curl -fsSL https://rpm.nodesource.com/setup_20.x | bash -
                dnf install -y nodejs || yum install -y nodejs
                ;;
            arch)
                pacman -S --noconfirm nodejs npm
                ;;
        esac
    fi
    
    if command -v node &> /dev/null; then
        NODE_VERSION=$(node --version)
        NPM_VERSION=$(npm --version)
        echo -e "${GREEN}Node.js: $NODE_VERSION${NC}"
        echo -e "${GREEN}npm: $NPM_VERSION${NC}"
    else
        echo -e "${RED}Failed to install Node.js${NC}"
        exit 1
    fi
}

# Main installation flow
main() {
    detect_os
    
    case "$OS" in
        ubuntu|debian)
            install_debian_ubuntu
            ;;
        centos|rhel|rocky)
            install_rhel_centos
            ;;
        arch)
            install_arch
            ;;
        *)
            echo -e "${RED}Unsupported OS: $OS${NC}"
            echo "Supported: Ubuntu, Debian, CentOS, RHEL, Rocky Linux, Arch Linux"
            exit 1
            ;;
    esac
    
    stop_conflicting_services
    create_directories
    check_nodejs
    verify_installations
    
    echo ""
    echo -e "${GREEN}======================================"
    echo "  Vexa Bootstrap Complete!"
    echo "======================================${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Configure Samba AD DC: cd api && go run ."
    echo "2. Build frontend: cd web && npm install && npm run dev"
    echo "3. Access the web interface on http://localhost:5173"
    echo ""
    echo -e "${YELLOW}Note: You may need to configure your firewall to allow:"
    echo "  - DNS (53/tcp, 53/udp)"
    echo "  - Kerberos (88/tcp, 88/udp)"
    echo "  - LDAP (389/tcp, 389/udp)"
    echo "  - SMB (445/tcp)"
    echo "  - Kerberos Password (464/tcp, 464/udp)"
    echo "  - LDAPS (636/tcp)"
    echo "  - Global Catalog (3268/tcp, 3269/tcp)"
    echo -e "${NC}"
}

# Run main function
main

