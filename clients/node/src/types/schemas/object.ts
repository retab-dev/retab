import { z } from 'zod';
import { ChatCompletionRetabMessageSchema } from '../chat.js';
import { generateSchemaDataId, generateSchemaId, getPatternAttribute, setPatternAttribute, loadJsonSchema } from '../../utils/json_schema_utils.js';
import { zodToJsonSchema } from '../../utils/zod_to_json_schema.js';

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

  constructor(data: { 
    json_schema?: Record<string, any> | string; 
    pydanticModel?: any; 
    zod_model?: z.ZodType<any>; 
    system_prompt?: string; 
    reasoning_prompts?: Record<string, string> 
  }) {
    this.created_at = new Date().toISOString();
    
    if (data.json_schema) {
      this.json_schema = loadJsonSchema(data.json_schema);
    } else if (data.pydanticModel) {
      // In a real implementation, this would extract schema from the model
      this.json_schema = {};
    } else if (data.zod_model) {
      this._zodModel = data.zod_model;
      // Convert Zod to JSON Schema using proper converter
      this.json_schema = zodToJsonSchema(data.zod_model);
      
      // Add system prompt if provided
      if (data.system_prompt) {
        this.json_schema['X-SystemPrompt'] = data.system_prompt;
      }
      
      // Add reasoning prompts if provided
      if (data.reasoning_prompts) {
        for (const [field, prompt] of Object.entries(data.reasoning_prompts)) {
          if (this.json_schema.properties && this.json_schema.properties[field]) {
            this.json_schema.properties[field]['X-ReasoningPrompt'] = prompt;
          }
        }
      }
    } else {
      throw new Error('Must provide either json_schema, pydanticModel, or zod_model');
    }
  }

  get dataId(): string {
    return generateSchemaDataId(this.json_schema);
  }

  get id(): string {
    return generateSchemaId(this.json_schema);
  }

  get inference_json_schema(): Record<string, any> {
    // Returns the schema formatted for structured output with OpenAI requirements
    if (this.strict) {
      // For strict schemas, convert to OpenAI-compatible format
      const inferenceSchema = this.jsonSchemaToStrictOpenaiSchema(JSON.parse(JSON.stringify(this._reasoningObjectSchema)));
      if (typeof inferenceSchema !== 'object' || inferenceSchema === null) {
        throw new Error('Validation Error: The inference_json_schema is not a dict');
      }
      return inferenceSchema;
    } else {
      // For non-strict schemas, return a deep copy of the reasoning schema
      return JSON.parse(JSON.stringify(this._reasoningObjectSchema));
    }
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
    // Returns TypeScript interface representation of the inference schema
    return this.jsonSchemaToTypescriptInterface(this._reasoningObjectSchema);
  }

  get inferenceNlpDataStructure(): string {
    // Returns NLP data structure representation of the inference schema
    return this.jsonSchemaToNlpDataStructure(this._reasoningObjectSchema);
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
    return this.expandRefs(JSON.parse(JSON.stringify(this.json_schema)));
  }

  private get _reasoningObjectSchema(): Record<string, any> {
    // Returns schema with inference-specific modifications (reasoning fields added)
    return this.createReasoningSchema(JSON.parse(JSON.stringify(this._expandedObjectSchema)));
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

  getPatternAttribute(pattern: string, attribute: 'X-FieldPrompt' | 'X-ReasoningPrompt' | 'type'): string | null {
    return getPatternAttribute(this.json_schema, pattern, attribute);
  }

  setPatternAttribute(
    pattern: string, 
    attribute: 'X-FieldPrompt' | 'X-ReasoningPrompt' | 'X-SystemPrompt' | 'description', 
    value: string
  ): void {
    setPatternAttribute(this.json_schema, pattern, attribute, value);
  }

  save(_path: string): void {
    // Save JSON schema to file
    // In a real implementation, this would write to filesystem
    throw new Error('save() method not implemented in browser environment');
  }

  static validate(data: any): Schema {
    return new Schema(data);
  }

  private createReasoningSchema(schema: Record<string, any>): Record<string, any> {
    // Add reasoning fields to the schema structure
    function addReasoningFields(obj: Record<string, any>, path: string = ''): void {
      if (obj.type === 'object' && obj.properties) {
        // Add reasoning field for the object itself
        const reasoningField = path ? `reasoning___${path.split('.').pop()}` : 'reasoning___root';
        obj.properties[reasoningField] = {
          type: 'string',
          description: 'Reasoning for this object'
        };

        // Process each property
        for (const [key, prop] of Object.entries(obj.properties)) {
          if (key.startsWith('reasoning___')) continue; // Skip already added reasoning fields
          
          const newPath = path ? `${path}.${key}` : key;
          if (typeof prop === 'object' && prop !== null) {
            const typedProp = prop as any;
            if (typedProp.type === 'array' && typedProp.items) {
              // Add reasoning field for the array
              obj.properties[`reasoning___${key}`] = {
                type: 'string',
                description: `Reasoning for array ${key}`
              };
              
              // Process array items
              if (typeof typedProp.items === 'object') {
                addReasoningFields(typedProp.items, `${newPath}.item`);
              }
            } else if (typedProp.type === 'object') {
              addReasoningFields(typedProp, newPath);
            } else {
              // Add reasoning field for leaf properties
              obj.properties[`reasoning___${key}`] = {
                type: 'string',
                description: `Reasoning for ${key}`
              };
            }
          }
        }
      }
    }

    addReasoningFields(schema);
    return schema;
  }

  private jsonSchemaToStrictOpenaiSchema(schema: Record<string, any>): Record<string, any> {
    // Convert schema to OpenAI strict format
    function makeStrict(obj: Record<string, any>): Record<string, any> {
      const result = { ...obj };
      
      // Ensure additionalProperties is false for objects
      if (result.type === 'object') {
        result.additionalProperties = false;
        
        if (result.properties) {
          const newProperties: Record<string, any> = {};
          for (const [key, prop] of Object.entries(result.properties)) {
            newProperties[key] = makeStrict(prop as Record<string, any>);
          }
          result.properties = newProperties;
        }
      }
      
      // Process array items
      if (result.type === 'array' && result.items) {
        result.items = makeStrict(result.items);
      }
      
      // Handle anyOf by taking the first non-null option
      if (result.anyOf) {
        const nonNullOptions = result.anyOf.filter((option: any) => option.type !== 'null');
        if (nonNullOptions.length > 0) {
          const chosen = makeStrict(nonNullOptions[0]);
          // Merge the chosen option into the result
          Object.assign(result, chosen);
          delete result.anyOf;
        }
      }
      
      // Remove unsupported fields
      delete result.format;
      
      return result;
    }
    
    return makeStrict(schema);
  }

  private expandRefs(schema: Record<string, any>): Record<string, any> {
    // Expand $ref references inline
    function expand(obj: any, defs: Record<string, any>): any {
      if (typeof obj !== 'object' || obj === null) {
        return obj;
      }
      
      if (Array.isArray(obj)) {
        return obj.map(item => expand(item, defs));
      }
      
      if (obj.$ref) {
        // Handle $ref
        const refPath = obj.$ref;
        if (refPath.startsWith('#/$defs/')) {
          const defName = refPath.substring('#/$defs/'.length);
          if (defs[defName]) {
            return expand(defs[defName], defs);
          }
        }
        return obj; // Return as-is if ref not found
      }
      
      const result: Record<string, any> = {};
      for (const [key, value] of Object.entries(obj)) {
        result[key] = expand(value, defs);
      }
      return result;
    }
    
    const defs = schema.$defs || {};
    return expand(schema, defs);
  }

  private jsonSchemaToTypescriptInterface(schema: Record<string, any>): string {
    // Convert JSON schema to TypeScript interface
    function convertType(obj: any, depth: number = 0): string {
      const indent = '  '.repeat(depth);
      
      if (!obj || typeof obj !== 'object') {
        return 'any';
      }
      
      if (obj.type === 'string') return 'string';
      if (obj.type === 'number' || obj.type === 'integer') return 'number';
      if (obj.type === 'boolean') return 'boolean';
      if (obj.type === 'null') return 'null';
      
      if (obj.type === 'array') {
        const itemType = obj.items ? convertType(obj.items, depth) : 'any';
        return `${itemType}[]`;
      }
      
      if (obj.type === 'object' && obj.properties) {
        const props = Object.entries(obj.properties)
          .map(([key, prop]: [string, any]) => {
            const optional = !obj.required?.includes(key) ? '?' : '';
            const type = convertType(prop, depth + 1);
            const desc = prop.description ? ` // ${prop.description}` : '';
            return `${indent}  ${key}${optional}: ${type};${desc}`;
          })
          .join('\n');
        
        return `{\n${props}\n${indent}}`;
      }
      
      if (obj.anyOf) {
        return obj.anyOf.map((subSchema: any) => convertType(subSchema, depth)).join(' | ');
      }
      
      return 'any';
    }
    
    const interfaceName = schema.title || 'Schema';
    const interfaceBody = convertType(schema, 0);
    
    return `interface ${interfaceName} ${interfaceBody}`;
  }

  private jsonSchemaToNlpDataStructure(schema: Record<string, any>): string {
    // Convert JSON schema to natural language data structure description
    function describe(obj: any, depth: number = 0): string {
      const indent = '  '.repeat(depth);
      
      if (!obj || typeof obj !== 'object') {
        return 'any value';
      }
      
      if (obj.description) {
        return obj.description;
      }
      
      if (obj.type === 'string') return 'text string';
      if (obj.type === 'number' || obj.type === 'integer') return 'number';
      if (obj.type === 'boolean') return 'true/false value';
      if (obj.type === 'null') return 'null value';
      
      if (obj.type === 'array') {
        const itemDesc = obj.items ? describe(obj.items, depth) : 'any item';
        return `array of ${itemDesc}`;
      }
      
      if (obj.type === 'object' && obj.properties) {
        const props = Object.entries(obj.properties)
          .map(([key, prop]: [string, any]) => {
            const optional = !obj.required?.includes(key) ? ' (optional)' : '';
            const desc = describe(prop, depth + 1);
            return `${indent}- ${key}${optional}: ${desc}`;
          })
          .join('\n');
        
        return `object containing:\n${props}`;
      }
      
      if (obj.anyOf) {
        return `one of: ${obj.anyOf.map((subSchema: any) => describe(subSchema, depth)).join(', ')}`;
      }
      
      return 'value';
    }
    
    return describe(schema, 0);
  }
}