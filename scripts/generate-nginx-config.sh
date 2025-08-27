#!/bin/bash
set -e

# Script for generating nginx configuration with partner domains
# Calls backend API to get current domain list

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Function for logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

echo "=========================================="
echo "🏗️  Generating nginx configuration with partner domains"
echo "=========================================="

# Function to get partner domains list from database
get_partner_domains() {
    log "📋 Getting partner domains from database..."
    
    # Use direct database connection to get domains
    local domains_query="SELECT domain FROM partners WHERE status = 'active' AND domain IS NOT NULL AND domain != '';"
    
    # Execute query via psql
    local domains
    domains=$(docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" postgres psql -U postgres -d mosaic -t -c "$domains_query" 2>/dev/null | tr -d ' ' | grep -v '^$' || echo "")
    
    if [ -n "$domains" ]; then
        echo "$domains"
        log "✅ Found partner domains: $(echo "$domains" | wc -l) domains"
    else
        log "ℹ️  No partner domains found in database"
        echo ""
    fi
}

# Function to generate nginx configuration
generate_nginx_config() {
    local partner_domains="$1"
    local config_path="$PROJECT_ROOT/deployments/nginx/production.conf"
    
    log "🔧 Generating nginx configuration..."
    
    # Create temporary file for new configuration
    local temp_config="$config_path.tmp"
    
    # Main nginx configuration matching existing structure
    cat > "$temp_config" << 'EOF'
events {
    worker_connections 1024;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # Logging
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for" '
                    'rt=$request_time uct="$upstream_connect_time" '
                    'uht="$upstream_header_time" urt="$upstream_response_time"';

    access_log /var/log/nginx/access.log main;

    # Basic Settings
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 100M;

    # Gzip Settings
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=general:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=api:10m rate=50r/s;

    # SSL settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;
    ssl_stapling on;
    ssl_stapling_verify on;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin";
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://www.google.com https://www.gstatic.com; frame-src 'self' https://www.google.com https://recaptcha.google.com; img-src 'self' data: https:; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' data: https://fonts.gstatic.com; connect-src 'self' https://www.google.com https://www.google.com/recaptcha/ https://unpkg.com;";

    # Upstream definitions
    upstream backend {
        least_conn;
        server backend:8080 max_fails=3 fail_timeout=30s;
    }

    upstream public-site {
        server public-site:80 max_fails=3 fail_timeout=30s;
    }

    upstream dashboards {
        server dashboards:80 max_fails=3 fail_timeout=30s;
    }

    # Default server block to handle unknown hosts
    server {
        listen 443 ssl default_server;
        server_name _;
        ssl_certificate /etc/letsencrypt/live/photo.doyoupaint.com-0001/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/photo.doyoupaint.com-0001/privkey.pem;
        return 444;
    }

    # HTTP to HTTPS redirect
    server {
        listen 80;
        server_name photo.doyoupaint.com adm.doyoupaint.com
EOF

    # Add partner domains to HTTP server_name
    if [ -n "$partner_domains" ]; then
        echo "$partner_domains" | while IFS= read -r domain; do
            if [ -n "$domain" ]; then
                echo " $domain" >> "$temp_config"
            fi
        done
    fi

    # Continue HTTP configuration
    cat >> "$temp_config" << 'EOF'
;
        
        # Let's Encrypt challenge
        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }
        
        # Redirect all other HTTP traffic to HTTPS
        location / {
            return 301 https://$server_name$request_uri;
        }
    }

    # Admin dashboard
    server {
        listen 443 ssl;
        http2 on;
        server_name adm.doyoupaint.com;

        ssl_certificate /etc/letsencrypt/live/photo.doyoupaint.com-0001/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/photo.doyoupaint.com-0001/privkey.pem;
        
        # API endpoints for admin dashboard
        location /api/ {
            limit_req zone=api burst=20 nodelay;
            
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache_bypass $http_upgrade;
            
            # CORS headers for API
            add_header Access-Control-Allow-Origin $http_origin always;
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
            add_header Access-Control-Allow-Headers "Authorization, Content-Type, X-Requested-With" always;
            add_header Access-Control-Allow-Credentials true always;
            
            if ($request_method = 'OPTIONS') {
                return 204;
            }
        }

        # MinIO Console proxy
        location /minio/ {
            rewrite ^/minio/(.*) /$1 break;
            proxy_pass http://minio:9001;

            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto https;
            proxy_set_header X-Forwarded-Host $host;
            proxy_set_header X-Forwarded-Port $server_port;
            proxy_set_header X-NginX-Proxy true;

            proxy_set_header X-Forwarded-Prefix /minio;
        
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_cache_bypass $http_upgrade;
            
            proxy_buffering off;
            proxy_redirect off;
            
            # Handle subpath properly
            proxy_cookie_path / /minio/;
        }

        # Grafana proxy
        location /grafana/ {
            proxy_pass http://grafana:3000/grafana/;

            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto http;
            proxy_set_header X-Forwarded-Host $host;
            proxy_set_header X-Forwarded-Port $server_port;

            proxy_set_header X-Forwarded-Prefix /grafana;
    
            proxy_redirect off;
            proxy_buffering off;
    
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_cache_bypass $http_upgrade;
            
            add_header Cache-Control "no-cache, no-store, must-revalidate";
            add_header Pragma "no-cache";
            add_header Expires "0";
        }

        # MinIO proxy for logos, images and chat data
        location /mosaic-logos/ {
            proxy_pass http://minio:9000/mosaic-logos/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache_bypass $http_upgrade;
            
            add_header Cache-Control "no-cache, no-store, must-revalidate";
            add_header Pragma "no-cache";
            add_header Expires "0";
        }

        location /mosaic-images/ {
            proxy_pass http://minio:9000/mosaic-images/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache_bypass $http_upgrade;
        }

        location /chat-data/ {
            proxy_pass http://minio:9000/chat-data/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache_bypass $http_upgrade;
        }

        location / {
            proxy_pass http://dashboards;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache_bypass $http_upgrade;
        }
    }
    
    # Main public site
    server {
        listen 443 ssl;
        http2 on;
        server_name photo.doyoupaint.com
EOF

    # Add partner domains to HTTPS server_name for public site
    if [ -n "$partner_domains" ]; then
        echo "$partner_domains" | while IFS= read -r domain; do
            if [ -n "$domain" ]; then
                echo " $domain" >> "$temp_config"
            fi
        done
    fi

    # Continue HTTPS configuration for public site
    cat >> "$temp_config" << 'EOF'
;

        ssl_certificate /etc/letsencrypt/live/photo.doyoupaint.com-0001/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/photo.doyoupaint.com-0001/privkey.pem;

        # Public site
        location / {
            proxy_pass http://public-site;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache_bypass $http_upgrade;
        }

        # API endpoints
        location /api/ {
            limit_req zone=api burst=50 nodelay;
            
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache_bypass $http_upgrade;
            
            # CORS headers for API
            add_header Access-Control-Allow-Origin $http_origin always;
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
            add_header Access-Control-Allow-Headers "Authorization, Content-Type, X-Requested-With" always;
            add_header Access-Control-Allow-Credentials true always;
            
            if ($request_method = 'OPTIONS') {
                return 204;
            }
        }

        # MinIO proxy for logos
        location /mosaic-logos/ {
            proxy_pass http://minio:9000/mosaic-logos/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache_bypass $http_upgrade;
            
            add_header Cache-Control "no-cache, no-store, must-revalidate";
            add_header Pragma "no-cache";
            add_header Expires "0";
        }

        location /mosaic-images/ {
            proxy_pass http://minio:9000/mosaic-images/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache_bypass $http_upgrade;
        }

        location /chat-data/ {
            proxy_pass http://minio:9000/chat-data/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache_bypass $http_upgrade;
        }
    }
}
EOF

    # Move temporary file to main location
    mv "$temp_config" "$config_path"
    
    log "✅ Nginx configuration generated: $config_path"
}

# Main function
main() {
    log "Starting nginx configuration generation..."
    
    # Check environment variables
    if [ -z "$POSTGRES_PASSWORD" ]; then
        log "❌ POSTGRES_PASSWORD environment variable is required"
        exit 1
    fi
    
    # Get partner domains list
    local domains
    domains=$(get_partner_domains)
    
    # Generate configuration
    generate_nginx_config "$domains"
    
    log "✅ Nginx configuration generation completed successfully!"
    
    if [ -n "$domains" ]; then
        log "📋 Partner domains included:"
        echo "$domains" | while IFS= read -r domain; do
            if [ -n "$domain" ]; then
                log "   - https://$domain"
            fi
        done
    else
        log "ℹ️  No partner domains found, using base configuration"
    fi
}

# Run main function
main "$@"