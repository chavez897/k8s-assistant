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

The Go module is named `k8s-assistant`. The CLI accepts an optional namespace flag and validates/trims the namespace and question through a helper that returns errors. Process-level reporting and exit handling stay in `main`. The table-driven test in `main_test.go` passes with `go test`.

## Local Kubernetes development

The local development cluster is kind, using Podman as its container provider. Podman 6.1.4, kind v0.33.0, and kubectl are available, and the `k8s-assistant` cluster has been created and is running.

The user's kubectl context is configured for the local cluster. The next implementation step is adding the Kubernetes Go client and using it to list pods in the selected namespace.

The Kubernetes Go client dependency (`k8s.io/client-go`) has been added. The CLI loads the current kubeconfig context and returns pod names, phases, container names, restart counts, and current waiting/termination reasons from the selected namespace; `main` formats the results. Pods with no container statuses are included. The user verified the CLI works against the local cluster. Application logic runs in `run`, and Kubernetes operation errors are wrapped with context.

The local Ollama chat API was verified with `curl` at `http://localhost:11434/api/chat`. The CLI defines a read-only `list_pods` tool and handles a tool-call response by fetching pod data and sending the result back to Ollama. The assistant tool-call type and function index are preserved in the follow-up request, and the tool result includes its `tool_name`. No cloud API key is needed for this local model.

Planned code organization: keep the project as one Go package for now and split responsibilities into `main.go`, `ollama.go`, `kubernetes.go`, `tools.go`, and focused test files. Add separate packages only after the application has grown enough to justify those boundaries.

Next implementation direction: add another read-only Kubernetes tool, then dispatch multiple requested tools concurrently with goroutines. Each goroutine will return a result to a pre-sized, index-matched slice; the main flow will wait for all results and append them to the Ollama conversation in tool-call order.

See [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md) for the decisions, roadmap, and a handoff summary for future work sessions.
