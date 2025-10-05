# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a simple HTTP backend server written in Go. Despite the repository name "python-backend-with-go", this is a pure Go project with no Python components.

## Architecture

Single-file application ([main.go](main.go)) with standard Go HTTP server using `net/http` package:
- HTTP handlers registered inline using `http.HandleFunc`
- No external dependencies beyond Go standard library
- Simple, flat structure with all logic in main package

## Common Commands

### Development
```bash
# Install/update dependencies
go mod tidy

# Run server directly
go run main.go
```

### Build
```bash
# Build executable
go build -o server main.go

# Run built server
./server
```

### Configuration
- `PORT` environment variable controls server port (default: 8080)
- Set via: `PORT=3000 go run main.go`

## API Endpoints

- `GET /` - Welcome message with request info
- `GET /health` - Health check endpoint
- `GET /api/hello` - JSON response example

## Requirements

- Go 1.25.1 (specified in go.mod)
- No external dependencies
