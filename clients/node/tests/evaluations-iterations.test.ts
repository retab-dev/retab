import { Retab, AsyncRetab } from '../src/index.js';
import { TEST_API_KEY } from './fixtures.js';

describe('Projects Iterations Tests', () => {
  let syncClient: Retab;
  let asyncClient: AsyncRetab;

  beforeEach(() => {
    syncClient = new Retab({ apiKey: TEST_API_KEY });
    asyncClient = new AsyncRetab({ apiKey: TEST_API_KEY });
  });

  describe('Resource Availability', () => {
    it('should have iterations resource on sync client', () => {
      expect(syncClient.evaluations.iterations).toBeDefined();
      expect(typeof syncClient.evaluations.iterations.create).toBe('function');
      expect(typeof syncClient.evaluations.iterations.list).toBe('function');
      expect(typeof syncClient.evaluations.iterations.get).toBe('function');
      expect(typeof syncClient.evaluations.iterations.patch).toBe('function');
      expect(typeof syncClient.evaluations.iterations.delete).toBe('function');
      expect(typeof syncClient.evaluations.iterations.process).toBe('function');
      expect(typeof syncClient.evaluations.iterations.documentStatus).toBe('function');
      expect(typeof syncClient.evaluations.iterations.addFromJsonl).toBe('function');
    });

    it('should have iterations resource on async client', () => {
      expect(asyncClient.evaluations.iterations).toBeDefined();
      expect(typeof asyncClient.evaluations.iterations.create).toBe('function');
      expect(typeof asyncClient.evaluations.iterations.list).toBe('function');
      expect(typeof asyncClient.evaluations.iterations.get).toBe('function');
      expect(typeof asyncClient.evaluations.iterations.patch).toBe('function');
      expect(typeof asyncClient.evaluations.iterations.delete).toBe('function');
      expect(typeof asyncClient.evaluations.iterations.process).toBe('function');
      expect(typeof asyncClient.evaluations.iterations.documentStatus).toBe('function');
      expect(typeof asyncClient.evaluations.iterations.addFromJsonl).toBe('function');
    });
  });

  describe('Request Preparation - Create', () => {
    it('should prepare create iteration request', () => {
      const evaluationId = 'eval-123';
      const createParams = {
        inference_settings: {
          temperature: 0.7,
          max_tokens: 1000
        },
        json_schema: {
          type: 'object',
          properties: {
            result: { type: 'string' }
          }
        }
      };

      const prepared = (syncClient.evaluations.iterations as any).mixin.prepareCreate(
        evaluationId,
        createParams
      );

      expect(prepared.method).toBe('POST');
      expect(prepared.url).toBe(`/v1/projects/${evaluationId}/iterations`);
      expect(prepared.data).toEqual(createParams);
    });

    it('should handle different evaluation IDs', () => {
      const evaluationIds = [
        'eval-123',
        'evaluation_456',
        'test-eval-789',
        'eval.with.dots'
      ];

      evaluationIds.forEach(evalId => {
        const prepared = (syncClient.evaluations.iterations as any).mixin.prepareCreate(
          evalId,
          { inference_settings: { temperature: 0.5 } }
        );

        expect(prepared.url).toBe(`/v1/projects/${evalId}/iterations`);
      });
    });

    it('should handle various create parameters', () => {
      const evaluationId = 'eval-123';
      const createParams = {
        inference_settings: {
          temperature: 0.7,
          max_tokens: 1000,
          top_p: 0.9
        },
        json_schema: {
          type: 'object',
          properties: {
            result: { type: 'string' },
            confidence: { type: 'number' }
          }
        },
        parent_id: 'iter-previous-123'
      };

      const prepared = (syncClient.evaluations.iterations as any).mixin.prepareCreate(
        evaluationId,
        createParams
      );

      expect(prepared.data).toEqual(createParams);
      expect(prepared.data.inference_settings).toEqual(createParams.inference_settings);
      expect(prepared.data.json_schema).toEqual(createParams.json_schema);
    });
  });

  describe('Request Preparation - List', () => {
    it('should prepare list request with default parameters', () => {
      const evaluationId = 'eval-123';
      const prepared = (syncClient.evaluations.iterations as any).mixin.prepareList(evaluationId);

      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe(`/v1/projects/${evaluationId}/iterations`);
      expect(prepared.params).toEqual({
        limit: 10,
        order: 'desc'
      });
    });

    it('should prepare list request with pagination parameters', () => {
      const evaluationId = 'eval-123';
      const listParams = {
        before: 'iter-before-123',
        after: 'iter-after-456',
        limit: 25,
        order: 'asc' as const
      };

      const prepared = (syncClient.evaluations.iterations as any).mixin.prepareList(
        evaluationId,
        listParams
      );

      expect(prepared.params).toEqual({
        before: 'iter-before-123',
        after: 'iter-after-456',
        limit: 25,
        order: 'asc'
      });
    });

    it('should handle partial pagination parameters', () => {
      const evaluationId = 'eval-123';

      // Test with only before
      const beforeOnly = (syncClient.evaluations.iterations as any).mixin.prepareList(
        evaluationId,
        { before: 'iter-123' }
      );
      expect(beforeOnly.params.before).toBe('iter-123');
      expect(beforeOnly.params.limit).toBe(10);
      expect(beforeOnly.params.order).toBe('desc');

      // Test with only limit
      const limitOnly = (syncClient.evaluations.iterations as any).mixin.prepareList(
        evaluationId,
        { limit: 50 }
      );
      expect(limitOnly.params.limit).toBe(50);
      expect(limitOnly.params.order).toBe('desc');
      expect(limitOnly.params.before).toBeUndefined();
    });

    it('should handle edge cases for pagination', () => {
      const evaluationId = 'eval-123';

      const edgeCases = [
        { limit: 0 },
        { limit: 1 },
        { limit: 1000 },
        { before: '' },
        { after: '' },
        { order: 'asc' as const },
        { order: 'desc' as const }
      ];

      edgeCases.forEach(params => {
        expect(() => {
          (syncClient.evaluations.iterations as any).mixin.prepareList(evaluationId, params);
        }).not.toThrow();
      });
    });
  });

  describe('Request Preparation - Get', () => {
    it('should prepare get iteration request', () => {
      const evaluationId = 'eval-123';
      const iterationId = 'iter-456';

      const prepared = (syncClient.evaluations.iterations as any).mixin.prepareGet(
        evaluationId,
        iterationId
      );

      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe(`/v1/projects/${evaluationId}/iterations/${iterationId}`);
      expect(prepared.data).toBeUndefined();
      expect(prepared.params).toBeUndefined();
    });

    it('should handle different ID formats', () => {
      const idPairs = [
        ['eval-123', 'iter-456'],
        ['evaluation_789', 'iteration_abc'],
        ['test.eval', 'test.iter'],
        ['eval-with-dashes', 'iter-with-dashes'],
        ['eval_with_underscores', 'iter_with_underscores']
      ];

      idPairs.forEach(([evalId, iterId]) => {
        const prepared = (syncClient.evaluations.iterations as any).mixin.prepareGet(
          evalId,
          iterId
        );

        expect(prepared.url).toBe(`/v1/projects/${evalId}/iterations/${iterId}`);
      });
    });
  });

  describe('Request Preparation - Patch', () => {
    it('should prepare patch iteration request', () => {
      const evaluationId = 'eval-123';
      const iterationId = 'iter-456';
      const patchParams = {
        inference_settings: {
          temperature: 0.8,
          max_tokens: 1500
        },
        version: 2
      };

      const prepared = (syncClient.evaluations.iterations as any).mixin.preparePatch(
        evaluationId,
        iterationId,
        patchParams
      );

      expect(prepared.method).toBe('PATCH');
      expect(prepared.url).toBe(`/v1/projects/${evaluationId}/iterations/${iterationId}`);
      expect(prepared.data).toEqual(patchParams);
    });

    it('should handle partial update parameters', () => {
      const evaluationId = 'eval-123';
      const iterationId = 'iter-456';

      const partialUpdates = [
        { inference_settings: { temperature: 0.9 } },
        { json_schema: { type: 'object' } },
        { version: 3 },
        { inference_settings: { max_tokens: 2000 } },
        { json_schema: { type: 'array' }, version: 4 }
      ];

      partialUpdates.forEach(params => {
        const prepared = (syncClient.evaluations.iterations as any).mixin.preparePatch(
          evaluationId,
          iterationId,
          params
        );

        expect(prepared.data).toEqual(params);
      });
    });
  });

  describe('Request Preparation - Delete', () => {
    it('should prepare delete iteration request', () => {
      const evaluationId = 'eval-123';
      const iterationId = 'iter-456';

      const prepared = (syncClient.evaluations.iterations as any).mixin.prepareDelete(
        evaluationId,
        iterationId
      );

      expect(prepared.method).toBe('DELETE');
      expect(prepared.url).toBe(`/v1/projects/${evaluationId}/iterations/${iterationId}`);
      expect(prepared.data).toBeUndefined();
      expect(prepared.params).toBeUndefined();
    });
  });

  describe('Request Preparation - Process', () => {
    it('should prepare process iteration request', () => {
      const evaluationId = 'eval-123';
      const iterationId = 'iter-456';
      const processParams = {
        document_ids: ['doc-1', 'doc-2', 'doc-3'],
        only_outdated: true
      };

      const prepared = (syncClient.evaluations.iterations as any).mixin.prepareProcess(
        evaluationId,
        iterationId,
        processParams
      );

      expect(prepared.method).toBe('POST');
      expect(prepared.url).toBe(`/v1/projects/${evaluationId}/iterations/${iterationId}/process`);
      expect(prepared.data).toEqual(processParams);
    });

    it('should handle different process parameters', () => {
      const evaluationId = 'eval-123';
      const iterationId = 'iter-456';

      const processConfigs = [
        {
          document_ids: ['doc-1'],
          only_outdated: false
        },
        {
          document_ids: ['doc-1', 'doc-2', 'doc-3', 'doc-4'],
          only_outdated: true
        },
        {
          only_outdated: true
        },
        {
          document_ids: []
        }
      ];

      processConfigs.forEach(config => {
        const prepared = (syncClient.evaluations.iterations as any).mixin.prepareProcess(
          evaluationId,
          iterationId,
          config
        );

        expect(prepared.data).toEqual(config);
      });
    });
  });

  describe('Request Preparation - Document Status', () => {
    it('should prepare document status request', () => {
      const evaluationId = 'eval-123';
      const iterationId = 'iter-456';

      const prepared = (syncClient.evaluations.iterations as any).mixin.prepareDocumentStatus(
        evaluationId,
        iterationId
      );

      expect(prepared.method).toBe('GET');
      expect(prepared.url).toBe(`/v1/projects/${evaluationId}/iterations/${iterationId}/document-status`);
      expect(prepared.data).toBeUndefined();
      expect(prepared.params).toBeUndefined();
    });
  });

  describe('Request Preparation - Add From JSONL', () => {
    it('should prepare add from JSONL request', () => {
      const evaluationId = 'eval-123';
      const jsonlParams = {
        jsonl_gcs_path: 'gs://bucket/path/to/data.jsonl'
      };

      const prepared = (syncClient.evaluations.iterations as any).mixin.prepareAddFromJsonl(
        evaluationId,
        jsonlParams
      );

      expect(prepared.method).toBe('POST');
      expect(prepared.url).toBe(`/v1/projects/${evaluationId}/iterations/add-from-jsonl`);
      expect(prepared.data).toEqual(jsonlParams);
    });

    it('should handle different GCS paths', () => {
      const evaluationId = 'eval-123';

      const gcsPaths = [
        'gs://bucket/data.jsonl',
        'gs://my-bucket/folder/subfolder/data.jsonl',
        'gs://bucket-with-dashes/path_with_underscores/file.jsonl',
        'gs://bucket/very/deep/nested/path/data.jsonl'
      ];

      gcsPaths.forEach(path => {
        const jsonlParams = { jsonl_gcs_path: path };
        const prepared = (syncClient.evaluations.iterations as any).mixin.prepareAddFromJsonl(
          evaluationId,
          jsonlParams
        );

        expect(prepared.data.jsonl_gcs_path).toBe(path);
      });
    });

    it('should handle invalid GCS paths gracefully', () => {
      const evaluationId = 'eval-123';
      const invalidPaths = [
        '',
        'not-a-gcs-path',
        'gs://',
        'gs://bucket',
        'http://not-gcs/path.jsonl',
        'file:///local/path.jsonl'
      ];

      // Should not throw during preparation (validation happens server-side)
      invalidPaths.forEach(path => {
        expect(() => {
          (syncClient.evaluations.iterations as any).mixin.prepareAddFromJsonl(
            evaluationId,
            { jsonl_gcs_path: path }
          );
        }).not.toThrow();
      });
    });
  });

  describe('URL Construction', () => {
    it('should construct correct URLs for all endpoints', () => {
      const evaluationId = 'eval-test';
      const iterationId = 'iter-test';

      const testCases = [
        {
          method: 'prepareCreate',
          params: [evaluationId, { inference_settings: { temperature: 0.5 } }],
          expectedUrl: `/v1/projects/${evaluationId}/iterations`
        },
        {
          method: 'prepareList',
          params: [evaluationId, {}],
          expectedUrl: `/v1/projects/${evaluationId}/iterations`
        },
        {
          method: 'prepareGet',
          params: [evaluationId, iterationId],
          expectedUrl: `/v1/projects/${evaluationId}/iterations/${iterationId}`
        },
        {
          method: 'preparePatch',
          params: [evaluationId, iterationId, { version: 2 }],
          expectedUrl: `/v1/projects/${evaluationId}/iterations/${iterationId}`
        },
        {
          method: 'prepareDelete',
          params: [evaluationId, iterationId],
          expectedUrl: `/v1/projects/${evaluationId}/iterations/${iterationId}`
        },
        {
          method: 'prepareProcess',
          params: [evaluationId, iterationId, { only_outdated: true }],
          expectedUrl: `/v1/projects/${evaluationId}/iterations/${iterationId}/process`
        },
        {
          method: 'prepareDocumentStatus',
          params: [evaluationId, iterationId],
          expectedUrl: `/v1/projects/${evaluationId}/iterations/${iterationId}/document-status`
        },
        {
          method: 'prepareAddFromJsonl',
          params: [evaluationId, { jsonl_gcs_path: 'gs://bucket/data.jsonl' }],
          expectedUrl: `/v1/projects/${evaluationId}/iterations/add-from-jsonl`
        }
      ];

      testCases.forEach(({ method, params, expectedUrl }) => {
        const prepared = (syncClient.evaluations.iterations as any).mixin[method](...params);
        expect(prepared.url).toBe(expectedUrl);
      });
    });
  });

  describe('Client Type Parity', () => {
    it('should have same methods for sync and async clients', () => {
      const syncMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(syncClient.evaluations.iterations));
      const asyncMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(asyncClient.evaluations.iterations));

      const requiredMethods = [
        'create',
        'list',
        'get',
        'patch',
        'delete',
        'process',
        'documentStatus',
        'addFromJsonl'
      ];

      requiredMethods.forEach(method => {
        expect(syncMethods).toContain(method);
        expect(asyncMethods).toContain(method);
      });
    });

    it('should have consistent method signatures', () => {
      expect(syncClient.evaluations.iterations.create.length).toBe(asyncClient.evaluations.iterations.create.length);
      expect(syncClient.evaluations.iterations.list.length).toBe(asyncClient.evaluations.iterations.list.length);
      expect(syncClient.evaluations.iterations.get.length).toBe(asyncClient.evaluations.iterations.get.length);
      expect(syncClient.evaluations.iterations.patch.length).toBe(asyncClient.evaluations.iterations.patch.length);
      expect(syncClient.evaluations.iterations.delete.length).toBe(asyncClient.evaluations.iterations.delete.length);
      expect(syncClient.evaluations.iterations.process.length).toBe(asyncClient.evaluations.iterations.process.length);
      expect(syncClient.evaluations.iterations.documentStatus.length).toBe(asyncClient.evaluations.iterations.documentStatus.length);
      expect(syncClient.evaluations.iterations.addFromJsonl.length).toBe(asyncClient.evaluations.iterations.addFromJsonl.length);
    });

    it('should return promises from both clients for most methods', () => {
      const evaluationId = 'eval-test';
      const iterationId = 'iter-test';

      // Test promise-returning methods
      const promiseMethodTests = [
        () => syncClient.evaluations.iterations.get(evaluationId, iterationId),
        () => asyncClient.evaluations.iterations.get(evaluationId, iterationId),
        () => syncClient.evaluations.iterations.patch(evaluationId, iterationId, { version: 2 }),
        () => asyncClient.evaluations.iterations.patch(evaluationId, iterationId, { version: 2 }),
        () => syncClient.evaluations.iterations.documentStatus(evaluationId, iterationId),
        () => asyncClient.evaluations.iterations.documentStatus(evaluationId, iterationId),
        () => syncClient.evaluations.iterations.addFromJsonl(evaluationId, { jsonl_gcs_path: 'gs://bucket/data.jsonl' }),
        () => asyncClient.evaluations.iterations.addFromJsonl(evaluationId, { jsonl_gcs_path: 'gs://bucket/data.jsonl' })
      ];

      promiseMethodTests.forEach(test => {
        const result = test();
        expect(result).toBeInstanceOf(Promise);
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle missing required parameters gracefully', () => {
      const testCases = [
        () => (syncClient.evaluations.iterations as any).mixin.prepareCreate('', {}),
        () => (syncClient.evaluations.iterations as any).mixin.prepareGet('', ''),
        () => (syncClient.evaluations.iterations as any).mixin.prepareList(''),
        () => (syncClient.evaluations.iterations as any).mixin.preparePatch('', '', {}),
        () => (syncClient.evaluations.iterations as any).mixin.prepareDelete('', ''),
        () => (syncClient.evaluations.iterations as any).mixin.prepareProcess('', '', {}),
        () => (syncClient.evaluations.iterations as any).mixin.prepareDocumentStatus('', ''),
        () => (syncClient.evaluations.iterations as any).mixin.prepareAddFromJsonl('', {})
      ];

      testCases.forEach(test => {
        expect(() => test()).not.toThrow();
      });
    });

    it('should handle special characters in IDs', () => {
      const specialIds = [
        'eval-with-dashes',
        'eval_with_underscores',
        'eval.with.dots',
        'eval@with@symbols',
        'eval with spaces',
        'eval/with/slashes'
      ];

      specialIds.forEach(evalId => {
        specialIds.forEach(iterId => {
          expect(() => {
            (syncClient.evaluations.iterations as any).mixin.prepareGet(evalId, iterId);
          }).not.toThrow();
        });
      });
    });
  });

  describe('Data Validation', () => {
    it('should handle various create parameter types', () => {
      const evaluationId = 'eval-123';
      const parameterSets = [
        { inference_settings: { temperature: 0.5 } },
        { inference_settings: { temperature: 0.7, max_tokens: 1000 }, json_schema: { type: 'object' } },
        { inference_settings: { temperature: 0.0 }, parent_id: 'iter-123' },
        { json_schema: { type: 'array', items: { type: 'string' } } },
        { inference_settings: { temperature: 1.0, top_p: 0.9 }, json_schema: null }
      ];

      parameterSets.forEach(params => {
        expect(() => {
          (syncClient.evaluations.iterations as any).mixin.prepareCreate(evaluationId, params);
        }).not.toThrow();
      });
    });

    it('should handle various process parameter configurations', () => {
      const evaluationId = 'eval-123';
      const iterationId = 'iter-456';

      const processConfigs = [
        { document_ids: ['doc-1', 'doc-2'], only_outdated: false },
        { document_ids: ['doc-1'], only_outdated: true },
        { only_outdated: true },
        { document_ids: [], only_outdated: false }
      ];

      processConfigs.forEach(config => {
        expect(() => {
          (syncClient.evaluations.iterations as any).mixin.prepareProcess(
            evaluationId,
            iterationId,
            config
          );
        }).not.toThrow();
      });
    });
  });

  describe('Response Type Safety', () => {
    it('should define proper iteration response structure', () => {
      const mockIteration = {
        id: 'iter-123',
        project_id: 'eval-456',
        name: 'Test Iteration',
        description: 'A test iteration',
        status: 'completed',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        document_count: 100,
        processed_count: 95,
        failed_count: 5
      };

      expect(mockIteration.id).toBeDefined();
      expect(mockIteration.project_id).toBeDefined();
      expect(typeof mockIteration.name).toBe('string');
      expect(typeof mockIteration.status).toBe('string');
      expect(typeof mockIteration.document_count).toBe('number');
    });

    it('should handle document status response structure', () => {
      const mockDocumentStatus = {
        total_documents: 100,
        pending_documents: 5,
        processing_documents: 10,
        completed_documents: 80,
        failed_documents: 5,
        status_breakdown: {
          'pending': 5,
          'processing': 10,
          'completed': 80,
          'failed': 5
        }
      };

      expect(typeof mockDocumentStatus.total_documents).toBe('number');
      expect(typeof mockDocumentStatus.status_breakdown).toBe('object');
      expect(mockDocumentStatus.status_breakdown.completed).toBeDefined();
    });
  });
});