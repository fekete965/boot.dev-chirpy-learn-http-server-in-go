# Chirpy - HTTP Server in Go

A modern, RESTful HTTP server built with Go for managing users and chirps (micro-posts). Features JWT authentication, refresh tokens, profanity filtering, and webhook support.

## ✨ Features

- 🔐 **User Authentication** - JWT-based authentication with refresh tokens
- 📝 **Chirps Management** - Create, read, update, and delete chirps
- 🛡️ **Profanity Filtering** - Automatic content moderation
- 🔄 **Token Refresh** - Secure token rotation mechanism
- 📊 **Metrics Tracking** - Admin metrics endpoint
- 🎣 **Webhook Support** - Polka webhook integration for user upgrades
- 📁 **Static File Serving** - Serve static assets and frontend files
- 🗄️ **PostgreSQL Integration** - Type-safe database queries with sqlc

## 🛠️ Tech Stack

- **Language**: Go 1.23+
- **Database**: PostgreSQL
- **Authentication**: JWT (golang-jwt/jwt)
- **Password Hashing**: Argon2id
- **Query Builder**: sqlc
- **UUID**: google/uuid

## 📋 Prerequisites

- Go 1.23 or higher
- PostgreSQL database
- Environment variables configured (see Configuration)

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
# Run migrations from sql/schema/
# Files should be executed in order: 001_users.sql, 002_chirps.sql, etc.
psql -U your_user -d your_database -f sql/schema/001_users.sql
psql -U your_user -d your_database -f sql/schema/002_chirps.sql
psql -U your_user -d your_database -f sql/schema/003_users.sql
psql -U your_user -d your_database -f sql/schema/004_refresh_tokens.sql
psql -U your_user -d your_database -f sql/schema/006_users.sql
```

### 4. Generate Database Code

If you modify SQL queries, regenerate the database code:

```bash
sqlc generate
```

### 5. Configure Environment Variables

Create a `.env` file in the root directory:

```env
PORT=8080
JWT_SECRET=your-super-secret-jwt-key-here
DATABASE_URL=postgres://user:password@localhost/dbname?sslmode=disable
POLKA_KEY=your-polka-webhook-secret
```

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
- `PUT /api/users` - Update user information
- `POST /api/login` - Login and receive JWT tokens
- `POST /api/refresh` - Refresh access token
- `POST /api/revoke` - Revoke refresh token

### Chirps
- `POST /api/chirps` - Create a new chirp (requires authentication)
- `GET /api/chirps` - Get all chirps
  - Query params: `author_id` (optional UUID), `sort` (optional: "asc" or "desc")
- `GET /api/chirps/{chirpID}` - Get a specific chirp by ID
- `DELETE /api/chirps/{chirpID}` - Delete a chirp (requires authentication, owner only)

### Webhooks
- `POST /api/polka/webhooks` - Handle Polka webhook events

### Admin
- `GET /admin/metrics` - View server metrics
- `POST /admin/reset` - Reset metrics counter

### Static Files
- `GET /app/` - Serve frontend application
- `GET /app/assets/` - Serve static assets

## 📁 Project Structure

```
.
├── internal/
│   ├── auth/           # Authentication utilities
│   │   ├── auth.go
│   │   └── auth_test.go
│   └── database/       # Database layer (sqlc generated + extensions)
│       ├── chirps.sql.go
│       ├── chirps_ext.go
│       ├── users.sql.go
│       ├── refresh_tokens.sql.go
│       └── models.go
├── sql/
│   ├── schema/         # Database migrations
│   └── queries/        # SQL queries for sqlc
├── static/             # Static files and frontend
│   ├── index.html
│   └── assets/
├── main.go             # Main application entry point
├── go.mod
├── go.sum
└── sqlc.yaml           # sqlc configuration
```

## 🔧 Development

### Running Tests

```bash
go test ./...
```

### Code Generation

If you modify SQL queries in `sql/queries/`, regenerate the Go code:

```bash
sqlc generate
```

### Building

```bash
go build -o chirpy main.go
```

## 🔒 Security Features

- **Password Hashing**: Uses Argon2id for secure password storage
- **JWT Tokens**: Secure token-based authentication
- **Refresh Tokens**: Long-lived refresh tokens with secure rotation
- **Input Validation**: UUID validation and parameter sanitization
- **SQL Injection Prevention**: Parameterized queries via sqlc

## 📝 Notes

- Chirps are automatically filtered for profanity (kerfuffle, sharbert, fornax)
- Access tokens expire after 1 hour
- Refresh tokens expire after 60 days
- Users can upgrade to "Chirpy Red" via Polka webhook integration

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

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

