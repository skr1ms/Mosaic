#!/bin/bash
set -e

# Script for generating nginx configuration via backend API
# This script calls the admin API to generate updated nginx configuration

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Function for logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

echo "=========================================="
echo "🔧 Nginx Configuration Generator"
echo "=========================================="

# Function to generate nginx config via API
generate_nginx_via_api() {
    local backend_url="${BACKEND_URL:-http://localhost:8080}"
    local admin_email="${DEFAULT_ADMIN_EMAIL:-admin@example.com}"
    local admin_password="${DEFAULT_ADMIN_PASSWORD}"
    
    log "Generating nginx configuration via backend API..."
    
    if [ -z "$admin_password" ]; then
        log "❌ DEFAULT_ADMIN_PASSWORD environment variable is required"
        return 1
    fi
    
    # Get admin JWT token
    log "Getting admin JWT token..."
    local login_response
    login_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"${admin_email}\",\"password\":\"${admin_password}\"}" \
        "${backend_url}/api/auth/login" 2>/dev/null || echo "")
    
    if [ -z "$login_response" ]; then
        log "❌ Failed to get login response"
        return 1
    fi
    
    local admin_token
    admin_token=$(echo "$login_response" | jq -r '.access_token' 2>/dev/null || echo "")
    
    if [ -z "$admin_token" ] || [ "$admin_token" = "null" ]; then
        log "❌ Failed to extract admin token"
        return 1
    fi
    
    log "✅ Admin token obtained"
    
    # Call backend API to deploy nginx config
    log "Calling nginx deploy API..."
    local nginx_response
    nginx_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $admin_token" \
        "${backend_url}/api/admin/nginx/deploy" 2>/dev/null || echo "")
    
    if [ -n "$nginx_response" ]; then
        log "✅ Nginx configuration updated via API: $nginx_response"
        return 0
    else
        log "❌ Failed to update nginx config via API"
        return 1
    fi
}

# Function to fallback to manual configuration copy
fallback_to_manual_copy() {
    log "⚠️ Falling back to manual configuration copy..."
    
    local source_config="$PROJECT_ROOT/deployments/nginx/production.conf"
    local target_config="/etc/nginx/nginx.conf"
    
    if [ -f "$source_config" ]; then
        log "Copying configuration from $source_config to $target_config..."
        
        # Backup current configuration
        if [ -f "$target_config" ]; then
            cp "$target_config" "$target_config.backup.$(date +%Y%m%d_%H%M%S)"
        fi
        
        # Copy new configuration
        cp "$source_config" "$target_config"
        
        log "✅ Configuration copied successfully"
        return 0
    else
        log "❌ Source configuration file not found: $source_config"
        return 1
    fi
}

# Function to validate nginx configuration
validate_nginx_config() {
    log "🔍 Validating nginx configuration..."
    
    if nginx -t 2>/dev/null; then
        log "✅ Nginx configuration is valid"
        return 0
    elif docker exec nginx nginx -t 2>/dev/null; then
        log "✅ Nginx configuration is valid (in container)"
        return 0
    else
        log "❌ Nginx configuration validation failed"
        return 1
    fi
}

# Main function
main() {
    local use_api="${1:-true}"
    
    log "Starting nginx configuration generation..."
    
    # Load environment variables
    if [ -f "$PROJECT_ROOT/.env" ]; then
        export $(cat "$PROJECT_ROOT/.env" | grep -v '^#' | xargs)
    fi
    
    if [ "$use_api" = "true" ]; then
        # Try API first
        if generate_nginx_via_api; then
            log "✅ Nginx configuration generated successfully via API"
        else
            log "⚠️ API method failed, trying manual copy..."
            if ! fallback_to_manual_copy; then
                log "❌ Manual copy also failed"
                exit 1
            fi
        fi
    else
        # Use manual copy
        if ! fallback_to_manual_copy; then
            log "❌ Manual configuration copy failed"
            exit 1
        fi
    fi
    
    # Validate the configuration
    if ! validate_nginx_config; then
        log "❌ Generated configuration is invalid"
        exit 1
    fi
    
    log "✅ Nginx configuration generation completed successfully!"
}

# Check command line parameters
case "${1:-api}" in
    "api")
        main true
        ;;
    "manual"|"copy")
        main false
        ;;
    "validate")
        validate_nginx_config
        ;;
    *)
        log "❌ Unknown mode: $1"
        echo "Usage: $0 [api|manual|validate]"
        echo ""
        echo "Modes:"
        echo "  api      - Generate config via backend API (default)"
        echo "  manual   - Copy config from deployments/nginx/production.conf"
        echo "  validate - Validate existing nginx configuration"
        exit 1
        ;;
esac
