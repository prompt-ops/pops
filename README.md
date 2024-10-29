# PromptOps

## Prerequisities

- `brew install upx`

## Local Setup

- `export GO111MODULE=on && go mod tidy && go build -ldflags="-s -w" -o pops && ./pops --help`

## Available Commands

- `pops connection create kubernetes my-k8s-connection`
- `pops connection create rdbms my-db`
- `pops connection list`

## Next Up

- `pops session create --connection my-db`
  - Needs a login.
  - This tells me that connection name must be unique.
- `pops session list`
  - May need a login.
- `pops session open my-rdbms-session`
  - Needs a login.
  - This should bring an existing session.
- `pops session delete`
  - Needs a login.
- `pops connection delete`

## Connection Types

- Kubernetes
- RDBMS
- NoSQL (Next)
- Cloud (Next)
  - Azure (Next)
  - AWS (Next)
  - GCP (Next)
