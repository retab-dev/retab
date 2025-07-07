import { z } from 'zod';
import { ChatCompletionRetabMessageSchema } from '../chat.js';

export const PartialSchemaSchema = z.object({
  object: z.literal('schema'),
  created_at: z.string().datetime(),
  json_schema: z.record(z.any()).default({}),
  strict: z.boolean().default(true),
});
export type PartialSchema = z.infer<typeof PartialSchemaSchema>;

export const PartialSchemaChunkSchema = z.object({
  object: z.literal('schema.chunk'),
  created_at: z.string().datetime(),
  delta_json_schema_flat: z.record(z.any()).default({}),
  streaming_error: z.custom<import('../standards.js').ErrorDetail>().nullable().optional(),
});
export type PartialSchemaChunk = z.infer<typeof PartialSchemaChunkSchema>;

export const SchemaSchema = PartialSchemaSchema.extend({}).transform((data) => new Schema(data));

export class Schema implements PartialSchema {
  object: 'schema' = 'schema';
  created_at: string;
  json_schema: Record<string, any> = {};
  strict: boolean = true;
  private _zodModel?: z.ZodType<any>;

  constructor(data: { json_schema?: Record<string, any>; pydanticModel?: any; zod_model?: z.ZodType<any>; system_prompt?: string; reasoning_prompts?: Record<string, string> }) {
    this.created_at = new Date().toISOString();
    
    if (data.json_schema) {
      this.json_schema = data.json_schema;
    } else if (data.pydanticModel) {
      // In a real implementation, this would extract schema from the model
      this.json_schema = {};
    } else if (data.zod_model) {
      this._zodModel = data.zod_model;
      // Convert Zod to JSON Schema (simplified)
      this.json_schema = {
        type: 'object',
        properties: {},
        additionalProperties: true
      };
    } else {
      throw new Error('Must provide either json_schema, pydanticModel, or zod_model');
    }
  }

  get dataId(): string {
    // In a real implementation, this would generate SHA1 hash of schema data
    return 'schema-data-id';
  }

  get id(): string {
    // In a real implementation, this would generate SHA1 hash of complete schema
    return 'schema-id';
  }

  get inference_json_schema(): Record<string, any> {
    // Returns the schema formatted for structured output with OpenAI requirements
    const schema = { ...this.json_schema };
    if (!schema.hasOwnProperty('additionalProperties')) {
      schema.additionalProperties = false;
    }
    return schema;
  }

  get inferenceJsonSchema(): Record<string, any> {
    // Alias for backwards compatibility
    return this.inference_json_schema;
  }

  get openaiMessages(): any[] {
    // Returns messages formatted for OpenAI's API
    return this.messages.map(msg => ({
      role: msg.role,
      content: msg.content
    }));
  }

  get anthropicSystemPrompt(): string {
    return 'Return your response as a JSON object following the provided schema.' + this.systemPrompt;
  }

  get anthropicMessages(): any[] {
    // Returns messages in Anthropic's Claude format
    return this.messages.slice(1); // Skip system message
  }

  get geminiSystemPrompt(): string {
    return this.systemPrompt;
  }

  get geminiMessages(): any[] {
    // Returns messages formatted for Google's Gemini API
    return this.messages.slice(1);
  }

  get inferenceGeminiJsonSchema(): Record<string, any> {
    // Convert schema for Gemini compatibility (no anyOf, etc.)
    const schema = { ...this._reasoningObjectSchema };
    // Implementation would remove unsupported fields like anyOf
    return schema;
  }

  get inferenceTypescriptInterface(): string {
    // Returns TypeScript interface representation
    return '// TypeScript interface would be generated here';
  }

  get inferenceNlpDataStructure(): string {
    // Returns NLP data structure representation
    return '// NLP data structure would be generated here';
  }

  get developerSystemPrompt(): string {
    return `
# General Instructions

You are an expert in data extraction and structured data outputs.

When provided with a **JSON schema** and a **document**, you must:

1. Carefully extract all relevant data from the provided document according to the given schema.
2. Return extracted data strictly formatted according to the provided schema.
3. Make sure that the extracted values are **UTF-8** encodable strings.
4. Avoid generating bytes, binary data, base64 encoded data, or other non-UTF-8 encodable data.

---

## Date and Time Formatting

When extracting date, time, or datetime values:

- **Always use ISO format** for dates and times (e.g., "2023-12-25", "14:30:00", "2023-12-25T14:30:00Z")
- **Include timezone information** when available (e.g., "2023-12-25T14:30:00+02:00")
- **Use UTC timezone** when timezone is not specified or unclear (e.g., "2023-12-25T14:30:00Z")
- **Maintain precision** as found in the source document (seconds, milliseconds if present)

---

## Handling Missing and Nullable Fields

### Nullable Leaf Attributes

- If valid data is missing or not explicitly present, set leaf attributes explicitly to \`null\`.
- **Do NOT** use empty strings (\`""\`), placeholder values, or fabricated data.

### Nullable Nested Objects

- If an entire nested object's data is missing or incomplete, **do NOT** set the object itself to \`null\`.
- Keep the object structure fully intact, explicitly setting each leaf attribute within to \`null\`.
- This preserves overall structure and explicitly communicates exactly which fields lack data.

---

## Reasoning Fields

Your schema includes special reasoning fields (\`reasoning___*\`) used exclusively to document your extraction logic. These fields are for detailed explanations and will not appear in final outputs.

You MUST include these details explicitly in your reasoning fields:

- **Explicit Evidence**: Quote specific lines or phrases from the document confirming your extraction.
- **Decision Justification**: Clearly justify why specific data was chosen or rejected.
- **Calculations/Transformations**: Document explicitly any computations, unit conversions, or normalizations.
- **Alternative Interpretations**: Explicitly describe any alternative data interpretations considered and why you rejected them.
- **Confidence and Assumptions**: Clearly state your confidence level and explicitly articulate any assumptions.

---

## Source Fields

Some leaf fields require you to explicitly provide the source of the data (verbatim from the document).
The idea is to simply provide a verbatim quote from the document, without any additional formatting or commentary, keeping it as close as possible to the original text.
Make sure to reasonably include some surrounding text to provide context about the quote.

You can easily identify the fields that require a source by the \`quote___[attributename]\` naming pattern.

---

# User Defined System Prompt

`;
  }

  get userSystemPrompt(): string {
    return this.json_schema['X-SystemPrompt'] || '';
  }

  get schemaSystemPrompt(): string {
    return (
      this.inferenceNlpDataStructure + 
      '\n---\n' + 
      '## Expected output schema as a TypeScript interface for better readability:\n\n' + 
      this.inferenceTypescriptInterface
    );
  }

  get systemPrompt(): string {
    return this.developerSystemPrompt + '\n\n' + this.userSystemPrompt + '\n\n' + this.schemaSystemPrompt;
  }

  get title(): string {
    return this.json_schema.title || 'NoTitle';
  }

  private get _expandedObjectSchema(): Record<string, any> {
    // Returns schema with all references expanded inline
    return this.json_schema; // Simplified for now
  }

  private get _reasoningObjectSchema(): Record<string, any> {
    // Returns schema with inference-specific modifications
    return this._expandedObjectSchema; // Simplified for now
  }


  get messages(): z.infer<typeof ChatCompletionRetabMessageSchema>[] {
    return [{ role: 'developer', content: this.systemPrompt }];
  }

  get openai_messages(): Array<{ role: string; content: string }> {
    return [{ role: 'developer', content: this.systemPrompt }];
  }

  get zod_model(): z.ZodType<any> {
    if (this._zodModel) {
      return this._zodModel;
    }
    // Convert JSON schema to basic Zod schema for validation
    return z.object({}).passthrough();
  }

  getPatternAttribute(_pattern: string, _attribute: 'X-FieldPrompt' | 'X-ReasoningPrompt' | 'type'): string | null {
    // Navigate schema and return the specified attribute
    return null; // Simplified for now
  }

  setPatternAttribute(
    _pattern: string, 
    _attribute: 'X-FieldPrompt' | 'X-ReasoningPrompt' | 'X-SystemPrompt' | 'description', 
    _value: string
  ): void {
    // Set attribute value at specific path in schema
    // Simplified for now
  }

  save(_path: string): void {
    // Save JSON schema to file
    // In a real implementation, this would write to filesystem
    throw new Error('save() method not implemented in browser environment');
  }

  static validate(data: any): Schema {
    return new Schema(data);
  }
}