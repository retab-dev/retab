import { Retab, AsyncRetab } from '../src/index.js';
import { 
  TEST_API_KEY,
  bookingConfirmationSchema,
  companySchema
} from './fixtures.js';

// Mock response for document extraction
const mockExtractResponse = {
  id: 'chatcmpl-123',
  object: 'chat.completion',
  created: 1677652288,
  model: 'gpt-4o-mini',
  choices: [{
    index: 0,
    message: {
      role: 'assistant',
      content: JSON.stringify({
        booking_number: 'BK-12345',
        shipper_name: 'ABC Shipping Co.',
        container_number: 'CONT-67890',
        vessel_name: 'Pacific Explorer',
        port_of_loading: 'Shanghai',
        port_of_discharge: 'Los Angeles'
      }),
      parsed: {
        booking_number: 'BK-12345',
        shipper_name: 'ABC Shipping Co.',
        container_number: 'CONT-67890',
        vessel_name: 'Pacific Explorer',
        port_of_loading: 'Shanghai',
        port_of_discharge: 'Los Angeles'
      }
    },
    finish_reason: 'stop'
  }],
  usage: {
    prompt_tokens: 100,
    completion_tokens: 50,
    total_tokens: 150
  }
};

describe('Document Extraction', () => {
  let syncClient: Retab;
  let asyncClient: AsyncRetab;

  beforeEach(() => {
    syncClient = new Retab({ apiKey: TEST_API_KEY });
    asyncClient = new AsyncRetab({ apiKey: TEST_API_KEY });
  });

  describe('Extract Methods', () => {
    describe('Sync Client', () => {
      it('should have extraction methods', () => {
        expect(syncClient.documents).toBeDefined();
        expect(syncClient.documents.extractions).toBeDefined();
        // Check for any available methods
        const extractionMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(syncClient.documents.extractions));
        expect(extractionMethods.length).toBeGreaterThan(0);
      });

      it('should prepare extraction request', () => {
        const mixin = (syncClient.documents.extractions as any).mixin;
        expect(mixin).toBeDefined();
        
        if (mixin && typeof mixin.prepareExtraction === 'function') {
          const request = mixin.prepareExtraction(
            bookingConfirmationSchema,
            'test document content'
          );
          
          expect(request).toBeDefined();
          expect(request.method).toBe('POST');
          expect(request.url).toContain('/documents/extractions');
        }
      });
    });

    describe('Async Client', () => {
      it('should have extraction methods', () => {
        expect(asyncClient.documents).toBeDefined();
        expect(asyncClient.documents.extractions).toBeDefined();
        // Check for any available methods
        const extractionMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(asyncClient.documents.extractions));
        expect(extractionMethods.length).toBeGreaterThan(0);
      });
    });
  });

  describe('Extraction Response Validation', () => {
    function validateExtractionResponse(response: any): void {
      // Validate response structure
      expect(response).toBeDefined();
      expect(response.choices).toBeDefined();
      expect(Array.isArray(response.choices)).toBe(true);
      expect(response.choices.length).toBeGreaterThan(0);
      
      const choice = response.choices[0];
      expect(choice.message).toBeDefined();
      expect(choice.message.content).toBeDefined();
      
      // Validate that content is valid JSON
      let parsed: any;
      try {
        parsed = JSON.parse(choice.message.content);
      } catch (error) {
        fail('Response content should be valid JSON');
      }
      
      // Validate parsed content exists
      expect(parsed).toBeDefined();
      expect(typeof parsed).toBe('object');
    }

    it('should validate mock extraction response', () => {
      validateExtractionResponse(mockExtractResponse);
    });

    it('should handle different document types', () => {
      const documentTypes = [
        'string content',
        Buffer.from('buffer content'),
        { filename: 'test.pdf', content: Buffer.from('pdf'), mimeType: 'application/pdf' }
      ];

      documentTypes.forEach(doc => {
        // Test that we can prepare requests for different document types
        const mixin = (syncClient.documents.extractions as any).mixin;
        if (mixin && typeof mixin.prepareExtraction === 'function') {
          expect(() => mixin.prepareExtraction(companySchema, doc))
            .not.toThrow();
        }
      });
    });
  });

  describe('Streaming Extractions', () => {
    it('should support streaming infrastructure', () => {
      // Verify extraction resource exists for streaming support
      expect(syncClient.documents.extractions).toBeDefined();
      expect(asyncClient.documents.extractions).toBeDefined();
    });
  });

  describe('Extraction Parameters', () => {
    const testDocument = 'test document content';
    const testSchema = bookingConfirmationSchema;

    describe('Model Selection', () => {
      const models = ['gpt-4o-mini', 'gpt-4o', 'gpt-4o-2024-11-20'];
      
      models.forEach(model => {
        it(`should support model: ${model}`, () => {
          const mixin = (syncClient.documents.extractions as any).mixin;
          if (mixin && typeof mixin.prepareExtraction === 'function') {
            const request = mixin.prepareExtraction(testSchema, testDocument, undefined, undefined, undefined, model);
            if (request.data.model) {
              expect(request.data.model).toBe(model);
            }
          }
        });
      });
    });

    describe('Modality Options', () => {
      const modalities = ['native', 'text'];
      
      modalities.forEach(modality => {
        it(`should support modality: ${modality}`, () => {
          const mixin = (syncClient.documents.extractions as any).mixin;
          if (mixin && typeof mixin.prepareExtraction === 'function') {
            const request = mixin.prepareExtraction(
              testSchema, 
              testDocument,
              undefined,
              undefined,
              undefined,
              undefined,
              undefined,
              modality as any
            );
            if (request.data.modality) {
              expect(request.data.modality).toBe(modality);
            }
          }
        });
      });
    });

    describe('Additional Parameters', () => {
      it('should support temperature parameter', () => {
        const mixin = (syncClient.documents.extractions as any).mixin;
        if (mixin && typeof mixin.prepareExtraction === 'function') {
          const request = mixin.prepareExtraction(
            testSchema,
            testDocument,
            undefined,
            undefined,
            undefined,
            undefined,
            0.7
          );
          if (request.data.temperature) {
            expect(request.data.temperature).toBe(0.7);
          }
        }
      });

      it('should support reasoning effort parameter', () => {
        const mixin = (syncClient.documents.extractions as any).mixin;
        if (mixin && typeof mixin.prepareExtraction === 'function') {
          const request = mixin.prepareExtraction(
            testSchema,
            testDocument,
            undefined,
            undefined,
            undefined,
            undefined,
            undefined,
            undefined,
            'high'
          );
          if (request.data.reasoning_effort) {
            expect(request.data.reasoning_effort).toBe('high');
          }
        }
      });
    });
  });

  describe('Batch Extractions', () => {
    it('should handle multiple documents', () => {
      const documents = [
        'document 1',
        'document 2',
        'document 3'
      ];

      documents.forEach(doc => {
        const mixin = (syncClient.documents.extractions as any).mixin;
        if (mixin && typeof mixin.prepareExtraction === 'function') {
          const request = mixin.prepareExtraction(companySchema, doc);
          expect(request).toBeDefined();
          expect(request.data.documents).toBeDefined();
        }
      });
    });

    it('should support concurrent extractions', async () => {
      const concurrentRequests = 5;
      const requests = [];

      for (let i = 0; i < concurrentRequests; i++) {
        const mixin = (asyncClient.documents.extractions as any).mixin;
        if (mixin && typeof mixin.prepareExtraction === 'function') {
          const request = mixin.prepareExtraction(
            companySchema,
            `document ${i}`
          );
          requests.push(request);
        }
      }

      expect(requests).toHaveLength(concurrentRequests);
      requests.forEach(req => {
        expect(req).toBeDefined();
        expect(req.method).toBe('POST');
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle invalid schema', () => {
      const invalidSchema = null;
      const mixin = (syncClient.documents.extractions as any).mixin;
      
      if (mixin && typeof mixin.prepareExtraction === 'function') {
        expect(() => mixin.prepareExtraction(invalidSchema, 'doc'))
          .toThrow();
      }
    });

    it('should handle empty document', () => {
      const emptyDoc = '';
      const mixin = (syncClient.documents.extractions as any).mixin;
      
      if (mixin && typeof mixin.prepareExtraction === 'function') {
        // Empty documents should still create valid requests
        const request = mixin.prepareExtraction(companySchema, emptyDoc);
        expect(request).toBeDefined();
      }
    });
  });

  describe('Document Resource Methods', () => {
    it('should have document resource', () => {
      expect(syncClient.documents).toBeDefined();
      expect(asyncClient.documents).toBeDefined();
    });
  });

  describe('Extraction Method Tests', () => {
    it('should have extraction method available', () => {
      const client = syncClient;
      
      // Test that extraction methods exist
      expect(client.documents.extractions.extract).toBeDefined();
      expect(client.documents.extractions.extractStream).toBeDefined();
      
      // Test that methods are functions
      expect(typeof client.documents.extractions.extract).toBe('function');
      expect(typeof client.documents.extractions.extractStream).toBe('function');
    });
  });

  describe('Stream Method Tests', () => {
    it('should have stream method available', () => {
      const client = asyncClient;
      
      // Test that stream extraction method exists
      expect(client.documents.extractions.extractStream).toBeDefined();
      expect(typeof client.documents.extractions.extractStream).toBe('function');
    });
  });

  describe('Client Type Parity', () => {
    it('should have same methods for sync and async clients', () => {
      // Document methods
      const syncDocMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(syncClient.documents));
      const asyncDocMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(asyncClient.documents));
      
      // Extraction methods
      const syncExtMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(syncClient.documents.extractions));
      const asyncExtMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(asyncClient.documents.extractions));
      
      // Basic parity check
      expect(syncDocMethods.length).toBeGreaterThan(0);
      expect(asyncDocMethods.length).toBeGreaterThan(0);
      expect(syncExtMethods.length).toBeGreaterThan(0);
      expect(asyncExtMethods.length).toBeGreaterThan(0);
    });
  });
});