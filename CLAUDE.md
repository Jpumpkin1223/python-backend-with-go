# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a simple HTTP backend server written in Go. Despite the repository name "python-backend-with-go", this is a pure Go project with no Python components.

## Architecture

Single-file application ([main.go](main.go)) with standard Go HTTP server using `net/http` package:
- Uses Go 1.22+ enhanced ServeMux with method-based routing (e.g., `GET /`, `POST /api/signup`)
- Structured logging with `log/slog` (JSON handler)
- Middleware chain: logging → recovery → CORS → security headers
- Graceful shutdown with 30-second timeout
- In-memory data storage (map-based, no database)
- Request/response handling follows standard patterns:
  - JSON encoding/decoding with `encoding/json`
  - Consistent error responses via `ErrorResponse` struct
  - Request context with UUID-based request IDs

**Key Implementation Details:**
- All handlers are thin and handle HTTP concerns only (parsing, validation, response formatting)
- User data stored in global `users` map (keyed by user ID)
- Password stored in plain text (in-memory only, not production-ready)
- No authentication or authorization implemented

## Common Commands

### Development
```bash
# Install/update dependencies
go mod tidy

# Run server directly
go run main.go

# Run with custom port
PORT=3000 go run main.go
```

### Build
```bash
# Build executable
go build -o server main.go

# Run built server
./server
```

## API Endpoints

- `GET /` - Welcome message with request info
- `GET /health` - Health check endpoint
- `GET /api/hello` - JSON response example
- `POST /api/signup` - User registration (name, email, password, profile)

### Example: User Signup
```bash
curl -X POST http://localhost:8080/api/signup \
  -H "Content-Type: application/json" \
  -d '{
    "name": "홍길동",
    "email": "hong@example.com",
    "password": "password123",
    "profile": "안녕하세요"
  }'
```

## Requirements

- Go 1.25.1 (specified in go.mod)
- External dependencies:
  - `github.com/google/uuid` - UUID generation for user IDs and request tracking

## Development Guidelines (from .cursor/rules)

This project follows Go standard library best practices for API development:

### Code Style
- Use Go 1.22+ ServeMux with method-based routing: `mux.HandleFunc("POST /api/signup", handler)`
- Extract path parameters with `r.PathValue("id")`
- Keep handlers thin - delegate to service layer for business logic (when applicable)
- Use `context.Context` for timeouts, cancellation, and request-scoped values
- Always handle errors explicitly with `fmt.Errorf` for wrapping

### HTTP Best Practices
- Set appropriate HTTP status codes (200, 201, 400, 404, 409, 500)
- Use consistent JSON error format via `ErrorResponse` struct
- Set `Content-Type: application/json` header for JSON responses
- Validate all input data before processing
- Use `json.NewDecoder(r.Body).Decode()` for request parsing

### Logging & Monitoring
- Use `log/slog` for structured logging with JSON output
- Include request_id, method, path, status, duration in request logs
- Log at appropriate levels: Info for requests, Error for failures
- Never log sensitive data (passwords, tokens)

### Server Configuration
- Set timeouts on `http.Server`: ReadTimeout, WriteTimeout, IdleTimeout
- Implement graceful shutdown with `srv.Shutdown(ctx)`
- Use middleware pattern: `func(http.Handler) http.Handler`
- Current middleware chain: logging → recovery → CORS → security headers

### Security
- Validate and sanitize all user input
- Set security headers (X-Content-Type-Options, X-Frame-Options, etc.)
- Handle CORS properly when needed
- Use HTTPS in production

## Git Submodules

This repository uses git submodules for Cursor AI rules:

### Initial Setup
```bash
# Clone with submodules
git clone --recurse-submodules https://github.com/Jpumpkin1223/python-backend-with-go.git

# Or if already cloned, initialize submodules
git submodule update --init --recursive
```

### Submodule Location
- `.cursor/rules/` - Cursor AI development rules (managed separately at https://github.com/Jpumpkin1223/cursor-go-rules)

### Updating Submodules
```bash
# Update submodule to latest commit
git submodule update --remote .cursor/rules

# Commit the submodule update
git add .cursor/rules
git commit -m "Update cursor rules submodule"
```
