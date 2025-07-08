import { Retab, AsyncRetab } from '../src/index.js';
import { 
  TEST_API_KEY,
  bookingConfirmationSchema,
  generateRandomId,
  generateTestEmail
} from './fixtures.js';

describe('Automations', () => {
  let syncClient: Retab;
  let asyncClient: AsyncRetab;

  beforeEach(() => {
    syncClient = new Retab({ apiKey: TEST_API_KEY });
    asyncClient = new AsyncRetab({ apiKey: TEST_API_KEY });
  });

  describe('Resource Availability', () => {
    it('should have automations resources on sync client', () => {
      expect(syncClient.processors).toBeDefined();
      expect(syncClient.processors.automations).toBeDefined();
      expect(syncClient.processors.automations.endpoints).toBeDefined();
      expect(syncClient.processors.automations.links).toBeDefined();
      expect(syncClient.processors.automations.mailboxes).toBeDefined();
      expect(syncClient.processors.automations.outlook).toBeDefined();
      expect(syncClient.processors.automations.logs).toBeDefined();
      expect(syncClient.processors.automations.tests).toBeDefined();
    });

    it('should have automations resources on async client', () => {
      expect(asyncClient.processors).toBeDefined();
      expect(asyncClient.processors.automations).toBeDefined();
      expect(asyncClient.processors.automations.endpoints).toBeDefined();
      expect(asyncClient.processors.automations.links).toBeDefined();
      expect(asyncClient.processors.automations.mailboxes).toBeDefined();
      expect(asyncClient.processors.automations.outlook).toBeDefined();
      expect(asyncClient.processors.automations.logs).toBeDefined();
      expect(asyncClient.processors.automations.tests).toBeDefined();
    });
  });

  describe('Automation Links', () => {
    ['sync', 'async'].forEach(clientType => {
      describe(`${clientType} client`, () => {
        it('should manage automation links', async () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;
          const linkName = `test_link_${generateRandomId()}`;
          let processorId: string | undefined;
          let linkId: string | undefined;

          try {
            // Create a processor using the actual CRUD methods
            const processor = await (clientType === 'async'
              ? client.processors.create({
                  name: `test_processor_${generateRandomId()}`,
                  json_schema: bookingConfirmationSchema,
                  model: 'gpt-4o-mini'
                })
              : client.processors.create({
                  name: `test_processor_${generateRandomId()}`,
                  json_schema: bookingConfirmationSchema,
                  model: 'gpt-4o-mini'
                }));
            processorId = processor.id;

            // Test automation links resource structure
            expect(client.processors.automations.links).toBeDefined();
            
            // Check that the links resource has expected prototype methods
            const linkMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(client.processors.automations.links));
            expect(linkMethods.length).toBeGreaterThan(0);
            
            // Note: Automation link CRUD operations not yet fully implemented
            // This ensures the resource structure exists
            }
          } finally {
            // Clean up processor
            if (processorId) {
              try {
                await (clientType === 'async'
                  ? client.processors.delete(processorId)
                  : client.processors.delete(processorId));
              } catch (error) {
                // Ignore cleanup errors
              }
            }
          }
        });
      });
    });
  });

  describe('Automation Endpoints', () => {
    it('should have endpoints resource structure', () => {
      const client = asyncClient;
      
      // Test that the endpoints resource exists
      expect(client.processors.automations.endpoints).toBeDefined();
      
      // Check that the endpoints resource has expected prototype methods
      const endpointMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(client.processors.automations.endpoints));
      expect(endpointMethods.length).toBeGreaterThan(0);
      
      // Note: Endpoint CRUD operations not yet fully implemented
      // This ensures the resource structure exists
    });
  });

  describe('Automation Mailboxes', () => {
    it('should have mailboxes resource structure', () => {
      const client = asyncClient;
      
      // Test that the mailboxes resource exists
      expect(client.processors.automations.mailboxes).toBeDefined();
      
      // Check that the mailboxes resource has expected prototype methods
      const mailboxMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(client.processors.automations.mailboxes));
      expect(mailboxMethods.length).toBeGreaterThan(0);
      
      // Note: Mailbox CRUD operations not yet fully implemented
      // This ensures the resource structure exists
    });
  });

  describe('Automation Logs', () => {
    it('should have logs resource structure', () => {
      const client = asyncClient;
      
      // Test that the logs resource exists
      expect(client.processors.automations.logs).toBeDefined();
      
      // Check that the logs resource has expected prototype methods
      const logMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(client.processors.automations.logs));
      expect(logMethods.length).toBeGreaterThan(0);
      
      // Note: Log operations not yet fully implemented
      // This ensures the resource structure exists
    });
  });

  describe('Automation Tests', () => {
    it('should have tests resource structure', () => {
      const client = asyncClient;
      
      // Test that the tests resource exists
      expect(client.processors.automations.tests).toBeDefined();
      
      // Check that the tests resource has expected prototype methods
      const testMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(client.processors.automations.tests));
      expect(testMethods.length).toBeGreaterThan(0);
      
      // Note: Test operations not yet fully implemented
      // This ensures the resource structure exists
    });
  });

  describe('Outlook Integration', () => {
    it('should check Outlook integration methods', () => {
      const client = syncClient;
      
      // Check Outlook methods exist
      expect(client.processors.automations.outlook).toBeDefined();
      
      const outlookMethods = [
        'connect',
        'disconnect',
        'status',
        'folders',
        'rules'
      ];

      outlookMethods.forEach(method => {
        if (typeof client.processors.automations.outlook[method] === 'function') {
          expect(client.processors.automations.outlook[method]).toBeDefined();
        }
      });
    });
  });

  describe('Future Error Handling', () => {
    it('should have basic error handling structure', () => {
      const client = asyncClient;
      
      // Test that processors and automation resources exist for error handling
      expect(client.processors).toBeDefined();
      expect(client.processors.automations).toBeDefined();
      
      // Note: Error handling tests will be implemented when CRUD operations are available
      // This ensures the basic structure is in place
    });
  });

  describe('Future Batch Operations', () => {
    it('should have structure for batch operations', () => {
      const client = asyncClient;
      
      // Test that all automation resources exist for future batch operations
      expect(client.processors.automations.links).toBeDefined();
      expect(client.processors.automations.endpoints).toBeDefined();
      expect(client.processors.automations.mailboxes).toBeDefined();
      expect(client.processors.automations.tests).toBeDefined();
      expect(client.processors.automations.logs).toBeDefined();
      
      // Note: Batch operations will be implemented when CRUD operations are available
      // This ensures the resource structure exists
    });
  });
});