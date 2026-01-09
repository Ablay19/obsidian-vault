#!/bin/bash

# Start SSH Management Server and Web Server together

echo "🚀 Starting Obsidian Automation Stack..."
echo "================================"

# Function to cleanup on exit
cleanup() {
    echo "🛑 Shutting down services..."
    docker stop obsidian-ssh-test obsidian-bot-test 2>/dev/null
    docker rm obsidian-ssh-test obsidian-bot-test 2>/dev/null
    echo "✅ Cleanup complete"
    exit 0
}

# Set trap for cleanup
trap cleanup SIGINT SIGTERM

# Check if .env exists
if [ ! -f .env ]; then
    echo "📋 Creating .env from .env.example..."
    cp .env.example .env
fi

# Build SSH Management Server
echo "🔨 Building SSH Management Server..."
docker build -f Dockerfile.ssh -t obsidian-ssh-server . > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ SSH Management Server built successfully"
else
    echo "❌ Failed to build SSH Management Server"
    exit 1
fi

# Build Web Server
echo "🔨 Building Web Server..."
docker build -f Dockerfile.production -t obsidian-bot . > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Web Server built successfully"
else
    echo "❌ Failed to build Web Server"
    exit 1
fi

# Stop any existing containers
echo "🔄 Stopping existing containers..."
docker stop obsidian-ssh-test obsidian-bot-test 2>/dev/null
docker rm obsidian-ssh-test obsidian-bot-test 2>/dev/null

# Start SSH Management Server
echo "🔧 Starting SSH Management Server on port 8081..."
docker run -d \
    --name obsidian-ssh-test \
    -p 8081:8081 \
    --env-file .env \
    obsidian-ssh-server

# Start Web Server
echo "🌐 Starting Web Server on port 8080..."
docker run -d \
    --name obsidian-bot-test \
    -p 8080:8080 \
    --env-file .env \
    --link obsidian-ssh-test:ssh-server \
    obsidian-bot

# Wait for services to be ready
echo "⏳ Waiting for services to start..."
sleep 5

# Check if services are running
echo "🔍 Checking service health..."
SSH_STATUS=$(curl -s http://localhost:8081/health 2>/dev/null | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
WEB_STATUS=$(curl -s http://localhost:8080/health 2>/dev/null | grep -o '"status":"[^"]*"' | cut -d'"' -f4)

if [ "$SSH_STATUS" = "healthy" ]; then
    echo "✅ SSH Management Server is healthy on http://localhost:8081"
else
    echo "❌ SSH Management Server is not responding"
fi

if [ "$WEB_STATUS" = "healthy" ]; then
    echo "✅ Web Server is healthy on http://localhost:8080"
else
    echo "❌ Web Server is not responding"
fi

echo ""
echo "🎉 Obsidian Automation Stack is running!"
echo "================================"
echo "📊 Web Server:     http://localhost:8080"
echo "🔧 SSH Management:  http://localhost:8081"
echo "📋 SSH API:        http://localhost:8081/api/ssh/status"
echo "🔍 Health Checks:"
echo "   Web:  curl http://localhost:8080/health"
echo "   SSH:   curl http://localhost:8081/health"
echo ""
echo "💡 Press Ctrl+C to stop all services"
echo "================================"

# Show logs
echo "📋 Showing logs (Ctrl+C to stop)..."
echo ""

# Function to show logs
show_logs() {
    while true; do
        clear
        echo "🎉 Obsidian Automation Stack - Running Services"
        echo "================================"
        echo "📊 Web Server (Port 8080):"
        echo "--------------------------------"
        docker logs --tail 10 obsidian-bot-test 2>/dev/null || echo "Web server not running"
        echo ""
        echo "🔧 SSH Management (Port 8081):"
        echo "--------------------------------"
        docker logs --tail 10 obsidian-ssh-test 2>/dev/null || echo "SSH server not running"
        echo ""
        echo "Press Ctrl+C to stop services..."
        sleep 10
    done
}

# Start showing logs
show_logs