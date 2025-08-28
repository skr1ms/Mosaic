#!/bin/bash
set -e

# Script for comprehensive post-deployment checks
# Performs full system validation including partner domains

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Function for logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Function for detailed check with logging
detailed_check() {
    local check_name="$1"
    local check_command="$2"
    local success_message="$3"
    local error_message="$4"
    
    log "🔍 $check_name..."
    
    if eval "$check_command"; then
        log "✅ $success_message"
        return 0
    else
        log "❌ $error_message"
        return 1
    fi
}

echo "=========================================="
echo "🚀 Post-Deploy Comprehensive Checks"
echo "=========================================="

# Function to check deployed image versions
check_deployed_versions() {
    log "📦 Checking deployed Docker image versions..."
    
    local services=("backend" "dashboards" "public-site")
    local failed_checks=0
    
    for service in "${services[@]}"; do
        local image_info
        image_info=$(docker inspect --format='{{.Config.Image}}' "$service" 2>/dev/null || echo "unknown")
        
        if [[ "$image_info" == *"$CI_COMMIT_SHA"* ]] || [[ "$image_info" == *"latest"* ]]; then
            log "✅ $service: $image_info"
        else
            log "❌ $service: unexpected image version - $image_info"
            failed_checks=$((failed_checks + 1))
        fi
    done
    
    return $failed_checks
}

# Function to check API endpoints
check_api_endpoints() {
    log "🔌 Checking API endpoints functionality..."
    
    local failed_checks=0
    
    # Check main API endpoints
    local api_endpoints=(
        "http://localhost:8080/api/health:status:healthy"
        "https://photo.doyoupaint.com/api/health:status:healthy"  
        "https://adm.doyoupaint.com/api/health:status:healthy"
    )
    
    for endpoint_config in "${api_endpoints[@]}"; do
        IFS=':' read -r url field expected_value <<< "$endpoint_config"
        
        log "Testing $url..."
        local response
        response=$(curl -s --max-time 15 "$url" 2>/dev/null || echo "")
        
        if [ -n "$response" ]; then
            local field_value
            field_value=$(echo "$response" | jq -r ".$field" 2>/dev/null || echo "")
            
            if [ "$field_value" = "$expected_value" ]; then
                log "✅ API $url is healthy"
            else
                log "❌ API $url returned unexpected value: $field_value"
                failed_checks=$((failed_checks + 1))
            fi
        else
            log "❌ API $url no response"
            failed_checks=$((failed_checks + 1))
        fi
    done
    
    return $failed_checks
}

# Function to check partner domains
check_partner_domains() {
    log "🏢 Checking partner domains functionality..."
    
    # Get active partner domains from database
    local partner_domains
    partner_domains=$(docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" postgres psql -U postgres -d mosaic -t -c "SELECT domain FROM partners WHERE status = 'active' AND domain IS NOT NULL AND domain != '';" 2>/dev/null | tr -d ' ' | grep -v '^$' || echo "")
    
    if [ -z "$partner_domains" ]; then
        log "ℹ️  No partner domains found in database"
        return 0
    fi
    
    local failed_checks=0
    local total_domains=0
    
    log "Found partner domains:"
    echo "$partner_domains" | while IFS= read -r domain; do
        if [ -n "$domain" ]; then
            log "   - $domain"
            total_domains=$((total_domains + 1))
        fi
    done
    
    echo "$partner_domains" | while IFS= read -r domain; do
        if [ -n "$domain" ]; then
            log "🔍 Testing partner domain: $domain"
            
            # Check HTTPS availability
            local https_url="https://$domain"
            if curl -s -f --max-time 15 "$https_url" > /dev/null 2>&1; then
                log "✅ HTTPS access OK for $domain"
            else
                log "❌ HTTPS access failed for $domain"
                failed_checks=$((failed_checks + 1))
            fi
            
            # Check API through partner domain
            local api_url="https://$domain/api/health"
            local api_response
            api_response=$(curl -s --max-time 15 "$api_url" 2>/dev/null || echo "")
            
            if [ -n "$api_response" ]; then
                local status
                status=$(echo "$api_response" | jq -r '.status' 2>/dev/null || echo "")
                if [ "$status" = "healthy" ]; then
                    log "✅ API access OK for $domain"
                else
                    log "❌ API returned unexpected status for $domain: $status"
                    failed_checks=$((failed_checks + 1))
                fi
            else
                log "❌ API access failed for $domain"
                failed_checks=$((failed_checks + 1))
            fi
            
            # Check White Label branding
            local branding_check
            branding_check=$(curl -s --max-time 15 "$https_url" | grep -i "partner\|brand" | head -1 || echo "")
            if [ -n "$branding_check" ]; then
                log "✅ White Label branding detected for $domain"
            else
                log "⚠️  White Label branding not detected for $domain (may be normal)"
            fi
        fi
    done
    
    if [ $failed_checks -eq 0 ]; then
        log "✅ All partner domains ($total_domains) are working correctly"
    else
        log "❌ $failed_checks issues found with partner domains"
    fi
    
    return $failed_checks
}

# Function to check SSL certificates
check_ssl_certificates() {
    log "🔐 Checking SSL certificates..."
    
    local failed_checks=0
    
    # Main domains
    local main_domains=("photo.doyoupaint.com" "adm.doyoupaint.com")
    
    for domain in "${main_domains[@]}"; do
        log "Checking SSL for $domain..."
        
        # Check SSL certificate
        local cert_info
        cert_info=$(echo | openssl s_client -connect "$domain:443" -servername "$domain" 2>/dev/null | openssl x509 -noout -dates 2>/dev/null || echo "")
        
        if [ -n "$cert_info" ]; then
            local not_after
            not_after=$(echo "$cert_info" | grep "notAfter" | cut -d= -f2)
            log "✅ SSL certificate valid for $domain until $not_after"
        else
            log "❌ SSL certificate check failed for $domain"
            failed_checks=$((failed_checks + 1))
        fi
    done
    
    # Partner domains
    local partner_domains
    partner_domains=$(docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" postgres psql -U postgres -d mosaic -t -c "SELECT domain FROM partners WHERE status = 'active' AND domain IS NOT NULL AND domain != '';" 2>/dev/null | tr -d ' ' | grep -v '^$' || echo "")
    
    if [ -n "$partner_domains" ]; then
        echo "$partner_domains" | while IFS= read -r domain; do
            if [ -n "$domain" ]; then
                log "Checking SSL for partner domain $domain..."
                
                local cert_info
                cert_info=$(echo | openssl s_client -connect "$domain:443" -servername "$domain" 2>/dev/null | openssl x509 -noout -dates 2>/dev/null || echo "")
                
                if [ -n "$cert_info" ]; then
                    local not_after
                    not_after=$(echo "$cert_info" | grep "notAfter" | cut -d= -f2)
                    log "✅ SSL certificate valid for $domain until $not_after"
                else
                    log "❌ SSL certificate check failed for $domain"
                    failed_checks=$((failed_checks + 1))
                fi
            fi
        done
    fi
    
    return $failed_checks
}

# Function to check monitoring
check_monitoring_services() {
    log "📊 Checking monitoring services..."
    
    local failed_checks=0
    
    # Check Prometheus
    log "Checking Prometheus..."
    if curl -s -f --max-time 10 "http://localhost:9090/-/healthy" > /dev/null 2>&1; then
        log "✅ Prometheus is healthy"
    else
        log "❌ Prometheus health check failed"
        failed_checks=$((failed_checks + 1))
    fi
    
    # Check Grafana
    log "Checking Grafana..."
    if curl -s -f --max-time 10 "http://localhost:3000/api/health" > /dev/null 2>&1; then
        log "✅ Grafana is healthy"
    else
        log "❌ Grafana health check failed"
        failed_checks=$((failed_checks + 1))
    fi
    
    # Check Loki
    log "Checking Loki..."
    if curl -s -f --max-time 10 "http://localhost:3100/ready" > /dev/null 2>&1; then
        log "✅ Loki is ready"
    else
        log "❌ Loki ready check failed"
        failed_checks=$((failed_checks + 1))
    fi
    
    return $failed_checks
}

# Function to check performance
check_performance() {
    log "⚡ Running performance checks..."
    
    local failed_checks=0
    
    # Check response time for main pages
    local urls=(
        "https://photo.doyoupaint.com"
        "https://adm.doyoupaint.com"
    )
    
    for url in "${urls[@]}"; do
        log "Checking response time for $url..."
        
        local response_time
        response_time=$(curl -o /dev/null -s -w '%{time_total}' --max-time 30 "$url" 2>/dev/null || echo "timeout")
        
        if [[ "$response_time" != "timeout" ]] && (( $(echo "$response_time < 5.0" | bc -l) )); then
            log "✅ Response time OK for $url: ${response_time}s"
        else
            log "❌ Slow response time for $url: ${response_time}s"
            failed_checks=$((failed_checks + 1))
        fi
    done
    
    return $failed_checks
}

# Function to check database integrity
check_database_integrity() {
    log "🗄️  Checking database integrity..."
    
    local failed_checks=0
    
    # Check database connection
    if docker exec postgres pg_isready -U postgres > /dev/null 2>&1; then
        log "✅ Database connection OK"
        
        # Check main tables
        local tables=("partners" "coupons" "users" "images")
        for table in "${tables[@]}"; do
            local count
            count=$(docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" postgres psql -U postgres -d mosaic -t -c "SELECT COUNT(*) FROM $table;" 2>/dev/null | tr -d ' ' || echo "error")
            
            if [[ "$count" =~ ^[0-9]+$ ]]; then
                log "✅ Table $table: $count records"
            else
                log "❌ Error checking table $table"
                failed_checks=$((failed_checks + 1))
            fi
        done
        
        # Check active partners
        local active_partners
        active_partners=$(docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" postgres psql -U postgres -d mosaic -t -c "SELECT COUNT(*) FROM partners WHERE status = 'active';" 2>/dev/null | tr -d ' ' || echo "error")
        
        if [[ "$active_partners" =~ ^[0-9]+$ ]]; then
            log "✅ Active partners: $active_partners"
        else
            log "❌ Error checking active partners"
            failed_checks=$((failed_checks + 1))
        fi
        
    else
        log "❌ Database connection failed"
        failed_checks=$((failed_checks + 1))
    fi
    
    return $failed_checks
}

# Function to check file system and storage
check_storage_services() {
    log "💾 Checking storage services..."
    
    local failed_checks=0
    
    # Check MinIO
    log "Checking MinIO service..."
    if curl -s -f --max-time 10 "http://localhost:9000/minio/health/live" > /dev/null 2>&1; then
        log "✅ MinIO is healthy"
    else
        log "❌ MinIO health check failed"
        failed_checks=$((failed_checks + 1))
    fi
    
    # Check Redis
    log "Checking Redis service..."
    if docker exec redis redis-cli ping | grep -q "PONG"; then
        log "✅ Redis is responding"
    else
        log "❌ Redis ping failed"
        failed_checks=$((failed_checks + 1))
    fi
    
    return $failed_checks
}

# Function to generate deployment report
generate_deploy_report() {
    local total_checks="$1"
    local failed_checks="$2"
    local start_time="$3"
    local end_time="$4"
    
    local report_file="$PROJECT_ROOT/deploy-report-$(date +%Y%m%d_%H%M%S).txt"
    local deploy_duration=$((end_time - start_time))
    
    cat > "$report_file" << EOF
========================================
DEPLOY REPORT - $(date)
========================================

Deploy Information:
- Commit SHA: ${CI_COMMIT_SHA:-"local"}
- Branch: ${CI_COMMIT_REF_NAME:-"unknown"}
- Duration: ${deploy_duration} seconds

Check Results:
- Total Checks: $total_checks  
- Passed: $((total_checks - failed_checks))
- Failed: $failed_checks
- Success Rate: $(( (total_checks - failed_checks) * 100 / total_checks ))%

System Status:
EOF
    
    if [ $failed_checks -eq 0 ]; then
        echo "✅ DEPLOY SUCCESSFUL - All systems operational" >> "$report_file"
    else
        echo "❌ DEPLOY ISSUES DETECTED - $failed_checks checks failed" >> "$report_file"
    fi
    
    echo "" >> "$report_file"
    echo "Generated at: $(date)" >> "$report_file"
    echo "========================================" >> "$report_file"
    
    log "📄 Deploy report generated: $report_file"
    
    # Send report to Slack/Teams/Email (if configured)
    if [ -n "$DEPLOY_WEBHOOK_URL" ]; then
        local status_emoji="✅"
        local status_text="SUCCESS"
        
        if [ $failed_checks -gt 0 ]; then
            status_emoji="❌"
            status_text="ISSUES DETECTED"
        fi
        
        local webhook_payload="{
            \"text\": \"Deploy Report - $status_emoji $status_text\",
            \"attachments\": [{
                \"color\": \"$([ $failed_checks -eq 0 ] && echo 'good' || echo 'danger')\",
                \"fields\": [
                    {\"title\": \"Commit\", \"value\": \"${CI_COMMIT_SHA:-local}\", \"short\": true},
                    {\"title\": \"Branch\", \"value\": \"${CI_COMMIT_REF_NAME:-unknown}\", \"short\": true},
                    {\"title\": \"Duration\", \"value\": \"${deploy_duration}s\", \"short\": true},
                    {\"title\": \"Success Rate\", \"value\": \"$(( (total_checks - failed_checks) * 100 / total_checks ))%\", \"short\": true}
                ]
            }]
        }"
        
        curl -s -X POST \
            -H "Content-Type: application/json" \
            -d "$webhook_payload" \
            "$DEPLOY_WEBHOOK_URL" > /dev/null 2>&1 || true
        
        log "📡 Deploy notification sent"
    fi
}

# Main function
main() {
    local start_time
    start_time=$(date +%s)
    local total_checks=0
    local total_failures=0
    
    log "Starting comprehensive post-deploy checks..."
    
    # Check environment variables
    if [ -z "$POSTGRES_PASSWORD" ]; then
        log "❌ POSTGRES_PASSWORD environment variable is required"
        exit 1
    fi
    
    echo ""
    echo "=========================================="
    echo "1. DEPLOYMENT VERIFICATION"
    echo "=========================================="
    
    # Check image versions
    log "📦 Verifying deployed versions..."
    local version_failures=0
    check_deployed_versions || version_failures=$?
    total_failures=$((total_failures + version_failures))
    total_checks=$((total_checks + 3)) # backend, dashboards, public-site
    
    echo ""
    echo "=========================================="
    echo "2. API ENDPOINTS"
    echo "=========================================="
    
    # Check API endpoints
    local api_failures=0
    check_api_endpoints || api_failures=$?
    total_failures=$((total_failures + api_failures))
    total_checks=$((total_checks + 3)) # health endpoints
    
    echo ""
    echo "=========================================="
    echo "3. PARTNER DOMAINS"
    echo "=========================================="
    
    # Check partner domains
    local partner_failures=0
    check_partner_domains || partner_failures=$?
    total_failures=$((total_failures + partner_failures))
    
    # Count partner domains for statistics
    local partner_count
    partner_count=$(docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" postgres psql -U postgres -d mosaic -t -c "SELECT COUNT(*) FROM partners WHERE status = 'active' AND domain IS NOT NULL AND domain != '';" 2>/dev/null | tr -d ' ' || echo "0")
    total_checks=$((total_checks + partner_count * 2)) # HTTPS + API for each partner
    
    echo ""
    echo "=========================================="
    echo "4. SSL CERTIFICATES"
    echo "=========================================="
    
    # Check SSL certificates
    local ssl_failures=0
    check_ssl_certificates || ssl_failures=$?
    total_failures=$((total_failures + ssl_failures))
    total_checks=$((total_checks + 2 + partner_count)) # main domains + partner domains
    
    echo ""
    echo "=========================================="
    echo "5. INFRASTRUCTURE SERVICES"
    echo "=========================================="
    
    # Check database
    local db_failures=0
    check_database_integrity || db_failures=$?
    total_failures=$((total_failures + db_failures))
    total_checks=$((total_checks + 5)) # connection + 4 tables
    
    # Check storage
    local storage_failures=0
    check_storage_services || storage_failures=$?
    total_failures=$((total_failures + storage_failures))
    total_checks=$((total_checks + 2)) # MinIO + Redis
    
    echo ""
    echo "=========================================="
    echo "6. MONITORING SERVICES"
    echo "=========================================="
    
    # Check monitoring
    local monitoring_failures=0
    check_monitoring_services || monitoring_failures=$?
    total_failures=$((total_failures + monitoring_failures))
    total_checks=$((total_checks + 3)) # Prometheus + Grafana + Loki
    
    echo ""
    echo "=========================================="
    echo "7. PERFORMANCE CHECKS"
    echo "=========================================="
    
    # Check performance
    local perf_failures=0
    check_performance || perf_failures=$?
    total_failures=$((total_failures + perf_failures))
    total_checks=$((total_checks + 2)) # main URLs
    
    # Completion
    local end_time
    end_time=$(date +%s)
    
    echo ""
    echo "=========================================="
    echo "FINAL RESULTS"
    echo "=========================================="
    
    log "📊 Total checks performed: $total_checks"
    log "✅ Passed: $((total_checks - total_failures))"
    log "❌ Failed: $total_failures"
    log "📈 Success rate: $(( (total_checks - total_failures) * 100 / total_checks ))%"
    log "⏱️  Total duration: $((end_time - start_time)) seconds"
    
    # Generate report
    generate_deploy_report "$total_checks" "$total_failures" "$start_time" "$end_time"
    
    if [ $total_failures -eq 0 ]; then
        echo ""
        log "🎉 🚀 DEPLOY COMPLETED SUCCESSFULLY! 🚀 🎉"
        log "All systems are operational and ready for production use."
        echo ""
        exit 0
    else
        echo ""
        log "⚠️  DEPLOY COMPLETED WITH ISSUES"
        log "$total_failures checks failed. Please review the issues above."
        echo ""
        exit 1
    fi
}

# Run main function
main "$@"
