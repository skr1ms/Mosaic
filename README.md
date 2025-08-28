# Mosaic Project

A comprehensive web application for managing photo mosaics and partner domains.

## Description

Mosaic is a full-stack application that provides:

- Photo mosaic generation and management
- Partner domain management
- SSL certificate automation
- Payment processing integration
- Admin dashboard
- Public-facing website

## Technology Stack

### Backend

- **Language**: Go
- **Database**: PostgreSQL
- **Cache**: Redis
- **Storage**: MinIO (S3-compatible)
- **AI**: Stable Diffusion integration

### Frontend

- **Admin Dashboard**: React.js
- **Public Site**: React.js with Vite
- **Styling**: SCSS, Tailwind CSS

### Infrastructure

- **Containerization**: Docker & Docker Compose
- **Web Server**: Nginx
- **SSL**: Let's Encrypt with automatic renewal
- **Monitoring**: Grafana, Prometheus, Loki

## Quick Start

1. Clone the repository:

```bash
git clone <repository-url>
cd Mosaic
```

2. Copy environment file:

```bash
cp .env.example .env
```

3. Configure your environment variables in `.env`

4. **Configure GitLab CI/CD Variables** (for production deployment):

   Go to your GitLab project → Settings → CI/CD → Variables and add the following variables:

   | Variable               | Description                       | Protected | Masked |
   | ---------------------- | --------------------------------- | --------- | ------ |
   | `POSTGRES_PASSWORD`    | Database password                 | ✅        | ✅     |
   | `PRODUCTION_SERVER_IP` | Production server IP address      | ✅        | ❌     |
   | `PRODUCTION_USER`      | SSH username for deployment       | ✅        | ❌     |
   | `SSH_PRIVATE_KEY`      | SSH private key for server access | ✅        | ✅     |

   **Note**: Mark sensitive variables as "Protected" and "Masked" for security.

5. Start the application:

```bash
# Development
docker compose --env-file .env -f deployments/docker-compose/docker-compose.dev.yml up --build

# Production
docker compose --env-file .env -f deployments/docker-compose/docker-compose.prod.yml up --build
```

## Environment Variables

See `.env.example` for all required environment variables.

## Development

### Prerequisites

- Docker & Docker Compose
- Go 1.24+
- Node.js 20+
- PostgreSQL
- Redis

### Running Tests

```bash
# Backend tests
cd backend
go test ./...
```

## Deployment

The project includes CI/CD pipeline configuration for GitLab with automatic domain management.

### Automatic CI/CD Pipeline

The system automatically triggers CI/CD pipelines when:

- **New partner is created** with a domain
- **Partner domain is updated**
- **Partner domain is removed**

The pipeline automatically:

- Updates nginx configuration with new domains
- Generates SSL certificates for new domains
- **Cleans up unused SSL certificates** (removes certificates for deleted domains)
- Updates monitoring configuration
- Reloads nginx configuration

### CI/CD Pipeline Architecture

The project uses a **dual-pipeline architecture**:

1. **Main Pipeline** (`.gitlab-ci.yml`):
   - Runs on `main`/`develop` branch pushes
   - Executes full test suite
   - Builds and deploys application
   - Includes business logic tests

2. **Domain Update Pipeline** (`.gitlab-ci-domain-update.yml`):
   - Triggered automatically from Go backend
   - Runs only when partner domains change
   - Updates nginx and SSL certificates
   - Cleans up unused certificates

### Domain Management Features

- **Automatic SSL Certificate Management**: New domains get SSL certificates automatically
- **SSL Certificate Cleanup**: Removed domains have their certificates cleaned up
- **Nginx Configuration Updates**: Automatic nginx config generation and deployment
- **Domain Validation**: DNS validation before SSL certificate generation
- **Error Handling**: Comprehensive error handling and logging

### Manual Deployment

```bash
# Update SSL certificates
./scripts/update-ssl-certificates.sh

# Clean up unused SSL certificates
./scripts/get-active-domains-from-db.sh
./scripts/cleanup-unused-ssl-certificates.sh

# Clean up specific domain SSL certificate
./scripts/cleanup-ssl-certificates.sh example.com

# Health check
./scripts/health-check.sh

# Post-deployment checks
./scripts/post-deploy-checks.sh
```

### SSL Certificate Management Scripts

The project includes several scripts for SSL certificate management:

- **`update-ssl-certificates.sh`**: Updates SSL certificates for all active domains
- **`cleanup-unused-ssl-certificates.sh`**: Removes SSL certificates for inactive domains
- **`cleanup-ssl-certificates.sh`**: Removes SSL certificate for a specific domain
- **`get-active-domains-from-db.sh`**: Retrieves list of active domains from database

### API Endpoints

#### Partner Management
- `POST /api/admin/partners` - Create new partner (triggers CI/CD)
- `PUT /api/admin/partners/:id` - Update partner (triggers CI/CD if domain changed)
- `DELETE /api/admin/partners/:id` - Delete partner (triggers CI/CD for cleanup)

#### Nginx Management
- `POST /api/admin/nginx/deploy` - Force nginx configuration update (used by CI/CD)

## Security Features

- **Environment Variable Protection**: Sensitive data stored in environment variables
- **SSL Certificate Automation**: Automatic SSL certificate generation and renewal
- **Domain Validation**: DNS validation before SSL certificate generation
- **Secure CI/CD**: Protected GitLab CI/CD variables with masking
- **Error Handling**: Comprehensive error handling and logging
- **Input Validation**: All API inputs are validated and sanitized

## Architecture

### Backend Architecture

The backend follows a clean architecture pattern:

- **Handlers**: HTTP request/response handling
- **Services**: Business logic and orchestration
- **Repositories**: Data access layer
- **Models**: Data structures and validation
- **Packages**: Utility packages (GitLab client, goroutine manager, etc.)

### Domain Management Flow

1. **Partner Creation/Update**: Admin creates or updates partner with domain
2. **Domain Validation**: System validates domain format and uniqueness
3. **CI/CD Trigger**: Backend triggers GitLab CI/CD pipeline asynchronously
4. **Nginx Update**: CI/CD pipeline calls Go API to update nginx configuration
5. **SSL Management**: Pipeline generates SSL certificates for new domains
6. **Cleanup**: Pipeline removes SSL certificates for deleted domains
7. **Monitoring Update**: Pipeline updates monitoring configuration

### Error Handling

- **Graceful Degradation**: System continues working even if CI/CD fails
- **Comprehensive Logging**: All operations are logged with timestamps
- **Timeout Management**: All async operations have timeouts
- **Retry Logic**: Failed operations are retried with exponential backoff

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
