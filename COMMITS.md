# Semantic Commit Messages for Heisenberg Service

Aşağıdaki commit mesajlarını sırayla kullanabilirsiniz:

## 1. Model Katmanı

```bash
git add model/aircraft.go
git commit -m "feat(model): add aircraft model with MAC address support"
```

```bash
git add model/anomaly.go
git commit -m "feat(model): add anomaly detection model with type definitions"
```

```bash
git add model/geofence.go
git commit -m "feat(model): add geofence model with point containment logic"
```

```bash
git add model/threshold.go
git commit -m "feat(model): add threshold model for telemetry metrics"
```

```bash
git add model/telemetry.go
git commit -m "feat(model): add telemetry model for timeseries data storage"
```

```bash
git add model/telemetry_dto.go
git commit -m "feat(model): add telemetry DTO for ingestion service integration"
```

## 2. Repository Katmanı

```bash
git add repository/aircraft_repo.go
git commit -m "feat(repository): implement aircraft repository with MAC address lookup"
```

```bash
git add repository/geofence_repo.go
git commit -m "feat(repository): implement geofence repository for active geofence queries"
```

```bash
git add repository/threshold_repo.go
git commit -m "feat(repository): implement threshold repository with fallback to defaults"
```

```bash
git add repository/telemetry_repo.go
git commit -m "feat(repository): implement telemetry repository with batch insert support"
```

## 3. Service Katmanı

```bash
git add service/aircraft_service.go
git commit -m "feat(service): implement aircraft service for MAC address resolution"
```

```bash
git add service/threshold_service.go
git commit -m "feat(service): implement threshold service for metric violation detection"
```

```bash
git add service/geofence_service.go
git commit -m "feat(service): implement geofence service for restricted area detection"
```

```bash
git add service/anomaly_service.go
git commit -m "feat(service): implement anomaly service combining threshold and geofence checks"
```

```bash
git add service/worker_service.go
git commit -m "feat(service): implement worker service for telemetry processing pipeline"
```

## 4. Consumer ve Publisher

```bash
git add consumer/stream_consumer.go
git commit -m "feat(consumer): implement Redis stream consumer with concurrent processing"
```

```bash
git add publisher/feed_publisher.go
git commit -m "feat(publisher): implement feed publisher for global telemetry and alerts"
```

## 5. Package Utilities

```bash
git add pkg/constant/constant.go
git commit -m "chore(pkg): add constant package placeholder"
```

```bash
git add pkg/logging/logging.go
git commit -m "feat(pkg): implement structured logging with zap logger"
```

```bash
git add pkg/postgres/postgres.go
git commit -m "feat(pkg): implement PostgreSQL client with GORM and auto-migration"
```

```bash
git add pkg/redis/redis.go
git commit -m "feat(pkg): implement Redis client with stream and pub/sub support"
```

## 6. Main Application

```bash
git add cmd/main.go
git commit -m "feat(cmd): implement main application with worker service orchestration"
```

## 7. Deployment ve Build

```bash
git add Dockerfile
git commit -m "build: add multi-stage Dockerfile for heisenberg service"
```

```bash
git add deployment/deployment.yml
git commit -m "feat(deployment): add Kubernetes deployment configuration"
```

```bash
git add deployment/service.yml
git commit -m "feat(deployment): add Kubernetes service configuration"
```

## 8. Go Modules

```bash
git add go.mod go.sum
git commit -m "chore: add Go module dependencies"
```

## 9. Documentation

```bash
git add README.md
git commit -m "docs: add README for heisenberg service"
```

