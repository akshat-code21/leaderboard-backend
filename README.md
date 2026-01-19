# Matiks Leaderboard Backend

A high-performance, scalable leaderboard API built with Go (Golang), featuring real-time ranking updates, efficient search capabilities, and tie-aware ranking algorithms.

## 🚀 Features

- **Tie-Aware Ranking**: Users with the same rating share the same rank
- **Efficient Search**: Fast username search with pagination support
- **Real-time Updates**: Background workers for non-blocking score updates
- **Redis Caching**: Optimized leaderboard queries using Redis Sorted Sets
- **PostgreSQL**: Robust data persistence with GORM
- **RESTful API**: Clean, well-structured API endpoints
- **Auto-sync**: Automatic data synchronization between PostgreSQL and Redis

## 📋 Prerequisites

- Go 1.23.3 or higher
- PostgreSQL 12+ 
- Redis 6+ (optional but recommended for performance)
- Environment variables configured (see Configuration section)

## 🛠️ Tech Stack

- **Framework**: Gin (HTTP web framework)
- **ORM**: GORM
- **Database**: PostgreSQL
- **Cache**: Redis (Sorted Sets for leaderboard)
- **Architecture**: Layered (Repository → Service → Controller → Handler)

## 📁 Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── config/
│   │   ├── database.go      # PostgreSQL connection
│   │   ├── migrate.go       # Database migrations
│   │   └── redis.go         # Redis connection
│   ├── models/
│   │   └── models.go        # Data models
│   ├── repository/
│   │   ├── repository.go   # PostgreSQL operations
│   │   └── redis_repository.go  # Redis operations
│   ├── service/
│   │   ├── leaderboard_service.go  # Leaderboard business logic
│   │   ├── user_service.go        # User search & rank logic
│   │   └── update_service.go      # Background update workers
│   ├── controllers/
│   │   ├── leaderboard_controller.go
│   │   ├── user_controller.go
│   │   └── update_controller.go
│   └── handlers/
│       ├── leaderboard_handler.go  # HTTP handlers
│       ├── user_handler.go
│       └── update_handler.go
├── scripts/
│   └── seed.go              # Database seeding script
├── go.mod
└── README.md
```

## ⚙️ Configuration

Create a `.env` file in the root directory:

```env
# Database Configuration
DATABASE_URL=postgres://username:password@localhost:5432/leaderboard_db?sslmode=disable

# Redis Configuration (optional but recommended)
REDIS_URL=redis://localhost:6379/0

# Server Configuration
PORT=8080
```

### Environment Variables

- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection URL (optional - app will run without Redis but with reduced performance)
- `PORT`: Server port (default: 8080)

## 🚀 Getting Started

### 1. Clone and Install Dependencies

```bash
cd backend
go mod download
```

### 2. Set Up Database

```bash
# Create PostgreSQL database
createdb leaderboard_db

# Or using psql
psql -U postgres
CREATE DATABASE leaderboard_db;
```

### 3. Configure Environment

Create a `.env` file with your database and Redis credentials (see Configuration section above).

### 4. Run Migrations

Migrations run automatically on server startup. The `AutoMigrate` function creates:
- `users` table with indexes on `username` and `rating`

### 5. Seed Database (Optional)

```bash
go run scripts/seed.go
```

This will create 10,000 users with random ratings between 100-5000.

### 6. Start the Server

```bash
go run cmd/server/main.go
```

Or build and run:

```bash
go build -o server ./cmd/server
./server
```

The server will start on `http://localhost:8080` (or your configured PORT).

## 📡 API Endpoints

### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected"
}
```

### Leaderboard

```http
GET /api/v1/leaderboard?page=1&limit=50
```

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Results per page (default: 50, max: 100)

**Response:**
```json
{
  "entries": [
    {
      "rank": 1,
      "username": "user_1",
      "rating": 5000
    }
  ],
  "page": 1,
  "limit": 50,
  "total": 10000
}
```

### Search Users

```http
GET /api/v1/users/search?q=user&page=1&limit=50
```

**Query Parameters:**
- `q` (required): Search query (case-insensitive)
- `page` (optional): Page number (default: 1)
- `limit` (optional): Results per page (default: 50, max: 100)

**Response:**
```json
{
  "users": [
    {
      "rank": 45,
      "username": "user_123",
      "rating": 3500
    }
  ],
  "page": 1,
  "limit": 50,
  "total": 150
}
```

### Get User Rank

```http
GET /api/v1/users/:username/rank
```

**Response:**
```json
{
  "username": "user_123",
  "rating": 3500,
  "rank": 45
}
```

### Admin Endpoints

#### Sync Redis

```http
POST /api/v1/admin/sync-redis
```

Manually sync all users from PostgreSQL to Redis.

**Response:**
```json
{
  "message": "Redis synced successfully",
  "count": 10000
}
```

#### Simulate Updates

```http
POST /api/v1/admin/simulate-updates?count=10
```

Trigger random rating updates for testing.

**Query Parameters:**
- `count` (optional): Number of users to update (default: 10)

**Response:**
```json
{
  "message": "Updates queued successfully",
  "count": 10
}
```

## 🎯 Ranking Algorithm

The system uses **tie-aware ranking**:

- Users with the same rating share the same rank
- The next rank skips appropriately
- Example:
  - Rank 1: Rating 5000
  - Rank 2: Rating 4999
  - Rank 3: Rating 4998, 4998 (tied)
  - Rank 5: Rating 4997 (skips rank 4)

## ⚡ Performance Optimizations

1. **Redis Sorted Sets**: O(log N) leaderboard queries
2. **Database Indexes**: B-tree indexes on `username` and `rating DESC`
3. **Background Workers**: Non-blocking score updates using goroutines
4. **Batch Operations**: Efficient Redis sync with batching
5. **Connection Pooling**: GORM connection pool for database

## 🔄 Background Updates

The system includes a background update service that:

- Processes updates asynchronously using worker goroutines
- Automatically syncs updates to both PostgreSQL and Redis
- Runs scheduled random updates every 5 minutes (configurable)
- Supports manual trigger via admin endpoint

## 🧪 Testing

### Using cURL

```bash
# Health check
curl http://localhost:8080/health

# Get leaderboard
curl "http://localhost:8080/api/v1/leaderboard?page=1&limit=10"

# Search users
curl "http://localhost:8080/api/v1/users/search?q=user&page=1&limit=10"

# Get user rank
curl http://localhost:8080/api/v1/users/user_123/rank

# Simulate updates
curl -X POST "http://localhost:8080/api/v1/admin/simulate-updates?count=5"

# Sync Redis
curl -X POST http://localhost:8080/api/v1/admin/sync-redis
```

## 🚢 Deployment

### Render.com

1. Create a new Web Service
2. Connect your GitHub repository
3. Set build command: `go build -o server ./cmd/server`
4. Set start command: `./server`
5. Add environment variables:
   - `DATABASE_URL`
   - `REDIS_URL`
   - `PORT` (optional, Render sets this automatically)

### Environment Variables on Render

```
DATABASE_URL=postgres://user:pass@host:5432/dbname
REDIS_URL=rediss://user:pass@host:6379/0
PORT=8080
```

## 📊 Database Schema

### Users Table

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    rating INTEGER NOT NULL CHECK (rating >= 100 AND rating <= 5000),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_rating ON users(rating DESC);
```