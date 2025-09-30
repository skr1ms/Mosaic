#!/bin/bash
# scripts/cleanup-ssl-certificates.sh
# Script for removing SSL certificates for a specific domain

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

DOMAIN_TO_REMOVE="$1"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >&2
}

if [ -z "$DOMAIN_TO_REMOVE" ]; then
    log "❌ Error: Domain to remove not specified"
    log "Usage: $0 <domain>"
    exit 1
fi

echo "=========================================="
echo "🗑️  SSL Certificate Cleanup Script"
echo "=========================================="

log "Cleaning up SSL certificates for domain: $DOMAIN_TO_REMOVE"

# Check that domain is not a system domain
SYSTEM_DOMAINS=("photo.doyoupaint.com" "adm.doyoupaint.com")
for sys_domain in "${SYSTEM_DOMAINS[@]}"; do
    if [ "$DOMAIN_TO_REMOVE" = "$sys_domain" ]; then
        log "⚠️  Warning: Skipping cleanup of system domain: $DOMAIN_TO_REMOVE"
        exit 0
    fi
done

# Check that domain is valid
if [[ ! "$DOMAIN_TO_REMOVE" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
    log "❌ Error: Invalid domain format: $DOMAIN_TO_REMOVE"
    exit 1
fi

# Remove Let's Encrypt certificates
if [ -d "/etc/letsencrypt/live/$DOMAIN_TO_REMOVE" ]; then
    log "🗑️  Removing Let's Encrypt certificates for $DOMAIN_TO_REMOVE..."
    
    # Use Docker for certbot, as in the main script
    if docker run --rm --name certbot-cleanup \
        -v "/etc/letsencrypt:/etc/letsencrypt" \
        -v "/var/lib/letsencrypt:/var/lib/letsencrypt" \
        certbot/certbot:latest \
        delete --cert-name "$DOMAIN_TO_REMOVE" --non-interactive 2>&1; then
        log "✅ Let's Encrypt certificate deleted successfully"
    else
        log "⚠️  Warning: Failed to delete Let's Encrypt certificate via certbot"
        log "ℹ️  Attempting manual cleanup..."
        
        # Manual file cleanup
        rm -rf "/etc/letsencrypt/live/$DOMAIN_TO_REMOVE" 2>/dev/null || true
        rm -rf "/etc/letsencrypt/archive/$DOMAIN_TO_REMOVE" 2>/dev/null || true
        log "✅ Manual cleanup completed"
    fi
else
    log "ℹ️  No Let's Encrypt certificate found for $DOMAIN_TO_REMOVE"
fi

# Remove nginx configuration files for the specific domain (if any)
if [ -f "/etc/nginx/sites-available/$DOMAIN_TO_REMOVE" ]; then
    log "🗑️  Removing nginx site config for $DOMAIN_TO_REMOVE..."
    rm -f "/etc/nginx/sites-available/$DOMAIN_TO_REMOVE"
    rm -f "/etc/nginx/sites-enabled/$DOMAIN_TO_REMOVE"
    log "✅ Nginx site config removed"
fi

# Reload nginx
log "🔄 Reloading nginx configuration..."
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

log "✅ SSL cleanup completed for domain: $DOMAIN_TO_REMOVE"
echo "=========================================="
