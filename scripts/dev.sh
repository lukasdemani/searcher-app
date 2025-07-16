#!/bin/bash

# Website Analyzer - Development Scripts
# Easy commands for development workflow

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_header() {
    echo -e "${BLUE}🚀 Website Analyzer - Development Tools${NC}"
    echo "============================================"
}

print_success() {
    echo -e "${GREEN}✅${NC} $1"
}

print_error() {
    echo -e "${RED}❌${NC} $1"
}

print_info() {
    echo -e "${YELLOW}ℹ️${NC} $1"
}

case "$1" in
    "setup")
        print_header
        print_info "Setting up development environment..."
        
        # Check if Go is installed
        if ! command -v go > /dev/null 2>&1; then
            print_error "Go is not installed. Installing with Homebrew..."
            if command -v brew > /dev/null 2>&1; then
                brew install go
                print_success "Go installed successfully"
            else
                print_error "Homebrew not found. Please install Go manually: https://golang.org/doc/install"
                exit 1
            fi
        else
            print_success "Go is already installed: $(go version)"
        fi
        
        # Setup environment file
        if [ ! -f ".env" ]; then
            cp env.template .env
            print_success "Environment file created from template"
        else
            print_info "Environment file already exists"
        fi
        
        # Install backend dependencies
        print_info "Installing Go dependencies..."
        cd backend && go mod download && cd ..
        print_success "Backend dependencies installed"
        
        # Install frontend dependencies
        print_info "Installing Node.js dependencies..."
        cd frontend && npm install && cd ..
        print_success "Frontend dependencies installed"
        
        print_success "Development environment setup complete!"
        echo ""
        echo "Next steps:"
        echo "  ./scripts/dev.sh start    # Start all services"
        echo "  ./scripts/dev.sh test     # Run all tests"
        ;;
        
    "start")
        print_header
        print_info "Starting development servers..."
        
        # Start database
        print_info "Starting MySQL database..."
        docker-compose up -d mysql
        
        # Wait for database
        print_info "Waiting for database to be ready..."
        sleep 5
        
        # Start backend in background
        print_info "Starting backend server..."
        cd backend
        go run cmd/main.go &
        BACKEND_PID=$!
        cd ..
        
        # Start frontend
        print_info "Starting frontend server..."
        cd frontend
        npm run dev &
        FRONTEND_PID=$!
        cd ..
        
        print_success "All services started!"
        echo ""
        echo "Services running:"
        echo "  📱 Frontend: http://localhost:3000"
        echo "  🔧 Backend:  http://localhost:8080"
        echo "  💾 Database: localhost:3306"
        echo ""
        echo "Press Ctrl+C to stop all services"
        
        # Wait for interrupt
        trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; docker-compose down; exit" INT
        wait
        ;;
        
    "stop")
        print_header
        print_info "Stopping all services..."
        
        # Kill backend and frontend processes
        pkill -f "go run cmd/main.go" 2>/dev/null || true
        pkill -f "react-scripts start" 2>/dev/null || true
        
        # Stop Docker containers
        docker-compose down
        
        print_success "All services stopped"
        ;;
        
    "test")
        print_header
        print_info "Running all tests..."
        
        # Test backend
        if command -v go > /dev/null 2>&1; then
            print_info "Running backend tests..."
            cd backend && go test ./... && cd ..
            print_success "Backend tests passed"
        else
            print_error "Go not installed - skipping backend tests"
        fi
        
        # Test frontend
        print_info "Running frontend tests..."
        cd frontend && npm test -- --coverage --watchAll=false && cd ..
        print_success "Frontend tests passed"
        
        # Integration tests
        print_info "Running integration tests..."
        ./test-setup-basic.sh
        ;;
        
    "build")
        print_header
        print_info "Building production assets..."
        
        # Build backend
        if command -v go > /dev/null 2>&1; then
            print_info "Building backend..."
            cd backend && go build -o ../dist/backend cmd/main.go && cd ..
            print_success "Backend built successfully"
        fi
        
        # Build frontend
        print_info "Building frontend..."
        cd frontend && npm run build && cd ..
        print_success "Frontend built successfully"
        
        print_success "Production build complete!"
        ;;
        
    "clean")
        print_header
        print_info "Cleaning project..."
        
        # Clean backend
        cd backend && go clean && cd ..
        
        # Clean frontend
        cd frontend && rm -rf build node_modules && cd ..
        
        # Clean Docker
        docker-compose down --volumes --remove-orphans
        
        print_success "Project cleaned"
        ;;
        
    *)
        print_header
        echo "Usage: $0 {setup|start|stop|test|build|clean}"
        echo ""
        echo "Commands:"
        echo "  setup  - Install dependencies and setup environment"
        echo "  start  - Start all development servers"
        echo "  stop   - Stop all running services"
        echo "  test   - Run all tests"
        echo "  build  - Build production assets"
        echo "  clean  - Clean build artifacts and dependencies"
        echo ""
        exit 1
        ;;
esac 