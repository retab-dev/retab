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
    it('should exist and be accessible', () => {
      expect(syncClient.models).toBeDefined();
      expect(asyncClient.models).toBeDefined();
    });
  });

  describe('Usage Resource', () => {
    it('should exist and be accessible', () => {
      expect(syncClient.usage).toBeDefined();
      expect(asyncClient.usage).toBeDefined();
    });
  });

  describe('FineTuning Resource', () => {
    it('should exist and be accessible', () => {
      expect(syncClient.fineTuning).toBeDefined();
      expect(asyncClient.fineTuning).toBeDefined();
    });
  });

  describe('Secrets Resource', () => {
    it('should exist and be accessible', () => {
      expect(syncClient.secrets).toBeDefined();
      expect(asyncClient.secrets).toBeDefined();
      expect(syncClient.secrets.external_api_keys).toBeDefined();
      expect(asyncClient.secrets.external_api_keys).toBeDefined();
    });
  });

  describe('Consensus Resource', () => {
    it('should exist and be accessible', () => {
      expect(syncClient.consensus).toBeDefined();
      expect(asyncClient.consensus).toBeDefined();
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
      expect(() => mixin.prepareGenerate(['test'], undefined, 'gpt-4o'))
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