# AGENTS.md - Coding Guidelines for github.com/yuhang-jieke/exam

## Project Overview
- **Language**: Go 1.25.8
- **Module**: `github.com/yuhang-jieke/exam`
- **Type**: Minimal Go project (currently empty)

## Build/Lint/Test Commands

### Build
```bash
# Build the project
go build ./...

# Build with verbose output
go build -v ./...

# Build for specific OS/arch
go build -o bin/app ./cmd/app  # if cmd/app exists
```

### Test
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a single test
go test -v -run TestFunctionName ./package

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests with race detector
go test -race ./...
```

### Lint
```bash
# Format code
go fmt ./...

# Vet (static analysis)
go vet ./...

# Use golangci-lint if installed
golangci-lint run ./...
```

### Dependencies
```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify

# Update dependencies
go get -u ./...
go mod tidy
```

## Code Style Guidelines

### Formatting
- Use `go fmt` for automatic formatting
- Use tabs for indentation (Go standard)
- Max line length: 120 characters (soft limit)
- Group related imports

### Imports
```go
// Standard library imports first
import (
    "fmt"
    "os"
    "time"
)

// Then third-party imports (blank line separator)
import (
    "github.com/some/package"
    "golang.org/x/exp/slices"
)

// Then local/project imports (blank line separator)
import (
    "github.com/yuhang-jieke/exam/internal/config"
    "github.com/yuhang-jieke/exam/pkg/utils"
)
```

Use `goimports` for automatic import management:
```bash
goimports -w .
```

### Naming Conventions

**Files**: `lowercase.go`, `snake_case_test.go` for tests

**Packages**:
- Short, lowercase, single word
- No underscores or mixedCaps
- Examples: `user`, `auth`, `httputil`

**Types**:
- `PascalCase` for exported types: `UserService`, `HTTPClient`
- `camelCase` for unexported types: `userConfig`, `httpTransport`

**Variables**:
- `PascalCase` for exported: `DefaultTimeout`, `MaxRetries`
- `camelCase` for unexported: `defaultTimeout`, `maxRetries`
- Short names for short scopes: `i`, `n` for loops
- Longer names for longer scopes: `userRepository`, `configLoader`

**Constants**:
- `PascalCase` for exported: `MaxBufferSize`
- `camelCase` for unexported: `maxBufferSize`
- Or `SCREAMING_SNAKE_CASE` for true constants: `DEFAULT_PORT`

**Interfaces**:
- Single method: `Reader`, `Writer`, `Closer`
- Multiple methods: `UserRepository`, `ConfigProvider`
- `-er` suffix for single method interfaces

**Functions/Methods**:
- `PascalCase` for exported: `GetUserByID`, `ParseConfig`
- `camelCase` for unexported: `getUserByID`, `parseConfig`
- Getters: `User()` not `GetUser()`
- Setters: `SetUser(user User)`

### Error Handling
```go
// Always check errors
data, err := ioutil.ReadFile("config.json")
if err != nil {
    return fmt.Errorf("failed to read config: %w", err)
}

// Use fmt.Errorf with %w for wrapping
if err := process(); err != nil {
    return fmt.Errorf("processing failed: %w", err)
}

// Define sentinel errors
var ErrNotFound = errors.New("not found")

// Check specific errors with errors.Is
if errors.Is(err, ErrNotFound) {
    // handle not found
}

// Check error types with errors.As
if err := doSomething(); err != nil {
    var pathErr *fs.PathError
    if errors.As(err, &pathErr) {
        // handle path error
    }
}

// Panic only in truly exceptional cases (programmer errors)
if invariant == nil {
    panic("invariant violated: must not be nil")
}
```

### Types
```go
// Prefer explicit types over `any`/`interface{}`
func Process(data []byte) error  // Good
func Process(data any) error     // Avoid unless necessary

// Use type aliases for clarity
type UserID string
type Timestamp time.Time

// Struct tags for serialization
type User struct {
    ID        int       `json:"id" db:"id"`
    Username  string    `json:"username" db:"username"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

### Comments
```go
// Package comment starts with "Package name ..."
// Package user provides user management functionality.
package user

// Function comment starts with function name
// GetUserByID retrieves a user by their ID.
func GetUserByID(id int) (*User, error) {
    // ...
}

// Exported types need comments
type UserService struct {
    // ...
}
```

### Testing
```go
// Test file: package_test.go
func TestFunctionName(t *testing.T) {
    // Arrange
    input := "test"
    expected := "result"

    // Act
    actual, err := Function(input)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if actual != expected {
        t.Errorf("expected %q, got %q", expected, actual)
    }
}

// Table-driven tests
func TestCalculate(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"positive", 5, 10},
        {"zero", 0, 0},
        {"negative", -3, -6},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Calculate(tt.input)
            if got != tt.expected {
                t.Errorf("Calculate(%d) = %d, want %d", tt.input, got, tt.expected)
            }
        })
    }
}

// Benchmarks
func BenchmarkFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Function("input")
    }
}
```

## Project Structure (Recommended)
```
.
├── cmd/                    # Main applications
│   └── app/
│       └── main.go
├── internal/               # Private application code
│   ├── config/
│   ├── db/
│   └── service/
├── pkg/                    # Public library code
│   └── utils/
├── api/                    # API definitions (OpenAPI, proto)
├── web/                    # Web assets
├── configs/                # Configuration files
├── scripts/                # Build scripts
├── docs/                   # Documentation
├── test/                   # Additional test data
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── AGENTS.md
```

## Common Patterns

### Context Usage
```go
// Always accept context as first parameter
func DoWork(ctx context.Context, param string) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    // ...
}
```

### Defer Usage
```go
// Close resources with defer immediately after opening
f, err := os.Open("file.txt")
if err != nil {
    return err
}
defer f.Close()
```

### Constructor Pattern
```go
// New functions for complex types
func NewUserService(repo UserRepository) *UserService {
    return &UserService{
        repo: repo,
    }
}
```

## CI/CD Commands
```bash
# Full CI pipeline
go mod verify
go build ./...
go vet ./...
go test -race -cover ./...
go fmt ./...
```
