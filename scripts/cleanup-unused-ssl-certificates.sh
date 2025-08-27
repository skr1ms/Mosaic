#!/bin/bash
# scripts/cleanup-unused-ssl-certificates.sh
# Script for comprehensive cleanup of unused SSL certificates

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

ACTIVE_DOMAINS_INPUT="$1"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >&2
}

if [ -z "$ACTIVE_DOMAINS_INPUT" ]; then
    log "❌ Error: Active domains not specified"
    log "Usage: $0 '<space-separated list of active domains>'"
    exit 1
fi

echo "=========================================="
echo "🧹 Unused SSL Certificates Cleanup Script"
echo "=========================================="

# Convert string to array
read -ra ACTIVE_DOMAINS <<< "$ACTIVE_DOMAINS_INPUT"

log "Active domains: ${ACTIVE_DOMAINS[*]}"

# Get list of all existing certificates
EXISTING_CERTS=()
if [ -d "/etc/letsencrypt/live" ]; then
    while IFS= read -r -d '' cert_dir; do
        cert_name=$(basename "$cert_dir")
        # Skip README and other service files
        if [ "$cert_name" != "README" ] && [ -f "$cert_dir/fullchain.pem" ]; then
            EXISTING_CERTS+=("$cert_name")
        fi
    done < <(find /etc/letsencrypt/live -maxdepth 1 -type d -print0 2>/dev/null || true)
fi

log "Existing certificates: ${EXISTING_CERTS[*]}"

# Check each existing certificate
for cert in "${EXISTING_CERTS[@]}"; do
    SHOULD_KEEP=false
    
    # Check if this domain is in the active list
    for active_domain in "${ACTIVE_DOMAINS[@]}"; do
        # Remove spaces and check exact match
        active_domain=$(echo "$active_domain" | xargs)
        if [ "$cert" = "$active_domain" ]; then
            SHOULD_KEEP=true
            break
        fi
    done
    
    # If certificate is not needed, remove it
    if [ "$SHOULD_KEEP" = false ]; then
        log "🗑️  Removing unused certificate: $cert"
        
        # Remove Let's Encrypt certificate via Docker
        if docker run --rm --name certbot-cleanup \
            -v "/etc/letsencrypt:/etc/letsencrypt" \
            -v "/var/lib/letsencrypt:/var/lib/letsencrypt" \
            certbot/certbot:latest \
            delete --cert-name "$cert" --non-interactive 2>&1; then
            log "✅ Certificate deleted via certbot: $cert"
        else
            log "⚠️  Warning: Failed to delete certificate $cert via certbot"
            log "ℹ️  Attempting manual cleanup..."
            
            # Manual file cleanup
            rm -rf "/etc/letsencrypt/live/$cert" 2>/dev/null || true
            rm -rf "/etc/letsencrypt/archive/$cert" 2>/dev/null || true
            log "✅ Manual cleanup completed for: $cert"
        fi
        
        # Remove nginx configuration files (if any)
        if [ -f "/etc/nginx/sites-available/$cert" ]; then
            log "🗑️  Removing nginx site config for $cert..."
            rm -f "/etc/nginx/sites-available/$cert"
            rm -f "/etc/nginx/sites-enabled/$cert"
        fi
        
        log "✅ Removed certificate: $cert"
    else
        log "✅ Keeping certificate: $cert"
    fi
done

# Cleanup certificate archive files older than 90 days
if [ -d "/etc/letsencrypt/archive" ]; then
    log "🧹 Cleaning up old certificate archives..."
    find /etc/letsencrypt/archive -type f -mtime +90 -name "*.pem" -delete 2>/dev/null || true
    log "✅ Old certificate archives cleaned up"
fi

# Cleanup certbot logs older than 30 days
if [ -d "/var/log/letsencrypt" ]; then
    log "🧹 Cleaning up old certbot logs..."
    find /var/log/letsencrypt -type f -mtime +30 -name "*.log*" -delete 2>/dev/null || true
    log "✅ Old certbot logs cleaned up"
fi

# Test and reload nginx
log "🔄 Testing and reloading nginx configuration..."
if docker exec nginx nginx -t 2>/dev/null; then
    docker exec nginx nginx -s reload 2>/dev/null || {
        log "⚠️  Warning: Failed to reload nginx via Docker"
        log "ℹ️  Attempting systemctl reload..."
        systemctl reload nginx 2>/dev/null || true
    }
    log "✅ Nginx configuration reloaded successfully"
else
    log "⚠️  Warning: Nginx configuration test failed"
    log "ℹ️  Skipping nginx reload"
fi

log "✅ Comprehensive SSL cleanup completed successfully"
echo "=========================================="
