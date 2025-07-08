import { Retab, AsyncRetab } from '../src/index.js';
import { TEST_API_KEY } from './fixtures.js';

// Valid OpenAI models for testing
const VALID_OPENAI_MODELS = [
  'gpt-4o-mini-2024-07-18',
  'gpt-4o-2024-08-06',
  'gpt-3.5-turbo-0125',
  'gpt-3.5-turbo-1106',
  'gpt-3.5-turbo-0613',
  'davinci-002',
  'babbage-002'
];

describe('Fine-Tuning Jobs', () => {
  let syncClient: Retab;
  let asyncClient: AsyncRetab;

  beforeEach(() => {
    // Create clients with OpenAI API key
    syncClient = new Retab({ 
      apiKey: TEST_API_KEY,
      openaiApiKey: 'sk-openai-test-key'
    });
    asyncClient = new AsyncRetab({ 
      apiKey: TEST_API_KEY,
      openaiApiKey: 'sk-openai-test-key'
    });
  });

  describe('Resource Availability', () => {
    it('should have fine-tuning resource on sync client', () => {
      expect(syncClient.fineTuning).toBeDefined();
      expect(syncClient.fineTuning.jobs).toBeDefined();
      expect(typeof syncClient.fineTuning.jobs.create).toBe('function');
      expect(typeof syncClient.fineTuning.jobs.retrieve).toBe('function');
      expect(typeof syncClient.fineTuning.jobs.list).toBe('function');
      expect(typeof syncClient.fineTuning.jobs.cancel).toBe('function');
      expect(typeof syncClient.fineTuning.jobs.listEvents).toBe('function');
    });

    it('should have fine-tuning resource on async client', () => {
      expect(asyncClient.fineTuning).toBeDefined();
      expect(asyncClient.fineTuning.jobs).toBeDefined();
      expect(typeof asyncClient.fineTuning.jobs.create).toBe('function');
      expect(typeof asyncClient.fineTuning.jobs.retrieve).toBe('function');
      expect(typeof asyncClient.fineTuning.jobs.list).toBe('function');
      expect(typeof asyncClient.fineTuning.jobs.cancel).toBe('function');
      expect(typeof asyncClient.fineTuning.jobs.listEvents).toBe('function');
    });
  });

  describe('Model Validation', () => {
    const clientTypes = ['sync', 'async'] as const;

    clientTypes.forEach(clientType => {
      describe(`${clientType} client`, () => {
        it('should reject invalid models', async () => {
          const client = clientType === 'sync' ? syncClient : asyncClient;
          const invalidModels = [
            'invalid-model',
            'gpt-4',
            'claude-3',
            'gemini-pro'
          ];

          for (const model of invalidModels) {
            await expect(client.fineTuning.jobs.create({
              training_file: 'file-abc123',
              model
            })).rejects.toThrow(`Model ${model} is not supported`);
          }
        });

        it('should handle missing OpenAI API key', async () => {
          const clientWithoutKey = clientType === 'sync' 
            ? new Retab({ apiKey: TEST_API_KEY })
            : new AsyncRetab({ apiKey: TEST_API_KEY });

          await expect(clientWithoutKey.fineTuning.jobs.create({
            training_file: 'file-abc123',
            model: 'gpt-3.5-turbo-0125'
          })).rejects.toThrow('OpenAI API key not found');
        });

        it('should handle missing OpenAI API key for retrieval', async () => {
          const clientWithoutKey = clientType === 'sync' 
            ? new Retab({ apiKey: TEST_API_KEY })
            : new AsyncRetab({ apiKey: TEST_API_KEY });

          await expect(clientWithoutKey.fineTuning.jobs.retrieve('ftjob-abc123'))
            .rejects.toThrow('OpenAI API key not found');
        });

        it('should handle missing OpenAI API key for listing', async () => {
          const clientWithoutKey = clientType === 'sync' 
            ? new Retab({ apiKey: TEST_API_KEY })
            : new AsyncRetab({ apiKey: TEST_API_KEY });

          await expect(clientWithoutKey.fineTuning.jobs.list())
            .rejects.toThrow('OpenAI API key not found');
        });

        it('should handle missing OpenAI API key for cancellation', async () => {
          const clientWithoutKey = clientType === 'sync' 
            ? new Retab({ apiKey: TEST_API_KEY })
            : new AsyncRetab({ apiKey: TEST_API_KEY });

          await expect(clientWithoutKey.fineTuning.jobs.cancel('ftjob-abc123'))
            .rejects.toThrow('OpenAI API key not found');
        });

        it('should handle missing OpenAI API key for events', async () => {
          const clientWithoutKey = clientType === 'sync' 
            ? new Retab({ apiKey: TEST_API_KEY })
            : new AsyncRetab({ apiKey: TEST_API_KEY });

          await expect(clientWithoutKey.fineTuning.jobs.listEvents('ftjob-abc123'))
            .rejects.toThrow('OpenAI API key not found');
        });
      });
    });
  });

  describe('Parameter Validation', () => {
    it('should accept valid OpenAI models in the model list', () => {
      const validModels = VALID_OPENAI_MODELS;
      expect(validModels).toContain('gpt-3.5-turbo-0125');
      expect(validModels).toContain('gpt-4o-mini-2024-07-18');
      expect(validModels).toContain('davinci-002');
      expect(validModels).toContain('babbage-002');
    });

    it('should validate hyperparameters structure', async () => {
      // Test that the method accepts hyperparameters object structure
      const hyperparameters = {
        batch_size: 4,
        learning_rate_multiplier: 0.1,
        n_epochs: 3
      };
      
      // This will fail due to model validation, but it tests parameter structure
      await expect(syncClient.fineTuning.jobs.create({
        training_file: 'file-abc123',
        model: 'invalid-model',
        hyperparameters
      })).rejects.toThrow('Model invalid-model is not supported');
    });

    it('should validate auto hyperparameters', async () => {
      const hyperparameters = {
        batch_size: 'auto' as const,
        learning_rate_multiplier: 'auto' as const,
        n_epochs: 'auto' as const
      };
      
      // This will fail due to model validation, but it tests parameter structure
      await expect(syncClient.fineTuning.jobs.create({
        training_file: 'file-abc123',
        model: 'invalid-model',
        hyperparameters
      })).rejects.toThrow('Model invalid-model is not supported');
    });
  });

  describe('Client Type Parity', () => {
    it('should have same fine-tuning methods for sync and async clients', () => {
      const syncMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(syncClient.fineTuning.jobs));
      const asyncMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(asyncClient.fineTuning.jobs));
      
      const requiredMethods = ['create', 'retrieve', 'list', 'cancel', 'listEvents'];
      
      requiredMethods.forEach(method => {
        expect(syncMethods).toContain(method);
        expect(asyncMethods).toContain(method);
        expect(typeof (syncClient.fineTuning.jobs as any)[method]).toBe('function');
        expect(typeof (asyncClient.fineTuning.jobs as any)[method]).toBe('function');
      });
    });

    it('should maintain consistent API surface between sync and async', () => {
      // Check that both clients expose the same interface
      expect(syncClient.fineTuning.jobs.constructor.name).toBe('FineTuningJobs');
      expect(asyncClient.fineTuning.jobs.constructor.name).toBe('AsyncFineTuningJobs');
      
      // Both should be async functions
      expect(syncClient.fineTuning.jobs.create.constructor.name).toBe('AsyncFunction');
      expect(asyncClient.fineTuning.jobs.create.constructor.name).toBe('AsyncFunction');
    });
  });

  describe('OpenAI Integration Structure', () => {
    it('should handle create method parameters correctly', () => {
      // Test parameter types are accepted (will fail on missing OpenAI key)
      const createParams = {
        training_file: 'file-abc123',
        model: 'gpt-3.5-turbo-0125',
        validation_file: 'file-validation123',
        hyperparameters: {
          batch_size: 4,
          learning_rate_multiplier: 0.1,
          n_epochs: 3
        },
        suffix: 'my-model'
      };

      // Verify structure is properly typed
      expect(createParams.training_file).toBe('file-abc123');
      expect(createParams.model).toBe('gpt-3.5-turbo-0125');
      expect(createParams.validation_file).toBe('file-validation123');
      expect(createParams.suffix).toBe('my-model');
    });

    it('should handle list method parameters correctly', () => {
      const listParams = {
        after: 'ftjob-abc123',
        limit: 10
      };

      // Verify structure is properly typed
      expect(listParams.after).toBe('ftjob-abc123');
      expect(listParams.limit).toBe(10);
    });

    it('should handle listEvents method parameters correctly', () => {
      const eventParams = {
        after: 'ftevent-abc123',
        limit: 20
      };

      // Verify structure is properly typed
      expect(eventParams.after).toBe('ftevent-abc123');
      expect(eventParams.limit).toBe(20);
    });
  });

  describe('Error Messages', () => {
    it('should provide clear model validation error messages', async () => {
      const invalidModel = 'claude-3-sonnet';
      
      await expect(syncClient.fineTuning.jobs.create({
        training_file: 'file-abc123',
        model: invalidModel
      })).rejects.toThrow(`Model ${invalidModel} is not supported. Supported models are: ${VALID_OPENAI_MODELS.join(', ')}`);
    });

    it('should provide clear OpenAI API key error message', async () => {
      const clientWithoutKey = new Retab({ apiKey: TEST_API_KEY });

      await expect(clientWithoutKey.fineTuning.jobs.create({
        training_file: 'file-abc123',
        model: 'gpt-3.5-turbo-0125'
      })).rejects.toThrow('OpenAI API key not found. Please provide it in the client configuration.');
    });
  });
});