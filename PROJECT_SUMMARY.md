# Project Summary

Use this document to pick up the project in another work session. Update it as decisions are made and implementation progresses.

## Goal

Learn Go by building a small DevOps tool that answers natural-language questions about Kubernetes pod status.

## Agreed decisions

- **First capability:** Answer questions related to pod status.
- **Example questions:** Identify pods that are not running, report a named pod's status, and find pods that are restarting.
- **AI runtime:** Ollama, running locally for the initial version.
- **Initial app shape:** A Go CLI run locally from GoLand, not a Kubernetes workload yet.
- **Cluster access:** Use the user's kubeconfig and read Kubernetes data; initially focus on one namespace.
- **Answer flow:** The Go app fetches pod information, sends a concise evidence summary plus the question to Ollama, and displays a plain-language response.
- **Safety:** Read-only Kubernetes access. The model must not execute commands or make cluster changes. Avoid sending secrets or unnecessary sensitive data.
- **Teaching style:** The user is new to Go but very familiar with Java and Spring. Walk through the work in small, focused steps, use Java/Spring comparisons when useful, and explain concepts as they arise. Give a coding task first and let the user try; don't provide the code solution unless they ask or need help. When reviewing work, cover correctness as well as idiomatic Go, GoLand/editor conventions, readability, error handling, and relevant good practices at an appropriate level. For terminal commands, explain and present the command for the user to run; do not run the command on their behalf.

## Not decided yet

- Whether kind is already installed.
- Exact kind/Podman setup required in this environment.
- Whether `gemma4:e2b` is suitable for the user's Mac, based on available memory and performance.
- CLI flags and exact user interaction.
- Whether to include Kubernetes events in the first version or only pod status.
- Whether to later add an HTTP API or deploy the app into Kubernetes.

## Roadmap

1. Check/install Go and create a minimal Go module. (Done)
2. Run a small Go program and learn the project structure. (Done; current CLI accepts and prints a question argument.)
3. Choose and start a local Kubernetes cluster.
4. Read pod status for a namespace using the Kubernetes Go client and kubeconfig.
5. Add a useful text/structured summary of the pod data.
6. Connect the Go app to the local Ollama API.
7. Test healthy pods, unhealthy/restarting pods, missing namespaces, and Ollama errors.
8. Document how to run the working CLI; consider API/deployment only after the MVP.

## Current state

- Go is installed: Go 1.27.3 on macOS `darwin/arm64` (as reported by the user).
- Initialized Go module `podqa`, then renamed it to `k8s-assistant` (`go.mod` declares Go 1.27.3).
- `main.go` parses the `-namespace` flag (default `default`), validates and trims question/namespace via `sanitizeInput(input, inputName) (string, error)`, and reports errors to stderr.
- Example invocation: `go run . -namespace kube-system "Which pods are restarting?"`. Quote questions with punctuation so the shell passes them literally.
- The user asked about `fmt`, imports/packages, Go constants, pointers, `flag.Args`, kubeconfig context versus `context.Context`, `context.TODO`, `metav1.ListOptions{}`, and `stop()` from `signal.NotifyContext`; explanations used Java/Spring comparisons.
- The user created `main_test.go` with a table-driven `TestSanitizeInput`, covering trimming, empty input, and whitespace-only input. It was reviewed as idiomatic; the user ran `go test` and confirmed it passed. User prefers moving on rather than adding more tests now.
- The user created `main_test.go` with a table-driven `TestSanitizeInput`, covering trimming, empty input, and whitespace-only input. The structure, named subtests, and error/value assertions are appropriate and idiomatic.
- The user ran `go test` and confirmed it passed.
- User preference: move on from tests rather than adding error-message assertions now. Existing basic helper tests pass; do not make test expansion the next task.
- User reports Podman version 6.1.4 on macOS and says the Podman machine looks healthy/running.
- User reports kind v0.33.0 is installed.
- User chose kind instead of Minikube for the local Kubernetes cluster. Use kind with Podman as the container provider; kind's Podman support may require its `KIND_EXPERIMENTAL_PROVIDER=podman` setting.
- `KIND_EXPERIMENTAL_PROVIDER=podman kind get clusters` returned no clusters.
- User confirms `kubectl` is installed.
- User created the named `k8s-assistant` kind cluster with Podman and reports it is up and running.
- User confirms kubectl context is pointing at the cluster.
- User knows Kubernetes and requests faster, more general K8s guidance: assume familiarity with cluster setup and avoid detailed installation/cluster instructions. Prefer concise checks/tasks such as “Make sure your Kubernetes cluster is up and running” and move directly to the Go/client implementation.
- Current learning step: begin Kubernetes API integration in small coding increments. User wants to write code themselves; provide a task and explain commands, not implementation code unless asked.
- User ran `go get k8s.io/client-go@latest`; dependency installation is reported complete. This adds the Kubernetes client module to `go.mod` and module checksums to `go.sum`.
- The user implemented `getK8sPods(ctx, namespace)` using client-go's default kubeconfig loading rules, creates a clientset, lists pods with `metav1.ListOptions{}`, and prints names/phases. The helper returns errors for failures and uses a Go context passed from `main`.
- Current-context config loader: `clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{}).ClientConfig()`. It honors `KUBECONFIG` and current-context like kubectl.
- Terminology note: Kubernetes kubeconfig `current-context` (cluster/user/namespace selection) is distinct from Go's `context.Context` passed to API calls for cancellation/deadlines/request-scoped values. Explain both when discussing “context.”
- Go request-context recommendation for the CLI: create a root context in `main` with `signal.NotifyContext(context.Background(), os.Interrupt)`, defer its cancel function, pass the context to `getK8sPods(ctx, namespace)`, and use it in the client-go `List` call. `context.Background()` and `context.TODO()` both have no cancellation/deadline; `Background` is the intentional root, while `TODO` signals that the right context is not yet decided.
- `signal.NotifyContext` returns a `stop` function that should be deferred. It cancels the context and unregisters the signal notification/releases associated resources when the command finishes; pressing Ctrl+C cancels the context so in-flight API calls can stop.
- Current code review: `main` creates a context with `signal.NotifyContext(context.Background(), os.Interrupt)` and defers `stop()`, then calls helpers. The Kubernetes helper prints the kubeconfig load error and returns it, so `main` prints it a second time. Also, `os.Exit(1)` in `main` skips deferred functions in `main` on error paths. Next refactor details: move application work (flag setup/parsing, input validation, context creation/defer, Kubernetes call) into `run() error`; make `main` call `run()`, print the returned error once, then call `os.Exit(1)` if needed. Since `run` has already returned, its deferred `stop()` has executed before `os.Exit`. In `getK8sPods`, remove stderr printing and just return contextual errors (e.g. wrap config/client/list failures with `fmt.Errorf("...: %w", err)`). User requested expanded explanation of this task before implementation.
- User knows Kubernetes and requests concise/general K8s setup guidance; assume the cluster is up unless asked otherwise.
- Ollama setup instructions are in `README.md`: download and open the macOS app, then try `ollama run gemma4:e2b` (about a 7.2 GB model download). This model has not yet been tested on the user's hardware.
- Official Ollama docs: [download](https://ollama.com/download) and [quickstart](https://docs.ollama.com/quickstart).

## Latest review and next step

- User refactored `getK8sPods` to return `[]PodInfo` rather than printing internally, and `run` returns that result to `main`, which formats it. This is a useful separation of data retrieval from presentation.
- Review observations: `run` still prints namespace/question itself; that is acceptable for now. `main`'s error branch calls `os.Exit(1)`, but `run` has already returned so its deferred signal cleanup has executed. `PodInfo` uses unexported fields, appropriate because it is internal to package `main`. The printing loop currently lives in `main`; this is fine, but a `String()` method on `PodInfo` could centralize formatting if the output grows.
- User now nil-checks both `containerStatus.State.Terminated` and `State.Waiting`, using `else if`, and assigns the current-state reason to `PodInfo.reason`; `main` includes the reason in output. This is correct and avoids nil dereference. The output is useful; blank container/reason fields for pods without status are a minor presentation issue only.
- User ran the CLI and confirmed it is working.
- User confirms `gemma4:e2b` appears in `ollama list`; local model is available.
- User verified Ollama's local chat API with `curl` against `http://localhost:11434/api/chat`; local inference works without a cloud API key.
- User implemented `ollamaRequest` using `net/http` and `encoding/json/v2`; request fields are exported with JSON tags. The Ollama base URL's leading space has been removed.
- Current behavior sends the question to Ollama and prints the raw response body. Next task: check for a non-2xx HTTP status and decode the response JSON to print `message.content` rather than the whole JSON. User should implement; explain response structs/tags and status codes as needed. Verify the corrected URL works before moving on.
- User ran `go run . -namespace default "Which pods are running?"` and confirmed it works against the local cluster.
- Current code builds `[]PodInfo` from each pod's `ContainerStatuses` and prints pod name, phase, container name, restart count. The nested loops and selected fields are reasonable and readable for this stage.
- Review issue: pods with zero `ContainerStatuses` are omitted entirely because `PodInfo` is appended only inside the inner loop. This can hide Pending/initializing pods. Ensure each pod is represented even if the status slice is empty, while still representing multiple containers.
- Waiting/termination reasons are not yet shown. Add an optional reason to `PodInfo`, extracting it only when `containerStatus.State.Waiting` or `.Terminated` is non-nil; retain phase and restart count. These state fields are pointers, so check nil before accessing.
- Next small task: update the output model/loop to preserve pods with no container statuses and include non-empty waiting/terminated reasons. Then run the CLI and inspect output with healthy and non-running pods.

## Latest Ollama review

- `ollamaRequest` now checks `resp.StatusCode != http.StatusOK`, reads/unmarshals the body into `OllamaResponse`, returns `(OllamaResponse, error)`, and `run` prints `response.Message.Content`. This completes decoding and is a good value-plus-error Go pattern.
- Review issue: `OllamaResponse.CreatedAt` uses JSON tag `createdAt`, but Ollama's API field is `created_at`; change the tag if you intend to use that timestamp. Other response fields use snake_case as expected.
- `ollamaRequest` currently sends only the question and is called before `getK8sPods`, so the AI does not yet receive cluster evidence. Next task: retrieve pods first, build a concise evidence string from `[]PodInfo`, and include it with the question in the Ollama message. User should implement in small steps and retain read-only behavior.
- GoLand reports the ignored error from `defer resp.Body.Close()`. Closing an HTTP response body is required, but its close error is commonly ignored after reading; an explicit `_ = resp.Body.Close()` inside a deferred closure can make that intentional to the IDE. Errors from closing output files/writers may need different handling.
- The user now fetches pods before the Ollama call, marshals the `[]PodInfo` to JSON, and sends that string plus the question in `OllamaMessage.Content`. This is a straightforward way to provide prompt context.
- The ignored Kubernetes error is now handled and wrapped; Ollama is called only after pod retrieval succeeds. `CreatedAt` now uses the correct `created_at` JSON tag.
- User ran the end-to-end CLI and confirmed it works; the model responds using the Kubernetes pod data.
- Next small task: make the prompt more explicit and grounded: label the serialized pod data and the user's question, and instruct the model to answer only from the provided data or say when the data is insufficient. Include the namespace in the prompt context. User should edit the prompt and retry a question; keep not adding model tools or cluster write capabilities.
- User added `buildPrompt`, which JSON-encodes pod data and includes a system-prompt label, namespace, pod status, and question. This addresses the prompt-context task.
- Review note: `SYSTEM_PROMPT` says “Used only...” (should be “Use only...”) and is embedded in a message whose actual role is `"user"`; the text label “System Prompt” does not make it an actual system-role message. Prefer a separate `OllamaMessage{Role: "system", Content: SYSTEM_PROMPT}` and a user message for namespace/data/question. Alternatively, keep one user message and remove the “System Prompt” label. Go constants conventionally use mixedCaps (e.g. `ollamaModel`), not all-caps underscores.
- Next step: make that prompt-role/wording cleanup, then run several pod questions and confirm responses are grounded in the supplied data. HTTP context/timeouts can follow.

Latest update: the system-prompt wording is corrected and is now sent as a separate `"system"` message; namespace, serialized pod data, and question are sent in the `"user"` message. The user reports this works well.
- User wants the next feature to expose pod listing as an Ollama tool and let the model decide from the initial question whether it needs cluster data. Desired flow: first send question plus a read-only `list_pods` tool definition; if the model requests that tool, the Go app validates the tool name, executes `getK8sPods` only for the already-selected CLI namespace, sends the tool result back as a tool message, and asks Ollama for the final answer. If no tool call is returned, use the direct answer. Never allow the model to execute arbitrary commands or modify Kubernetes.
- Tool calls are requests from the model, not direct function execution; the Go app must dispatch only explicitly allowed tool names/arguments. Confirm `gemma4:e2b` supports tool calling; if it does not, switch to a compatible local model or retain the current always-fetch flow.
- Next small coding task: represent a single `list_pods` tool in the Ollama request's `tools` field (read-only, namespace comes from CLI, no model-controlled namespace), then inspect the model's response shape/tool call without executing it yet. Follow with a separate task to dispatch that call and send results back.
- User is currently defining the Go structs for the tool schema. The `Parameters` nested value needs a typed composite literal; a named `FunctionParameters` struct with `Type`, `Properties map[string]any`, and `Required []string` will make construction straightforward. Empty values should be initialized as `map[string]any{}` and `[]string{}` to serialize as `{}` and `[]`.
- Tool-call response handling: check `response.Message.ToolCalls` (JSON field `message.tool_calls`). An empty slice means no tool call; a non-empty slice means the model requested one. Decode tool-call function name/arguments and dispatch only a known allowlisted name such as `list_pods`; never execute arbitrary names.
- Tool-call round trip: keep the original system/user messages; append the assistant message containing `tool_calls`, execute the allowlisted Go function, append its result as a message with role `"tool"`, then send the conversation again to Ollama for the final answer. The tool result is not put into the original question or the top-level `tools` definitions.
- Current `run` line 77 declares a new block-local `response` with `:=`, shadowing the outer `response` returned by the first call; the new local is unused, while output after the `if` still reads the first response. Use assignment to the outer variable if that is the intended result.
- More importantly, the second `ollamaRequest(prompt)` builds a fresh request, so it discards the appended assistant tool-call and tool-result messages. Refactor the request sender to accept the updated `OllamaRequest` (or otherwise preserve the conversation), then use the second response for final output.

Latest review of tool calling:
- User refactored `ollamaRequest` to accept/return an `OllamaRequest`; `run` appends the assistant tool-call message and a tool result, then resends the updated request. This preserves the conversation across the round trip.
- JSON tags are present on `Tools`, `Function`, and `FunctionParameters`; the previous concern that they were missing was based on an older file view and is resolved.
- `SYSTEM_PROMPT` hardcodes the namespace as `default`, which conflicts with the `-namespace` option; remove that claim or make it dynamic.
- The current implementation handles one tool-call round; chained calls can be supported later.
- Next: fix the JSON tags and namespace prompt inconsistency, then test one question that should invoke `list_pods` and one that should not. Verify tool results reach the model and unknown tools fail safely.

Current tool-calling code review:
- The second request now uses the updated `OllamaRequest`, so the assistant tool-call message and tool result are retained.
- JSON tags are present on the tool-definition structs.
- Ollama's documented tool-call response includes `"type":"function"` on each tool call and an `index` in its nested function object. Since the app appends and reserializes the assistant message, add those fields to the response structs to preserve the tool-call shape.
- Tool result messages should include `tool_name` matching the invoked tool; add `ToolName string `json:"tool_name,omitempty"` to the message type and set it for role `"tool"`.
- `SYSTEM_PROMPT` hardcodes namespace `default`, conflicting with `-namespace`; remove that claim or make it dynamic. Unknown tool names now return an error.
- Next: fix the JSON/schema round-trip fields and tool-name association, make the namespace instruction accurate, then test one question that should invoke `list_pods` and one that should not. Verify unknown tool names fail safely. Handling multiple tool-call rounds can follow later.

Latest review update: the request/response structs preserve the tool call's `type` and nested function `index`, and role `"tool"` messages include `tool_name`. A focused JSON round-trip test covers these fields; `go test ./...` passes. End-to-end tool-call behavior against Ollama still needs user verification.

Next planned improvements:
- The code is split into packages: `main.go` delegates to `internal/app`; Kubernetes queries live in `internal/k8s`; Ollama client and protocol types live in `internal/ollama`; `internal/tools` owns tool definitions and allowlisted dispatch. Package boundaries require exported identifiers for cross-package calls and imports based on the `module` path in `go.mod`.
- Add another read-only Kubernetes tool, starting with namespace-scoped deployment listing.
- Refactor tool dispatch so Ollama-requested tools execute concurrently with goroutines.
- Preserve tool-call order and avoid concurrent writes to the shared request/messages state.
- Improve Ollama HTTP cancellation and add a timeout once the tool flow is stable.
- Send pod/deployment data to the model as JSON rather than formatted Go struct strings.

## Session handoff

Latest status: the pod-listing flow, Ollama chat integration, and native tool-call round trip are implemented. Tool-call protocol fields for `type`, nested function `index`, and `tool_name` are modeled. `internal/tools` now owns the model-visible tool definition and dispatches the allowlisted `list_pods` call; `main.go` calls the exported `app.Run`. Next, add a second read-only Kubernetes tool, then implement concurrent dispatch while preserving tool-call order.

Start from the current state above. The Podman-backed kind cluster `k8s-assistant` is running and kubectl context is configured. Continue in small, focused coding tasks: user wants to implement code and receive review, not get full solutions upfront. Explain Go through Java/Spring comparisons, mention idioms and good practices, and give terminal commands for the user to run rather than running them. Don't assume Ollama model download has completed.
