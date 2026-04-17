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
import { Retab, raiseForStatus } from "@retab/node";

const retab = new Retab({
  apiKey: "your-api-key",
});

const run = await retab.workflows.runs.createAndWait({
  workflowId: "wf_abc123",
  documents: { "start-node-1": "invoice.pdf" },
  onStatus: (currentRun) => console.log(currentRun.status),
});

raiseForStatus(run);

const steps = await retab.workflows.runs.steps.list(run.id);
const outputs = await retab.workflows.runs.steps.getMany(run.id, ["extract-1"]);

console.log(run.final_outputs);
console.log(steps.map((step) => `${step.block_id}: ${step.status}`));
console.log(outputs.outputs["extract-1"]?.handle_outputs);
```

## Support

- Documentation: https://docs.retab.com
- Issues: https://github.com/retab-inc/retab/issues
- Discord: https://discord.gg/retab
