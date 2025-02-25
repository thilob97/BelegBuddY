# BelegBuddY - Developer Guidelines

## Build & Test Commands
- Build: `go build -o belegbuddy ./cmd/belegbuddy`
- Run: `go run ./cmd/belegbuddy`
- Test all: `go test ./...`
- Test single package: `go test ./pkg/modulname`
- Test single function: `go test -run TestFunctionName ./pkg/modulname`
- Lint: `golangci-lint run`

## Code Style Guidelines
- Use Go modules for dependency management
- Follow standard Go formatting with `gofmt`
- Organize imports alphabetically: standard library, then third-party, then internal
- Use meaningful variable names in lowerCamelCase; package names in lowercase
- Struct fields should use PascalCase for exported fields
- Error handling: always check errors; use custom error types for domain-specific errors
- Prefer early returns over nested if-else blocks
- Use context for cancellation and timeouts when appropriate
- Implementation follows domain-driven design with clear boundaries
- Document all exported functions, types, and constants

## Project Specifications
- GUI Framework: Fyne or Wails
- Database: SQLite with GORM
- OCR: Tesseract via gosseract
- Logging: logrus for structured logging
- File structure: cmd/, pkg/, internal/ directories following Go project layout conventions