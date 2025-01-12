# Add a new Connection Type

## Introduction

Prompt-Ops has connection types and connection subtypes. Some examples are as follows:

| Connection Type | Connection Subtype                                        |
| --------------- | --------------------------------------------------------- |
| Database        | PostgreSQL                                                |
| Database        | MySQL                                                     |
| Database        | MongoDB                                                   |
| Cloud Provider  | Azure                                                     |
| Cloud Provider  | AWS                                                       |
| Kubernetes      | (No subtype - any cluster selected during initialization) |

## How to add a new connection type

1. Create a new file under `pkg/connection` for the new connection type. For example, `mqs.go` for message queues.
2. Define the new connection type in `pkg/connection/mqs.go`.
3. Implement the `ConnectionInterface` for the new connection type.
4. For an example, please see `DatabaseConnectionType` in `pkg/connection/db.go`.
5. Now you can create subtypes for the new connection.
6. Suggest improvements if you think the code structure can be enhanced. We welcome new ideas.
7. Consider creating new files for new connection types to keep the main connection file manageable.
