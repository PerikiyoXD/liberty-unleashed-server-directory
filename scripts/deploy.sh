#!/bin/bash

# Deployment script for Liberty Unleashed Server Directory
# This script deploys the application to a Linux server

set -e

# Configuration
SERVER_HOST=${1:-"localhost"}
SERVER_USER=${2:-"lusd"}
SERVICE_NAME="lusd-server"
INSTALL_DIR="/opt/lusd"
CONFIG_DIR="/etc/lusd"

echo "Deploying Liberty Unleashed Server Directory to $SERVER_HOST"

# Check if binary exists
if [ ! -f "build/lusd-linux-amd64" ]; then
    echo "Error: Binary not found. Please run build script first."
    exit 1
fi

# Create deployment package
echo "Creating deployment package..."
mkdir -p deploy
cp build/lusd-linux-amd64 deploy/lusd
cp configs/config.example.json deploy/config.json
cp web/index.html deploy/
cp systemd/lusd-server.service deploy/

# Upload files to server
echo "Uploading files to server..."
scp -r deploy/ $SERVER_USER@$SERVER_HOST:/tmp/lusd-deploy/

# Execute deployment commands on server
echo "Installing on server..."
ssh $SERVER_USER@$SERVER_HOST << 'EOF'
    # Stop service if running
    sudo systemctl stop lusd-server || true
    
    # Create user if doesn't exist
    if ! id lusd &>/dev/null; then
        sudo useradd -r -s /bin/false lusd
    fi
    
    # Create directories
    sudo mkdir -p /opt/lusd /etc/lusd /var/log/lusd
    
    # Copy files
    sudo cp /tmp/lusd-deploy/lusd /opt/lusd/
    sudo cp /tmp/lusd-deploy/index.html /opt/lusd/
    sudo cp /tmp/lusd-deploy/config.json /etc/lusd/
    sudo cp /tmp/lusd-deploy/lusd-server.service /etc/systemd/system/
    
    # Set permissions
    sudo chown -R lusd:lusd /opt/lusd /var/log/lusd
    sudo chmod +x /opt/lusd/lusd
    sudo chmod 644 /etc/lusd/config.json
    
    # Link config
    sudo ln -sf /etc/lusd/config.json /opt/lusd/config.json
    
    # Reload systemd and start service
    sudo systemctl daemon-reload
    sudo systemctl enable lusd-server
    sudo systemctl start lusd-server
    
    # Cleanup
    rm -rf /tmp/lusd-deploy
    
    echo "Deployment completed successfully!"
    sudo systemctl status lusd-server
EOF

echo "Deployment finished!"
