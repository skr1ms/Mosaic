#!/bin/bash
set -e

# Script for updating monitoring configuration with partner domains
# Used in CI/CD pipeline and can be run manually

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Function for logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

echo "=========================================="
echo "📊 Updating monitoring configuration with partner domains"
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

# Function to update Prometheus configuration
update_prometheus_config() {
    local partner_domains="$1"
    local prometheus_config="$PROJECT_ROOT/monitoring/prometheus.yml"
    local temp_config="$prometheus_config.tmp"
    
    log "🔧 Updating Prometheus configuration..."
    
    # Read existing config as base
    cp "$prometheus_config" "$temp_config"
    
    # Create partner domains targets list
    local targets="'photo.doyoupaint.com', 'adm.doyoupaint.com'"
    
    if [ -n "$partner_domains" ]; then
        while IFS= read -r domain; do
            if [ -n "$domain" ]; then
                targets="$targets, '$domain'"
            fi
        done <<< "$partner_domains"
    fi
    
    # Replace targets in health monitoring job
    sed -i "s|- 'photo.doyoupaint.com'|          - $targets|g" "$temp_config"
    sed -i '/- photo.doyoupaint.com/d' "$temp_config"
    sed -i '/- adm.doyoupaint.com/d' "$temp_config"
    
    # Move temp file to main location
    mv "$temp_config" "$prometheus_config"
    
    log "✅ Prometheus configuration updated"
}

# Function to restart Prometheus
restart_prometheus() {
    log "🔄 Restarting Prometheus service..."
    
    if docker-compose -f "$PROJECT_ROOT/monitoring/docker-compose.monitoring.yml" restart prometheus; then
        log "✅ Prometheus restarted successfully"
    else
        log "❌ Failed to restart Prometheus"
        return 1
    fi
    
    # Wait for Prometheus to be ready
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "http://localhost:9090/-/ready" > /dev/null 2>&1; then
            log "✅ Prometheus is ready (attempt $attempt)"
            return 0
        fi
        
        log "⏳ Waiting for Prometheus to be ready (attempt $attempt/$max_attempts)..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log "❌ Prometheus readiness check timeout"
    return 1
}

# Function to create Grafana dashboard for partner domains
create_grafana_dashboard() {
    local partner_domains="$1"
    local dashboard_file="$PROJECT_ROOT/monitoring/grafana/dashboards/partner-domains.json"
    
    log "📊 Creating Grafana dashboard for partner domains..."
    
    # Create basic dashboard JSON
    cat > "$dashboard_file" << 'EOF'
{
  "dashboard": {
    "id": null,
    "title": "Partner Domains Monitoring",
    "tags": ["partner", "domains", "monitoring"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Partner Domains Health",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=\"partner-domains-health\"}",
            "legendFormat": "{{ instance }}"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "mappings": [
              {
                "options": {
                  "0": { "text": "DOWN", "color": "red" },
                  "1": { "text": "UP", "color": "green" }
                }
              }
            ]
          }
        },
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Domain Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "probe_duration_seconds{job=\"partner-domains-health\"}",
            "legendFormat": "{{ instance }}"
          }
        ],
        "yAxes": [
          {
            "label": "Seconds",
            "min": 0
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0}
      },
      {
        "id": 3,
        "title": "SSL Certificate Days Until Expiry",
        "type": "graph",
        "targets": [
          {
            "expr": "(probe_ssl_earliest_cert_expiry{job=\"partner-domains-health\"} - time()) / 86400",
            "legendFormat": "{{ instance }}"
          }
        ],
        "yAxes": [
          {
            "label": "Days",
            "min": 0
          }
        ],
        "gridPos": {"h": 8, "w": 24, "x": 0, "y": 8}
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}
EOF
    
    log "✅ Grafana dashboard created"
}

# Function to reload Grafana dashboards
reload_grafana_dashboards() {
    log "🔄 Reloading Grafana dashboards..."
    
    # Restart Grafana to pickup new dashboard
    if docker-compose -f "$PROJECT_ROOT/monitoring/docker-compose.monitoring.yml" restart grafana; then
        log "✅ Grafana restarted successfully"
    else
        log "❌ Failed to restart Grafana"
        return 1
    fi
    
    # Wait for Grafana to be ready
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "http://localhost:3000/api/health" > /dev/null 2>&1; then
            log "✅ Grafana is ready (attempt $attempt)"
            return 0
        fi
        
        log "⏳ Waiting for Grafana to be ready (attempt $attempt/$max_attempts)..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log "❌ Grafana readiness check timeout"
    return 1
}

# Function to test monitoring setup
test_monitoring_setup() {
    local partner_domains="$1"
    local failed_tests=0
    
    log "🧪 Testing monitoring setup..."
    
    # Test Prometheus targets
    log "Testing Prometheus targets..."
    local prometheus_targets
    prometheus_targets=$(curl -s "http://localhost:9090/api/v1/targets" | jq -r '.data.activeTargets[] | select(.labels.job=="partner-domains-health") | .discoveredLabels.__address__')
    
    if [ -n "$prometheus_targets" ]; then
        log "✅ Prometheus targets configured:"
        echo "$prometheus_targets" | while IFS= read -r target; do
            if [ -n "$target" ]; then
                log "   - $target"
            fi
        done
    else
        log "❌ No Prometheus targets found"
        failed_tests=$((failed_tests + 1))
    fi
    
    # Test Grafana dashboard
    log "Testing Grafana dashboard..."
    if curl -s -f "http://localhost:3000/api/search?query=Partner%20Domains" | jq -e '.[] | select(.title=="Partner Domains Monitoring")' > /dev/null; then
        log "✅ Grafana dashboard found"
    else
        log "❌ Grafana dashboard not found"
        failed_tests=$((failed_tests + 1))
    fi
    
    return $failed_tests
}

# Main function
main() {
    log "Starting monitoring configuration update..."
    
    # Check required environment variables
    if [ -z "$POSTGRES_PASSWORD" ]; then
        log "❌ POSTGRES_PASSWORD environment variable is required"
        exit 1
    fi
    
    # Check if monitoring services are running
    if ! docker-compose -f "$PROJECT_ROOT/monitoring/docker-compose.monitoring.yml" ps | grep -q "Up"; then
        log "⚠️  Starting monitoring services..."
        docker-compose -f "$PROJECT_ROOT/monitoring/docker-compose.monitoring.yml" up -d
        sleep 30
    fi
    
    # Get partner domains list
    local domains
    domains=$(get_partner_domains)
    
    # Update Prometheus configuration
    update_prometheus_config "$domains"
    
    # Restart Prometheus
    if ! restart_prometheus; then
        log "❌ Failed to restart Prometheus"
        exit 1
    fi
    
    # Create Grafana dashboard
    create_grafana_dashboard "$domains"
    
    # Reload Grafana dashboards
    if ! reload_grafana_dashboards; then
        log "❌ Failed to reload Grafana dashboards"
        exit 1
    fi
    
    # Test monitoring setup
    if ! test_monitoring_setup "$domains"; then
        log "⚠️  Some monitoring tests failed, but configuration was updated"
    fi
    
    log "✅ Monitoring configuration update completed successfully!"
    
    if [ -n "$domains" ]; then
        log "📋 Monitoring enabled for partner domains:"
        echo "$domains" | while IFS= read -r domain; do
            if [ -n "$domain" ]; then
                log "   - https://$domain"
            fi
        done
    else
        log "ℹ️  No partner domains found, monitoring base configuration only"
    fi
    
    log "🌐 Access monitoring:"
    log "   - Prometheus: http://localhost:9090"
    log "   - Grafana: http://localhost:3000 (admin/admin)"
    log "   - Partner Domains Dashboard: http://localhost:3000/d/partner-domains"
}

# Run main function
main "$@"
