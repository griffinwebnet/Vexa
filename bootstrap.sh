#!/bin/bash
set -e

# Vexa Bootstrap Script
echo "======================================"
echo "  Vexa Bootstrap Installer  v0.1.14"
echo "======================================"
echo ""

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Parse command line arguments
NIGHTLY=false
if [ "$1" == "--nightly" ]; then
    NIGHTLY=true
    echo -e "${YELLOW}NIGHTLY MODE: Installing from main branch${NC}"
    echo ""
fi

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}ERROR: This script must be run as root${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

echo -e "${GREEN}Installing system packages...${NC}"
apt-get update
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
    wget \
    nginx \
    ddclient \
    jq \
    pamtester \
    vim-nox \
    golang-go \
    nodejs \
    npm

echo -e "${GREEN}Packages installed${NC}"

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
    echo "Installing nightly build: $CURRENT_VERSION"
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

# Create installation directory
mkdir -p /var/www/vexa

# Copy source files
echo "Copying source files..."
cp -r api /var/www/vexa/
cp -r web /var/www/vexa/

# Build Go API
echo -e "${YELLOW}Building API server...${NC}"
cd /var/www/vexa/api
go build -o /usr/local/bin/vexa-api
chmod +x /usr/local/bin/vexa-api
echo -e "${GREEN}API built${NC}"

# Build React frontend
echo -e "${YELLOW}Building web interface...${NC}"
cd /var/www/vexa/web
npm ci
npm run build
echo -e "${GREEN}Frontend built${NC}"

# Configure Nginx
echo -e "${YELLOW}Configuring Nginx...${NC}"

cat > /etc/nginx/sites-available/vexa << 'EOF'
server {
    listen 80;
    server_name _;
    
    root /var/www/vexa/web/dist;
    index index.html;
    
    location /api/ {
        proxy_pass http://localhost:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    location /health {
        proxy_pass http://localhost:8080/health;
    }
    
    location / {
        try_files $uri $uri/ /index.html;
    }
}
EOF

ln -sf /etc/nginx/sites-available/vexa /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default
nginx -t && systemctl reload nginx
echo -e "${GREEN}Nginx configured${NC}"

# Create systemd service
echo -e "${YELLOW}Creating systemd service...${NC}"

cat > /etc/systemd/system/vexa-api.service << 'EOF'
[Unit]
Description=Vexa API Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/var/www/vexa/api
ExecStart=/usr/local/bin/vexa-api
Restart=always
RestartSec=5
Environment=ENV=production

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable vexa-api
systemctl start vexa-api
systemctl enable nginx

echo -e "${GREEN}Service started${NC}"

# Done
echo ""
echo -e "${GREEN}======================================"
echo "  Installation Complete!"
echo "======================================${NC}"
echo ""
echo "Vexa version: $CURRENT_VERSION"
echo "Running at: http://$(hostname -I | awk '{print $1}')"
echo ""
echo "Services:"
echo "  - vexa-api: systemctl status vexa-api"
echo "  - nginx: systemctl status nginx"
echo ""
echo "To start the API with dev mode (test credentials):"
echo "  systemctl stop vexa-api"
echo "  /usr/local/bin/vexa-api --dev"
echo ""
echo "To update Vexa in the future:"
echo "  wget https://raw.githubusercontent.com/griffinwebnet/Vexa/master/update.sh"
echo "  chmod +x update.sh"
echo "  sudo ./update.sh"
echo ""
