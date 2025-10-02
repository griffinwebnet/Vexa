#!/bin/bash

# Development Setup Script for Vexa API
# This script sets up a minimal Samba AD DC environment for development

set -e

echo "🚀 Setting up Vexa API development environment..."

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    echo "❌ Please don't run this script as root"
    exit 1
fi

# Check if samba-tool is available
if ! command -v samba-tool &> /dev/null; then
    echo "❌ samba-tool not found. Please install Samba first:"
    echo "   Ubuntu/Debian: sudo apt install samba samba-tools"
    echo "   CentOS/RHEL: sudo yum install samba samba-client"
    echo "   Arch: sudo pacman -S samba"
    exit 1
fi

# Check if domain is already provisioned
if samba-tool domain info 127.0.0.1 &> /dev/null; then
    echo "✅ Domain already provisioned"
else
    echo "📋 No domain found. You'll need to provision one using the API:"
    echo "   POST /api/v1/domain/provision"
    echo "   {"
    echo "     \"domain\": \"DEV\","
    echo "     \"realm\": \"dev.local\","
    echo "     \"admin_password\": \"DevPass123!\""
    echo "   }"
fi

# Set development environment
export ENV=development

echo "✅ Development environment configured"
echo ""
echo "🎯 Next steps:"
echo "   1. Start the API server: go run main.go"
echo "   2. Access the web interface at http://localhost:3000"
echo "   3. Login with admin/DevPass123! (after domain is provisioned)"
echo ""
echo "💡 The API will now use real Samba commands instead of dummy data"
