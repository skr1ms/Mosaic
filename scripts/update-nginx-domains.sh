#!/bin/bash
set -e

# Script for updating nginx configuration with partner domains
# Used in CI/CD pipeline during deployment

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
NGINX_CONFIG_PATH="$PROJECT_ROOT/deployments/nginx/production.conf"
BACKUP_DIR="$PROJECT_ROOT/deployments/nginx/backups"

# Create backup directory
mkdir -p "$BACKUP_DIR"

echo "=========================================="
echo "🔧 Starting nginx domains configuration update"
echo "=========================================="

# Function for logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Function to create backup
create_backup() {
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_file="$BACKUP_DIR/production.conf.backup.$timestamp"
    
    if [ -f "$NGINX_CONFIG_PATH" ]; then
        cp "$NGINX_CONFIG_PATH" "$backup_file"
        log "✅ Backup created: $backup_file"
    else
        log "⚠️  No existing nginx config found, skipping backup"
    fi
}

# Function to call backend API for generating new configuration
generate_config_from_api() {
    log "🔄 Calling backend API to generate nginx config with partner domains..."
    
    # Get admin JWT token (assume it's in environment variables)
    if [ -z "$ADMIN_JWT_TOKEN" ]; then
        log "❌ ADMIN_JWT_TOKEN not set. Attempting to get token..."
        
        # Try to get token via API login
        local login_response
        login_response=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -d "{\"email\":\"${DEFAULT_ADMIN_EMAIL:-admin@example.com}\",\"password\":\"${DEFAULT_ADMIN_PASSWORD}\"}" \
            "http://localhost:8080/api/auth/login" || echo "")
        
        if [ -z "$login_response" ]; then
            log "❌ Failed to get admin token"
            return 1
        fi
        
        ADMIN_JWT_TOKEN=$(echo "$login_response" | jq -r '.access_token' 2>/dev/null || echo "")
        
        if [ -z "$ADMIN_JWT_TOKEN" ] || [ "$ADMIN_JWT_TOKEN" = "null" ]; then
            log "❌ Failed to extract token from login response"
            return 1
        fi
        
        log "✅ Successfully obtained admin token"
    fi
    
    # Call API to generate and deploy nginx configuration
    local api_response
    api_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_JWT_TOKEN" \
        "http://localhost:8080/api/admin/nginx/deploy" || echo "")
    
    if [ -z "$api_response" ]; then
        log "❌ Failed to call nginx deploy API"
        return 1
    fi
    
    # Check response success
    local success=$(echo "$api_response" | jq -r '.message' 2>/dev/null || echo "")
    if [[ "$success" == *"successfully"* ]]; then
        log "✅ Nginx configuration generated successfully via API"
        return 0
    else
        log "❌ API returned error: $api_response"
        return 1
    fi
}

# Function to validate nginx configuration
validate_nginx_config() {
    log "🔍 Validating nginx configuration..."
    
    if ! [ -f "$NGINX_CONFIG_PATH" ]; then
        log "❌ Nginx config file not found: $NGINX_CONFIG_PATH"
        return 1
    fi
    
    # Test syntax via nginx container
    if docker exec nginx nginx -t 2>/dev/null; then
        log "✅ Nginx configuration is valid"
        return 0
    else
        log "❌ Nginx configuration validation failed"
        return 1
    fi
}

# Function to apply configuration
apply_nginx_config() {
    log "🔄 Applying nginx configuration..."
    
    if docker exec nginx nginx -s reload; then
        log "✅ Nginx configuration reloaded successfully"
        return 0
    else
        log "❌ Failed to reload nginx configuration"
        return 1
    fi
}

# Function to restore from backup
restore_from_backup() {
    log "🔄 Restoring from backup..."
    
    local latest_backup=$(ls -t "$BACKUP_DIR"/production.conf.backup.* 2>/dev/null | head -1)
    
    if [ -n "$latest_backup" ] && [ -f "$latest_backup" ]; then
        cp "$latest_backup" "$NGINX_CONFIG_PATH"
        log "✅ Restored from backup: $latest_backup"
        
        if validate_nginx_config && apply_nginx_config; then
            log "✅ Backup configuration applied successfully"
            return 0
        else
            log "❌ Failed to apply backup configuration"
            return 1
        fi
    else
        log "❌ No backup found to restore from"
        return 1
    fi
}

# Function to check health after update
health_check() {
    log "🏥 Performing health check..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "http://localhost/health" > /dev/null 2>&1; then
            log "✅ Health check passed (attempt $attempt)"
            return 0
        fi
        
        log "⏳ Health check failed (attempt $attempt/$max_attempts), waiting 2 seconds..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log "❌ Health check failed after $max_attempts attempts"
    return 1
}

# Function to clean up old backups
cleanup_old_backups() {
    log "🧹 Cleaning up old backups (keeping last 10)..."
    
    local backups_count=$(ls -1 "$BACKUP_DIR"/production.conf.backup.* 2>/dev/null | wc -l)
    
    if [ "$backups_count" -gt 10 ]; then
        ls -t "$BACKUP_DIR"/production.conf.backup.* | tail -n +11 | xargs rm -f
        log "✅ Cleaned up $((backups_count - 10)) old backups"
    else
        log "✅ No old backups to clean up ($backups_count total)"
    fi
}

# Main function
main() {
    log "Starting nginx domains update process..."
    
    # Check required environment variables
    if [ -z "$DEFAULT_ADMIN_PASSWORD" ]; then
        log "❌ DEFAULT_ADMIN_PASSWORD environment variable is required"
        exit 1
    fi
    
    # Create backup of current configuration
    create_backup
    
    # Generate new configuration via API
    if ! generate_config_from_api; then
        log "❌ Failed to generate nginx configuration"
        exit 1
    fi
    
    # Validate new configuration
    if ! validate_nginx_config; then
        log "❌ New configuration is invalid, restoring from backup..."
        if ! restore_from_backup; then
            log "❌ Failed to restore from backup"
            exit 1
        fi
        exit 1
    fi
    
    # Apply new configuration
    if ! apply_nginx_config; then
        log "❌ Failed to apply new configuration, restoring from backup..."
        if ! restore_from_backup; then
            log "❌ Failed to restore from backup"
            exit 1
        fi
        exit 1
    fi
    
    # Check health after update
    if ! health_check; then
        log "❌ Health check failed after configuration update, restoring from backup..."
        if ! restore_from_backup; then
            log "❌ Failed to restore from backup"
            exit 1
        fi
        exit 1
    fi
    
    # Clean up old backups
    cleanup_old_backups
    
    log "✅ Nginx domains configuration updated successfully!"
}

# Run main function
main "$@"
