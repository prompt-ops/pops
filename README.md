# 🤖 Prompt-Ops

**Prompt-Ops** is a CLI tool that makes managing your infrastructure—like Kubernetes clusters, databases, and cloud environments—effortless and intuitive. By translating natural language into precise commands, it eliminates the need to memorize complex syntax or juggle multiple tools.

With features like interactive flows, intelligent suggestions, and broad connection support, Prompt-Ops streamlines operations, saves time, and makes managing complex systems more approachable.

## 🚀 Key Features

- 🔍 **Natural Language Commands**: Interact with your services using plain English.
- ⚡ **Interactive Workflows**: Step-by-step prompts for setup and operations.
- 🌐 **Broad Compatibility**: Supports Kubernetes, databases, cloud services, and more.
- 🔮 **AI-Powered Suggestions**: Get guided next steps and smart command completions.

## 🛠️ Setup

To get started locally:

```bash
make install
```

## 📜 Available Commands

### 🌍 General

- `pops connection create`: Create a new connection interactively.
- `pops connection list`: List all connections.
- `pops connection open [conn-name]`: Open a specific connection.
- `pops connection delete [conn-name]`: Delete a specific connection.
- `pops connection types`: Show available connection types.

### 🌥️ Cloud

- `pops connection cloud create`: Create a cloud connection interactively.
- `pops connection cloud list`: List all cloud connections.
- `pops connection cloud open [conn-name]`: Open a specific cloud connection.
- `pops connection cloud delete [conn-name]`: Delete a specific cloud connection.
- `pops connection cloud types`: Show supported cloud providers.

### 🚆 Kubernetes

- `pops connection kubernetes create`: Create a Kubernetes connection.
- `pops connection kubernetes list`: List Kubernetes connections.
- `pops connection kubernetes open [conn-name]`: Open a specific Kubernetes connection.
- `pops connection kubernetes delete [conn-name]`: Delete a Kubernetes connection.

### 💿 Database

- `pops connection db create`: Create a database connection.
- `pops connection db list`: List database connections.
- `pops connection db open [conn-name]`: Open a specific database connection.
- `pops connection db delete [conn-name]`: Delete a database connection.
- `pops connection db types`: Show supported database types.

## 〄 Supported Connection Types

### Available Now

- **Kubernetes**
- **Databases**:
  - PostgreSQL
  - MySQL
  - MongoDB
- **Cloud**:
  - Azure

### Coming Soon

- **Cloud Providers**: AWS, GCP
- **Message Queues**: Kafka, RabbitMQ, AWS SQS
- **Object Storage**: AWS S3, Azure Blob, GCP Storage
- **Monitoring & Logging**: Prometheus, Elasticsearch, Datadog, Splunk
- **CI/CD**: Jenkins, GitLab CI, GitHub Actions, CircleCI
- **Cache**: Redis, Memcached

## 🎯 Planned Features

- **Message Queues**: pops connection mq for Kafka, RabbitMQ.
- **Storage**: pops connection storage for object storage (e.g., S3, Azure Blob).
- **Monitoring**: pops connection monitoring for logging and metrics (e.g., Prometheus).
- **Sessions**: Keep track of prompts, commands, and history.
- **CI/CD Pipelines**: Integrations with popular tools like Jenkins and GitHub Actions.
