# Relay Project Structure & Guidelines

## Overview
Relay is a one-click deployment platform where users can deploy their websites by simply sharing their GitHub repository URL. This document defines the project structure and coding standards that all contributors must follow strictly.

## Project Structure

```
relay/
├── configs/          # Server and application configuration
├── handlers/         # HTTP request handlers (controllers)
├── services/         # Business logic layer
├── repositories/     # Database interaction layer
├── middlewares/      # HTTP middlewares
├── models/           # Database models and data structures
├── errors/           # Custom error types and handlers
├── utils/            # Reusable utility functions
├── main.go           # Application entry point
├── go.mod            # Go module dependencies
├── .env              # Environment variables (not committed)
└── template.env      # Environment template for reference
```

## Directory Responsibilities

### `configs/`
- Contains all server and application configuration logic
- Loads environment variables using `godotenv`
- Exports configuration structs for use across the application
- Example: `serverConfig.go` handles server port, host, and environment settings

### `handlers/`
- HTTP request handlers that process incoming requests
- Responsible for request validation and response formatting
- Calls service layer functions for business logic
- Should NOT contain business logic or database calls
- Returns appropriate HTTP status codes and JSON responses

### `services/`
- Contains all business logic
- Called by handlers to perform operations
- Orchestrates calls to repositories
- Handles data transformation and validation
- Should NOT directly interact with HTTP requests/responses

### `repositories/`
- All database interactions happen here
- CRUD operations and complex queries
- Returns domain models or errors
- Should NOT contain business logic

### `middlewares/`
- HTTP middleware functions
- Authentication, authorization, logging, CORS, etc.
- Applied to routes in main.go or route groups

### `models/`
- Database models and data structures
- Request/response DTOs (Data Transfer Objects)
- Shared types used across layers

### `errors/`
- Custom error types
- Error handling utilities
- Standardized error responses

### `utils/`
- Reusable utility functions
- Helper functions used across the application
- Should be pure functions when possible

## Coding Standards

### Logging with `slog`
- Use Go's standard `log/slog` package extensively throughout the application
- Log at appropriate levels: `Debug`, `Info`, `Warn`, `Error`
- **NEVER log sensitive data** (passwords, tokens, API keys, PII, etc.)
- Include contextual information using structured logging

**Example:**
```go
slog.Info("User created successfully", 
    slog.String("user_id", userID),
    slog.String("email", "[REDACTED]"), // Don't log actual email
)

slog.Error("Database connection failed",
    slog.Any("error", err),
)
```

### Environment Configuration
- Use `.env` file for local development
- Reference `template.env` for required environment variables
- Never commit `.env` to version control
- For testing purposes, use the `development` environment

**Required Environment Variables:**
```
HOST="localhost"
PORT=:3000
ENV="development"
```

### Error Handling
- Always handle errors explicitly
- Use custom error types from `errors/` package
- Log errors with context before returning
- Return meaningful error messages to clients

### Testing
- Write tests for all services and repositories
- Use table-driven tests where appropriate
- Mock external dependencies
- Test files should be named `*_test.go`

### Code Organization
- Follow the layered architecture strictly: `handlers → services → repositories`
- Keep functions small and focused
- Use dependency injection for better testability
- Avoid circular dependencies between packages

### Naming Conventions
- Use camelCase for variables and functions
- Use PascalCase for exported types and functions
- Use descriptive names that convey intent
- Prefix interfaces with `I` if it improves clarity (optional)

### HTTP Responses
- Use consistent JSON response format
- Include appropriate HTTP status codes
- Handle errors gracefully with meaningful messages

**Example Response Format:**
```json
{
    "success": true,
    "data": {},
    "message": "Operation successful"
}
```

## Getting Started

### Installation
```bash
# Install dependencies
go mod download

# Install Chi router
go get -u github.com/go-chi/chi/v5

# Install godotenv
go get github.com/joho/godotenv
```

### Running the Application
```bash
# Copy template.env to .env
cp template.env .env

# Run the application
go run main.go
```

### Development Workflow
1. Create feature branch from `main`
2. Follow the project structure strictly
3. Write code with extensive logging using `slog`
4. Test your changes in development environment
5. Ensure no sensitive data is logged
6. Submit pull request for review

## Important Reminders

- **Strictly follow the folder structure** - do not create files outside designated directories
- **Use `slog` for all logging** - no `fmt.Println` or `log.Println` in production code
- **Never log sensitive data** - redact passwords, tokens, keys, PII
- **Test in development environment** - set `ENV="development"` in `.env`
- **Handle errors properly** - log and return meaningful messages
- **Keep layers separated** - handlers → services → repositories

## Questions?
If you're unsure about where to place code or how to structure something, refer to this document first. When in doubt, ask the team lead before proceeding.
