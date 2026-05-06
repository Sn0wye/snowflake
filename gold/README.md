# Core Service - Payment & Balance Management

Core microservice responsible for managing user balances and FLAKE-style instant transactions.

## Features

- **Balance Management**: Track and manage user account balances
- **FLAKE Transactions**: Instant transfers using Flake keys (email, phone, CPF, random)

---

## Architecture: Ledger Pattern

This service follows the **ledger-based architecture** where:

- **Ledger entries are the source of truth** - All financial movements are immutable entries
- **Balance is a materialized view** - Derived from ledger, cached for performance
- **Double-entry bookkeeping** - Every transaction creates debit and credit entries
- **Eventual consistency** - Balance can be recalculated from ledger at any time
- **Audit-first design** - Complete transaction history for compliance

---

## Database Resources

### 1. **accounts**

Represents user accounts in the system.

```sql
- id (UUID, PK)
- user_id (UUID, FK to helium users, unique)
- balance (BIGINT, default 0)            -- In cents, cached value
- last_reconciled_at (TIMESTAMP, nullable)      -- Last time balance was verified
- status (ENUM: active, suspended, closed)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
```

**Note**: `balance` is stored in **cents** (smallest currency unit) as an integer. The true balance is always calculated from `transaction_history`.

### 2. **flake_keys**

Maps Flake keys to accounts for easy transfers.

```sql
- id (UUID, PK)
- account_id (UUID, FK to accounts)
- key_type (ENUM: email, phone, cpf, cnpj, random)
- key_value (VARCHAR(255), unique)
- status (ENUM: active, inactive)
- created_at (TIMESTAMP)
- updated_at (TIMESTAMP)
```

**Indexes:**

- Unique index on `key_value`
- Index on `account_id`

### 3. **transactions**

Represents the high-level transaction intent (transfer, deposit, etc.).

```sql
- id (UUID, PK)
- type (ENUM: transfer, deposit, withdrawal, refund)
- status (ENUM: pending, completed, failed, reversed)
- amount (BIGINT)                               -- In cents
- sender_account_id (UUID, FK to accounts, nullable)
- receiver_account_id (UUID, FK to accounts, nullable)
- flake_key_used (VARCHAR(255), nullable)
- description (TEXT)
- idempotency_key (UUID, unique)
- metadata (JSONB)
- created_at (TIMESTAMP)
- completed_at (TIMESTAMP, nullable)
```

**Indexes:**

- Index on `sender_account_id`
- Index on `receiver_account_id`
- Unique index on `idempotency_key`
- Index on `created_at` (for queries)
- Index on `status`

### 4. **transaction_history**

Immutable double-entry ledger. Every transaction creates 2 entries (debit + credit).

```sql
- id (UUID, PK)
- transaction_id (UUID, FK to transactions)
- account_id (UUID, FK to accounts)
- entry_type (ENUM: debit, credit)
- amount (BIGINT)                               -- In cents, always positive
- balance_before (BIGINT)                       -- In cents, snapshot for verification
- balance_after (BIGINT)                        -- In cents, snapshot for verification
- description (TEXT)
- created_at (TIMESTAMP)                        -- Immutable timestamp
```

**Indexes:**

- Index on `account_id, created_at` (for balance calculation)
- Index on `transaction_id`
- Composite index on `account_id, id` (for ordered queries)

**Key Properties:**

- **Immutable**: Entries are never updated or deleted
- **Ordered**: Entries have sequential timestamps per account
- **Double-entry**: Each transaction creates offsetting debit/credit entries
- **Balance verification**: `balance_after` allows point-in-time verification

---

## Business Rules & Requirements

### Balance Management

1. **Non-Negative Balance Rule**
   - Accounts cannot have negative balances
   - Validate sufficient funds before processing withdrawals/transfers
   - Return clear error when insufficient funds

2. **Integer-Based Amounts (Cents)**
   - Store all amounts as BIGINT in cents (smallest currency unit)
   - $100.50 → 10050 cents
   - $0.01 → 1 cent
   - Never use floating-point or DECIMAL types
   - Convert to decimal/currency format only in presentation layer
   - Benefits: exact precision, better performance, no rounding errors

3. **Ledger is Source of Truth**
   - Balance is ALWAYS calculated from `transaction_history`: `SUM(credits) - SUM(debits)`
   - `balance` is a denormalized field for read performance only
   - Never trust `balance` alone - always reconcile against ledger
   - On conflicts, ledger calculation wins

4. **Double-Entry Bookkeeping**
   - Every transaction creates exactly 2 ledger entries (debit + credit)
   - Transfer from A to B: debit on A, credit on B
   - Deposit to A: credit on A (debit from external source)
   - Total debits must equal total credits (balances to zero)

5. **Immutability of Ledger**
   - Ledger entries are NEVER updated or deleted
   - To reverse a transaction, create new offsetting entries
   - Maintain complete audit trail for compliance

### Flake Key Management

6. **Key Uniqueness**
   - Each Flake key can only be registered to ONE account
   - Validate key doesn't exist before registration
   - Support multiple keys per account (different types)

7. **Key Format Validation**
   - **Email**: Valid email format
   - **Phone**: Brazilian format (+55XXXXXXXXXXX)
   - **CPF**: 11 digits, valid checksum
   - **CNPJ**: 14 digits, valid checksum
   - **Random**: UUID v4 format

8. **Key Limits**
   - Maximum 5 keys per account
   - Each key type can only be registered once per account
   - Random keys: unlimited per account

9. **Key Lifecycle**
   - Keys can be deactivated but not deleted (for audit)
   - Only active keys can be used for transactions
   - Support key portability (deactivate + reactivate on new account)

### Transaction Processing

10. **ACID Transactions**

- Use database transactions for all balance updates
- Debit sender and credit receiver atomically
- Rollback completely on any failure

11. **Idempotency**

- Every transaction request must include an idempotency_key
- Duplicate requests with same key return original transaction
- Store idempotency keys for at least 24 hours

12. **Transaction Validation**
    - Amount must be positive (> 0)
    - Amount must have max 2 decimal places
    - Sender and receiver cannot be the same account
    - Both accounts must have 'active' status

13. **Transaction Limits**
    - Minimum transaction: 1 cent ( $ 0.01)
    - Maximum transaction: 10000000 cents ($ 100,000.00, configurable)
    - Daily limit per account (configurable, default 5000000 = $ 50,000.00)
    - Maximum 100 transactions per day per account

14. **Concurrency Control**
    - Use row-level locking (`SELECT ... FOR UPDATE`) on accounts
    - Prevent race conditions in concurrent transactions
    - Implement retry logic with exponential backoff

### Security & Compliance

15. **Authentication & Authorization**
    - Verify user owns the sender account via JWT token
    - Only account owner can initiate transfers from their account
    - Admin role can view all transactions (not initiate)

<!-- 14. **Data Privacy**
    - Don't expose full Flake keys in transaction history (mask them)
    - Log access to sensitive transaction data
    - Implement data retention policies

15. **Fraud Prevention**
    - Rate limiting: max 10 transactions per minute per account
    - Flag suspicious patterns (many failed transactions, rapid succession)
    - Implement transaction velocity checks -->

### Integration & Communication

16. **Event Publishing**
    - Publish events to RabbitMQ on transaction completion
    - Events: `transaction.created`, `transaction.completed`, `transaction.failed`
    - Include minimal data (IDs, status, timestamp)

17. **Notification Service Integration**
    - Send notification on transaction completion
    - Include transaction details (amount, sender, receiver)

18. **Reconciliation & Balance Verification**

- Hourly job: recalculate `balance` from ledger entries
- Formula: `SUM(credits) - SUM(debits)` for each account
- Compare calculated balance with `balance`
- If mismatch: log alert, update cache, investigate discrepancy
- On-demand reconciliation API for debugging
- Verify double-entry: total system debits = total system credits

---

## API Endpoints

### Balance Endpoints

- `GET /account/balance` - Get current user balance
- `GET /account/balance/history` - Get balance change history

### Flake Endpoints

- `POST /account/flake` - Register new Flake key
- `GET /account/flake` - List user's Flake keys
- `DELETE /account/flake/:keyId` - Deactivate Flake key
- `GET /account/flake/lookup?keyValue=:keyValue` - Lookup account by Flake key value (query param)

### Transaction Endpoints

- `POST /account/transactions/transfer` - Execute FLAKE transfer
- `GET /account/transactions` - List user's transactions (paginated)
- `GET /account/transactions/:id` - Get transaction details
- `POST /account/transactions/deposit` - Deposit funds (admin/test only)

---

## Implementation Steps

### Phase 1: Database Setup

1. Create database migration files
2. Set up connection pool and ORM/query builder
3. Create repository layer for each entity

### Phase 2: Core Business Logic

1. Implement account service (create, get balance)
2. Implement Flake key service (register, validate, lookup)
3. Implement transaction service with locking mechanism
4. Add validation logic for all business rules

### Phase 3: API Layer

1. Set up HTTP router (Fiber/Gin)
2. Create middleware (auth, rate limiting, logging)
3. Implement REST endpoints
4. Add request/response DTOs with validation

### Phase 4: Integration

1. Set up gRPC service for inter-service communication
2. Configure RabbitMQ producer for events
3. Add health check endpoint

### Phase 5: Testing & Deployment

1. Unit tests for business logic
2. Integration tests for API endpoints
3. Load testing for concurrent transactions
4. Docker containerization
5. Add to docker-compose.yml

---

## Technology Stack

- **Language**: Go 1.23+
- **Framework**: Fiber (HTTP) + gRPC
- **Database**: PostgreSQL 14+
- **Message Queue**: RabbitMQ
- **Validation**: go-playground/validator
- **Migration**: golang-migrate or goose
- **ORM**: sqlx or GORM

---

## Environment Variables

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=snowflake_core
DB_USER=postgres
DB_PASSWORD=postgres

RABBITMQ_URL=amqp://guest:guest@localhost:5672/

JWT_SECRET=<shared-with-helium>

GRPC_PORT=50052
HTTP_PORT=8082

MAX_TRANSACTION_AMOUNT=100000.00
DAILY_TRANSACTION_LIMIT=50000.00
MAX_DAILY_TRANSACTIONS=100
```

---

## Notes

- All monetary amounts are in USD (or local currency) but stored as integer cents
- Timestamp fields use UTC timezone
- API versioning follows semver (currently v1)
- Rate limiting uses sliding window algorithm
