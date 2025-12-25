# OCAPI - OpenCart API

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![OpenCart](https://img.shields.io/badge/OpenCart-3.x%20%7C%204.x-23A8E0?style=flat)](https://www.opencart.com/)

A high-performance REST API service for OpenCart e-commerce platform. OCAPI enables external systems to manage products, categories, orders, and other OpenCart entities without direct database access.

## Features

- **Product Management** - Create, update, and retrieve products with full attribute support
- **Category Management** - Manage product categories and hierarchies
- **Order Processing** - Retrieve orders and update order statuses
- **Batch Operations** - Process multiple products in a single request with batch synchronization
- **Image Handling** - Upload and manage product images via base64 encoding
- **Multi-language Support** - Handle product descriptions in multiple languages
- **Currency Updates** - Update currency exchange rates programmatically
- **Graceful Shutdown** - Clean shutdown with in-flight request handling
- **Health Checks** - Built-in health endpoint for load balancer integration

## Tech Stack

- **Language:** Go 1.21+
- **Router:** [chi](https://github.com/go-chi/chi) - Lightweight, idiomatic HTTP router
- **Database:** MySQL (OpenCart database)
- **Config:** YAML with environment variable override
- **Logging:** Structured logging with slog

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Running OpenCart installation (3.x or 4.x)
- MySQL database access

### Installation

```bash
# Clone the repository
git clone https://github.com/nomadus/ocapi.git
cd ocapi

# Build the application
go build -v -o ocapi ./cmd/opencart

# Copy and configure
cp config.yml.example config.yml
# Edit config.yml with your database credentials
```

### Configuration

Create a `config.yml` file with your settings:

```yaml
env: prod

listen:
  bind_ip: 127.0.0.1
  port: 9800
  key: your-api-key

sql:
  enabled: true
  hostname: localhost
  username: opencart_user
  password: your_password
  database: opencart_db
  port: 3306
  prefix: oc_

images:
  path: /var/www/opencart/image/catalog/
  url: catalog/

product:
  custom_fields:
    - points
    - sort_order
```

See [Configuration Documentation](docs/config.md) for all options.

### Running

```bash
# Run directly
./ocapi -conf=config.yml -log=./

# Or with custom paths
./ocapi -conf=/etc/ocapi/config.yml -log=/var/log/
```

### Systemd Service

```ini
[Unit]
Description=OpenCart API Service
After=network.target mysql.service

[Service]
Type=simple
ExecStart=/usr/local/bin/ocapi -conf=/etc/ocapi/config.yml -log=/var/log/
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## API Overview

All API endpoints require Bearer token authentication (except `/health`).

```bash
curl -H "Authorization: Bearer your-api-key" \
     -H "Content-Type: application/json" \
     http://localhost:9800/api/v1/product/abc-123
```

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check (no auth) |
| `POST` | `/api/v1/product` | Create/update products |
| `GET` | `/api/v1/product/{uid}` | Get product by UID |
| `POST` | `/api/v1/product/description` | Update product descriptions |
| `POST` | `/api/v1/product/image` | Upload product image |
| `POST` | `/api/v1/category` | Create/update categories |
| `GET` | `/api/v1/order/{id}` | Get order details |
| `POST` | `/api/v1/order` | Update order status |
| `GET` | `/api/v1/orders/{statusId}` | List orders by status |
| `GET` | `/api/v1/batch/{uid}` | Get batch processing results |

See [API Documentation](docs/apiv1.md) for complete reference.

## Documentation

- [API v1 Reference](docs/apiv1.md) - Complete API documentation
- [Configuration](docs/config.md) - Configuration file options
- [Database Analysis](docs/database-analysis.md) - Database tables and operations

## Development

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -v -o ocapi ./cmd/opencart

# Run locally
go run ./cmd/opencart -conf=config.yml -log=./
```

## Deployment

The project includes GitHub Actions workflows for CI/CD:

- `deploy.yml` - Deploys to production on push to `master`
- `deploy-dev.yml` - Deploys to development on push to `develop`

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact

For questions or inquiries, contact the developer at [dev@nomadus.net](mailto:dev@nomadus.net).
