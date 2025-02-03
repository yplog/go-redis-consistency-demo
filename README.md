# go-redis-consistency-demo

A Go-based demo comparing Strong Consistency and Eventual Consistency using Redis replication. 

## Install Redis With Docker Compose

```bash
# Start Redis containers
docker compose up -d

# Verify replication status
docker compose exec redis-replica redis-cli info replication
```

## Run Go Application with Eventual Consistency

```bash
go run src/main.go -mode=eventual
```

## Run Go Application with Strong Consistency

```bash
go run src/main.go -mode=strong
```

## Test

```bash
curl http://localhost:8080/increment
```

```bash
curl http://localhost:8080/get
```

## Stop Redis Containers

```bash
docker compose down
```