import { Retab, AsyncRetab } from '../src/index.js';
import {
  TEST_API_KEY
} from './fixtures.js';

describe('Projects', () => {
  let syncClient: Retab;
  let asyncClient: AsyncRetab;

  beforeEach(() => {
    syncClient = new Retab({ apiKey: TEST_API_KEY });
    asyncClient = new AsyncRetab({ apiKey: TEST_API_KEY });
  });

  describe('Resource Availability', () => {
    it('should have evaluations resource on sync client', () => {
      expect(syncClient.evaluations).toBeDefined();
      expect(syncClient.evaluations.documents).toBeDefined();
      expect(syncClient.evaluations.iterations).toBeDefined();
    });

    it('should have evaluations resource on async client', () => {
      expect(asyncClient.evaluations).toBeDefined();
      expect(asyncClient.evaluations.documents).toBeDefined();
      expect(asyncClient.evaluations.iterations).toBeDefined();
    });
  });

  describe('Basic Operations Placeholder', () => {
    ['sync', 'async'].forEach(clientType => {
      describe(`${clientType} client`, () => {
        it('should have evaluation resource structure', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;

          // Test that the evaluations resource exists and has the expected structure
          expect(client.evaluations).toBeDefined();
          expect(client.evaluations.documents).toBeDefined();
          expect(client.evaluations.iterations).toBeDefined();

          // Note: CRUD operations not yet implemented in Node SDK
          // This test ensures resource structure exists for future implementation
        });
      });
    });
  });

  describe('Document Resources Structure', () => {
    ['sync', 'async'].forEach(clientType => {
      describe(`${clientType} client`, () => {
        it('should have document resource structure', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;

          // Test that the document resources exist
          expect(client.evaluations.documents).toBeDefined();

          // Check that the documents resource has expected prototype methods
          const documentMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(client.evaluations.documents));
          expect(documentMethods.length).toBeGreaterThan(0);

          // Note: CRUD operations not yet implemented in Node SDK
          // This test ensures resource structure exists for future implementation
        });
      });
    });
  });

  describe('Iteration Resources Structure', () => {
    ['sync', 'async'].forEach(clientType => {
      describe(`${clientType} client`, () => {
        it('should have iteration resource structure', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;

          // Test that the iteration resources exist
          expect(client.evaluations.iterations).toBeDefined();

          // Check that the iterations resource has expected prototype methods
          const iterationMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(client.evaluations.iterations));
          expect(iterationMethods.length).toBeGreaterThan(0);

          // Note: CRUD operations not yet implemented in Node SDK
          // This test ensures resource structure exists for future implementation
        });
      });
    });
  });

  describe('Future Batch Operations', () => {
    it('should have structure for future batch operations', () => {
      const client = asyncClient;

      // Test that resources exist for future batch operations
      expect(client.evaluations).toBeDefined();
      expect(client.evaluations.documents).toBeDefined();

      // Note: Batch operations not yet implemented in Node SDK
      // This test ensures resource structure exists for future implementation
    });
  });

  describe('Future Error Handling', () => {
    it('should have basic error handling structure', () => {
      const client = syncClient;

      // Test that basic client structure exists for error handling
      expect(client.evaluations).toBeDefined();

      // Note: Error handling tests will be implemented when CRUD operations are available
      // This ensures the basic structure is in place
    });
  });

  describe('Future Concurrent Operations', () => {
    it('should have structure for concurrent operations', () => {
      const client = asyncClient;

      // Test that async client exists for future concurrent operations
      expect(client.evaluations).toBeDefined();

      // Note: Concurrent operations tests will be implemented when CRUD operations are available
      // This ensures the async structure is in place
    });
  });
});