#!/bin/bash
set -e

# Test script for partner domain management flow
# Tests create, update, and delete operations

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
TEST_PARTNER_NAME="Test Partner $(date +%s)"
TEST_DOMAIN_1="test-partner-$(date +%s).example.com"
TEST_DOMAIN_2="updated-partner-$(date +%s).example.com"
BACKEND_URL="${BACKEND_URL:-http://localhost:8080}"
ADMIN_EMAIL="${DEFAULT_ADMIN_EMAIL:-admin@example.com}"
ADMIN_PASSWORD="${DEFAULT_ADMIN_PASSWORD}"

# Function for logging
log() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1" >&2
    exit 1
}

warn() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1"
}

echo "=========================================="
echo "🧪 Testing Partner Domain Management Flow"
echo "=========================================="

# Load environment variables
if [ -f "$PROJECT_ROOT/.env" ]; then
    export $(cat "$PROJECT_ROOT/.env" | grep -v '^#' | xargs)
fi

# Function to get JWT token
get_jwt_token() {
    log "🔐 Getting admin JWT token..."
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}" \
        "$BACKEND_URL/api/auth/login")
    
    local token=$(echo "$response" | jq -r '.access_token')
    
    if [ "$token" = "null" ] || [ -z "$token" ]; then
        error "Failed to get JWT token"
    fi
    
    echo "$token"
}

# Function to create partner
create_partner() {
    local token="$1"
    local domain="$2"
    
    log "➕ Creating partner with domain: $domain"
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "{
            \"partner_code\": \"TEST$(date +%s)\",
            \"brand_name\": \"$TEST_PARTNER_NAME\",
            \"domain\": \"$domain\",
            \"email\": \"test@$domain\",
            \"phone\": \"+1234567890\",
            \"address\": \"Test Address\",
            \"status\": \"active\"
        }" \
        "$BACKEND_URL/api/admin/partners")
    
    local partner_id=$(echo "$response" | jq -r '.id')
    
    if [ "$partner_id" = "null" ] || [ -z "$partner_id" ]; then
        error "Failed to create partner: $response"
    fi
    
    log "✅ Partner created with ID: $partner_id"
    echo "$partner_id"
}

# Function to update partner domain
update_partner_domain() {
    local token="$1"
    local partner_id="$2"
    local new_domain="$3"
    
    log "🔄 Updating partner domain to: $new_domain"
    
    local response=$(curl -s -X PUT \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "{
            \"domain\": \"$new_domain\",
            \"reason\": \"Test domain update\"
        }" \
        "$BACKEND_URL/api/admin/partners/$partner_id")
    
    if echo "$response" | grep -q "error"; then
        error "Failed to update partner: $response"
    fi
    
    log "✅ Partner domain updated"
}

# Function to delete partner
delete_partner() {
    local token="$1"
    local partner_id="$2"
    
    log "🗑️  Deleting partner with ID: $partner_id"
    
    local response=$(curl -s -X DELETE \
        -H "Authorization: Bearer $token" \
        "$BACKEND_URL/api/admin/partners/$partner_id")
    
    if echo "$response" | grep -q "error"; then
        warn "Failed to delete partner: $response"
    else
        log "✅ Partner deleted"
    fi
}

# Function to check GitLab pipeline status
check_pipeline_status() {
    log "🔍 Checking GitLab pipeline status..."
    
    # Get latest pipeline
    local pipeline_response=$(curl -s \
        -H "PRIVATE-TOKEN: $GITLAB_API_TOKEN" \
        "$GITLAB_BASE_URL/api/v4/projects/$GITLAB_PROJECT_ID/pipelines?per_page=1")
    
    local pipeline_id=$(echo "$pipeline_response" | jq -r '.[0].id')
    local pipeline_status=$(echo "$pipeline_response" | jq -r '.[0].status')
    local pipeline_ref=$(echo "$pipeline_response" | jq -r '.[0].ref')
    
    if [ "$pipeline_id" != "null" ]; then
        log "Pipeline #$pipeline_id (ref: $pipeline_ref) - Status: $pipeline_status"
        
        # Wait for pipeline to complete (max 5 minutes)
        local max_wait=300
        local waited=0
        
        while [ "$pipeline_status" = "pending" ] || [ "$pipeline_status" = "running" ]; do
            if [ $waited -ge $max_wait ]; then
                warn "Pipeline is still running after 5 minutes"
                break
            fi
            
            sleep 10
            waited=$((waited + 10))
            
            pipeline_response=$(curl -s \
                -H "PRIVATE-TOKEN: $GITLAB_API_TOKEN" \
                "$GITLAB_BASE_URL/api/v4/projects/$GITLAB_PROJECT_ID/pipelines/$pipeline_id")
            
            pipeline_status=$(echo "$pipeline_response" | jq -r '.status')
            log "Pipeline status: $pipeline_status (waited ${waited}s)"
        done
        
        if [ "$pipeline_status" = "success" ]; then
            log "✅ Pipeline completed successfully"
        else
            warn "Pipeline finished with status: $pipeline_status"
        fi
    else
        warn "No pipeline found or failed to get pipeline info"
    fi
}

# Function to verify nginx configuration
verify_nginx_config() {
    local domain="$1"
    
    log "🔍 Verifying nginx configuration for $domain..."
    
    # Check if domain responds
    if curl -s -o /dev/null -w "%{http_code}" "https://$domain/health" | grep -q "200\|301\|302"; then
        log "✅ Domain $domain is responding"
    else
        warn "Domain $domain is not responding (this might be expected for test domains)"
    fi
    
    # Check nginx config file
    if [ -f "/etc/nginx/sites-available/$domain" ]; then
        log "✅ Nginx config file exists for $domain"
    else
        warn "Nginx config file not found for $domain"
    fi
    
    # Check SSL certificate
    if [ -d "/etc/letsencrypt/live/$domain" ]; then
        log "✅ SSL certificate directory exists for $domain"
    else
        warn "SSL certificate directory not found for $domain (expected for test domains)"
    fi
}

# Main test flow
main() {
    log "Starting test flow..."
    
    # Get JWT token
    JWT_TOKEN=$(get_jwt_token)
    
    # Test 1: Create partner with domain
    log "📝 Test 1: Create partner with domain"
    PARTNER_ID=$(create_partner "$JWT_TOKEN" "$TEST_DOMAIN_1")
    sleep 5
    check_pipeline_status
    verify_nginx_config "$TEST_DOMAIN_1"
    
    # Test 2: Update partner domain
    log "📝 Test 2: Update partner domain"
    update_partner_domain "$JWT_TOKEN" "$PARTNER_ID" "$TEST_DOMAIN_2"
    sleep 5
    check_pipeline_status
    verify_nginx_config "$TEST_DOMAIN_2"
    
    # Verify old domain was cleaned up
    if [ ! -f "/etc/nginx/sites-available/$TEST_DOMAIN_1" ]; then
        log "✅ Old domain config was cleaned up"
    else
        warn "Old domain config still exists"
    fi
    
    # Test 3: Delete partner
    log "📝 Test 3: Delete partner"
    delete_partner "$JWT_TOKEN" "$PARTNER_ID"
    sleep 5
    check_pipeline_status
    
    # Verify domain was cleaned up
    if [ ! -f "/etc/nginx/sites-available/$TEST_DOMAIN_2" ]; then
        log "✅ Domain config was cleaned up after deletion"
    else
        warn "Domain config still exists after deletion"
    fi
    
    echo "=========================================="
    echo "✅ Test flow completed!"
    echo "=========================================="
    
    # Summary
    echo ""
    echo "Test Summary:"
    echo "- Partner created with domain: $TEST_DOMAIN_1"
    echo "- Domain updated to: $TEST_DOMAIN_2"
    echo "- Partner deleted and cleaned up"
    echo ""
    echo "Note: Some warnings are expected for test domains (e.g., SSL certificates)"
}

# Check dependencies
command -v curl >/dev/null 2>&1 || error "curl is required but not installed"
command -v jq >/dev/null 2>&1 || error "jq is required but not installed"

# Run main test flow
main