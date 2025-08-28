#!/bin/bash
# scripts/get-active-domains-from-db.sh
# Script for getting active domains from database

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >&2
}

echo "=========================================="
echo "📋 Active Domains from Database Script"
echo "=========================================="

# Check that PostgreSQL container is running
if ! docker ps --format "table {{.Names}}" | grep -q "postgres" 2>/dev/null; then
    log "❌ Error: PostgreSQL container is not running"
    exit 1
fi

# Check environment variables
if [ -z "$POSTGRES_PASSWORD" ]; then
    log "❌ Error: POSTGRES_PASSWORD environment variable is required"
    exit 1
fi

log "📋 Getting active domains from database..."

# Connect to database and get active domains
ACTIVE_DOMAINS=$(docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" postgres psql -U postgres -d mosaic -t -c "SELECT domain FROM partners WHERE status = 'active' AND domain IS NOT NULL AND domain != '' AND domain ~ '^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$';" 2>/dev/null | sed 's/[[:space:]]//g' | grep -v '^$' || echo "")

# Add system domains
SYSTEM_DOMAINS="photo.doyoupaint.com adm.doyoupaint.com"

# Combine and clean domains
ALL_DOMAINS=""
if [ -n "$ACTIVE_DOMAINS" ]; then
    ALL_DOMAINS="$ACTIVE_DOMAINS"
fi

# Add system domains
for sys_domain in $SYSTEM_DOMAINS; do
    if [ -z "$ALL_DOMAINS" ]; then
        ALL_DOMAINS="$sys_domain"
    else
        ALL_DOMAINS="$ALL_DOMAINS
$sys_domain"
    fi
done

# Remove duplicates and output
FINAL_DOMAINS=$(echo "$ALL_DOMAINS" | sort -u | grep -v '^$' | tr '\n' ' ')

if [ -n "$FINAL_DOMAINS" ]; then
    log "✅ Active domains found:"
    echo "$FINAL_DOMAINS" | tr ' ' '\n' | while read -r domain; do
        if [ -n "$domain" ]; then
            log "   - $domain"
        fi
    done
    
    # Output result for use in other scripts
    echo "$FINAL_DOMAINS"
else
    log "⚠️  No active domains found"
    echo ""
fi

echo "=========================================="
