#!/bin/bash
set -e

# Enhanced script for managing partner domains
# Handles add/update/delete operations with proper cleanup

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
NGINX_CONFIG_PATH="$PROJECT_ROOT/deployments/nginx/production.conf"
SSL_CERT_PATH="/etc/letsencrypt/live"
NGINX_SITES_PATH="$PROJECT_ROOT/deployments/nginx/sites"

# Operation mode: add, update, delete
OPERATION="${1:-add}"
OLD_DOMAIN="${2:-}"
NEW_DOMAIN="${3:-}"

# Function for logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

error() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $1" >&2
    exit 1
}

# Load environment variables
if [ -f "$PROJECT_ROOT/.env" ]; then
    export $(cat "$PROJECT_ROOT/.env" | grep -v '^#' | xargs)
fi

echo "=========================================="
echo "🔧 Partner Domain Management Script"
echo "Operation: $OPERATION"
echo "=========================================="

# Function to get all active partner domains from database
get_active_domains() {
    # All logging to stderr to avoid contaminating domain output
    {
        log "📋 Getting active partner domains from database..."
        
        # First, let's check what partners exist at all
        local count_query="SELECT COUNT(*) FROM partners;"
        local active_count_query="SELECT COUNT(*) FROM partners WHERE status = 'active';"
        local with_domains_query="SELECT COUNT(*) FROM partners WHERE domain IS NOT NULL AND domain != '';"
        
        # Main query for active domains
        local query="SELECT domain FROM partners WHERE status = 'active' AND domain IS NOT NULL AND domain != '' ORDER BY domain;"
        
        # Execute diagnostic queries first
        if command -v psql &> /dev/null; then
            log "Using psql client to query database..."
            
            log "Running diagnostic queries..."
            local total_partners=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h localhost -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "$count_query" 2>/dev/null | xargs || echo "0")
            local active_partners=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h localhost -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "$active_count_query" 2>/dev/null | xargs || echo "0")
            local with_domains=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h localhost -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "$with_domains_query" 2>/dev/null | xargs || echo "0")
            
            log "Total partners: $total_partners, Active: $active_partners, With domains: $with_domains"
            
            domains=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h localhost -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "$query" 2>/dev/null | xargs | tr ' ' '\n' | grep -v '^$' || echo "")
            
        elif docker ps | grep -q postgres; then
            log "Using docker exec to query postgres container..."
            
            log "Running diagnostic queries..."
            local total_partners=$(docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "$count_query" 2>/dev/null | xargs || echo "0")
            local active_partners=$(docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "$active_count_query" 2>/dev/null | xargs || echo "0")
            local with_domains=$(docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "$with_domains_query" 2>/dev/null | xargs || echo "0")
            
            log "Total partners: $total_partners, Active: $active_partners, With domains: $with_domains"
            
            domains=$(docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "$query" 2>/dev/null | xargs | tr ' ' '\n' | grep -v '^$' || echo "")
            
        else
            log "ERROR: Cannot connect to database - no psql client or postgres container found"
            echo "" # Return empty string on stdout
            return 1
        fi
        
        log "Database query completed, found domains: [$domains]"
    } >&2
    
    # Only return domains to stdout
    echo "$domains"
}

# Function to check if domain already exists in nginx configuration
check_domain_exists() {
    local domain="$1"
    
    if [ -z "$domain" ]; then
        return 1
    fi
    
    # Check if nginx config file already exists
    if [ -f "$NGINX_SITES_PATH/$domain" ]; then
        log "⚠️  Domain configuration already exists: $NGINX_SITES_PATH/$domain" >&2
        return 0
    fi
    
    # Check if domain exists in main nginx.conf
    if [ -f "$NGINX_CONFIG_PATH" ] && grep -q "server_name.*$domain" "$NGINX_CONFIG_PATH"; then
        log "⚠️  Domain already exists in main nginx.conf: $domain" >&2
        return 0
    fi
    
    # Check if domain is enabled (symlink exists)
    if [ -L "/etc/nginx/sites-enabled/$domain" ]; then
        log "⚠️  Domain is already enabled: /etc/nginx/sites-enabled/$domain" >&2
        return 0
    fi
    
    return 1
}

# Function to cleanup old domain configuration
cleanup_old_domain() {
    local domain="$1"
    
    if [ -z "$domain" ]; then
        return 0
    fi
    
    log "🗑️  Cleaning up configuration for old domain: $domain"
    
    # Remove SSL certificate
    if [ -d "$SSL_CERT_PATH/$domain" ]; then
        log "Removing SSL certificate for $domain..."
        certbot delete --cert-name "$domain" --non-interactive 2>/dev/null || true
    fi
    
    # Remove nginx site config if exists
    if [ -f "$NGINX_SITES_PATH/$domain" ]; then
        log "Removing nginx site config for $domain..."
        rm -f "$NGINX_SITES_PATH/$domain"
        rm -f "/etc/nginx/sites-enabled/$domain"
    fi
    
    log "✅ Cleanup completed for $domain"
}

# Function to generate SSL certificate for domain
generate_ssl_certificate() {
    local domain="$1"
    
    if [ -z "$domain" ]; then
        return 1
    fi
    
    log "🔐 Generating SSL certificate for $domain..."
    
    # Check if certificate already exists
    if [ -d "$SSL_CERT_PATH/$domain" ]; then
        log "SSL certificate already exists for $domain"
        return 0
    fi
    
    # Generate certificate using certbot
    certbot certonly \
        --nginx \
        --non-interactive \
        --agree-tos \
        --email "${ADMIN_EMAIL:-admin@example.com}" \
        --domains "$domain" \
        --expand \
        || {
            log "⚠️  Failed to generate SSL certificate for $domain, trying webroot method..."
            certbot certonly \
                --webroot \
                --webroot-path=/var/www/certbot \
                --non-interactive \
                --agree-tos \
                --email "${ADMIN_EMAIL:-admin@example.com}" \
                --domains "$domain" \
                || error "Failed to generate SSL certificate for $domain"
        }
    
    log "✅ SSL certificate generated for $domain"
}

# Function to generate nginx server block for partner domain
generate_partner_nginx_config() {
    local domain="$1"
    local config_file="$NGINX_SITES_PATH/$domain"
    
    log "📝 Generating nginx configuration for $domain..."
    
    # Create sites directories if they don't exist
    mkdir -p "$NGINX_SITES_PATH"
    mkdir -p "/etc/nginx/sites-enabled"
    
    # Generate nginx server block
    cat > "$config_file" << EOF
# Partner domain configuration: $domain
# Auto-generated by manage-partner-domains.sh

server {
    listen 80;
    server_name $domain;
    
    # Let's Encrypt challenge
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }
    
    # Redirect to HTTPS
    location / {
        return 301 https://\$server_name\$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name $domain;
    
    # SSL Configuration
    ssl_certificate $SSL_CERT_PATH/$domain/fullchain.pem;
    ssl_certificate_key $SSL_CERT_PATH/$domain/privkey.pem;
    
    # SSL Security Settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;
    ssl_stapling on;
    ssl_stapling_verify on;
    
    # Security Headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    
    # CORS Headers for partner domain
    add_header Access-Control-Allow-Origin "https://$domain" always;
    add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
    add_header Access-Control-Allow-Headers "Authorization, Content-Type, Accept, Origin, X-Requested-With" always;
    add_header Access-Control-Allow-Credentials "true" always;
    
    # Handle preflight requests
    if (\$request_method = 'OPTIONS') {
        add_header Access-Control-Allow-Origin "https://$domain" always;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
        add_header Access-Control-Allow-Headers "Authorization, Content-Type, Accept, Origin, X-Requested-With" always;
        add_header Access-Control-Max-Age 86400;
        add_header Content-Length 0;
        add_header Content-Type text/plain;
        return 204;
    }
    
    # Rate limiting
    limit_req zone=general burst=10 nodelay;
    
    # Client settings
    client_max_body_size 100M;
    
    # API endpoints
    location /api/ {
        limit_req zone=api burst=20 nodelay;
        
        proxy_pass http://backend:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header X-Partner-Domain \$host;
        
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_cache_bypass \$http_upgrade;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # Public site (partner branded)
    location / {
        proxy_pass http://public-site:80;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header X-Partner-Domain \$host;
        
        # Cache static assets
        location ~* \.(jpg|jpeg|png|gif|ico|css|js|svg|woff|woff2|ttf|eot)$ {
            proxy_pass http://public-site:80;
            proxy_cache_valid 200 30d;
            expires 30d;
            add_header Cache-Control "public, immutable";
        }
    }
    
    # Health check endpoint
    location /health {
        access_log off;
        return 200 "OK";
        add_header Content-Type text/plain;
    }
}
EOF
    
    # Enable the site
    ln -sf "$config_file" "/etc/nginx/sites-enabled/$domain"
    
    log "✅ Nginx configuration created for $domain"
}

# Function to update main nginx config to include partner domains
update_main_nginx_config() {
    log "🔧 Updating main nginx configuration..."
    
    local active_domains=$(get_active_domains)
    
    # Backup current configuration
    cp "$NGINX_CONFIG_PATH" "$NGINX_CONFIG_PATH.backup.$(date +%Y%m%d_%H%M%S)"
    
    # Generate nginx config via script
    log "Calling nginx config generator..."
    if "$SCRIPT_DIR/generate-nginx-config.sh" api; then
        log "✅ Nginx configuration generated successfully"
    else
        log "❌ Failed to generate nginx configuration"
        return 1
    fi
    
    log "✅ Main nginx configuration updated"
}

# Function to validate nginx configuration
validate_nginx_config() {
    log "🔍 Validating nginx configuration..."
    
    # Try Docker nginx container first (preferred method)
    if docker ps --format "table {{.Names}}" | grep -q "nginx" 2>/dev/null; then
        if docker exec nginx nginx -t 2>/dev/null; then
            log "✅ Nginx configuration is valid (Docker)"
            return 0
        else
            log "❌ Nginx configuration validation failed (Docker)"
            docker exec nginx nginx -t 2>&1 | while IFS= read -r line; do
                log "  $line" >&2
            done
            return 1
        fi
    # Fallback to local nginx if available
    elif command -v nginx >/dev/null 2>&1; then
        if nginx -c "$NGINX_CONFIG_PATH" -t 2>/dev/null; then
            log "✅ Nginx configuration is valid (Local)"
            return 0
        else
            log "❌ Nginx configuration validation failed (Local)"
            nginx -c "$NGINX_CONFIG_PATH" -t 2>&1 | while IFS= read -r line; do
                log "  $line" >&2
            done
            return 1
        fi
    else
        log "⚠️  Cannot validate nginx configuration - no nginx command or container found"
        return 0
    fi
}

# Function to reload nginx
reload_nginx() {
    log "🔄 Reloading nginx..."
    
    # Try Docker nginx container first (preferred method)
    if docker ps --format "table {{.Names}}" | grep -q "nginx" 2>/dev/null; then
        if docker exec nginx nginx -s reload 2>/dev/null; then
            log "✅ Nginx reloaded successfully (Docker)"
            return 0
        else
            log "❌ Failed to reload nginx in container"
            return 1
        fi
    # Fallback to systemctl if nginx service is running
    elif systemctl is-active --quiet nginx 2>/dev/null; then
        if systemctl reload nginx 2>/dev/null; then
            log "✅ Nginx reloaded successfully (Systemctl)"
            return 0
        else
            log "❌ Failed to reload nginx via systemctl"
            return 1
        fi
    else
        log "⚠️  Cannot reload nginx - no nginx container or service found"
        return 1
    fi
}

# Function to update CORS in backend
update_backend_cors() {
    local domain="$1"
    
    log "🌐 Updating CORS configuration for $domain..."
    
    # The backend should dynamically handle CORS based on X-Partner-Domain header
    # This is already implemented in the nginx config above
    
    log "✅ CORS configuration updated"
}

# Main execution logic
case "$OPERATION" in
    add)
        if [ -z "$NEW_DOMAIN" ]; then
            NEW_DOMAIN="$OLD_DOMAIN"
        fi
        
        if [ -z "$NEW_DOMAIN" ]; then
            error "Domain is required for add operation"
        fi
        
        log "➕ Adding new partner domain: $NEW_DOMAIN"
        
        # Check if domain already exists
        if check_domain_exists "$NEW_DOMAIN"; then
            log "⚠️  Domain $NEW_DOMAIN already exists in nginx configuration"
            log "🔄 Updating existing domain configuration instead..."
            
            # Clean up existing configuration first
            cleanup_old_domain "$NEW_DOMAIN"
        else
            log "✅ Domain $NEW_DOMAIN is new, proceeding with configuration..."
        fi
        
        # Generate SSL certificate first
        generate_ssl_certificate "$NEW_DOMAIN"
        
        # Generate nginx config for the domain
        generate_partner_nginx_config "$NEW_DOMAIN"
        
        # Update CORS
        update_backend_cors "$NEW_DOMAIN"
        
        # Update main nginx config
        update_main_nginx_config
        
        # Validate and reload
        validate_nginx_config
        reload_nginx
        
        log "✅ Successfully added partner domain: $NEW_DOMAIN"
        ;;
        
    update)
        if [ -z "$OLD_DOMAIN" ] || [ -z "$NEW_DOMAIN" ]; then
            error "Both old and new domains are required for update operation"
        fi
        
        log "🔄 Updating partner domain from $OLD_DOMAIN to $NEW_DOMAIN"
        
        # Cleanup old domain
        cleanup_old_domain "$OLD_DOMAIN"
        
        # Add new domain
        generate_ssl_certificate "$NEW_DOMAIN"
        generate_partner_nginx_config "$NEW_DOMAIN"
        update_backend_cors "$NEW_DOMAIN"
        
        # Update main nginx config
        update_main_nginx_config
        
        # Validate and reload
        validate_nginx_config
        reload_nginx
        
        log "✅ Successfully updated partner domain from $OLD_DOMAIN to $NEW_DOMAIN"
        ;;
        
    delete)
        if [ -z "$OLD_DOMAIN" ]; then
            error "Domain is required for delete operation"
        fi
        
        log "🗑️  Deleting partner domain: $OLD_DOMAIN"
        
        # Cleanup domain
        cleanup_old_domain "$OLD_DOMAIN"
        
        # Update main nginx config
        update_main_nginx_config
        
        # Validate and reload
        validate_nginx_config
        reload_nginx
        
        log "✅ Successfully deleted partner domain: $OLD_DOMAIN"
        ;;
        
    refresh)
        log "🔄 Refreshing all partner domains configuration..."
        
        # Get all active domains and regenerate configs
        active_domains=$(get_active_domains)
        
        # Debug: show what domains we got
        log "🔍 Debug: Active domains retrieved:" >&2
        echo "DEBUG: [$active_domains]" >&2
        
        if [ -n "$active_domains" ]; then
            while IFS= read -r domain; do
                # Clean up domain - remove timestamps, emoji, brackets
                domain=$(echo "$domain" | grep -E '^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' || echo "")
                
                if [ -n "$domain" ]; then
                    log "🔍 Processing domain: $domain" >&2
                    
                    # Check if domain already exists and warn, but continue
                    if check_domain_exists "$domain"; then
                        log "🔄 Domain $domain already configured, refreshing..." >&2
                        cleanup_old_domain "$domain"
                    fi
                    
                    # Check if SSL exists, if not generate
                    if [ ! -d "$SSL_CERT_PATH/$domain" ]; then
                        generate_ssl_certificate "$domain"
                    fi
                    
                    # Regenerate nginx config
                    generate_partner_nginx_config "$domain"
                else
                    log "⚠️  Skipping invalid domain line" >&2
                fi
            done <<< "$active_domains"
        else
            log "⚠️  No active domains found" >&2
        fi
        
        # Update main config
        update_main_nginx_config
        
        # Validate and reload
        validate_nginx_config
        reload_nginx
        
        log "✅ Successfully refreshed all partner domains"
        ;;
        
    *)
        error "Invalid operation. Use: add, update, delete, or refresh"
        ;;
esac

echo "=========================================="
echo "✅ Partner domain management completed successfully!"
echo "=========================================="