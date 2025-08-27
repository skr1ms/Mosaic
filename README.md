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

4. Start the application:
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

The project includes CI/CD pipeline configuration for GitLab.

### Manual Deployment
```bash
# Update SSL certificates
./scripts/update-ssl-certificates.sh

# Health check
./scripts/health-check.sh

# Post-deployment checks
./scripts/post-deploy-checks.sh
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.