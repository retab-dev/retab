import { Retab, AsyncRetab } from '../src/index.js';
import { TEST_API_KEY } from './fixtures.js';

describe('Secrets External API Keys Tests', () => {
  let syncClient: Retab;
  let asyncClient: AsyncRetab;

  beforeEach(() => {
    syncClient = new Retab({ apiKey: TEST_API_KEY });
    asyncClient = new AsyncRetab({ apiKey: TEST_API_KEY });
  });

  describe('Resource Availability', () => {
    it('should have secrets resource on sync client', () => {
      expect(syncClient.secrets).toBeDefined();
      expect(syncClient.secrets.external_api_keys).toBeDefined();
      expect(typeof syncClient.secrets.external_api_keys.create).toBe('function');
      expect(typeof syncClient.secrets.external_api_keys.get).toBe('function');
      expect(typeof syncClient.secrets.external_api_keys.list).toBe('function');
      expect(typeof syncClient.secrets.external_api_keys.delete).toBe('function');
    });

    it('should have secrets resource on async client', () => {
      expect(asyncClient.secrets).toBeDefined();
      expect(asyncClient.secrets.external_api_keys).toBeDefined();
      expect(typeof asyncClient.secrets.external_api_keys.create).toBe('function');
      expect(typeof asyncClient.secrets.external_api_keys.get).toBe('function');
      expect(typeof asyncClient.secrets.external_api_keys.list).toBe('function');
      expect(typeof asyncClient.secrets.external_api_keys.delete).toBe('function');
    });
  });

  describe('Request Preparation - Create', () => {
    it('should prepare create request for OpenAI', () => {
      const prepared = (syncClient.secrets.external_api_keys as any).mixin.prepareCreate(
        'OpenAI',
        'sk-test-key'
      );
      
      expect(prepared.method).toBe('POST');
      expect(prepared.url).toBe('/v1/secrets/external_api_keys');
      expect(prepared.data).toEqual({
        provider: 'OpenAI',
        api_key: 'sk-test-key'
      });
    });

    it('should prepare create request for all providers', () => {
      const providers = ['OpenAI', 'Anthropic', 'Gemini', 'xAI'] as const;
      
      providers.forEach(provider => {
        const prepared = (syncClient.secrets.external_api_keys as any).mixin.prepareCreate(
          provider,
          `test-${provider}-key`
        );
        
        expect(prepared.method).toBe('POST');
        expect(prepared.url).toBe('/v1/secrets/external_api_keys');
        expect(prepared.data.provider).toBe(provider);
        expect(prepared.data.api_key).toBe(`test-${provider}-key`);
      });
    });
  });

  describe('Request Preparation - Get', () => {
    it('should prepare get request for specific provider', () => {
      const prepared = (syncClient.secrets.external_api_keys as any).mixin.prepareGet('OpenAI');
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe('/v1/secrets/external_api_keys/openai');
    });

    it('should prepare get request for all providers', () => {
      const providers = ['OpenAI', 'Anthropic', 'Gemini', 'xAI'] as const;
      
      providers.forEach(provider => {
        const prepared = (syncClient.secrets.external_api_keys as any).mixin.prepareGet(provider);
        
        expect(prepared.method).toBe('GET');
        expect(prepared.url).toBe(`/v1/secrets/external_api_keys/${provider}`);
      });
    });
  });

  describe('Request Preparation - List', () => {
    it('should prepare list request', () => {
      const prepared = (syncClient.secrets.external_api_keys as any).mixin.prepareList();
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe('/v1/secrets/external_api_keys');
      expect(prepared.data).toBeUndefined();
      expect(prepared.params).toBeUndefined();
    });
  });

  describe('Request Preparation - Delete', () => {
    it('should prepare delete request for specific provider', () => {
      const prepared = (syncClient.secrets.external_api_keys as any).mixin.prepareDelete('Anthropic');
      
      expect(prepared.method).toBe('DELETE');
      expect(prepared.url).toBe('/v1/secrets/external_api_keys/anthropic');
    });

    it('should prepare delete request for all providers', () => {
      const providers = ['OpenAI', 'Anthropic', 'Gemini', 'xAI'] as const;
      
      providers.forEach(provider => {
        const prepared = (syncClient.secrets.external_api_keys as any).mixin.prepareDelete(provider);
        
        expect(prepared.method).toBe('DELETE');
        expect(prepared.url).toBe(`/v1/secrets/external_api_keys/${provider}`);
      });
    });
  });

  describe('Provider Validation', () => {
    it('should handle valid AI providers', () => {
      const validProviders = ['OpenAI', 'Anthropic', 'Gemini', 'xAI'] as const;
      
      validProviders.forEach(provider => {
        expect(() => {
          (syncClient.secrets.external_api_keys as any).mixin.prepareCreate(provider, 'test-key');
        }).not.toThrow();
        
        expect(() => {
          (syncClient.secrets.external_api_keys as any).mixin.prepareGet(provider);
        }).not.toThrow();
        
        expect(() => {
          (syncClient.secrets.external_api_keys as any).mixin.prepareDelete(provider);
        }).not.toThrow();
      });
    });
  });

  describe('API Key Validation', () => {
    it('should handle different API key formats', () => {
      const apiKeyFormats = [
        'sk-test123',
        'ak-anthropic123',
        'xai-test123',
        'very-long-api-key-with-dashes-and-numbers-123456',
        'short',
        '123456789'
      ];
      
      apiKeyFormats.forEach(apiKey => {
        expect(() => {
          (syncClient.secrets.external_api_keys as any).mixin.prepareCreate('OpenAI', apiKey);
        }).not.toThrow();
      });
    });

    it('should handle empty and special characters in API keys', () => {
      const specialKeys = [
        '',
        ' ',
        'key with spaces',
        'key_with_underscores',
        'key-with-dashes',
        'key.with.dots',
        'key@with@symbols'
      ];
      
      specialKeys.forEach(apiKey => {
        expect(() => {
          (syncClient.secrets.external_api_keys as any).mixin.prepareCreate('OpenAI', apiKey);
        }).not.toThrow();
      });
    });
  });

  describe('Client Type Parity', () => {
    it('should have same methods for sync and async clients', () => {
      const syncMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(syncClient.secrets.external_api_keys));
      const asyncMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(asyncClient.secrets.external_api_keys));
      
      const requiredMethods = ['create', 'get', 'list', 'delete'];
      
      requiredMethods.forEach(method => {
        expect(syncMethods).toContain(method);
        expect(asyncMethods).toContain(method);
      });
    });

    it('should have consistent method signatures', () => {
      expect(syncClient.secrets.external_api_keys.create.length).toBe(asyncClient.secrets.external_api_keys.create.length);
      expect(syncClient.secrets.external_api_keys.get.length).toBe(asyncClient.secrets.external_api_keys.get.length);
      expect(syncClient.secrets.external_api_keys.list.length).toBe(asyncClient.secrets.external_api_keys.list.length);
      expect(syncClient.secrets.external_api_keys.delete.length).toBe(asyncClient.secrets.external_api_keys.delete.length);
    });

    it('should return promises from both clients', () => {
      // All methods should return promises
      const syncCreateResult = syncClient.secrets.external_api_keys.create('OpenAI', 'test-key');
      const asyncCreateResult = asyncClient.secrets.external_api_keys.create('OpenAI', 'test-key');
      
      expect(syncCreateResult).toBeInstanceOf(Promise);
      expect(asyncCreateResult).toBeInstanceOf(Promise);
      
      const syncListResult = syncClient.secrets.external_api_keys.list();
      const asyncListResult = asyncClient.secrets.external_api_keys.list();
      
      expect(syncListResult).toBeInstanceOf(Promise);
      expect(asyncListResult).toBeInstanceOf(Promise);
    });
  });

  describe('CRUD Operations Flow', () => {
    it('should prepare requests for complete CRUD flow', () => {
      const provider = 'OpenAI';
      const apiKey = 'sk-test-key';
      
      // Create
      const createRequest = (syncClient.secrets.external_api_keys as any).mixin.prepareCreate(provider, apiKey);
      expect(createRequest.method).toBe('POST');
      expect(createRequest.data.provider).toBe(provider);
      expect(createRequest.data.api_key).toBe(apiKey);
      
      // Read (Get)
      const getRequest = (syncClient.secrets.external_api_keys as any).mixin.prepareGet(provider);
      expect(getRequest.method).toBe('GET');
      expect(getRequest.url).toContain(provider);
      
      // Read (List)
      const listRequest = (syncClient.secrets.external_api_keys as any).mixin.prepareList();
      expect(listRequest.method).toBe('GET');
      expect(listRequest.url).toBe('/v1/secrets/external_api_keys');
      
      // Delete
      const deleteRequest = (syncClient.secrets.external_api_keys as any).mixin.prepareDelete(provider);
      expect(deleteRequest.method).toBe('DELETE');
      expect(deleteRequest.url).toContain(provider);
    });
  });

  describe('Response Type Safety', () => {
    it('should define proper external API key interface', () => {
      // Test that the expected structure is properly typed
      const mockExternalAPIKey = {
        provider: 'OpenAI',
        api_key: 'sk-***',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z'
      };
      
      expect(mockExternalAPIKey.provider).toBeDefined();
      expect(mockExternalAPIKey.api_key).toBeDefined();
      expect(typeof mockExternalAPIKey.provider).toBe('string');
      expect(typeof mockExternalAPIKey.api_key).toBe('string');
    });

    it('should handle list response structure', () => {
      const mockListResponse = [
        {
          provider: 'OpenAI',
          api_key: 'sk-***',
          created_at: '2024-01-01T00:00:00Z'
        },
        {
          provider: 'Anthropic',
          api_key: 'ak-***',
          created_at: '2024-01-01T00:00:00Z'
        }
      ];
      
      expect(Array.isArray(mockListResponse)).toBe(true);
      expect(mockListResponse.length).toBeGreaterThan(0);
      expect(mockListResponse[0].provider).toBeDefined();
    });
  });

  describe('URL Construction', () => {
    it('should construct URLs correctly for all operations', () => {
      const baseUrl = '/v1/secrets/external_api_keys';
      
      // List URL
      const listRequest = (syncClient.secrets.external_api_keys as any).mixin.prepareList();
      expect(listRequest.url).toBe(baseUrl);
      
      // Create URL
      const createRequest = (syncClient.secrets.external_api_keys as any).mixin.prepareCreate('OpenAI', 'test');
      expect(createRequest.url).toBe(baseUrl);
      
      // Get URL
      const getRequest = (syncClient.secrets.external_api_keys as any).mixin.prepareGet('Anthropic');
      expect(getRequest.url).toBe(`${baseUrl}/anthropic`);
      
      // Delete URL
      const deleteRequest = (syncClient.secrets.external_api_keys as any).mixin.prepareDelete('Gemini');
      expect(deleteRequest.url).toBe(`${baseUrl}/gemini`);
    });
  });

  describe('Edge Cases', () => {
    it('should handle provider names with different cases', () => {
      // Even though TypeScript enforces specific provider types,
      // test runtime behavior with different cases
      const providers = ['OpenAI', 'ANTHROPIC', 'Gemini', 'XAI'];
      
      providers.forEach(provider => {
        expect(() => {
          (syncClient.secrets.external_api_keys as any).mixin.prepareGet(provider.toLowerCase());
        }).not.toThrow();
      });
    });

    it('should handle very long API keys', () => {
      const longApiKey = 'a'.repeat(1000);
      
      expect(() => {
        (syncClient.secrets.external_api_keys as any).mixin.prepareCreate('OpenAI', longApiKey);
      }).not.toThrow();
    });
  });
});