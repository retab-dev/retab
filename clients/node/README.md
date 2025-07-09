# Retab Node.js SDK

Official Node.js SDK for the Retab API - extract structured data from documents using AI.

## Installation

### From npm Registry

```bash
npm install @retab/node
```

### For Local Development

To install and test the SDK locally using npm link:

```bash
# From the SDK directory
cd open-source/sdk/clients/node
npm install
npm run build
npm link

# From your project directory
npm link @retab/node
```

## Quick Start

```typescript
import { Retab } from '@retab/node';

// Initialize the client
const retab = new Retab({
  apiKey: 'your-api-key', // Or set RETAB_API_KEY environment variable
});

// Extract data from a document
const extraction = await retab.documents.extract({
  document: 'path/to/invoice.pdf',
  schema: {
    json_schema: {
      type: 'object',
      properties: {
        invoice_number: { type: 'string' },
        total_amount: { type: 'number' },
        due_date: { type: 'string', format: 'date' }
      }
    }
  }
});

console.log(extraction.extracted_data);
```

## Key Features

### Document Processing
Extract structured data from various document formats:

```typescript
// From file path
const result = await retab.documents.extract({
  document: 'path/to/document.pdf',
  schema: mySchema
});

// From buffer
const buffer = fs.readFileSync('document.pdf');
const result = await retab.documents.extract({
  document: buffer,
  schema: mySchema
});

// With specific AI model
const result = await retab.documents.extract({
  document: 'document.pdf',
  schema: mySchema,
  model: 'gpt-4.1'
});
```

### Schema Management
Generate and enhance JSON schemas for your documents:

```typescript
// Generate schema from sample documents
const schema = await retab.schemas.generate({
  documents: ['invoice1.pdf', 'invoice2.pdf'],
  instructions: 'Extract invoice data including line items'
});

// Enhance existing schema with ground truth data
const enhanced = await retab.schemas.enhance({
  schema_id: 'existing-schema-id',
  documents: ['sample.pdf'],
  ground_truths: [{ invoice_total: 150.00 }]
});
```

### Multi-Provider AI Support
Use different AI providers by passing API keys via headers:

```typescript
const retab = new Retab({
  apiKey: 'your-retab-api-key',
  defaultHeaders: {
    'X-OpenAI-Api-Key': 'your-openai-key',
    'X-Anthropic-Api-Key': 'your-anthropic-key',
    'X-xAI-Api-Key': 'your-xai-key',
    'X-Gemini-Api-Key': 'your-gemini-key'
  }
});
```

### Consensus Extraction
Get multiple AI opinions for higher accuracy:

```typescript
const consensus = await retab.consensus.extract({
  document: 'contract.pdf',
  schema: contractSchema,
  num_extractions: 3,
  models: ['gpt-4o', 'claude-3-5-sonnet-latest', 'grok-2-latest']
});
```

### Automation & Webhooks
Set up automated document processing:

```typescript
// Create webhook endpoint
const automation = await retab.processors.endpoints.create({
  name: 'Invoice Processor',
  processor_id: 'proc_9dfnjkf2912',
  webhook_url: 'https://your-app.com/webhook'
});

```

### Schema Types
The SDK supports multiple schema formats:

```typescript
import { Schema } from '@retab/node';

// From JSON Schema
const schema = Schema.from_json_schema({
  type: 'object',
  properties: {
    name: { type: 'string' }
  }
});

// From Pydantic model (as JSON)
const schema = Schema.from_pydantic_json({
  type: 'object',
  properties: { /* ... */ }
});

// From Zod schema (as JSON)
const schema = Schema.from_zod_json({
  /* zod schema JSON representation */
});
```

### Streaming Responses
For real-time updates:

```typescript
const stream = await retab.documents.extract({
  document: 'large-document.pdf',
  schema: mySchema,
  stream: true
});

for await (const chunk of stream) {
  console.log('Progress:', chunk);
}
```

### Idempotency
Ensure requests are processed only once:

```typescript
const result = await retab.documents.extract({
  document: 'invoice.pdf',
  schema: mySchema
}, {
  headers: {
    'Idempotency-Key': 'unique-request-id'
  }
});
```

## Configuration

### Client Options

```typescript
const retab = new Retab({
  apiKey: 'your-api-key',           // Required
  baseUrl: 'https://api.retab.com', // Optional, defaults to production
  timeout: 240000,                  // Optional, request timeout in ms
  maxRetries: 3,                    // Optional, retry count
  defaultHeaders: {                 // Optional, additional headers
    'X-Custom-Header': 'value'
  }
});
```

### Environment Variables

- `RETAB_API_KEY`: Your Retab API key
- `RETAB_API_BASE_URL`: Override the default API base URL
- `OPENAI_API_KEY`: OpenAI API key (if not using headers)
- `ANTHROPIC_API_KEY`: Anthropic API key (if not using headers)
- `XAI_API_KEY`: xAI API key (if not using headers)
- `GEMINI_API_KEY`: Gemini API key (if not using headers)

## TypeScript Support

The SDK is written in TypeScript and provides comprehensive type definitions:

```typescript
import type {
  Document,
  Schema,
  ExtractionResult,
  Model,
  BrowserCanvasSize
} from '@retab/node';
```

## Error Handling

The SDK provides detailed error types:

```typescript
import { RetabError } from '@retab/node';

try {
  const result = await retab.documents.extract({
    document: 'file.pdf',
    schema: mySchema
  });
} catch (error) {
  if (error instanceof RetabError) {
    console.error('API Error:', error.message);
    console.error('Status:', error.status);
    console.error('Details:', error.details);
  }
}
```

## Development

### Building from Source

```bash
npm install
npm run build
```

### Running Tests

```bash
npm test              # Run all tests
npm run test:watch    # Watch mode
npm run test:coverage # With coverage
```

### Linting and Formatting

```bash
npm run lint    # Check for linting errors
npm run format  # Format code with Prettier
```

## Requirements

- Node.js >= 16.0.0
- TypeScript >= 5.0.0 (for development)

## License

MIT

## Support

- Documentation: https://docs.retab.com
- Issues: https://github.com/retab-inc/retab/issues
- Discord: https://discord.gg/retab