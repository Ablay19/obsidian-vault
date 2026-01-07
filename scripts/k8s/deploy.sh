#!/bin/bash

set -e

echo "☸️  Obsidian Bot Kubernetes Deployment"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Configuration
ENVIRONMENT=${1:-production}
NAMESPACE=${2:-obsidian-system}
REGISTRY=${3:-gcr.io/obsidian-bot-prod}
IMAGE_TAG=${4:-v1.0.0}

print_info "Configuration:"
echo "  Environment: $ENVIRONMENT"
echo "  Namespace: $NAMESPACE"
echo "  Registry: $REGISTRY"
echo "  Image Tag: $IMAGE_TAG"

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed"
        print_info "Install from: https://kubernetes.io/docs/tasks/tools/"
        exit 1
    fi
    
    # Check kustomize
    if ! command -v kustomize &> /dev/null; then
        print_error "kustomize is not installed"
        print_info "Install from: https://kubectl.docs.kubernetes.io/installation/kustomize/"
        exit 1
    fi
    
    # Check cluster access
    if ! kubectl cluster-info &> /dev/null; then
        print_error "Cannot access Kubernetes cluster"
        print_info "Check your kubeconfig file"
        exit 1
    fi
    
    print_status "Prerequisites satisfied"
}

# Build and push Docker image
build_image() {
    print_info "Building Docker image..."
    
    # Build image
    docker build -f Dockerfile.production -t ${REGISTRY}/obsidian-bot:${IMAGE_TAG} .
    
    print_status "Docker image built"
    
    # Push image
    print_info "Pushing Docker image..."
    
    # Check if logged in to registry
    if [ "$REGISTRY" = "gcr.io"* ]; then
        if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "."; then
            print_error "Not logged in to Google Cloud"
            print_info "Run: gcloud auth login"
            exit 1
        fi
        
        gcloud auth configure-docker ${REGISTRY}
        docker push ${REGISTRY}/obsidian-bot:${IMAGE_TAG}
    else
        docker push ${REGISTRY}/obsidian-bot:${IMAGE_TAG}
    fi
    
    print_status "Docker image pushed"
}

# Create namespace
create_namespace() {
    print_info "Creating namespace..."
    
    if ! kubectl get namespace $NAMESPACE &> /dev/null; then
        kubectl create namespace $NAMESPACE
        print_status "Namespace $NAMESPACE created"
    else
        print_status "Namespace $NAMESPACE already exists"
    fi
}

# Apply secrets
apply_secrets() {
    print_info "Applying secrets..."
    
    # Check if secrets file exists
    if [ ! -f "k8s/base/secrets.yaml" ]; then
        print_error "Secrets file not found: k8s/base/secrets.yaml"
        print_info "Create secrets with: ./scripts/create-secrets.sh"
        exit 1
    fi
    
    # Apply secrets
    kubectl apply -f k8s/base/secrets.yaml -n $NAMESPACE
    print_status "Secrets applied"
}

# Apply configurations
apply_config() {
    print_info "Applying configurations..."
    
    # Apply using kustomize
    if [ -d "k8s/overlays/$ENVIRONMENT" ]; then
        print_info "Using $ENVIRONMENT overlay"
        kustomize build k8s/overlays/$ENVIRONMENT | kubectl apply -f -
    else
        print_warning "Environment overlay not found, using base"
        kubectl apply -f k8s/base/ -n $NAMESPACE
    fi
    
    print_status "Configurations applied"
}

# Wait for rollout
wait_for_rollout() {
    print_info "Waiting for deployment rollout..."
    
    # Wait for deployment to be ready
    kubectl rollout status deployment/obsidian-bot -n $NAMESPACE --timeout=300s
    
    print_status "Deployment completed"
}

# Verify deployment
verify_deployment() {
    print_info "Verifying deployment..."
    
    # Check pod status
    local pod_status=$(kubectl get pods -n $NAMESPACE -l app=obsidian-bot -o jsonpath='{.items[0].status.phase}')
    if [ "$pod_status" = "Running" ]; then
        print_status "Pods are running"
    else
        print_warning "Pods status: $pod_status"
    fi
    
    # Check service status
    if kubectl get service obsidian-bot-service -n $NAMESPACE &> /dev/null; then
        local service_ip=$(kubectl get service obsidian-bot-service -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        if [ -n "$service_ip" ]; then
            print_status "Service accessible at: http://$service_ip"
        else
            local node_port=$(kubectl get service obsidian-bot-service -n $NAMESPACE -o jsonpath='{.spec.ports[0].nodePort}')
            if [ -n "$node_port" ]; then
                print_status "Service accessible via NodePort: $node_port"
            else
                print_status "Service created (internal access only)"
            fi
        fi
    fi
}

# Show status
show_status() {
    print_info "Deployment status:"
    echo ""
    kubectl get pods -n $NAMESPACE -l app=obsidian-bot
    echo ""
    kubectl get services -n $NAMESPACE -l app=obsidian-bot
    echo ""
    kubectl get hpa -n $NAMESPACE -l app=obsidian-bot
    echo ""
    
    # Show logs if requested
    if [ "$SHOW_LOGS" = "true" ]; then
        print_info "Recent logs:"
        kubectl logs -n $NAMESPACE -l app=obsidian-bot --tail=50
    fi
}

# Cleanup function
cleanup() {
    print_info "Cleaning up deployment..."
    
    kubectl delete -f k8s/overlays/$ENVIRONMENT/ -n $NAMESPACE --ignore-not-found=true || \
    kubectl delete -f k8s/base/ -n $NAMESPACE --ignore-not-found=true
    
    print_status "Cleanup completed"
}

# Main deployment function
main() {
    case "${1:-deploy}" in
        "deploy")
            check_prerequisites
            build_image
            create_namespace
            apply_secrets
            apply_config
            wait_for_rollout
            verify_deployment
            show_status
            ;;
        "status")
            SHOW_LOGS=${SHOW_LOGS:-false}
            show_status
            ;;
        "logs")
            SHOW_LOGS=true
            show_status
            ;;
        "cleanup")
            cleanup
            ;;
        "build")
            build_image
            ;;
        "secrets")
            apply_secrets
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [deploy|status|logs|cleanup|build|secrets|help] [environment] [namespace] [registry] [tag]"
            echo ""
            echo "Commands:"
            echo "  deploy   - Full deployment (default)"
            echo "  status   - Show deployment status"
            echo "  logs     - Show recent logs"
            echo "  cleanup  - Remove deployment"
            echo "  build    - Build and push Docker image"
            echo "  secrets  - Apply secrets only"
            echo "  help     - Show this help"
            echo ""
            echo "Environments:"
            echo "  production (default)"
            echo "  staging"
            echo "  development"
            echo ""
            echo "Examples:"
            echo "  $0 deploy production obsidian-system gcr.io/obsidian-bot-prod v1.0.0"
            echo "  $0 status"
            echo "  $0 logs"
            ;;
        *)
            print_error "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Parse command line arguments
if [ $# -gt 0 ]; then
    main "$@"
else
    main "deploy" "$ENVIRONMENT" "$NAMESPACE" "$REGISTRY" "$IMAGE_TAG"
fi