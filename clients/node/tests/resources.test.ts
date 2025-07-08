import { Retab, AsyncRetab, Schema } from '../src/index.js';
import { 
  TEST_API_KEY, 
  companySchema
} from './fixtures.js';

describe('Resource Methods', () => {
  let syncClient: Retab;
  let asyncClient: AsyncRetab;

  beforeEach(() => {
    syncClient = new Retab({ apiKey: TEST_API_KEY });
    asyncClient = new AsyncRetab({ apiKey: TEST_API_KEY });
  });

  describe('Schemas Resource', () => {
    describe('Load Method', () => {
      it('should load schema from JSON schema object', () => {
        const schema = syncClient.schemas.load(companySchema);
        expect(schema).toBeInstanceOf(Schema);
        expect(schema.json_schema).toEqual(companySchema);
      });

      it('should load schema from JSON schema string', () => {
        const schemaString = JSON.stringify(companySchema);
        const schema = syncClient.schemas.load(schemaString);
        expect(schema).toBeInstanceOf(Schema);
        expect(schema.json_schema).toEqual(companySchema);
      });

      it('should load schema from pydantic model', () => {
        const mockPydanticModel = {
          model_json_schema: () => companySchema
        };
        const schema = syncClient.schemas.load(null, mockPydanticModel);
        expect(schema).toBeInstanceOf(Schema);
        expect(schema.json_schema).toEqual(companySchema);
      });

      it('should throw error when neither parameter provided', () => {
        expect(() => syncClient.schemas.load())
          .toThrow('Either json_schema or pydantic_model must be provided');
      });

      it('should work with async client', () => {
        const schema = asyncClient.schemas.load(companySchema);
        expect(schema).toBeInstanceOf(Schema);
        expect(schema.json_schema).toEqual(companySchema);
      });
    });

    describe('Prepare Methods', () => {
      const testDocuments = ['test document content'];
      
      describe('Prepare Generate', () => {
        it('should prepare generate request with defaults', () => {
          const mixin = (syncClient.schemas as any).mixin;
          const request = mixin.prepareGenerate(testDocuments);
          
          expect(request.method).toBe('POST');
          expect(request.url).toBe('/v1/schemas/generate');
          expect(request.data).toBeDefined();
          expect(request.data.documents).toBeDefined();
          expect(request.data.model).toBe('gpt-4o-2024-11-20');
          expect(request.data.temperature).toBe(0);
          expect(request.data.modality).toBe('native');
          expect(request.data.reasoning_effort).toBe('medium');
        });

        it('should prepare generate request with custom parameters', () => {
          const mixin = (syncClient.schemas as any).mixin;
          const request = mixin.prepareGenerate(
            testDocuments,
            'Custom instructions',
            'gpt-4o-mini',
            0.5,
            'text',
            'high'
          );
          
          expect(request.data.instructions).toBe('Custom instructions');
          expect(request.data.model).toBe('gpt-4o-mini');
          expect(request.data.temperature).toBe(0.5);
          expect(request.data.modality).toBe('text');
          expect(request.data.reasoning_effort).toBe('high');
        });
      });

      describe('Prepare Evaluate', () => {
        it('should prepare evaluate request with defaults', () => {
          const mixin = (syncClient.schemas as any).mixin;
          const request = mixin.prepareEvaluate(testDocuments, companySchema);
          
          expect(request.method).toBe('POST');
          expect(request.url).toBe('/v1/schemas/evaluate');
          expect(request.data).toBeDefined();
          expect(request.data.documents).toBeDefined();
          expect(request.data.json_schema).toEqual(companySchema);
          expect(request.data.model).toBe('gpt-4o-mini');
          expect(request.data.reasoning_effort).toBe('medium');
          expect(request.data.modality).toBe('native');
          expect(request.data.image_resolution_dpi).toBe(96);
          expect(request.data.browser_canvas).toBe('A4');
          expect(request.data.n_consensus).toBe(1);
        });

        it('should prepare evaluate request with ground truths', () => {
          const groundTruths = [{ name: 'Test Company', type: 'startup' }];
          const mixin = (syncClient.schemas as any).mixin;
          const request = mixin.prepareEvaluate(
            testDocuments, 
            companySchema, 
            groundTruths,
            'gpt-4o',
            'low',
            'text',
            150,
            'A3',
            3
          );
          
          expect(request.data.ground_truths).toEqual(groundTruths);
          expect(request.data.model).toBe('gpt-4o');
          expect(request.data.reasoning_effort).toBe('low');
          expect(request.data.modality).toBe('text');
          expect(request.data.image_resolution_dpi).toBe(150);
          expect(request.data.browser_canvas).toBe('A3');
          expect(request.data.n_consensus).toBe(3);
        });
      });

      describe('Prepare Enhance', () => {
        it('should prepare enhance request with defaults', () => {
          const mixin = (syncClient.schemas as any).mixin;
          const request = mixin.prepareEnhance(companySchema, testDocuments);
          
          expect(request.method).toBe('POST');
          expect(request.url).toBe('/v1/schemas/enhance');
          expect(request.data).toBeDefined();
          expect(request.data.json_schema).toEqual(companySchema);
          expect(request.data.documents).toBeDefined();
          expect(request.data.model).toBe('gpt-4o-mini');
          expect(request.data.temperature).toBe(0.0);
          expect(request.data.modality).toBe('native');
        });

        it('should prepare enhance request with custom parameters', () => {
          const groundTruths = [{ name: 'Test Company' }];
          const instructions = 'Enhance the schema for better extraction';
          const flatLikelihoods = { 'name': 0.95, 'type': 0.8 };
          const toolsConfig = { enhance_descriptions: true };
          
          const mixin = (syncClient.schemas as any).mixin;
          const request = mixin.prepareEnhance(
            companySchema,
            testDocuments,
            groundTruths,
            instructions,
            'gpt-4o',
            0.3,
            'text',
            flatLikelihoods,
            toolsConfig
          );
          
          expect(request.data.ground_truths).toEqual(groundTruths);
          expect(request.data.instructions).toBe(instructions);
          expect(request.data.model).toBe('gpt-4o');
          expect(request.data.temperature).toBe(0.3);
          expect(request.data.modality).toBe('text');
          expect(request.data.flat_likelihoods).toEqual(flatLikelihoods);
          expect(request.data.tools_config).toEqual(toolsConfig);
        });

        it('should handle schema as string', () => {
          const schemaString = JSON.stringify(companySchema);
          const mixin = (syncClient.schemas as any).mixin;
          const request = mixin.prepareEnhance(schemaString, testDocuments);
          
          expect(request.data.json_schema).toEqual(companySchema);
        });
      });

      describe('Prepare Get', () => {
        it('should prepare get request', () => {
          const schemaId = 'schema-123';
          const mixin = (syncClient.schemas as any).mixin;
          const request = mixin.prepareGet(schemaId);
          
          expect(request.method).toBe('GET');
          expect(request.url).toBe(`/v1/schemas/${schemaId}`);
        });
      });
    });
  });

  describe('Files Resource', () => {
    it('should exist and be accessible', () => {
      expect(syncClient.files).toBeDefined();
      expect(asyncClient.files).toBeDefined();
    });
  });

  describe('Models Resource', () => {
    describe('Prepare Methods', () => {
      it('should prepare list request', () => {
        const mixin = (syncClient.models as any).mixin;
        const request = mixin.prepareList();
        
        expect(request.method).toBe('GET');
        expect(request.url).toBe('/v1/models');
      });

      it('should prepare get request', () => {
        const modelId = 'gpt-4o';
        const mixin = (syncClient.models as any).mixin;
        const request = mixin.prepareGet(modelId);
        
        expect(request.method).toBe('GET');
        expect(request.url).toBe(`/v1/models/${modelId}`);
      });
    });
  });

  describe('Usage Resource', () => {
    describe('Prepare Methods', () => {
      it('should prepare get request with defaults', () => {
        const mixin = (syncClient.usage as any).mixin;
        const request = mixin.prepareGet();
        
        expect(request.method).toBe('GET');
        expect(request.url).toBe('/v1/usage');
        expect(request.params).toEqual({
          group_by: 'day',
          start_time: undefined,
          end_time: undefined
        });
      });

      it('should prepare get request with parameters', () => {
        const startTime = '2024-01-01T00:00:00Z';
        const endTime = '2024-01-31T23:59:59Z';
        const mixin = (syncClient.usage as any).mixin;
        const request = mixin.prepareGet('month', startTime, endTime);
        
        expect(request.params).toEqual({
          group_by: 'month',
          start_time: startTime,
          end_time: endTime
        });
      });
    });
  });

  describe('FineTuning Resource', () => {
    describe('Prepare Methods', () => {
      it('should prepare create request', () => {
        const trainingFileId = 'file-123';
        const model = 'gpt-4o-mini';
        const mixin = (syncClient.fineTuning as any).mixin;
        const request = mixin.prepareCreate(trainingFileId, model);
        
        expect(request.method).toBe('POST');
        expect(request.url).toBe('/v1/fine_tuning/jobs');
        expect(request.data).toEqual({
          training_file: trainingFileId,
          model: model,
          hyperparameters: undefined,
          suffix: undefined,
          validation_file: undefined,
          integrations: undefined,
          seed: undefined
        });
      });

      it('should prepare create request with hyperparameters', () => {
        const trainingFileId = 'file-123';
        const model = 'gpt-4o-mini';
        const hyperparameters = { n_epochs: 3, batch_size: 1 };
        const suffix = 'custom-model';
        const validationFile = 'file-456';
        const seed = 42;
        
        const mixin = (syncClient.fineTuning as any).mixin;
        const request = mixin.prepareCreate(
          trainingFileId, 
          model, 
          hyperparameters, 
          suffix, 
          validationFile, 
          undefined, 
          seed
        );
        
        expect(request.data).toEqual({
          training_file: trainingFileId,
          model: model,
          hyperparameters: hyperparameters,
          suffix: suffix,
          validation_file: validationFile,
          integrations: undefined,
          seed: seed
        });
      });

      it('should prepare list request', () => {
        const mixin = (syncClient.fineTuning as any).mixin;
        const request = mixin.prepareList();
        
        expect(request.method).toBe('GET');
        expect(request.url).toBe('/v1/fine_tuning/jobs');
        expect(request.params).toEqual({
          after: undefined,
          limit: 20
        });
      });

      it('should prepare list request with parameters', () => {
        const mixin = (syncClient.fineTuning as any).mixin;
        const request = mixin.prepareList('cursor-123', 50);
        
        expect(request.params).toEqual({
          after: 'cursor-123',
          limit: 50
        });
      });

      it('should prepare get request', () => {
        const jobId = 'job-123';
        const mixin = (syncClient.fineTuning as any).mixin;
        const request = mixin.prepareGet(jobId);
        
        expect(request.method).toBe('GET');
        expect(request.url).toBe(`/v1/fine_tuning/jobs/${jobId}`);
      });

      it('should prepare cancel request', () => {
        const jobId = 'job-123';
        const mixin = (syncClient.fineTuning as any).mixin;
        const request = mixin.prepareCancel(jobId);
        
        expect(request.method).toBe('POST');
        expect(request.url).toBe(`/v1/fine_tuning/jobs/${jobId}/cancel`);
      });
    });
  });

  describe('Secrets Resource', () => {
    describe('External API Keys Methods', () => {
      it('should prepare create external API key request', () => {
        const mixin = (syncClient.secrets.external_api_keys as any).mixin;
        const request = mixin.prepareCreate('OpenAI', 'sk-test-key');
        
        expect(request.method).toBe('POST');
        expect(request.url).toBe('/v1/secrets/external_api_keys');
        expect(request.data).toEqual({
          provider: 'OpenAI',
          api_key: 'sk-test-key'
        });
      });

      it('should prepare list external API keys request', () => {
        const mixin = (syncClient.secrets.external_api_keys as any).mixin;
        const request = mixin.prepareList();
        
        expect(request.method).toBe('GET');
        expect(request.url).toBe('/v1/secrets/external_api_keys');
      });

      it('should prepare get external API key request', () => {
        const keyId = 'key-123';
        const mixin = (syncClient.secrets.external_api_keys as any).mixin;
        const request = mixin.prepareGet(keyId);
        
        expect(request.method).toBe('GET');
        expect(request.url).toBe(`/v1/secrets/external_api_keys/${keyId}`);
      });

      it('should prepare update external API key request', () => {
        const keyId = 'key-123';
        const mixin = (syncClient.secrets.external_api_keys as any).mixin;
        const request = mixin.prepareUpdate(keyId, 'sk-new-key');
        
        expect(request.method).toBe('PATCH');
        expect(request.url).toBe(`/v1/secrets/external_api_keys/${keyId}`);
        expect(request.data).toEqual({
          api_key: 'sk-new-key'
        });
      });

      it('should prepare delete external API key request', () => {
        const keyId = 'key-123';
        const mixin = (syncClient.secrets.external_api_keys as any).mixin;
        const request = mixin.prepareDelete(keyId);
        
        expect(request.method).toBe('DELETE');
        expect(request.url).toBe(`/v1/secrets/external_api_keys/${keyId}`);
      });
    });
  });

  describe('Consensus Resource', () => {
    describe('Prepare Methods', () => {
      const testSchema = companySchema;
      const testDocuments = ['test document'];
      
      it('should prepare create request with defaults', () => {
        const mixin = (syncClient.consensus as any).mixin;
        const request = mixin.prepareCreate(testSchema, testDocuments);
        
        expect(request.method).toBe('POST');
        expect(request.url).toBe('/v1/consensus');
        expect(request.data.json_schema).toEqual(testSchema);
        expect(request.data.documents).toBeDefined();
        expect(request.data.model).toBe('gpt-4o-mini');
        expect(request.data.n_consensus).toBe(3);
        expect(request.data.temperature).toBe(0);
        expect(request.data.reasoning_effort).toBe('medium');
        expect(request.data.modality).toBe('native');
      });

      it('should prepare create request with custom parameters', () => {
        const mixin = (syncClient.consensus as any).mixin;
        const request = mixin.prepareCreate(
          testSchema,
          testDocuments,
          'gpt-4o',
          5,
          0.2,
          'high',
          'text'
        );
        
        expect(request.data.model).toBe('gpt-4o');
        expect(request.data.n_consensus).toBe(5);
        expect(request.data.temperature).toBe(0.2);
        expect(request.data.reasoning_effort).toBe('high');
        expect(request.data.modality).toBe('text');
      });
    });
  });
});

describe('Error Handling in Resources', () => {
  let client: Retab;

  beforeEach(() => {
    client = new Retab({ apiKey: TEST_API_KEY });
  });

  describe('Model Validation', () => {
    it('should validate model for schema generation', () => {
      const mixin = (client.schemas as any).mixin;
      
      // Should not throw for valid models
      expect(() => mixin.prepareGenerate(['test'], undefined, 'gpt-4o-2024-11-20'))
        .not.toThrow();
      expect(() => mixin.prepareGenerate(['test'], undefined, 'gpt-4o-mini'))
        .not.toThrow();
      expect(() => mixin.prepareGenerate(['test'], undefined, 'claude-3-5-sonnet-latest'))
        .not.toThrow();
    });
  });

  describe('Document Preparation', () => {
    it('should handle string documents', () => {
      const mixin = (client.schemas as any).mixin;
      const request = mixin.prepareGenerate(['text document']);
      expect(request.data.documents).toBeDefined();
      expect(Array.isArray(request.data.documents)).toBe(true);
    });

    it('should handle buffer documents', () => {
      const buffer = Buffer.from('test content');
      const mixin = (client.schemas as any).mixin;
      const request = mixin.prepareGenerate([buffer]);
      expect(request.data.documents).toBeDefined();
      expect(Array.isArray(request.data.documents)).toBe(true);
    });

    it('should handle mixed document types', () => {
      const documents = [
        'text document',
        Buffer.from('binary content'),
        { filename: 'test.pdf', content: Buffer.from('pdf'), mimeType: 'application/pdf' }
      ];
      const mixin = (client.schemas as any).mixin;
      const request = mixin.prepareGenerate(documents);
      expect(request.data.documents).toBeDefined();
      expect(Array.isArray(request.data.documents)).toBe(true);
    });
  });
});