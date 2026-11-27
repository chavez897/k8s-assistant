# Kubernetes Pod Status Q&A

A read-only Go CLI that answers natural-language questions about Kubernetes pods and deployments. It uses the current kubeconfig context to read data from one namespace and a local Ollama model to explain the results.

## What it does

- Accepts a question and an optional namespace (`default` by default).
- Sends the question and available read-only tool definitions to Ollama.
- Lets the model request pod or deployment status through an allowlisted tool registry.
- Returns concise Kubernetes summaries to Ollama and prints its final answer.
- Executes up to four tool calls concurrently per response, while preserving their original order in the conversation.
- Cancels in-flight Ollama HTTP and Kubernetes API requests when interrupted with Ctrl+C.

The model cannot execute shell commands or modify cluster resources. Keep the local Ollama service private and grant the kubeconfig identity only the read permissions the app needs.

## Requirements

- Go 1.27.3 or compatible with the module version in `go.mod`.
- A reachable Kubernetes cluster and kubeconfig with read access to pods and deployments in the selected namespace.
- Ollama running locally at `http://localhost:11434`, with the configured model `gemma4:e2b` available.

The project has been developed with a kind cluster using Podman. Other Kubernetes distributions should work through the current kubeconfig context.

## Run

Start Ollama and ensure the desired Kubernetes context is active, then run from the project directory:

```sh
go run . -namespace kube-system "Which pods are restarting?"
```

The `-namespace` flag is optional and defaults to `default`. Questions can also be passed as multiple arguments; the CLI joins them with spaces. Quoting the question is recommended, especially when it contains shell punctuation.

Run the package build/tests with:

```sh
go test ./...
```

There are currently no test files, so this command verifies compilation but does not yet provide behavioral test coverage.

## Project layout

- `main.go` — CLI entry point and top-level error reporting.
- `internal/app` — input handling, Ollama conversation/tool-call loop, and bounded concurrent tool execution.
- `internal/k8s` — kubeconfig-based client and concise pod/deployment status queries.
- `internal/ollama` — Ollama HTTP API client and request/response types.
- `internal/tools` — tool definitions and allowlisted dispatch.
- `internal/tools/commands` — command interface and individual pod/deployment commands.

Tool definitions and dispatch are built from an ordered command registry. To add a tool, implement the command interface and register the command in `internal/tools`.

## Current limitations

- Ollama requests are cancellable but do not yet have an explicit timeout.
- The tool-call conversation loop has no maximum-round limit.
- Deployment summaries currently include match-label selectors only; selector expressions and deployment conditions are not populated.
- The test suite needs focused unit tests for input validation, Kubernetes summaries, command dispatch, serialization, and cancellation/concurrency behavior.
- End-to-end verification is still needed for multi-tool rounds and cancellation while Ollama or Kubernetes requests are in flight.

Before expanding the tool set, prioritize those reliability items: focused behavioral tests, an explicit Ollama timeout, a maximum tool-call-round limit, and end-to-end checks for concurrency and cancellation. Once these are covered, new tools can be added through the command registry.

See [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) for development decisions, learning preferences, and the current roadmap.
