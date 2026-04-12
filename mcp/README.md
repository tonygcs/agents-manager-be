# agents-manager MCP server

MCP server that exposes the agents-manager API as tools, allowing AI assistants to list secrets, browse worker configurations, and deploy workers.

## Configuration

The server accepts one configuration parameter:

| Flag | Env variable | Default | Description |
|---|---|---|---|
| `--server` | `AGENTS_MANAGER_URL` | `http://localhost:8080` | Base URL of the agents-manager API |

## Tools

### `list_secrets`

Returns the names of all configured secrets. Values are never returned.

### `list_worker_configs`

Returns all worker configuration definitions, including image, description, command, labels, and required secrets.

### `deploy_worker`

Deploys a new worker instance from an existing worker configuration.

| Parameter | Type | Required | Description |
| --------- | ---- | -------- | ----------- |
| `worker_name` | string | yes | Name of the worker configuration to deploy (e.g. `opencode-node`) |
| `env` | string | no | JSON object with environment variables to pass to the worker (e.g. `{"KEY": "value"}`) |

Returns the container ID of the deployed worker.

## Running with Docker

Build the image (or use `make build-mcp` from the repository root):

```sh
docker build -t agents-manager-mcp ./mcp
```

Run passing the agents-manager URL:

```sh
docker run agents-manager-mcp --server http://your-agents-manager:8080
```

Or via environment variable:

```sh
docker run -e AGENTS_MANAGER_URL=http://your-agents-manager:8080 agents-manager-mcp
```

## Claude Code configuration

Add the following to your `.mcp.json`:

```json
{
  "mcpServers": {
    "agents-manager": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "AGENTS_MANAGER_URL=http://your-agents-manager:8080",
        "agents-manager-mcp"
      ]
    }
  }
}
```

If running the binary directly instead of Docker:

```json
{
  "mcpServers": {
    "agents-manager": {
      "command": "/path/to/mcp-server",
      "args": ["--server", "http://your-agents-manager:8080"]
    }
  }
}
```
