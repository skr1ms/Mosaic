# Partner Domain Management System

## Overview

The Partner Domain Management System enables White Label partners to use their own custom domains while leveraging the main platform infrastructure. Each partner gets a personalized version of the site under their brand and domain.

## Features

- **Automatic Domain Configuration**: When a partner is created/updated/deleted, the system automatically configures nginx and SSL certificates
- **SSL Certificate Management**: Automatic generation and cleanup of Let's Encrypt SSL certificates
- **CORS Support**: Dynamic CORS configuration for partner domains
- **Domain Updates**: Seamless domain changes with automatic cleanup of old configurations
- **CI/CD Integration**: GitLab CI/CD pipeline triggers for all domain operations

## Architecture

### Components

1. **Backend API** (`/api/admin/partners`)
   - Manages partner CRUD operations
   - Triggers GitLab CI/CD pipeline for domain updates
   - Handles CORS dynamically for partner domains

2. **GitLab CI/CD Pipeline** (`.gitlab-ci-domain-update.yml`)
   - Triggered by backend API with operation details
   - Executes domain management scripts on the server
   - Handles SSL certificate generation/cleanup

3. **Domain Management Script** (`scripts/manage-partner-domains.sh`)
   - Manages nginx configurations
   - Handles SSL certificates via certbot
   - Performs cleanup operations

4. **Nginx Configuration**
   - Dynamic server blocks for each partner domain
   - Proper SSL configuration
   - CORS headers for API access
   - Proxy to backend and frontend services

## Operations

### 1. Adding a New Partner Domain

When a partner is created with a domain:

```bash
# Backend triggers pipeline with:
DOMAIN_OPERATION=add
NEW_DOMAIN=partner.example.com
```

The system will:
1. Generate SSL certificate for the domain
2. Create nginx server block configuration
3. Update CORS settings
4. Reload nginx

### 2. Updating Partner Domain

When a partner's domain is changed:

```bash
# Backend triggers pipeline with:
DOMAIN_OPERATION=update
OLD_DOMAIN=old.example.com
NEW_DOMAIN=new.example.com
```

The system will:
1. Remove SSL certificate for old domain
2. Remove nginx config for old domain
3. Generate SSL certificate for new domain
4. Create nginx config for new domain
5. Update CORS settings
6. Reload nginx

### 3. Deleting Partner

When a partner is deleted:

```bash
# Backend triggers pipeline with:
DOMAIN_OPERATION=delete
OLD_DOMAIN=partner.example.com
```

The system will:
1. Remove SSL certificate
2. Remove nginx configuration
3. Clean up any related files
4. Reload nginx

## Configuration

### Environment Variables

Required environment variables in `.env`:

```bash
# GitLab Configuration
GITLAB_BASE_URL=https://gitlab.com
GITLAB_API_TOKEN=glpat-xxxxx
GITLAB_TRIGGER_TOKEN=glptt-xxxxx
GITLAB_PROJECT_ID=73841570

# Server Configuration
PRODUCTION_SERVER_IP=your.server.ip
PRODUCTION_USER=lumirin
SSH_PRIVATE_KEY=your-ssh-key

# Database Configuration
POSTGRES_USER=postgres
POSTGRES_PASSWORD=yourpassword
POSTGRES_DB=mosaic

# Admin Configuration
DEFAULT_ADMIN_EMAIL=admin@example.com
DEFAULT_ADMIN_PASSWORD=yourpassword
```

### Nginx Configuration Structure

Each partner domain gets its own nginx server block:

```nginx
server {
    listen 443 ssl http2;
    server_name partner.example.com;
    
    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/partner.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/partner.example.com/privkey.pem;
    
    # CORS Headers for partner domain
    add_header Access-Control-Allow-Origin "https://partner.example.com";
    
    # Proxy to backend
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header X-Partner-Domain $host;
    }
    
    # Proxy to frontend
    location / {
        proxy_pass http://public-site:80;
        proxy_set_header X-Partner-Domain $host;
    }
}
```

## Testing

### Manual Testing

Use the test script to verify the complete flow:

```bash
chmod +x scripts/test-partner-domain-flow.sh
./scripts/test-partner-domain-flow.sh
```

This will:
1. Create a test partner with a domain
2. Update the partner's domain
3. Delete the partner
4. Verify each operation

### Verification Steps

1. **Check GitLab Pipeline**:
   ```bash
   curl -H "PRIVATE-TOKEN: $GITLAB_API_TOKEN" \
     "$GITLAB_BASE_URL/api/v4/projects/$GITLAB_PROJECT_ID/pipelines"
   ```

2. **Check Nginx Configuration**:
   ```bash
   nginx -t
   ls -la /etc/nginx/sites-available/
   ```

3. **Check SSL Certificates**:
   ```bash
   certbot certificates
   ls -la /etc/letsencrypt/live/
   ```

4. **Test Domain Response**:
   ```bash
   curl -I https://partner.example.com/health
   ```

## Troubleshooting

### Pipeline Not Triggering

1. Check GitLab trigger token is correct
2. Verify environment variables are set
3. Check backend logs for trigger attempts
4. Ensure GitLab project ID is correct

### SSL Certificate Generation Fails

1. Ensure domain DNS points to server
2. Check port 80 is accessible
3. Verify certbot is installed
4. Check rate limits (Let's Encrypt has limits)

### Nginx Configuration Issues

1. Always validate config: `nginx -t`
2. Check error logs: `tail -f /var/log/nginx/error.log`
3. Ensure all proxy targets are accessible
4. Verify file permissions

### CORS Issues

1. Check browser console for CORS errors
2. Verify partner domain is in database
3. Check nginx headers are being sent
4. Ensure backend CORS middleware is working

## Security Considerations

1. **SSL/TLS**: All partner domains use HTTPS with modern TLS protocols
2. **Rate Limiting**: Nginx implements rate limiting for API endpoints
3. **Security Headers**: HSTS, X-Frame-Options, CSP headers are configured
4. **Domain Validation**: Only active partners with valid domains are configured
5. **Cleanup**: Old domains and certificates are automatically removed

## Monitoring

The system updates monitoring configuration automatically:

- Prometheus targets for each domain
- Health check endpoints
- SSL certificate expiry monitoring
- Nginx access/error logs

## Maintenance

### Regular Tasks

1. **Certificate Renewal**: Certbot auto-renewal via cron
2. **Log Rotation**: Nginx logs are rotated automatically
3. **Database Cleanup**: Remove inactive partner domains
4. **Monitoring**: Check pipeline success rates

### Manual Operations

```bash
# Refresh all partner domains
./scripts/manage-partner-domains.sh refresh

# Cleanup unused certificates
./scripts/cleanup-unused-ssl-certificates.sh

# Update monitoring config
./scripts/update-monitoring-config.sh
```

## API Endpoints

### Create Partner
```http
POST /api/admin/partners
Authorization: Bearer {token}

{
  "partner_code": "PART001",
  "brand_name": "Partner Brand",
  "domain": "partner.example.com",
  "email": "contact@partner.com",
  "status": "active"
}
```

### Update Partner
```http
PUT /api/admin/partners/{id}
Authorization: Bearer {token}

{
  "domain": "new-partner.example.com",
  "reason": "Domain change requested by partner"
}
```

### Delete Partner
```http
DELETE /api/admin/partners/{id}
Authorization: Bearer {token}
```

## Support

For issues or questions:
1. Check logs in `/var/log/nginx/`
2. Review GitLab pipeline logs
3. Check backend logs with `docker logs backend`
4. Verify database state with partner queries