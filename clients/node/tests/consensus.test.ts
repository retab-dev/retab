import { Retab, AsyncRetab } from '../src/index.js';
import { 
  TEST_API_KEY,
  bookingConfirmationSchema,
  companySchema
} from './fixtures.js';

// Mock consensus data for testing
const mockExtractionResults = [
  {
    booking_number: 'BK-12345',
    shipper_name: 'ABC Shipping Co.',
    container_number: 'CONT-67890',
    vessel_name: 'Pacific Explorer',
    port_of_loading: 'Shanghai',
    port_of_discharge: 'Los Angeles'
  },
  {
    booking_number: 'BK-12345',
    shipper_name: 'ABC Shipping Company',
    container_number: 'CONT-67890',
    vessel_name: 'Pacific Explorer',
    port_of_loading: 'Shanghai Port',
    port_of_discharge: 'Los Angeles Port'
  },
  {
    booking_number: 'BK-12345',
    shipper_name: 'ABC Shipping Co.',
    container_number: 'CONT-67890',
    vessel_name: 'Pacific Explorer',
    port_of_loading: 'Shanghai',
    port_of_discharge: 'Los Angeles'
  }
];

const mockConsensusResponse = {
  consensus_dict: {
    booking_number: 'BK-12345',
    shipper_name: 'ABC Shipping Co.',
    container_number: 'CONT-67890',
    vessel_name: 'Pacific Explorer',
    port_of_loading: 'Shanghai',
    port_of_discharge: 'Los Angeles'
  },
  likelihoods: {
    booking_number: 1.0,
    shipper_name: 0.85,
    container_number: 1.0,
    vessel_name: 1.0,
    port_of_loading: 0.9,
    port_of_discharge: 0.9
  }
};

describe('Consensus API', () => {
  let syncClient: Retab;
  let asyncClient: AsyncRetab;

  beforeEach(() => {
    syncClient = new Retab({ apiKey: TEST_API_KEY });
    asyncClient = new AsyncRetab({ apiKey: TEST_API_KEY });
  });

  describe('Resource Availability', () => {
    it('should have consensus resource on sync client', () => {
      expect(syncClient.consensus).toBeDefined();
      expect(typeof syncClient.consensus.reconcile).toBe('function');
      expect(typeof syncClient.consensus.extract).toBe('function');
    });

    it('should have consensus resource on async client', () => {
      expect(asyncClient.consensus).toBeDefined();
      expect(typeof asyncClient.consensus.reconcile).toBe('function');
      expect(typeof asyncClient.consensus.extract).toBe('function');
    });

    it('should have consensus mixin available', () => {
      const syncMixin = (syncClient.consensus as any).mixin;
      const asyncMixin = (asyncClient.consensus as any).mixin;
      
      expect(syncMixin).toBeDefined();
      expect(asyncMixin).toBeDefined();
      expect(typeof syncMixin.prepareReconcile).toBe('function');
      expect(typeof syncMixin.prepareExtract).toBe('function');
      expect(typeof asyncMixin.prepareReconcile).toBe('function');
      expect(typeof asyncMixin.prepareExtract).toBe('function');
    });
  });

  describe('Consensus Reconciliation', () => {
    const clientTypes = ['sync', 'async'] as const;

    clientTypes.forEach(clientType => {
      describe(`${clientType} client`, () => {
        it('should prepare reconciliation request with default parameters', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;
          const mixin = (client.consensus as any).mixin;
          
          const request = mixin.prepareReconcile({
            list_dicts: mockExtractionResults
          });
          
          expect(request).toBeDefined();
          expect(request.method).toBe('POST');
          expect(request.url).toBe('/v1/consensus/reconcile');
          expect(request.data).toBeDefined();
          expect(request.data.list_dicts).toEqual(mockExtractionResults);
          expect(request.data.mode).toBe('direct'); // default mode
          expect(request.data.reference_schema).toBeUndefined();
        });

        it('should prepare reconciliation request with all parameters', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;
          const mixin = (client.consensus as any).mixin;
          const idempotencyKey = 'consensus-test-' + Date.now();
          
          const request = mixin.prepareReconcile({
            list_dicts: mockExtractionResults,
            reference_schema: bookingConfirmationSchema,
            mode: 'aligned',
            idempotency_key: idempotencyKey
          });
          
          expect(request).toBeDefined();
          expect(request.method).toBe('POST');
          expect(request.url).toBe('/v1/consensus/reconcile');
          expect(request.data.list_dicts).toEqual(mockExtractionResults);
          expect(request.data.reference_schema).toEqual(bookingConfirmationSchema);
          expect(request.data.mode).toBe('aligned');
          expect(request.idempotency_key).toBe(idempotencyKey);
        });

        it('should handle different reconciliation modes', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;
          const mixin = (client.consensus as any).mixin;
          const modes = ['direct', 'aligned'] as const;
          
          modes.forEach(mode => {
            const request = mixin.prepareReconcile({
              list_dicts: mockExtractionResults,
              mode
            });
            
            expect(request.data.mode).toBe(mode);
          });
        });

        it('should handle empty extraction results array', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;
          const mixin = (client.consensus as any).mixin;
          
          const request = mixin.prepareReconcile({
            list_dicts: []
          });
          
          expect(request).toBeDefined();
          expect(request.data.list_dicts).toEqual([]);
        });

        it('should handle single extraction result', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;
          const mixin = (client.consensus as any).mixin;
          
          const request = mixin.prepareReconcile({
            list_dicts: [mockExtractionResults[0]]
          });
          
          expect(request).toBeDefined();
          expect(request.data.list_dicts).toHaveLength(1);
          expect(request.data.list_dicts[0]).toEqual(mockExtractionResults[0]);
        });
      });
    });
  });

  describe('Consensus Extraction', () => {
    const clientTypes = ['sync', 'async'] as const;

    clientTypes.forEach(clientType => {
      describe(`${clientType} client`, () => {
        it('should prepare consensus extract request with default parameters', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;
          const mixin = (client.consensus as any).mixin;
          
          const request = mixin.prepareExtract({
            document: 'test document content'
          });
          
          expect(request).toBeDefined();
          expect(request.method).toBe('POST');
          expect(request.url).toBe('/v1/consensus/extract');
          expect(request.form_data).toBeDefined();
          expect(request.files).toBeDefined();
          expect(request.files.document).toBe('test document content');
          
          // Check default values
          expect(request.form_data.model).toBe('gpt-4o-mini');
          expect(request.form_data.n_consensus).toBe(3);
          expect(request.form_data.modality).toBe('native');
          expect(request.form_data.temperature).toBe(0.0);
          expect(request.form_data.reasoning_effort).toBe('medium');
          expect(request.form_data.image_resolution_dpi).toBe(96);
          expect(request.form_data.browser_canvas).toBe('A4');
        });

        it('should prepare consensus extract request with all parameters', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;
          const mixin = (client.consensus as any).mixin;
          const idempotencyKey = 'extract-consensus-' + Date.now();
          
          const request = mixin.prepareExtract({
            document: 'test document for consensus extraction',
            model: 'gpt-4o',
            n_consensus: 5,
            json_schema: bookingConfirmationSchema,
            modality: 'text',
            image_resolution_dpi: 150,
            browser_canvas: 'Letter',
            temperature: 0.2,
            reasoning_effort: 'high',
            idempotency_key: idempotencyKey
          });
          
          expect(request).toBeDefined();
          expect(request.form_data.model).toBe('gpt-4o');
          expect(request.form_data.n_consensus).toBe(5);
          expect(request.form_data.json_schema).toEqual(bookingConfirmationSchema);
          expect(request.form_data.modality).toBe('text');
          expect(request.form_data.image_resolution_dpi).toBe(150);
          expect(request.form_data.browser_canvas).toBe('Letter');
          expect(request.form_data.temperature).toBe(0.2);
          expect(request.form_data.reasoning_effort).toBe('high');
          expect(request.idempotency_key).toBe(idempotencyKey);
          expect(request.files.document).toBe('test document for consensus extraction');
        });

        it('should handle different consensus values', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;
          const mixin = (client.consensus as any).mixin;
          const consensusValues = [1, 3, 5, 7, 10];
          
          consensusValues.forEach(nConsensus => {
            const request = mixin.prepareExtract({
              document: 'test document',
              n_consensus: nConsensus
            });
            
            expect(request.form_data.n_consensus).toBe(nConsensus);
          });
        });

        it('should handle different model types', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;
          const mixin = (client.consensus as any).mixin;
          const models = ['gpt-4o-mini', 'gpt-4o', 'gpt-4o-2024-11-20'];
          
          models.forEach(model => {
            const request = mixin.prepareExtract({
              document: 'test document',
              model
            });
            
            expect(request.form_data.model).toBe(model);
          });
        });

        it('should handle different modalities', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;
          const mixin = (client.consensus as any).mixin;
          const modalities = ['native', 'text', 'image'];
          
          modalities.forEach(modality => {
            const request = mixin.prepareExtract({
              document: 'test document',
              modality
            });
            
            expect(request.form_data.modality).toBe(modality);
          });
        });

        it('should handle temperature values', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;
          const mixin = (client.consensus as any).mixin;
          const temperatures = [0.0, 0.1, 0.5, 0.7, 1.0];
          
          temperatures.forEach(temperature => {
            const request = mixin.prepareExtract({
              document: 'test document',
              temperature
            });
            
            expect(request.form_data.temperature).toBe(temperature);
          });
        });

        it('should handle reasoning effort levels', () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;
          const mixin = (client.consensus as any).mixin;
          const efforts = ['low', 'medium', 'high'];
          
          efforts.forEach(effort => {
            const request = mixin.prepareExtract({
              document: 'test document',
              reasoning_effort: effort
            });
            
            expect(request.form_data.reasoning_effort).toBe(effort);
          });
        });
      });
    });
  });

  describe('Consensus Response Validation', () => {
    function validateConsensusResponse(response: any): void {
      expect(response).toBeDefined();
      expect(response.consensus_dict).toBeDefined();
      expect(typeof response.consensus_dict).toBe('object');
      expect(response.likelihoods).toBeDefined();
      expect(typeof response.likelihoods).toBe('object');
      
      // Validate that likelihoods are probabilities (0-1)
      Object.values(response.likelihoods).forEach((likelihood: any) => {
        expect(typeof likelihood).toBe('number');
        expect(likelihood).toBeGreaterThanOrEqual(0);
        expect(likelihood).toBeLessThanOrEqual(1);
      });
    }

    it('should validate mock consensus response structure', () => {
      validateConsensusResponse(mockConsensusResponse);
    });

    it('should validate consensus response with different confidence levels', () => {
      const variedConfidenceResponse = {
        consensus_dict: {
          field1: 'value1',
          field2: 'value2',
          field3: 'value3'
        },
        likelihoods: {
          field1: 1.0,   // Perfect confidence
          field2: 0.85,  // High confidence
          field3: 0.6    // Medium confidence
        }
      };
      
      validateConsensusResponse(variedConfidenceResponse);
    });

    it('should validate empty consensus response', () => {
      const emptyResponse = {
        consensus_dict: {},
        likelihoods: {}
      };
      
      validateConsensusResponse(emptyResponse);
    });
  });

  describe('Consensus Parameter Validation', () => {
    it('should handle consensus with different data types', () => {
      const mixedDataResults = [
        {
          string_field: 'text value',
          number_field: 123,
          boolean_field: true,
          null_field: null,
          array_field: ['item1', 'item2']
        },
        {
          string_field: 'text value',
          number_field: 124,
          boolean_field: true,
          null_field: null,
          array_field: ['item1', 'item2', 'item3']
        }
      ];

      const mixin = (syncClient.consensus as any).mixin;
      const request = mixin.prepareReconcile({
        list_dicts: mixedDataResults,
        mode: 'direct'
      });
      
      expect(request).toBeDefined();
      expect(request.data.list_dicts).toEqual(mixedDataResults);
    });

    it('should handle nested object structures in consensus', () => {
      const nestedResults = [
        {
          basic_field: 'value1',
          nested_object: {
            sub_field1: 'nested_value1',
            sub_field2: 42
          }
        },
        {
          basic_field: 'value1',
          nested_object: {
            sub_field1: 'nested_value1',
            sub_field2: 43
          }
        }
      ];

      const mixin = (syncClient.consensus as any).mixin;
      const request = mixin.prepareReconcile({
        list_dicts: nestedResults,
        reference_schema: companySchema,
        mode: 'aligned'
      });
      
      expect(request).toBeDefined();
      expect(request.data.list_dicts).toEqual(nestedResults);
      expect(request.data.reference_schema).toEqual(companySchema);
    });

    it('should handle large consensus datasets', () => {
      // Create a large dataset for consensus
      const largeResults = Array.from({ length: 50 }, (_, i) => ({
        id: i,
        value: `result_${i}`,
        score: Math.random(),
        category: i % 3 === 0 ? 'A' : i % 3 === 1 ? 'B' : 'C'
      }));

      const mixin = (syncClient.consensus as any).mixin;
      const request = mixin.prepareReconcile({
        list_dicts: largeResults,
        mode: 'direct'
      });
      
      expect(request).toBeDefined();
      expect(request.data.list_dicts).toHaveLength(50);
    });
  });

  describe('Consensus Error Handling', () => {
    it('should handle missing list_dicts parameter', () => {
      const mixin = (syncClient.consensus as any).mixin;
      
      // Should not throw but will create request with undefined list_dicts
      const request = mixin.prepareReconcile({});
      expect(request).toBeDefined();
      expect(request.data.list_dicts).toBeUndefined();
    });

    it('should handle invalid mode parameter', () => {
      const mixin = (syncClient.consensus as any).mixin;
      
      // This should not throw but should default to 'direct'
      const request = mixin.prepareReconcile({
        list_dicts: mockExtractionResults,
        mode: 'invalid_mode' as any
      });
      
      expect(request).toBeDefined();
      // The request is created but the API would reject invalid mode
    });

    it('should handle missing document parameter for extraction', () => {
      const mixin = (syncClient.consensus as any).mixin;
      
      // Should not throw but will create request with undefined document
      const request = mixin.prepareExtract({});
      expect(request).toBeDefined();
      expect(request.files.document).toBeUndefined();
    });

    it('should handle invalid consensus count', () => {
      const mixin = (syncClient.consensus as any).mixin;
      
      // Test with various edge cases
      const edgeCases = [-1, 0, 0.5, 'invalid' as any];
      
      edgeCases.forEach(nConsensus => {
        const request = mixin.prepareExtract({
          document: 'test',
          n_consensus: nConsensus
        });
        
        // Request is created but API would validate
        expect(request).toBeDefined();
      });
    });
  });

  describe('Consensus Idempotency', () => {
    it('should create consistent reconciliation requests with same idempotency key', () => {
      const idempotencyKey = 'consensus-reconcile-' + Date.now();
      const mixin = (syncClient.consensus as any).mixin;
      
      const request1 = mixin.prepareReconcile({
        list_dicts: mockExtractionResults,
        reference_schema: bookingConfirmationSchema,
        mode: 'direct',
        idempotency_key: idempotencyKey
      });
      
      const request2 = mixin.prepareReconcile({
        list_dicts: mockExtractionResults,
        reference_schema: bookingConfirmationSchema,
        mode: 'direct',
        idempotency_key: idempotencyKey
      });
      
      expect(request1.idempotency_key).toBe(request2.idempotency_key);
      expect(request1.data).toEqual(request2.data);
    });

    it('should create consistent extraction requests with same idempotency key', () => {
      const idempotencyKey = 'consensus-extract-' + Date.now();
      const mixin = (syncClient.consensus as any).mixin;
      
      const request1 = mixin.prepareExtract({
        document: 'test document',
        model: 'gpt-4o-mini',
        n_consensus: 3,
        temperature: 0.1,
        idempotency_key: idempotencyKey
      });
      
      const request2 = mixin.prepareExtract({
        document: 'test document',
        model: 'gpt-4o-mini',
        n_consensus: 3,
        temperature: 0.1,
        idempotency_key: idempotencyKey
      });
      
      expect(request1.idempotency_key).toBe(request2.idempotency_key);
      expect(request1.form_data).toEqual(request2.form_data);
      expect(request1.files).toEqual(request2.files);
    });
  });

  describe('Consensus Performance and Scalability', () => {
    it('should handle concurrent consensus reconciliations', () => {
      const concurrentRequests = 5;
      const requests = [];

      for (let i = 0; i < concurrentRequests; i++) {
        const mixin = (asyncClient.consensus as any).mixin;
        const request = mixin.prepareReconcile({
          list_dicts: mockExtractionResults,
          mode: 'direct',
          idempotency_key: `concurrent-reconcile-${i}-${Date.now()}`
        });
        requests.push(request);
      }

      expect(requests).toHaveLength(concurrentRequests);
      requests.forEach((req, index) => {
        expect(req).toBeDefined();
        expect(req.idempotency_key).toContain(`concurrent-reconcile-${index}`);
      });
    });

    it('should handle concurrent consensus extractions', () => {
      const concurrentRequests = 3;
      const requests = [];

      for (let i = 0; i < concurrentRequests; i++) {
        const mixin = (asyncClient.consensus as any).mixin;
        const request = mixin.prepareExtract({
          document: `test document ${i}`,
          n_consensus: 3,
          idempotency_key: `concurrent-extract-${i}-${Date.now()}`
        });
        requests.push(request);
      }

      expect(requests).toHaveLength(concurrentRequests);
      requests.forEach((req, index) => {
        expect(req).toBeDefined();
        expect(req.form_data.n_consensus).toBe(3);
        expect(req.idempotency_key).toContain(`concurrent-extract-${index}`);
      });
    });

    it('should handle varying consensus sizes', () => {
      const consensusSizes = [1, 3, 5, 7, 10, 15];
      const mixin = (syncClient.consensus as any).mixin;
      
      consensusSizes.forEach(size => {
        const request = mixin.prepareExtract({
          document: 'test document',
          n_consensus: size,
          temperature: size > 1 ? 0.1 : 0.0 // Handle temperature validation
        });
        
        expect(request.form_data.n_consensus).toBe(size);
      });
    });
  });

  describe('Client Type Parity for Consensus', () => {
    it('should have same consensus methods for sync and async clients', () => {
      // Check that both clients have the same method names
      const syncMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(syncClient.consensus));
      const asyncMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(asyncClient.consensus));
      
      expect(syncMethods).toContain('reconcile');
      expect(syncMethods).toContain('extract');
      expect(asyncMethods).toContain('reconcile');
      expect(asyncMethods).toContain('extract');
      
      // Check method types
      expect(typeof syncClient.consensus.reconcile).toBe('function');
      expect(typeof syncClient.consensus.extract).toBe('function');
      expect(typeof asyncClient.consensus.reconcile).toBe('function');
      expect(typeof asyncClient.consensus.extract).toBe('function');
    });

    it('should have matching mixin methods for consensus', () => {
      const syncMixin = (syncClient.consensus as any).mixin;
      const asyncMixin = (asyncClient.consensus as any).mixin;
      
      expect(syncMixin).toBeDefined();
      expect(asyncMixin).toBeDefined();
      
      // Both should have the same mixin methods
      expect(typeof syncMixin.prepareReconcile).toBe('function');
      expect(typeof syncMixin.prepareExtract).toBe('function');
      expect(typeof asyncMixin.prepareReconcile).toBe('function');
      expect(typeof asyncMixin.prepareExtract).toBe('function');
    });
  });
});