import { Retab, AsyncRetab } from '../src/index.js';
import { TEST_API_KEY } from './fixtures.js';

describe('Usage Resource Tests', () => {
  let syncClient: Retab;
  let asyncClient: AsyncRetab;

  beforeEach(() => {
    syncClient = new Retab({ apiKey: TEST_API_KEY });
    asyncClient = new AsyncRetab({ apiKey: TEST_API_KEY });
  });

  describe('Resource Availability', () => {
    it('should have usage resource on sync client', () => {
      expect(syncClient.usage).toBeDefined();
      expect(typeof syncClient.usage.monthlyCreditsUsage).toBe('function');
      expect(typeof syncClient.usage.total).toBe('function');
      expect(typeof syncClient.usage.mailbox).toBe('function');
      expect(typeof syncClient.usage.link).toBe('function');
      expect(typeof syncClient.usage.schema).toBe('function');
      expect(typeof syncClient.usage.schemaData).toBe('function');
      expect(typeof syncClient.usage.log).toBe('function');
    });

    it('should have usage resource on async client', () => {
      expect(asyncClient.usage).toBeDefined();
      expect(typeof asyncClient.usage.monthlyCreditsUsage).toBe('function');
      expect(typeof asyncClient.usage.total).toBe('function');
      expect(typeof asyncClient.usage.mailbox).toBe('function');
      expect(typeof asyncClient.usage.link).toBe('function');
      expect(typeof asyncClient.usage.schema).toBe('function');
      expect(typeof asyncClient.usage.schemaData).toBe('function');
      expect(typeof asyncClient.usage.log).toBe('function');
    });
  });

  describe('Monthly Credits Usage', () => {
    it('should prepare monthly credits request', () => {
      const prepared = (syncClient.usage as any).mixin.prepareMonthlyCreditsUsage();
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe('/v1/usage/monthly_credits');
      expect(prepared.data).toBeUndefined();
      expect(prepared.params).toBeUndefined();
    });
  });

  describe('Total Usage', () => {
    it('should prepare total usage request with no dates', () => {
      const prepared = (syncClient.usage as any).mixin.prepareTotal();
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe('/v1/usage/total');
      expect(prepared.params).toEqual({});
    });

    it('should prepare total usage request with start date', () => {
      const startDate = new Date('2024-01-01');
      const prepared = (syncClient.usage as any).mixin.prepareTotal({ start_date: startDate });
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe('/v1/usage/total');
      expect(prepared.params).toEqual({
        start_date: startDate.toISOString()
      });
    });

    it('should prepare total usage request with both dates', () => {
      const startDate = new Date('2024-01-01');
      const endDate = new Date('2024-01-31');
      const prepared = (syncClient.usage as any).mixin.prepareTotal({ 
        start_date: startDate, 
        end_date: endDate 
      });
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe('/v1/usage/total');
      expect(prepared.params).toEqual({
        start_date: startDate.toISOString(),
        end_date: endDate.toISOString()
      });
    });

    it('should return amount structure for total usage', () => {
      const result = syncClient.usage.total();
      
      expect(result).toHaveProperty('value');
      expect(result).toHaveProperty('currency');
      expect(typeof result.value).toBe('number');
      expect(result.currency).toBe('USD');
    });
  });

  describe('Mailbox Usage', () => {
    it('should prepare mailbox usage request', () => {
      const email = 'test@example.com';
      const prepared = (syncClient.usage as any).mixin.prepareMailbox({ email });
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe(`/v1/processors/automations/mailboxes/${email}/usage`);
      expect(prepared.params).toEqual({});
    });

    it('should prepare mailbox usage request with dates', () => {
      const email = 'test@example.com';
      const startDate = new Date('2024-01-01');
      const endDate = new Date('2024-01-31');
      
      const prepared = (syncClient.usage as any).mixin.prepareMailbox({ 
        email, 
        start_date: startDate, 
        end_date: endDate 
      });
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe(`/v1/processors/automations/mailboxes/${email}/usage`);
      expect(prepared.params).toEqual({
        start_date: startDate.toISOString(),
        end_date: endDate.toISOString()
      });
    });

    it('should handle different email formats', () => {
      const emails = [
        'simple@example.com',
        'user+tag@example.com',
        'user.name@sub.example.com',
        'user_name@example-domain.com'
      ];
      
      emails.forEach(email => {
        expect(() => {
          (syncClient.usage as any).mixin.prepareMailbox({ email });
        }).not.toThrow();
      });
    });
  });

  describe('Link Usage', () => {
    it('should prepare link usage request', () => {
      const linkId = 'link-123';
      const prepared = (syncClient.usage as any).mixin.prepareLink({ link_id: linkId });
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe(`/v1/processors/automations/links/${linkId}/usage`);
      expect(prepared.params).toEqual({});
    });

    it('should prepare link usage request with dates', () => {
      const linkId = 'link-123';
      const startDate = new Date('2024-01-01');
      const endDate = new Date('2024-01-31');
      
      const prepared = (syncClient.usage as any).mixin.prepareLink({ 
        link_id: linkId, 
        start_date: startDate, 
        end_date: endDate 
      });
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe(`/v1/processors/automations/links/${linkId}/usage`);
      expect(prepared.params).toEqual({
        start_date: startDate.toISOString(),
        end_date: endDate.toISOString()
      });
    });
  });

  describe('Schema Usage', () => {
    it('should prepare schema usage request', () => {
      const schemaId = 'schema-123';
      const prepared = (syncClient.usage as any).mixin.prepareSchema({ schema_id: schemaId });
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe(`/v1/schemas/${schemaId}/usage`);
      expect(prepared.params).toEqual({});
    });

    it('should prepare schema usage request with dates', () => {
      const schemaId = 'schema-123';
      const startDate = new Date('2024-01-01');
      const endDate = new Date('2024-01-31');
      
      const prepared = (syncClient.usage as any).mixin.prepareSchema({ 
        schema_id: schemaId, 
        start_date: startDate, 
        end_date: endDate 
      });
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe(`/v1/schemas/${schemaId}/usage`);
      expect(prepared.params).toEqual({
        start_date: startDate.toISOString(),
        end_date: endDate.toISOString()
      });
    });
  });

  describe('Schema Data Usage', () => {
    it('should prepare schema data usage request', () => {
      const schemaDataId = 'schema-data-123';
      const prepared = (syncClient.usage as any).mixin.prepareSchemaData({ schema_data_id: schemaDataId });
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe(`/v1/schemas/${schemaDataId}/usage_data`);
      expect(prepared.params).toEqual({});
    });

    it('should prepare schema data usage request with dates', () => {
      const schemaDataId = 'schema-data-123';
      const startDate = new Date('2024-01-01');
      const endDate = new Date('2024-01-31');
      
      const prepared = (syncClient.usage as any).mixin.prepareSchemaData({ 
        schema_data_id: schemaDataId, 
        start_date: startDate, 
        end_date: endDate 
      });
      
      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe(`/v1/schemas/${schemaDataId}/usage_data`);
      expect(prepared.params).toEqual({
        start_date: startDate.toISOString(),
        end_date: endDate.toISOString()
      });
    });
  });

  describe('Usage Logging', () => {
    it('should prepare log request with valid response format', () => {
      const responseFormat = {
        json_schema: {
          schema: {
            type: 'object',
            properties: {
              name: { type: 'string' }
            }
          }
        }
      };
      
      const completion = {
        id: 'chatcmpl-123',
        model: 'gpt-4o',
        usage: { total_tokens: 100 }
      };
      
      const prepared = (syncClient.usage as any).mixin.prepareLog({ 
        response_format: responseFormat, 
        completion 
      });
      
      expect(prepared.method).toBe('POST');
      expect(prepared.url).toBe('/v1/usage/log');
      expect(prepared.data).toEqual({
        json_schema: responseFormat.json_schema.schema,
        completion
      });
    });

    it('should handle invalid response format', () => {
      const invalidFormats = [
        null,
        undefined,
        {},
        { invalid: 'format' },
        { json_schema: {} },
        { json_schema: { invalid: 'schema' } }
      ];
      
      const completion = { id: 'test' };
      
      invalidFormats.forEach(responseFormat => {
        expect(() => {
          (syncClient.usage as any).mixin.prepareLog({ 
            response_format: responseFormat, 
            completion 
          });
        }).toThrow('Invalid response format');
      });
    });
  });

  describe('Date Handling', () => {
    it('should properly format dates to ISO strings', () => {
      const testDate = new Date('2024-03-15T10:30:00.000Z');
      const prepared = (syncClient.usage as any).mixin.prepareTotal({ start_date: testDate });
      
      expect(prepared.params.start_date).toBe(testDate.toISOString());
    });

    it('should handle various date formats', () => {
      const dates = [
        new Date('2024-01-01'),
        new Date('2024-12-31T23:59:59.999Z'),
        new Date(Date.now()),
        new Date(0) // Unix epoch
      ];
      
      dates.forEach(date => {
        expect(() => {
          (syncClient.usage as any).mixin.prepareTotal({ start_date: date });
        }).not.toThrow();
      });
    });
  });

  describe('Client Type Parity', () => {
    it('should have same methods for sync and async clients', () => {
      const syncMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(syncClient.usage));
      const asyncMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(asyncClient.usage));
      
      const requiredMethods = [
        'monthlyCreditsUsage',
        'total',
        'mailbox',
        'link',
        'schema',
        'schemaData',
        'log'
      ];
      
      requiredMethods.forEach(method => {
        expect(syncMethods).toContain(method);
        expect(asyncMethods).toContain(method);
      });
    });

    it('should have consistent return types', () => {
      // Most methods should return promises
      const promiseMethods = [
        'monthlyCreditsUsage',
        'mailbox',
        'link',
        'schema',
        'schemaData',
        'log'
      ];
      
      promiseMethods.forEach(method => {
        const syncResult = (syncClient.usage as any)[method]({ email: 'test@example.com', link_id: 'test', schema_id: 'test', schema_data_id: 'test', response_format: { json_schema: { schema: {} } }, completion: {} });
        const asyncResult = (asyncClient.usage as any)[method]({ email: 'test@example.com', link_id: 'test', schema_id: 'test', schema_data_id: 'test', response_format: { json_schema: { schema: {} } }, completion: {} });
        
        expect(syncResult).toBeInstanceOf(Promise);
        expect(asyncResult).toBeInstanceOf(Promise);
      });
    });

    it('should handle total method differently (sync vs async)', () => {
      const syncResult = syncClient.usage.total();
      const asyncResult = asyncClient.usage.total();
      
      // Sync returns direct value, async returns promise
      expect(syncResult).toHaveProperty('value');
      expect(syncResult).toHaveProperty('currency');
      expect(asyncResult).toBeInstanceOf(Promise);
    });
  });

  describe('Parameter Validation', () => {
    it('should handle missing email for mailbox usage', () => {
      // The code should handle missing email gracefully (may produce invalid URL but not crash)
      expect(() => {
        (syncClient.usage as any).mixin.prepareMailbox({});
      }).not.toThrow();
    });

    it('should handle missing link_id for link usage', () => {
      // The code should handle missing link_id gracefully (may produce invalid URL but not crash)
      expect(() => {
        (syncClient.usage as any).mixin.prepareLink({});
      }).not.toThrow();
    });

    it('should handle missing schema_id for schema usage', () => {
      // The code should handle missing schema_id gracefully (may produce invalid URL but not crash)
      expect(() => {
        (syncClient.usage as any).mixin.prepareSchema({});
      }).not.toThrow();
    });

    it('should handle missing schema_data_id for schema data usage', () => {
      // The code should handle missing schema_data_id gracefully (may produce invalid URL but not crash)
      expect(() => {
        (syncClient.usage as any).mixin.prepareSchemaData({});
      }).not.toThrow();
    });
  });

  describe('URL Construction', () => {
    it('should construct correct URLs for all endpoints', () => {
      const testCases = [
        {
          method: 'prepareMonthlyCreditsUsage',
          params: [],
          expectedUrl: '/v1/usage/monthly_credits'
        },
        {
          method: 'prepareTotal',
          params: [{}],
          expectedUrl: '/v1/usage/total'
        },
        {
          method: 'prepareMailbox',
          params: [{ email: 'test@example.com' }],
          expectedUrl: '/v1/processors/automations/mailboxes/test@example.com/usage'
        },
        {
          method: 'prepareLink',
          params: [{ link_id: 'link-123' }],
          expectedUrl: '/v1/processors/automations/links/link-123/usage'
        },
        {
          method: 'prepareSchema',
          params: [{ schema_id: 'schema-123' }],
          expectedUrl: '/v1/schemas/schema-123/usage'
        },
        {
          method: 'prepareSchemaData',
          params: [{ schema_data_id: 'data-123' }],
          expectedUrl: '/v1/schemas/data-123/usage_data'
        }
      ];
      
      testCases.forEach(({ method, params, expectedUrl }) => {
        const prepared = (syncClient.usage as any).mixin[method](...params);
        expect(prepared.url).toBe(expectedUrl);
      });
    });
  });

  describe('Response Type Structures', () => {
    it('should define proper Amount interface', () => {
      const amount = syncClient.usage.total();
      
      expect(amount).toHaveProperty('value');
      expect(amount).toHaveProperty('currency');
      expect(typeof amount.value).toBe('number');
      expect(typeof amount.currency).toBe('string');
    });

    it('should handle global cost tracking', () => {
      // Test that total cost starts at 0
      const initialTotal = syncClient.usage.total();
      expect(initialTotal.value).toBe(0);
      expect(initialTotal.currency).toBe('USD');
    });
  });

  describe('Edge Cases', () => {
    it('should handle special characters in IDs and emails', () => {
      const specialCases = [
        { email: 'user+tag@example.com' },
        { link_id: 'link-with-dashes-123' },
        { schema_id: 'schema_with_underscores' },
        { schema_data_id: 'data.with.dots' }
      ];
      
      specialCases.forEach(params => {
        if ('email' in params) {
          expect(() => {
            (syncClient.usage as any).mixin.prepareMailbox(params);
          }).not.toThrow();
        }
        if ('link_id' in params) {
          expect(() => {
            (syncClient.usage as any).mixin.prepareLink(params);
          }).not.toThrow();
        }
        if ('schema_id' in params) {
          expect(() => {
            (syncClient.usage as any).mixin.prepareSchema(params);
          }).not.toThrow();
        }
        if ('schema_data_id' in params) {
          expect(() => {
            (syncClient.usage as any).mixin.prepareSchemaData(params);
          }).not.toThrow();
        }
      });
    });

    it('should handle empty strings in required parameters', () => {
      const emptyParams = [
        { email: '' },
        { link_id: '' },
        { schema_id: '' },
        { schema_data_id: '' }
      ];
      
      // These should not throw (URL construction will handle empty strings)
      emptyParams.forEach(params => {
        if ('email' in params) {
          expect(() => {
            (syncClient.usage as any).mixin.prepareMailbox(params);
          }).not.toThrow();
        }
      });
    });
  });
});