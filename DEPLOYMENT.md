# Production Deployment Guide

This guide covers deploying Quards to production with proper database setup.

## Prerequisites

- PostgreSQL database server
- Go 1.21+ (for building from source)

## Quick Start

1. **Clone and build the application:**
   ```bash
   git clone <repository-url>
   cd quards
   make build
   ```

2. **Set up environment variables:**
   ```bash
   export DATABASE_URL="postgres://username:password@host:5432/database?sslmode=require"
   export PORT="8080"
   export HOST="0.0.0.0"
   export ENVIRONMENT="production"
   ```

3. **Run database migrations:**
   ```bash
   ./bin/migrate
   ```

4. **Start the application:**
   ```bash
   ./bin/quards
   ```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://quards:quards@localhost/quards_db?sslmode=disable` | Yes |
| `PORT` | Server port | `8080` | No |
| `HOST` | Server host | `localhost` | No |
| `ENVIRONMENT` | Application environment | `development` | No |

## Database Setup

### New Database Setup

1. Create a PostgreSQL database:
   ```sql
   CREATE DATABASE your_app_db;
   CREATE USER your_app_user WITH PASSWORD 'secure_password';
   GRANT ALL PRIVILEGES ON DATABASE your_app_db TO your_app_user;
   ```

2. Set the DATABASE_URL:
   ```bash
   export DATABASE_URL="postgres://your_app_user:secure_password@localhost:5432/your_app_db?sslmode=require"
   ```

3. Run migrations:
   ```bash
   ./bin/migrate
   ```

### Migration Management

The migration system:
- Automatically tracks applied migrations
- Only runs each migration once
- Supports rollback via transactions
- Creates necessary tables and indexes

Available commands:
```bash
# Run all pending migrations
./bin/migrate

# Show migration help
./bin/migrate -h

# Using make (during development)
make migrate
```

## Docker Deployment (Optional)

Create a `Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o bin/quards main.go
RUN go build -o bin/migrate cmd/migrate/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/ ./
COPY --from=builder /app/migrations/ ./migrations/
COPY --from=builder /app/static/ ./static/
CMD ["./quards"]
```

Build and run:
```bash
docker build -t quards .
docker run -p 8080:8080 -e DATABASE_URL="your_db_url" quards
```

## Security Considerations

1. **Database Security:**
   - Use strong passwords
   - Enable SSL/TLS (`sslmode=require`)
   - Restrict database access by IP
   - Use connection pooling

2. **Application Security:**
   - Set `ENVIRONMENT=production`
   - Use reverse proxy (nginx/Apache)
   - Enable HTTPS
   - Set appropriate firewall rules

3. **Environment Variables:**
   - Never commit `.env` files
   - Use secure secret management
   - Rotate credentials regularly

## Health Checks

The application provides basic health monitoring:
- API endpoint: `GET /api/decks` (returns 200 if database is accessible)
- Database connectivity is verified on startup

## Monitoring

Consider adding:
- Log aggregation (ELK stack, Splunk)
- Application metrics (Prometheus + Grafana)
- Database monitoring
- Uptime monitoring

## Backup Strategy

1. **Database Backups:**
   ```bash
   # Create backup
   pg_dump -h host -U user -d database > backup.sql
   
   # Restore backup
   psql -h host -U user -d database < backup.sql
   ```

2. **Application Data:**
   - Static files in `/static/`
   - Configuration files
   - Migration files

## Troubleshooting

### Common Issues

1. **Database Connection Failed:**
   - Verify DATABASE_URL format
   - Check database server status
   - Verify network connectivity
   - Check firewall settings

2. **Migration Failed:**
   - Check database permissions
   - Verify migration file syntax
   - Check migration logs
   - Manually inspect `schema_migrations` table

3. **Application Won't Start:**
   - Check environment variables
   - Verify binary permissions
   - Check port availability
   - Review application logs

### Debug Commands

```bash
# Check database connectivity
./bin/migrate -h

# Test API endpoints
curl http://localhost:8080/api/decks

# Check application status
ps aux | grep quards
```

## Support

For issues:
1. Check logs for error messages
2. Verify all environment variables are set
3. Test database connectivity separately
4. Review this deployment guide