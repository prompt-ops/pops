# Add a new Connection Subtype

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

## How to add a new connection subtype

1. Implement the `ConnectionInterface` in `pkg/connection/types.go` under the proper connection type.
2. For an example, please see `PostgreSQLConnection` in `pkg/connection/db.go`.
3. Suggest improvements if you think the code structure can be enhanced. We welcome new ideas.
4. Consider creating new files for new subtypes to keep the main connection file manageable.
