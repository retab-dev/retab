# Retab Node.js SDK

Official Node.js SDK for the Retab API.

## Installation

```bash
npm install @retab/node
```

## Local Development

```bash
cd open-source/sdk/clients/node
bun install
bun run build
bun test tests/
```

## Quick Start

```typescript
import fs from "fs";
import { Retab } from "@retab/node";

const retab = new Retab({
  apiKey: "your-api-key",
});

const extraction = await retab.extractions.create({
  document: "path/to/invoice.pdf",
  model: "retab-small",
  json_schema: {
    type: "object",
    properties: {
      invoice_number: { type: "string" },
      total_amount: { type: "number" },
      due_date: { type: "string", format: "date" },
    },
  },
});

console.log(extraction.output);

const buffer = fs.readFileSync("document.pdf");
const parse = await retab.parses.create({
  document: buffer,
  model: "retab-small",
});

console.log(parse.output.pages[0]);
```

## Workflows

```typescript
import { Retab } from "@retab/node";

const retab = new Retab({
  apiKey: "your-api-key",
});

const run = await retab.workflows.runs.create({
  workflowId: "wf_abc123",
  documents: { "start_document-node-1": "invoice.pdf" },
});

const currentRun = await retab.workflows.runs.get(run.id);
if (currentRun.lifecycle.status === "error") {
  throw new Error(currentRun.lifecycle.message);
}
if (currentRun.lifecycle.status === "cancelled") {
  throw new Error(currentRun.lifecycle.reason ?? "Workflow run was cancelled");
}

const steps = await retab.workflows.steps.list(currentRun.id);
const extractStep = await retab.workflows.steps.get(currentRun.id, "extract-1");
const runArtifacts = await retab.workflows.artifacts.list({
  runId: currentRun.id,
});

console.log(steps.data.map((step) => `${step.block_id}: ${step.lifecycle.status}`));
console.log(extractStep.handle_outputs["output-json-0"]?.data);
console.log(artifact);
console.log(runArtifacts);
```

### Workflow Spec

Use `client.workflows.spec` to validate, plan, apply, and export declarative workflow YAML.

```typescript
const validation = await retab.workflows.spec.validate(yamlDefinition);
const plan = await retab.workflows.spec.plan(yamlDefinition);
const result = await retab.workflows.spec.apply(yamlDefinition);
const exported = await retab.workflows.spec.export(result.workflow_id);
```

A declarative spec uses `apiVersion: workflows.retab.com/v1alpha2` and explicit edge handles:

```yaml
edges:
  - from:
      block: start_document-node
      handle: output-file-0
    to:
      block: extract-node
      handle: input-file-source_doc
```

## Support

- Documentation: https://docs.retab.com
- Issues: https://github.com/retab-inc/retab/issues
- Discord: https://discord.gg/retab
