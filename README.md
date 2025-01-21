# 🤖 Prompt-Ops

**Prompt-Ops** is a CLI tool that makes managing your infrastructure—like Kubernetes clusters, databases, and cloud environments—effortless and intuitive. By translating natural language into precise commands, it eliminates the need to memorize complex syntax or juggle multiple tools.

With features like interactive flows, intelligent suggestions, and broad connection support, Prompt-Ops streamlines operations, saves time, and makes managing complex systems more approachable.

## Table of Contents

- [🤖 Prompt-Ops](#-prompt-ops)
  - [Table of Contents](#table-of-contents)
  - [🚀 Key Features](#-key-features)
  - [🛠️ Installation](#️-installation)
    - [Using Curl](#using-curl)
    - [Using Homebrew (WIP)](#using-homebrew-wip)
    - [Using Make](#using-make)
  - [🎮 Usage](#-usage)
  - [📜 Available Commands](#-available-commands)
    - [🌍 General](#-general)
    - [🌥️ Cloud](#️-cloud)
    - [🚆 Kubernetes](#-kubernetes)
    - [💿 Database](#-database)
  - [〄 Supported Connection Types](#-supported-connection-types)
    - [Available Now](#available-now)
    - [Coming Soon](#coming-soon)
  - [🎯 Planned Features](#-planned-features)
  - [🤝 Contributing](#-contributing)
  - [🪪 License](#-license)
  - [📚 Examples](#-examples)

## 🚀 Key Features

- 🔍 **Natural Language Commands**: Interact with your services using plain English.
- ⚡ **Interactive Workflows**: Step-by-step prompts for setup and operations.
- 🌐 **Broad Compatibility**: Supports Kubernetes, databases, cloud services, and more.
- 🔮 **AI-Powered Suggestions**: Get guided next steps and smart command completions.

## 🛠️ Installation

You can install **Prompt-Ops** using one of the following methods:

### Using Curl

Run the installation script using **curl**:

```bash
curl -fsSL https://raw.githubusercontent.com/prompt-ops/pops/main/scripts/install.sh | bash
```

### Using Homebrew (WIP)

You can also install Prompt-Ops via Homebrew:

```bash
brew tap prompt-ops/homebrew-tap
brew install pops
```

### Using Make

To install locally using `make`:

```bash
make install
```

## 🎮 Usage

You need to have `OPENAI_API_KEY` in the environment variables to be able to run certain features of Prompt-Ops. You can set it as follows:

```bash
export OPENAI_API_KEY=your_api_key_here
```

## 📜 Available Commands

### 🌍 General

- `pops conn create`: Create a new connection interactively.
- `pops conn list`: List all connections.
- `pops conn open [conn-name]`: Open a specific connection.
- `pops conn delete [conn-name]`: Delete a specific connection.
- `pops conn types`: Show available connection types.

### 🌥️ Cloud

- `pops conn cloud create`: Create a cloud connection interactively.
- `pops conn cloud list`: List all cloud connections.
- `pops conn cloud open [conn-name]`: Open a specific cloud connection.
- `pops conn cloud delete [conn-name]`: Delete a specific cloud connection.
- `pops conn cloud types`: Show supported cloud providers.

### 🚆 Kubernetes

- `pops conn kubernetes create`: Create a Kubernetes connection.
- `pops conn kubernetes list`: List Kubernetes connections.
- `pops conn kubernetes open [conn-name]`: Open a specific Kubernetes connection.
- `pops conn kubernetes delete [conn-name]`: Delete a Kubernetes connection.

### 💿 Database

- `pops conn db create`: Create a database connection.
- `pops conn db list`: List database connections.
- `pops conn db open [conn-name]`: Open a specific database connection.
- `pops conn db delete [conn-name]`: Delete a database connection.
- `pops conn db types`: Show supported database types.

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

## 🤝 Contributing

We welcome contributions! Please see our [CONTRIBUTING.md](docs/contributing/CONTRIBUTING.md) for guidelines on how to get started.

## 🪪 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📚 Examples

Please see [Prompt-Ops examples](docs/examples/README.md) for details.
