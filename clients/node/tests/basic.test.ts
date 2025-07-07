import { Retab, AsyncRetab, Schema } from '../src/index.js';

describe('Retab SDK', () => {
  describe('Client Initialization', () => {
    it('should create sync client with API key', () => {
      const client = new Retab({ apiKey: 'test-api-key' });
      expect(client).toBeInstanceOf(Retab);
    });

    it('should create async client with API key', () => {
      const client = new AsyncRetab({ apiKey: 'test-api-key' });
      expect(client).toBeInstanceOf(AsyncRetab);
    });

    it('should throw error when no API key provided', () => {
      // Clear env var for test
      const originalApiKey = process.env.RETAB_API_KEY;
      delete process.env.RETAB_API_KEY;

      expect(() => new Retab()).toThrow('No API key provided');
      
      // Restore env var
      if (originalApiKey) {
        process.env.RETAB_API_KEY = originalApiKey;
      }
    });
  });

  describe('Schema', () => {
    it('should create schema from json_schema', () => {
      const jsonSchema = {
        type: 'object',
        properties: {
          name: { type: 'string' }
        }
      };
      
      const schema = new Schema({ json_schema: jsonSchema });
      expect(schema.json_schema).toEqual(jsonSchema);
      expect(schema.object).toBe('schema');
    });

    it('should throw error when neither json_schema nor pydanticModel provided', () => {
      expect(() => new Schema({})).toThrow('Must provide either json_schema or pydanticModel');
    });

    it('should throw error when both json_schema and pydanticModel provided', () => {
      expect(() => new Schema({ 
        json_schema: {}, 
        pydanticModel: {} 
      })).toThrow('Cannot provide both json_schema and pydanticModel');
    });
  });

  describe('Resources', () => {
    let client: Retab;

    beforeEach(() => {
      client = new Retab({ apiKey: 'test-api-key' });
    });

    it('should have all expected resources', () => {
      expect(client.schemas).toBeDefined();
      expect(client.documents).toBeDefined();
      expect(client.files).toBeDefined();
      expect(client.fineTuning).toBeDefined();
      expect(client.models).toBeDefined();
      expect(client.processors).toBeDefined();
      expect(client.secrets).toBeDefined();
      expect(client.usage).toBeDefined();
      expect(client.consensus).toBeDefined();
      expect(client.evaluations).toBeDefined();
    });

    it('should be able to load schema', () => {
      const jsonSchema = {
        type: 'object',
        properties: {
          name: { type: 'string' }
        }
      };
      
      const schema = client.schemas.load({ jsonSchema });
      expect(schema).toBeInstanceOf(Schema);
      expect(schema.json_schema).toEqual(jsonSchema);
    });
  });
});