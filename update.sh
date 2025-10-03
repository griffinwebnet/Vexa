#!/bin/bash
set -e

# Vexa Update Script
echo "======================================"
echo "  Vexa Update Script  v0.1.20"
echo "======================================"
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Parse command line arguments
NIGHTLY=false
if [ "$1" == "--nightly" ]; then
    NIGHTLY=true
    echo -e "${YELLOW}NIGHTLY MODE: Updating to main branch${NC}"
    echo ""
fi

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}ERROR: This script must be run as root${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

# Ensure Samba is held at 4.15 (don't let it upgrade to buggy 4.19.5)
echo -e "${YELLOW}Checking Samba version...${NC}"
CURRENT_SAMBA=$(dpkg -l | grep "^ii  samba " | awk '{print $3}')

if [[ "$CURRENT_SAMBA" == *"4.15"* ]]; then
    echo -e "${GREEN}Samba 4.15 installed - good!${NC}"
    # Ensure packages are held
    apt-mark hold samba samba-dsdb-modules samba-common samba-common-bin samba-libs winbind libwbclient0 2>/dev/null || true
else
    echo -e "${RED}WARNING: Samba $CURRENT_SAMBA detected - need 4.15 for LXC!${NC}"
    echo -e "${YELLOW}Downgrading to Samba 4.15...${NC}"
    
    cd /tmp
    rm -rf samba_debs 2>/dev/null || true
    mkdir -p samba_debs
    cd samba_debs
    
    SAMBA_VERSION="4.15.13+dfsg-0ubuntu1.6"
    
    # Download Samba 4.15 packages
    echo -e "${YELLOW}Downloading Samba 4.15 packages...${NC}"
    wget -q http://archive.ubuntu.com/ubuntu/pool/main/s/samba/samba_${SAMBA_VERSION}_amd64.deb
    wget -q http://archive.ubuntu.com/ubuntu/pool/main/s/samba/samba-common_${SAMBA_VERSION}_all.deb
    wget -q http://archive.ubuntu.com/ubuntu/pool/main/s/samba/samba-common-bin_${SAMBA_VERSION}_amd64.deb
    wget -q http://archive.ubuntu.com/ubuntu/pool/main/s/samba/samba-dsdb-modules_${SAMBA_VERSION}_amd64.deb
    wget -q http://archive.ubuntu.com/ubuntu/pool/main/s/samba/samba-libs_${SAMBA_VERSION}_amd64.deb
    wget -q http://archive.ubuntu.com/ubuntu/pool/main/s/samba/winbind_${SAMBA_VERSION}_amd64.deb
    wget -q http://archive.ubuntu.com/ubuntu/pool/main/s/samba/libwbclient0_${SAMBA_VERSION}_amd64.deb
    
    # Stop Samba services before downgrade
    echo -e "${YELLOW}Stopping Samba services...${NC}"
    systemctl stop samba-ad-dc 2>/dev/null || true
    systemctl stop smbd nmbd winbind 2>/dev/null || true
    
    # Force install downgraded packages
    echo -e "${YELLOW}Installing Samba 4.15...${NC}"
    DEBIAN_FRONTEND=noninteractive dpkg --force-downgrade --force-depends -i *.deb
    DEBIAN_FRONTEND=noninteractive apt-get install -f -y
    
    # Hold packages
    echo -e "${YELLOW}Holding Samba packages...${NC}"
    apt-mark hold samba samba-dsdb-modules samba-common samba-common-bin samba-libs winbind libwbclient0
    
    cd /tmp
    rm -rf samba_debs
    
    INSTALLED_VERSION=$(dpkg -l | grep "^ii  samba " | awk '{print $3}')
    echo -e "${GREEN}Samba downgraded to: $INSTALLED_VERSION${NC}"
    
    if [[ ! "$INSTALLED_VERSION" == *"4.15"* ]]; then
        echo -e "${RED}ERROR: Samba 4.15 installation failed!${NC}"
        exit 1
    fi
fi

# Fetch Vexa source
cd /tmp
rm -rf Vexa 2>/dev/null || true

if [ "$NIGHTLY" = true ]; then
    echo -e "${YELLOW}Cloning nightly (main branch)...${NC}"
    git clone https://github.com/griffinwebnet/Vexa.git
    cd Vexa
    # Try master first, fallback to main
    git checkout master 2>/dev/null || git checkout main 2>/dev/null || true
    CURRENT_VERSION="nightly-$(git rev-parse --short HEAD)"
    echo "Updating to nightly build: $CURRENT_VERSION"
else
    echo -e "${YELLOW}Fetching latest release...${NC}"
    LATEST_RELEASE=$(curl -s https://api.github.com/repos/griffinwebnet/Vexa/releases/latest | jq -r '.tag_name')
    
    if [ -z "$LATEST_RELEASE" ] || [ "$LATEST_RELEASE" = "null" ]; then
        echo -e "${YELLOW}No releases found, falling back to main branch...${NC}"
        git clone https://github.com/griffinwebnet/Vexa.git
        cd Vexa
        git checkout master 2>/dev/null || git checkout main 2>/dev/null || true
        CURRENT_VERSION="main-$(git rev-parse --short HEAD)"
    else
        echo "Latest release: $LATEST_RELEASE"
        git clone --branch "$LATEST_RELEASE" --depth 1 https://github.com/griffinwebnet/Vexa.git
        cd Vexa
        CURRENT_VERSION="$LATEST_RELEASE"
    fi
fi

# Stop the API service
echo -e "${YELLOW}Stopping vexa-api service...${NC}"
systemctl stop vexa-api

# Build Go API
echo -e "${YELLOW}Building API server...${NC}"
cd /tmp/Vexa/api
go build -o /usr/local/bin/vexa-api
chmod +x /usr/local/bin/vexa-api
echo -e "${GREEN}API built${NC}"

# Build React frontend
echo -e "${YELLOW}Building web interface...${NC}"
cd /tmp/Vexa/web
npm ci
npm run build

# Update web files
echo -e "${YELLOW}Updating web files...${NC}"
rm -rf /var/www/vexa/web/dist
mkdir -p /var/www/vexa/web
cp -r dist /var/www/vexa/web/
echo -e "${GREEN}Frontend updated${NC}"

# Copy updated API source (for WorkingDirectory in systemd)
echo -e "${YELLOW}Updating API files...${NC}"
rm -rf /var/www/vexa/api
cp -r /tmp/Vexa/api /var/www/vexa/
echo -e "${GREEN}API files updated${NC}"

# Start the API service
echo -e "${YELLOW}Starting vexa-api service...${NC}"
systemctl start vexa-api

# Reload nginx
systemctl reload nginx

# Cleanup
rm -rf /tmp/Vexa

echo ""
echo -e "${GREEN}======================================"
echo "  Update Complete!"
echo "======================================${NC}"
echo ""
echo "Vexa version: $CURRENT_VERSION"
echo "Running at: http://$(hostname -I | awk '{print $1}')"
echo ""
echo "Services restarted:"
echo "  - vexa-api: $(systemctl is-active vexa-api)"
echo "  - nginx: $(systemctl is-active nginx)"
echo ""

