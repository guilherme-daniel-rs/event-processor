# Event Processor (Go)

[![CI](https://github.com/guilherme-daniel-rs/event-processor/actions/workflows/ci.yml/badge.svg)](https://github.com/guilherme-daniel-rs/event-processor/actions/workflows/ci.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=guilherme-daniel-rs_event-processor&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=guilherme-daniel-rs_event-processor)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=guilherme-daniel-rs_event-processor&metric=coverage)](https://sonarcloud.io/summary/new_code?id=guilherme-daniel-rs_event-processor)

A Go worker built to process events from multiple sources (SQS) and persist them to a database (DynamoDB). The focus here was creating a resilient system with smart error handling that's easy to run locally.

![Architecture Diagram](diagram-c4.png)

## How it works

The main binary listens to an SQS queue. When a message arrives, it:
1. Validates the header (Tenant, Client, ID, etc.).
2. Uses an internal **Schema Registry** to identify the event type (`user.created`, `order.placed`, `payment.processed`).
3. Validates the event body using the specific version's schema.
4. Persists the data in DynamoDB with a "processed" status.

### Resilience (Retries & DLQ)
Processors need to handle failures gracefully:
- **Validation Errors:** If the JSON is broken or mandatory fields are missing, we `Ack` it immediately. There's no point in retrying something that will never pass validation.
- **Infrastructure Failures:** If the database is down or the network flickers, we use **Exponential Backoff**. The system waits for a delay that doubles with each attempt (30s, 60s, 120s...) up to a 5-minute limit.
- **DLQ:** If it still fails after X retries (default 5), we let the message go to the Dead Letter Queue for manual inspection.

---

## Getting Started

### Local Infrastructure
We use LocalStack to simulate AWS. To spin everything up:

```bash
# Start LocalStack
docker-compose up -d

# Create queues and tables using Terraform
cd infra/terraform
tflocal init
tflocal apply --auto-approve
cd ../..
```

### The App
There's a `Makefile` to handle daily commands:

```bash
# Build the project
make build

# Run the worker
make run

# Script to send a few test events to the queue
make send-events
or
go run cmd/send-events/main.go -count=100 -type=payment.processed
```

### Tests & Coverage
```bash
make test      # Runs unit tests
make coverage  # Generates the HTML coverage report
```

---

## Project Structure

```text
├── cmd/
│   ├── worker/         # Processor entrypoint
│   └── send-events/    # Helper to test the queue
├── internal/
│   ├── domain/         # Schemas and validations
│   ├── app/            # Main processing logic
│   ├── adapters/       # SQS and DynamoDB integrations
│   └── ports/          # Interfaces and error definitions
├── Dockerfile          # Multi-stage build (final image is scratch)
└── Makefile            # Command shortcuts
```

---

## Automation (CI/CD)
The project is configured with **GitHub Actions**:
- **Build/Test:** Every PR runs tests and checks the build.
- **SonarCloud:** Quality analysis and coverage checks.
- **Docker:** On push to `main`, it builds a new image and pushes it to **GHCR** (GitHub Container Registry).

---
Built with ❤️ by [guilherme-daniel-rs](https://github.com/guilherme-daniel-rs)