import {
  Links,
  AsyncLinks,
  LinksMixin
} from '../src/resources/processors/automations/links.js';
import { Link, ListLinks } from '../src/types/automations/links.js';

// Mock console.log to avoid cluttering test output
const originalConsoleLog = console.log;
beforeAll(() => {
  console.log = jest.fn();
});

afterAll(() => {
  console.log = originalConsoleLog;
});

describe('Automation Links Tests', () => {
  // Test data fixtures
  const mockLink: Link = {
    object: 'automation.link',
    id: 'link_12345',
    name: 'Test Link',
    processor_id: 'processor_67890',
    updated_at: '2024-01-01T12:00:00Z',
    default_language: 'en',
    webhook_url: 'https://example.com/webhook',
    webhook_headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer token123'
    },
    need_validation: true,
    password: 'secret123'
  };

  const mockListLinks: ListLinks = {
    data: [mockLink],
    list_metadata: {
      before: null,
      after: 'cursor_def456'
    }
  };

  const mockClient = {
    _preparedRequest: jest.fn()
  };

  describe('LinksMixin', () => {
    let mixin: LinksMixin;

    beforeEach(() => {
      mixin = new LinksMixin();
    });

    describe('Base URL Configuration', () => {
      it('should have correct base URL', () => {
        expect(mixin.linksBaseUrl).toBe('/v1/processors/automations/links');
      });
    });

    describe('prepareCreate', () => {
      it('should prepare create request with required parameters', () => {
        const params = {
          processor_id: 'processor_123',
          name: 'Test Link',
          webhook_url: 'https://example.com/webhook'
        };

        const result = mixin.prepareCreate(params);

        expect(result).toEqual({
          method: 'POST',
          url: '/v1/processors/automations/links',
          data: {
            processor_id: 'processor_123',
            name: 'Test Link',
            webhook_url: 'https://example.com/webhook'
          }
        });
      });

      it('should prepare create request with all optional parameters', () => {
        const params = {
          processor_id: 'processor_123',
          name: 'Test Link',
          webhook_url: 'https://example.com/webhook',
          webhook_headers: {
            'Authorization': 'Bearer token',
            'Custom-Header': 'value'
          },
          need_validation: false,
          password: 'secret123'
        };

        const result = mixin.prepareCreate(params);

        expect(result).toEqual({
          method: 'POST',
          url: '/v1/processors/automations/links',
          data: {
            processor_id: 'processor_123',
            name: 'Test Link',
            webhook_url: 'https://example.com/webhook',
            webhook_headers: {
              'Authorization': 'Bearer token',
              'Custom-Header': 'value'
            },
            need_validation: false,
            password: 'secret123'
          }
        });
      });

      it('should exclude undefined optional parameters', () => {
        const params = {
          processor_id: 'processor_123',
          name: 'Test Link',
          webhook_url: 'https://example.com/webhook',
          webhook_headers: undefined,
          need_validation: undefined,
          password: undefined
        };

        const result = mixin.prepareCreate(params);

        expect(result.data).not.toHaveProperty('webhook_headers');
        expect(result.data).not.toHaveProperty('need_validation');
        expect(result.data).not.toHaveProperty('password');
        expect(result.data).toEqual({
          processor_id: 'processor_123',
          name: 'Test Link',
          webhook_url: 'https://example.com/webhook'
        });
      });

      it('should handle empty string password', () => {
        const params = {
          processor_id: 'processor_123',
          name: 'Test Link',
          webhook_url: 'https://example.com/webhook',
          password: ''
        };

        const result = mixin.prepareCreate(params);

        expect(result.data.password).toBe('');
      });
    });

    describe('prepareList', () => {
      it('should prepare list request with no parameters', () => {
        const result = mixin.prepareList();

        expect(result).toEqual({
          method: 'GET',
          url: '/v1/processors/automations/links',
          params: {
            limit: 10,
            order: 'desc'
          }
        });
      });

      it('should prepare list request with empty parameters object', () => {
        const result = mixin.prepareList({});

        expect(result).toEqual({
          method: 'GET',
          url: '/v1/processors/automations/links',
          params: {
            limit: 10,
            order: 'desc'
          }
        });
      });

      it('should prepare list request with all parameters', () => {
        const params = {
          before: 'cursor_before',
          after: 'cursor_after',
          limit: 25,
          order: 'asc' as const,
          processor_id: 'processor_123',
          name: 'Test Link'
        };

        const result = mixin.prepareList(params);

        expect(result).toEqual({
          method: 'GET',
          url: '/v1/processors/automations/links',
          params: {
            before: 'cursor_before',
            after: 'cursor_after',
            limit: 25,
            order: 'asc',
            processor_id: 'processor_123',
            name: 'Test Link'
          }
        });
      });

      it('should use default values for limit and order', () => {
        const params = { processor_id: 'processor_123' };
        const result = mixin.prepareList(params);

        expect(result.params.limit).toBe(10);
        expect(result.params.order).toBe('desc');
      });

      it('should exclude undefined optional parameters', () => {
        const params = {
          before: undefined,
          after: undefined,
          processor_id: 'processor_123',
          name: undefined
        };

        const result = mixin.prepareList(params);

        expect(result.params).not.toHaveProperty('before');
        expect(result.params).not.toHaveProperty('after');
        expect(result.params).not.toHaveProperty('name');
        expect(result.params).toEqual({
          limit: 10,
          order: 'desc',
          processor_id: 'processor_123'
        });
      });

      it('should handle custom limit and order values', () => {
        const params = {
          limit: 50,
          order: 'asc' as const
        };

        const result = mixin.prepareList(params);

        expect(result.params.limit).toBe(50);
        expect(result.params.order).toBe('asc');
      });

      it('should handle zero limit', () => {
        const params = { limit: 0 };
        const result = mixin.prepareList(params);

        expect(result.params.limit).toBe(0);
      });
    });

    describe('prepareGet', () => {
      it('should prepare get request with link ID', () => {
        const linkId = 'link_12345';

        const result = mixin.prepareGet(linkId);

        expect(result).toEqual({
          method: 'GET',
          url: '/v1/processors/automations/links/link_12345'
        });
      });

      it('should handle various link ID formats', () => {
        const testIds = [
          'link_123',
          'lnk_abc-def-ghi',
          'automation-link-456',
          '12345'
        ];

        testIds.forEach(id => {
          const result = mixin.prepareGet(id);
          expect(result.url).toBe(`/v1/processors/automations/links/${id}`);
        });
      });
    });

    describe('prepareUpdate', () => {
      it('should prepare update request with minimal parameters', () => {
        const params = {
          link_id: 'link_123',
          name: 'Updated Link'
        };

        const result = mixin.prepareUpdate(params);

        expect(result).toEqual({
          method: 'PUT',
          url: '/v1/processors/automations/links/link_123',
          data: {
            name: 'Updated Link'
          }
        });
      });

      it('should prepare update request with all parameters', () => {
        const params = {
          link_id: 'link_123',
          name: 'Updated Link',
          webhook_url: 'https://updated.example.com/webhook',
          webhook_headers: {
            'Authorization': 'Bearer new_token',
            'Content-Type': 'application/json'
          },
          need_validation: false,
          password: 'newsecret123'
        };

        const result = mixin.prepareUpdate(params);

        expect(result).toEqual({
          method: 'PUT',
          url: '/v1/processors/automations/links/link_123',
          data: {
            name: 'Updated Link',
            webhook_url: 'https://updated.example.com/webhook',
            webhook_headers: {
              'Authorization': 'Bearer new_token',
              'Content-Type': 'application/json'
            },
            need_validation: false,
            password: 'newsecret123'
          }
        });
      });

      it('should exclude undefined parameters from update data', () => {
        const params = {
          link_id: 'link_123',
          name: 'Updated Link',
          webhook_url: undefined,
          webhook_headers: undefined,
          need_validation: undefined,
          password: undefined
        };

        const result = mixin.prepareUpdate(params);

        expect(result.data).toEqual({
          name: 'Updated Link'
        });
        expect(result.data).not.toHaveProperty('webhook_url');
        expect(result.data).not.toHaveProperty('webhook_headers');
        expect(result.data).not.toHaveProperty('need_validation');
        expect(result.data).not.toHaveProperty('password');
      });

      it('should handle empty update data', () => {
        const params = {
          link_id: 'link_123'
        };

        const result = mixin.prepareUpdate(params);

        expect(result.data).toEqual({});
      });

      it('should handle boolean false and empty string values correctly', () => {
        const params = {
          link_id: 'link_123',
          need_validation: false,
          password: ''
        };

        const result = mixin.prepareUpdate(params);

        expect(result.data.need_validation).toBe(false);
        expect(result.data.password).toBe('');
      });
    });

    describe('prepareDelete', () => {
      it('should prepare delete request with link ID', () => {
        const linkId = 'link_12345';

        const result = mixin.prepareDelete(linkId);

        expect(result).toEqual({
          method: 'DELETE',
          url: '/v1/processors/automations/links/link_12345',
          raiseForStatus: true
        });
      });

      it('should include raiseForStatus flag', () => {
        const result = mixin.prepareDelete('link_123');

        expect(result.raiseForStatus).toBe(true);
      });

      it('should handle various link ID formats for deletion', () => {
        const testIds = [
          'link_789',
          'lnk_xyz-123-abc',
          'automation-link-999'
        ];

        testIds.forEach(id => {
          const result = mixin.prepareDelete(id);
          expect(result.url).toBe(`/v1/processors/automations/links/${id}`);
        });
      });
    });
  });

  describe('Links (Sync)', () => {
    let links: Links;

    beforeEach(() => {
      links = new Links(mockClient as any);
      jest.clearAllMocks();
    });

    describe('create', () => {
      it('should create link and return promise', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockLink);

        const params = {
          processor_id: 'processor_123',
          name: 'Test Link',
          webhook_url: 'https://example.com/webhook'
        };

        const result = links.create(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'POST',
          url: '/v1/processors/automations/links',
          data: {
            processor_id: 'processor_123',
            name: 'Test Link',
            webhook_url: 'https://example.com/webhook'
          }
        });

        const link = await result;
        expect(link).toEqual(mockLink);
      });

      it('should handle create with all optional parameters', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockLink);

        const params = {
          processor_id: 'processor_123',
          name: 'Test Link',
          webhook_url: 'https://example.com/webhook',
          webhook_headers: { 'Authorization': 'Bearer token' },
          need_validation: true,
          password: 'secret123'
        };

        await links.create(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'POST',
          url: '/v1/processors/automations/links',
          data: expect.objectContaining({
            webhook_headers: { 'Authorization': 'Bearer token' },
            need_validation: true,
            password: 'secret123'
          })
        });
      });

      it('should log console message on successful creation', async () => {
        const mockResponse = { ...mockLink, id: 'link_new123' };
        mockClient._preparedRequest.mockResolvedValue(mockResponse);

        const params = {
          processor_id: 'processor_123',
          name: 'Test Link',
          webhook_url: 'https://example.com/webhook'
        };

        await links.create(params);

        // Give time for the promise chain to resolve
        await new Promise(resolve => setTimeout(resolve, 0));

        expect(console.log).toHaveBeenCalledWith(
          'Link Created. Url: https://www.retab.com/dashboard/processors/automations/link_new123'
        );
      });

      it('should handle create error gracefully for console log', async () => {
        mockClient._preparedRequest.mockRejectedValue(new Error('Creation failed'));

        const params = {
          processor_id: 'processor_123',
          name: 'Test Link',
          webhook_url: 'https://example.com/webhook'
        };

        await expect(links.create(params)).rejects.toThrow('Creation failed');

        // Console.log error handler should not throw
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      it('should handle missing ID in response for console log', async () => {
        const responseWithoutId = { ...mockLink };
        delete (responseWithoutId as any).id;
        mockClient._preparedRequest.mockResolvedValue(responseWithoutId);

        const params = {
          processor_id: 'processor_123',
          name: 'Test Link',
          webhook_url: 'https://example.com/webhook'
        };

        await links.create(params);

        // Give time for the promise chain to resolve
        await new Promise(resolve => setTimeout(resolve, 0));

        expect(console.log).toHaveBeenCalledWith(
          'Link Created. Url: https://www.retab.com/dashboard/processors/automations/unknown'
        );
      });
    });

    describe('list', () => {
      it('should list links with no parameters', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockListLinks);

        const result = await links.list();

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'GET',
          url: '/v1/processors/automations/links',
          params: {
            limit: 10,
            order: 'desc'
          }
        });

        expect(result).toEqual(mockListLinks);
      });

      it('should list links with empty parameters object', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockListLinks);

        const result = await links.list({});

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'GET',
          url: '/v1/processors/automations/links',
          params: {
            limit: 10,
            order: 'desc'
          }
        });

        expect(result).toEqual(mockListLinks);
      });

      it('should list links with all parameters', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockListLinks);

        const params = {
          before: 'cursor_before',
          after: 'cursor_after',
          limit: 25,
          order: 'asc' as const,
          processor_id: 'processor_123',
          name: 'Test'
        };

        await links.list(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'GET',
          url: '/v1/processors/automations/links',
          params: params
        });
      });

      it('should list links with partial parameters', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockListLinks);

        const params = {
          processor_id: 'processor_123',
          limit: 20
        };

        await links.list(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'GET',
          url: '/v1/processors/automations/links',
          params: {
            limit: 20,
            order: 'desc',
            processor_id: 'processor_123'
          }
        });
      });
    });

    describe('get', () => {
      it('should get link by ID', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockLink);

        const result = await links.get('link_123');

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'GET',
          url: '/v1/processors/automations/links/link_123'
        });

        expect(result).toEqual(mockLink);
      });
    });

    describe('update', () => {
      it('should update link with partial data', async () => {
        const updatedLink = { ...mockLink, name: 'Updated Name' };
        mockClient._preparedRequest.mockResolvedValue(updatedLink);

        const params = {
          link_id: 'link_123',
          name: 'Updated Name'
        };

        const result = await links.update(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'PUT',
          url: '/v1/processors/automations/links/link_123',
          data: { name: 'Updated Name' }
        });

        expect(result).toEqual(updatedLink);
      });

      it('should update link with all parameters', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockLink);

        const params = {
          link_id: 'link_123',
          name: 'Updated Name',
          webhook_url: 'https://new.example.com/webhook',
          webhook_headers: { 'New-Header': 'value' },
          password: 'newsecret',
          need_validation: false
        };

        await links.update(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'PUT',
          url: '/v1/processors/automations/links/link_123',
          data: {
            name: 'Updated Name',
            webhook_url: 'https://new.example.com/webhook',
            webhook_headers: { 'New-Header': 'value' },
            password: 'newsecret',
            need_validation: false
          }
        });
      });
    });

    describe('delete', () => {
      it('should delete link', async () => {
        mockClient._preparedRequest.mockResolvedValue(undefined);

        await links.delete('link_123');

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'DELETE',
          url: '/v1/processors/automations/links/link_123',
          raiseForStatus: true
        });
      });
    });
  });

  describe('AsyncLinks', () => {
    let asyncLinks: AsyncLinks;

    beforeEach(() => {
      asyncLinks = new AsyncLinks(mockClient as any);
      jest.clearAllMocks();
    });

    describe('create', () => {
      it('should create link asynchronously', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockLink);

        const params = {
          processor_id: 'processor_123',
          name: 'Test Link',
          webhook_url: 'https://example.com/webhook'
        };

        const result = await asyncLinks.create(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'POST',
          url: '/v1/processors/automations/links',
          data: {
            processor_id: 'processor_123',
            name: 'Test Link',
            webhook_url: 'https://example.com/webhook'
          }
        });

        expect(result).toEqual(mockLink);
        expect(console.log).toHaveBeenCalledWith(
          'Link Created. Url: https://www.retab.com/dashboard/processors/automations/link_12345'
        );
      });

      it('should handle create with all optional parameters', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockLink);

        const params = {
          processor_id: 'processor_123',
          name: 'Test Link',
          webhook_url: 'https://example.com/webhook',
          webhook_headers: { 'Authorization': 'Bearer token' },
          need_validation: true,
          password: 'secret123'
        };

        const result = await asyncLinks.create(params);

        expect(result).toEqual(mockLink);
        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'POST',
          url: '/v1/processors/automations/links',
          data: expect.objectContaining({
            webhook_headers: { 'Authorization': 'Bearer token' },
            need_validation: true,
            password: 'secret123'
          })
        });
      });
    });

    describe('list', () => {
      it('should list links asynchronously', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockListLinks);

        const result = await asyncLinks.list();

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'GET',
          url: '/v1/processors/automations/links',
          params: {
            limit: 10,
            order: 'desc'
          }
        });

        expect(result).toEqual(mockListLinks);
      });

      it('should list links with parameters asynchronously', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockListLinks);

        const params = {
          processor_id: 'processor_123',
          limit: 15,
          order: 'asc' as const
        };

        const result = await asyncLinks.list(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'GET',
          url: '/v1/processors/automations/links',
          params: {
            processor_id: 'processor_123',
            limit: 15,
            order: 'asc'
          }
        });

        expect(result).toEqual(mockListLinks);
      });
    });

    describe('get', () => {
      it('should get link asynchronously', async () => {
        mockClient._preparedRequest.mockResolvedValue(mockLink);

        const result = await asyncLinks.get('link_123');

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'GET',
          url: '/v1/processors/automations/links/link_123'
        });

        expect(result).toEqual(mockLink);
      });
    });

    describe('update', () => {
      it('should update link asynchronously', async () => {
        const updatedLink = { ...mockLink, name: 'Updated Async' };
        mockClient._preparedRequest.mockResolvedValue(updatedLink);

        const params = {
          link_id: 'link_123',
          name: 'Updated Async',
          password: 'newpassword'
        };

        const result = await asyncLinks.update(params);

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'PUT',
          url: '/v1/processors/automations/links/link_123',
          data: {
            name: 'Updated Async',
            password: 'newpassword'
          }
        });

        expect(result).toEqual(updatedLink);
      });
    });

    describe('delete', () => {
      it('should delete link asynchronously', async () => {
        mockClient._preparedRequest.mockResolvedValue(undefined);

        await asyncLinks.delete('link_123');

        expect(mockClient._preparedRequest).toHaveBeenCalledWith({
          method: 'DELETE',
          url: '/v1/processors/automations/links/link_123',
          raiseForStatus: true
        });
      });
    });
  });

  describe('Mixin Integration', () => {
    it('should use the same mixin logic in both sync and async classes', () => {
      const syncLinks = new Links(mockClient as any);
      const asyncLinks = new AsyncLinks(mockClient as any);

      expect(syncLinks.mixin).toBeInstanceOf(LinksMixin);
      expect(asyncLinks.mixin).toBeInstanceOf(LinksMixin);

      // Test that both use the same preparation logic
      const createParams = {
        processor_id: 'processor_123',
        name: 'Test',
        webhook_url: 'https://example.com'
      };

      const syncPrepared = syncLinks.mixin.prepareCreate(createParams);
      const asyncPrepared = asyncLinks.mixin.prepareCreate(createParams);

      expect(syncPrepared).toEqual(asyncPrepared);
    });

    it('should use consistent base URL across all methods', () => {
      const mixin = new LinksMixin();
      const baseUrl = '/v1/processors/automations/links';

      expect(mixin.linksBaseUrl).toBe(baseUrl);
      expect(mixin.prepareCreate({ processor_id: 'p', name: 'n', webhook_url: 'w' }).url).toBe(baseUrl);
      expect(mixin.prepareList().url).toBe(baseUrl);
      expect(mixin.prepareGet('link_123').url).toBe(`${baseUrl}/link_123`);
      expect(mixin.prepareUpdate({ link_id: 'link_123' }).url).toBe(`${baseUrl}/link_123`);
      expect(mixin.prepareDelete('link_123').url).toBe(`${baseUrl}/link_123`);
    });
  });

  describe('Error Handling and Edge Cases', () => {
    let links: Links;

    beforeEach(() => {
      links = new Links(mockClient as any);
      jest.clearAllMocks();
    });

    it('should handle network errors in requests', async () => {
      const networkError = new Error('Network error');
      mockClient._preparedRequest.mockRejectedValue(networkError);

      const params = {
        processor_id: 'processor_123',
        name: 'Test Link',
        webhook_url: 'https://example.com/webhook'
      };

      await expect(links.create(params)).rejects.toThrow('Network error');
    });

    it('should handle empty response gracefully', async () => {
      mockClient._preparedRequest.mockResolvedValue(null);

      const result = await links.get('link_123');
      expect(result).toBeNull();
    });

    it('should handle malformed responses', async () => {
      mockClient._preparedRequest.mockResolvedValue('invalid response');

      const result = await links.get('link_123');
      expect(result).toBe('invalid response');
    });

    it('should handle special characters in link IDs', async () => {
      const specialId = 'link_123-abc@example.com';
      mockClient._preparedRequest.mockResolvedValue(mockLink);

      await links.get(specialId);

      expect(mockClient._preparedRequest).toHaveBeenCalledWith({
        method: 'GET',
        url: `/v1/processors/automations/links/${specialId}`
      });
    });

    it('should handle empty webhook headers', () => {
      const mixin = new LinksMixin();
      const params = {
        processor_id: 'processor_123',
        name: 'Test',
        webhook_url: 'https://example.com',
        webhook_headers: {}
      };

      const result = mixin.prepareCreate(params);
      expect(result.data.webhook_headers).toEqual({});
    });

    it('should handle very long passwords', () => {
      const mixin = new LinksMixin();
      const longPassword = 'A'.repeat(1000);
      const params = {
        processor_id: 'processor_123',
        name: 'Test',
        webhook_url: 'https://example.com',
        password: longPassword
      };

      const result = mixin.prepareCreate(params);
      expect(result.data.password).toBe(longPassword);
    });

    it('should handle Unicode characters in parameters', () => {
      const mixin = new LinksMixin();
      const params = {
        processor_id: 'processor_123',
        name: '测试链接 🔗',
        webhook_url: 'https://example.com/webhook',
        password: 'пароль123'
      };

      const result = mixin.prepareCreate(params);
      expect(result.data.name).toBe('测试链接 🔗');
      expect(result.data.password).toBe('пароль123');
    });

    it('should handle boolean values in list parameters', () => {
      const mixin = new LinksMixin();
      const params = {
        limit: 0, // Edge case: zero limit
        processor_id: ''  // Edge case: empty string
      };

      const result = mixin.prepareList(params);
      expect(result.params.limit).toBe(0);
      expect(result.params.processor_id).toBe('');
    });
  });

  describe('Password Handling', () => {
    let mixin: LinksMixin;

    beforeEach(() => {
      mixin = new LinksMixin();
    });

    it('should include password in create request', () => {
      const params = {
        processor_id: 'processor_123',
        name: 'Test Link',
        webhook_url: 'https://example.com/webhook',
        password: 'secret123'
      };

      const result = mixin.prepareCreate(params);
      expect(result.data.password).toBe('secret123');
    });

    it('should include password in update request', () => {
      const params = {
        link_id: 'link_123',
        password: 'newsecret'
      };

      const result = mixin.prepareUpdate(params);
      expect(result.data.password).toBe('newsecret');
    });

    it('should handle empty password correctly', () => {
      const createParams = {
        processor_id: 'processor_123',
        name: 'Test Link',
        webhook_url: 'https://example.com/webhook',
        password: ''
      };

      const updateParams = {
        link_id: 'link_123',
        password: ''
      };

      const createResult = mixin.prepareCreate(createParams);
      const updateResult = mixin.prepareUpdate(updateParams);

      expect(createResult.data.password).toBe('');
      expect(updateResult.data.password).toBe('');
    });

    it('should exclude password when undefined', () => {
      const createParams = {
        processor_id: 'processor_123',
        name: 'Test Link',
        webhook_url: 'https://example.com/webhook',
        password: undefined
      };

      const updateParams = {
        link_id: 'link_123',
        password: undefined
      };

      const createResult = mixin.prepareCreate(createParams);
      const updateResult = mixin.prepareUpdate(updateParams);

      expect(createResult.data).not.toHaveProperty('password');
      expect(updateResult.data).not.toHaveProperty('password');
    });
  });
});