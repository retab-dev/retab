import { generateSchemaDataId, generateSchemaId, loadJsonSchema } from '../src/utils/json_schema_utils.js';
import { zodToJsonSchema } from '../src/utils/zod_to_json_schema.js';
import { z } from 'zod';
import { companySchema, PersonZodSchema, PersonWithAddressZodSchema } from './fixtures.js';

describe('Utility Functions', () => {
  describe('JSON Schema Utils', () => {
    describe('loadJsonSchema', () => {
      it('should load JSON schema from object', () => {
        const result = loadJsonSchema(companySchema);
        expect(result).toEqual(companySchema);
      });

      it('should load JSON schema from string', () => {
        const schemaString = JSON.stringify(companySchema);
        const result = loadJsonSchema(schemaString);
        expect(result).toEqual(companySchema);
      });

      it('should handle nested schema objects', () => {
        const nestedSchema = {
          type: 'object',
          properties: {
            company: companySchema,
            metadata: {
              type: 'object',
              properties: {
                created_at: { type: 'string', format: 'date-time' },
                updated_at: { type: 'string', format: 'date-time' }
              }
            }
          }
        };

        const result = loadJsonSchema(nestedSchema);
        expect(result).toEqual(nestedSchema);
        expect(result.properties.company).toEqual(companySchema);
      });

      it('should handle schema with $defs', () => {
        const schemaWithDefs = {
          type: 'object',
          properties: {
            company: { $ref: '#/$defs/Company' }
          },
          $defs: {
            Company: companySchema
          }
        };

        const result = loadJsonSchema(schemaWithDefs);
        expect(result).toEqual(schemaWithDefs);
        expect(result.$defs.Company).toEqual(companySchema);
      });

      it('should throw error for invalid JSON string', () => {
        const invalidJson = '{ invalid json }';
        expect(() => loadJsonSchema(invalidJson)).toThrow();
      });
    });

    describe('generateSchemaDataId', () => {
      it('should generate consistent data ID for same schema', () => {
        const id1 = generateSchemaDataId(companySchema);
        const id2 = generateSchemaDataId(companySchema);
        
        expect(id1).toBe(id2);
        expect(typeof id1).toBe('string');
        expect(id1.length).toBeGreaterThan(0);
      });

      it('should generate different IDs for different schemas', () => {
        const schema1 = { type: 'object', properties: { name: { type: 'string' } } };
        const schema2 = { type: 'object', properties: { title: { type: 'string' } } };
        
        const id1 = generateSchemaDataId(schema1);
        const id2 = generateSchemaDataId(schema2);
        
        expect(id1).not.toBe(id2);
      });

      it('should ignore description fields in data ID generation', () => {
        const schema1 = {
          type: 'object',
          properties: {
            name: { type: 'string', description: 'Person name' }
          }
        };
        
        const schema2 = {
          type: 'object',
          properties: {
            name: { type: 'string', description: 'Different description' }
          }
        };
        
        const id1 = generateSchemaDataId(schema1);
        const id2 = generateSchemaDataId(schema2);
        
        expect(id1).toBe(id2);
      });

      it('should ignore X- custom fields in data ID generation', () => {
        const schema1 = {
          type: 'object',
          properties: {
            name: { 
              type: 'string',
              'X-FieldPrompt': 'Extract the name',
              'X-ReasoningPrompt': 'Look carefully'
            }
          }
        };
        
        const schema2 = {
          type: 'object',
          properties: {
            name: { 
              type: 'string',
              'X-FieldPrompt': 'Different prompt',
              'X-ReasoningPrompt': 'Different reasoning'
            }
          }
        };
        
        const id1 = generateSchemaDataId(schema1);
        const id2 = generateSchemaDataId(schema2);
        
        expect(id1).toBe(id2);
      });

      it('should handle empty schema', () => {
        const emptySchema = {};
        const id = generateSchemaDataId(emptySchema);
        
        expect(typeof id).toBe('string');
        expect(id.length).toBeGreaterThan(0);
      });
    });

    describe('generateSchemaId', () => {
      it('should generate consistent ID for same complete schema', () => {
        const id1 = generateSchemaId(companySchema);
        const id2 = generateSchemaId(companySchema);
        
        expect(id1).toBe(id2);
        expect(typeof id1).toBe('string');
        expect(id1.length).toBeGreaterThan(0);
      });

      it('should generate different IDs for different schemas', () => {
        const schema1 = { type: 'object', properties: { name: { type: 'string' } } };
        const schema2 = { type: 'object', properties: { title: { type: 'string' } } };
        
        const id1 = generateSchemaId(schema1);
        const id2 = generateSchemaId(schema2);
        
        expect(id1).not.toBe(id2);
      });

      it('should include description fields in complete ID generation', () => {
        const schema1 = {
          type: 'object',
          properties: {
            name: { type: 'string', description: 'Person name' }
          }
        };
        
        const schema2 = {
          type: 'object',
          properties: {
            name: { type: 'string', description: 'Different description' }
          }
        };
        
        const id1 = generateSchemaId(schema1);
        const id2 = generateSchemaId(schema2);
        
        expect(id1).not.toBe(id2);
      });

      it('should include X- custom fields in complete ID generation', () => {
        const schema1 = {
          type: 'object',
          properties: {
            name: { 
              type: 'string',
              'X-FieldPrompt': 'Extract the name'
            }
          }
        };
        
        const schema2 = {
          type: 'object',
          properties: {
            name: { 
              type: 'string',
              'X-FieldPrompt': 'Different prompt'
            }
          }
        };
        
        const id1 = generateSchemaId(schema1);
        const id2 = generateSchemaId(schema2);
        
        expect(id1).not.toBe(id2);
      });

      it('should handle complex nested schemas', () => {
        const complexSchema = {
          type: 'object',
          properties: {
            user: {
              type: 'object',
              properties: {
                profile: {
                  type: 'object',
                  properties: {
                    name: { type: 'string' },
                    contacts: {
                      type: 'array',
                      items: {
                        type: 'object',
                        properties: {
                          type: { type: 'string', enum: ['email', 'phone'] },
                          value: { type: 'string' }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        };
        
        const id = generateSchemaId(complexSchema);
        expect(typeof id).toBe('string');
        expect(id.length).toBeGreaterThan(0);
      });
    });
  });

  describe('Zod to JSON Schema Conversion', () => {
    describe('zodToJsonSchema', () => {
      it('should convert simple Zod schema to JSON schema', () => {
        const jsonSchema = zodToJsonSchema(PersonZodSchema);
        
        expect(jsonSchema).toBeDefined();
        expect(jsonSchema.type).toBe('object');
        expect(jsonSchema.properties).toBeDefined();
        expect(jsonSchema.properties.name).toBeDefined();
        expect(jsonSchema.properties.age).toBeDefined();
        expect(jsonSchema.properties.email).toBeDefined();
        expect(jsonSchema.properties.isActive).toBeDefined();
      });

      it('should handle Zod schema with nested objects', () => {
        const jsonSchema = zodToJsonSchema(PersonWithAddressZodSchema);
        
        expect(jsonSchema.type).toBe('object');
        expect(jsonSchema.properties.person).toBeDefined();
        expect(jsonSchema.properties.address).toBeDefined();
        expect(jsonSchema.properties.tags).toBeDefined();
        
        // Check nested person object
        expect(jsonSchema.properties.person.type).toBe('object');
        expect(jsonSchema.properties.person.properties).toBeDefined();
      });

      it('should handle Zod enums', () => {
        const StatusEnum = z.enum(['active', 'inactive', 'pending']);
        const SchemaWithEnum = z.object({
          status: StatusEnum
        });
        
        const jsonSchema = zodToJsonSchema(SchemaWithEnum);
        
        expect(jsonSchema.properties.status.enum).toEqual(['active', 'inactive', 'pending']);
      });

      it('should handle Zod arrays', () => {
        const ArraySchema = z.object({
          tags: z.array(z.string()),
          numbers: z.array(z.number())
        });
        
        const jsonSchema = zodToJsonSchema(ArraySchema);
        
        expect(jsonSchema.properties.tags.type).toBe('array');
        expect(jsonSchema.properties.tags.items.type).toBe('string');
        expect(jsonSchema.properties.numbers.type).toBe('array');
        expect(jsonSchema.properties.numbers.items.type).toBe('number');
      });

      it('should handle optional fields', () => {
        const OptionalSchema = z.object({
          required: z.string(),
          optional: z.string().optional()
        });
        
        const jsonSchema = zodToJsonSchema(OptionalSchema);
        
        expect(jsonSchema.required).toContain('required');
        expect(jsonSchema.required).not.toContain('optional');
      });

      it('should handle default values', () => {
        const DefaultSchema = z.object({
          name: z.string(),
          active: z.boolean().default(true),
          count: z.number().default(0)
        });
        
        const jsonSchema = zodToJsonSchema(DefaultSchema);
        
        // Note: zod-to-json-schema may not always preserve defaults
        // This is implementation dependent
        expect(jsonSchema.properties.active).toBeDefined();
        expect(jsonSchema.properties.count).toBeDefined();
      });

      it('should handle Zod descriptions', () => {
        const DescribedSchema = z.object({
          name: z.string().describe('The person\'s full name'),
          age: z.number().describe('Age in years')
        });
        
        const jsonSchema = zodToJsonSchema(DescribedSchema);
        
        expect(jsonSchema.properties.name.description).toBe('The person\'s full name');
        expect(jsonSchema.properties.age.description).toBe('Age in years');
      });

      it('should handle Zod unions', () => {
        const UnionSchema = z.object({
          value: z.union([z.string(), z.number()])
        });
        
        const jsonSchema = zodToJsonSchema(UnionSchema);
        
        expect(jsonSchema.properties.value.anyOf).toBeDefined();
        expect(jsonSchema.properties.value.anyOf).toHaveLength(2);
      });

      it('should handle Zod literals', () => {
        const LiteralSchema = z.object({
          type: z.literal('user'),
          status: z.literal('active')
        });
        
        const jsonSchema = zodToJsonSchema(LiteralSchema);
        
        expect(jsonSchema.properties.type.const).toBe('user');
        expect(jsonSchema.properties.status.const).toBe('active');
      });

      it('should handle complex nested structures', () => {
        const ComplexSchema = z.object({
          user: z.object({
            id: z.string().uuid(),
            profile: z.object({
              name: z.string(),
              email: z.string().email(),
              preferences: z.object({
                theme: z.enum(['light', 'dark']),
                notifications: z.boolean().default(true)
              })
            }).optional(),
            tags: z.array(z.string()).default([])
          }),
          metadata: z.record(z.string(), z.any()).optional()
        });
        
        const jsonSchema = zodToJsonSchema(ComplexSchema);
        
        expect(jsonSchema.type).toBe('object');
        expect(jsonSchema.properties.user.type).toBe('object');
        expect(jsonSchema.properties.user.properties.profile.type).toBe('object');
        expect(jsonSchema.properties.user.properties.profile.properties.preferences.type).toBe('object');
      });
    });
  });

  describe('Hash Generation', () => {
    it('should generate consistent hashes for same input', () => {
      // This would test the Blake2b hash function if we could access it directly
      // For now, we test it indirectly through the schema ID generation
      const id1 = generateSchemaDataId(companySchema);
      const id2 = generateSchemaDataId(companySchema);
      
      expect(id1).toBe(id2);
    });

    it('should generate different hashes for different inputs', () => {
      const schema1 = { type: 'object', properties: { a: { type: 'string' } } };
      const schema2 = { type: 'object', properties: { b: { type: 'string' } } };
      
      const id1 = generateSchemaDataId(schema1);
      const id2 = generateSchemaDataId(schema2);
      
      expect(id1).not.toBe(id2);
    });
  });

  describe('Edge Cases', () => {
    it('should handle null and undefined inputs gracefully', () => {
      expect(() => generateSchemaDataId(null as any)).toThrow();
      expect(() => generateSchemaDataId(undefined as any)).toThrow();
      expect(() => generateSchemaId(null as any)).toThrow();
      expect(() => generateSchemaId(undefined as any)).toThrow();
    });

    it('should handle very large schemas', () => {
      const largeSchema = {
        type: 'object',
        properties: Object.fromEntries(
          Array.from({ length: 100 }, (_, i) => [
            `field_${i}`,
            {
              type: 'object',
              properties: {
                id: { type: 'string' },
                name: { type: 'string' },
                data: {
                  type: 'array',
                  items: {
                    type: 'object',
                    properties: {
                      value: { type: 'number' },
                      label: { type: 'string' }
                    }
                  }
                }
              }
            }
          ])
        )
      };
      
      const id = generateSchemaDataId(largeSchema);
      expect(typeof id).toBe('string');
      expect(id.length).toBeGreaterThan(0);
    });

    it('should handle schemas with circular references', () => {
      const circularSchema = {
        type: 'object',
        properties: {
          name: { type: 'string' },
          children: {
            type: 'array',
            items: { $ref: '#' }
          }
        }
      };
      
      // Should not throw even with circular references
      const id = generateSchemaDataId(circularSchema);
      expect(typeof id).toBe('string');
      expect(id.length).toBeGreaterThan(0);
    });
  });
});