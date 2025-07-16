import {
  calculateCost,
  calculateAudioCost,
  calculateVideoProjectCost,
  CalculationConfig
} from '../src/utils/cost_calculation.js';
import { hashString } from '../src/utils/hash.js';
import { loadJsonSchema } from '../src/utils/json_schema_utils.js';
import { zodToJsonSchema } from '../src/utils/zod_to_json_schema.js';
import { z } from 'zod';

describe('Utility Functions Tests', () => {
  describe('Cost Calculation Utilities', () => {
    describe('calculateCost', () => {
      it('should calculate basic text cost correctly', () => {
        const config: CalculationConfig = {
          model: 'gpt-4o',
          inputTokens: 1000,
          outputTokens: 500
        };

        const cost = calculateCost(config);
        expect(typeof cost).toBe('number');
        expect(cost).toBeGreaterThan(0);
      });

      it('should handle different models', () => {
        const models = ['gpt-4o', 'gpt-4o-mini', 'claude-3-5-sonnet-latest'];

        models.forEach(model => {
          const config: CalculationConfig = {
            model,
            inputTokens: 1000,
            outputTokens: 500
          };

          const cost = calculateCost(config);
          expect(typeof cost).toBe('number');
          expect(cost).toBeGreaterThan(0);
        });
      });

      it('should handle zero tokens', () => {
        const config: CalculationConfig = {
          model: 'gpt-4o',
          inputTokens: 0,
          outputTokens: 0
        };

        const cost = calculateCost(config);
        expect(cost).toBe(0);
      });

      it('should handle large token counts', () => {
        const config: CalculationConfig = {
          model: 'gpt-4o',
          inputTokens: 1000000,
          outputTokens: 500000
        };

        const cost = calculateCost(config);
        expect(typeof cost).toBe('number');
        expect(cost).toBeGreaterThan(0);
      });

      it('should include image costs when provided', () => {
        const configWithImage: CalculationConfig = {
          model: 'gpt-4o',
          inputTokens: 1000,
          outputTokens: 500,
          imageCount: 2
        };

        const configWithoutImage: CalculationConfig = {
          model: 'gpt-4o',
          inputTokens: 1000,
          outputTokens: 500
        };

        const costWithImage = calculateCost(configWithImage);
        const costWithoutImage = calculateCost(configWithoutImage);

        expect(costWithImage).toBeGreaterThan(costWithoutImage);
      });

      it('should handle different image sizes', () => {
        const imageSizes = ['low', 'high'] as const;

        imageSizes.forEach(size => {
          const config: CalculationConfig = {
            model: 'gpt-4o',
            inputTokens: 1000,
            outputTokens: 500,
            imageCount: 1,
            imageSize: size
          };

          const cost = calculateCost(config);
          expect(typeof cost).toBe('number');
          expect(cost).toBeGreaterThan(0);
        });
      });
    });

    describe('calculateAudioCost', () => {
      it('should calculate audio cost correctly', () => {
        const cost = calculateAudioCost('whisper-1', 60); // 60 seconds
        expect(typeof cost).toBe('number');
        expect(cost).toBeGreaterThan(0);
      });

      it('should handle zero duration', () => {
        const cost = calculateAudioCost('whisper-1', 0);
        expect(cost).toBe(0);
      });

      it('should handle fractional duration', () => {
        const cost = calculateAudioCost('whisper-1', 30.5);
        expect(typeof cost).toBe('number');
        expect(cost).toBeGreaterThan(0);
      });

      it('should handle large durations', () => {
        const cost = calculateAudioCost('whisper-1', 3600); // 1 hour
        expect(typeof cost).toBe('number');
        expect(cost).toBeGreaterThan(0);
      });
    });

    describe('calculateVideoProjectCost', () => {
      it('should calculate video evaluation cost correctly', () => {
        const cost = calculateVideoProjectCost(120); // 2 minutes
        expect(typeof cost).toBe('number');
        expect(cost).toBeGreaterThan(0);
      });

      it('should handle zero duration', () => {
        const cost = calculateVideoProjectCost(0);
        expect(cost).toBe(0);
      });

      it('should handle fractional duration', () => {
        const cost = calculateVideoProjectCost(90.5);
        expect(typeof cost).toBe('number');
        expect(cost).toBeGreaterThan(0);
      });

      it('should scale with duration', () => {
        const shortCost = calculateVideoProjectCost(60);
        const longCost = calculateVideoProjectCost(120);

        expect(longCost).toBeGreaterThan(shortCost);
      });
    });
  });

  describe('Hash Utilities', () => {
    describe('hashString', () => {
      it('should generate consistent hashes for same input', () => {
        const input = 'test string';
        const hash1 = hashString(input);
        const hash2 = hashString(input);

        expect(hash1).toBe(hash2);
        expect(typeof hash1).toBe('string');
        expect(hash1.length).toBeGreaterThan(0);
      });

      it('should generate different hashes for different inputs', () => {
        const hash1 = hashString('input1');
        const hash2 = hashString('input2');

        expect(hash1).not.toBe(hash2);
      });

      it('should handle empty string', () => {
        const hash = hashString('');
        expect(typeof hash).toBe('string');
        expect(hash.length).toBeGreaterThan(0);
      });

      it('should handle special characters', () => {
        const specialStrings = [
          'unicode: 🚀 ñ ü',
          'symbols: !@#$%^&*()',
          'newlines\nand\ttabs',
          'very long string'.repeat(100)
        ];

        specialStrings.forEach(str => {
          const hash = hashString(str);
          expect(typeof hash).toBe('string');
          expect(hash.length).toBeGreaterThan(0);
        });
      });

      it('should be case sensitive', () => {
        const hash1 = hashString('Test');
        const hash2 = hashString('test');

        expect(hash1).not.toBe(hash2);
      });

      it('should handle very long strings', () => {
        const longString = 'a'.repeat(10000);
        const hash = hashString(longString);

        expect(typeof hash).toBe('string');
        expect(hash.length).toBeGreaterThan(0);
      });
    });
  });

  describe('JSON Schema Utilities', () => {
    describe('loadJsonSchema', () => {
      it('should load object schema correctly', () => {
        const schema = {
          type: 'object',
          properties: {
            name: { type: 'string' },
            age: { type: 'number' }
          }
        };

        const loaded = loadJsonSchema(schema);
        expect(loaded).toEqual(schema);
      });

      it('should handle string input (JSON)', () => {
        const schema = {
          type: 'object',
          properties: {
            name: { type: 'string' }
          }
        };

        const jsonString = JSON.stringify(schema);
        const loaded = loadJsonSchema(jsonString);

        expect(loaded).toEqual(schema);
      });

      it('should handle invalid JSON string', () => {
        expect(() => {
          loadJsonSchema('invalid json string');
        }).toThrow();
      });

      it('should handle complex nested schemas', () => {
        const complexSchema = {
          type: 'object',
          properties: {
            user: {
              type: 'object',
              properties: {
                name: { type: 'string' },
                addresses: {
                  type: 'array',
                  items: {
                    type: 'object',
                    properties: {
                      street: { type: 'string' },
                      city: { type: 'string' }
                    }
                  }
                }
              }
            }
          },
          required: ['user']
        };

        const loaded = loadJsonSchema(complexSchema);
        expect(loaded).toEqual(complexSchema);
      });

      it('should preserve schema properties', () => {
        const schema = {
          type: 'object',
          properties: {
            name: {
              type: 'string',
              description: 'User name',
              minLength: 1,
              maxLength: 100
            }
          },
          required: ['name'],
          additionalProperties: false
        };

        const loaded = loadJsonSchema(schema);
        expect(loaded).toEqual(schema);
        expect(loaded.properties.name.description).toBe('User name');
        expect(loaded.required).toEqual(['name']);
        expect(loaded.additionalProperties).toBe(false);
      });
    });
  });

  describe('Zod to JSON Schema Conversion', () => {
    describe('zodToJsonSchema', () => {
      it('should convert basic Zod schema to JSON schema', () => {
        const zodSchema = z.object({
          name: z.string(),
          age: z.number()
        });

        const jsonSchema = zodToJsonSchema(zodSchema);

        expect(jsonSchema.type).toBe('object');
        expect(jsonSchema.properties).toBeDefined();
        expect(jsonSchema.properties.name).toEqual({ type: 'string' });
        expect(jsonSchema.properties.age).toEqual({ type: 'number' });
      });

      it('should handle string validations', () => {
        const zodSchema = z.object({
          email: z.string().email(),
          name: z.string().min(1).max(100)
        });

        const jsonSchema = zodToJsonSchema(zodSchema);

        expect(jsonSchema.properties.email).toBeDefined();
        expect(jsonSchema.properties.name).toBeDefined();
        expect(jsonSchema.properties.name.minLength).toBe(1);
        expect(jsonSchema.properties.name.maxLength).toBe(100);
      });

      it('should handle number validations', () => {
        const zodSchema = z.object({
          age: z.number().min(0).max(120),
          score: z.number().positive()
        });

        const jsonSchema = zodToJsonSchema(zodSchema);

        expect(jsonSchema.properties.age.minimum).toBe(0);
        expect(jsonSchema.properties.age.maximum).toBe(120);
        expect(jsonSchema.properties.score.minimum).toBeGreaterThan(0);
      });

      it('should handle array schemas', () => {
        const zodSchema = z.object({
          tags: z.array(z.string()),
          scores: z.array(z.number()).min(1).max(10)
        });

        const jsonSchema = zodToJsonSchema(zodSchema);

        expect(jsonSchema.properties.tags.type).toBe('array');
        expect(jsonSchema.properties.tags.items.type).toBe('string');
        expect(jsonSchema.properties.scores.type).toBe('array');
        expect(jsonSchema.properties.scores.items.type).toBe('number');
        expect(jsonSchema.properties.scores.minItems).toBe(1);
        expect(jsonSchema.properties.scores.maxItems).toBe(10);
      });

      it('should handle nested objects', () => {
        const zodSchema = z.object({
          user: z.object({
            name: z.string(),
            profile: z.object({
              bio: z.string().optional(),
              avatar: z.string().url()
            })
          })
        });

        const jsonSchema = zodToJsonSchema(zodSchema);

        expect(jsonSchema.properties.user.type).toBe('object');
        expect(jsonSchema.properties.user.properties.name.type).toBe('string');
        expect(jsonSchema.properties.user.properties.profile.type).toBe('object');
        expect(jsonSchema.properties.user.properties.profile.properties.bio.type).toBe('string');
      });

      it('should handle optional fields', () => {
        const zodSchema = z.object({
          required: z.string(),
          optional: z.string().optional()
        });

        const jsonSchema = zodToJsonSchema(zodSchema);

        expect(jsonSchema.required).toContain('required');
        expect(jsonSchema.required).not.toContain('optional');
      });

      it('should handle enums', () => {
        const zodSchema = z.object({
          status: z.enum(['active', 'inactive', 'pending']),
          priority: z.union([z.literal('low'), z.literal('medium'), z.literal('high')])
        });

        const jsonSchema = zodToJsonSchema(zodSchema);

        expect(jsonSchema.properties.status.enum).toEqual(['active', 'inactive', 'pending']);
      });

      it('should handle complex validation rules', () => {
        const zodSchema = z.object({
          password: z.string().min(8).regex(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/),
          confirmPassword: z.string()
        }).refine(data => data.password === data.confirmPassword, {
          message: "Passwords don't match"
        });

        const jsonSchema = zodToJsonSchema(zodSchema);

        expect(jsonSchema.properties.password.minLength).toBe(8);
        expect(jsonSchema.properties.password.pattern).toBeDefined();
      });

      it('should handle different Zod types', () => {
        const zodSchema = z.object({
          string: z.string(),
          number: z.number(),
          boolean: z.boolean(),
          date: z.date(),
          uuid: z.string().uuid(),
          url: z.string().url(),
          email: z.string().email()
        });

        const jsonSchema = zodToJsonSchema(zodSchema);

        expect(jsonSchema.properties.string.type).toBe('string');
        expect(jsonSchema.properties.number.type).toBe('number');
        expect(jsonSchema.properties.boolean.type).toBe('boolean');
        expect(jsonSchema.properties.uuid.format).toBe('uuid');
        expect(jsonSchema.properties.url.format).toBe('uri');
        expect(jsonSchema.properties.email.format).toBe('email');
      });
    });
  });

  describe('Utility Function Integration', () => {
    it('should work together in realistic scenarios', () => {
      // Create a Zod schema
      const zodSchema = z.object({
        name: z.string(),
        email: z.string().email(),
        age: z.number().min(0)
      });

      // Convert to JSON schema
      const jsonSchema = zodToJsonSchema(zodSchema);

      // Load the JSON schema
      const loadedSchema = loadJsonSchema(jsonSchema);

      // Hash the schema for caching
      const schemaHash = hashString(JSON.stringify(loadedSchema));

      // Calculate cost for processing this schema
      const cost = calculateCost({
        model: 'gpt-4o',
        inputTokens: 1000,
        outputTokens: 500
      });

      expect(loadedSchema).toEqual(jsonSchema);
      expect(typeof schemaHash).toBe('string');
      expect(typeof cost).toBe('number');
      expect(cost).toBeGreaterThan(0);
    });

    it('should handle schema hashing for deduplication', () => {
      const schema1 = { type: 'object', properties: { name: { type: 'string' } } };
      const schema2 = { type: 'object', properties: { name: { type: 'string' } } };
      const schema3 = { type: 'object', properties: { age: { type: 'number' } } };

      const hash1 = hashString(JSON.stringify(loadJsonSchema(schema1)));
      const hash2 = hashString(JSON.stringify(loadJsonSchema(schema2)));
      const hash3 = hashString(JSON.stringify(loadJsonSchema(schema3)));

      expect(hash1).toBe(hash2); // Same schemas should have same hash
      expect(hash1).not.toBe(hash3); // Different schemas should have different hashes
    });

    it('should calculate costs for different scenarios', () => {
      const scenarios = [
        { model: 'gpt-4o-mini', inputTokens: 100, outputTokens: 50 },
        { model: 'gpt-4o', inputTokens: 1000, outputTokens: 500, imageCount: 1 },
        { model: 'claude-3-5-sonnet-latest', inputTokens: 2000, outputTokens: 1000 }
      ];

      scenarios.forEach(scenario => {
        const cost = calculateCost(scenario);
        expect(typeof cost).toBe('number');
        expect(cost).toBeGreaterThan(0);
      });
    });
  });

  describe('Edge Cases and Error Handling', () => {
    it('should handle edge cases in cost calculation', () => {
      const edgeCases = [
        { model: 'unknown-model', inputTokens: 100, outputTokens: 50 },
        { model: 'gpt-4o', inputTokens: -1, outputTokens: 50 },
        { model: 'gpt-4o', inputTokens: 100, outputTokens: -1 },
        { model: 'gpt-4o', inputTokens: 100, outputTokens: 50, imageCount: -1 }
      ];

      edgeCases.forEach(scenario => {
        expect(() => calculateCost(scenario)).not.toThrow();
      });
    });

    it('should handle invalid JSON schema inputs gracefully', () => {
      const invalidInputs = [
        null,
        undefined,
        123,
        [],
        { invalid: 'schema without type' }
      ];

      invalidInputs.forEach(input => {
        if (input === null || input === undefined) {
          expect(() => loadJsonSchema(input as any)).toThrow();
        } else {
          // For other invalid inputs, should handle gracefully or throw meaningful error
          expect(() => loadJsonSchema(input as any)).not.toThrow();
        }
      });
    });

    it('should handle complex Zod schemas without errors', () => {
      const complexSchema = z.object({
        metadata: z.record(z.string(), z.any()),
        tags: z.array(z.string()).optional(),
        nested: z.object({
          deep: z.object({
            value: z.union([z.string(), z.number(), z.boolean()])
          })
        }).optional()
      });

      expect(() => zodToJsonSchema(complexSchema)).not.toThrow();
    });
  });
});