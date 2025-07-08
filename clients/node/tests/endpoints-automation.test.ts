import {
  Endpoints,
  AsyncEndpoints,
  EndpointsMixin
} from '../src/resources/processors/automations/endpoints.js';
import { Endpoint, ListEndpoints } from '../src/types/automations/endpoints.js';

// Mock console.log to avoid cluttering test output
const originalConsoleLog = console.log;
beforeAll(() => {
  console.log = jest.fn();
});

afterAll(() => {
  console.log = originalConsoleLog;
});

describe('Automation Endpoints Tests', () => {
  // Test data fixtures
  const mockEndpoint: Endpoint = {
    object: 'automation.endpoint',
    id: 'endpoint_12345',
    name: 'Test Endpoint',
    processor_id: 'processor_67890',
    updated_at: '2024-01-01T12:00:00Z',
    default_language: 'en',
    webhook_url: 'https://example.com/webhook',
    webhook_headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer token123'
    },
    need_validation: true
  };

  const mockListEndpoints: ListEndpoints = {
    data: [mockEndpoint],
    list_metadata: {
      before: null,
      after: 'cursor_abc123'
    }
  };

  const mockClient = {
    _preparedRequest: jest.fn()
  };

  describe('EndpointsMixin', () => {
    let mixin: EndpointsMixin;

    beforeEach(() => {
      mixin = new EndpointsMixin();
    });

    describe('prepareCreate', () => {
      it('should prepare create request with required parameters', () => {
        const params = {
          processor_id: 'processor_123',
          name: 'Test Endpoint',
          webhook_url: 'https://example.com/webhook'
        };

        const result = mixin.prepareCreate(params);

        expect(result).toEqual({
          method: 'POST',
          url: '/v1/processors/automations/endpoints',
          data: {
            processor_id: 'processor_123',
            name: 'Test Endpoint',
            webhook_url: 'https://example.com/webhook',
            model: 'gpt-4o-mini'
          }
        });
      });

      it('should prepare create request with all optional parameters', () => {
        const params = {
          processor_id: 'processor_123',
          name: 'Test Endpoint',
          webhook_url: 'https://example.com/webhook',
          model: 'claude-3-haiku-20240307',
          webhook_headers: {
            'Authorization': 'Bearer token',
            'Custom-Header': 'value'
          },
          need_validation: false
        };

        const result = mixin.prepareCreate(params);

        expect(result).toEqual({
          method: 'POST',
          url: '/v1/processors/automations/endpoints',
          data: {
            processor_id: 'processor_123',
            name: 'Test Endpoint',
            webhook_url: 'https://example.com/webhook',
            model: 'claude-3-haiku-20240307',
            webhook_headers: {
              'Authorization': 'Bearer token',
              'Custom-Header': 'value'
            },
            need_validation: false
          }
        });
      });

      it('should use default model when not specified', () => {
        const params = {
          processor_id: 'processor_123',
          name: 'Test Endpoint',
          webhook_url: 'https://example.com/webhook'
        };

        const result = mixin.prepareCreate(params);

        expect(result.data.model).toBe('gpt-4o-mini');
      });

      it('should exclude undefined optional parameters', () => {
        const params = {
          processor_id: 'processor_123',
          name: 'Test Endpoint',
          webhook_url: 'https://example.com/webhook',
          webhook_headers: undefined,
          need_validation: undefined
        };

        const result = mixin.prepareCreate(params);

        expect(result.data).not.toHaveProperty('webhook_headers');
        expect(result.data).not.toHaveProperty('need_validation');
      });
    });

    describe('prepareList', () => {
      it('should prepare list request with required parameters', () => {
        const params = {
          processor_id: 'processor_123'
        };

        const result = mixin.prepareList(params);

        expect(result).toEqual({
          method: 'GET',
          url: '/v1/processors/automations/endpoints',
          params: {
            processor_id: 'processor_123',
            limit: 10,
            order: 'desc'
          }
        });
      });

      it('should prepare list request with all parameters', () => {
        const params = {
          processor_id: 'processor_123',
          before: 'cursor_before',
          after: 'cursor_after',
          limit: 50,
          order: 'asc' as const,
          name: 'Test Endpoint',
          webhook_url: 'https://example.com'
        };

        const result = mixin.prepareList(params);

        expect(result).toEqual({
          method: 'GET',
          url: '/v1/processors/automations/endpoints',
          params: {
            processor_id: 'processor_123',
            before: 'cursor_before',
            after: 'cursor_after',
            limit: 50,
            order: 'asc',
            name: 'Test Endpoint',
            webhook_url: 'https://example.com'
          }
        });
      });

      it('should use default values for limit and order', () => {
        const params = {
          processor_id: 'processor_123'
        };

        const result = mixin.prepareList(params);

        expect(result.params.limit).toBe(10);
        expect(result.params.order).toBe('desc');
      });

      it('should exclude undefined optional parameters', () => {
        const params = {
          processor_id: 'processor_123',
          before: undefined,
          after: undefined,
          name: undefined,
          webhook_url: undefined
        };

        const result = mixin.prepareList(params);

        expect(result.params).not.toHaveProperty('before');
        expect(result.params).not.toHaveProperty('after');
        expect(result.params).not.toHaveProperty('name');
        expect(result.params).not.toHaveProperty('webhook_url');
        expect(result.params).toEqual({
          processor_id: 'processor_123',
          limit: 10,
          order: 'desc'
        });
      });

      it('should handle order parameter variations', () => {
        const ascParams = { processor_id: 'processor_123', order: 'asc' as const };
        const descParams = { processor_id: 'processor_123', order: 'desc' as const };

        const ascResult = mixin.prepareList(ascParams);
        const descResult = mixin.prepareList(descParams);

        expect(ascResult.params.order).toBe('asc');
        expect(descResult.params.order).toBe('desc');
      });
    });

    describe('prepareGet', () => {
      it('should prepare get request with endpoint ID', () => {
        const endpointId = 'endpoint_12345';

        const result = mixin.prepareGet(endpointId);

        expect(result).toEqual({
          method: 'GET',
          url: '/v1/processors/automations/endpoints/endpoint_12345'
        });
      });

      it('should handle various endpoint ID formats', () => {
        const testIds = [
          'endpoint_123',
          'ep_abc-def-ghi',
          'automation-endpoint-456',
          '12345'
        ];

        testIds.forEach(id => {
          const result = mixin.prepareGet(id);
          expect(result.url).toBe(`/v1/processors/automations/endpoints/${id}`);
        });
      });
    });

    describe('prepareUpdate', () => {
      it('should prepare update request with minimal parameters', () => {
        const params = {
          endpoint_id: 'endpoint_123',
          name: 'Updated Endpoint'
        };

        const result = mixin.prepareUpdate(params);

        expect(result).toEqual({
          method: 'PUT',
          url: '/v1/processors/automations/endpoints/endpoint_123',
          data: {
            name: 'Updated Endpoint'
          }
        });
      });

      it('should prepare update request with all parameters', () => {
        const params = {
          endpoint_id: 'endpoint_123',
          name: 'Updated Endpoint',
          default_language: 'es',
          webhook_url: 'https://updated.example.com/webhook',
          webhook_headers: {
            'Authorization': 'Bearer new_token',
            'Content-Type': 'application/json'
          },
          need_validation: false
        };

        const result = mixin.prepareUpdate(params);

        expect(result).toEqual({
          method: 'PUT',
          url: '/v1/processors/automations/endpoints/endpoint_123',
          data: {
            name: 'Updated Endpoint',
            default_language: 'es',
            webhook_url: 'https://updated.example.com/webhook',
            webhook_headers: {
              'Authorization': 'Bearer new_token',
              'Content-Type': 'application/json'
            },
            need_validation: false
          }
        });
      });

      it('should exclude undefined parameters from update data', () => {
        const params = {
          endpoint_id: 'endpoint_123',
          name: 'Updated Endpoint',
          default_language: undefined,
          webhook_url: undefined,
          webhook_headers: undefined,
          need_validation: undefined
        };

        const result = mixin.prepareUpdate(params);

        expect(result.data).toEqual({
          name: 'Updated Endpoint'
        });
        expect(result.data).not.toHaveProperty('default_language');
        expect(result.data).not.toHaveProperty('webhook_url');
        expect(result.data).not.toHaveProperty('webhook_headers');
        expect(result.data).not.toHaveProperty('need_validation');
      });

      it('should handle empty update data', () => {
        const params = {
          endpoint_id: 'endpoint_123'
        };

        const result = mixin.prepareUpdate(params);

        expect(result.data).toEqual({});
      });

      it('should handle boolean false values correctly', () => {
        const params = {
          endpoint_id: 'endpoint_123',
          need_validation: false
        };

        const result = mixin.prepareUpdate(params);

        expect(result.data.need_validation).toBe(false);
      });
    });

    describe('prepareDelete', () => {
      it('should prepare delete request with endpoint ID', () => {
        const endpointId = 'endpoint_12345';

        const result = mixin.prepareDelete(endpointId);

        expect(result).toEqual({
          method: 'DELETE',
          url: '/v1/processors/automations/endpoints/endpoint_12345'
        });
      });

      it('should handle various endpoint ID formats for deletion', () => {
        const testIds = [
          'endpoint_789',
          'ep_xyz-123-abc',
          'automation-endpoint-999'
        ];

        testIds.forEach(id => {
          const result = mixin.prepareDelete(id);
          expect(result.url).toBe(`/v1/processors/automations/endpoints/${id}`);
        });
      });
    });
  });

  describe('Endpoints (Sync)', () => {
    let endpoints: Endpoints;

    beforeEach(() => {
      endpoints = new Endpoints(mockClient as any);
      jest.clearAllMocks();
    });

    describe('create', () => {
      it('should create endpoint and return promise', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockEndpoint);

        const params = {
          processor_id: 'processor_123',
          name: 'Test Endpoint',
          webhook_url: 'https://example.com/webhook'
        };

        const result = endpoints.create(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'POST',
          url: '/v1/processors/automations/endpoints',
          data: {
            processor_id: 'processor_123',
            name: 'Test Endpoint',
            webhook_url: 'https://example.com/webhook',
            model: 'gpt-4o-mini'
          }
        });

        const endpoint = await result;
        expect(endpoint).toEqual(mockEndpoint);
      });

      it('should handle create with optional parameters', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockEndpoint);

        const params = {
          processor_id: 'processor_123',
          name: 'Test Endpoint',
          webhook_url: 'https://example.com/webhook',
          webhook_headers: { 'Authorization': 'Bearer token' },
          need_validation: true
        };

        await endpoints.create(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'POST',
          url: '/v1/processors/automations/endpoints',
          data: expect.objectContaining({
            webhook_headers: { 'Authorization': 'Bearer token' },
            need_validation: true
          })
        });
      });

      it('should log console message on successful creation', async () => {
        const mockResponse = { ...mockEndpoint, id: 'endpoint_new123' };
        mockClient._preparedRequest.mockResolvedValue(mockResponse);

        const params = {
          processor_id: 'processor_123',
          name: 'Test Endpoint',
          webhook_url: 'https://example.com/webhook'
        };

        await endpoints.create(params);

        // Give time for the promise chain to resolve
        await new Promise(resolve => setTimeout(resolve, 0));

        expect(console.log).toHaveBeenCalledWith(
          'Endpoint Created. Url: https://www.retab.com/dashboard/processors/automations/endpoint_new123'
        );
      });

      it('should handle create error gracefully for console log', async () => {
        mockClient._preparedRequest.mockRejectedValue(new Error('Creation failed'));

        const params = {
          processor_id: 'processor_123',
          name: 'Test Endpoint',
          webhook_url: 'https://example.com/webhook'
        };

        await expect(endpoints.create(params)).rejects.toThrow('Creation failed');

        // Console.log error handler should not throw
        await new Promise(resolve => setTimeout(resolve, 0));
      });
    });

    describe('list', () => {
      it('should list endpoints with minimal parameters', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockListEndpoints);

        const params = { processor_id: 'processor_123' };
        const result = await endpoints.list(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'GET',
          url: '/v1/processors/automations/endpoints',
          params: {
            processor_id: 'processor_123',
            limit: 10,
            order: 'desc'
          }
        });

        expect(result).toEqual(mockListEndpoints);
      });

      it('should list endpoints with all parameters', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockListEndpoints);

        const params = {
          processor_id: 'processor_123',
          before: 'cursor_before',
          after: 'cursor_after',
          limit: 25,
          order: 'asc' as const,
          name: 'Test',
          webhook_url: 'https://example.com'
        };

        await endpoints.list(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'GET',
          url: '/v1/processors/automations/endpoints',
          params: params
        });
      });
    });

    describe('get', () => {
      it('should get endpoint by ID', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockEndpoint);

        const result = await endpoints.get('endpoint_123');

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'GET',
          url: '/v1/processors/automations/endpoints/endpoint_123'
        });

        expect(result).toEqual(mockEndpoint);
      });
    });

    describe('update', () => {
      it('should update endpoint with partial data', async () => {
        const updatedEndpoint = { ...mockEndpoint, name: 'Updated Name' };
        mockClient._preparedRequest.mockResolvedValue(updatedEndpoint);

        const params = {
          endpoint_id: 'endpoint_123',
          name: 'Updated Name'
        };

        const result = await endpoints.update(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'PUT',
          url: '/v1/processors/automations/endpoints/endpoint_123',
          data: { name: 'Updated Name' }
        });

        expect(result).toEqual(updatedEndpoint);
      });

      it('should update endpoint with all parameters', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockEndpoint);

        const params = {
          endpoint_id: 'endpoint_123',
          name: 'Updated Name',
          default_language: 'fr',
          webhook_url: 'https://new.example.com/webhook',
          webhook_headers: { 'New-Header': 'value' },
          need_validation: false
        };

        await endpoints.update(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'PUT',
          url: '/v1/processors/automations/endpoints/endpoint_123',
          data: {
            name: 'Updated Name',
            default_language: 'fr',
            webhook_url: 'https://new.example.com/webhook',
            webhook_headers: { 'New-Header': 'value' },
            need_validation: false
          }
        });
      });
    });

    describe('delete', () => {
      it('should delete endpoint and log message', async () => {
        mockClient._preparedRequest.mockResolvedValue(undefined);

        await endpoints.delete('endpoint_123');

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'DELETE',
          url: '/v1/processors/automations/endpoints/endpoint_123'
        });

        expect(console.log).toHaveBeenCalledWith('Endpoint Deleted. ID: endpoint_123');
      });
    });
  });

  describe('AsyncEndpoints', () => {
    let asyncEndpoints: AsyncEndpoints;

    beforeEach(() => {
      asyncEndpoints = new AsyncEndpoints(mockClient as any);
      jest.clearAllMocks();
    });

    describe('create', () => {
      it('should create endpoint asynchronously', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockEndpoint);

        const params = {
          processor_id: 'processor_123',
          name: 'Test Endpoint',
          webhook_url: 'https://example.com/webhook'
        };

        const result = await asyncEndpoints.create(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'POST',
          url: '/v1/processors/automations/endpoints',
          data: {
            processor_id: 'processor_123',
            name: 'Test Endpoint',
            webhook_url: 'https://example.com/webhook',
            model: 'gpt-4o-mini'
          }
        });

        expect(result).toEqual(mockEndpoint);
        expect(console.log).toHaveBeenCalledWith(
          'Endpoint Created. Url: https://www.retab.com/dashboard/processors/automations/endpoint_12345'
        );
      });

      it('should handle undefined response ID gracefully', async () => {
        const responseWithoutId = { ...mockEndpoint };
        delete (responseWithoutId as any).id;
        mockClient._preparedRequest.mockResolvedValue(responseWithoutId);

        const params = {
          processor_id: 'processor_123',
          name: 'Test Endpoint',
          webhook_url: 'https://example.com/webhook'
        };

        await asyncEndpoints.create(params);

        expect(console.log).toHaveBeenCalledWith(
          'Endpoint Created. Url: https://www.retab.com/dashboard/processors/automations/unknown'
        );
      });
    });

    describe('list', () => {
      it('should list endpoints asynchronously', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockListEndpoints);

        const params = { processor_id: 'processor_123' };
        const result = await asyncEndpoints.list(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'GET',
          url: '/v1/processors/automations/endpoints',
          params: {
            processor_id: 'processor_123',
            limit: 10,
            order: 'desc'
          }
        });

        expect(result).toEqual(mockListEndpoints);
      });
    });

    describe('get', () => {
      it('should get endpoint asynchronously', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockEndpoint);

        const result = await asyncEndpoints.get('endpoint_123');

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'GET',
          url: '/v1/processors/automations/endpoints/endpoint_123'
        });

        expect(result).toEqual(mockEndpoint);
      });
    });

    describe('update', () => {
      it('should update endpoint asynchronously', async () => {
        const updatedEndpoint = { ...mockEndpoint, name: 'Updated Async' };
        mockClient._preparedRequest.mockResolvedValue(updatedEndpoint);

        const params = {
          endpoint_id: 'endpoint_123',
          name: 'Updated Async',
          need_validation: false
        };

        const result = await asyncEndpoints.update(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'PUT',
          url: '/v1/processors/automations/endpoints/endpoint_123',
          data: {
            name: 'Updated Async',
            need_validation: false
          }
        });

        expect(result).toEqual(updatedEndpoint);
      });
    });

    describe('delete', () => {
      it('should delete endpoint asynchronously', async () => {
        mockClient._preparedRequest.mockResolvedValue(undefined);

        await asyncEndpoints.delete('endpoint_123');

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'DELETE',
          url: '/v1/processors/automations/endpoints/endpoint_123'
        });

        expect(console.log).toHaveBeenCalledWith('Endpoint Deleted. ID: endpoint_123');
      });
    });
  });

  describe('Mixin Integration', () => {
    it('should use the same mixin logic in both sync and async classes', () => {
      const syncEndpoints = new Endpoints(mockClient as any);
      const asyncEndpoints = new AsyncEndpoints(mockClient as any);

      expect(syncEndpoints.mixin).toBeInstanceOf(EndpointsMixin);
      expect(asyncEndpoints.mixin).toBeInstanceOf(EndpointsMixin);

      // Test that both use the same preparation logic
      const createParams = {
        processor_id: 'processor_123',
        name: 'Test',
        webhook_url: 'https://example.com'
      };

      const syncPrepared = syncEndpoints.mixin.prepareCreate(createParams);
      const asyncPrepared = asyncEndpoints.mixin.prepareCreate(createParams);

      expect(syncPrepared).toEqual(asyncPrepared);
    });
  });

  describe('Error Handling and Edge Cases', () => {
    let endpoints: Endpoints;

    beforeEach(() => {
      endpoints = new Endpoints(mockClient as any);
      jest.clearAllMocks();
    });

    it('should handle network errors in requests', async () => {
      const networkError = new Error('Network error');
      mockClient._preparedRequest.mockRejectedValue(networkError);

      const params = {
        processor_id: 'processor_123',
        name: 'Test Endpoint',
        webhook_url: 'https://example.com/webhook'
      };

      await expect(endpoints.create(params)).rejects.toThrow('Network error');
    });

    it('should handle empty response gracefully', async () => {
      mockClient._preparedRequest.mockResolvedValue(null);

      const result = await endpoints.get('endpoint_123');
      expect(result).toBeNull();
    });

    it('should handle malformed responses', async () => {
      mockClient._preparedRequest.mockResolvedValue('invalid response');

      const result = await endpoints.get('endpoint_123');
      expect(result).toBe('invalid response');
    });

    it('should handle special characters in endpoint IDs', async () => {
      const specialId = 'endpoint_123-abc@example.com';
      mockClient._preparedRequest.mockResolvedValue(mockEndpoint);

      await endpoints.get(specialId);

      expect(mockClient._preparedRequest).toHaveBeenCalledWith({
        method: 'GET',
        url: `/v1/processors/automations/endpoints/${specialId}`
      });
    });

    it('should handle empty webhook headers', () => {
      const mixin = new EndpointsMixin();
      const params = {
        processor_id: 'processor_123',
        name: 'Test',
        webhook_url: 'https://example.com',
        webhook_headers: {}
      };

      const result = mixin.prepareCreate(params);
      expect(result.data.webhook_headers).toEqual({});
    });

    it('should handle very long endpoint names', () => {
      const mixin = new EndpointsMixin();
      const longName = 'A'.repeat(1000);
      const params = {
        processor_id: 'processor_123',
        name: longName,
        webhook_url: 'https://example.com'
      };

      const result = mixin.prepareCreate(params);
      expect(result.data.name).toBe(longName);
    });
  });
});