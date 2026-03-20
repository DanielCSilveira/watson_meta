#!/bin/bash

# Deploy Script for Watson WhatsApp API
# Usage: ./deploy.sh [build|run|stop|restart|logs|clean]

set -e

PROJECT_NAME="watson-whatsapp-api"
CONTAINER_NAME="watson-api"
IMAGE_NAME="watson-whatsapp-api"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker not found. Please install Docker first."
        exit 1
    fi
    log_success "Docker is installed"
}

build_image() {
    log_info "Building Docker image..."
    cd api
    docker build -t $IMAGE_NAME .
    cd ..
    log_success "Image built successfully"
}

run_container() {
    log_info "Starting container..."
    
    if [ ! -f .env ]; then
        log_warning ".env file not found. Using .env.example"
        cp .env.example .env
        log_error "Please edit .env file with your credentials and run again"
        exit 1
    fi
    
    # Check if container already exists
    if [ "$(docker ps -aq -f name=$CONTAINER_NAME)" ]; then
        log_warning "Container already exists. Stopping and removing..."
        docker stop $CONTAINER_NAME 2>/dev/null || true
        docker rm $CONTAINER_NAME 2>/dev/null || true
    fi
    
    docker run -d \
        --name $CONTAINER_NAME \
        -p 8080:8080 \
        --env-file .env \
        --restart unless-stopped \
        $IMAGE_NAME
    
    log_success "Container started: $CONTAINER_NAME"
    log_info "API running at http://localhost:8080"
    log_info "Swagger at http://localhost:8080/swagger/"
}

stop_container() {
    log_info "Stopping container..."
    docker stop $CONTAINER_NAME 2>/dev/null || log_warning "Container not running"
    log_success "Container stopped"
}

restart_container() {
    stop_container
    sleep 2
    run_container
}

show_logs() {
    log_info "Showing logs (Ctrl+C to exit)..."
    docker logs -f $CONTAINER_NAME
}

clean_all() {
    log_warning "Cleaning all containers and images..."
    docker stop $CONTAINER_NAME 2>/dev/null || true
    docker rm $CONTAINER_NAME 2>/dev/null || true
    docker rmi $IMAGE_NAME 2>/dev/null || true
    log_success "Cleanup complete"
}

compose_up() {
    log_info "Starting with Docker Compose..."
    docker-compose up -d
    log_success "Services started"
    docker-compose ps
}

compose_down() {
    log_info "Stopping Docker Compose services..."
    docker-compose down
    log_success "Services stopped"
}

compose_logs() {
    log_info "Showing Docker Compose logs (Ctrl+C to exit)..."
    docker-compose logs -f
}

show_status() {
    log_info "Container Status:"
    docker ps -f name=$CONTAINER_NAME
    echo ""
    log_info "Recent Logs:"
    docker logs --tail 20 $CONTAINER_NAME
}

show_help() {
    echo "Watson WhatsApp API - Deploy Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  build       Build Docker image"
    echo "  run         Run container"
    echo "  stop        Stop container"
    echo "  restart     Restart container"
    echo "  logs        Show container logs"
    echo "  status      Show container status"
    echo "  clean       Remove container and image"
    echo ""
    echo "Docker Compose:"
    echo "  up          Start with docker-compose"
    echo "  down        Stop docker-compose services"
    echo "  compose-logs Show docker-compose logs"
    echo ""
    echo "Examples:"
    echo "  $0 build    # Build image"
    echo "  $0 run      # Start container"
    echo "  $0 logs     # View logs"
}

# Main
check_docker

case "${1:-help}" in
    build)
        build_image
        ;;
    run)
        build_image
        run_container
        ;;
    stop)
        stop_container
        ;;
    restart)
        restart_container
        ;;
    logs)
        show_logs
        ;;
    status)
        show_status
        ;;
    clean)
        clean_all
        ;;
    up|compose-up)
        compose_up
        ;;
    down|compose-down)
        compose_down
        ;;
    compose-logs)
        compose_logs
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
