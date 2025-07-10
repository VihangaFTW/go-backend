# SimpleBank - Go Backend API

A production-ready banking backend system built with Go, implementing secure user authentication, multi-currency account management, and money transfers with modern banking practices.

> **Note**: This project is based on the excellent [Backend Master Class course by Tech School](https://www.udemy.com/course/backend-master-class-golang-postgresql-kubernetes/) on Udemy. It serves as my hands-on introduction to backend development and DevOps practices using Go. While following the course structure, I have taken the liberty to add changes and new features where I felt like experimenting and learning beyond the original scope. The original course repository can be found at: https://github.com/techschool/simplebank

## 🏗️ Architecture

- **Language**: Go 1.24.4
- **Database**: PostgreSQL with SQLC for type-safe queries
- **APIs**: gRPC + gRPC-Gateway for HTTP/JSON translation
- **Authentication**: PASETO tokens with refresh token support
- **Documentation**: Embedded Swagger UI
- **Deployment**: Docker + Kubernetes (AWS EKS)
- **CI/CD**: GitHub Actions with AWS ECR

## 🚀 Features

### Core Banking Operations

- **User Management**: Secure registration and login with bcrypt password hashing
- **Multi-Currency Accounts**: Support for USD, EUR, CAD with balance tracking
- **Money Transfers**: Atomic transfers between accounts with double-entry bookkeeping
- **Session Management**: Secure session handling with refresh tokens

### Security & Authentication

- PASETO & JWT token-based authentication
- Authorization middleware for protected endpoints
- Client metadata tracking (IP, User-Agent)
- Session blocking capabilities
- Input validation with custom validators

### API Design

- Dual protocol support (gRPC + REST)
- OpenAPI/Swagger documentation
- Comprehensive error handling
- Type-safe database operations

## 🛠️ Tech Stack

| Component         | Technology     |
| ----------------- | -------------- |
| Backend           | Go 1.24.4      |
| Database          | PostgreSQL     |
| ORM/Query Builder | SQLC           |
| API Framework     | gRPC + Gin     |
| Authentication    | PASETO         |
| Documentation     | Swagger UI     |
| Containerization  | Docker         |
| Orchestration     | Kubernetes     |
| Cloud Platform    | AWS EKS        |
| CI/CD             | GitHub Actions |

## 🗄️ Database Schema

```
Users (username, hashed_password, full_name, email)
├── Accounts (id, owner, balance, currency)
│   ├── Entries (id, account_id, amount)
│   └── Transfers (id, from_account_id, to_account_id, amount)
└── Sessions (id, username, refresh_token, user_agent, client_ip)
```

## 🚦 Getting Started

### Prerequisites

- Go 1.24.4+
- PostgreSQL
- Docker (optional)
- Make

### Local Development

1. **Clone the repository**

   ```bash
   git clone https://github.com/VihangaFTW/Go-Backend.git
   cd Go-Backend
   ```

2. **Start PostgreSQL database**

   ```bash
   make startdb
   make createdb
   ```

3. **Run database migrations**

   ```bash
   make migrateup
   ```

4. **Set up environment variables**

   ```bash
   cp app.env.template app.env
   # Edit app.env with your configuration
   ```

5. **Start the server**
   ```bash
   make server
   ```

The API will be available at:

- gRPC: `localhost:9090`
- HTTP Gateway: `localhost:8080`
- Swagger UI: `localhost:8080/swagger/`

### Docker Development

```bash
docker-compose up
```

## 📋 Available Commands

```bash
# Database
make startdb          # Start PostgreSQL container
make createdb         # Create database
make dropdb          # Drop database
make migrateup       # Run migrations
make migratedown     # Rollback migrations

# Development
make server          # Start the server
make test           # Run tests
make sqlc           # Generate SQLC code
make mock           # Generate mocks

# Documentation
make db_docs        # Generate database docs
make db_schema      # Generate SQL schema
make proto          # Generate protobuf code
```

## 🔗 API Endpoints

### Authentication

- `POST /v1/create_user` - Register new user
- `POST /v1/login_user` - User login
- `POST /tokens/renew_access` - Refresh access token

### Accounts (Protected)

- `POST /accounts` - Create account
- `GET /accounts/:id` - Get account details
- `GET /accounts` - List user accounts

### Transfers (Protected)

- `POST /transfers` - Create money transfer

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific test
go test -v ./api/...
go test -v ./db/sqlc/...
```

## 🚀 Deployment

### AWS EKS Deployment

The project includes Kubernetes manifests for AWS EKS deployment:

```bash
# Deploy to EKS
kubectl apply -f eks/
```

### CI/CD Pipeline

GitHub Actions automatically:

- Runs tests on pull requests
- Builds and pushes Docker images to AWS ECR
- Deploys to AWS EKS on main branch

## 🔮 Roadmap

Planned enhancements include:

- **Async Processing**: Redis-based background workers for email notifications
- **Enhanced Security**: Role-based access control (RBAC) and CORS support
- **Performance**: Migration to pgx driver for better PostgreSQL performance
- **Monitoring**: Comprehensive logging and metrics collection

## 🎓 Learning Journey

This project represents my practical journey into backend development and DevOps. Through building this banking system, I've gained hands-on experience with:

- **Backend Development**: RESTful and gRPC APIs, database design, authentication systems
- **DevOps Practices**: Docker containerization, Kubernetes orchestration, CI/CD pipelines
- **Cloud Technologies**: AWS services (EKS, ECR, Secrets Manager)
- **Security**: Token-based authentication, secure password handling, session management
- **Testing**: Unit testing, integration testing, mocking strategies

## 🙏 Acknowledgments

Special thanks to [Tech School](https://github.com/techschool) for creating an outstanding course that provides real-world, production-ready examples. The course covers everything from basic Go concepts to advanced DevOps practices.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👨‍💻 Author

**Vihanga Malaviarachchi**

- GitHub: [@VihangaFTW](https://github.com/VihangaFTW)
- Email: vihaaanga.mihiranga@gmail.com

---

⭐ Star this repository if you found it helpful!
