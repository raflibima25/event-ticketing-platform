# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Event Ticketing Platform - A microservices-based ticket sales platform supporting multiple payment channels via Xendit integration.

**Tech Stack:**
- Backend: Golang with gRPC/REST APIs
- Frontend: Next.js (React)
- Infrastructure: Google Kubernetes Engine (GKE)
- Database: PostgreSQL (Cloud SQL)
- Cache: Redis (Cloud Memorystore)
- Message Queue: Cloud Pub/Sub
- Payment Gateway: Xendit
- Storage: Google Cloud Storage
- Email: SendGrid

## Development Commands

### Backend (Golang Services)

```bash
# Run tests
go test ./... -cover -race

# Run tests for specific service
go test ./services/auth/... -v

# Run with coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Linting
golangci-lint run

# Run locally with hot reload (using Air)
air

# Database migrations
migrate -path db/migrations -database "${DATABASE_URL}" up
migrate -path db/migrations -database "${DATABASE_URL}" down 1
migrate create -ext sql -dir db/migrations -seq migration_name
```

### Frontend (Next.js)

```bash
# Development server
npm run dev

# Build
npm run build

# Production mode
npm run start

# Linting
npm run lint

# Type checking
npm run type-check
```

### Docker & Kubernetes

```bash
# Build Docker image
docker build -t service-name:tag -f services/service-name/Dockerfile .

# Local development with Docker Compose
docker-compose up -d
docker-compose down

# Kubernetes deployment
kubectl apply -f k8s/
kubectl get pods
kubectl logs -f pod-name
kubectl describe pod pod-name

# Port forwarding for local testing
kubectl port-forward service/api-gateway 8080:8080
```

### Testing

```bash
# Integration tests (with testcontainers)
go test ./tests/integration/... -v

# Load testing with K6
k6 run tests/load/ticket-purchase.js

# Security scanning
trivy image service-name:tag
```

## Architecture Overview

### Microservices Structure

```
services/
├── auth-service/          # User registration, login, JWT management
├── event-service/         # Event CRUD, ticket tiers, search/filtering
├── ticketing-service/     # Ticket reservation, inventory locking, QR generation
├── payment-service/       # Xendit integration, webhook handling, refunds
├── notification-service/  # Email via SendGrid, Pub/Sub consumers
└── api-gateway/          # Request routing, auth middleware, rate limiting
```

Each service is independently deployable with its own:
- Database schema (shared PostgreSQL with service-specific tables)
- gRPC/REST endpoints
- Health checks (liveness/readiness probes)
- HPA configuration

### Critical Business Logic

#### Ticket Reservation & Inventory Management

**Key Challenge:** Prevent overselling while handling high concurrency

**Implementation Requirements:**
- Use PostgreSQL row-level locking: `SELECT ... FOR UPDATE` when reserving tickets
- Database constraint: `CHECK (sold_count <= quota)` on ticket_tiers table
- 15-minute reservation timeout from initial cart, extended to 30 minutes after payment method selection
- Background job to release expired reservations every 5 minutes
- Redis distributed locks for cross-service coordination

**File locations (when implemented):**
- `services/ticketing-service/internal/service/reservation.go` - Core reservation logic
- `services/ticketing-service/internal/repository/ticket_tier.go` - Database operations with locking

#### Payment Processing & Transaction Consistency

**Critical Flow:** Payment success must atomically trigger multiple operations

**Saga Pattern Implementation:**
1. Webhook receives `invoice.paid` event from Xendit
2. Verify webhook signature (security requirement)
3. Check idempotency: `webhook_events` table with unique constraint on `webhook_id`
4. Update order status to 'paid'
5. Reduce ticket quota (`UPDATE ticket_tiers SET sold_count = sold_count + X`)
6. Publish to Pub/Sub: `payment.success` event
7. Notification service consumes event → generates e-ticket → sends email

**Compensation Logic:**
- If ticket generation fails after payment: Retry 5x with exponential backoff
- If still failing: Log to dead letter queue, create manual admin task, notify customer of 24h delay
- If email fails: Queue for retry (max 10 attempts)
- Refund processing failure: Rollback order status, restore quota

**File locations (when implemented):**
- `services/payment-service/internal/handler/webhook.go` - Xendit webhook handling
- `services/payment-service/internal/service/idempotency.go` - Duplicate prevention
- `services/ticketing-service/internal/service/ticket_generator.go` - E-ticket generation with retry

#### Database Transaction Isolation

- Use `READ COMMITTED` isolation level (PostgreSQL default)
- Critical sections use explicit transactions with `SELECT FOR UPDATE`
- Avoid long-running transactions (max 5 seconds)

### Database Schema Highlights

**Critical Constraints:**
- `ticket_tiers.sold_count <= ticket_tiers.quota` - Prevents overselling at DB level
- `orders.reservation_expires_at` - Indexed for efficient cleanup queries
- `webhook_events.webhook_id UNIQUE` - Idempotency guarantee
- `events.start_date < end_date` - Data integrity

**Important Indexes:**
- `idx_orders_reservation ON orders(reservation_expires_at) WHERE status = 'reserved'` - For timeout cleanup
- `idx_events_search USING gin(to_tsvector('indonesian', title || ' ' || description))` - Full-text search
- `idx_ticket_tiers_quota ON ticket_tiers(event_id, sold_count, quota)` - Availability queries

### Xendit Payment Integration

**Payment Methods Supported:**
- QRIS (unified QR code)
- Virtual Account (BCA, Mandiri, BNI, BRI)
- E-Wallet (OVO, DANA, GoPay, LinkAja)

**Invoice Creation:**
- Expiry: 1800 seconds (30 minutes) from checkout
- External ID format: `ORDER-{order_id}`
- Callback URL must verify signature using Xendit callback token

**Webhook Events to Handle:**
- `invoice.paid` → Generate tickets, send email
- `invoice.expired` → Release reservation
- `disbursement.completed` → Refund confirmation to customer

**File locations (when implemented):**
- `services/payment-service/pkg/xendit/client.go` - Xendit API wrapper
- `services/payment-service/internal/handler/webhook.go` - Webhook signature verification

### Financial Logic

**Platform Fee Calculation:**
- Customer pays: `ticket_price + (ticket_price * 5%) + Rp 2,500`
- Event Organizer receives (T+7 after event): `ticket_price - (ticket_price * 10%)`
- Platform revenue: `(ticket_price * 10%) + Rp 2,500`

**Refund Policy:**
- Customer can request: Max 7 days before event
- Refund amount: `ticket_price - (ticket_price * 10%)` (customer loses platform fee)
- EO auto-refund if event cancelled: 100% refund to customer, EO pays 5% penalty

### Time Zone Handling

**Critical Requirement:**
- Events stored with `TIMESTAMPTZ` in UTC + timezone field (e.g., "Asia/Jakarta")
- Frontend displays event time in event's local timezone
- User's browser timezone used for "time until event" calculations
- Email notifications show both local time and UTC offset

**Implementation:**
- Database: `events.start_date TIMESTAMPTZ, events.timezone VARCHAR(50)`
- API responses include ISO8601 with timezone: `2025-12-24T19:00:00+07:00`
- Frontend uses `date-fns-tz` or `luxon` for timezone conversion

## Security Requirements

### Authentication & Authorization

- JWT tokens expire in 24 hours, refresh tokens in 7 days
- Password hashing: bcrypt with cost factor 12
- OAuth support for Google Sign-In
- RBAC: customer, organizer, admin roles
- Middleware validates JWT on all protected endpoints

### Critical Security Controls

**Rate Limiting:**
- Read endpoints: 100 req/min per IP
- Write endpoints: 20 req/min per IP
- Payment webhook: No rate limit (but verify signature)

**Input Validation:**
- All user inputs sanitized to prevent XSS
- Parameterized queries only (prevent SQL injection)
- File uploads: Validate MIME type, max 5MB for images
- QR code validation hash to prevent forgery

**Data Protection:**
- PII masking in logs: `email: u***@example.com, phone: 08****123`
- Secrets in GCP Secret Manager, never in code/env files committed
- Audit logging for: payments, refunds, ticket validation, user deletion
- HTTPS only (TLS 1.3), CORS whitelist configured

**Webhook Security:**
- Always verify Xendit webhook signature before processing
- Implement replay attack protection (timestamp validation)
- Log all webhook payloads for audit

## Infrastructure & Deployment

### GKE Configuration

**Production Cluster:**
- Multi-zone deployment (3 zones minimum)
- Node pools with autoscaling (2-10 nodes)
- Pod disruption budgets: `minAvailable: 1` for critical services
- Resource limits: CPU request 100m, limit 500m; Memory request 128Mi, limit 512Mi

**HPA Settings:**
- Target CPU utilization: 70%
- Min replicas: 2, Max replicas: 10
- Scale-down stabilization: 5 minutes

### Database

**Cloud SQL Configuration:**
- Instance: db-n1-standard-2 (production), db-f1-micro (staging)
- High availability: Failover replica in different zone
- Automated backups: Daily, 30-day retention
- Point-in-time recovery: 7 days

**Connection Management:**
- Connection pooling: Max 20 connections per service instance
- Use Cloud SQL Proxy for secure connections
- Connection timeout: 30 seconds

### Redis (Cloud Memorystore)

**Usage:**
- Distributed locks for ticket reservation (TTL: 15 minutes)
- Session cache (TTL: 24 hours)
- Event listing cache (TTL: 5 minutes)
- Available quota cache (TTL: 30 seconds)

**Cache Invalidation:**
- Ticket purchase → invalidate event quota cache
- Event update → invalidate event detail + listing cache
- Refund → invalidate order cache + quota cache

### Cloud Pub/Sub Topics

```
payment.success        → Notification Service (generate e-ticket, send email)
ticket.created         → Analytics Service (update stats)
refund.processed       → Notification Service (send refund confirmation)
event.cancelled        → Payment Service (trigger auto-refunds)
```

**Consumer Configuration:**
- Acknowledgement deadline: 60 seconds
- Max retry: 5 attempts with exponential backoff
- Dead letter queue after max retries

## Monitoring & Observability

### Critical Metrics to Track

**Application Metrics:**
- Payment success rate (target: >95%)
- API response time P95 (target: <500ms)
- Ticket overselling incidents (target: 0)
- Order abandonment rate (checkout started but not completed)

**Infrastructure Metrics:**
- Pod CPU/Memory usage (alert threshold: >80%)
- Database connection pool utilization (alert: >80%)
- Redis hit rate (target: >90%)
- Pub/Sub message processing lag (alert: >5 minutes)

### Logging Standards

**Structured Logging (JSON format):**
```json
{
  "timestamp": "2025-12-24T10:00:00Z",
  "level": "info",
  "service": "payment-service",
  "correlation_id": "uuid",
  "user_id": "uuid",
  "message": "Payment processed successfully",
  "metadata": {
    "order_id": "uuid",
    "amount": 150000,
    "payment_method": "QRIS"
  }
}
```

**Correlation ID:**
- Generate UUID at API Gateway, propagate through all services via gRPC metadata
- Include in all logs for request tracing
- Return in API response header: `X-Correlation-ID`

### Alerting Thresholds

- Error rate >1% sustained for 5 minutes
- P95 latency >1 second for 5 minutes
- Payment webhook processing lag >5 minutes
- Database CPU >80% for 10 minutes
- Any overselling incident (immediate alert)

## Development Workflow

### Branching Strategy

```
main              → Production-ready code
staging           → Deployed to staging environment
feature/*         → New features
bugfix/*          → Bug fixes
hotfix/*          → Emergency production fixes
```

### CI/CD Pipeline

**On Pull Request:**
1. Lint (golangci-lint)
2. Unit tests (all services in parallel)
3. Integration tests
4. Security scan (Trivy)
5. Build verification

**On Merge to Staging:**
1. Build Docker images
2. Push to Artifact Registry
3. Deploy to staging cluster
4. Run E2E tests
5. Smoke tests

**On Merge to Main (Production):**
1. Manual approval required
2. Canary deployment (10% → 50% → 100%)
3. Automated rollback on health check failure
4. Post-deployment verification

### Code Quality Standards

- Unit test coverage: Minimum 80% for business logic
- No critical or high security vulnerabilities
- All exported functions must have Go documentation
- Database migrations must be reversible (up/down)
- API changes require OpenAPI spec update

## Common Pitfalls to Avoid

1. **Never bypass row-level locking for ticket reservation** - This will cause overselling
2. **Always verify Xendit webhook signatures** - Prevent fraudulent payment confirmations
3. **Don't ignore idempotency** - Duplicate webhook processing will cause double ticket generation
4. **Never store Xendit API keys in code** - Use Secret Manager
5. **Don't forget timezone conversion** - Event times must display correctly for all users
6. **Always set reservation timeouts** - Prevent indefinite inventory locks
7. **Test payment flows in Xendit sandbox** - Before any production deployment
8. **Implement graceful shutdown** - Prevent mid-transaction failures during deployments
9. **Use database transactions for multi-step operations** - Especially payment confirmation flow
10. **Monitor dead letter queues** - Failed async operations need manual intervention

## Project Status

This is a greenfield project currently in specification phase. The SRS.md document contains complete requirements. When implementing:

1. Start with Phase 1 (Foundation): Auth Service + API Gateway
2. Implement database schema exactly as specified in SRS.md sections 5.2
3. Follow the 12-week development timeline in SRS.md section 11
4. Ensure all non-functional requirements (NFR-*) are met
5. Reference SRS.md for detailed functional requirements (FR-*)
