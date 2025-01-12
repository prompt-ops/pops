# Add a new AI Model

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

Prompt-Ops also has access to AI models that help power this application.

| Creator | Model  |
| ------- | ------ |
| OpenAI  | gpt-4o |

## How to add a new AI model

1. Create a new file under `pkg/ai` for the new AI model. For example, `meta.go` for Meta being the creator of `Llama 3.1`.
2. Implement the `AIModel` interface in `pkg/ai/types.go` for the new AI model. Define the functions needed by the interface.
3. For an example, please see `OpenAIModel` in `pkg/ai/openai.go`.
4. Suggest improvements if you think the code structure can be enhanced. We welcome new ideas.
5. Naming may not sound right but as time passes we are going to improve.
