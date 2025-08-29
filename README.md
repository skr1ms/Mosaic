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

   #### **Step 1: Create Pipeline Trigger Token**

Go to your GitLab project → Settings → CI/CD → Pipeline triggers:

- Click "New pipeline trigger"
- Name: `domain-update-trigger`
- Click "Create pipeline trigger"
- Copy the generated token (starts with `glptt-`)

#### **Step 2: Add CI/CD Variables**

Go to your GitLab project → Settings → CI/CD → Variables and add the following variables:

#### **Required Variables for Deployment:**

| Variable               | Description                       | Protected | Masked | Default Value |
| ---------------------- | -------------------------------- | --------- | ------ | ------------- |
| `PRODUCTION_SERVER_IP` | Production server IP address     | ✅        | ❌     | -             |
| `PRODUCTION_USER`      | SSH username for deployment      | ✅        | ❌     | -             |
| `SSH_PRIVATE_KEY`      | SSH private key for server access| ✅        | ✅     | -             |

#### **Domain Update Pipeline Variables:**

| Variable           | Description                                | Protected | Masked | Default Value |
| ------------------ | ------------------------------------------- | --------- | ------ | ------------- |
| `DOMAIN_UPDATE`    | Flag to trigger domain update pipeline      | ✅        | ❌     | `false`       |
| `DOMAIN_OPERATION` | Type of domain operation (refresh, update)  | ✅        | ❌     | `refresh`     |
| `OLD_DOMAIN`       | Previous domain name (for updates)          | ❌        | ❌     | (empty)       |
| `NEW_DOMAIN`       | New domain name (for updates)               | ❌        | ❌     | (empty)       |
| `BACKEND_URL`      | Backend API URL for nginx config updates     | ❌        | ❌     | `https://photo.doyoupaint.com` |
| `DEFAULT_ADMIN_EMAIL` | Admin email for API authentication        | ✅        | ❌     | -             |
| `DEFAULT_ADMIN_PASSWORD` | Admin password for API authentication   | ✅        | ✅     | -             |

#### **Optional Variables:**

| Variable            | Description                              | Protected | Masked | Default Value |
| ------------------- | ---------------------------------------- | --------- | ------ | ------------- |
| `SLACK_WEBHOOK_URL` | Slack webhook for pipeline notifications | ❌        | ❌     | (empty)       |

**Note**:

- Mark sensitive variables as "Protected" and "Masked" for security
- Domain variables are automatically set by the backend when triggering pipelines
- SSH_PRIVATE_KEY must be masked to prevent exposure in logs

5. Start the application:

```bash
# Development
docker compose --env-file .env -f deployments/docker-compose/docker-compose.dev.yml up --build

# Production
docker compose --env-file .env -f deployments/docker-compose/docker-compose.prod.yml up --build
```

## Environment Variables

See `.env.example` for all required environment variables.

### GitLab Configuration Variables

For the backend to trigger GitLab pipelines, you need these variables in your `.env` file:

```bash
# GitLab Configuration
GITLAB_BASE_URL=https://gitlab.com
GITLAB_API_TOKEN=glpat-your-api-token-here
GITLAB_PROJECT_ID=your-project-id

GITLAB_TRIGGER_TOKEN=glptt-your-trigger-token-here
```

**Important**:

- **GitLab API Token** (`GITLAB_API_TOKEN`) - for basic GitLab API operations
- **GitLab Trigger Token** (`GITLAB_TRIGGER_TOKEN`) - for launching pipelines with variables
- API token must have `api` scope and `write_repository` permissions
- Trigger token is created in GitLab: `Settings` → `CI/CD` → `Pipeline triggers`

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
   - **Automatically triggered** from Go backend via GitLab API
   - **Uses trigger token** for variable support
   - Runs only when partner domains change
   - Updates nginx and SSL certificates
   - Cleans up unused certificates
   - **Variables are automatically set** by the backend when triggering

#### **Pipeline Triggering Process:**

1. **Backend Action**: Admin creates/updates/removes partner domain
2. **API Call**: Go backend calls GitLab API with domain variables and `CI_CONFIG_PATH`
3. **Pipeline Creation**: GitLab creates new pipeline using `.gitlab-ci-domain-update.yml` file
4. **Automatic Execution**: Pipeline runs with domain-specific parameters
5. **Domain Management**: Scripts update nginx, SSL, and monitoring
6. **Notification**: Success/failure notifications sent to Slack (if configured)

### Domain Management Features

- **Automatic SSL Certificate Management**: New domains get SSL certificates automatically
- **SSL Certificate Cleanup**: Removed domains have their certificates cleaned up
- **Nginx Configuration Updates**: Automatic nginx config generation and deployment
- **Domain Validation**: DNS validation before SSL certificate generation
- **Error Handling**: Comprehensive error handling and logging

### Domain Update Pipeline Variables

The domain update pipeline uses specific variables that are automatically set by the backend:

- **`DOMAIN_UPDATE`**: Set to `true` when the backend triggers a domain update pipeline
- **`DOMAIN_OPERATION`**: Specifies the operation type:
  - `refresh`: Refresh all domains (default)
  - `update`: Update specific domain
  - `add`: Add new domain
  - `remove`: Remove domain
- **`OLD_DOMAIN`**: Previous domain name (for update operations)
- **`NEW_DOMAIN`**: New domain name (for add/update operations)
- **`BACKEND_URL`**: Backend API URL for nginx config updates (default: `https://photo.doyoupaint.com`)
- **`DEFAULT_ADMIN_EMAIL`**: Admin email for API authentication
- **`DEFAULT_ADMIN_PASSWORD`**: Admin password for API authentication

These variables are automatically populated by the Go backend when calling the GitLab API, so you don't need to set them manually in the GitLab interface.

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

### GitLab CI/CD Security

- **Dual Token Authentication**: Uses both API token and trigger token for different purposes
- **API Token**: For reading pipeline status and other GitLab API operations
- **Trigger Token**: For launching pipelines with variables (required by GitLab design)
- **Variable Masking**: Sensitive variables (SSH keys) are masked in logs
- **Protected Variables**: Critical variables are protected and only available on protected branches

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

## ⚠️ Important Note

**This is an open-source version of the Mosaic project. The mosaic generation functionality requires proprietary commercial components that are not included:**

- Python scripts for mosaic generation
- Excel palette files with color schemes
- These components must be obtained separately to enable full mosaic generation capabilities

**Required directory structure for mosaic generation:**

```
backend/
└── scripts/
    ├── mosaic_cli.py          # Main Python script
    ├── pallete_bw.xlsx        # Black & White palette
    ├── pallete_fl.xlsx        # Full color palette
    ├── pallete_max.xlsx       # Maximum color palette
    └── pallete_tl.xlsx        # Limited color palette
```

**All other functionality works normally without these commercial components.**
