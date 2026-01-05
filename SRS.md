# Software Requirements Specification (SRS)
## Event Ticketing Platform

**Version:** 1.0  
**Date:** December 24, 2025  
**Author:** Bima

---

## 1. Introduction

### 1.1 Purpose
Dokumen ini menjelaskan spesifikasi lengkap untuk Event Ticketing Platform, sebuah sistem penjualan tiket event online yang mendukung berbagai jenis event dengan multiple payment channels.

### 1.2 Scope
Platform ini memungkinkan:
- Event organizers membuat dan mengelola event
- Customers membeli tiket dengan berbagai metode pembayaran
- Automated ticket generation dan distribution
- Real-time inventory management
- Payment processing melalui Xendit

### 1.3 Definitions and Acronyms
- **EO**: Event Organizer
- **VA**: Virtual Account
- **QRIS**: Quick Response Code Indonesian Standard
- **HPA**: Horizontal Pod Autoscaler
- **RBAC**: Role-Based Access Control

---

## 2. Overall Description

### 2.1 Product Perspective
Sistem ini terdiri dari:
- **Frontend**: Web application (Next.js)
- **Backend**: Microservices architecture (Golang)
- **Infrastructure**: Google Kubernetes Engine (GKE)
- **Database**: PostgreSQL on Cloud SQL
- **Cache & Session**: Redis (Cloud Memorystore)
- **Message Queue**: Cloud Pub/Sub
- **Payment Gateway**: Xendit
- **Storage**: Google Cloud Storage
- **Email Service**: SendGrid

### 2.2 User Classes and Characteristics

#### 2.2.1 Guest User
- Browse events
- View event details
- No authentication required

#### 2.2.2 Customer
- Register/Login
- Purchase tickets
- View purchase history
- Download e-tickets
- Request refunds

#### 2.2.3 Event Organizer
- Create and manage events
- Set ticket tiers and pricing
- View sales analytics
- Manage refunds
- Export attendee lists

#### 2.2.4 Admin
- Manage users
- Platform analytics
- Handle disputes
- System configuration

### 2.3 Operating Environment
- **Client**: Web browsers (Chrome, Firefox, Safari, Edge)
- **Server**: GKE clusters on GCP
- **Database**: PostgreSQL 15+
- **Payment**: Xendit API v2

---

## 3. System Features

### 3.1 User Management

#### 3.1.1 Registration & Authentication
**Priority:** High  
**Description:** User dapat mendaftar dan login menggunakan email/password atau social login.

**Functional Requirements:**
- FR-UM-001: System harus support registrasi dengan email dan password
- FR-UM-002: System harus validasi email dengan OTP
- FR-UM-003: System harus support Google OAuth
- FR-UM-004: System harus implementasi JWT untuk session management
- FR-UM-005: Password harus di-hash menggunakan bcrypt
- FR-UM-006: System harus support forgot password flow

#### 3.1.2 Profile Management
**Priority:** Medium  
**Functional Requirements:**
- FR-UM-007: User dapat update profile (nama, phone, foto)
- FR-UM-008: User dapat change password
- FR-UM-009: User dapat delete account

---

### 3.2 Event Management

#### 3.2.1 Create Event
**Priority:** High  
**Description:** Event Organizer dapat membuat event baru.

**Functional Requirements:**
- FR-EM-001: EO dapat create event dengan informasi: title, description, date, time, location, category, timezone
- FR-EM-002: EO dapat upload event banner/poster (max 5MB, JPG/PNG)
- FR-EM-003: EO dapat set event status: Draft, Published, Cancelled
- FR-EM-004: Event banner harus disimpan di Cloud Storage
- FR-EM-005: System harus generate unique event slug/URL
- FR-EM-006: System harus store event datetime dalam UTC dan timezone information (Asia/Jakarta, etc)
- FR-EM-007: Frontend harus display event time dalam timezone event location

#### 3.2.2 Ticket Tier Management
**Priority:** High
**Functional Requirements:**
- FR-EM-008: EO dapat create multiple ticket tiers (VIP, Regular, Early Bird, dll)
- FR-EM-009: Setiap tier memiliki: name, price, quota, description
- FR-EM-010: EO dapat set early bird pricing dengan start/end date
- FR-EM-011: EO dapat set maximum tickets per transaction
- FR-EM-012: System harus validasi quota tidak exceed capacity

#### 3.2.3 Event Discovery
**Priority:** High
**Functional Requirements:**
- FR-EM-013: User dapat browse events dengan pagination
- FR-EM-014: User dapat filter events by: category, date range, location, price range
- FR-EM-015: User dapat search events by keyword
- FR-EM-016: System harus show upcoming events di homepage
- FR-EM-017: Event detail page harus show: info lengkap, available tickets, terms & conditions
- FR-EM-018: Event time harus di-display sesuai timezone user (with timezone label)

---

### 3.3 Ticketing System

#### 3.3.1 Ticket Purchase Flow
**Priority:** Critical  
**Description:** Core flow pembelian tiket dengan inventory management.

**Functional Requirements:**
- FR-TS-001: User dapat select ticket tier dan quantity
- FR-TS-002: System harus implementasi seat/ticket reservation dengan timeout (15 menit)
- FR-TS-003: System harus handle race condition dengan row-level locking (SELECT FOR UPDATE)
- FR-TS-004: User harus input attendee information (nama, email per ticket)
- FR-TS-005: System harus calculate total price (base price + platform fee)
- FR-TS-006: User dapat apply promo code jika tersedia
- FR-TS-007: System harus release reserved tickets jika payment timeout
- FR-TS-008: System harus extend reservation jika payment method dipilih (total timeout 30 menit dari initial reservation)

#### 3.3.2 Inventory Management
**Priority:** Critical
**Functional Requirements:**
- FR-TS-009: System harus track available quota secara real-time
- FR-TS-010: System harus prevent overselling dengan database constraints (CHECK: sold_count <= quota)
- FR-TS-011: System harus show "Sold Out" badge jika quota habis
- FR-TS-012: System harus update quota setelah successful payment
- FR-TS-013: System harus restore quota jika refund approved

#### 3.3.3 E-Ticket Generation
**Priority:** High
**Functional Requirements:**
- FR-TS-014: System harus generate unique QR code per ticket
- FR-TS-015: QR code harus contain: ticket_id, event_id, attendee_name, validation_hash
- FR-TS-016: System harus generate PDF e-ticket
- FR-TS-017: E-ticket PDF harus include: event details, QR code, terms
- FR-TS-018: E-ticket harus dikirim via email setelah payment success
- FR-TS-019: User dapat download e-ticket dari dashboard
- FR-TS-020: E-ticket PDF harus disimpan di Cloud Storage untuk re-download

---

### 3.4 Payment System

#### 3.4.1 Payment Processing
**Priority:** Critical  
**Description:** Integration dengan Xendit untuk multiple payment channels.

**Functional Requirements:**
- FR-PS-001: System harus support payment methods: QRIS, Virtual Account (BCA, Mandiri, BNI, BRI), E-Wallet (OVO, DANA, GoPay, LinkAja)
- FR-PS-002: User dapat pilih payment method saat checkout
- FR-PS-003: System harus create Xendit invoice dengan expiry time (30 menit dari reservation)
- FR-PS-004: System harus display payment instructions setelah checkout
- FR-PS-005: System harus handle Xendit webhook untuk payment status updates
- FR-PS-006: Payment status: Pending, Paid, Expired, Failed
- FR-PS-007: System harus validate payment completion sebelum reservation timeout

#### 3.4.2 Payment Webhook Handling
**Priority:** Critical
**Functional Requirements:**
- FR-PS-008: System harus verify Xendit webhook signature
- FR-PS-009: System harus implementasi idempotency dengan webhook_events table (unique constraint pada webhook_id)
- FR-PS-010: System harus update order status setelah payment success
- FR-PS-011: System harus trigger e-ticket generation setelah payment success
- FR-PS-012: System harus retry failed webhook processing (max 3x dengan exponential backoff)
- FR-PS-013: System harus log semua webhook events ke database dan Cloud Logging
- FR-PS-014: System harus handle duplicate webhook callbacks dengan idempotent response

#### 3.4.3 Refund Processing
**Priority:** High
**Functional Requirements:**
- FR-PS-015: Customer dapat request refund (max 7 hari sebelum event)
- FR-PS-016: EO dapat approve/reject refund request dengan reason
- FR-PS-017: System harus process refund via Xendit disbursement
- FR-PS-018: Refund amount = ticket price - platform fee (10%)
- FR-PS-019: System harus update order status setelah refund processed
- FR-PS-020: System harus notify customer via email setelah refund success/rejected
- FR-PS-021: System harus handle automatic refund jika event dibatalkan oleh EO

#### 3.4.4 Transaction Consistency & Compensation
**Priority:** Critical
**Description:** Saga pattern untuk distributed transaction management
**Functional Requirements:**
- FR-PS-022: Payment success HARUS atomic: update order status → reduce quota → trigger ticket generation
- FR-PS-023: Jika ticket generation gagal setelah payment success, system harus:
  - Retry ticket generation (max 5x dengan exponential backoff)
  - Jika masih gagal, log error dan create manual ticket generation task untuk admin
  - Notify customer bahwa e-ticket akan dikirim dalam 24 jam
- FR-PS-024: Jika email sending gagal, system harus queue email untuk retry (retry max 10x)
- FR-PS-025: Compensation transaction: Jika refund processing gagal, rollback order status dan restore quota
- FR-PS-026: Database transactions harus menggunakan isolation level READ COMMITTED
- FR-PS-027: Idempotency keys untuk semua critical operations (payment, refund, payout)
- FR-PS-028: Dead letter queue untuk failed async operations (Pub/Sub)

---

### 3.5 Financial & Business Logic

#### 3.5.1 Platform Fee & Commission
**Priority:** High
**Functional Requirements:**
- FR-FN-001: Platform fee = 5% dari ticket price + Rp 2.500 per ticket (biaya admin)
- FR-FN-002: Platform fee harus ditampilkan secara transparan saat checkout
- FR-FN-003: Total amount yang dibayar customer = ticket price + platform fee
- FR-FN-004: EO menerima = ticket price - platform commission (10% dari ticket price)
- FR-FN-005: System harus calculate dan store platform revenue per transaction

#### 3.5.2 Payout Management
**Priority:** High
**Functional Requirements:**
- FR-FN-006: Payout ke EO dilakukan T+7 setelah event selesai
- FR-FN-007: EO harus input bank account information untuk payout
- FR-FN-008: System harus validate bank account menggunakan Xendit account validation
- FR-FN-009: Payout otomatis trigger 7 hari setelah event end_date
- FR-FN-010: Payout amount = total ticket sales - platform commission - refunded amount
- FR-FN-011: EO dapat view payout history dan status (pending, processing, completed, failed)
- FR-FN-012: System harus kirim notification ke EO saat payout completed

#### 3.5.3 Event Cancellation Policy
**Priority:** Medium
**Functional Requirements:**
- FR-FN-013: EO dapat cancel event minimal 14 hari sebelum event date
- FR-FN-014: Jika EO cancel event, automatic full refund ke semua customers (100%, no admin fee)
- FR-FN-015: EO dikenakan penalty fee 5% dari total sales jika cancel event
- FR-FN-016: System harus notify semua ticket holders via email jika event cancelled
- FR-FN-017: Admin dapat force cancel event dengan approval

---

### 3.6 Notification System

#### 3.6.1 Email Notifications
**Priority:** High
**Description:** Email service menggunakan SendGrid API
**Functional Requirements:**
- FR-NS-001: System harus kirim email verification saat registrasi
- FR-NS-002: System harus kirim order confirmation setelah checkout
- FR-NS-003: System harus kirim e-ticket setelah payment success
- FR-NS-004: System harus kirim payment reminder jika pending (15 menit sebelum expire)
- FR-NS-005: System harus kirim event reminder (H-1 sebelum event)
- FR-NS-006: System harus kirim refund confirmation
- FR-NS-007: Email template harus responsive dan mobile-friendly
- FR-NS-008: System harus track email delivery status (sent, delivered, bounced, opened)

#### 3.6.2 Push Notifications (Future Enhancement)
**Priority:** Low
**Functional Requirements:**
- FR-NS-009: System dapat kirim push notification untuk promotional events

---

### 3.7 Analytics & Reporting

#### 3.7.1 EO Dashboard
**Priority:** Medium
**Functional Requirements:**
- FR-AR-001: EO dapat view total sales per event
- FR-AR-002: EO dapat view sold tickets breakdown by tier
- FR-AR-003: EO dapat view revenue chart (daily/weekly/monthly)
- FR-AR-004: EO dapat export attendee list (CSV/Excel) dengan columns: name, email, ticket tier, order date
- FR-AR-005: EO dapat view payment status distribution
- FR-AR-006: EO dapat view real-time ticket sales dashboard

#### 3.7.2 Admin Dashboard
**Priority:** Medium
**Functional Requirements:**
- FR-AR-007: Admin dapat view platform-wide statistics
- FR-AR-008: Admin dapat view total transactions dan revenue
- FR-AR-009: Admin dapat view top selling events
- FR-AR-010: Admin dapat view payment method distribution
- FR-AR-011: Admin dapat export financial reports untuk accounting

---

## 4. Non-Functional Requirements

### 4.1 Performance
- NFR-P-001: API response time harus < 500ms untuk 95% requests
- NFR-P-002: System harus handle 1000 concurrent ticket purchases
- NFR-P-003: Database query time harus < 100ms untuk read operations
- NFR-P-004: Page load time harus < 3 seconds

### 4.2 Scalability
- NFR-S-001: System harus support horizontal scaling dengan Kubernetes HPA
- NFR-S-002: Ticketing service harus auto-scale berdasarkan CPU usage (target: 70%)
- NFR-S-003: System harus handle traffic spike saat flash sale (10x normal load)

### 4.3 Security
- NFR-SEC-001: All API endpoints harus menggunakan HTTPS (TLS 1.3)
- NFR-SEC-002: Sensitive data (payment info, personal data) harus encrypted at rest menggunakan AES-256
- NFR-SEC-003: System harus implementasi rate limiting: 100 req/min untuk read endpoints, 20 req/min untuk write endpoints
- NFR-SEC-004: JWT tokens harus expire dalam 24 jam dengan refresh token mechanism (7 hari)
- NFR-SEC-005: Webhook endpoints harus verify signature (Xendit callback token verification)
- NFR-SEC-006: System harus log semua payment transactions dengan audit trail
- NFR-SEC-007: Database credentials harus disimpan di GCP Secret Manager
- NFR-SEC-008: System harus implementasi CORS policy dengan whitelist domains
- NFR-SEC-009: Frontend harus implement XSS protection (Content Security Policy headers)
- NFR-SEC-010: All forms harus protected dengan CSRF tokens
- NFR-SEC-011: API communication antar microservices harus menggunakan mTLS atau service mesh (Istio)
- NFR-SEC-012: Audit logging untuk critical operations: payment, refund, ticket validation, user deletion
- NFR-SEC-013: PII data harus di-mask di logs (email: u***@example.com, phone: 08****123)
- NFR-SEC-014: SQL injection protection dengan prepared statements/parameterized queries
- NFR-SEC-015: Password policy: minimum 8 karakter, kombinasi huruf besar/kecil, angka, simbol

### 4.4 Reliability
- NFR-R-001: System uptime harus 99.5% (measured monthly)
- NFR-R-002: Database harus memiliki automated backups (daily) dengan retention 30 hari
- NFR-R-003: System harus implement health checks (liveness & readiness probes) untuk semua services
- NFR-R-004: Failed payments harus di-retry dengan exponential backoff (max 3 attempts)
- NFR-R-005: Database point-in-time recovery capability hingga 7 hari ke belakang
- NFR-R-006: Backup restoration testing dilakukan quarterly
- NFR-R-007: GKE cluster harus multi-zone deployment untuk high availability
- NFR-R-008: Cloud SQL harus memiliki failover replica di zone berbeda
- NFR-R-009: Circuit breaker pattern untuk external API calls (Xendit, SendGrid)
- NFR-R-010: System harus graceful shutdown untuk rolling updates

### 4.5 Maintainability
- NFR-M-001: Code harus follow Go best practices dan standards (golangci-lint)
- NFR-M-002: API harus memiliki OpenAPI/Swagger documentation (auto-generated)
- NFR-M-003: All services harus memiliki structured logging (JSON format dengan correlation ID)
- NFR-M-004: System harus memiliki comprehensive monitoring (Prometheus + Grafana)
- NFR-M-005: APM tools untuk distributed tracing (Cloud Trace atau Jaeger)
- NFR-M-006: Error tracking dengan Sentry atau Cloud Error Reporting
- NFR-M-007: Alert thresholds: Error rate > 1%, P95 latency > 1s, CPU > 80%, Memory > 85%
- NFR-M-008: Unit test coverage minimum 80% untuk business logic
- NFR-M-009: Database migration menggunakan golang-migrate dengan versioning

### 4.6 Usability
- NFR-U-001: UI harus responsive (mobile, tablet, desktop)
- NFR-U-002: Payment flow harus maksimal 3 steps
- NFR-U-003: Error messages harus user-friendly dan actionable
- NFR-U-004: Accessibility compliance (WCAG 2.1 Level AA) untuk screen readers
- NFR-U-005: Support untuk bahasa Indonesia dan English (i18n ready)

### 4.7 Data Privacy & Compliance
- NFR-DP-001: User dapat request data export (semua data personal dalam format JSON/CSV)
- NFR-DP-002: User dapat delete account (right to be forgotten) - soft delete dengan anonymization
- NFR-DP-003: Data retention policy: Transaction data disimpan 7 tahun (compliance pajak), logs 90 hari
- NFR-DP-004: Privacy policy dan Terms of Service harus di-accept saat registration
- NFR-DP-005: Email opt-out mechanism untuk marketing emails (unsubscribe link)
- NFR-DP-006: Cookie consent banner untuk compliance (GDPR-ready)
- NFR-DP-007: PCI DSS compliance checklist (meskipun Xendit yang handle payment)

### 4.8 Caching Strategy
- NFR-CACHE-001: Redis cache untuk event listings (TTL: 5 menit)
- NFR-CACHE-002: Redis cache untuk available ticket quota (TTL: 30 detik)
- NFR-CACHE-003: CDN caching untuk static assets (event images, CSS, JS)
- NFR-CACHE-004: API response caching untuk read-heavy endpoints (GET /events)
- NFR-CACHE-005: Cache invalidation strategy untuk real-time data updates

---

## 5. System Architecture

### 5.1 Microservices

#### 5.1.1 Auth Service
**Responsibilities:**
- User registration & login
- JWT token generation & validation
- Password reset
- OAuth integration

**Tech Stack:** Golang, gRPC, PostgreSQL, Firebase Auth (optional)

#### 5.1.2 Event Service
**Responsibilities:**
- Event CRUD operations
- Ticket tier management
- Event search & filtering
- Image upload to Cloud Storage

**Tech Stack:** Golang, gRPC/REST, PostgreSQL, Cloud Storage

#### 5.1.3 Ticketing Service
**Responsibilities:**
- Ticket reservation with timeout
- Inventory management dengan locking
- Order creation
- QR code generation

**Tech Stack:** Golang, gRPC, PostgreSQL, Redis (for distributed locks)

#### 5.1.4 Payment Service
**Responsibilities:**
- Xendit integration
- Payment invoice creation
- Webhook handling
- Refund processing

**Tech Stack:** Golang, REST, PostgreSQL, Xendit SDK

#### 5.1.5 Notification Service
**Responsibilities:**
- Email sending (SendGrid/Mailgun)
- Event-driven notifications via Pub/Sub
- Template management

**Tech Stack:** Golang, Cloud Pub/Sub, SendGrid API

#### 5.1.6 API Gateway
**Responsibilities:**
- Request routing
- Authentication middleware
- Rate limiting
- Request/response logging

**Tech Stack:** Golang (Gin/Echo), gRPC-Gateway

### 5.2 Database Schema (Detailed)

#### Users Table
```sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255),
  full_name VARCHAR(255) NOT NULL,
  phone VARCHAR(20),
  role VARCHAR(20) NOT NULL CHECK (role IN ('customer', 'organizer', 'admin')),
  is_email_verified BOOLEAN DEFAULT FALSE,
  oauth_provider VARCHAR(50),
  oauth_id VARCHAR(255),
  is_deleted BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email) WHERE NOT is_deleted;
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_id);
```

#### Events Table
```sql
CREATE TABLE events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  organizer_id UUID NOT NULL REFERENCES users(id),
  title VARCHAR(255) NOT NULL,
  slug VARCHAR(255) UNIQUE NOT NULL,
  description TEXT,
  category VARCHAR(50) NOT NULL,
  location VARCHAR(255) NOT NULL,
  venue VARCHAR(255),
  start_date TIMESTAMPTZ NOT NULL,
  end_date TIMESTAMPTZ NOT NULL,
  timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Jakarta',
  banner_url VARCHAR(500),
  status VARCHAR(20) NOT NULL CHECK (status IN ('draft', 'published', 'cancelled')),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  CONSTRAINT valid_dates CHECK (end_date >= start_date)
);

CREATE INDEX idx_events_organizer ON events(organizer_id);
CREATE INDEX idx_events_slug ON events(slug);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_category ON events(category);
CREATE INDEX idx_events_dates ON events(start_date, end_date);
CREATE INDEX idx_events_search ON events USING gin(to_tsvector('indonesian', title || ' ' || description));
```

#### Ticket Tiers Table
```sql
CREATE TABLE ticket_tiers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  price DECIMAL(12,2) NOT NULL CHECK (price >= 0),
  quota INTEGER NOT NULL CHECK (quota > 0),
  sold_count INTEGER DEFAULT 0 CHECK (sold_count >= 0),
  max_per_order INTEGER DEFAULT 5,
  early_bird_price DECIMAL(12,2) CHECK (early_bird_price >= 0),
  early_bird_end_date TIMESTAMPTZ,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  CONSTRAINT no_overselling CHECK (sold_count <= quota)
);

CREATE INDEX idx_ticket_tiers_event ON ticket_tiers(event_id);
CREATE INDEX idx_ticket_tiers_quota ON ticket_tiers(event_id, sold_count, quota);
```

#### Orders Table
```sql
CREATE TABLE orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id),
  event_id UUID NOT NULL REFERENCES events(id),
  total_amount DECIMAL(12,2) NOT NULL CHECK (total_amount >= 0),
  platform_fee DECIMAL(12,2) NOT NULL CHECK (platform_fee >= 0),
  promo_code VARCHAR(50),
  discount_amount DECIMAL(12,2) DEFAULT 0,
  status VARCHAR(20) NOT NULL CHECK (status IN ('reserved', 'pending_payment', 'paid', 'expired', 'refunded', 'cancelled')),
  payment_method VARCHAR(50),
  xendit_invoice_id VARCHAR(255),
  reservation_expires_at TIMESTAMPTZ,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_event ON orders(event_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_xendit ON orders(xendit_invoice_id);
CREATE INDEX idx_orders_reservation ON orders(reservation_expires_at) WHERE status = 'reserved';
```

#### Order Items Table
```sql
CREATE TABLE order_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  ticket_tier_id UUID NOT NULL REFERENCES ticket_tiers(id),
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  price DECIMAL(12,2) NOT NULL CHECK (price >= 0),
  attendee_name VARCHAR(255) NOT NULL,
  attendee_email VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_order_items_order ON order_items(order_id);
CREATE INDEX idx_order_items_tier ON order_items(ticket_tier_id);
```

#### Tickets Table
```sql
CREATE TABLE tickets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_item_id UUID NOT NULL REFERENCES order_items(id),
  qr_code VARCHAR(255) UNIQUE NOT NULL,
  validation_hash VARCHAR(255) NOT NULL,
  pdf_url VARCHAR(500),
  is_validated BOOLEAN DEFAULT FALSE,
  validated_at TIMESTAMPTZ,
  validated_by UUID REFERENCES users(id),
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_tickets_qr ON tickets(qr_code);
CREATE INDEX idx_tickets_order_item ON tickets(order_item_id);
CREATE INDEX idx_tickets_validated ON tickets(is_validated, validated_at);
```

#### Payments Table
```sql
CREATE TABLE payments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL REFERENCES orders(id),
  xendit_invoice_id VARCHAR(255) UNIQUE NOT NULL,
  amount DECIMAL(12,2) NOT NULL CHECK (amount >= 0),
  status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'paid', 'expired', 'failed')),
  payment_method VARCHAR(50),
  payment_channel VARCHAR(50),
  paid_at TIMESTAMPTZ,
  expired_at TIMESTAMPTZ,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_payments_order ON payments(order_id);
CREATE UNIQUE INDEX idx_payments_xendit ON payments(xendit_invoice_id);
CREATE INDEX idx_payments_status ON payments(status);
```

#### Webhook Events Table
```sql
CREATE TABLE webhook_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  webhook_id VARCHAR(255) UNIQUE NOT NULL,
  event_type VARCHAR(50) NOT NULL,
  payment_id UUID REFERENCES payments(id),
  payload JSONB NOT NULL,
  processed BOOLEAN DEFAULT FALSE,
  retry_count INTEGER DEFAULT 0,
  error_message TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  processed_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_webhook_id ON webhook_events(webhook_id);
CREATE INDEX idx_webhook_payment ON webhook_events(payment_id);
CREATE INDEX idx_webhook_processed ON webhook_events(processed, created_at);
```

#### Refunds Table
```sql
CREATE TABLE refunds (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL REFERENCES orders(id),
  user_id UUID NOT NULL REFERENCES users(id),
  amount DECIMAL(12,2) NOT NULL CHECK (amount >= 0),
  reason TEXT,
  status VARCHAR(20) NOT NULL CHECK (status IN ('requested', 'approved', 'rejected', 'completed', 'failed')),
  approved_by UUID REFERENCES users(id),
  rejection_reason TEXT,
  xendit_disbursement_id VARCHAR(255),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_refunds_order ON refunds(order_id);
CREATE INDEX idx_refunds_user ON refunds(user_id);
CREATE INDEX idx_refunds_status ON refunds(status);
```

### 5.3 GCP Services Usage

#### Cloud SQL
- PostgreSQL database dengan automated backups
- Read replicas untuk analytics queries

#### Google Kubernetes Engine (GKE)
- Host semua microservices
- HPA untuk auto-scaling
- Pod disruption budgets untuk high availability

#### Cloud Storage
- Store event banners/posters
- Store generated e-ticket PDFs

#### Cloud Pub/Sub
- Async communication antar services
- Topics: `payment.success`, `ticket.created`, `refund.processed`

#### Cloud Secret Manager
- Store database credentials
- Store Xendit API keys
- Store JWT signing keys

#### Cloud Logging & Monitoring
- Centralized logging
- Application metrics
- Alerting untuk critical errors

#### Cloud Scheduler
- Payment expiry checker (runs every 5 minutes)
- Event reminder sender (runs daily)

---

## 6. API Endpoints (High-Level)

### Auth Service
```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh-token
POST   /api/v1/auth/forgot-password
POST   /api/v1/auth/reset-password
```

### Event Service
```
GET    /api/v1/events (list with filters)
GET    /api/v1/events/:id
POST   /api/v1/events (EO only)
PUT    /api/v1/events/:id (EO only)
DELETE /api/v1/events/:id (EO only)
GET    /api/v1/events/:id/tickets (available tiers)
```

### Ticketing Service
```
POST   /api/v1/orders/reserve (create reservation)
POST   /api/v1/orders/:id/checkout
GET    /api/v1/orders/:id
GET    /api/v1/orders/my-orders
GET    /api/v1/tickets/:id/download
POST   /api/v1/tickets/:id/validate (untuk gate scanning)
```

### Payment Service
```
POST   /api/v1/payments/create-invoice
GET    /api/v1/payments/:id/status
POST   /api/v1/payments/webhook (Xendit callback)
POST   /api/v1/refunds/request
PUT    /api/v1/refunds/:id/approve (EO only)
```

### Analytics Service
```
GET    /api/v1/analytics/events/:id/sales
GET    /api/v1/analytics/events/:id/attendees
GET    /api/v1/analytics/platform/overview (Admin only)
```

---

## 7. Xendit Integration Details

### 7.1 Payment Channels
- **QRIS**: Single QR for all e-wallets
- **Virtual Account**: BCA, Mandiri, BNI, BRI, Permata
- **E-Wallet**: OVO, DANA, GoPay, LinkAja

### 7.2 Invoice Creation
```go
POST https://api.xendit.co/v2/invoices
{
  "external_id": "ORDER-{order_id}",
  "amount": 150000,
  "payer_email": "customer@email.com",
  "description": "Ticket for Music Festival 2025",
  "invoice_duration": 1800,  // 30 minutes (in seconds)
  "payment_methods": ["QRIS", "BANK_TRANSFER", "EWALLET"],
  "success_redirect_url": "https://app.com/payment/success",
  "failure_redirect_url": "https://app.com/payment/failed"
}
```

### 7.3 Webhook Events
System harus handle webhook events:
- `invoice.paid`: Payment berhasil → trigger ticket generation
- `invoice.expired`: Payment expired → release reservation
- `disbursement.completed`: Refund selesai → notify customer

### 7.4 Refund via Disbursement
```go
POST https://api.xendit.co/disbursements
{
  "external_id": "REFUND-{order_id}",
  "bank_code": "BCA",
  "account_holder_name": "John Doe",
  "account_number": "1234567890",
  "description": "Refund for ORDER-123",
  "amount": 135000
}
```

---

## 8. Kubernetes Configuration

### 8.1 Deployment Strategy
- Rolling updates dengan maxUnavailable: 1, maxSurge: 1
- Health checks (liveness & readiness probes)
- Resource requests & limits

### 8.2 HPA Configuration
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: ticketing-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: ticketing-service
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### 8.3 Services
- ClusterIP untuk internal communication (gRPC)
- LoadBalancer untuk API Gateway (exposed to internet)

---

## 9. Testing Strategy

### 9.1 Unit Testing
- **Framework**: Go testing package, testify
- **Coverage Target**: Minimum 80% untuk business logic
- **Scope**:
  - Service layer functions
  - Repository layer (dengan mock database)
  - Utility functions dan helpers
  - Validation logic
- **Running**: `go test ./... -cover -race`

### 9.2 Integration Testing
- **Framework**: Go testing dengan testcontainers
- **Scope**:
  - Database operations (real PostgreSQL container)
  - Redis operations
  - gRPC client-server communication
  - API endpoint testing
- **Environment**: Isolated test database per test suite

### 9.3 End-to-End Testing
- **Framework**: Playwright atau Cypress
- **Scope**:
  - Complete ticket purchase flow
  - Payment integration (Xendit sandbox)
  - Email delivery verification
  - User registration & login
- **Environment**: Staging environment

### 9.4 Load Testing
- **Tool**: K6
- **Scenarios**:
  - Normal load: 100 concurrent users
  - Flash sale: 1000 concurrent users purchasing tickets
  - Sustained load: 500 users over 30 minutes
- **Metrics**: Response time, error rate, throughput
- **Pass Criteria**: P95 < 500ms, error rate < 0.1%

### 9.5 Security Testing
- **Tools**:
  - OWASP ZAP untuk vulnerability scanning
  - golangci-lint untuk static analysis
  - Trivy untuk container scanning
- **Scope**:
  - SQL injection testing
  - XSS vulnerability testing
  - Authentication & authorization testing
  - Rate limiting verification

### 9.6 CI/CD Testing Pipeline
```yaml
Pipeline Steps:
1. Lint (golangci-lint)
2. Unit Tests (parallel per service)
3. Integration Tests (sequential)
4. Build Docker Images
5. Security Scan (Trivy)
6. Deploy to Staging
7. E2E Tests on Staging
8. Load Testing (weekly)
9. Deploy to Production (manual approval)
```

---

## 10. Environment Setup

### 10.1 Development Environment
**Infrastructure:**
- Local Kubernetes (Minikube atau Kind)
- PostgreSQL (Docker Compose)
- Redis (Docker Compose)
- SendGrid sandbox account

**Configuration:**
- Environment variables via `.env` files
- Mock Xendit API untuk development
- Hot reload dengan Air (Golang)
- Next.js dev server dengan HMR

**Developer Tools:**
- VSCode dengan Go extensions
- Postman collection untuk API testing
- pgAdmin untuk database management

### 10.2 Staging Environment
**Infrastructure:**
- GKE cluster (1 node pool, e2-medium)
- Cloud SQL (db-f1-micro)
- Cloud Memorystore Redis (1GB)
- Shared Cloud Storage bucket

**Purpose:**
- Integration testing
- E2E testing
- Client demos
- Performance testing

**Data:**
- Anonymized production data (monthly refresh)
- Xendit sandbox mode
- SendGrid test mode

### 10.3 Production Environment
**Infrastructure:**
- GKE cluster (multi-zone, 3 nodes minimum)
- Cloud SQL (db-n1-standard-2 dengan failover replica)
- Cloud Memorystore Redis (5GB dengan HA)
- Cloud Storage (multi-region)
- Cloud CDN untuk static assets

**Monitoring:**
- Cloud Logging (retention: 90 days)
- Cloud Monitoring dengan custom dashboards
- Cloud Trace untuk distributed tracing
- Uptime checks (every 1 minute)
- PagerDuty untuk alerting

**Backup:**
- Database: Automated daily backups + point-in-time recovery
- Configuration: Git repository
- Secrets: GCP Secret Manager dengan versioning

### 10.4 Environment Variables Management
```
Development: .env files (not committed)
Staging: GCP Secret Manager
Production: GCP Secret Manager + Kubernetes Secrets

Required Variables:
- DATABASE_URL
- REDIS_URL
- XENDIT_API_KEY
- SENDGRID_API_KEY
- JWT_SECRET_KEY
- STORAGE_BUCKET_NAME
- CORS_ALLOWED_ORIGINS
```

### 10.5 CI/CD Pipeline
**Tools:**
- GitHub Actions atau Cloud Build
- Artifact Registry untuk Docker images
- Kubernetes manifest files (Helm charts)

**Workflow:**
```
Development → PR → CI Tests → Merge to main → Deploy to Staging → Manual Testing → Deploy to Production
```

**Deployment Strategy:**
- Rolling updates dengan zero downtime
- Canary deployment untuk critical services (10% → 50% → 100%)
- Automatic rollback pada health check failures

---

## 11. Development Phases

**Total Timeline: 12 weeks** (termasuk 3 minggu buffer untuk bug fixes, testing, dan optimization)

### Phase 1: Foundation (Week 1-2)
- Setup GKE cluster & Cloud SQL
- Setup CI/CD pipeline (GitHub Actions + Cloud Build)
- Setup development environment (Docker Compose)
- Database schema migration setup (golang-migrate)
- Implement Auth Service (registration, login, JWT)
- Implement API Gateway (routing, auth middleware, rate limiting)
- Basic Next.js setup dengan routing
- **Deliverable**: Working authentication flow

### Phase 2: Core Features (Week 3-5)
- Implement Event Service (CRUD, image upload, search)
- Implement Ticketing Service (reservation dengan timeout, inventory locking)
- Database indexes optimization
- Redis integration untuk distributed locks
- Frontend: Event listing dengan filtering & pagination
- Frontend: Event detail page dengan ticket selection
- Frontend: Ticket purchase flow dengan reservation countdown
- **Deliverable**: Complete ticket browsing and reservation flow

### Phase 3: Payment Integration (Week 6-7)
- Implement Payment Service
- Xendit integration (invoice creation, multiple payment methods)
- Webhook handling dengan idempotency dan retry mechanism
- Transaction consistency implementation (saga pattern)
- Frontend: Payment method selection page
- Frontend: Payment status tracking page
- Unit tests untuk payment logic
- **Deliverable**: End-to-end payment flow dengan Xendit sandbox

### Phase 4: Post-Payment Features (Week 8)
- E-ticket generation (QR code generation, PDF creation)
- Cloud Storage integration untuk PDF storage
- Implement Notification Service
- Email notifications via Pub/Sub (SendGrid integration)
- Frontend: Order history page
- Frontend: Ticket download page
- **Deliverable**: Automated e-ticket delivery system

### Phase 5: Advanced Features (Week 9-10)
- Refund system (request, approval, Xendit disbursement)
- Payout system untuk Event Organizers
- Analytics dashboard untuk EO (sales, attendee stats)
- Admin panel (user management, platform analytics)
- Event cancellation flow dengan auto-refund
- Performance testing dengan K6
- Integration tests
- **Deliverable**: Complete business operations features

### Phase 6: Production Readiness (Week 11)
- Security hardening (OWASP ZAP scanning, penetration testing)
- Monitoring & alerting setup (Grafana dashboards, PagerDuty)
- Load testing (1000 concurrent users scenario)
- E2E testing dengan Playwright
- API documentation (Swagger/OpenAPI)
- Database optimization (query analysis, indexing)
- **Deliverable**: Production-ready system

### Phase 7: Buffer & Polish (Week 12)
- Bug fixes dari testing phase
- Performance optimization berdasarkan load test results
- UI/UX improvements
- Documentation completion (deployment guide, runbook)
- Security audit fixes
- Final staging environment testing
- Production deployment preparation
- **Deliverable**: Fully tested and optimized system ready for launch

---

## 12. Success Metrics

### 12.1 Technical Metrics
- API response time p95 < 500ms
- System uptime > 99.5%
- Zero overselling incidents
- Payment success rate > 95%
- Unit test coverage > 80%
- Security vulnerability score: 0 critical, 0 high

### 12.2 Business Metrics
- Support 100+ concurrent events
- Handle 10,000+ ticket purchases per day
- Average checkout time < 2 minutes
- Customer satisfaction score > 4.5/5
- Platform transaction fee revenue tracking

---

## 13. Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Xendit API downtime | High | Circuit breaker pattern, monitoring alerts, customer communication plan |
| Race condition overselling | Critical | Row-level locking (SELECT FOR UPDATE), database constraints, real-time monitoring |
| GKE node failures | Medium | Multi-zone deployment, pod disruption budgets, auto-healing |
| Email delivery failures | Medium | Retry mechanism dengan exponential backoff, dead letter queue, alternative delivery channel |
| Database bottleneck | High | Connection pooling, read replicas, query optimization, caching layer (Redis) |
| Security breach | Critical | Regular security audits, penetration testing, WAF, rate limiting, audit logging |
| Data loss | Critical | Automated backups, point-in-time recovery, multi-region replication for critical data |
| Flash sale traffic spike | High | Auto-scaling (HPA), load testing, CDN for static assets, queue system |

---

## 14. Future Enhancements
- Mobile app (Flutter/React Native)
- Seat selection untuk venue dengan seating plan
- Secondary ticket marketplace
- Loyalty program & points system
- Integration dengan Google Calendar
- WhatsApp notifications
- Multi-language support
- Multi-currency support

---

**Document Approval:**

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Developer | Bima | _______ | Dec 24, 2025 |