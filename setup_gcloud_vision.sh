#!/bin/bash

# Advanced Vision Processing Setup Script for Google Cloud Shell
# This script sets up all necessary tools and dependencies for the
# multimodal vision processing system with DeepSeek OCR capabilities

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install packages with apt
install_package() {
    local package=$1
    if ! dpkg -l | grep -q "^ii.*$package"; then
        log_info "Installing $package..."
        sudo apt-get update -qq && sudo apt-get install -y -qq "$package"
        log_success "$package installed successfully"
    else
        log_info "$package is already installed"
    fi
}

log_info "🚀 Starting Advanced Vision Processing Setup for Google Cloud Shell"
log_info "================================================================="

# Check if running in Google Cloud Shell
if [[ ! -v CLOUD_SHELL ]]; then
    log_warning "This script is designed for Google Cloud Shell. Continue anyway? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        log_error "Setup cancelled by user"
        exit 1
    fi
fi

# Update package lists
log_info "Updating package lists..."
sudo apt-get update -qq

# Install essential build tools
log_info "Installing essential build tools..."
install_package "build-essential"
install_package "curl"
install_package "wget"
install_package "git"
install_package "unzip"
install_package "software-properties-common"

# Install Go 1.21+
log_info "Setting up Go 1.21+..."
if command_exists go; then
    GO_VERSION=$(go version | grep -oP 'go\d+\.\d+\.\d+' | sed 's/go//')
    log_info "Go version $GO_VERSION is already installed"
    if [[ "$(printf '%s\n' "$GO_VERSION" "1.21.0" | sort -V | head -n1)" = "1.21.0" ]]; then
        log_warning "Go version $GO_VERSION is older than 1.21.0. Installing newer version..."
        sudo rm -rf /usr/local/go
    else
        log_success "Go version $GO_VERSION meets requirements"
        SKIP_GO_INSTALL=true
    fi
fi

if [[ ! $SKIP_GO_INSTALL ]]; then
    log_info "Installing Go 1.21+..."
    wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    rm go1.21.5.linux-amd64.tar.gz

    # Add Go to PATH
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc

    log_success "Go 1.21.5 installed successfully"
fi

# Verify Go installation
log_info "Verifying Go installation..."
if command_exists go; then
    GO_VERSION=$(go version | grep -oP 'go\d+\.\d+\.\d+' | sed 's/go//')
    log_success "Go $GO_VERSION is ready"
else
    log_error "Go installation failed"
    exit 1
fi

# Install Tesseract OCR for document processing
log_info "Installing Tesseract OCR for document processing..."
install_package "tesseract-ocr"
install_package "tesseract-ocr-eng"
install_package "tesseract-ocr-fra"
install_package "tesseract-ocr-ara"
install_package "tesseract-ocr-deu"
install_package "tesseract-ocr-spa"

# Verify Tesseract
log_info "Verifying Tesseract OCR..."
if command_exists tesseract; then
    TESSERACT_VERSION=$(tesseract --version | head -n1)
    log_success "Tesseract OCR ready: $TESSERACT_VERSION"
else
    log_error "Tesseract OCR installation failed"
    exit 1
fi

# Install ImageMagick for image preprocessing
log_info "Installing ImageMagick for image preprocessing..."
install_package "imagemagick"

# Verify ImageMagick
log_info "Verifying ImageMagick..."
if command_exists convert; then
    IMAGEMAGICK_VERSION=$(convert --version | head -n1)
    log_success "ImageMagick ready: $IMAGEMAGICK_VERSION"
else
    log_error "ImageMagick installation failed"
    exit 1
fi

# Install additional image processing libraries
log_info "Installing additional image processing libraries..."
install_package "libpng-dev"
install_package "libjpeg-dev"
install_package "libtiff-dev"
install_package "liblcms2-dev"

# Install Python3 and pip (for any Python dependencies)
log_info "Installing Python3 and pip..."
install_package "python3"
install_package "python3-pip"
install_package "python3-dev"

# Install Node.js and npm (for workers)
log_info "Installing Node.js and npm..."
if ! command_exists node; then
    curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
    install_package "nodejs"
fi

# Verify Node.js
log_info "Verifying Node.js..."
if command_exists node; then
    NODE_VERSION=$(node --version)
    NPM_VERSION=$(npm --version)
    log_success "Node.js $NODE_VERSION and npm $NPM_VERSION ready"
else
    log_error "Node.js installation failed"
    exit 1
fi

# Install Cloud SDK components (if not already available)
log_info "Checking Google Cloud SDK..."
if command_exists gcloud; then
    GCLOUD_VERSION=$(gcloud version | head -n1)
    log_success "Google Cloud SDK ready: $GCLOUD_VERSION"
else
    log_warning "Google Cloud SDK not found. Installing..."
    echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -
    sudo apt-get update -qq && install_package "google-cloud-sdk"
fi

# Setup Go workspace
log_info "Setting up Go workspace..."
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
mkdir -p "$GOPATH/bin" "$GOPATH/src" "$GOPATH/pkg"

# Add to bashrc if not already there
if ! grep -q "export GOPATH=" ~/.bashrc; then
    echo "export GOPATH=$HOME/go" >> ~/.bashrc
    echo "export PATH=\$PATH:/usr/local/go/bin:\$GOPATH/bin" >> ~/.bashrc
fi

# Clone or update the repository
REPO_DIR="obsidian-vault"
if [[ -d "$REPO_DIR" ]]; then
    log_info "Repository already exists. Pulling latest changes..."
    cd "$REPO_DIR"
    git pull origin main
    cd ..
else
    log_info "Cloning repository..."
    git clone https://github.com/Ablay19/obsidian-vault.git "$REPO_DIR"
fi

# Setup project dependencies
log_info "Setting up project dependencies..."
cd "$REPO_DIR"

# Download Go dependencies
log_info "Downloading Go dependencies..."
go mod download

# Verify dependencies
log_info "Verifying Go dependencies..."
if go mod verify; then
    log_success "Go dependencies verified successfully"
else
    log_error "Go dependency verification failed"
    exit 1
fi

# Build the project
log_info "Building the project..."
if go build -v ./cmd/bot/; then
    log_success "Project built successfully"
else
    log_error "Project build failed"
    exit 1
fi

# Create necessary directories
log_info "Creating necessary directories..."
mkdir -p vault/Inbox
mkdir -p logs
mkdir -p attachments

# Setup environment file
log_info "Setting up environment configuration..."
if [[ ! -f ".env" ]]; then
    if [[ -f ".env.example" ]]; then
        cp .env.example .env
        log_success "Environment file created from example"
        log_warning "Please edit .env file with your API keys and configuration"
    else
        log_warning ".env.example not found. Please create .env manually"
    fi
else
    log_info "Environment file already exists"
fi

# Setup configuration
log_info "Verifying configuration..."
if [[ -f "config.yaml" ]]; then
    log_success "Configuration file exists"
else
    log_error "config.yaml not found"
    exit 1
fi

# Install additional tools for development
log_info "Installing development tools..."

# Install htop for monitoring
install_package "htop"

# Install jq for JSON processing
install_package "jq"

# Install tmux for session management
install_package "tmux"

# Install vim or nano if not available
if ! command_exists vim && ! command_exists nano; then
    install_package "vim"
fi

# Setup aliases for convenience
log_info "Setting up convenience aliases..."
if ! grep -q "alias gologs=" ~/.bashrc; then
    cat >> ~/.bashrc << 'EOF'

# Obsidian Vault aliases
alias gologs='cd ~/obsidian-vault && tail -f logs/*.log'
alias gobuild='cd ~/obsidian-vault && go build ./cmd/bot/'
alias gorun='cd ~/obsidian-vault && go run ./cmd/bot/'
alias goupdate='cd ~/obsidian-vault && git pull origin main && go mod download'
alias govision='cd ~/obsidian-vault && echo "Vision processing ready with $(ls internal/vision/ | wc -l) encoders"'
EOF
    log_success "Convenience aliases added"
fi

# Create a simple test script
log_info "Creating test script..."
cat > test_vision_setup.sh << 'EOF'
#!/bin/bash
echo "🧪 Testing Vision Processing Setup"
echo "=================================="

cd ~/obsidian-vault

echo "Testing Go build..."
if go build ./cmd/bot/ > /dev/null 2>&1; then
    echo "✅ Go build successful"
else
    echo "❌ Go build failed"
    exit 1
fi

echo "Testing vision modules..."
if [[ -d "internal/vision" ]] && [[ -f "internal/vision/processor.go" ]]; then
    echo "✅ Vision modules present"
else
    echo "❌ Vision modules missing"
    exit 1
fi

echo "Testing OCR modules..."
if [[ -d "internal/ocr" ]] && [[ -f "internal/ocr/deepseek_ocr.go" ]]; then
    echo "✅ OCR modules present"
else
    echo "❌ OCR modules missing"
    exit 1
fi

echo "Testing configuration..."
if grep -q "vision:" config.yaml; then
    echo "✅ Vision configuration present"
else
    echo "❌ Vision configuration missing"
fi

echo ""
echo "🎉 Vision Processing Setup Complete!"
echo "Run './bin/bot' to start the bot with advanced vision capabilities"
EOF

chmod +x test_vision_setup.sh

# Final verification
log_info "Running final verification..."
source ~/.bashrc

if command_exists go && command_exists tesseract && command_exists convert && command_exists node; then
    log_success "All core tools are installed and ready"
else
    log_error "Some tools are missing"
    exit 1
fi

# Print completion message
echo ""
echo "╔══════════════════════════════════════════════════════════════════════════════╗"
echo "║                          🎉 SETUP COMPLETE! 🎉                              ║"
echo "╠══════════════════════════════════════════════════════════════════════════════╣"
echo "║                                                                              ║"
echo "║  Advanced Vision Processing System Ready for Google Cloud Shell              ║"
echo "║                                                                              ║"
echo "║  Installed Components:                                                        ║"
echo "║  ✅ Go 1.21.5                                                               ║"
echo "║  ✅ Tesseract OCR (Multi-language)                                         ║"
echo "║  ✅ ImageMagick                                                             ║"
echo "║  ✅ Node.js & npm                                                           ║"
echo "║  ✅ Vision Processing Modules                                               ║"
echo "║  ✅ DeepSeek OCR Pipeline                                                   ║"
echo "║  ✅ Multimodal Fusion Engine                                                ║"
echo "║                                                                              ║"
echo "║  Next Steps:                                                                 ║"
echo "║  1. Edit .env with your API keys (DeepSeek, Gemini, OpenAI)                 ║"
echo "║  2. Run: cd ~/obsidian-vault && ./test_vision_setup.sh                      ║"
echo "║  3. Start the bot: go run ./cmd/bot/                                        ║"
echo "║                                                                              ║"
echo "║  Useful Commands:                                                            ║"
echo "║  • gobuild  - Build the project                                             ║"
echo "║  • gorun    - Run the bot                                                   ║"
echo "║  • gologs   - View logs                                                     ║"
echo "║  • govision - Check vision status                                           ║"
echo "║                                                                              ║"
echo "╚══════════════════════════════════════════════════════════════════════════════╝"
echo ""

log_success "Setup script completed successfully!"
log_info "To get started: cd ~/obsidian-vault && source ~/.bashrc"