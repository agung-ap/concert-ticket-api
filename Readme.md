# Concert Ticket Reservation API

A high-performance API service for concert ticket reservations built with Go, PostgreSQL, and gRPC/REST.

## Features

- Search available concerts with filtering options
- Book concert tickets within specified time windows
- Handle high concurrency (500+ users per second)
- Prevent race conditions and overbooking
- Support both REST and gRPC interfaces
- Database transaction management
- Comprehensive test coverage

## Technical Overview

This service is built with the following components:

- **API Layer**: Supports both REST (using Gin) and gRPC interfaces
- **Service Layer**: Contains business logic and coordination
- **Repository Layer**: Manages data access to the database
- **Database**: PostgreSQL with proper indexing and transaction management

### Key Technical Approaches

#### Handling Concurrent Bookings

To handle 500+ concurrent users trying to book tickets without overbooking:

1. **Optimistic Locking**: Uses version field to detect and handle concurrent modifications
2. **Database Transactions**: Ensures atomicity for booking operations
3. **Row-Level Locking**: Uses SELECT FOR UPDATE to prevent race conditions
4. **Retry Mechanism**: Implements automatic retries for concurrent booking conflicts

#### Preventing Overbooking

1. **Atomic Updates**: Uses SQL conditions to check available tickets during updates
2. **Transaction Isolation**: Uses proper transaction isolation levels
3. **Validation**: Multiple validation layers (service, repository, database constraints)

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌───────────────┐     ┌────────────────┐
│  Client App │────▶│  API Layer  │────▶│ Service Layer │────▶│ Repository     │
│             │     │ REST / gRPC │     │               │     │ Layer          │
└─────────────┘     └─────────────┘     └───────────────┘     └────────┬───────┘
                                                                       │
                                                                       ▼
                                                              ┌────────────────┐
                                                              │   PostgreSQL   │
                                                              │   Database     │
                                                              └────────────────┘
```

## Database Schema

### Concerts Table
```sql
CREATE TABLE concerts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    artist VARCHAR(255) NOT NULL,
    venue VARCHAR(255) NOT NULL,
    concert_date TIMESTAMP NOT NULL,
    total_tickets INT NOT NULL,
    available_tickets INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    booking_start_time TIMESTAMP NOT NULL,
    booking_end_time TIMESTAMP NOT NULL,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Bookings Table
```sql
CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    concert_id INT NOT NULL REFERENCES concerts(id),
    user_id VARCHAR(255) NOT NULL,
    ticket_count INT NOT NULL,
    booking_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'confirmed',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_ticket_count CHECK (ticket_count > 0)
);
```

## API Endpoints

### REST API

#### Concerts
- `GET /api/v1/concerts` - List concerts with filtering and pagination
- `GET /api/v1/concerts/:id` - Get a specific concert
- `POST /api/v1/concerts` - Create a new concert
- `PUT /api/v1/concerts/:id` - Update a concert

#### Bookings
- `POST /api/v1/bookings` - Book tickets for a concert
- `GET /api/v1/bookings/:id` - Get a specific booking
- `GET /api/v1/bookings?userID=123` - Get bookings for a user
- `POST /api/v1/bookings/:id/cancel` - Cancel a booking

### gRPC API

The service also provides a gRPC API with the following methods:

#### ConcertService
- `GetConcert`
- `ListConcerts`
- `CreateConcert`
- `UpdateConcert`

#### BookingService
- `GetBooking`
- `GetUserBookings`
- `BookTickets`
- `CancelBooking`

## Getting Started

### Prerequisites

- Go 1.18 or higher (for local development)
- PostgreSQL 12 or higher (for local development)
- Docker and Docker Compose (for containerized deployment)

### Local Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/concert-ticket-api.git
   cd concert-ticket-api
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up the database:
   ```bash
   createdb concert_tickets
   ```

4. Run database migrations:
   ```bash
   go run cmd/server/main.go -migrate
   ```

5. Configure the application:
   ```bash
   cp config/config.example.yaml config/config.yaml
   # Edit config.yaml with your settings
   ```

6. Start the server:
   ```bash
   go run cmd/server/main.go
   ```

### Docker Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/concert-ticket-api.git
   cd concert-ticket-api
   ```

2. Build and start the containers:
   ```bash
   docker-compose up -d
   ```

   This will:
    - Build the application Docker image
    - Start the PostgreSQL container
    - Start the application container
    - Run database migrations automatically
    - Connect the application to the database

3. Check the container status:
   ```bash
   docker-compose ps
   ```

4. View application logs:
   ```bash
   docker-compose logs -f app
   ```

5. Access the API:
    - REST API: http://localhost:8080
    - gRPC: localhost:50051

6. Test the API with curl:
   ```bash
   # Check health endpoint
   curl http://localhost:8080/health
   
   # List concerts
   curl http://localhost:8080/api/v1/concerts
   
   # Create a concert (replace the dates with future dates)
   curl -X POST http://localhost:8080/api/v1/concerts \
     -H "Content-Type: application/json" \
     -d '{
       "name": "Rock Concert",
       "artist": "Band Name",
       "venue": "Concert Hall",
       "concert_date": "2025-06-01T20:00:00Z",
       "total_tickets": 1000,
       "price": 50.0,
       "booking_start_time": "2025-05-01T10:00:00Z",
       "booking_end_time": "2025-05-01T10:20:00Z"
     }'
   ```

7. For load testing (from inside the Docker network):
   ```bash
   # Build a custom Docker image with test tools
   docker build -f Dockerfile.test -t concert-load-test .
   
   # Run the load test
   docker run --network concert-network concert-load-test
   ```

8. Stop and remove the containers:
   ```bash
   docker-compose down
   ```

   To remove volumes as well (this will delete all data):
   ```bash
   docker-compose down -v
   ```

### Docker Environment Variables

When running with Docker, you can customize the application by setting environment variables:

| Environment Variable          | Description                  | Default Value      |
|-------------------------------|------------------------------|-------------------|
| APP_LOG_LEVEL                 | Logging level                | info              |
| APP_REST_PORT                 | REST API port                | 8080              |
| APP_GRPC_PORT                 | gRPC port                    | 50051             |
| APP_MAX_RETRIES               | Max retries for booking      | 3                 |
| APP_DATABASE_HOST             | Database hostname            | db                |
| APP_DATABASE_PORT             | Database port                | 5432              |
| APP_DATABASE_USERNAME         | Database username            | postgres          |
| APP_DATABASE_PASSWORD         | Database password            | postgres          |
| APP_DATABASE_NAME             | Database name                | concert_tickets   |
| APP_DATABASE_SSLMODE          | Database SSL mode            | disable           |

Example:
```bash
APP_LOG_LEVEL=debug APP_REST_PORT=9090 docker-compose up -d
```

### Testing

Run the test suite:
```bash
go test ./...
```

Run integration tests:
```bash
go test ./test/integration/...
```

Run load tests:
```bash
go test ./test/load/...
```

## Design Decisions

### Optimistic vs. Pessimistic Locking

We use a hybrid approach with optimistic locking for general concert updates and row-level locking (SELECT FOR UPDATE) within transactions for booking operations. This provides a good balance between performance and data integrity.

### Transaction Management

Booking operations use database transactions to ensure that ticket count updates and booking creation are atomic. This prevents scenarios where tickets could be deducted but the booking not created, or vice versa.

### Retry Mechanism

The booking service implements an automatic retry mechanism for handling concurrent booking attempts. This helps to resolve temporary conflicts without requiring client-side retries.

### Database Isolation Level

We use the default PostgreSQL transaction isolation level (Read Committed) which provides a good balance between consistency and performance. For especially high-concurrency scenarios, you might consider using Serializable isolation, but be aware of the performance trade-offs.

## Performance Considerations

This service is designed to handle high concurrency with the following optimizations:

1. Connection pooling for database access
2. Efficient indexing strategy
3. Optimistic locking to reduce contention
4. Rate limiting middleware to protect against overload
5. Caching opportunities (not implemented, but prepared for in the architecture)

## Kubernetes Deployment

The application can also be deployed to a Kubernetes cluster using the provided configuration files.

### Prerequisites

- A running Kubernetes cluster
- kubectl configured to interact with your cluster
- Docker registry to store your images
- set of kubernetes config (we don't have it now)

### Deployment Steps

1. Build and push the Docker image to your registry:
   ```bash
   docker build -t your-registry/concert-ticket-api:latest .
   docker push your-registry/concert-ticket-api:latest
   ```

2. Update the image reference in `kubernetes/concert-api-deployment.yaml`:
   ```yaml
   image: your-registry/concert-ticket-api:latest
   ```

3. Deploy the PostgreSQL database:
   ```bash
   kubectl apply -f kubernetes/postgres-deployment.yaml
   ```

4. Deploy the Concert Ticket API:
   ```bash
   kubectl apply -f kubernetes/concert-api-deployment.yaml
   ```

5. Verify the deployment:
   ```bash
   kubectl get pods
   kubectl get services
   ```

6. For production deployments, consider updating:
    - The Secret for database credentials (use a more secure password)
    - The Ingress hostname and TLS configuration
    - Resource limits based on your workload
    - Horizontal Pod Autoscaler for automatic scaling

### Scaling Considerations

The application is designed to be horizontally scalable. When running multiple instances:

1. Database connection pooling helps manage the connection load
2. Optimistic locking prevents race conditions between API instances
3. Stateless design allows for easy scaling and load balancing

You can scale the deployment with:
```bash
kubectl scale deployment concert-ticket-api --replicas=5
```

Or set up an Horizontal Pod Autoscaler:
```bash
kubectl autoscale deployment concert-ticket-api --cpu-percent=70 --min=3 --max=10
```

## Future Improvements

- Distributed locking using Redis for multi-instance deployments
- Caching layer for frequently accessed concerts
- Event-driven architecture for notification systems
- Advanced monitoring and observability with Prometheus and Grafana
- CI/CD pipeline integration for automated testing and deployment