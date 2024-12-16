# 🤖 Prompt-Ops

**PromptOps** is a CLI tool that lets you interact with your infrastructure—like Kubernetes clusters, databases, and cloud environments—using natural language. Instead of memorizing commands and juggling multiple tools, you can simply “ask” for what you need in plain English, and PromptOps will translate your request into the right kubectl commands, SQL queries, or cloud CLI instructions. With built-in interactive flows, suggested next steps, and support for various connection types, PromptOps streamlines operations, saves time, and makes managing complex systems more intuitive.

## 🛠️ Local Setup

- `export GO111MODULE=on && go mod tidy && go build -ldflags="-s -w" -o pops && ./pops --help`

## 🎯 Available Commands

### 🌍 General Connection Management

- `pops connection create`: Interactive creation of a new connection (asks for type).
- `pops connection list`: Lists all connections.
- `pops connection open`: Interactive selection and opening of a connection.
- `pops connection open [conn-name]`: Opens a specific connection by name.
- `pops connection delete`: Interactive selection and deletion of a connection.
- `pops connection delete [conn-name]`: Deletes a specific connection by name.
- `pops connection types`: Lists all available connection types (e.g., cloud, kubernetes, db).

### 🌥️ Cloud Connections

- `pops connection cloud create`: Interactive creation of a cloud connection.
- `pops connection cloud list`: Lists all cloud connections.
- `pops connection cloud open`: Interactive selection and opening of a cloud connection.
- `pops connection cloud open [conn-name]`: Opens a specific cloud connection.
- `pops connection cloud delete`: Interactive selection and deletion of a cloud connection.
- `pops connection cloud delete [conn-name]`: Deletes a specific cloud connection.
- `pops connection cloud types`: Lists available cloud providers.

### 🚆 Kubernetes Connections

- `pops connection kubernetes create`: Interactive creation of a Kubernetes connection.
- `pops connection kubernetes list`: Lists all Kubernetes connections.
- `pops connection kubernetes open`: Interactive selection and opening of a Kubernetes connection.
- `pops connection kubernetes open [conn-name]`: Opens a specific Kubernetes connection.
- `pops connection kubernetes delete`: Interactive selection and deletion of a Kubernetes connection.
- `pops connection kubernetes delete [conn-name]`: Deletes a specific Kubernetes connection.
- `pops connection kubernetes types`: Lists available Kubernetes connection types.

### 💿 Database Connections

- `pops connection db create`: Interactive creation of a database connection.
- `pops connection db list`: Lists all database connections.
- `pops connection db open`: Interactive selection and opening of a database connection.
- `pops connection db open [conn-name]`: Opens a specific database connection.
- `pops connection db delete`: Interactive selection and deletion of a database connection.
- `pops connection db delete [conn-name]`: Deletes a specific database connection.
- `pops connection db types`: Lists available database connection types.

### 🔮 Coming Soon (Commands)

- `pops connection delete`
- `pops connection mq` or `pops connection queue` for Message Queues.
- `pops connection storage` for Object Storage.
- `pops connection monitoring` or `pops connection logging` for Monitoring & Logging.
- `pops connection cicd` for CI/CD Pipelines.
- `pops connection cache` or `pops connection kv` for Cache & Key-Value Stores.
- `pops session` to keep track of all prompts and commands...

## 〄 Connection Types

### ⏚ Available Now

- Kubernetes
- DB
  - PostgreSQL
  - MySQL
  - MongoDB
- Cloud
  - Azure

### 🔮 Coming Soon (Types)

- Cloud
  - AWS
  - GCP
- Message Queues
  - Kafka
  - RabbitMQ
  - AWS SQS
- Object Storage
  - AWS S3
  - GCP Storage
  - Azure Blob Storage
- Monitoring & Logging
  - Prometheus
  - Elasticsearch
  - Datadog
  - Splunk
- CI/CD Pipelines
  - Jenkins
  - GitLab CI
  - GitHub Actions
  - CircleCI
- Cache & Key-Value Stores
  - Redis
  - Memcached
