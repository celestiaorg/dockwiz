# DockWiz

DockWiz is a tool that offers an API server for seamless Docker image construction. It empowers developers to effortlessly build Docker images through a user-friendly API.

## Build Instructions

### Prerequisites
- Go installed on your machine
- Docker installed on your machine

### Building the Tool

To build the DockWiz tool, follow these steps:

1. Clone the repository:

    ```bash
    git clone https://github.com/celestiaorg/dockwiz.git
    ```

2. Change into the project directory:

    ```bash
    cd dockwiz
    ```

3. Run the build command:

    ```bash
    make build
    ```

This will generate the binary in the `bin/` directory.

## Docker Support

DockWiz also includes Docker support for easy containerization.

### Building the Docker Image

To build the Docker image, run:

```bash
make docker
```

### Development Environment

For a development environment, use the following command:

```bash
make dev
```

This command will build the development Docker image, create and run a Redis container, and launch a shell within a container with the necessary volume mounts.

## Usage

Once the tool is built, you can use it with the following command:

```bash
./bin/dockwiz serve [flags]
```

Command Flags

*    `--log-level`: Set the log level (e.g., debug, info, warn, error, dpanic, panic, fatal). Default is "info".
*    `--origin-allowed`: Set the allowed origin for CORS. Default is "*".
*    `--production-mode`: Enable production mode to disable debug logs.
*    `--redis-addr`: Set the Redis server address. Default is "localhost:6379".
*    `--redis-db`: Set the Redis database.
*    `--redis-password`: Set the Redis password.
*    `--serve-addr`: Set the address to serve on. Default is ":9007".

For example:

```bash
./bin/dockwiz serve --redis-addr 172.17.0.2:6379 --serve-addr :8080
```

**Warning:** Never run this binary outside a container as `root` because it might mess with your file system and damage your OS.

### API Usage Examples

```bash
curl -X POST -H "Content-Type: application/json" --data '{"git_options" : {"url": "https://github.com/celestiaorg/bittwister/"}}' http://localhost:8080/api/v1/build
```

```bash
curl -X POST -H "Content-Type: application/json" --data '{"git_options" : {"url": "https://github.com/celestiaorg/celestia-app"}}' http://localhost:8080/api/v1/build
```

which will give this output:
```json
{
  "image_name": "c830a947-44c0-40ff-bda2-29ff95423463",
  "image_tag": "1h"
}
```

`image_name` can be used as a reference to query the build status.

Check Build Status:

```bash
curl http://localhost:8080/api/v1/status/c830a947-44c0-40ff-bda2-29ff95423463
```

sample output:
```json
{
  "status": 2,
  "status_string": "building",
  "error": "",
  "start_time": "2024-01-11T16:31:54.892884307Z",
  "end_time": "0001-01-01T00:00:00Z",
  "logs": "Building image c830a947-44c0-40ff-bda2-29ff95423463:1h\n\u001b[36mINFO\u001b[0m[0044] Resolved base name docker.io/golang:1.21-alpine3.18 to develop... \n"
}
```

By default the status is kept in the system for 24 hours, so users can query their build status.
