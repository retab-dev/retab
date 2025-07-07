# Retab Node.js SDK

Official Node.js SDK for Retab API

## Installation

```bash
npm install retab
```

## Usage

```typescript
import { Retab } from 'retab';

const retab = new Retab({
  apiKey: 'your-api-key', // Or set RETAB_API_KEY environment variable
});

// Example: Generate a schema
const schema = await retab.schemas.generate({
  documents: ['path/to/document.pdf'],
  instructions: 'Extract invoice data',
});
```

## Configuration

The SDK can be configured with the following options:

- `apiKey`: Your Retab API key (required)
- `baseUrl`: API base URL (defaults to https://api.retab.com)
- `timeout`: Request timeout in milliseconds (defaults to 240000)
- `maxRetries`: Maximum number of retries (defaults to 3)

## Environment Variables

- `RETAB_API_KEY`: Your Retab API key
- `RETAB_API_BASE_URL`: Override the default API base URL
- `OPENAI_API_KEY`: OpenAI API key (optional)
- `GEMINI_API_KEY`: Gemini API key (optional)

## License

MIT