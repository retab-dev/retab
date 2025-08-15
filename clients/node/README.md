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
cd open_source/sdk/clients/node
bun install
bun run build
bun link

# From your project directory
bun link @retab/node

bun test tests/
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
```

## Key Features




## Support

- Documentation: https://docs.retab.com
- Issues: https://github.com/retab-inc/retab/issues
- Discord: https://discord.gg/retab