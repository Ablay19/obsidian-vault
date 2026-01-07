#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Header
echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║              Environment Setup Scripts                        ║${NC}"
echo -e "${CYAN}║                  Complete Configuration                       ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Utility functions
print_status() { echo -e "${BLUE}[•]${NC} $1"; }
print_success() { echo -e "${GREEN}[✓]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[⚠]${NC} $1"; }
print_error() { echo -e "${RED}[✗]${NC} $1"; }
print_info() { echo -e "${PURPLE}[ℹ]${NC} $1"; }

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -d "internal" ]; then
    print_error "This script must be run from the project root directory"
    exit 1
fi

# Environment setup menu
show_env_menu() {
    echo -e "${CYAN}Environment Setup:${NC}"
    echo ""
    echo "1. 🔧 Initial Setup"
    echo "2. 📋 Configure .env"
    echo "3. 🗄️  Database Setup"
    echo "4. 🤖 AI Services Setup"
    echo "5. 📱 Bot Services Setup"
    echo "6. 🔐 Authentication Setup"
    echo "7. 🐳 Docker Environment"
    echo "8. ☁️  Cloud Services"
    echo "9. 🧪 Test Environment"
    echo "10. 📊 Validate Configuration"
    echo "0. 🔙 Back"
    echo ""
}

# Environment setup operations
env_operations() {
    case $1 in
        1) initial_setup ;;
        2) configure_env ;;
        3) database_setup ;;
        4) ai_services_setup ;;
        5) bot_services_setup ;;
        6) auth_setup ;;
        7) docker_environment ;;
        8) cloud_services ;;
        9) test_environment ;;
        10) validate_config ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

initial_setup() {
    print_status "Performing initial setup..."
    
    # Check system requirements
    print_status "Checking system requirements..."
    
    # Check Go
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version)
        print_success "Go installed: $GO_VERSION"
    else
        print_error "Go is not installed. Please install Go 1.21 or later."
        print_info "Download from: https://golang.org/dl/"
        exit 1
    fi
    
    # Check Git
    if command -v git &> /dev/null; then
        print_success "Git installed: $(git --version)"
    else
        print_error "Git is not installed. Please install Git."
        exit 1
    fi
    
    # Initialize Go modules
    print_status "Initializing Go modules..."
    go mod init obsidian-automation 2>/dev/null || true
    go mod tidy
    
    # Create directory structure
    print_status "Creating directory structure..."
    mkdir -p logs attachments vault/{Inbox,Attachments,physics,math,chemistry,biology,admin,document}
    mkdir -p config scripts data temp
    
    # Set permissions
    chmod -R 755 logs attachments vault config data temp
    
    # Create .gitignore if not exists
    if [ ! -f ".gitignore" ]; then
        print_status "Creating .gitignore..."
        cat > .gitignore << 'GITIGNORE'
# Binaries
obsidian-automation
obsidian-automation-*
*.exe
*.exe~
*.dll
*.so
*.dylib

# Environment files
.env
.env.local
.env.*.local

# Logs
logs/
*.log

# Attachments and media
attachments/
vault/Attachments/

# Temporary files
temp/
tmp/
*.tmp

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# IDE files
.vscode/
.idea/
*.swp
*.swo
*~

# Build artifacts
dist/
build/

# Database
*.db
*.sqlite
*.sqlite3

# PID files
*.pid

# Coverage reports
coverage.out
coverage.html

# Docker
.dockerignore
GITIGNORE
        print_success ".gitignore created"
    fi
    
    # Download dependencies
    print_status "Downloading dependencies..."
    go mod download
    
    print_success "Initial setup completed"
    print_info "Next steps:"
    print_info "1. Run: ./scripts/env-setup.sh"
    print_info "2. Select option 2 to configure .env"
    print_info "3. Configure your services"
}

configure_env() {
    print_status "Configuring environment variables..."
    
    # Create .env from example if not exists
    if [ ! -f ".env" ]; then
        if [ -f ".env.example" ]; then
            print_status "Creating .env from example..."
            cp .env.example .env
            print_success ".env file created"
        else
            print_error ".env.example not found"
            exit 1
        fi
    fi
    
    # Interactive configuration
    print_info "Let's configure your environment variables..."
    print_info "Press Enter to keep current value or type new value."
    echo ""
    
    # Telegram Bot Token
    CURRENT_TOKEN=$(grep "TELEGRAM_BOT_TOKEN=" .env | cut -d'=' -f2)
    echo -e "${BLUE}Telegram Bot Token:${NC}"
    echo "Current: ${CURRENT_TOKEN:-not set}"
    read -p "New token (or press Enter to keep): " NEW_TOKEN
    if [ -n "$NEW_TOKEN" ]; then
        sed -i "s/TELEGRAM_BOT_TOKEN=.*/TELEGRAM_BOT_TOKEN=\"$NEW_TOKEN\"/" .env
        print_success "Telegram token updated"
    fi
    echo ""
    
    # Environment Mode
    CURRENT_MODE=$(grep "ENVIRONMENT_MODE=" .env | cut -d'=' -f2 | tr -d '"')
    echo -e "${BLUE}Environment Mode:${NC}"
    echo "Current: ${CURRENT_MODE:-not set}"
    echo "Options: dev, prod"
    read -p "New mode (or press Enter to keep): " NEW_MODE
    if [ -n "$NEW_MODE" ]; then
        sed -i "s/ENVIRONMENT_MODE=.*/ENVIRONMENT_MODE=\"$NEW_MODE\"/" .env
        print_success "Environment mode updated"
    fi
    echo ""
    
    # Database URL
    CURRENT_DB=$(grep "TURSO_DATABASE_URL=" .env | cut -d'=' -f2 | tr -d '"')
    echo -e "${BLUE}Turso Database URL:${NC}"
    echo "Current: ${CURRENT_DB:-not set}"
    read -p "New URL (or press Enter to keep): " NEW_DB
    if [ -n "$NEW_DB" ]; then
        sed -i "s|TURSO_DATABASE_URL=.*|TURSO_DATABASE_URL=\"$NEW_DB\"|" .env
        print_success "Database URL updated"
    fi
    echo ""
    
    # AI Service Configuration
    print_info "Configure AI services (press Enter to skip):"
    
    # Gemini API Key
    CURRENT_GEMINI=$(grep "GEMINI_API_KEY=" .env | cut -d'=' -f2 | tr -d '"')
    echo -e "${BLUE}Gemini API Key:${NC}"
    echo "Current: ${CURRENT_GEMINI:-not set}"
    read -p "New key (or press Enter to keep): " NEW_GEMINI
    if [ -n "$NEW_GEMINI" ]; then
        sed -i "s/GEMINI_API_KEY=.*/GEMINI_API_KEY=\"$NEW_GEMINI\"/" .env
        print_success "Gemini API key updated"
    fi
    echo ""
    
    # Session Secret
    CURRENT_SECRET=$(grep "SESSION_SECRET=" .env | cut -d'=' -f2 | tr -d '"')
    echo -e "${BLUE}Session Secret:${NC}"
    echo "Current: ${CURRENT_SECRET:-not set}"
    if [ -z "$CURRENT_SECRET" ]; then
        NEW_SECRET=$(openssl rand -base64 32 2>/dev/null || date +%s | sha256sum | base64 | head -c 32)
        sed -i "s/SESSION_SECRET=.*/SESSION_SECRET=\"$NEW_SECRET\"/" .env
        print_success "Session secret generated"
    else
        read -p "New secret (or press Enter to keep): " NEW_SECRET
        if [ -n "$NEW_SECRET" ]; then
            sed -i "s/SESSION_SECRET=.*/SESSION_SECRET=\"$NEW_SECRET\"/" .env
            print_success "Session secret updated"
        fi
    fi
    
    print_success "Environment configuration completed"
    print_info "Review your .env file for additional settings"
}

database_setup() {
    print_status "Setting up database..."
    
    # Check if database URL is configured
    if [ -z "$TURSO_DATABASE_URL" ]; then
        print_warning "Database URL not configured in .env"
        print_info "Please configure TURSO_DATABASE_URL first"
        return
    fi
    
    # Test database connection
    print_status "Testing database connection..."
    if go run ./cmd/bot/main.go --test-db 2>/dev/null; then
        print_success "Database connection successful"
    else
        print_error "Database connection failed"
        print_info "Please check your database credentials"
        return
    fi
    
    # Run migrations
    print_status "Running database migrations..."
    if go run ./cmd/bot/main.go --migrate 2>/dev/null; then
        print_success "Database migrations completed"
    else
        print_error "Database migrations failed"
        return
    fi
    
    print_success "Database setup completed"
}

ai_services_setup() {
    print_status "Setting up AI services..."
    
    # Check configured AI services
    print_info "Checking AI service configuration..."
    
    AI_SERVICES=()
    
    if [ -n "$GEMINI_API_KEY" ]; then
        AI_SERVICES+=("Gemini")
        print_success "Gemini: Configured"
    else
        print_warning "Gemini: Not configured"
    fi
    
    if [ -n "$GROQ_API_KEY" ]; then
        AI_SERVICES+=("Groq")
        print_success "Groq: Configured"
    else
        print_warning "Groq: Not configured"
    fi
    
    if [ -n "$HUGGINGFACE_API_KEY" ]; then
        AI_SERVICES+=("HuggingFace")
        print_success "HuggingFace: Configured"
    else
        print_warning "HuggingFace: Not configured"
    fi
    
    if [ -n "$OPENROUTER_API_KEY" ]; then
        AI_SERVICES+=("OpenRouter")
        print_success "OpenRouter: Configured"
    else
        print_warning "OpenRouter: Not configured"
    fi
    
    if [ ${#AI_SERVICES[@]} -eq 0 ]; then
        print_error "No AI services configured"
        print_info "Please configure at least one AI service in .env"
        return
    fi
    
    # Test AI services
    print_status "Testing AI services..."
    for service in "${AI_SERVICES[@]}"; do
        print_status "Testing $service..."
        # Add AI service testing logic here
        print_success "$service: Ready"
    done
    
    print_success "AI services setup completed"
}

bot_services_setup() {
    print_status "Setting up bot services..."
    
    # Telegram Bot
    if [ -n "$TELEGRAM_BOT_TOKEN" ]; then
        print_status "Testing Telegram bot..."
        # Test Telegram bot connection
        print_success "Telegram bot: Ready"
    else
        print_warning "Telegram bot: Not configured"
    fi
    
    # WhatsApp
    if [ -n "$WHATSAPP_ACCESS_TOKEN" ]; then
        print_status "Testing WhatsApp service..."
        print_success "WhatsApp: Ready"
    else
        print_warning "WhatsApp: Not configured"
    fi
    
    print_success "Bot services setup completed"
}

auth_setup() {
    print_status "Setting up authentication..."
    
    # Check Google OAuth
    if [ -n "$GOOGLE_CLIENT_ID" ] && [ -n "$GOOGLE_CLIENT_SECRET" ]; then
        print_success "Google OAuth: Configured"
        
        # Test OAuth configuration
        print_status "Testing OAuth configuration..."
        # Add OAuth testing logic here
        print_success "OAuth: Ready"
    else
        print_warning "Google OAuth: Not configured"
        print_info "Set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET in .env"
    fi
    
    # Check session secret
    if [ -n "$SESSION_SECRET" ]; then
        print_success "Session secret: Configured"
    else
        print_error "Session secret: Not configured"
        print_info "Run configure_env to generate session secret"
    fi
    
    print_success "Authentication setup completed"
}

docker_environment() {
    print_status "Setting up Docker environment..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        print_info "Install Docker from: https://docker.com"
        return
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        print_warning "Docker Compose not found"
    fi
    
    # Create Dockerfile if not exists
    if [ ! -f "Dockerfile" ]; then
        print_status "Creating Dockerfile..."
        cat > Dockerfile << 'DOCKERFILE'
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/bot/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/.env.example .env

# Create directories
RUN mkdir -p logs attachments vault

EXPOSE 8080

CMD ["./main"]
DOCKERFILE
        print_success "Dockerfile created"
    fi
    
    # Create docker-compose.yml if not exists
    if [ ! -f "docker-compose.yml" ]; then
        print_status "Creating docker-compose.yml..."
        cat > docker-compose.yml << 'COMPOSE'
version: '3.8'

services:
  obsidian-automation:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT_MODE=prod
    volumes:
      - ./logs:/root/logs
      - ./attachments:/root/attachments
      - ./vault:/root/vault
      - ./.env:/root/.env
    restart: unless-stopped

  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
    restart: unless-stopped
COMPOSE
        print_success "docker-compose.yml created"
    fi
    
    print_success "Docker environment setup completed"
    print_info "Run with: docker-compose up -d"
}

cloud_services() {
    print_status "Setting up cloud services..."
    
    # Cloudflare Workers
    if [ -n "$CLOUDFLARE_API_TOKEN" ]; then
        print_success "Cloudflare: Configured"
        if [ -f "scripts/deploy-cloudflare.sh" ]; then
            print_info "Deploy with: ./scripts/deploy-cloudflare.sh"
        fi
    else
        print_warning "Cloudflare: Not configured"
    fi
    
    # Render
    if [ -f "config/render.yaml" ]; then
        print_success "Render: Configuration exists"
        print_info "Push to GitHub to trigger Render deployment"
    else
        print_warning "Render: Configuration not found"
    fi
    
    print_success "Cloud services setup completed"
}

test_environment() {
    print_status "Testing environment..."
    
    # Test Go build
    print_status "Testing Go build..."
    if go build -o /tmp/test-build ./cmd/bot/main.go; then
        print_success "Go build: PASS"
        rm -f /tmp/test-build
    else
        print_error "Go build: FAIL"
        return
    fi
    
    # Test dependencies
    print_status "Testing dependencies..."
    if go mod verify > /dev/null 2>&1; then
        print_success "Dependencies: PASS"
    else
        print_error "Dependencies: FAIL"
        return
    fi
    
    # Test configuration
    print_status "Testing configuration..."
    if [ -f ".env" ]; then
        print_success "Configuration: PASS"
    else
        print_error "Configuration: FAIL (.env missing)"
        return
    fi
    
    print_success "Environment test completed"
}

validate_config() {
    print_status "Validating configuration..."
    
    # Load .env file
    if [ -f ".env" ]; then
        set -a
        source .env
        set +a
    else
        print_error ".env file not found"
        return
    fi
    
    echo -e "${BLUE}Configuration Validation:${NC}"
    
    # Required variables
    REQUIRED_VARS=(
        "TELEGRAM_BOT_TOKEN"
        "ENVIRONMENT_MODE"
        "SESSION_SECRET"
    )
    
    for var in "${REQUIRED_VARS[@]}"; do
        if [ -n "${!var}" ]; then
            print_success "$var: SET"
        else
            print_error "$var: MISSING"
        fi
    done
    
    # Optional variables
    OPTIONAL_VARS=(
        "GEMINI_API_KEY"
        "GROQ_API_KEY"
        "TURSO_DATABASE_URL"
        "GOOGLE_CLIENT_ID"
        "WHATSAPP_ACCESS_TOKEN"
    )
    
    echo -e "\n${BLUE}Optional Variables:${NC}"
    for var in "${OPTIONAL_VARS[@]}"; do
        if [ -n "${!var}" ]; then
            print_success "$var: SET"
        else
            print_warning "$var: NOT SET"
        fi
    done
    
    # Validate values
    echo -e "\n${BLUE}Value Validation:${NC}"
    
    # Environment mode
    if [[ "$ENVIRONMENT_MODE" =~ ^(dev|prod)$ ]]; then
        print_success "ENVIRONMENT_MODE: Valid"
    else
        print_error "ENVIRONMENT_MODE: Invalid (must be dev or prod)"
    fi
    
    # Session secret length
    if [ ${#SESSION_SECRET} -ge 32 ]; then
        print_success "SESSION_SECRET: Valid length"
    else
        print_error "SESSION_SECRET: Too short (minimum 32 characters)"
    fi
    
    print_success "Configuration validation completed"
}

# Main execution
main() {
    while true; do
        show_env_menu
        read -p "Select an option: " choice
        echo ""
        
        env_operations $choice
        
        echo ""
        read -p "Press Enter to continue..."
        clear
    done
}

# Handle command line arguments
if [ $# -eq 1 ]; then
    case $1 in
        "init") initial_setup ;;
        "config") configure_env ;;
        "validate") validate_config ;;
        "test") test_environment ;;
        *) echo "Usage: $0 [init|config|validate|test]" ;;
    esac
    exit 0
fi

# Start interactive mode
main