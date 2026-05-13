# Retab Go SDK

Official Go client for the Retab API.

This Go SDK mirrors the Node SDK service tree:

- `client.Files`, `client.Schemas`
- `client.Extractions`, `client.Parses`, `client.Splits`, `client.Classifications`, `client.Partitions`
- `client.Edits` and `client.Edits.Templates`
- `client.Jobs`
- `client.Workflows`, including `Runs`, `Runs.Steps`, `Blocks`, `Edges`, `Tests`, and `Tests.Runs`

`Steps.List` returns the step roster/status summaries for a run. `Steps.Get` requires both the run ID and block ID and returns the full execution record for that single step.

## Install

```bash
go get github.com/retab-dev/retab/clients/go
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	retab "github.com/retab-dev/retab/clients/go"
)

func main() {
	ctx := context.Background()

	client, err := retab.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	run, err := client.Workflows.Runs.Get(ctx, "run_abc123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(run.ID, run.Lifecycle.Status)

	stepSummaries, err := client.Workflows.Runs.Steps.List(ctx, run.ID)
	if err != nil {
		log.Fatal(err)
	}
	for _, summary := range stepSummaries {
		step, err := client.Workflows.Runs.Steps.Get(ctx, run.ID, summary.BlockID)
		if err != nil {
			log.Fatal(err)
		}
		if len(step.HandleOutputs) > 0 {
			fmt.Println(step.BlockID, step.HandleOutputs)
		}
	}
}
```

Set `RETAB_API_KEY` or pass the API key to `NewClient`.

```go
client, err := retab.NewClient("sk_retab_...")
```

Use `WithBaseURL` for local development:

```go
client, err := retab.NewClient(
	"sk_retab_...",
	retab.WithBaseURL("http://localhost:4000/v1"),
)
```

## Processing Documents

```go
document, err := retab.InferMIMEData("invoice.pdf")
if err != nil {
	log.Fatal(err)
}

extraction, err := client.Extractions.Create(ctx, retab.ExtractionCreateRequest{
	Document:   document,
	JSONSchema: map[string]any{"type": "object"},
	Model:      "retab-small",
})
if err != nil {
	log.Fatal(err)
}
fmt.Println((*extraction)["id"])
```

Upload a local file to Retab storage:

```go
uploaded, err := client.Files.Upload(ctx, "invoice.pdf")
if err != nil {
	log.Fatal(err)
}
fmt.Println(uploaded.Filename, uploaded.URL, uploaded.ID())
```

## Workflows

```go
run, err := client.Workflows.Runs.Create(ctx, retab.CreateWorkflowRunRequest{
	WorkflowID: "workflow_abc123",
	Documents: map[string]any{
		"start-1": "invoice.pdf",
	},
	JSONInputs: map[string]any{
		"vendor": "Retab",
	},
})
if err != nil {
	log.Fatal(err)
}

step, err := client.Workflows.Runs.Steps.Get(ctx, run.ID, "extract-1")
if err != nil {
	log.Fatal(err)
}
fmt.Println(step.HandleInputs, step.HandleOutputs)
```

Streams must be closed. This snippet uses `io.EOF` from the standard library:

```go
stream, err := client.Extractions.CreateStream(ctx, retab.ExtractionCreateRequest{
	Document:   document,
	JSONSchema: map[string]any{"type": "object"},
	Model:      "retab-small",
})
if err != nil {
	log.Fatal(err)
}
defer stream.Close()

for {
	chunk, err := stream.Next()
	if err == io.EOF {
		break
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(chunk)
}
```

Per-call request overrides mirror the Node SDK's `RequestOptions` escape hatch:

```go
params := url.Values{}
params.Set("debug", "1")

partition, err := client.Partitions.Create(
	ctx,
	retab.PartitionCreateRequest{
		Document:     retab.MIMEData{Filename: "invoice.pdf", URL: "https://example.com/invoice.pdf"},
		Key:          "vendor",
		Instructions: "group by vendor",
		Model:        "retab-small",
	},
	retab.WithRequestParams(params),
	retab.WithRequestHeader("X-Debug", "true"),
	retab.WithRequestBody(map[string]any{"model": "retab-large"}),
)
if err != nil {
	log.Fatal(err)
}
_ = partition
```
