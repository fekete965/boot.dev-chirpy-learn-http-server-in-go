# Chirpy - HTTP Server in Go

A production-ready RESTful HTTP server built with Go for managing users and chirps (micro-posts). Features JWT authentication, refresh tokens, profanity filtering, webhook support, and comprehensive test coverage.

## ✨ Features

- 🔐 **User Authentication** - JWT-based authentication with refresh tokens
- 📝 **Chirps Management** - Create, list, fetch-by-id, and delete chirps with authorization
- 🛡️ **Profanity Filtering** - Automatic content moderation
- 🔄 **Token Refresh** - Secure token rotation mechanism with revocation support
- 📊 **Metrics Tracking** - Admin metrics endpoint
- 🎣 **Webhook Support** - Polka webhook integration for user upgrades
- 📁 **Static File Serving** - Serve static assets and frontend files
- 🗄️ **PostgreSQL Integration** - Type-safe database queries with sqlc
- 🏗️ **Clean Architecture** - Layered architecture with handlers, services, and database layers
- 🧪 **Comprehensive Testing** - Unit and integration tests with testcontainers

## 🛠️ Tech Stack

- **Language**: Go 1.24+
- **Database**: PostgreSQL
- **Authentication**: JWT (golang-jwt/jwt/v4)
- **Password Hashing**: Argon2id
- **Query Builder**: sqlc (type-safe SQL)
- **UUID**: google/uuid
- **Testing**: testify, testcontainers-go
- **Migrations**: goose/v3
- **Environment**: godotenv

## 📋 Prerequisites

- Go 1.24 or higher
- PostgreSQL database (or Docker for testcontainers)
- Environment variables configured (see Configuration section)

## 🚀 Getting Started

### 1. Clone the Repository

```bash
git clone <repository-url>
cd boot.dev-chirpy-learn-http-server-in-go
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Set Up Database

Create a PostgreSQL database and run the migration files in order:

```bash
# Option 1: Using psql directly
psql -U your_user -d your_database -f sql/schema/001_users.sql
psql -U your_user -d your_database -f sql/schema/002_chirps.sql
psql -U your_user -d your_database -f sql/schema/003_users.sql
psql -U your_user -d your_database -f sql/schema/004_refresh_tokens.sql
psql -U your_user -d your_database -f sql/schema/006_users.sql

# Option 2: Using goose (recommended)
goose -dir sql/schema postgres "postgres://user:password@localhost/dbname?sslmode=disable" up
```

### 4. Generate Database Code

If you modify SQL queries in `sql/queries/`, regenerate the database code:

```bash
sqlc generate
```

### 5. Configure Environment Variables

Create a `.env` file in the root directory:

```env
PORT=8080
JWT_SECRET=your-super-secret-jwt-key-here-min-32-chars
DB_URL=postgres://user:password@localhost/dbname?sslmode=disable
PLATFORM=dev
POLKA_KEY=your-polka-webhook-secret
```

**Required Environment Variables:**

| Variable     | Description                                      | Default  |
| ------------ | ------------------------------------------------ | -------- |
| `PORT`       | Server port                                      | `8080`   |
| `JWT_SECRET` | Secret key for JWT token signing (min 32 chars)  | Required |
| `DB_URL`     | PostgreSQL database connection string            | Required |
| `PLATFORM`   | Platform environment (e.g., "dev", "production") | Required |
| `POLKA_KEY`  | Secret key for Polka webhook authentication      | Required |

### 6. Run the Server

```bash
go run main.go
```

The server will start on the port specified in your `.env` file (default: 8080).

## 📡 API Endpoints

### Health Check

- `GET /api/healthz` - Health check endpoint

### Authentication

- `POST /api/users` - Create a new user
  - Request body: `{ "email": "user@example.com", "password": "password" }`
  - Returns: User object with ID, email, and timestamps
- `PUT /api/users` - Update user information (requires authentication)
  - Headers: `Authorization: Bearer <access_token>`
  - Request body: `{ "email": "new@example.com", "password": "newpassword" }`
- `POST /api/login` - Login and receive JWT tokens
  - Request body: `{ "email": "user@example.com", "password": "password" }`
  - Returns: User object with `token` (access token) and `refresh_token`
- `POST /api/refresh` - Refresh access token (requires refresh token)
  - Headers: `Authorization: Bearer <refresh_token>`
  - Returns: `{ "token": "<new_access_token>" }`
- `POST /api/revoke` - Revoke refresh token (requires refresh token)
  - Headers: `Authorization: Bearer <refresh_token>`
  - Returns: 204 No Content

### Chirps

- `POST /api/chirps` - Create a new chirp (requires authentication)
  - Headers: `Authorization: Bearer <access_token>`
  - Request body: `{ "body": "Your chirp content here" }`
  - Max length: 140 characters
  - Profanity filtering: Automatic replacement of banned words
- `GET /api/chirps` - Get all chirps
  - Query params:
    - `author_id` (optional UUID) - Filter by author
    - `sort` (optional: "asc" or "desc") - Sort order, default: "asc"
- `GET /api/chirps/{chirpID}` - Get a specific chirp by ID
  - Returns 404 if chirp not found
- `DELETE /api/chirps/{chirpID}` - Delete a chirp (requires authentication, owner only)
  - Headers: `Authorization: Bearer <access_token>`
  - Returns 403 if user is not the owner
  - Returns 204 on success

### Webhooks

- `POST /api/polka/webhooks` - Handle Polka webhook events (requires API key)
  - Headers: `Authorization: ApiKey <polka_key>`
  - Request body: `{ "event": "user.upgraded", "data": { "user_id": "<uuid>" } }`
  - Upgrades user to "Chirpy Red" status

### Admin

- `GET /admin/metrics` - View server metrics
  - Returns HTML page with fileserver hit count
- `POST /admin/reset` - Reset metrics counter (dev only)
  - Resets the fileserver hit counter to 0

### Static Files

- `GET /app/` - Serve frontend application
- `GET /app/assets/` - Serve static assets

## 📁 Project Structure

```
.
├── internal/
│   ├── auth/                    # Authentication utilities
│   │   ├── auth.go              # JWT, password hashing, token utilities
│   │   └── auth_test.go         # Authentication unit tests
│   ├── config/                  # Configuration management
│   │   ├── config.go            # Config struct and environment loading
│   │   └── config_test.go       # Configuration tests
│   ├── constants/               # Centralized constants
│   │   ├── app.go               # Application-level constants
│   │   ├── auth.go              # Auth-related constants (token expiration)
│   │   └── database.go          # Database constants
│   ├── database/                # Database layer (sqlc generated + extensions)
│   │   ├── chirps.sql.go        # Generated chirp queries
│   │   ├── chirps_ext.go        # Extended chirp queries
│   │   ├── users.sql.go         # Generated user queries
│   │   ├── refresh_tokens.sql.go # Generated refresh token queries
│   │   ├── models.go            # Database models
│   │   └── db.go                # Database connection and queries struct
│   ├── handlers/                # HTTP request handlers
│   │   ├── handlers.go          # Handler struct with services and config
│   │   ├── admin_handler.go     # Admin endpoints (metrics, reset, webhooks)
│   │   ├── auth_handler.go      # Authentication endpoints
│   │   ├── chirp_handler.go     # Chirp CRUD endpoints
│   │   ├── user_handler.go      # User management endpoints
│   │   ├── health.go            # Health check endpoint
│   │   └── app.go               # Static file handlers
│   ├── middlewares/             # HTTP middleware
│   │   ├── middlewares.go       # Middlewares struct with config
│   │   ├── logger.go            # Request logging middleware
│   │   └── metric.go            # Metrics tracking middleware
│   ├── models/                  # API models (DTOs)
│   │   ├── user.go              # User request/response models
│   │   ├── chirp.go             # Chirp request/response models
│   │   ├── auth.go              # Auth request/response models
│   │   ├── env.go               # Environment variables model
│   │   ├── polka.go             # Polka webhook models
│   │   └── user.go              # User models
│   ├── service_errors/          # Custom error types
│   │   └── service_errors.go    # Service error types with HTTP status codes
│   ├── services/                # Business logic layer
│   │   ├── service.go           # Services container
│   │   ├── user_service.go      # UserService interface
│   │   ├── user_service_impl.go # UserService implementation
│   │   ├── user_service_impl_test.go # UserService integration tests
│   │   ├── chirp_service.go     # ChirpService interface
│   │   ├── chirp_service_impl.go # ChirpService implementation
│   │   ├── chirp_service_impl_test.go # ChirpService integration tests
│   │   ├── auth_service.go      # AuthService interface
│   │   └── auth_service_impl.go # AuthService implementation
│   ├── testdb/                  # Test database utilities
│   │   ├── postgres.go          # PostgreSQL testcontainer setup with migrations
│   │   └── generators.go        # Test data generators using gofakeit
│   ├── testutils/               # Test utilities and helpers
│   │   └── testutils.go         # Test setup helpers, transaction utilities, test config
│   └── utils/                   # Utility functions
│       ├── response.go          # HTTP response helpers
│       ├── response_test.go     # Response utility tests
│       ├── request.go           # Request parsing utilities
│       ├── request_test.go      # Request utility tests
│       ├── validation.go        # Validation utilities
│       ├── validation_test.go  # Validation utility tests
│       ├── service_errors.go   # Error mapper utility
│       └── service_errors_test.go # Error mapper tests
├── sql/
│   ├── schema/                  # Database migrations
│   │   ├── 001_users.sql
│   │   ├── 002_chirps.sql
│   │   ├── 003_users.sql
│   │   ├── 004_refresh_tokens.sql
│   │   └── 006_users.sql
│   └── queries/                 # SQL queries for sqlc
│       ├── chirps.sql
│       ├── users.sql
│       └── refresh_tokens.sql
├── static/                      # Static files and frontend
│   ├── index.html
│   └── assets/
│       └── logo.png
├── main.go                      # Main application entry point (~65 lines)
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── sqlc.yaml                    # sqlc configuration
└── README.md                    # This file
```

## 🔧 Development

### Project Architecture

This project follows idiomatic Go patterns with a clean layered architecture:

```
┌─────────────────────────────────────────┐
│           HTTP Layer                    │
│  (handlers - request/response)          │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│        Service Layer                     │
│  (business logic, validation)            │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│        Database Layer                    │
│  (sqlc generated queries)               │
└─────────────────────────────────────────┘
```

**Key Components:**

- **Handlers** (`internal/handlers/`) - Handle HTTP requests/responses, delegate to services
- **Services** (`internal/services/`) - Business logic layer, contains all domain logic
- **Database** (`internal/database/`) - Database queries and models (sqlc generated)
- **Models** (`internal/models/`) - API request/response DTOs
- **Utils** (`internal/utils/`) - Reusable utility functions
- **Middleware** (`internal/middlewares/`) - HTTP middleware (logging, metrics)
- **Config** (`internal/config/`) - Configuration management
- **Constants** (`internal/constants/`) - Centralized application constants
- **Service Errors** (`internal/service_errors/`) - Custom error types with HTTP status codes

### Key Design Patterns

- **Handler Struct Pattern** - Handlers struct holds services and config dependencies
- **Service Layer** - All business logic encapsulated in services
- **Dependency Injection** - Services and handlers receive dependencies via constructors
- **Custom Error Types** - Service errors include HTTP status codes for proper error mapping
- **Generic Utilities** - Type-safe request body decoding with generics (`DecodeRequestBody[T]()`)
- **Interface-Based Design** - Service interfaces enable easy testing and mocking

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./internal/services/...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Test Coverage:**

- ✅ Authentication utilities (`internal/auth/`)
- ✅ Configuration loading (`internal/config/`)
- ✅ User service integration tests (`internal/services/user_service_impl_test.go`)
  - 12 test cases covering: Create, FindByEmail, FindByID, Update, DeleteAll, IsChirpyRed updates
  - Tests include conflict handling, not found scenarios, and edge cases
- ✅ Chirp service integration tests (`internal/services/chirp_service_impl_test.go`)
  - 17 test cases covering: Create, GetAll, GetByID, Delete operations
  - Tests include profanity filtering, validation, authorization, sorting, and filtering
- ✅ Auth service integration tests (`internal/services/auth_service_impl_test.go`)
  - 13 test cases covering: Login, RefreshToken, RevokeToken, UpgradeUser operations
  - Tests include authentication, token refresh, revocation (including already revoked/expired scenarios), user upgrades, and error scenarios
- ✅ User handler HTTP integration tests (`internal/handlers/user_handler_test.go`)
  - 18 test cases covering: CreateUser and UpdateUser endpoints
  - CreateUser tests: Happy path, malformed payload, invalid email, duplicate email, empty email/password validation
  - UpdateUser tests: Happy path, malformed payload, invalid email, duplicate email, user not found, authentication errors (no token, invalid token, expired token, malformed header), updating to same email, empty email/password validation
- ✅ Request utilities (`internal/utils/request.go`)
- ✅ Response utilities (`internal/utils/response.go`)
- ✅ Validation utilities (`internal/utils/validation.go`)
- ✅ Error handling (`internal/utils/service_errors.go`)

**Testing Infrastructure:**

- **testcontainers-go**: Integration tests with real PostgreSQL instances
- **testdb package**: PostgreSQL container setup and test data generators
  - `SetupPostgres()`: Singleton pattern for container initialization
  - `Generator`: Interface-based test data generation using gofakeit
  - Supports generating users, chirps, and custom types
- **testutils package**: Test helpers and utilities
  - `NewTestHelper()`: Complete test environment setup
  - `WithTx()`: Transaction-based test isolation with automatic rollback
  - `GetTestApiConfig()`: Test configuration factory
- **Transaction isolation**: All integration tests use transactions that are automatically rolled back

### Code Generation

If you modify SQL queries in `sql/queries/`, regenerate the Go code:

```bash
sqlc generate
```

### Building

```bash
# Build the binary
go build -o chirpy main.go

# Build for production (with optimizations)
go build -ldflags="-s -w" -o chirpy main.go

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o chirpy-linux main.go
```

### Development Workflow

1. Make changes to SQL queries in `sql/queries/`
2. Run `sqlc generate` to regenerate database code
3. Update services/handlers as needed
4. Write/update tests: `go test ./...`
5. Build and test: `go run main.go`
6. Run linter: `golangci-lint run` (if configured)

## 🔒 Security Features

- **Password Hashing**: Uses Argon2id for secure password storage
- **JWT Tokens**: Secure token-based authentication with expiration
- **Refresh Tokens**: Long-lived refresh tokens with secure rotation and revocation
  - Defensive validation at both SQL and service layers
  - Clear error messages for revoked, expired, and invalid tokens
- **Input Validation**: UUID validation, parameter sanitization, and length checks
- **SQL Injection Prevention**: Parameterized queries via sqlc
- **Error Handling**: Secure error messages that don't leak internal details
- **Authorization**: Resource-level authorization (users can only delete their own chirps)
- **Webhook Authentication**: API key validation for webhook endpoints

## 📝 Important Notes

### Profanity Filtering

Chirps are automatically filtered for profanity using the `utils.CleanChirp` function. Banned words are replaced with asterisks:

- `kerfuffle` → `****`
- `sharbert` → `****`
- `fornax` → `****`

### Token Expiration

- **Access tokens**: Expire after 1 hour
- **Refresh tokens**: Expire after 60 days
- Refresh tokens can be revoked and are checked for expiration on each use
- **Defensive Token Handling**: Both SQL queries and service layer validate token status before operations
  - Service layer provides clear error messages ("already revoked", "already expired", "not found")
  - SQL queries include defensive WHERE clauses to prevent invalid operations
  - Double-layer validation ensures data integrity and better error reporting
- **JWT Token Uniqueness**: Each JWT token includes a unique ID (jti claim) to ensure tokens are always unique, even when generated in the same second

### User Upgrades

Users can upgrade to "Chirpy Red" status via Polka webhook integration. The upgrade is triggered by a webhook event and updates the user's `is_chirpy_red` flag.

### Architecture Benefits

- **Separation of Concerns**: Clear boundaries between layers
- **Testability**: Services can be easily unit tested with mock dependencies
- **Maintainability**: Business logic centralized in service layer
- **Scalability**: Easy to add new features following established patterns
- **Type Safety**: sqlc-generated database code prevents SQL errors at compile time
- **Clean main.go**: Minimal wiring code (~65 lines)

### Request Utilities

- `DecodeRequestBody[T]()` - Generic function for type-safe JSON body decoding
- `GetAuthenticatedUserID()` - Extract and validate JWT access tokens
- `GetBearerToken()` - Extract Bearer token from Authorization header
- `GetQueryParam()` - Extract and parse query parameters with defaults

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Write tests for new functionality
5. Run tests (`go test ./...`)
6. Ensure code follows Go conventions
7. Commit your changes (`git commit -m 'Add some amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

## 📄 License

MIT License

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

---

Built with ❤️ using Go
