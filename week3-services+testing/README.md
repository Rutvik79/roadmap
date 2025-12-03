## User Management API

A production-ready RESTful API built with Go and Gin framwork, featuring JWT authentication, pagination, filtering, and comprehensive testing.

## Features

- JWT Authentication (Register/Login)
- Complete CRUD Operations
- Pagination Support
- Filtering and Sorting
- Rate Limiting
- Request Validation
- Comprehensive Testing (70%+ coverage)
- Custom Middleware
- Error Handling

## Tech Stack

- **Framework:** Gin
- **Authentication:** JWT (golang-jwt)
- **Validation:** go-playground/validator
- **Testing:** testify
- **Password Hashing:** bcrypt

## API Endpoints

### Authentication

#### Register

```http
POST /api/v1/auth/register
Content-Type: application/json

{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123",
    "age": 30,
}
```

#### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{
    "email": "john@example.com",
    "password": "password123"
}
```

### Users (Protected)

#### Get All Users (with pagination)

```http
GET /api/v1/users?page=1&page_size=10&search=john&sort_by=name&order=asc
Authorization: Bearer
```

#### Get User by ID

```http
GET /api/v1/users/:id
Authorization: Bearer
```

#### Create User

```http
POST /api/v1/users
Authorization: Bearer
Content-Type: application/json

{
    "name": "Jane Doe",
    "email": "jane@example.com",
    "age": 28
}
```

#### Update User

```http
PUT /api/v1/users/:id
Authorization: Bearer
Content-Type: application/json

{
    "name": "Jane Smith",
    "email": "jane.smith@example.com",
    "age": 29
}
```

#### Delete User

```http
DELETE /api/v1/users/
Authorization: Bearer
```

## Query Parameters

### Pagination

- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 10, max: 100)

### Filtering

- `search` - Search by name or email
- `min_age` - Minimum age filter
- `max_age` - Maximum age filter

### Sorting

- `sort_by` - Field to sort by (name, email, age)
- `order` - Sort order(asc, desc)

## Installation

```bash
# Clone repository
git clone https://github.com/Rutvik79/go-learning-journey/week3-services+testing/api.git

# Install dependencies
go mod download

# Run server
go run api/main.go
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

# Project Struct

```
week3-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── auth/
│   │   ├── jwt.go
│   │   ├── jwt_test.go
│   │   ├── password.go
│   │   └── password_test.go
│   ├── handlers/
│   │   ├── auth.go
│   │   ├── auth_test.go
│   │   ├── user.go
│   │   └── user_test.go
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── auth_test.go
│   │   ├── cors.go
│   │   ├── errors.go
│   │   ├── logger.go
│   │   ├── ratelimit.go
│   │   └── recovery.go
│   └── models/
│       ├── auth.go
│       ├── filter.go
│       ├── pagination.go
│       └── user.go
├── go.mod
├── go.sum
└── README.md
```

## Environment Variables

Create a `.env` file:

```env
PORT=8080
JWT_SECRET=your-secret-key
```

## License

MIT
