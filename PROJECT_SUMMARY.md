# Project Summary

Use this document to continue the project and Go-learning sessions.

## Goal and decisions

- Learn Go by building a CLI that answers natural-language questions about Kubernetes pod and deployment status, with a limited Deployment-creation capability.
- Use the user's kubeconfig/current context and focus on one namespace per invocation.
- Use Ollama locally (`gemma4:e2b` at `http://localhost:11434`) for the initial AI runtime.
- The model may request only explicitly registered tools. It cannot run shell commands. Read tools list pods and deployments; the `create_deployment` write tool can create a Deployment in the CLI-selected namespace without a confirmation prompt.
- Deployment creation accepts a name, container image, and 1–5 replicas. The image is run in the cluster, so only use the write tool where model-requested changes are acceptable.
- Send concise resource summaries rather than complete Kubernetes API objects.
- The user is new to Go and experienced with Java/Spring. Teach incrementally with small tasks, using Java comparisons when useful. Let the user implement first; provide code only when requested. Review correctness, idiomatic Go, readability, error handling, and relevant practices at the user's current level.
- For terminal steps, explain and provide commands for the user to run; do not run commands on the user's behalf.
- The user knows Kubernetes; keep cluster setup guidance concise and assume familiarity unless asked.

## Current implementation

- Go module: `k8s-assistant`; module version in `go.mod` is Go 1.27.3.
- `main.go` calls `internal/app.Run()` and reports errors to stderr with a non-zero exit status.
- The CLI accepts `-namespace` (default `default`) and a question. It trims and validates both inputs.
- `internal/app` builds the initial Ollama request and handles repeated tool-call rounds. The Kubernetes client is initialized lazily when the model first requests a tool and reused across rounds.
- Each round can execute up to four tool calls concurrently using a worker pool. Results are stored by original call index and appended to the Ollama conversation in request order. Tool failures are returned as errors.
- The agent loop stops with an error if it exceeds 10 tool-call rounds.
- Ctrl+C cancels the shared context. That context is passed to Ollama HTTP requests and Kubernetes API calls.
- Ollama is configured with a local base URL and model in `internal/ollama/client.go`. Requests use `http.NewRequestWithContext` and a reusable HTTP client with a two-minute timeout per request.
- `internal/tools` contains an ordered registry used to build tool definitions and dispatch by command name. `internal/tools/commands` defines the command interface and the `list_pods`, `list_deployments`, and `create_deployment` implementations.
- Tool results are JSON-encoded concise status summaries. Pod summaries include pod name/phase and per-container name, restart count, and waiting/terminated reason when present; pods without container statuses are still represented.
- Deployment summaries include replica counts, images, labels, match-label selectors, and strategy. Conditions are not populated, and selector match expressions are not represented.
- `go test ./...` and `go test -race ./...` have passed, but the current repository has no test files, so they only verify compilation and do not provide behavioral coverage.

## Environment and run command

The user reports a macOS `darwin/arm64` development machine, a Podman-backed kind cluster named `k8s-assistant`, and a configured kubectl context. Ollama with `gemma4:e2b` and its local API were previously verified by the user.

Example:

```sh
go run . -namespace kube-system "Which pods are restarting?"
```

The namespace flag is optional. The CLI joins multiple question arguments with spaces; quoting the question is recommended.

## Package layout

- `main.go`: process entry point.
- `internal/app`: CLI orchestration, Ollama agent loop, and bounded worker pool.
- `internal/k8s`: kubeconfig client and Kubernetes resource summaries.
- `internal/ollama`: Ollama HTTP client and protocol types.
- `internal/tools`: command registry, tool definitions, and dispatch.
- `internal/tools/commands`: individual read/write command implementations and shared tool-message construction.

## Suggested roadmap

1. Add focused behavioral tests, starting with command definitions/dispatch and tool-message serialization; then cover input validation and Kubernetes summary mapping.
2. Complete deployment summaries by populating conditions and representing selectors including match expressions.
3. Exercise end-to-end flows with Ollama and a local Kubernetes cluster, including cancellation, unknown tool names, one and multiple tools per round, timeout behavior, tool errors, and the Deployment-creation path.
4. After reliability and safety behavior are covered, expand the available tools through the command registry.

## Local safety notes

- Grant the kubeconfig identity only the permissions needed for the selected namespace, including `create` on deployments if enabling the write tool.
- The model is not allowed to perform arbitrary commands. It can create Deployments through the explicit allowlisted tool; there is currently no confirmation step.
- Keep Ollama bound to localhost unless there is a deliberate, secured reason to expose it.
- Avoid passing secrets or unnecessary metadata to the model. Review the requested image and resource impact when using the Deployment-creation capability.
