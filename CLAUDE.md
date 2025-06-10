# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

Build the project:
```bash
make build
```

Run unit tests:
```bash
make unit-test
```

Run linter:
```bash
make lint
```

Format code and organize imports:
```bash
make organize-imports
```

Install locally:
```bash
make install
```

## Architecture Overview

Prompt-Ops follows a clean, layered architecture with clear separation of concerns. The architecture implements Domain-Driven Design with hexagonal architecture patterns.

### Architectural Layers

**Domain Layer** (`pkg/domain/`):
- Contains core business entities: `Connection`, `Command`, `Session`
- Defines interfaces for repositories and adapters
- No external dependencies - pure business logic
- Includes comprehensive unit tests

**Application Layer** (`pkg/services/`):
- Business logic services: `ConnectionService`, `CommandService`, `ShellService`
- Orchestrates domain entities and coordinates with infrastructure
- Dependency injection through interfaces
- Comprehensive service tests with mocks

**Infrastructure Layer** (`pkg/adapters/`):
- Concrete implementations of domain interfaces
- Database adapters, AI providers, external service integrations
- Includes fake implementations for testing (e.g., `FakeKubernetesAdapter`)
- Clean adapter pattern isolates external dependencies

**Presentation Layer** (`pkg/ui/`, `cmd/`):
- Modern TUI shell with Claude-Code-like experience
- Cobra CLI commands for traditional CLI usage
- Clean separation between UI state and business logic

### Clean Architecture Principles

**Dependency Rule**: Dependencies point inward - domain layer has no external dependencies, application layer depends only on domain interfaces, infrastructure implements domain interfaces.

**Interface Segregation**: Small, focused interfaces (`ConnectionRepository`, `ConnectionAdapter`, `CommandGenerator`)

**Dependency Injection**: Services receive dependencies through constructor injection, not global state

**Single Responsibility**: Each service and domain entity has one reason to change

### Key Components

**Domain Models**:
- `Connection`: Simplified model with clean configuration structure
- `Command`: Includes safety levels and impact assessment
- `Session`: Tracks user interactions and conversation history

**Service Layer**:
- `ConnectionService`: Manages connection lifecycle
- `CommandService`: Handles AI command generation and execution
- `ShellService`: Provides beautiful context display with emojis and icons

**Testing Strategy**:
- Domain: Unit tests for business logic
- Services: Mock-based testing with comprehensive scenarios
- Adapters: Integration tests with fake implementations
- Fake Kubernetes clusters for realistic testing

## Environment Requirements

Set `OPENAI_API_KEY` environment variable to enable AI features:
```bash
export OPENAI_API_KEY=your_api_key_here
```

## Testing

Run tests with:
```bash
go test ./... -v
```

The project uses standard Go testing with testify for assertions.