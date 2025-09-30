#!/bin/bash
set -e

# Enhanced script for generating nginx configuration with partner domains
# This script uses templates and generates optimized nginx configurations

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Configuration paths
TEMPLATES_DIR="$PROJECT_ROOT/deployments/nginx/templates"
SNIPPETS_DIR="$PROJECT_ROOT/deployments/nginx/snippets"
OUTPUT_DIR="/etc/nginx"
PARTNERS_CONF_DIR="$OUTPUT_DIR/conf.d/partners"

# SSL Configuration
SSL_CERT_NAME="photo.doyoupaint.com-0001"
DEFAULT_SSL_CERT="/etc/letsencrypt/live/$SSL_CERT_NAME/fullchain.pem"
DEFAULT_SSL_KEY="/etc/letsencrypt/live/$SSL_CERT_NAME/privkey.pem"

# Function for logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

error() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ❌ ERROR: $1" >&2
    exit 1
}

success() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ✅ $1"
}

echo "=========================================="
echo "🔧 Enhanced Nginx Configuration Generator"
echo "=========================================="

# Function to check if running in Docker or on host
check_environment() {
    log "Checking execution environment..."
    
    if [ -f "/.dockerenv" ]; then
        log "Running inside Docker container"
        DOCKER_MODE=true
    else
        log "Running on host system"
        DOCKER_MODE=false
    fi
}

# Function to get active partner domains from database
get_active_domains() {
    # All logging to stderr to avoid contaminating domain output
    {
        log "📋 Getting active partner domains from database..."
        
        # Load environment variables if available
        if [ -f "$PROJECT_ROOT/.env" ]; then
            export $(cat "$PROJECT_ROOT/.env" | grep -v '^#' | xargs)
        fi
        
        local query="SELECT domain FROM partners WHERE status = 'active' AND domain IS NOT NULL AND domain != '' AND domain ~ '^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' ORDER BY domain;"
        
        local domains=""
        
        # Try different methods to connect to database
        if command -v psql >/dev/null 2>&1 && [ -n "$POSTGRES_PASSWORD" ]; then
            log "Using direct psql connection..."
            domains=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h localhost -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-mosaic}" -t -c "$query" 2>/dev/null | sed 's/[[:space:]]//g' | grep -v '^$' || echo "")
        elif docker ps --format "table {{.Names}}" | grep -q "postgres" 2>/dev/null; then
            log "Using Docker postgres container..."
            domains=$(docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" postgres psql -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-mosaic}" -t -c "$query" 2>/dev/null | sed 's/[[:space:]]//g' | grep -v '^$' || echo "")
        else
            log "⚠️  Cannot connect to database, no partner domains will be configured"
            # Still return empty string to stdout
            echo ""
            return 0
        fi
        
        if [ -n "$domains" ]; then
            log "✅ Found $(echo "$domains" | wc -l) active partner domains"
        else
            log "ℹ️  No active partner domains found"
        fi
    } >&2
    
    # Only return domains to stdout
    echo "$domains"
}

# Function to validate domain format
validate_domain() {
    local domain="$1"
    
    if [[ ! "$domain" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
        log "⚠️  Invalid domain format: $domain"
        return 1
    fi
    
    return 0
}

# Function to check SSL certificate availability
check_ssl_certificate() {
    local domain="$1"
    local cert_path="/etc/letsencrypt/live/$SSL_CERT_NAME/fullchain.pem"
    local key_path="/etc/letsencrypt/live/$SSL_CERT_NAME/privkey.pem"
    
    if [ -f "$cert_path" ] && [ -f "$key_path" ]; then
        # Check if domain is included in certificate
        if openssl x509 -in "$cert_path" -noout -text 2>/dev/null | grep -q "$domain"; then
            log "✅ SSL certificate includes domain: $domain"
            return 0
        else
            log "⚠️  Domain $domain not found in SSL certificate"
            return 1
        fi
    else
        log "⚠️  SSL certificate files not found"
        return 1
    fi
}

# Function to generate partner domain configuration
generate_partner_config() {
    local domain="$1"
    local output_file="$PARTNERS_CONF_DIR/${domain}.conf"
    
    log "🔧 Generating configuration for domain: $domain"
    
    if ! validate_domain "$domain"; then
        log "⚠️  Skipping invalid domain: $domain"
        return 1
    fi
    
    # Check SSL certificate
    if ! check_ssl_certificate "$domain"; then
        log "⚠️  SSL certificate issue for domain $domain, using default certificate"
    fi
    
    # Create partner config directory if it doesn't exist
    mkdir -p "$PARTNERS_CONF_DIR"
    
    # Generate configuration from template
    if [ -f "$TEMPLATES_DIR/partner-domain.conf.template" ]; then
        cp "$TEMPLATES_DIR/partner-domain.conf.template" "$output_file"
        
        # Replace template variables
        sed -i "s|{{DOMAIN}}|$domain|g" "$output_file"
        sed -i "s|{{SSL_CERT_PATH}}|$DEFAULT_SSL_CERT|g" "$output_file"
        sed -i "s|{{SSL_KEY_PATH}}|$DEFAULT_SSL_KEY|g" "$output_file"
        
        success "Generated configuration for $domain"
        return 0
    else
        error "Partner domain template not found: $TEMPLATES_DIR/partner-domain.conf.template"
    fi
}

# Function to generate main nginx configuration
generate_main_config() {
    log "🔧 Generating main nginx configuration..."
    
    local main_config="$OUTPUT_DIR/nginx.conf"
    local backup_config="$main_config.backup.$(date +%Y%m%d_%H%M%S)"
    
    # Create backup of current configuration
    if [ -f "$main_config" ]; then
        cp "$main_config" "$backup_config"
        log "Created backup: $backup_config"
    fi
    
    # Generate main configuration from template
    if [ -f "$TEMPLATES_DIR/main.conf.template" ]; then
        cp "$TEMPLATES_DIR/main.conf.template" "$main_config"
        
        # Replace template variables
        sed -i "s|{{DEFAULT_SSL_CERT}}|$DEFAULT_SSL_CERT|g" "$main_config"
        sed -i "s|{{DEFAULT_SSL_KEY}}|$DEFAULT_SSL_KEY|g" "$main_config"
        
        success "Generated main nginx configuration"
    else
        error "Main configuration template not found: $TEMPLATES_DIR/main.conf.template"
    fi
}

# Function to copy nginx snippets
copy_snippets() {
    log "📋 Copying nginx snippets..."
    
    local target_snippets_dir="$OUTPUT_DIR/snippets"
    mkdir -p "$target_snippets_dir"
    
    if [ -d "$SNIPPETS_DIR" ]; then
        cp -r "$SNIPPETS_DIR"/* "$target_snippets_dir/"
        success "Copied nginx snippets"
    else
        log "⚠️  Snippets directory not found: $SNIPPETS_DIR"
    fi
}

# Function to clean up old partner configurations
cleanup_old_configs() {
    log "🧹 Cleaning up old partner configurations..."
    
    if [ -d "$PARTNERS_CONF_DIR" ]; then
        # Get current active domains
        local active_domains=$(get_active_domains)
        
        # Find all existing config files
        for config_file in "$PARTNERS_CONF_DIR"/*.conf; do
            if [ -f "$config_file" ]; then
                local domain=$(basename "$config_file" .conf)
                local domain_active=false
                
                # Check if domain is still active
                while IFS= read -r active_domain; do
                    if [ -n "$active_domain" ] && [ "$domain" = "$active_domain" ]; then
                        domain_active=true
                        break
                    fi
                done <<< "$active_domains"
                
                if [ "$domain_active" = false ]; then
                    log "🗑️  Removing configuration for inactive domain: $domain"
                    rm -f "$config_file"
                fi
            fi
        done
    fi
}

# Function to validate generated nginx configuration
validate_nginx_config() {
    log "🔍 Validating nginx configuration..."
    
    if command -v nginx >/dev/null 2>&1; then
        if nginx -t 2>/dev/null; then
            success "Nginx configuration is valid"
            return 0
        else
            log "❌ Nginx configuration validation failed"
            nginx -t 2>&1 | while IFS= read -r line; do
                log "  $line"
            done
            return 1
        fi
    elif docker ps --format "table {{.Names}}" | grep -q "nginx" 2>/dev/null; then
        if docker exec nginx nginx -t 2>/dev/null; then
            success "Nginx configuration is valid (Docker)"
            return 0
        else
            log "❌ Nginx configuration validation failed (Docker)"
            docker exec nginx nginx -t 2>&1 | while IFS= read -r line; do
                log "  $line"
            done
            return 1
        fi
    else
        log "⚠️  Cannot validate nginx configuration - nginx not available"
        return 0
    fi
}

# Function to reload nginx
reload_nginx() {
    log "🔄 Reloading nginx..."
    
    if systemctl is-active --quiet nginx 2>/dev/null; then
        if systemctl reload nginx; then
            success "Nginx reloaded successfully (systemctl)"
            return 0
        else
            log "❌ Failed to reload nginx via systemctl"
            return 1
        fi
    elif docker ps --format "table {{.Names}}" | grep -q "nginx" 2>/dev/null; then
        if docker exec nginx nginx -s reload 2>/dev/null; then
            success "Nginx reloaded successfully (Docker)"
            return 0
        else
            log "❌ Failed to reload nginx via Docker"
            return 1
        fi
    else
        log "⚠️  Cannot reload nginx - not running or not accessible"
        return 1
    fi
}

# Function to generate configuration via backend API
generate_via_api() {
    log "🌐 Attempting to generate configuration via backend API..."
    
    local backend_url="${BACKEND_URL:-http://localhost:8080}"
    local admin_email="${DEFAULT_ADMIN_EMAIL:-admin@example.com}"
    local admin_password="${DEFAULT_ADMIN_PASSWORD}"
    
    if [ -z "$admin_password" ]; then
        log "⚠️  DEFAULT_ADMIN_PASSWORD not set, skipping API method"
        return 1
    fi
    
    # Get admin JWT token
    log "Authenticating with backend API..."
    local login_response
    login_response=$(timeout 30s curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"${admin_email}\",\"password\":\"${admin_password}\"}" \
        "${backend_url}/api/auth/login" 2>/dev/null || echo "")
    
    if [ -z "$login_response" ]; then
        log "⚠️  Failed to get login response from API"
        return 1
    fi
    
    local admin_token
    admin_token=$(echo "$login_response" | jq -r '.access_token' 2>/dev/null || echo "")
    
    if [ -z "$admin_token" ] || [ "$admin_token" = "null" ]; then
        log "⚠️  Failed to extract admin token from API response"
        return 1
    fi
    
    success "Authenticated with backend API"
    
    # Call backend API to deploy nginx config
    log "Calling nginx deployment API..."
    local nginx_response
    nginx_response=$(timeout 60s curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $admin_token" \
        "${backend_url}/api/admin/nginx/deploy" 2>/dev/null || echo "")
    
    if [ -n "$nginx_response" ]; then
        success "Configuration generated via backend API"
        log "Response: $nginx_response"
        return 0
    else
        log "⚠️  Failed to generate configuration via backend API"
        return 1
    fi
}

# Main execution function
main() {
    local method="${1:-template}"
    
    log "Starting nginx configuration generation (method: $method)..."
    
    # Check environment
    check_environment
    
    case "$method" in
        "api")
            log "Using backend API method..."
            if generate_via_api; then
                success "Configuration generated successfully via API"
            else
                log "⚠️  API method failed, falling back to template method"
                method="template"
            fi
            ;;
        "template")
            log "Using template method..."
            ;;
        *)
            error "Unknown method: $method. Use 'api' or 'template'"
            ;;
    esac
    
    if [ "$method" = "template" ]; then
        # Copy nginx snippets
        copy_snippets
        
        # Generate main configuration
        generate_main_config
        
        # Get active domains and generate partner configurations
        local domains
        domains=$(get_active_domains)
        
        # Clean up old configurations first
        cleanup_old_configs
        
        # Generate new partner configurations
        if [ -n "$domains" ]; then
            log "Generating partner domain configurations..."
            while IFS= read -r domain; do
                if [ -n "$domain" ]; then
                    generate_partner_config "$domain"
                fi
            done <<< "$domains"
        else
            log "No partner domains to configure"
        fi
    fi
    
    # Validate configuration
    if ! validate_nginx_config; then
        error "Generated configuration is invalid"
    fi
    
    # Reload nginx if requested
    if [ "${RELOAD_NGINX:-true}" = "true" ]; then
        if ! reload_nginx; then
            log "⚠️  Configuration generated but nginx reload failed"
        fi
    fi
    
    success "Nginx configuration generation completed successfully!"
    
    # Display summary
    echo "=========================================="
    log "📊 Configuration Summary:"
    log "Main config: $OUTPUT_DIR/nginx.conf"
    log "Snippets: $OUTPUT_DIR/snippets/"
    
    if [ -d "$PARTNERS_CONF_DIR" ]; then
        local partner_count=$(find "$PARTNERS_CONF_DIR" -name "*.conf" 2>/dev/null | wc -l)
        log "Partner domains: $partner_count configured"
        
        if [ "$partner_count" -gt 0 ]; then
            log "Partner configurations:"
            find "$PARTNERS_CONF_DIR" -name "*.conf" 2>/dev/null | while read -r conf_file; do
                local domain=$(basename "$conf_file" .conf)
                log "  - $domain"
            done
        fi
    fi
    echo "=========================================="
}

# Check command line parameters
case "${1:-template}" in
    "api")
        main api
        ;;
    "template")
        main template
        ;;
    "validate")
        validate_nginx_config
        ;;
    "reload")
        reload_nginx
        ;;
    *)
        echo "Usage: $0 [api|template|validate|reload]"
        echo ""
        echo "Methods:"
        echo "  api      - Generate config via backend API"
        echo "  template - Generate config using templates (default)"
        echo "  validate - Validate existing nginx configuration"
        echo "  reload   - Reload nginx configuration"
        echo ""
        echo "Environment variables:"
        echo "  BACKEND_URL              - Backend API URL (default: http://localhost:8080)"
        echo "  DEFAULT_ADMIN_EMAIL      - Admin email for API authentication"
        echo "  DEFAULT_ADMIN_PASSWORD   - Admin password for API authentication"
        echo "  POSTGRES_PASSWORD        - Database password"
        echo "  POSTGRES_USER            - Database user (default: postgres)"
        echo "  POSTGRES_DB              - Database name (default: mosaic)"
        echo "  RELOAD_NGINX             - Whether to reload nginx after generation (default: true)"
        exit 1
        ;;
esac
