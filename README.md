# counter-service

* [1 Prerequisites](#1-prerequisites)
* [2 Prepare the environment](#2-prepare-the-environment)
* [3 Run the service](#3-run-the-service)
  * [3.1 Run locally](#31-run-locally)
  * [3.2 Run using Docker](#32-run-using-docker)

## 1 Prerequisites

* [Go](https://go.dev/) (`v1.22` or later)
* [Docker](https://www.docker.com/)
* [goose](https://github.com/pressly/goose) (for running migrations)

`goose` can be installed by the following command (Go language must be already installed on the machine):

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

`goose` can also be installed using `brew` (for MacOS):

```bash
brew install goose
```

## 2 Prepare the environment

#### Step 1 - Up containers

Run the following command to up all infrastructure containers.

```bash
docker compose -f ./docker-compose-infra.yml up -d
```

#### Step 2 - Run migrations

Execute the commands below to apply migrations:

```bash
goose -dir ./migrations/ postgres "host=localhost port=35432 user=postgres password=123 dbname=postgres" up
```

## 3 Run the service

Depending on your preferences, you can run the service using one of the following ways:

* Run locally (Go must be installed)
* Run using Docker

### 3.1 Run locally

Run the following command:

```bash
go run ./cmd/app/main.go --config ./config/local.yml
```

### 3.2 Run using Docker

Run the following command:

```bash
docker compose -f ./docker-compose-service.yml up -d
```
