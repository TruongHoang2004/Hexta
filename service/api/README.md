# Ngoc Lam ZMP Backend

Backend service for the Ngoc Lam ZMP application, built with Go and Gin.

## 🛠 Tech Stack

- **Language**: [Go](https://go.dev/) (v1.24)
- **Framework**: [Gin](https://github.com/gin-gonic/gin)
- **Database**: PostgreSQL
- **ORM**: [Gorm](https://gorm.io/)
- **Dependency Injection**: [Uber Fx](https://github.com/uber-go/fx)
- **Migrations**: [Atlas](https://atlasgo.io/)
- **Logging**: [Zap](https://github.com/uber-go/zap)

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.24 or higher
- [Docker](https://www.docker.com/) & Docker Compose
- [Make](https://www.gnu.org/software/make/)

### Installation

1. **Clone the repository**

   ```bash
   git clone https://gitlab.com/ecommercehub1/api.git
   cd ngoclam-zmp-backend
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Environment Configuration**
   Copy the example environment file (if available) or create a `.env` file:
   ```bash
   cp .env.example .env
   ```
   _Note: Update `.env` with your database credentials and other configurations._

### Running the Application

**Development Mode**
Run the application with hot-reload (if configured) or standard dev tags:

```bash
make dev
```

**Production Build**
Build and run the binary:

```bash
make build
make run
```

**Using Docker**
Start the database and application using Docker Compose:

```bash
docker-compose up -d
```

## 🗄 Database Migrations

This project uses **Atlas** for database migrations with Gorm.

| Command               | Description                                        |
| :-------------------- | :------------------------------------------------- |
| `make migrate-init`   | Initialize a new migration file                    |
| `make migrate-diff`   | Generate a migration diff based on code changes    |
| `make migrate-apply`  | Apply pending migrations to the database           |
| `make migrate-status` | Check the status of migrations                     |
| `make migrate-prod`   | Apply migrations to the production database (Neon) |

## 📂 Project Structure

```
.
├── cmd/                # Application entry point
├── config/             # Configuration loading
├── internal/           # Private application code
│   ├── bootstrap/      # App initialization (DI, Server, DB)
│   ├── common/         # Shared utilities
│   ├── infrastructure/ # Infrastructure layer
│   ├── present/        # Presentation layer (Controllers, HTTP)
│   └── services/       # Business logic
├── bin/                # Compiled binaries
├── docker-compose.yml  # Docker services
└── Makefile            # Build and utility commands
```

## 📝 License

[MIT](LICENSE)
