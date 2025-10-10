#!/bin/bash
set -e

# Vexa Update Script
echo "======================================"
echo "  Vexa Update Script  v0.3.124"
echo "======================================"
echo ""
echo "Usage: $0 [--nightly] [--fast]"
echo "  --nightly  Update to main branch (development)"
echo "  --fast     Skip npm cache clean and dependency reinstall"
echo "  --help     Show this help message"
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Parse command line arguments
NIGHTLY=false
FAST=false

for arg in "$@"; do
    case $arg in
        --nightly)
            NIGHTLY=true
            echo -e "${YELLOW}NIGHTLY MODE: Updating to main branch${NC}"
            ;;
        --fast)
            FAST=true
            echo -e "${YELLOW}FAST MODE: Skipping npm cache clean and dependency reinstall${NC}"
            ;;
        --help)
            echo "Vexa Update Script"
            echo ""
            echo "Usage: $0 [--nightly] [--fast]"
            echo ""
            echo "Options:"
            echo "  --nightly  Update to main branch (development version)"
            echo "  --fast     Skip npm cache clean and dependency reinstall for faster builds"
            echo "  --help     Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                    # Update to latest stable release"
            echo "  $0 --nightly          # Update to main branch"
            echo "  $0 --fast             # Fast update (skip npm clean)"
            echo "  $0 --nightly --fast   # Fast nightly update"
            echo ""
            exit 0
            ;;
    esac
done

if [ "$NIGHTLY" = true ] || [ "$FAST" = true ]; then
    echo ""
fi

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}ERROR: This script must be run as root${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

# Check for unprivileged container issues
if [ -f /proc/1/status ] && grep -q "CapEff.*0000000000000000" /proc/1/status 2>/dev/null; then
    CURRENT_SAMBA=$(dpkg -l | grep "^ii  samba " | awk '{print $3}')
    if [[ "$CURRENT_SAMBA" == *"4.19"* ]]; then
        echo ""
        echo -e "${YELLOW}⚠️  WARNING: Samba 4.19 in unprivileged container${NC}"
        echo -e "Domain provisioning may fail. Consider migrating to:"
        echo -e "  - Debian 12 (Bookworm) LXC, or"
        echo -e "  - Full KVM/QEMU VM"
        echo ""
    fi
fi

# Ensure required packages are installed
echo -e "${YELLOW}Ensuring required packages...${NC}"
apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -y \
    samba-common-bin \
    smbclient \
    pamtester \
    libpam-modules
echo -e "${GREEN}Required packages installed${NC}"

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
    
    # Try to get the latest release with better error handling
    RELEASE_RESPONSE=$(curl -s -w "%{http_code}" https://api.github.com/repos/griffinwebnet/Vexa/releases/latest)
    HTTP_CODE="${RELEASE_RESPONSE: -3}"
    RELEASE_BODY="${RELEASE_RESPONSE%???}"
    
    echo "DEBUG: HTTP response code: $HTTP_CODE"
    
    if [ "$HTTP_CODE" = "200" ]; then
        LATEST_RELEASE=$(echo "$RELEASE_BODY" | jq -r '.tag_name' 2>/dev/null)
        echo "DEBUG: Latest release tag: $LATEST_RELEASE"
        
        if [ -n "$LATEST_RELEASE" ] && [ "$LATEST_RELEASE" != "null" ] && [ "$LATEST_RELEASE" != "" ]; then
            echo "Latest release found: $LATEST_RELEASE"
            git clone --branch "$LATEST_RELEASE" --depth 1 https://github.com/griffinwebnet/Vexa.git
            cd Vexa
            CURRENT_VERSION="$LATEST_RELEASE"
        else
            echo -e "${YELLOW}Could not parse release tag, using main branch...${NC}"
            git clone https://github.com/griffinwebnet/Vexa.git
            cd Vexa
            git checkout master 2>/dev/null || git checkout main 2>/dev/null || true
            CURRENT_VERSION="main-$(git rev-parse --short HEAD)"
            echo "Using development version: $CURRENT_VERSION"
        fi
    else
        echo -e "${YELLOW}Could not fetch releases (HTTP $HTTP_CODE), using main branch...${NC}"
        git clone https://github.com/griffinwebnet/Vexa.git
        cd Vexa
        git checkout master 2>/dev/null || git checkout main 2>/dev/null || true
        CURRENT_VERSION="main-$(git rev-parse --short HEAD)"
        echo "Using development version: $CURRENT_VERSION"
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

if [ "$FAST" = false ]; then
    # Full clean build - clean npm cache and dependencies to fix rollup module issues
    echo -e "${YELLOW}Cleaning npm dependencies...${NC}"
    rm -rf node_modules package-lock.json
    npm cache clean --force
    
    # Install dependencies
    echo -e "${YELLOW}Installing dependencies...${NC}"
    npm install
else
    # Fast build - check if dependencies need updating
    if [ ! -d "node_modules" ]; then
        echo -e "${YELLOW}Installing dependencies (node_modules missing)...${NC}"
        npm install
    elif [ -f "package-lock.json" ] && [ -f "/var/www/vexa/web/package-lock.json" ]; then
        # Check if package-lock.json has changed
        if ! cmp -s "package-lock.json" "/var/www/vexa/web/package-lock.json"; then
            echo -e "${YELLOW}Dependencies changed, reinstalling...${NC}"
            rm -rf node_modules
            npm install
        else
            echo -e "${YELLOW}Skipping dependency install (no changes detected)${NC}"
        fi
    else
        echo -e "${YELLOW}Skipping dependency install (using existing node_modules)${NC}"
    fi
fi

# Build the frontend
echo -e "${YELLOW}Building frontend...${NC}"
npm run build

# Update web files
echo -e "${YELLOW}Updating web files...${NC}"
rm -rf /var/www/vexa/web/dist
mkdir -p /var/www/vexa/web
cp -r dist /var/www/vexa/web/

# Preserve package files for future dependency comparison
if [ -f "package.json" ]; then
    cp package.json /var/www/vexa/web/
fi
if [ -f "package-lock.json" ]; then
    cp package-lock.json /var/www/vexa/web/
fi

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
echo "Build mode: $([ "$FAST" = true ] && echo "FAST (skipped npm clean)" || echo "FULL (clean build)")"
echo "Running at: http://$(hostname -I | awk '{print $1}')"
echo ""
echo "Services restarted:"
echo "  - vexa-api: $(systemctl is-active vexa-api)"
echo "  - nginx: $(systemctl is-active nginx)"
echo ""

