# Event Ticketing Platform - MVP

Event Ticketing Platform dengan microservices architecture menggunakan Golang (backend) dan Next.js (frontend).

## 📋 Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local development)
- Node.js 20+ (for local development)
- Xendit Sandbox Account (for payment testing)

## 🚀 Quick Start

### 1. Clone dan Setup Environment

```bash
# Copy environment variables
cp .env.example .env

# Edit .env file:
# - Database credentials (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)
# - Redis configuration (REDIS_HOST, REDIS_PORT, REDIS_PASSWORD, REDIS_DB)
# - JWT secret key (JWT_SECRET)
# - Xendit API keys (dapatkan dari: https://dashboard.xendit.co/settings/developers)
```

**Environment Variables Structure:**

```env
# Database - Separate variables for flexibility
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=ticketing_platform
DB_SSL_MODE=disable

# Redis - Separate variables
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRY=24h

# Xendit (Payment Gateway)
XENDIT_API_KEY=your-api-key
XENDIT_WEBHOOK_TOKEN=your-webhook-token
```

### 2. Start All Services dengan Docker Compose

```bash
# Build dan start semua services
docker-compose up --build

# Atau run di background
docker-compose up -d --build

# Lihat logs
docker-compose logs -f

# Stop semua services
docker-compose down

# Dev
docker-compose -f docker-compose.dev.yml logs -f
```

### 3. Verify Services

Setelah semua container running, verify dengan mengakses:

- **Frontend**: http://localhost:3000
- **API Gateway**: http://localhost:8080/health
- **Auth Service**: http://localhost:8081/health
- **Event Service**: http://localhost:8082/health
- **Ticketing Service**: http://localhost:8083/health
- **Payment Service**: http://localhost:8084/health
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

## 🏗️ Architecture

```
┌─────────────┐
│   Frontend  │ (Next.js - Port 3000)
│  (React)    │
└──────┬──────┘
       │
       │ HTTP
       ▼
┌─────────────┐
│ API Gateway │ (Port 8080)
└──────┬──────┘
       │
       ├─────────┬─────────┬──────────┬────────────┐
       │         │         │          │            │
       ▼         ▼         ▼          ▼            ▼
   ┌──────┐ ┌───────┐ ┌─────────┐ ┌────────┐ ┌─────────┐
   │ Auth │ │ Event │ │Ticketing│ │Payment │ │  Redis  │
   │ :8081│ │ :8082 │ │  :8083  │ │ :8084  │ │  :6379  │
   └───┬──┘ └───┬───┘ └────┬────┘ └───┬────┘ └─────────┘
       │        │          │          │
       └────────┴──────────┴──────────┘
                        │
                        ▼
                 ┌─────────────┐
                 │ PostgreSQL  │
                 │   :5432     │
                 └─────────────┘
```

## 📁 Project Structure

```
event-ticketing-platform/
├── backend/
│   ├── services/
│   │   ├── auth/          # Authentication & User Management
│   │   ├── event/         # Event CRUD & Ticket Tiers
│   │   ├── ticketing/     # Reservation & Inventory
│   │   ├── payment/       # Xendit Integration
│   │   └── gateway/       # API Gateway & Routing
│   ├── shared/            # Shared packages (database, middleware)
│   └── migrations/        # Database migrations
├── frontend/              # Next.js application
├── proto/                 # Protocol Buffer definitions (future)
├── pkg/                   # Shared libraries
├── docker-compose.yml
├── .env.example
├── SRS.md                 # Software Requirements Specification
└── CLAUDE.md             # Development Guide for Claude Code
```

## 🛠️ Development

### Backend Development

```bash
# Install Go dependencies
go mod download

# Run specific service locally (example: auth service)
cd backend/services/auth
PORT=8081 DATABASE_URL="postgres://postgres:postgres@localhost:5432/ticketing_platform?sslmode=disable" go run main.go

# Run tests
go test ./...

# Run with coverage
go test ./... -cover
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Run linter
npm run lint
```

### Database Migrations

```bash
# Run migrations manually
docker exec ticketing-migrate migrate -path=/migrations -database="postgres://postgres:postgres@postgres:5432/ticketing_platform?sslmode=disable" up

# Rollback last migration
docker exec ticketing-migrate migrate -path=/migrations -database="postgres://postgres:postgres@postgres:5432/ticketing_platform?sslmode=disable" down 1
```

## 🧪 Testing

### Test API Endpoints

```bash
# Health check semua services
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
curl http://localhost:8084/health

# Test register (placeholder)
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","full_name":"Test User"}'

# Test list events (placeholder)
curl http://localhost:8080/api/v1/events
```

## 📚 API Documentation

### Authentication Endpoints

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh-token` - Refresh JWT token

### Event Endpoints

- `GET /api/v1/events` - List events (with filters)
- `GET /api/v1/events/:id` - Get event details
- `POST /api/v1/events` - Create event (Organizer)
- `PUT /api/v1/events/:id` - Update event (Organizer)
- `GET /api/v1/events/:id/tickets` - Get ticket tiers

### Ticketing Endpoints

- `POST /api/v1/orders/reserve` - Reserve tickets
- `POST /api/v1/orders/:id/checkout` - Checkout
- `GET /api/v1/orders/my-orders` - Get user orders
- `GET /api/v1/tickets/:id/download` - Download e-ticket

### Payment Endpoints

- `POST /api/v1/payments/create-invoice` - Create Xendit invoice
- `GET /api/v1/payments/:id/status` - Get payment status
- `POST /api/v1/payments/webhook` - Xendit webhook callback

## 🔒 Security Notes

- Jangan commit file `.env` ke repository
- Gunakan strong JWT secret di production
- Xendit API keys harus disimpan dengan aman
- Enable HTTPS di production
- Implement rate limiting di production

## 📝 Current Status

✅ **Phase 1 COMPLETED**: Foundation & Infrastructure Setup

- [x] Monorepo structure
- [x] Docker Compose configuration
- [x] Database migrations
- [x] Minimal service scaffolding
- [x] Frontend basic setup

🚧 **Next Steps**:

- [ ] Implement Auth Service (registration, login, JWT)
- [ ] Implement Event Service (CRUD, search)
- [ ] Implement Ticketing Service (reservation with locking)
- [ ] Implement Payment Service (Xendit integration)
- [ ] Build Frontend UI components

See [CLAUDE.md](./CLAUDE.md) for detailed development guide and [SRS.md](./SRS.md) for complete requirements.

## 🤝 Contributing

This is an MVP project. Refer to CLAUDE.md for development guidelines and architecture decisions.

## 📄 License

This project is for educational/development purposes.
