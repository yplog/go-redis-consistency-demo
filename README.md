# Go Redis Consistency Demo

A Go-based demo comparing Strong Consistency and Eventual Consistency using Redis replication. 

This demo showcases two different consistency models:

- **Strong Consistency**: Ensures data is up-to-date across all replicas.
- **Eventual Consistency**: Guarantees data will be consistent across replicas over time.

## Redis Setup (Using Docker Compose)

```bash
# Start Redis containers
docker compose up

# Verify replication status
docker compose exec redis-replica redis-cli info replication
```

### Running the Go Application

For Eventual Consistency mode:
```bash
go run src/main.go -mode=eventual
```

For Strong Consistency mode:
```bash
go run src/main.go -mode=strong
```

## API Endpoints

### Increment Counter
```bash
curl http://localhost:8080/increment
```

### Get Counter Value
```bash
curl http://localhost:8080/get
```

## Stopping Services

```bash
docker compose down
```