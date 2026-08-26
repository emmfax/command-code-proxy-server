# CommandCode Proxy Server

English | [简体中文](README.zh-CN.md)

OpenAI-compatible proxy server for the CommandCode API. It exposes `/v1/chat/completions` and `/v1/models` endpoints so OpenAI-compatible clients can call CommandCode models through a local HTTP server.

Repository: https://github.com/dev2k6/command-code-proxy-server

Version: `v1.0.9`

## Features

- OpenAI-compatible chat completions endpoint
- Streaming and non-streaming responses
- OpenAI-compatible model list endpoint
- Short model name mapping
- Optional default API key from CLI
- Per-request API key via `Authorization` header
- **Client key whitelist** (`-auth-keys` / env `CCP_AUTH_KEYS`) — restrict who may use this proxy; upstream CommandCode key stays server-side
- Empty/null message content is normalized (never serialized as `"content": null`)
- Oversized request bodies rejected with a clear 413 error
- Upstream empty responses surfaced as a clear 502 error with guidance instead of a mysterious empty reply
- Configurable host and port
- Checks GitHub tags for a newer proxy version and displays it next to the current version

## Requirements

- Go 1.26.2 or newer

## Run

```bash
go run main.go
```

Default server address:

```text
http://127.0.0.1:55990
```

## CLI options

```bash
go run main.go [options]
```

| Option | Default | Description |
| --- | --- | --- |
| `-host` | `127.0.0.1` | Host to bind the server to |
| `-port` | `55990` | Port to run the server on |
| `-api-key` | empty | CommandCode API key used for upstream calls (kept server-side) |
| `-auth-keys` | empty | Comma-separated client keys allowed to use this proxy; empty = open proxy. Env: `CCP_AUTH_KEYS` |
| `-version` | `false` | Print version and exit |

Examples:

```bash
# Run on default host and port
go run main.go

# Run on a custom port
go run main.go -port 8080

# Expose on all interfaces
go run main.go -host 0.0.0.0

# Use a default API key for all upstream calls that do not include Authorization
go run main.go -api-key your-commandcode-api-key

# Require clients to present one of these keys (upstream key stays secret on the server)
go run main.go -auth-keys "client-key-1,client-key-2" -api-key your-commandcode-api-key

# Same via environment variable
CCP_AUTH_KEYS="client-key-1,client-key-2" go run main.go

# Print version
go run main.go -version
```

## Build

Build for the current platform:

```bash
go build -o bin/command-code-proxy
```

Cross-compile for Windows and Linux:

```bash
GOOS=windows GOARCH=amd64 go build -o bin/command-code-proxy.exe
GOOS=linux GOARCH=amd64 go build -o bin/command-code-proxy
```

## API key behavior

Two separate keys are involved:

1. **Client key** — the `Authorization` header sent by whoever calls this proxy.
   When `-auth-keys` (or env `CCP_AUTH_KEYS`) is configured, this key must be in
   the whitelist, otherwise the request is rejected with `401`.
   When no whitelist is configured, any client key is accepted (open proxy).
2. **Upstream key** — the key used against `api.commandcode.ai`. Resolution order:
   1. `-api-key` CLI value (recommended: keep it server-side and secret)
   2. The client's `Authorization` key (pass-through)

Header format:

```http
Authorization: Bearer your-key
```

If neither a whitelist nor any key is available, the request returns `401 Unauthorized`.

## Content normalization

OpenAI clients sometimes send messages with `"content": null` or `"content": ""`
(e.g. assistant turns that only contain tool calls). The CommandCode upstream
rejects null content, so this proxy guarantees every outgoing message carries
valid content parts — empty ones become a single text part.

## Request size limit

Request bodies are capped at 32 MB. Larger uploads are rejected with a clear
`413` error instead of hanging or being silently dropped by intermediate layers.

## Empty upstream responses

When the CommandCode API answers `200` with an empty body (observed when the
conversation context exceeds the model limit), the proxy converts it into a
`502` with an explanatory message so clients can react (e.g. compact history)
instead of hanging on a dead stream.

## Endpoints

### Health check

```http
GET /health
```

Response:

```json
{"status":"ok"}
```

### List models

```http
GET /v1/models
```

Returns an OpenAI-compatible model list.

### Chat completions

```http
POST /v1/chat/completions
```

Example non-streaming request:

```bash
curl http://127.0.0.1:55990/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-commandcode-api-key" \
  -d '{
    "model": "deepseek-v4-pro",
    "messages": [
      {"role": "system", "content": "You are helpful."},
      {"role": "user", "content": "Hello"}
    ],
    "stream": false
  }'
```

Example streaming request:

```bash
curl -N http://127.0.0.1:55990/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-commandcode-api-key" \
  -d '{
    "model": "deepseek-v4-pro",
    "messages": [
      {"role": "user", "content": "Write a short poem."}
    ],
    "stream": true
  }'
```

## Supported model aliases

The proxy accepts full model IDs and these short aliases:

| Alias | Maps to |
| --- | --- |
| `deepseek-v4-pro`, `deepseek-v4`, `deepseek-pro` | `deepseek/deepseek-v4-pro` |
| `deepseek-v4-flash`, `deepseek-flash` | `deepseek/deepseek-v4-flash` |
| `minimax-m2.7`, `minimax2.7` | `MiniMaxAI/MiniMax-M2.7` |
| `minimax-m2.5`, `minimax2.5`, `minimax` | `MiniMaxAI/MiniMax-M2.5` |
| `glm-5.1` | `zai-org/GLM-5.1` |
| `glm-5` | `zai-org/GLM-5` |
| `kimi-k2.6`, `kimi2.6` | `moonshotai/Kimi-K2.6` |
| `kimi-k2.5`, `kimi2.5` | `moonshotai/Kimi-K2.5` |
| `qwen-3.6-max-preview`, `qwen3.6-max` | `Qwen/Qwen3.6-Max-Preview` |
| `qwen-3.6-plus`, `qwen3.6-plus`, `qwen3.6` | `Qwen/Qwen3.6-Plus` |
| `step-3.5-flash`, `step3.5` | `stepfun/Step-3.5-Flash` |
| `gemini-3.1-flash-lite`, `gemini-flash-lite` | `google/gemini-3.1-flash-lite` |
| `minimax-m3`, `minimax3` | `MiniMaxAI/MiniMax-M3` |
| `qwen-3.7-max-free`, `qwen3.7-max-free` | `Qwen/Qwen3.7-Max-Free` |
| `qwen-3.7-max`, `qwen3.7-max` | `Qwen/Qwen3.7-Max` |
| `step-3.7-flash`, `step3.7` | `stepfun/Step-3.7-Flash` |
| `mimo-v2.5-pro`, `mimo-pro` | `xiaomi/mimo-v2.5-pro` |
| `mimo-v2.5`, `mimo` | `xiaomi/mimo-v2.5` |

Unknown model names are passed through unchanged.

## Project structure

```text
.
├── README.md
├── go.mod
├── go.sum
├── main.go
├── bin
│   ├── command-code-proxy
│   └── command-code-proxy.exe
└── internal
    ├── api
    │   ├── commandcode.go
    │   └── openai.go
    ├── proxy
    │   ├── convert.go
    │   ├── model.go
    │   └── proxy.go
    ├── server
    │   └── server.go
    ├── update
    │   └── update.go
    └── version
        └── version.go
```

## How it works

1. Client sends an OpenAI-compatible request to the local proxy.
2. The proxy extracts system messages, maps the model name, and converts messages to CommandCode format.
3. The proxy sends the request to `https://api.commandcode.ai/alpha/generate`.
4. CommandCode streaming NDJSON events are converted back to OpenAI-compatible SSE chunks or collected into a single JSON response.

## Version check

On startup and when running `-version`, the proxy calls:

```text
https://api.github.com/repos/dev2k6/command-code-proxy-server/tags
```

If the latest GitHub tag is newer than the current app version, the version line is displayed as:

```text
v1.0.8 (latest: v1.x.x)
```

## CommandCode version header

The upstream request includes:

```http
x-command-code-version: <latest npm command-code version>
```

The value is fetched from:

```text
https://registry.npmjs.org/command-code/latest
```

The fetched version is cached for 30 minutes. If the registry request fails, the proxy uses the last cached version, or `unknown` if no version has been fetched yet.
