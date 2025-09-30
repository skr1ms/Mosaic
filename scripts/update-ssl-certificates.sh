#!/bin/bash
set -e

# Script for automatic SSL certificate updates for new partner domains
# Used in CI/CD pipeline

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Configuration
CERTBOT_EMAIL="${LETSENCRYPT_EMAIL:-admin@doyoupaint.com}"
WEBROOT_PATH="/var/www/certbot"
CERT_NAME="photo.doyoupaint.com-0001"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >&2
}

echo "=========================================="
echo "🔐 SSL Certificates Update Script"
echo "=========================================="

get_current_cert_domains() {
    local cert_path="/etc/letsencrypt/live/$CERT_NAME/fullchain.pem"
    
    if [ -f "$cert_path" ]; then
        # Extract domains from certificate
        openssl x509 -in "$cert_path" -noout -text 2>/dev/null | \
        grep -A1 "Subject Alternative Name" | \
        tail -1 | \
        sed 's/DNS://g' | \
        sed 's/, /\n/g' | \
        sed 's/[[:space:]]//g' | \
        grep -E '^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' | \
        sort | uniq
    fi
}

get_active_domains() {
    # Main domains (always included)
    local base_domains="photo.doyoupaint.com
adm.doyoupaint.com"
    
    # Partner domains
    local partner_domains=""
    if docker ps --format "table {{.Names}}" | grep -q "postgres" 2>/dev/null; then
        partner_domains=$(docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" postgres psql -U postgres -d mosaic -t -c "SELECT domain FROM partners WHERE status = 'active' AND domain IS NOT NULL AND domain != '' AND domain ~ '^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$';" 2>/dev/null | sed 's/[[:space:]]//g' | grep -v '^$' || echo "")
    fi
    
    # Combine and clean domains
    {
        echo "$base_domains"
        if [ -n "$partner_domains" ]; then
            echo "$partner_domains"
        fi
    } | sed 's/[[:space:]]//g' | \
    grep -E '^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$' | \
    sort | uniq
}

check_dns_records() {
    local domains="$1"
    local temp_file=$(mktemp)
    
    while IFS= read -r domain; do
        if [ -n "$domain" ] && [[ "$domain" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
            # Get A record with timeout and error handling
            local ip_address
            ip_address=$(timeout 10s dig +short "$domain" A 2>/dev/null | grep -E '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$' | head -1 || echo "")
            
            if [ -n "$ip_address" ]; then
                echo "$domain" >> "$temp_file"
            fi
        fi
    done <<< "$domains"
    
    if [ -f "$temp_file" ]; then
        cat "$temp_file"
        rm -f "$temp_file"
    fi
}

setup_webroot() {
    log "📁 Setting up webroot directory..."
    
    # Create directory if it doesn't exist
    if docker ps --format "table {{.Names}}" | grep -q "nginx" 2>/dev/null; then
        # Create the full path including .well-known/acme-challenge
        docker exec nginx mkdir -p "$WEBROOT_PATH/.well-known/acme-challenge" 2>/dev/null
        docker exec nginx chown -R nginx:nginx "$WEBROOT_PATH" 2>/dev/null || true
        docker exec nginx chmod -R 755 "$WEBROOT_PATH" 2>/dev/null || true
        log "✅ Webroot directory ready: $WEBROOT_PATH/.well-known/acme-challenge"
        
        # Verify directory exists
        if docker exec nginx ls -la "$WEBROOT_PATH/.well-known/acme-challenge" >/dev/null 2>&1; then
            log "✅ Challenge directory confirmed"
            return 0
        else
            log "❌ Failed to create challenge directory"
            return 1
        fi
    else
        log "❌ Nginx container not found"
        return 1
    fi
}

generate_certificate() {
    local domains="$1"
    local is_renewal="${2:-false}"
    
    if [ -z "$domains" ]; then
        log "❌ No domains provided for certificate generation"
        return 1
    fi
    
    log "🔐 Generating SSL certificate..."
    
    # Create domain string for certbot
    local domain_args=""
    local domain_count=0
    
    while IFS= read -r domain; do
        if [ -n "$domain" ] && [[ "$domain" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
            domain_args="$domain_args -d $domain"
            domain_count=$((domain_count + 1))
        fi
    done <<< "$domains"
    
    if [ $domain_count -eq 0 ]; then
        log "❌ No valid domains found for certificate generation"
        return 1
    fi
    
    log "Certificate will be generated for $domain_count domains:$domain_args"
    
    # Certbot command
    local certbot_cmd="certonly --webroot --webroot-path=$WEBROOT_PATH --email $CERTBOT_EMAIL --agree-tos --no-eff-email --non-interactive --cert-name $CERT_NAME"
    
    if [ "$is_renewal" = "true" ]; then
        certbot_cmd="$certbot_cmd --expand --force-renewal"
        log "📱 Expanding existing certificate with new domains (force renewal)"
    else
        log "🆕 Creating new certificate"
    fi
    
    log "Running: certbot $certbot_cmd $domain_args"
    
    # Execute certbot through Docker with proper error handling
    # Find the correct certbot volume name
    local certbot_volume
    certbot_volume=$(docker volume ls --format "table {{.Name}}" | grep "certbot_www" | head -n1 || echo "certbot_www")
    log "📂 Using certbot volume: $certbot_volume"
    
    if docker run --rm --name certbot \
        -v "/etc/letsencrypt:/etc/letsencrypt" \
        -v "/var/lib/letsencrypt:/var/lib/letsencrypt" \
        -v "${certbot_volume}:$WEBROOT_PATH" \
        certbot/certbot:latest \
        $certbot_cmd $domain_args 2>&1; then
        
        log "✅ SSL certificate generated successfully"
        return 0
    else
        log "❌ Failed to generate SSL certificate"
        return 1
    fi
}

check_certificate_expiry() {
    local cert_path="/etc/letsencrypt/live/$CERT_NAME/fullchain.pem"
    
    if [ -f "$cert_path" ]; then
        local expiry_date
        expiry_date=$(openssl x509 -in "$cert_path" -noout -enddate 2>/dev/null | cut -d= -f2)
        local expiry_timestamp
        expiry_timestamp=$(date -d "$expiry_date" +%s 2>/dev/null || echo "0")
        local current_timestamp
        current_timestamp=$(date +%s)
        
        if [ "$expiry_timestamp" -gt 0 ]; then
            local days_until_expiry
            days_until_expiry=$(( (expiry_timestamp - current_timestamp) / 86400 ))
            
            log "📅 Certificate expires in $days_until_expiry days ($expiry_date)"
            
            if [ $days_until_expiry -lt 30 ]; then
                log "⚠️  Certificate expires soon (less than 30 days)"
                return 1
            else
                log "✅ Certificate is valid for $days_until_expiry more days"
                return 0
            fi
        else
            log "❌ Could not parse certificate expiry date"
            return 1
        fi
    else
        log "❌ Certificate file not found: $cert_path"
        return 1
    fi
}

reload_nginx() {
    log "🔄 Reloading nginx configuration..."
    
    if ! docker ps --format "table {{.Names}}" | grep -q "nginx" 2>/dev/null; then
        log "❌ Nginx container not found"
        return 1
    fi
    
    # Test nginx configuration
    if docker exec nginx nginx -t >/dev/null 2>&1; then
        # Reload nginx
        if docker exec nginx nginx -s reload >/dev/null 2>&1; then
            log "✅ Nginx reloaded successfully"
            return 0
        else
            log "❌ Failed to reload nginx"
            return 1
        fi
    else
        log "❌ Nginx configuration test failed"
        return 1
    fi
}

backup_certificate() {
    local backup_dir="/etc/letsencrypt/backups/$(date +%Y%m%d_%H%M%S)"
    
    log "💾 Creating certificate backup..."
    
    if [ -d "/etc/letsencrypt/live/$CERT_NAME" ]; then
        mkdir -p "$backup_dir" 2>/dev/null
        cp -r "/etc/letsencrypt/live/$CERT_NAME" "$backup_dir/" 2>/dev/null || true
        cp -r "/etc/letsencrypt/archive/$CERT_NAME" "$backup_dir/" 2>/dev/null || true
        
        log "✅ Certificate backup created: $backup_dir"
        
        # Clean up old backups (keep last 5)
        if [ -d "/etc/letsencrypt/backups" ]; then
            local backup_count
            backup_count=$(ls -1 /etc/letsencrypt/backups/ 2>/dev/null | wc -l)
            if [ "$backup_count" -gt 5 ]; then
                ls -t /etc/letsencrypt/backups/ | tail -n +6 | xargs -I {} rm -rf "/etc/letsencrypt/backups/{}" 2>/dev/null || true
                log "🧹 Cleaned up old backups"
            fi
        fi
    else
        log "⚠️  No existing certificate to backup"
    fi
}

verify_domains_ready() {
    local domains="$1"
    local failed_count=0
    local total_count=0
    
    log "🔍 Verifying all domains are ready for SSL..."
    
    # Create test file content
    local test_content="acme-test-$(date +%s)"
    local test_file="$WEBROOT_PATH/.well-known/acme-challenge/test-$(date +%s)"
    
    # Ensure .well-known/acme-challenge directory exists
    docker exec nginx mkdir -p "$(dirname "$test_file")" 2>/dev/null || true
    
    while IFS= read -r domain; do
        if [ -n "$domain" ] && [[ "$domain" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
            total_count=$((total_count + 1))
            log "Verifying $domain..."
            
            # Create test file
            docker exec nginx sh -c "echo '$test_content' > '$test_file'" 2>/dev/null || true
            
            # Test domain availability via HTTP
            local test_url="http://$domain/.well-known/acme-challenge/$(basename "$test_file")"
            local response
            response=$(timeout 10s curl -s --max-time 10 "$test_url" 2>/dev/null || echo "")
            
            if [ "$response" = "$test_content" ]; then
                log "✅ Domain verification OK: $domain"
            else
                log "❌ Domain verification failed: $domain"
                failed_count=$((failed_count + 1))
            fi
            
            # Remove test file
            docker exec nginx rm -f "$test_file" 2>/dev/null || true
        fi
    done <<< "$domains"
    
    if [ $failed_count -gt 0 ]; then
        log "⚠️  $failed_count out of $total_count domains failed verification"
        log "⚠️  Proceeding with certificate generation anyway"
    else
        log "✅ All domains verified successfully"
    fi
    
    return 0
}

domains_changed() {
    local current_domains="$1"
    local active_domains="$2"
    
    # Sort both lists and compare
    local current_sorted
    current_sorted=$(echo "$current_domains" | sort | tr '\n' ' ')
    local active_sorted
    active_sorted=$(echo "$active_domains" | sort | tr '\n' ' ')
    
    if [ "$current_sorted" != "$active_sorted" ]; then
        log "🔄 Domain list has changed, certificate update required"
        return 0
    else
        log "✅ Domain list unchanged"
        return 1
    fi
}

log_domains_list() {
    local title="$1"
    local domains="$2"
    
    log "$title"
    if [ -n "$domains" ]; then
        while IFS= read -r domain; do
            if [ -n "$domain" ]; then
                log "   - $domain"
            fi
        done <<< "$domains"
    else
        log "   (none)"
    fi
}

log_dns_check_results() {
    local domains="$1"
    local valid_domains="$2"
    
    log "🔍 Checking DNS records for all domains..."
    
    while IFS= read -r domain; do
        if [ -n "$domain" ] && [[ "$domain" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
            log "Checking DNS for $domain..."
            
            # Check if domain is in valid list
            if echo "$valid_domains" | grep -q "^$domain$"; then
                # Get IP for logging
                local ip_address
                ip_address=$(timeout 10s dig +short "$domain" A 2>/dev/null | grep -E '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$' | head -1 || echo "")
                log "✅ DNS OK for $domain: $ip_address"
            else
                log "❌ DNS failed for $domain"
            fi
        fi
    done <<< "$domains"
}

main() {
    local force_update="${1:-false}"
    
    log "Starting SSL certificates update process..."
    
    # Check required environment variables
    if [ -z "$POSTGRES_PASSWORD" ]; then
        log "❌ POSTGRES_PASSWORD environment variable is required"
        exit 1
    fi
    
    # Setup webroot
    if ! setup_webroot; then
        log "❌ Failed to setup webroot directory"
        exit 1
    fi
    
    # Get current domains from certificate
    log "📋 Getting current domains from certificate..."
    local current_domains
    current_domains=$(get_current_cert_domains)
    log_domains_list "✅ Current certificate domains:" "$current_domains"
    
    # Get all active domains
    log "📋 Getting active domains from database..."
    local active_domains
    active_domains=$(get_active_domains)
    log_domains_list "✅ Active domains from database:" "$active_domains"
    
    if [ -z "$active_domains" ]; then
        log "❌ No active domains found"
        exit 1
    fi
    
    # Check DNS records for all domains
    local valid_domains
    valid_domains=$(check_dns_records "$active_domains")
    log_dns_check_results "$active_domains" "$valid_domains"
    
    if [ -z "$valid_domains" ]; then
        log "❌ No domains with valid DNS records found"
        exit 1
    fi
    
    # Check if domains have changed
    local domains_have_changed=false
    if domains_changed "$current_domains" "$valid_domains"; then
        domains_have_changed=true
    fi
    
    # Check certificate expiry
    local cert_expiring=false
    if ! check_certificate_expiry; then
        cert_expiring=true
        log "🔄 Certificate needs renewal due to expiry"
    fi
    
    # Determine if update is needed
    if [ "$domains_have_changed" = "true" ] || [ "$cert_expiring" = "true" ] || [ "$force_update" = "true" ]; then
        
        log "🔐 Certificate update is required"
        
        # Create backup
        backup_certificate
        
        # Check domain readiness
        verify_domains_ready "$valid_domains"
        
        # Determine update type
        local is_renewal="false"
        if [ -n "$current_domains" ]; then
            is_renewal="true"
        fi
        
        # Generate new certificate
        if generate_certificate "$valid_domains" "$is_renewal"; then
            
            # Reload nginx
            if reload_nginx; then
                log "✅ SSL certificates updated successfully!"
                
                # Test HTTPS for each domain
                while IFS= read -r domain; do
                    if [ -n "$domain" ] && [[ "$domain" =~ ^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
                        log "🔍 Testing HTTPS for $domain..."
                        if timeout 10s curl -s -f --max-time 10 "https://$domain/health" > /dev/null 2>&1; then
                            log "✅ HTTPS OK for $domain"
                        else
                            log "⚠️  HTTPS test failed for $domain (may need time to propagate)"
                        fi
                    fi
                done <<< "$valid_domains"
                
            else
                log "❌ Failed to reload nginx after certificate update"
                exit 1
            fi
            
        else
            log "❌ Failed to generate SSL certificate"
            exit 1
        fi
        
    else
        log "✅ No certificate update required"
        log "Current certificate is valid and covers all active domains"
    fi
    
    log "✅ SSL certificate update process completed successfully!"
}

force_update() {
    log "🔄 Forcing SSL certificate update..."
    main "true"
}

check_status() {
    log "📊 Checking SSL certificate status..."
    
    echo "=========================================="
    echo "Current Certificate Status"
    echo "=========================================="
    
    # Get current domains
    local current_domains
    current_domains=$(get_current_cert_domains)
    if [ -n "$current_domains" ]; then
        echo "$current_domains"
    else
        echo "No certificate found"
    fi
    
    check_certificate_expiry 2>/dev/null || true
    
    echo ""
    echo "=========================================="
    echo "Active Domains in Database"
    echo "=========================================="
    
    # Get active domains
    local active_domains
    active_domains=$(get_active_domains)
    if [ -n "$active_domains" ]; then
        echo "$active_domains"
    else
        echo "No active domains found"
    fi
}

# Check command line parameters
case "${1:-update}" in
    "update")
        main
        ;;
    "force")
        force_update
        ;;
    "status")
        check_status
        ;;
    *)
        log "❌ Unknown command: $1"
        echo "Usage: $0 [update|force|status]"
        echo ""
        echo "Commands:"
        echo "  update  - Update certificates if needed (default)"
        echo "  force   - Force certificate update"
        echo "  status  - Show certificate status"
        exit 1
        ;;
esac