# Kubernetes Pod Status Q&A

A beginner-friendly Go project to build a read-only question-and-answer tool for Kubernetes pod status. The app will retrieve pod information from a Kubernetes cluster and use a local Ollama model to explain the results in plain language.

## Initial scope

The first capability is answering questions about pod status, for example:

- Which pods are not running?
- What is the status of a particular pod?
- Are any pods restarting?

The first version will focus on one namespace at a time. It will report Kubernetes data and explain it; it will not create, update, or delete cluster resources.

## Planned first version

Initially, this will be a Go command-line application run locally from GoLand. It will:

1. Connect to Kubernetes using the user's kubeconfig.
2. Read pod status in the selected namespace.
3. Pass only the relevant, concise status data and question to Ollama running locally.
4. Display the answer, with the underlying pod facts available for verification.

The local Kubernetes distribution (such as kind or Minikube), model, and exact command-line interface have not been selected yet.

## Learning approach

Build incrementally, explaining Go concepts as they are introduced. The user is familiar with Java and Spring, so use those as comparisons when they help clarify Go concepts. Give small, focused coding tasks instead of writing the solution immediately; let the user try first and answer specific questions or review their attempt when asked. Reviews should cover not only whether code works, but also idiomatic Go, GoLand/editor conventions, readability, error handling, and relevant good practices at the user's current learning level. For terminal steps, provide the command and explain what it does; the user will run it themselves. Start with a minimal Go program, then add Kubernetes access, and finally connect Ollama. Do not require Kubernetes or AI setup before the basic Go project is working.

## Local Ollama setup

1. Download Ollama for macOS from [ollama.com/download](https://ollama.com/download), install it, and open the Ollama app.
2. In Terminal, run `ollama run gemma4:e2b`. This downloads the model (about 7.2 GB) the first time and opens a prompt to chat with it. The model needs substantial memory; responses may be slow on less capable hardware.
3. Type a prompt, press Enter, and use `/bye` to exit.

The Ollama app provides the local API at `http://localhost:11434` while it is running. Local models do not require an API key. Model choice can be revisited depending on available Mac memory and performance.

## Safety and privacy

- Keep Kubernetes access read-only and limited to the resources needed for pod status.
- Do not let the model execute commands or change cluster resources.
- Send only the information needed to answer the question; avoid secrets and sensitive data.
- Keep Ollama local and do not expose it publicly by default.

## Run the current app

From the project directory, provide a question as command-line arguments:

```sh
go run . -namespace kube-system "Which pods are restarting?"
```

The optional `-namespace` flag defaults to `default`. Quote the question so the shell passes it as one argument and treats punctuation such as `?` literally. The CLI also joins multiple unquoted question arguments with spaces. If no question is supplied, it prints a prompt to stderr and exits with status 1. Kubernetes and Ollama integration will be added in later steps.

## Current status

The Go module is named `k8s-assistant`. The CLI accepts an optional namespace flag, validates the namespace and question, and delegates application flow to `internal/app`. `main.go` reports errors and exits with a non-zero status.

## Local Kubernetes development

The local development cluster is kind, using Podman as its container provider. The `k8s-assistant` cluster and kubectl context are configured locally.

The Kubernetes Go client dependency (`k8s.io/client-go`) is used through `internal/k8s.Client`. The app creates one client when Ollama requests a tool and reuses it for each call in that tool round. The package can list pod status and deployment replica status in the selected namespace.

The local Ollama chat API was verified with `curl` at `http://localhost:11434/api/chat`. The CLI exposes the read-only `list_pods` and `list_deployments` tools. It dispatches only allowlisted tool names, sends JSON-encoded Kubernetes data back to Ollama, and preserves the tool-call fields and `tool_name` in the follow-up conversation. No cloud API key is needed for this local model.

The project is organized into Go packages: `main.go` is the CLI entry point; `internal/app` coordinates the application flow; `internal/k8s` contains the reusable Kubernetes client and queries; `internal/ollama` contains Ollama API code and protocol types; and `internal/tools` defines and dispatches model-visible tools. Tool calls are currently handled sequentially and in a single round.

Next improvements: restore focused tests; populate deployment conditions and represent full selectors (including match expressions); make Ollama HTTP requests cancellable and time-limited; support additional tool-call rounds; then execute multiple requested tools concurrently while preserving their original order in the conversation.

See [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) for the decisions, roadmap, and a handoff summary for future work sessions.
