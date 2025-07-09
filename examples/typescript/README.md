# Retab TypeScript Examples

This directory contains TypeScript versions of all Retab SDK examples, demonstrating type-safe usage of the SDK.

## Setup

```bash
bun install
```

## Structure

- `automations/` - Webhook endpoints, email automation, and link management examples
- `consensus/` - Multi-model consensus extraction examples
- `documents/` - Document extraction and processing examples
- `processors/` - Processor management examples
- `schemas/` - Schema creation examples (JSON Schema, Pydantic, Zod)
- `utilities/` - Utility functions and helpers

## Running Examples

You can run any example directly with Bun:

```bash
# Document extraction
bun run documents/extract_api.ts

# Schema examples
bun run schemas/json_schema_calendar_event.ts

# Consensus extraction
bun run consensus/extract_api.ts
```

## Environment Variables

Create a `.env` file with:

```
RETAB_API_KEY=your-api-key
OPENAI_API_KEY=your-openai-key  # Optional
ANTHROPIC_API_KEY=your-anthropic-key  # Optional
XAI_API_KEY=your-xai-key  # Optional
GEMINI_API_KEY=your-gemini-key  # Optional
```

## Type Checking

Run TypeScript type checking:

```bash
bun run typecheck
```

## Notes

- All examples use ES modules and modern TypeScript features
- The SDK provides full TypeScript support with comprehensive type definitions
- Examples demonstrate both sync and async patterns
- Some examples require specific API keys or resources to run

## Known Issues

- The blake2 native module may have compatibility issues with Bun on some platforms
- For production use, consider using Node.js with `tsx` or `ts-node`

## Alternative Running Methods

If you encounter issues with Bun, you can also run the examples with Node.js:

```bash
# Install tsx globally
npm install -g tsx

# Run an example
tsx documents/extract_api.ts
```