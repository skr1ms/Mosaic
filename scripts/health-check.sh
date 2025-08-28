#!/bin/bash
set -e

# Script for checking service health after deployment
# Used in CI/CD pipeline

# Function for logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Function to check URL
check_url() {
    local url="$1"
    local timeout="${2:-10}"
    local max_attempts="${3:-5}"
    local attempt=1
    
    log "🔍 Checking $url (timeout: ${timeout}s, max attempts: $max_attempts)"
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f --max-time "$timeout" "$url" > /dev/null 2>&1; then
            log "✅ $url is healthy (attempt $attempt)"
            return 0
        else
            log "⚠️  $url check failed (attempt $attempt/$max_attempts)"
            if [ $attempt -lt $max_attempts ]; then
                sleep 2
            fi
        fi
        attempt=$((attempt + 1))
    done
    
    log "❌ $url health check failed after $max_attempts attempts"
    return 1
}

# Function to check API endpoint with JSON response
check_api_endpoint() {
    local url="$1"
    local expected_field="${2:-status}"
    local expected_value="${3:-ok}"
    local timeout="${4:-10}"
    local max_attempts="${5:-5}"
    local attempt=1
    
    log "🔍 Checking API endpoint $url (expecting $expected_field: $expected_value)"
    
    while [ $attempt -le $max_attempts ]; do
        local response
        response=$(curl -s --max-time "$timeout" "$url" 2>/dev/null || echo "")
        
        if [ -n "$response" ]; then
            # Check JSON response
            local field_value
            field_value=$(echo "$response" | jq -r ".$expected_field" 2>/dev/null || echo "")
            
            if [ "$field_value" = "$expected_value" ]; then
                log "✅ API $url is healthy (attempt $attempt)"
                return 0
            else
                log "⚠️  API $url returned unexpected value: $field_value (attempt $attempt/$max_attempts)"
            fi
        else
            log "⚠️  API $url no response (attempt $attempt/$max_attempts)"
        fi
        
        if [ $attempt -lt $max_attempts ]; then
            sleep 2
        fi
        attempt=$((attempt + 1))
    done
    
    log "❌ API $url health check failed after $max_attempts attempts"
    return 1
}

# Function to check Docker container
check_docker_container() {
    local container_name="$1"
    local timeout="${2:-30}"
    
    log "🐳 Checking Docker container: $container_name"
    
    # Check that container is running
    if ! docker ps --filter "name=$container_name" --filter "status=running" | grep -q "$container_name"; then
        log "❌ Container $container_name is not running"
        return 1
    fi
    
    # Check health status if defined
    local health_status
    health_status=$(docker inspect --format='{{.State.Health.Status}}' "$container_name" 2>/dev/null || echo "none")
    
    if [ "$health_status" = "healthy" ]; then
        log "✅ Container $container_name is healthy"
        return 0
    elif [ "$health_status" = "none" ]; then
        log "✅ Container $container_name is running (no health check defined)"
        return 0
    elif [ "$health_status" = "starting" ]; then
        log "⏳ Container $container_name is starting, waiting..."
        local attempt=1
        local max_attempts=$((timeout / 2))
        
        while [ $attempt -le $max_attempts ]; do
            health_status=$(docker inspect --format='{{.State.Health.Status}}' "$container_name" 2>/dev/null || echo "none")
            
            if [ "$health_status" = "healthy" ]; then
                log "✅ Container $container_name is now healthy (attempt $attempt)"
                return 0
            elif [ "$health_status" = "unhealthy" ]; then
                log "❌ Container $container_name became unhealthy"
                return 1
            fi
            
            sleep 2
            attempt=$((attempt + 1))
        done
        
        log "❌ Container $container_name health check timeout"
        return 1
    else
        log "❌ Container $container_name is unhealthy: $health_status"
        return 1
    fi
}

# Function to check database
check_database() {
    local max_attempts="${1:-10}"
    local attempt=1
    
    log "🗄️  Checking PostgreSQL database connection"
    
    while [ $attempt -le $max_attempts ]; do
        if docker exec postgres pg_isready -U postgres > /dev/null 2>&1; then
            log "✅ Database is ready (attempt $attempt)"
            return 0
        else
            log "⏳ Database not ready (attempt $attempt/$max_attempts)"
            if [ $attempt -lt $max_attempts ]; then
                sleep 2
            fi
        fi
        attempt=$((attempt + 1))
    done
    
    log "❌ Database connection failed after $max_attempts attempts"
    return 1
}

# Function to check Redis
check_redis() {
    local max_attempts="${1:-10}"
    local attempt=1
    
    log "🔴 Checking Redis connection"
    
    while [ $attempt -le $max_attempts ]; do
        if docker exec redis redis-cli ping | grep -q "PONG"; then
            log "✅ Redis is ready (attempt $attempt)"
            return 0
        else
            log "⏳ Redis not ready (attempt $attempt/$max_attempts)"
            if [ $attempt -lt $max_attempts ]; then
                sleep 2
            fi
        fi
        attempt=$((attempt + 1))
    done
    
    log "❌ Redis connection failed after $max_attempts attempts"
    return 1
}

# Function for complete system check
full_health_check() {
    local failed_checks=0
    
    log "🏥 Starting comprehensive health check..."
    echo "=========================================="
    
    # Check Docker containers
    log "📋 Checking Docker containers..."
    
    local containers=("postgres" "redis" "backend" "public-site" "dashboards" "nginx")
    for container in "${containers[@]}"; do
        if ! check_docker_container "$container"; then
            failed_checks=$((failed_checks + 1))
        fi
    done
    
    # Check database services
    echo "----------------------------------------"
    log "📋 Checking database services..."
    
    if ! check_database; then
        failed_checks=$((failed_checks + 1))
    fi
    
    if ! check_redis; then
        failed_checks=$((failed_checks + 1))
    fi
    
    # Check main URLs
    echo "----------------------------------------"
    log "📋 Checking main URLs..."
    
    local main_urls=(
        "https://photo.doyoupaint.com/health"
        "https://adm.doyoupaint.com/health" 
        "https://photo.doyoupaint.com/api/health"
        "https://adm.doyoupaint.com/api/health"
    )
    
    for url in "${main_urls[@]}"; do
        if ! check_url "$url" 15 3; then
            failed_checks=$((failed_checks + 1))
        fi
    done
    
    # Check API endpoints
    echo "----------------------------------------"
    log "📋 Checking API endpoints..."
    
    if ! check_api_endpoint "http://localhost:8080/api/health" "status" "healthy"; then
        failed_checks=$((failed_checks + 1))
    fi
    
    # Check partner domains (if any)
    echo "----------------------------------------"
    log "📋 Checking partner domains..."
    
    # Get partner domains list from database
    local partner_domains
    partner_domains=$(docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" postgres psql -U postgres -d mosaic -t -c "SELECT domain FROM partners WHERE status = 'active' AND domain IS NOT NULL AND domain != '';" 2>/dev/null | tr -d ' ' | grep -v '^$' || echo "")
    
    if [ -n "$partner_domains" ]; then
        echo "$partner_domains" | while IFS= read -r domain; do
            if [ -n "$domain" ]; then
                local partner_url="https://$domain/health"
                if ! check_url "$partner_url" 15 3; then
                    failed_checks=$((failed_checks + 1))
                fi
            fi
        done
    else
        log "ℹ️  No partner domains found to check"
    fi
    
    echo "=========================================="
    
    if [ $failed_checks -eq 0 ]; then
        log "✅ All health checks passed successfully!"
        return 0
    else
        log "❌ $failed_checks health check(s) failed!"
        return 1
    fi
}

# Function for quick check
quick_health_check() {
    local failed_checks=0
    
    log "⚡ Starting quick health check..."
    
    # Check main URLs
    local main_urls=(
        "https://photo.doyoupaint.com/health"
        "https://adm.doyoupaint.com/health"
    )
    
    for url in "${main_urls[@]}"; do
        if ! check_url "$url" 10 2; then
            failed_checks=$((failed_checks + 1))
        fi
    done
    
    if [ $failed_checks -eq 0 ]; then
        log "✅ Quick health check passed!"
        return 0
    else
        log "❌ Quick health check failed!"
        return 1
    fi
}

# Main function
main() {
    local check_type="${1:-quick}"
    
    echo "=========================================="
    echo "🏥 Health Check Script"
    echo "=========================================="
    
    case "$check_type" in
        "full")
            full_health_check
            ;;
        "quick")
            quick_health_check
            ;;
        "url")
            if [ -z "$2" ]; then
                log "❌ URL parameter required for url check"
                echo "Usage: $0 url <URL> [timeout] [attempts]"
                exit 1
            fi
            check_url "$2" "${3:-10}" "${4:-3}"
            ;;
        "api")
            if [ -z "$2" ]; then
                log "❌ API URL parameter required for api check"
                echo "Usage: $0 api <API_URL> [field] [expected_value] [timeout] [attempts]"
                exit 1
            fi
            check_api_endpoint "$2" "${3:-status}" "${4:-ok}" "${5:-10}" "${6:-3}"
            ;;
        "container")
            if [ -z "$2" ]; then
                log "❌ Container name required for container check"
                echo "Usage: $0 container <CONTAINER_NAME> [timeout]"
                exit 1
            fi
            check_docker_container "$2" "${3:-30}"
            ;;
        *)
            log "❌ Unknown check type: $check_type"
            echo "Usage: $0 [full|quick|url|api|container]"
            echo ""
            echo "Examples:"
            echo "  $0 full                                          # Full health check"
            echo "  $0 quick                                         # Quick health check"
            echo "  $0 url https://example.com/health                # Check specific URL"
            echo "  $0 api http://localhost:8080/api/health          # Check API endpoint"
            echo "  $0 container backend                             # Check Docker container"
            exit 1
            ;;
    esac
}

# Run main function with parameters
main "$@"
