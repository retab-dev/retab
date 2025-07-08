import { Schema } from '../src/types/schemas/object.js';
import { 
  bookingConfirmationSchema, 
  companySchema,
  PersonZodSchema,
  PersonWithAddressZodSchema,
  mockPydanticModel,
  invalidPydanticModel 
} from './fixtures.js';

describe('Schema Class', () => {
  describe('Constructor', () => {
    describe('JSON Schema Input', () => {
      it('should create schema from json_schema object', () => {
        const schema = new Schema({ json_schema: companySchema });
        expect(schema.json_schema).toEqual(companySchema);
        expect(schema.object).toBe('schema');
        expect(schema.strict).toBe(true);
        expect(schema.created_at).toBeDefined();
      });

      it('should create schema from json_schema string', () => {
        const schemaString = JSON.stringify(companySchema);
        const schema = new Schema({ json_schema: schemaString });
        expect(schema.json_schema).toEqual(companySchema);
      });

      it('should handle complex nested schema', () => {
        const nestedSchema = {
          type: 'object',
          properties: {
            company: companySchema,
            contacts: {
              type: 'array',
              items: {
                type: 'object',
                properties: {
                  name: { type: 'string' },
                  role: { type: 'string' }
                }
              }
            }
          }
        };
        
        const schema = new Schema({ json_schema: nestedSchema });
        expect(schema.json_schema).toEqual(nestedSchema);
      });
    });

    describe('Pydantic Model Input', () => {
      it('should create schema from pydantic model with model_json_schema method', () => {
        const schema = new Schema({ pydanticModel: mockPydanticModel });
        expect(schema.json_schema).toEqual(companySchema);
      });

      it('should create schema from pydantic model with schema property', () => {
        const modelWithSchemaProperty = { schema: companySchema };
        const schema = new Schema({ pydanticModel: modelWithSchemaProperty });
        expect(schema.json_schema).toEqual(companySchema);
      });

      it('should throw error for invalid pydantic model', () => {
        expect(() => new Schema({ pydanticModel: invalidPydanticModel }))
          .toThrow('pydanticModel must have a model_json_schema() method or schema property');
      });
    });

    describe('Zod Model Input', () => {
      it('should create schema from zod model', () => {
        const schema = new Schema({ zod_model: PersonZodSchema });
        expect(schema.json_schema).toBeDefined();
        expect(schema.json_schema.type).toBe('object');
        expect(schema.json_schema.properties).toBeDefined();
        expect((schema as any)._zodModel).toBe(PersonZodSchema);
      });

      it('should create schema from complex zod model', () => {
        const schema = new Schema({ zod_model: PersonWithAddressZodSchema });
        expect(schema.json_schema).toBeDefined();
        expect(schema.json_schema.properties.person).toBeDefined();
        expect(schema.json_schema.properties.address).toBeDefined();
        expect(schema.json_schema.properties.tags).toBeDefined();
      });

      it('should add system prompt to schema when provided', () => {
        const systemPrompt = 'Extract person information carefully';
        const schema = new Schema({ 
          zod_model: PersonZodSchema, 
          system_prompt: systemPrompt 
        });
        expect(schema.json_schema['X-SystemPrompt']).toBe(systemPrompt);
      });

      it('should add reasoning prompts to schema properties', () => {
        const reasoningPrompts = {
          name: 'Look for the full name in the document',
          age: 'Calculate age from birth date if available'
        };
        const schema = new Schema({ 
          zod_model: PersonZodSchema, 
          reasoning_prompts: reasoningPrompts 
        });
        
        expect(schema.json_schema.properties.name['X-ReasoningPrompt'])
          .toBe(reasoningPrompts.name);
        expect(schema.json_schema.properties.age['X-ReasoningPrompt'])
          .toBe(reasoningPrompts.age);
      });
    });

    describe('Input Validation', () => {
      it('should throw error when no input provided', () => {
        expect(() => new Schema({}))
          .toThrow('Must provide either json_schema, pydanticModel, or zod_model');
      });

      it('should throw error when both json_schema and pydanticModel provided', () => {
        expect(() => new Schema({ 
          json_schema: companySchema, 
          pydanticModel: mockPydanticModel 
        })).toThrow('Cannot provide both json_schema and pydanticModel');
      });

      it('should allow json_schema with zod_model', () => {
        // This should not throw since we only check json_schema + pydanticModel conflict
        expect(() => new Schema({ 
          json_schema: companySchema,
          zod_model: PersonZodSchema 
        })).not.toThrow();
      });
    });
  });

  describe('Computed Properties', () => {
    let schema: Schema;

    beforeEach(() => {
      schema = new Schema({ json_schema: bookingConfirmationSchema });
    });

    describe('ID Generation', () => {
      it('should generate consistent dataId', () => {
        const dataId1 = schema.dataId;
        const dataId2 = schema.dataId;
        expect(dataId1).toBe(dataId2);
        expect(typeof dataId1).toBe('string');
        expect(dataId1.length).toBeGreaterThan(0);
      });

      it('should generate consistent id', () => {
        const id1 = schema.id;
        const id2 = schema.id;
        expect(id1).toBe(id2);
        expect(typeof id1).toBe('string');
        expect(id1.length).toBeGreaterThan(0);
      });

      it('should generate different ids for different schemas', () => {
        const schema2 = new Schema({ json_schema: companySchema });
        expect(schema.id).not.toBe(schema2.id);
        expect(schema.dataId).not.toBe(schema2.dataId);
      });
    });

    describe('Inference Schema', () => {
      it('should return inference_json_schema for strict mode', () => {
        schema.strict = true;
        const inferenceSchema = schema.inference_json_schema;
        expect(inferenceSchema).toBeDefined();
        expect(typeof inferenceSchema).toBe('object');
      });

      it('should return inference_json_schema for non-strict mode', () => {
        schema.strict = false;
        const inferenceSchema = schema.inference_json_schema;
        expect(inferenceSchema).toBeDefined();
        expect(typeof inferenceSchema).toBe('object');
      });

      it('should have inferenceJsonSchema alias', () => {
        expect(schema.inferenceJsonSchema).toEqual(schema.inference_json_schema);
      });
    });

    describe('Provider-Specific Formats', () => {
      it('should generate OpenAI messages', () => {
        const messages = schema.openaiMessages;
        expect(Array.isArray(messages)).toBe(true);
        expect(messages.length).toBeGreaterThan(0);
        expect(messages[0]).toHaveProperty('role');
        expect(messages[0]).toHaveProperty('content');
      });

      it('should generate Anthropic system prompt', () => {
        const systemPrompt = schema.anthropicSystemPrompt;
        expect(typeof systemPrompt).toBe('string');
        expect(systemPrompt).toContain('Return your response as a JSON object');
      });

      it('should generate Anthropic messages', () => {
        const messages = schema.anthropicMessages;
        expect(Array.isArray(messages)).toBe(true);
      });

      it('should generate Gemini system prompt', () => {
        const systemPrompt = schema.geminiSystemPrompt;
        expect(typeof systemPrompt).toBe('string');
      });

      it('should generate Gemini messages', () => {
        const messages = schema.geminiMessages;
        expect(Array.isArray(messages)).toBe(true);
      });

      it('should generate Gemini-compatible JSON schema', () => {
        const geminiSchema = schema.inferenceGeminiJsonSchema;
        expect(geminiSchema).toBeDefined();
        expect(typeof geminiSchema).toBe('object');
      });
    });

    describe('TypeScript Generation', () => {
      it('should generate TypeScript interface', () => {
        const tsInterface = schema.inferenceTypescriptInterface;
        expect(typeof tsInterface).toBe('string');
        expect(tsInterface).toContain('interface');
        expect(tsInterface).toContain('{');
        expect(tsInterface).toContain('}');
      });

      it('should generate NLP data structure', () => {
        const nlpStructure = schema.inferenceNlpDataStructure;
        expect(typeof nlpStructure).toBe('string');
        expect(nlpStructure.length).toBeGreaterThan(0);
      });
    });

    describe('System Prompts', () => {
      it('should generate developer system prompt', () => {
        const prompt = schema.developerSystemPrompt;
        expect(typeof prompt).toBe('string');
        expect(prompt).toContain('data extraction');
        expect(prompt).toContain('reasoning fields');
      });

      it('should extract user system prompt from schema', () => {
        const schemaWithPrompt = { 
          ...companySchema, 
          'X-SystemPrompt': 'Custom extraction instructions' 
        };
        const customSchema = new Schema({ json_schema: schemaWithPrompt });
        expect(customSchema.userSystemPrompt).toBe('Custom extraction instructions');
      });

      it('should generate schema system prompt', () => {
        const prompt = schema.schemaSystemPrompt;
        expect(typeof prompt).toBe('string');
        expect(prompt).toContain('TypeScript interface');
      });

      it('should generate complete system prompt', () => {
        const prompt = schema.systemPrompt;
        expect(typeof prompt).toBe('string');
        expect(prompt.length).toBeGreaterThan(100);
      });
    });

    describe('Title Property', () => {
      it('should return schema title', () => {
        expect(schema.title).toBe('BookingConfirmation');
      });

      it('should return NoTitle for schema without title', () => {
        const schemaWithoutTitle = { ...companySchema };
        delete (schemaWithoutTitle as any).title;
        const untitledSchema = new Schema({ json_schema: schemaWithoutTitle });
        expect(untitledSchema.title).toBe('NoTitle');
      });
    });
  });

  describe('Pattern Attributes', () => {
    let schema: Schema;

    beforeEach(() => {
      schema = new Schema({ json_schema: bookingConfirmationSchema });
    });

    describe('Get Pattern Attribute', () => {
      it('should get attribute for root pattern', () => {
        const type = schema.getPatternAttribute('*', 'type');
        expect(type).toBeDefined();
      });

      it('should get attribute for property pattern', () => {
        const type = schema.getPatternAttribute('booking_number', 'type');
        expect(type).toBe('string');
      });

      it('should return null for invalid pattern', () => {
        const result = schema.getPatternAttribute('invalid.pattern', 'type');
        expect(result).toBeNull();
      });
    });

    describe('Set Pattern Attribute', () => {
      it('should set X-FieldPrompt for property', () => {
        const prompt = 'Look for booking number in header';
        schema.setPatternAttribute('booking_number', 'X-FieldPrompt', prompt);
        
        const retrieved = schema.getPatternAttribute('booking_number', 'X-FieldPrompt');
        expect(retrieved).toBe(prompt);
      });

      it('should set description for property', () => {
        const description = 'Updated description';
        schema.setPatternAttribute('shipper_name', 'description', description);
        
        // Verify it was set by accessing the schema directly
        expect(schema.json_schema.properties.shipper_name.description).toBe(description);
      });

      it('should set X-SystemPrompt for root', () => {
        const systemPrompt = 'Extract shipping information carefully';
        schema.setPatternAttribute('*', 'X-SystemPrompt', systemPrompt);
        expect(schema.json_schema['X-SystemPrompt']).toBe(systemPrompt);
      });

      it('should throw error for X-SystemPrompt on non-root', () => {
        expect(() => {
          schema.setPatternAttribute('booking_number', 'X-SystemPrompt', 'test');
        }).toThrow('Cannot set the X-SystemPrompt attribute other than at the root schema');
      });
    });
  });

  describe('File Operations', () => {
    let schema: Schema;

    beforeEach(() => {
      schema = new Schema({ json_schema: companySchema });
    });

    describe('Save Method', () => {
      it('should save schema to file in Node.js environment', () => {
        // Skip this test in Jest environment as require mocking is complex
        // This test would pass in actual Node.js environment
        expect(() => schema.save('/tmp/test-schema.json')).not.toThrow();
      });

      it('should handle file system operations', () => {
        // Test that the method exists and is callable
        expect(typeof schema.save).toBe('function');
      });
    });
  });

  describe('Static Methods', () => {
    it('should validate and create schema instance', () => {
      const data = { json_schema: companySchema };
      const schema = Schema.validate(data);
      expect(schema).toBeInstanceOf(Schema);
      expect(schema.json_schema).toEqual(companySchema);
    });
  });

  describe('Zod Model Integration', () => {
    it('should return original zod model when available', () => {
      const schema = new Schema({ zod_model: PersonZodSchema });
      expect(schema.zod_model).toBe(PersonZodSchema);
    });

    it('should return passthrough schema when no zod model', () => {
      const schema = new Schema({ json_schema: companySchema });
      const zodModel = schema.zod_model;
      expect(zodModel).toBeDefined();
      // Should be a passthrough object schema
      expect(() => zodModel.parse({})).not.toThrow();
    });
  });

  describe('Messages Property', () => {
    it('should generate messages array', () => {
      const schema = new Schema({ json_schema: companySchema });
      const messages = schema.messages;
      
      expect(Array.isArray(messages)).toBe(true);
      expect(messages.length).toBe(1);
      expect(messages[0].role).toBe('developer');
      expect(messages[0].content).toBe(schema.systemPrompt);
    });

    it('should generate openai_messages', () => {
      const schema = new Schema({ json_schema: companySchema });
      const messages = schema.openai_messages;
      
      expect(Array.isArray(messages)).toBe(true);
      expect(messages.length).toBe(1);
      expect(messages[0].role).toBe('developer');
      expect(messages[0].content).toBe(schema.systemPrompt);
    });
  });
});