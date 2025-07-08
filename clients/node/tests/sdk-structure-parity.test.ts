/**
 * SDK Structure Parity Tests
 * 
 * Tests that compare the structure and basic functionality between Python and Node.js SDKs
 * without actually running Python code to avoid timeout issues.
 */

import { Retab, AsyncRetab } from '../src/index.js';
import { TEST_API_KEY } from './fixtures.js';

describe('SDK Structure Parity Tests', () => {
  let syncClient: Retab;
  let asyncClient: AsyncRetab;

  beforeAll(() => {
    syncClient = new Retab({ 
      apiKey: TEST_API_KEY,
      openaiApiKey: 'test-openai-key'
    });
    asyncClient = new AsyncRetab({ 
      apiKey: TEST_API_KEY,
      openaiApiKey: 'test-openai-key'
    });
  });

  describe('Client Configuration Structure', () => {
    it('should have the same configuration properties as Python SDK', () => {
      // Based on Python SDK analysis, these properties should be accessible
      expect((syncClient as any).apiKey).toBe(TEST_API_KEY);
      expect((syncClient as any).baseUrl).toContain('api.retab.com');
      expect(typeof (syncClient as any).timeout).toBe('number');
      expect(typeof (syncClient as any).maxRetries).toBe('number');
      
      // Should have OpenAI API key in headers
      expect(syncClient.getHeaders()['OpenAI-Api-Key']).toBe('test-openai-key');
    });

    it('should handle missing API key the same as Python SDK', () => {
      // Both SDKs should throw an error when no API key is provided
      expect(() => new Retab()).toThrow();
      expect(() => new AsyncRetab()).toThrow();
      
      // Error message should mention API key
      try {
        new Retab();
      } catch (error: any) {
        expect(error.message.toLowerCase()).toContain('api key');
      }
    });

    it('should handle custom base URL configuration', () => {
      const customUrl = 'https://custom.api.url';
      const customClient = new Retab({ 
        apiKey: TEST_API_KEY,
        baseUrl: customUrl
      });
      
      expect((customClient as any).baseUrl).toBe(customUrl);
    });
  });

  describe('Resource Structure Parity', () => {
    it('should have the same resources as Python SDK', () => {
      // Core resources that should exist in both SDKs
      expect(syncClient.schemas).toBeDefined();
      expect(syncClient.documents).toBeDefined();
      expect(syncClient.files).toBeDefined();
      expect(syncClient.fineTuning).toBeDefined();
      expect(syncClient.consensus).toBeDefined();
      expect(syncClient.processors).toBeDefined();
      
      // Async client should have the same resources
      expect(asyncClient.schemas).toBeDefined();
      expect(asyncClient.documents).toBeDefined();
      expect(asyncClient.files).toBeDefined();
      expect(asyncClient.fineTuning).toBeDefined();
      expect(asyncClient.consensus).toBeDefined();
      expect(asyncClient.processors).toBeDefined();
    });

    it('should have schemas resource with correct methods', () => {
      const schemasSync = syncClient.schemas;
      const schemasAsync = asyncClient.schemas;
      
      // Both should have load method
      expect(typeof schemasSync.load).toBe('function');
      expect(typeof schemasAsync.load).toBe('function');
      
      // Both should have generate method
      expect(typeof schemasSync.generate).toBe('function');
      expect(typeof schemasAsync.generate).toBe('function');
      
      // Both should have evaluate method
      expect(typeof schemasSync.evaluate).toBe('function');
      expect(typeof schemasAsync.evaluate).toBe('function');
      
      // Both should have enhance method
      expect(typeof schemasSync.enhance).toBe('function');
      expect(typeof schemasAsync.enhance).toBe('function');
    });

    it('should have documents resource with extractions', () => {
      expect(syncClient.documents.extractions).toBeDefined();
      expect(asyncClient.documents.extractions).toBeDefined();
      
      // Should have extraction methods
      expect(typeof syncClient.documents.extractions.extract).toBe('function');
      expect(typeof asyncClient.documents.extractions.extract).toBe('function');
    });

    it('should have fine-tuning resource structure', () => {
      expect(syncClient.fineTuning.jobs).toBeDefined();
      expect(asyncClient.fineTuning.jobs).toBeDefined();
      
      // Should have job management methods
      expect(typeof syncClient.fineTuning.jobs.create).toBe('function');
      expect(typeof syncClient.fineTuning.jobs.retrieve).toBe('function');
      expect(typeof syncClient.fineTuning.jobs.list).toBe('function');
      expect(typeof syncClient.fineTuning.jobs.cancel).toBe('function');
      expect(typeof syncClient.fineTuning.jobs.listEvents).toBe('function');
    });

    it('should have processor automations structure', () => {
      expect(syncClient.processors.automations).toBeDefined();
      expect(asyncClient.processors.automations).toBeDefined();
      
      // Should have automation sub-resources
      expect(syncClient.processors.automations.links).toBeDefined();
      expect(syncClient.processors.automations.endpoints).toBeDefined();
      expect(syncClient.processors.automations.mailboxes).toBeDefined();
      expect(syncClient.processors.automations.tests).toBeDefined();
      expect(syncClient.processors.automations.logs).toBeDefined();
    });
  });

  describe('Schema Operations Structure', () => {
    it('should create schemas with the same interface as Python SDK', () => {
      const testSchema = {
        type: 'object',
        properties: {
          name: { type: 'string' },
          age: { type: 'number' }
        },
        required: ['name', 'age']
      };
      
      // Should be able to load schema from JSON
      const schema = syncClient.schemas.load(testSchema);
      expect(schema.json_schema).toEqual(testSchema);
      expect(typeof schema.save).toBe('function');
      expect(schema.constructor.name).toBe('Schema');
    });

    it('should handle pydantic model simulation', () => {
      // Simulate how pydantic models would be handled in Node.js
      const mockPydanticModel = {
        model_json_schema: () => ({
          type: 'object',
          properties: {
            name: { type: 'string' },
            age: { type: 'integer' }
          },
          required: ['name', 'age']
        })
      };
      
      const schema = syncClient.schemas.load(null, mockPydanticModel);
      expect(schema.json_schema).toBeDefined();
      expect(schema.json_schema.properties).toHaveProperty('name');
      expect(schema.json_schema.properties).toHaveProperty('age');
    });

    it('should validate required parameters like Python SDK', () => {
      // Should throw when no parameters provided
      expect(() => syncClient.schemas.load()).toThrow();
      
      const error = (() => {
        try {
          syncClient.schemas.load();
        } catch (e: any) {
          return e.message;
        }
      })();
      
      expect(error.toLowerCase()).toContain('provided');
    });
  });

  describe('Error Handling Parity', () => {
    it('should handle invalid model names consistently', () => {
      const invalidModel = 'invalid-model-name';
      
      // Should throw error for invalid model in schema generation
      expect(() => {
        (syncClient.schemas as any).mixin.prepareGenerate(
          ['test.pdf'],
          null,
          invalidModel
        );
      }).toThrow();
    });

    it('should provide similar error messages', () => {
      // Test missing API key error message
      try {
        new Retab();
      } catch (error: any) {
        expect(error.message).toContain('API key');
        expect(error.message).toContain('retab.com');
      }
      
      // Test missing parameters error message
      try {
        syncClient.schemas.load();
      } catch (error: any) {
        expect(error.message).toContain('provided');
      }
    });
  });

  describe('API Request Structure', () => {
    it('should generate requests with same structure as Python SDK', () => {
      // Test schema generation request structure
      const prepared = (syncClient.schemas as any).mixin.prepareGenerate(
        ['test.pdf'],
        'Test instructions',
        'gpt-4o-2024-11-20',
        0.5
      );
      
      expect(prepared.method).toBe('POST');
      expect(prepared.url).toBe('/v1/schemas/generate');
      expect(prepared.data).toBeDefined();
      expect(prepared.data.documents).toBeDefined();
      expect(prepared.data.model).toBe('gpt-4o-2024-11-20');
      expect(prepared.data.temperature).toBe(0.5);
      expect(prepared.data.instructions).toBe('Test instructions');
    });

    it('should prepare document extraction with same structure', () => {
      const prepared = (syncClient.documents.extractions as any).mixin.prepareExtraction(
        { type: 'object', properties: { name: { type: 'string' } } },
        null,
        ['test.pdf'],
        undefined,
        undefined,
        'gpt-4o-mini',
        undefined,
        'native'
      );
      
      expect(prepared.method).toBe('POST');
      expect(prepared.url).toBe('/v1/documents/extractions');
      expect(prepared.data).toBeDefined();
      expect(prepared.data.json_schema).toBeDefined();
      expect(prepared.data.documents).toBeDefined();
      expect(prepared.data.model).toBe('gpt-4o-mini');
      expect(prepared.data.modality).toBe('native');
    });
  });

  describe('Client Type Consistency', () => {
    it('should have consistent naming patterns', () => {
      // Check client class names
      expect(syncClient.constructor.name).toBe('Retab');
      expect(asyncClient.constructor.name).toBe('AsyncRetab');
      
      // Check resource class names
      expect(syncClient.schemas.constructor.name).toBe('Schemas');
      expect(asyncClient.schemas.constructor.name).toBe('AsyncSchemas');
      
      expect(syncClient.documents.constructor.name).toBe('Documents');
      expect(asyncClient.documents.constructor.name).toBe('AsyncDocuments');
    });

    it('should have async functions consistently', () => {
      // All main methods should be async functions
      expect(syncClient.schemas.generate.constructor.name).toBe('AsyncFunction');
      expect(asyncClient.schemas.generate.constructor.name).toBe('AsyncFunction');
      
      expect(syncClient.documents.extractions.extract.constructor.name).toBe('AsyncFunction');
      expect(asyncClient.documents.extractions.extract.constructor.name).toBe('AsyncFunction');
    });
  });

  describe('Parameter Validation Consistency', () => {
    it('should handle document parameters the same way', () => {
      // Should accept string array
      expect(() => {
        (syncClient.schemas as any).mixin.prepareGenerate(['test.pdf']);
      }).not.toThrow();
      
      // Should handle invalid parameter types appropriately
      const testCases = [
        { input: ['test.pdf'], shouldSucceed: true },
        { input: 'test.pdf', shouldSucceed: false }, // Should be array
      ];
      
      testCases.forEach(({ input, shouldSucceed }) => {
        if (shouldSucceed) {
          expect(() => {
            (syncClient.schemas as any).mixin.prepareGenerate(input);
          }).not.toThrow();
        } else {
          expect(() => {
            (syncClient.schemas as any).mixin.prepareGenerate(input);
          }).toThrow();
        }
      });
    });
  });

  describe('OpenAI Integration Consistency', () => {
    it('should handle OpenAI API key the same way', () => {
      // Client with OpenAI key should have it in headers
      expect(syncClient.getHeaders()['OpenAI-Api-Key']).toBe('test-openai-key');
      expect(asyncClient.getHeaders()['OpenAI-Api-Key']).toBe('test-openai-key');
      
      // Client without OpenAI key should not have it
      const clientWithoutOpenAI = new Retab({ apiKey: TEST_API_KEY });
      expect(clientWithoutOpenAI.getHeaders()['OpenAI-Api-Key']).toBeUndefined();
    });

    it('should validate OpenAI models consistently', async () => {
      // Valid OpenAI models should be accepted for fine-tuning
      const validModels = [
        'gpt-4o-mini-2024-07-18',
        'gpt-4o-2024-08-06',
        'gpt-3.5-turbo-0125'
      ];
      
      validModels.forEach(model => {
        expect(() => {
          syncClient.fineTuning.jobs.create({
            training_file: 'file-abc123',
            model
          });
        }).not.toThrow(); // Should not throw validation error
      });
      
      // Invalid models should be rejected
      await expect(
        syncClient.fineTuning.jobs.create({
          training_file: 'file-abc123',
          model: 'claude-3-sonnet'
        })
      ).rejects.toThrow();
    });
  });
});