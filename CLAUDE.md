# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OCAPI (OpenCart API) is a Go-based REST API service that provides programmatic access to OpenCart e-commerce database operations. It enables external systems to manage products, categories, orders, and other OpenCart entities without direct database access.

## Build and Run Commands

### Build the application
```bash
go build -v -o ocapi ./cmd/opencart
```

### Run the application
```bash
# With default config location (config.yml in current directory)
./ocapi

# With custom config and log paths
./ocapi -conf=/etc/conf/config.yml -log=/var/log/

# Run directly with go
go run ./cmd/opencart -conf=config.yml
```

### Development
```bash
# Install dependencies
go mod download

# Verify dependencies
go mod verify

# Run with local config
go run ./cmd/opencart -conf=config.yml -log=./
```

### Testing
No test files are currently present in the project.

## Architecture

### Layered Architecture Pattern

The codebase follows a clean layered architecture with clear separation of concerns:

**Entry Point** (`cmd/opencart/main.go`)
- Bootstraps the application: config, logging, database, HTTP server
- Initializes the Core handler with dependencies
- Starts a goroutine that logs MySQL connection stats every 30 minutes

**Core Business Logic** (`impl/core/`)
- Contains all business logic for products, categories, orders, attributes
- `Core` struct is the main handler that implements all business operations
- Depends on `Repository` interface (implemented by SQL client) and optional `MessageService` interface (Telegram bot)
- Key files:
  - `core.go`: Core struct definition, Repository and MessageService interfaces
  - `products.go`: Product creation/update, image handling (base64 decoding, file operations)
  - `category.go`: Category management
  - `order.go`: Order retrieval and status updates
  - `attribute.go`: Product attribute management
  - `auth.go`: API key validation (checks both config file and database)

**Database Layer** (`internal/database/`)
- `sql-client.go`: Large file (~1500 lines) implementing the Repository interface
- Direct SQL queries using `database/sql` with MySQL driver
- Handles OpenCart database schema including table prefixes
- `statements.go`: SQL query templates
- `table-structure.go`: Database schema definitions

**HTTP Layer** (`internal/http-server/`)
- `api/api.go`: Chi router setup, route definitions, middleware chain
- `handlers/`: One subdirectory per resource (product, category, order, etc.)
  - Each handler receives logger and Core interface
  - Validates input using go-playground/validator
  - Uses chi/render for JSON responses
- `middleware/`:
  - `authenticate`: Bearer token authentication (checks config key or queries database)
  - `timeout`: 5-second request timeout

**Entity Layer** (`entity/`)
- Plain Go structs representing domain objects
- JSON and validation tags for API serialization
- Separate structs for data transfer vs database records

**Configuration** (`internal/config/`)
- YAML-based configuration using cleanenv library
- Singleton pattern with sync.Once
- Environment variables override YAML values

### Data Flow

1. HTTP request → Chi router → Middleware (auth, timeout) → Handler
2. Handler validates request → Calls Core method
3. Core applies business logic → Calls Repository method
4. Repository executes SQL → Returns data
5. Core processes result → Handler formats response → JSON to client

### Key Design Patterns

**Repository Pattern**: `Repository` interface in `impl/core/core.go` abstracts database operations

**Dependency Injection**: Core receives Repository and MessageService via setter methods, not constructor

**Handler Pattern**: Each HTTP handler is a standalone function that takes logger and Core interface

**Interface Segregation**: The main Handler interface (`internal/http-server/api/api.go`) composes multiple smaller interfaces (product.Core, category.Core, order.Core, etc.)

## Important Implementation Details

### OpenCart Database Schema
- All tables use a configurable prefix (default: `prefix_`)
- Products identified by both `product_id` (auto-increment) and `product_uid` (UUID)
- Categories have similar dual identification
- Multi-language support: separate description tables with `language_id`

### Product Image Handling
- Images received as base64-encoded strings in API requests
- Core decodes and writes to filesystem at configured `images.path`
- Database stores relative path (e.g., `catalog/product/image.png`)
- Batch finalization (`FinishBatch`) cleans up orphaned image files

### Authentication
- Two-stage: config file API key OR database lookup (`oc_api` table)
- Bearer token in Authorization header
- Implemented in `impl/core/auth.go` and `internal/http-server/middleware/authenticate`

### Error Handling
- Structured logging with slog throughout
- HTTP handlers return proper status codes and JSON error responses
- Database errors logged with context (module, operation, parameters)

### Logging
- Structured logging using Go's slog package
- Log level determined by `env` config field (local = debug)
- Custom logger setup in `internal/lib/logger/`
- Each component adds module context: `log.With(sl.Module("core"))`

## Configuration

Configuration file structure is documented in `docs/config.md`. Key sections:
- `listen`: Bind address, port, API key
- `sql`: Database connection details, table prefix
- `images`: Filesystem path and URL prefix for product images
- `telegram`: Optional Telegram bot integration (currently commented out in main.go)

Two config files exist:
- `config.yml`: Default/development configuration
- `ocapi-config.yml`: Template for CI/CD (uses ${VAR} placeholders)

## API Documentation

Full API documentation is in `docs/apiv1.md`. The API follows RESTful patterns with these key endpoints:

- `POST /api/v1/product` - Upsert product data (creates if not exists)
- `POST /api/v1/product/description` - Upsert product descriptions
- `POST /api/v1/category` - Upsert categories
- `POST /api/v1/category/description` - Upsert category descriptions
- `GET /api/v1/product/{uid}` - Get product by UUID
- `GET /api/v1/order/{orderId}` - Get order with full details
- `GET /api/v1/orders/{orderStatusId}` - Get order IDs by status
- `POST /api/v1/order` - Update order status
- `GET /api/v1/batch/{batchUid}` - Get batch processing results

All requests require Bearer token authentication. Responses use consistent JSON format with `success`, `status_message`, `timestamp`, and `data` fields.

## Deployment

GitHub Actions workflows handle CI/CD:
- `deploy.yml`: Deploys to production on push to master branch
- `deploy-dev.yml`: Deploys to dev server on push to develop branch
- Build process: prepares config file (substitutes secrets), builds Go binary, SCPs to server, restarts systemd service

The application is designed to run as a systemd service (`ocapi.service`) on the target server.
