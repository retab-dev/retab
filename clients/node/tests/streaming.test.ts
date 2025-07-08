import { Retab, AsyncRetab } from '../src/index.js';
import { 
  TEST_API_KEY,
  bookingConfirmationSchema,
  companySchema
} from './fixtures.js';

// Mock streaming response chunks
const mockStreamingChunks = [
  {
    id: 'chatcmpl-123',
    object: 'chat.completion.chunk',
    created: 1677652288,
    model: 'gpt-4o-mini',
    choices: [{
      index: 0,
      delta: {
        content: '{"booking'
      },
      finish_reason: null
    }]
  },
  {
    id: 'chatcmpl-123',
    object: 'chat.completion.chunk',
    created: 1677652289,
    model: 'gpt-4o-mini',
    choices: [{
      index: 0,
      delta: {
        content: '_number": "BK-'
      },
      finish_reason: null
    }]
  },
  {
    id: 'chatcmpl-123',
    object: 'chat.completion.chunk',
    created: 1677652290,
    model: 'gpt-4o-mini',
    choices: [{
      index: 0,
      delta: {
        content: '12345"}'
      },
      finish_reason: 'stop'
    }]
  }
];

describe('Streaming Extractions', () => {
  let syncClient: Retab;
  let asyncClient: AsyncRetab;

  beforeEach(() => {
    syncClient = new Retab({ apiKey: TEST_API_KEY });
    asyncClient = new AsyncRetab({ apiKey: TEST_API_KEY });
  });

  describe('Basic Streaming Functionality', () => {
    const testCases = [
      { clientType: 'sync' as const, model: 'gpt-4o-mini' },
      { clientType: 'async' as const, model: 'gpt-4o-mini' },
    ];

    testCases.forEach(({ clientType, model }) => {
      it(`should have streaming method available for ${clientType} client`, () => {
        const client = clientType === 'sync' ? syncClient : asyncClient;
        
        // Test that streaming extraction method exists
        expect(client.documents.extractions.extractStream).toBeDefined();
        expect(typeof client.documents.extractions.extractStream).toBe('function');
      });

      it(`should create valid streaming request for ${clientType} client`, () => {
        const client = clientType === 'sync' ? syncClient : asyncClient;
        const mixin = (client.documents.extractions as any).mixin;
        
        if (mixin && typeof mixin.prepareExtraction === 'function') {
          const request = mixin.prepareExtraction(
            bookingConfirmationSchema,
            'test document content',
            undefined, // documents
            undefined, // imageResolutionDpi
            undefined, // browserCanvas
            model,
            undefined, // temperature
            'native', // modality
            undefined, // reasoningEffort
            true, // stream = true
            undefined, // nConsensus
            false, // store
            undefined // idempotencyKey
          );
          
          expect(request).toBeDefined();
          expect(request.method).toBe('POST');
          expect(request.url).toContain('/documents/extractions');
          expect(request.data.stream).toBe(true);
          expect(request.data.model).toBe(model);
          expect(request.data.modality).toBe('native');
        }
      });
    });
  });

  describe('Streaming Response Validation', () => {
    function validateStreamingChunk(chunk: any): void {
      expect(chunk).toBeDefined();
      
      // Basic chunk structure validation
      expect(chunk.id).toBeDefined();
      expect(chunk.object).toBeDefined();
      expect(chunk.choices).toBeDefined();
      expect(Array.isArray(chunk.choices)).toBe(true);
      expect(chunk.choices.length).toBeGreaterThan(0);
      
      const choice = chunk.choices[0];
      expect(choice).toBeDefined();
      expect(choice.index).toBeDefined();
      
      // Delta or message should be present
      expect(choice.delta !== undefined || choice.message !== undefined).toBe(true);
    }

    it('should validate mock streaming chunks', () => {
      mockStreamingChunks.forEach((chunk, index) => {
        validateStreamingChunk(chunk);
        
        // Additional validation for specific chunks
        if (index === 0) {
          expect(chunk.choices[0].delta.content).toBe('{"booking');
          expect(chunk.choices[0].finish_reason).toBe(null);
        } else if (index === mockStreamingChunks.length - 1) {
          expect(chunk.choices[0].finish_reason).toBe('stop');
        }
      });
    });

    it('should validate complete streaming response structure', () => {
      // Simulate accumulating chunks into a complete response
      let accumulatedContent = '';
      
      mockStreamingChunks.forEach(chunk => {
        if (chunk.choices[0].delta && chunk.choices[0].delta.content) {
          accumulatedContent += chunk.choices[0].delta.content;
        }
      });
      
      // Validate accumulated content is valid JSON
      expect(() => JSON.parse(accumulatedContent)).not.toThrow();
      
      const parsed = JSON.parse(accumulatedContent);
      expect(parsed).toBeDefined();
      expect(typeof parsed).toBe('object');
      expect(parsed.booking_number).toBe('BK-12345');
    });
  });

  describe('Streaming Parameters', () => {
    const streamingParams = [
      { model: 'gpt-4o-mini', modality: 'text' },
      { model: 'gpt-4o-mini', modality: 'native' },
      { model: 'gpt-4o', modality: 'text' },
    ];

    streamingParams.forEach(({ model, modality }) => {
      it(`should prepare streaming request with model: ${model}, modality: ${modality}`, () => {
        const mixin = (syncClient.documents.extractions as any).mixin;
        
        if (mixin && typeof mixin.prepareExtraction === 'function') {
          const request = mixin.prepareExtraction(
            bookingConfirmationSchema,
            'test document',
            undefined, // documents
            undefined, // imageResolutionDpi
            undefined, // browserCanvas
            model,
            0.7, // temperature
            modality,
            undefined, // reasoningEffort
            true, // stream = true
            1, // nConsensus
            false, // store
            undefined // idempotencyKey
          );
          
          expect(request).toBeDefined();
          expect(request.data.stream).toBe(true);
          expect(request.data.model).toBe(model);
          expect(request.data.modality).toBe(modality);
          expect(request.data.temperature).toBe(0.7);
          expect(request.data.n_consensus).toBe(1);
        }
      });
    });

    it('should handle streaming with temperature parameter', () => {
      const mixin = (syncClient.documents.extractions as any).mixin;
      
      if (mixin && typeof mixin.prepareExtraction === 'function') {
        const request = mixin.prepareExtraction(
          companySchema,
          'test document',
          undefined, // documents
          undefined, // imageResolutionDpi
          undefined, // browserCanvas
          'gpt-4o-mini',
          0.5, // temperature
          'native',
          undefined, // reasoningEffort
          true, // stream = true
          undefined, // nConsensus
          false, // store
          undefined // idempotencyKey
        );
        
        expect(request.data.temperature).toBe(0.5);
        expect(request.data.stream).toBe(true);
      }
    });

    it('should handle streaming with reasoning effort parameter', () => {
      const mixin = (syncClient.documents.extractions as any).mixin;
      
      if (mixin && typeof mixin.prepareExtraction === 'function') {
        const request = mixin.prepareExtraction(
          companySchema,
          'test document',
          undefined, // documents
          undefined, // imageResolutionDpi
          undefined, // browserCanvas
          'gpt-4o-mini',
          undefined, // temperature
          'native',
          'high', // reasoningEffort
          true, // stream = true
          undefined, // nConsensus
          false, // store
          undefined // idempotencyKey
        );
        
        expect(request.data.reasoning_effort).toBe('high');
        expect(request.data.stream).toBe(true);
      }
    });
  });

  describe('Streaming with Multiple Documents', () => {
    it('should prepare streaming request for multiple documents', () => {
      const documents = [
        'document 1 content',
        'document 2 content',
        'document 3 content'
      ];

      const mixin = (syncClient.documents.extractions as any).mixin;
      
      if (mixin && typeof mixin.prepareExtraction === 'function') {
        const request = mixin.prepareExtraction(
          companySchema,
          undefined, // single document
          documents, // multiple documents
          undefined, // imageResolutionDpi
          undefined, // browserCanvas
          'gpt-4o-mini',
          undefined, // temperature
          'native',
          undefined, // reasoningEffort
          true, // stream = true
          undefined, // nConsensus
          false, // store
          undefined // idempotencyKey
        );
        
        expect(request).toBeDefined();
        expect(request.data.stream).toBe(true);
        expect(request.data.documents).toBeDefined();
        expect(Array.isArray(request.data.documents)).toBe(true);
        expect(request.data.documents.length).toBe(3);
      }
    });
  });

  describe('Streaming Error Handling', () => {
    it('should handle invalid schema in streaming request', () => {
      const mixin = (syncClient.documents.extractions as any).mixin;
      
      if (mixin && typeof mixin.prepareExtraction === 'function') {
        expect(() => mixin.prepareExtraction(
          null, // Invalid schema
          'test document',
          undefined, // documents
          undefined, // imageResolutionDpi
          undefined, // browserCanvas
          'gpt-4o-mini',
          undefined, // temperature
          'native',
          undefined, // reasoningEffort
          true, // stream = true
          undefined, // nConsensus
          false, // store
          undefined // idempotencyKey
        )).toThrow();
      }
    });

    it('should handle missing required parameters in streaming request', () => {
      const mixin = (syncClient.documents.extractions as any).mixin;
      
      if (mixin && typeof mixin.prepareExtraction === 'function') {
        expect(() => mixin.prepareExtraction(
          companySchema,
          'test document',
          undefined, // documents
          undefined, // imageResolutionDpi
          undefined, // browserCanvas
          undefined, // model - required
          undefined, // temperature
          undefined, // modality - required
          undefined, // reasoningEffort
          true, // stream = true
          undefined, // nConsensus
          false, // store
          undefined // idempotencyKey
        )).toThrow();
      }
    });

    it('should handle conflicting document parameters', () => {
      const mixin = (syncClient.documents.extractions as any).mixin;
      
      if (mixin && typeof mixin.prepareExtraction === 'function') {
        expect(() => mixin.prepareExtraction(
          companySchema,
          'single document', // Both single and multiple provided
          ['doc1', 'doc2'], // documents
          undefined, // imageResolutionDpi
          undefined, // browserCanvas
          'gpt-4o-mini',
          undefined, // temperature
          'native',
          undefined, // reasoningEffort
          true, // stream = true
          undefined, // nConsensus
          false, // store
          undefined // idempotencyKey
        )).toThrow();
      }
    });
  });

  describe('Streaming with Idempotency', () => {
    it('should prepare streaming request with idempotency key', () => {
      const idempotencyKey = 'test-streaming-' + Date.now();
      const mixin = (syncClient.documents.extractions as any).mixin;
      
      if (mixin && typeof mixin.prepareExtraction === 'function') {
        const request = mixin.prepareExtraction(
          bookingConfirmationSchema,
          'test document',
          undefined, // documents
          undefined, // imageResolutionDpi
          undefined, // browserCanvas
          'gpt-4o-mini',
          undefined, // temperature
          'native',
          undefined, // reasoningEffort
          true, // stream = true
          undefined, // nConsensus
          false, // store
          idempotencyKey
        );
        
        expect(request).toBeDefined();
        expect(request.data.stream).toBe(true);
        expect(request.idempotencyKey).toBe(idempotencyKey);
      }
    });

    it('should handle idempotent streaming requests consistently', () => {
      const idempotencyKey = 'test-streaming-consistent-' + Date.now();
      const mixin = (syncClient.documents.extractions as any).mixin;
      
      if (mixin && typeof mixin.prepareExtraction === 'function') {
        const request1 = mixin.prepareExtraction(
          bookingConfirmationSchema,
          'test document',
          undefined, // documents
          undefined, // imageResolutionDpi
          undefined, // browserCanvas
          'gpt-4o-mini',
          0.7, // temperature
          'native',
          undefined, // reasoningEffort
          true, // stream = true
          undefined, // nConsensus
          false, // store
          idempotencyKey
        );
        
        const request2 = mixin.prepareExtraction(
          bookingConfirmationSchema,
          'test document',
          undefined, // documents
          undefined, // imageResolutionDpi
          undefined, // browserCanvas
          'gpt-4o-mini',
          0.7, // temperature
          'native',
          undefined, // reasoningEffort
          true, // stream = true
          undefined, // nConsensus
          false, // store
          idempotencyKey
        );
        
        // Requests should be identical for same idempotency key
        expect(request1.idempotencyKey).toBe(request2.idempotencyKey);
        expect(request1.data.stream).toBe(request2.data.stream);
        expect(request1.data.model).toBe(request2.data.model);
        expect(request1.data.temperature).toBe(request2.data.temperature);
      }
    });
  });

  describe('Streaming Performance and Concurrency', () => {
    it('should handle multiple concurrent streaming requests', () => {
      const concurrentRequests = 3;
      const requests = [];

      for (let i = 0; i < concurrentRequests; i++) {
        const mixin = (asyncClient.documents.extractions as any).mixin;
        if (mixin && typeof mixin.prepareExtraction === 'function') {
          const request = mixin.prepareExtraction(
            companySchema,
            `streaming document ${i}`,
            undefined, // documents
            undefined, // imageResolutionDpi
            undefined, // browserCanvas
            'gpt-4o-mini',
            undefined, // temperature
            'native',
            undefined, // reasoningEffort
            true, // stream = true
            undefined, // nConsensus
            false, // store
            `stream-${i}-${Date.now()}` // unique idempotency key
          );
          requests.push(request);
        }
      }

      expect(requests).toHaveLength(concurrentRequests);
      requests.forEach((req, index) => {
        expect(req).toBeDefined();
        expect(req.method).toBe('POST');
        expect(req.data.stream).toBe(true);
        expect(req.idempotencyKey).toContain(`stream-${index}`);
      });
    });

    it('should handle streaming with store parameter', () => {
      const mixin = (syncClient.documents.extractions as any).mixin;
      
      if (mixin && typeof mixin.prepareExtraction === 'function') {
        const request = mixin.prepareExtraction(
          bookingConfirmationSchema,
          'test document for storage',
          undefined, // documents
          undefined, // imageResolutionDpi
          undefined, // browserCanvas
          'gpt-4o-mini',
          undefined, // temperature
          'native',
          undefined, // reasoningEffort
          true, // stream = true
          undefined, // nConsensus
          true, // store = true
          undefined // idempotencyKey
        );
        
        expect(request.data.stream).toBe(true);
        expect(request.data.store).toBe(true);
      }
    });
  });

  describe('Streaming vs Non-Streaming Comparison', () => {
    it('should create different requests for streaming vs non-streaming', () => {
      const mixin = (syncClient.documents.extractions as any).mixin;
      
      if (mixin && typeof mixin.prepareExtraction === 'function') {
        const streamingRequest = mixin.prepareExtraction(
          bookingConfirmationSchema,
          'test document',
          undefined, // documents
          undefined, // imageResolutionDpi
          undefined, // browserCanvas
          'gpt-4o-mini',
          undefined, // temperature
          'native',
          undefined, // reasoningEffort
          true, // stream = true
          undefined, // nConsensus
          false, // store
          undefined // idempotencyKey
        );
        
        const nonStreamingRequest = mixin.prepareExtraction(
          bookingConfirmationSchema,
          'test document',
          undefined, // documents
          undefined, // imageResolutionDpi
          undefined, // browserCanvas
          'gpt-4o-mini',
          undefined, // temperature
          'native',
          undefined, // reasoningEffort
          false, // stream = false
          undefined, // nConsensus
          false, // store
          undefined // idempotencyKey
        );
        
        expect(streamingRequest.data.stream).toBe(true);
        expect(nonStreamingRequest.data.stream).toBe(false);
        
        // Other parameters should be the same
        expect(streamingRequest.data.model).toBe(nonStreamingRequest.data.model);
        expect(streamingRequest.data.modality).toBe(nonStreamingRequest.data.modality);
        expect(streamingRequest.url).toBe(nonStreamingRequest.url);
        expect(streamingRequest.method).toBe(nonStreamingRequest.method);
      }
    });
  });

  describe('Client Type Parity for Streaming', () => {
    it('should have same streaming methods for sync and async clients', () => {
      // Streaming methods
      const syncStreamMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(syncClient.documents.extractions));
      const asyncStreamMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(asyncClient.documents.extractions));
      
      // Both should have extractStream method
      expect(syncStreamMethods).toContain('extractStream');
      expect(asyncStreamMethods).toContain('extractStream');
      
      // Both should be async generator functions
      expect(typeof syncClient.documents.extractions.extractStream).toBe('function');
      expect(typeof asyncClient.documents.extractions.extractStream).toBe('function');
    });

    it('should have matching mixin methods for streaming', () => {
      const syncMixin = (syncClient.documents.extractions as any).mixin;
      const asyncMixin = (asyncClient.documents.extractions as any).mixin;
      
      expect(syncMixin).toBeDefined();
      expect(asyncMixin).toBeDefined();
      
      // Both should have prepareExtraction method
      expect(typeof syncMixin.prepareExtraction).toBe('function');
      expect(typeof asyncMixin.prepareExtraction).toBe('function');
    });
  });
});