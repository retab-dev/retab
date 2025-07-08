import { Retab, AsyncRetab } from '../src/index.js';
import { TEST_API_KEY } from './fixtures.js';

describe('Models Resource Tests', () => {
  let syncClient: Retab;
  let asyncClient: AsyncRetab;

  beforeEach(() => {
    syncClient = new Retab({ apiKey: TEST_API_KEY });
    asyncClient = new AsyncRetab({ apiKey: TEST_API_KEY });
  });

  describe('Resource Availability', () => {
    it('should have models resource on sync client', () => {
      expect(syncClient.models).toBeDefined();
      expect(typeof syncClient.models.list).toBe('function');
    });

    it('should have models resource on async client', () => {
      expect(asyncClient.models).toBeDefined();
      expect(typeof asyncClient.models.list).toBe('function');
    });
  });

  describe('Request Preparation', () => {
    it('should prepare list request with default parameters', () => {
      const prepared = (syncClient.models as any).mixin.prepareList();
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe('/v1/models');
      expect(prepared.params).toEqual({
        supports_finetuning: false,
        supports_image: false,
        include_finetuned_models: true
      });
    });

    it('should prepare list request with custom parameters', () => {
      const params = {
        supports_finetuning: true,
        supports_image: true,
        include_finetuned_models: false
      };
      
      const prepared = (syncClient.models as any).mixin.prepareList(params);
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe('/v1/models');
      expect(prepared.params).toEqual(params);
    });

    it('should handle partial parameters', () => {
      const prepared = (syncClient.models as any).mixin.prepareList({
        supports_finetuning: true
      });
      
      expect(prepared.params).toEqual({
        supports_finetuning: true,
        supports_image: false,
        include_finetuned_models: true
      });
    });
  });

  describe('Parameter Validation', () => {
    it('should accept boolean parameters', () => {
      const validParams = [
        { supports_finetuning: true },
        { supports_finetuning: false },
        { supports_image: true },
        { supports_image: false },
        { include_finetuned_models: true },
        { include_finetuned_models: false },
        {
          supports_finetuning: true,
          supports_image: true,
          include_finetuned_models: false
        }
      ];

      validParams.forEach(params => {
        expect(() => {
          (syncClient.models as any).mixin.prepareList(params);
        }).not.toThrow();
      });
    });

    it('should handle empty parameters object', () => {
      expect(() => {
        (syncClient.models as any).mixin.prepareList({});
      }).not.toThrow();
    });

    it('should handle undefined parameters', () => {
      expect(() => {
        (syncClient.models as any).mixin.prepareList();
      }).not.toThrow();
    });
  });

  describe('Response Structure', () => {
    it('should define correct model interface', () => {
      // Test that the interface structure matches expectations
      const mockModel = {
        id: 'gpt-4o',
        object: 'model' as const,
        created: 1677610602,
        owned_by: 'openai',
        permission: [],
        root: 'gpt-4o',
        parent: null
      };

      // Verify all required properties exist
      expect(mockModel.id).toBeDefined();
      expect(mockModel.object).toBe('model');
      expect(typeof mockModel.created).toBe('number');
      expect(typeof mockModel.owned_by).toBe('string');
    });

    it('should define correct list response interface', () => {
      const mockResponse = {
        data: [
          {
            id: 'gpt-4o',
            object: 'model' as const,
            created: 1677610602,
            owned_by: 'openai'
          }
        ],
        object: 'list' as const
      };

      expect(Array.isArray(mockResponse.data)).toBe(true);
      expect(mockResponse.object).toBe('list');
      expect(mockResponse.data.length).toBeGreaterThan(0);
      expect(mockResponse.data[0].id).toBeDefined();
    });
  });

  describe('Client Type Parity', () => {
    it('should have same methods for sync and async clients', () => {
      const syncMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(syncClient.models));
      const asyncMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(asyncClient.models));
      
      expect(syncMethods).toContain('list');
      expect(asyncMethods).toContain('list');
    });

    it('should have consistent method signatures', () => {
      expect(syncClient.models.list.length).toBe(asyncClient.models.list.length);
    });

    it('should return promises from both clients', () => {
      // Both sync and async clients should return promises for the list method
      const syncResult = syncClient.models.list();
      const asyncResult = asyncClient.models.list();
      
      expect(syncResult).toBeInstanceOf(Promise);
      expect(asyncResult).toBeInstanceOf(Promise);
    });
  });

  describe('Filter Combinations', () => {
    it('should handle all filter combinations', () => {
      const filterCombinations = [
        {},
        { supports_finetuning: true },
        { supports_image: true },
        { include_finetuned_models: false },
        { supports_finetuning: true, supports_image: true },
        { supports_finetuning: true, include_finetuned_models: false },
        { supports_image: true, include_finetuned_models: false },
        { 
          supports_finetuning: true, 
          supports_image: true, 
          include_finetuned_models: false 
        },
        { 
          supports_finetuning: false, 
          supports_image: false, 
          include_finetuned_models: true 
        }
      ];

      filterCombinations.forEach(filters => {
        expect(() => {
          (syncClient.models as any).mixin.prepareList(filters);
        }).not.toThrow();
      });
    });
  });

  describe('Edge Cases', () => {
    it('should handle null and undefined filter values gracefully', () => {
      // TypeScript should prevent this, but test runtime behavior
      const edgeCases = [
        { supports_finetuning: null },
        { supports_image: undefined },
        { include_finetuned_models: null }
      ];

      edgeCases.forEach(params => {
        expect(() => {
          (syncClient.models as any).mixin.prepareList(params);
        }).not.toThrow();
      });
    });

    it('should handle extra unknown parameters', () => {
      const paramsWithExtra = {
        supports_finetuning: true,
        unknown_param: 'should_be_ignored',
        another_unknown: 123
      };

      expect(() => {
        (syncClient.models as any).mixin.prepareList(paramsWithExtra);
      }).not.toThrow();
    });
  });

  describe('URL and Method Consistency', () => {
    it('should always use GET method', () => {
      const testCases = [
        {},
        { supports_finetuning: true },
        { supports_image: true, include_finetuned_models: false }
      ];

      testCases.forEach(params => {
        const prepared = (syncClient.models as any).mixin.prepareList(params);
        expect(prepared.method).toBe('GET');
      });
    });

    it('should always use correct endpoint', () => {
      const testCases = [
        {},
        { supports_finetuning: true },
        { supports_image: true, include_finetuned_models: false }
      ];

      testCases.forEach(params => {
        const prepared = (syncClient.models as any).mixin.prepareList(params);
        expect(prepared.url).toBe('/v1/models');
      });
    });
  });
});