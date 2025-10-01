# SimpleBank - Go Backend API

Production-ready banking backend with secure authentication, multi-currency accounts, and atomic money transfers. Built with Go, PostgreSQL, gRPC, and deployed on AWS EKS.

> Based on [Backend Master Class by Tech School](https://www.udemy.com/course/backend-master-class-golang-postgresql-kubernetes/)

## 🏗️ Stack

**Backend**: Go 1.24.4 • PostgreSQL • SQLC • gRPC + REST  
**Auth**: PASETO tokens or JWTs with session management  
**Observability**: Structured logging (zerolog) for HTTP & gRPC  
**Infra**: Docker • Kubernetes (AWS EKS) • GitHub Actions CI/CD

## 🚀 Features

- User authentication with bcrypt hashing and refresh tokens
- Multi-currency accounts (USD, EUR, CAD) with atomic transfers
- Dual protocol support (gRPC + REST via gRPC-Gateway)
- Request/response logging with status codes, duration, and metadata
- Authorization middleware and input validation
- Swagger documentation at `/swagger/`

## 🗄️ Database Schema

```
Users (username, hashed_password, full_name, email)
├── Accounts (id, owner, balance, currency)
│   ├── Entries (id, account_id, amount)
│   └── Transfers (id, from_account_id, to_account_id, amount)
└── Sessions (id, username, refresh_token, user_agent, client_ip)
```

## 📊 Logging

Structured logs using **zerolog** for both HTTP and gRPC:

- **HTTP**: Middleware logs method, path, status, duration, error bodies
- **gRPC**: Interceptor logs method, status, duration, errors
- **Format**: Pretty console (dev) with local time, JSON (prod) with UTC


## 🚦 Quick Start

```bash
# Clone and setup
git clone https://github.com/VihangaFTW/Go-Backend.git
cd Go-Backend

# Start database
make startdb createdb migrateup

# Configure and run
cp app.env.template app.env  # Edit with your config
make server

# Or use Docker
docker-compose up
```

**Endpoints**:

- gRPC: `localhost:9090`
- HTTP: `localhost:8080`
- Swagger: `localhost:8080/swagger/`

## 📋 Commands

```bash
make server          # Start server
make test            # Run tests
make migrateup       # Run migrations
make sqlc            # Generate SQLC code
make proto           # Generate protobuf code
```

## 🚀 Deployment

Deploys to AWS EKS via GitHub Actions. On push to main: runs tests → builds Docker image → pushes to ECR → deploys to Kubernetes.

```bash
kubectl apply -f eks/  # Manual deployment
```

## 👨‍💻 Author

**Vihanga Malaviarachchi**  
GitHub: [@VihangaFTW](https://github.com/VihangaFTW) • Email: vihaaanga.mihiranga@gmail.com

---

