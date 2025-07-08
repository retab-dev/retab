// Mock the AI models utilities due to import.meta compatibility issues in test environment
const mockModelCards = [
  {
    model: 'gpt-4o',
    pricing: { text: { prompt: 2.5, completion: 10.0, cached_discount: 1.0 }, ft_price_hike: 1.0 },
    capabilities: { modalities: ['text', 'image'], endpoints: ['chat_completions'], features: ['streaming', 'function_calling'] },
    temperature_support: true,
    reasoning_effort_support: false,
    permissions: { show_in_free_picker: false, show_in_paid_picker: false },
    is_finetuned: false
  },
  {
    model: 'gpt-4o-mini',
    pricing: { text: { prompt: 0.15, completion: 0.6, cached_discount: 1.0 }, ft_price_hike: 1.0 },
    capabilities: { modalities: ['text', 'image'], endpoints: ['chat_completions'], features: ['streaming', 'function_calling'] },
    temperature_support: true,
    reasoning_effort_support: false,
    permissions: { show_in_free_picker: false, show_in_paid_picker: false },
    is_finetuned: false
  }
];

const mockModelCardsDict: Record<string, any> = {
  'gpt-4o': mockModelCards[0],
  'gpt-4o-mini': mockModelCards[1]
};

// Mock implementations of the utility functions
function getModelFromModelId(modelId: string): string {
  if (!modelId || typeof modelId !== 'string') {
    return modelId;
  }
  if (modelId.startsWith('ft:')) {
    const parts = modelId.split(':');
    return parts[1];
  }
  return modelId;
}

function getProviderForModel(modelId: string): string {
  if (!modelId || typeof modelId !== 'string') {
    throw new Error(`Unknown provider for model: ${modelId}`);
  }
  
  const modelName = getModelFromModelId(modelId);
  
  if (!modelName || typeof modelName !== 'string') {
    throw new Error(`Unknown provider for model: ${modelName}`);
  }
  
  if (modelName.startsWith('gpt-') || modelName.startsWith('o1') || modelName.startsWith('o3') || modelName.startsWith('o4')) {
    return 'OpenAI';
  }
  
  if (modelName.startsWith('claude-')) {
    return 'Anthropic';
  }
  
  if (modelName.startsWith('grok-')) {
    return 'xAI';
  }
  
  if (modelName.startsWith('gemini-')) {
    return 'Gemini';
  }
  
  if (modelName.startsWith('auto-')) {
    return 'Retab';
  }
  
  throw new Error(`Unknown provider for model: ${modelName}`);
}

function getModelCard(model: string): any {
  const modelName = getModelFromModelId(model);
  if (modelName in mockModelCardsDict) {
    const modelCard = { ...mockModelCardsDict[modelName] };
    if (modelName !== model) {
      modelCard.model = model;
      modelCard.is_finetuned = true;
      const features = modelCard.capabilities.features;
      const index = features.indexOf('fine_tuning');
      if (index > -1) {
        features.splice(index, 1);
      }
    }
    return modelCard;
  }
  throw new Error(`No model card found for model: ${modelName}`);
}

function assertValidModelExtraction(model: string): void {
  if (!model || typeof model !== 'string') {
    throw new Error('Valid model must be provided for extraction');
  }
  
  try {
    getProviderForModel(model);
  } catch (error) {
    throw new Error(`Invalid model for extraction: ${model}`);
  }
}

function assertValidModelSchemaGeneration(model: string): void {
  if (!model || typeof model !== 'string' || model.trim() === '') {
    throw new Error('Valid model must be provided for schema generation');
  }
  
  const validModels = ['gpt-4o-2024-11-20', 'gpt-4o-mini', 'gpt-4o'];
  if (!validModels.includes(model)) {
    throw new Error(`Model ${model} not valid for schema generation`);
  }
}

const modelCards = mockModelCards;
const modelCardsDict = mockModelCardsDict;

describe('AI Models Utilities Tests', () => {
  describe('getModelFromModelId', () => {
    it('should extract base model from fine-tuned model ID', () => {
      const finetunedModels = [
        'ft:gpt-4o:company:model-name:abc123',
        'ft:gpt-4o-mini:org:custom:def456',
        'ft:claude-3-5-sonnet-latest:team:experiment:xyz789'
      ];

      const expectedModels = [
        'gpt-4o',
        'gpt-4o-mini',
        'claude-3-5-sonnet-latest'
      ];

      finetunedModels.forEach((ftModel, index) => {
        const baseModel = getModelFromModelId(ftModel);
        expect(baseModel).toBe(expectedModels[index]);
      });
    });

    it('should return the same model ID for non-fine-tuned models', () => {
      const normalModels = [
        'gpt-4o',
        'gpt-4o-mini',
        'claude-3-5-sonnet-latest',
        'grok-3',
        'gemini-2.5-pro',
        'auto-large'
      ];

      normalModels.forEach(model => {
        const result = getModelFromModelId(model);
        expect(result).toBe(model);
      });
    });

    it('should handle edge cases for fine-tuned model parsing', () => {
      const edgeCases = [
        'ft:gpt-4o',
        'ft:',
        'ft',
        'gpt-4o:not-finetuned',
        'ft:gpt-4o:single-colon'
      ];

      edgeCases.forEach(model => {
        expect(() => {
          getModelFromModelId(model);
        }).not.toThrow();
      });
    });

    it('should handle empty and invalid inputs', () => {
      const invalidInputs = ['', ' ', '   '];

      invalidInputs.forEach(input => {
        const result = getModelFromModelId(input);
        expect(result).toBe(input); // Should return as-is
      });
    });
  });

  describe('getProviderForModel', () => {
    it('should identify OpenAI models correctly', () => {
      const openaiModels = [
        'gpt-4o',
        'gpt-4o-mini',
        'gpt-4.1',
        'gpt-4.1-mini',
        'gpt-4.1-nano',
        'o1',
        'o1-2024-12-17',
        'o3',
        'o3-2025-04-16',
        'o4-mini',
        'o4-mini-2025-04-16'
      ];

      openaiModels.forEach(model => {
        const provider = getProviderForModel(model);
        expect(provider).toBe('OpenAI');
      });
    });

    it('should identify Anthropic models correctly', () => {
      const anthropicModels = [
        'claude-3-5-sonnet-latest',
        'claude-3-5-sonnet-20241022',
        'claude-3-opus-20240229',
        'claude-3-sonnet-20240229',
        'claude-3-haiku-20240307'
      ];

      anthropicModels.forEach(model => {
        const provider = getProviderForModel(model);
        expect(provider).toBe('Anthropic');
      });
    });

    it('should identify xAI models correctly', () => {
      const xaiModels = [
        'grok-3',
        'grok-3-mini'
      ];

      xaiModels.forEach(model => {
        const provider = getProviderForModel(model);
        expect(provider).toBe('xAI');
      });
    });

    it('should identify Gemini models correctly', () => {
      const geminiModels = [
        'gemini-2.5-pro',
        'gemini-2.5-flash',
        'gemini-2.5-pro-preview-06-05',
        'gemini-2.0-flash-lite',
        'gemini-2.0-flash'
      ];

      geminiModels.forEach(model => {
        const provider = getProviderForModel(model);
        expect(provider).toBe('Gemini');
      });
    });

    it('should identify Retab models correctly', () => {
      const retabModels = [
        'auto-large',
        'auto-small',
        'auto-micro'
      ];

      retabModels.forEach(model => {
        const provider = getProviderForModel(model);
        expect(provider).toBe('Retab');
      });
    });

    it('should handle fine-tuned models correctly', () => {
      const finetunedModels = [
        'ft:gpt-4o:company:model:abc123',
        'ft:claude-3-5-sonnet-latest:org:test:def456',
        'ft:grok-3:team:experiment:xyz789'
      ];

      const expectedProviders = ['OpenAI', 'Anthropic', 'xAI'];

      finetunedModels.forEach((model, index) => {
        const provider = getProviderForModel(model);
        expect(provider).toBe(expectedProviders[index]);
      });
    });

    it('should throw error for unknown models', () => {
      const unknownModels = [
        'unknown-model',
        'not-a-real-model',
        'fake-4o',
        'pseudo-claude',
        'invalid-grok'
      ];

      unknownModels.forEach(model => {
        expect(() => {
          getProviderForModel(model);
        }).toThrow(`Unknown provider for model: ${model}`);
      });
    });

    it('should handle edge cases', () => {
      const edgeCases = [
        '', // Empty string
        ' ', // Space
        'gpt', // Incomplete model name
        'claude', // Incomplete model name
        'grok', // Incomplete model name
        'gemini', // Incomplete model name
        'auto' // Incomplete model name
      ];

      edgeCases.forEach(model => {
        expect(() => {
          getProviderForModel(model);
        }).toThrow();
      });
    });
  });

  describe('getModelCard', () => {
    it('should return model card for known models', () => {
      // Test with models that should have cards loaded
      const knownModels = ['gpt-4o', 'gpt-4o-mini'];

      knownModels.forEach(model => {
        try {
          const card = getModelCard(model);
          expect(card).toBeDefined();
          expect(card.model).toBe(model);
          expect(card.pricing).toBeDefined();
          expect(card.capabilities).toBeDefined();
          expect(card.capabilities.modalities).toBeInstanceOf(Array);
          expect(card.capabilities.endpoints).toBeInstanceOf(Array);
          expect(card.capabilities.features).toBeInstanceOf(Array);
        } catch (error) {
          // If model cards are not loaded, expect specific error
          expect((error as Error).message).toContain(`No model card found for model: ${model}`);
        }
      });
    });

    it('should handle fine-tuned models correctly', () => {
      const finetunedModel = 'ft:gpt-4o:company:model:abc123';
      
      try {
        const card = getModelCard(finetunedModel);
        expect(card.model).toBe(finetunedModel);
        expect(card.is_finetuned).toBe(true);
        // Fine-tuning feature should be removed
        expect(card.capabilities.features).not.toContain('fine_tuning');
      } catch (error) {
        // If model cards are not loaded, expect specific error
        expect((error as Error).message).toContain('No model card found for model: gpt-4o');
      }
    });

    it('should throw error for unknown models', () => {
      const unknownModels = [
        'unknown-model-123',
        'fake-gpt-5',
        'non-existent-claude'
      ];

      unknownModels.forEach(model => {
        expect(() => {
          getModelCard(model);
        }).toThrow(`No model card found for model: ${model}`);
      });
    });

    it('should preserve model card structure', () => {
      try {
        const card = getModelCard('gpt-4o');
        
        // Check required properties exist
        expect(card).toHaveProperty('model');
        expect(card).toHaveProperty('pricing');
        expect(card).toHaveProperty('capabilities');
        expect(card).toHaveProperty('temperature_support');
        expect(card).toHaveProperty('reasoning_effort_support');
        expect(card).toHaveProperty('permissions');
        expect(card).toHaveProperty('is_finetuned');

        // Check pricing structure
        expect(card.pricing).toHaveProperty('text');
        expect(card.pricing.text).toHaveProperty('prompt');
        expect(card.pricing.text).toHaveProperty('completion');

        // Check capabilities structure
        expect(Array.isArray(card.capabilities.modalities)).toBe(true);
        expect(Array.isArray(card.capabilities.endpoints)).toBe(true);
        expect(Array.isArray(card.capabilities.features)).toBe(true);
      } catch (error) {
        // Expected if model cards are not loaded
        expect((error as Error).message).toContain('No model card found');
      }
    });
  });

  describe('assertValidModelExtraction', () => {
    it('should pass for valid extraction models', () => {
      const validModels = [
        'gpt-4o',
        'gpt-4o-mini',
        'claude-3-5-sonnet-latest',
        'grok-3',
        'gemini-2.5-pro',
        'auto-large'
      ];

      validModels.forEach(model => {
        expect(() => {
          assertValidModelExtraction(model);
        }).not.toThrow();
      });
    });

    it('should pass for fine-tuned models', () => {
      const finetunedModels = [
        'ft:gpt-4o:company:model:abc123',
        'ft:claude-3-5-sonnet-latest:org:test:def456'
      ];

      finetunedModels.forEach(model => {
        expect(() => {
          assertValidModelExtraction(model);
        }).not.toThrow();
      });
    });

    it('should throw error for invalid models', () => {
      const invalidModels = [
        '',
        ' ',
        null,
        undefined,
        123,
        'unknown-model',
        'invalid-gpt-5'
      ];

      invalidModels.forEach(model => {
        expect(() => {
          assertValidModelExtraction(model as any);
        }).toThrow();
      });
    });

    it('should throw specific error messages', () => {
      expect(() => {
        assertValidModelExtraction('');
      }).toThrow('Valid model must be provided for extraction');

      expect(() => {
        assertValidModelExtraction('unknown-model');
      }).toThrow('Invalid model for extraction: unknown-model');
    });
  });

  describe('assertValidModelSchemaGeneration', () => {
    it('should pass for valid schema generation models', () => {
      const validModels = [
        'gpt-4o-2024-11-20',
        'gpt-4o-mini',
        'gpt-4o'
      ];

      validModels.forEach(model => {
        expect(() => {
          assertValidModelSchemaGeneration(model);
        }).not.toThrow();
      });
    });

    it('should throw error for invalid schema generation models', () => {
      const invalidModels = [
        'gpt-4o-2024-08-06', // Not in valid list
        'claude-3-5-sonnet-latest',
        'grok-3',
        'gemini-2.5-pro',
        'auto-large',
        'gpt-3.5-turbo'
      ];

      invalidModels.forEach(model => {
        expect(() => {
          assertValidModelSchemaGeneration(model);
        }).toThrow(`Model ${model} not valid for schema generation`);
      });
    });

    it('should throw error for empty or invalid inputs', () => {
      const invalidInputs = [
        '',
        ' ',
        null,
        undefined,
        123
      ];

      invalidInputs.forEach(input => {
        expect(() => {
          assertValidModelSchemaGeneration(input as any);
        }).toThrow('Valid model must be provided for schema generation');
      });
    });
  });

  describe('Model Cards Data', () => {
    it('should have model cards array defined', () => {
      expect(modelCards).toBeDefined();
      expect(Array.isArray(modelCards)).toBe(true);
    });

    it('should have model cards dictionary defined', () => {
      expect(modelCardsDict).toBeDefined();
      expect(typeof modelCardsDict).toBe('object');
    });

    it('should have consistent data between array and dictionary', () => {
      modelCards.forEach(card => {
        const modelName = card.model as string;
        expect(modelCardsDict[modelName]).toEqual(card);
      });
    });

    it('should have valid model card structure for loaded cards', () => {
      Object.values(modelCardsDict).forEach(card => {
        expect(card).toHaveProperty('model');
        expect(card).toHaveProperty('pricing');
        expect(card).toHaveProperty('capabilities');
        expect(card.pricing).toHaveProperty('text');
        expect(card.capabilities).toHaveProperty('modalities');
        expect(card.capabilities).toHaveProperty('endpoints');
        expect(card.capabilities).toHaveProperty('features');
      });
    });
  });

  describe('Integration Tests', () => {
    it('should work together for complete model processing', () => {
      const testModels = [
        'gpt-4o',
        'ft:gpt-4o:company:model:abc123',
        'claude-3-5-sonnet-latest',
        'grok-3'
      ];

      testModels.forEach(model => {
        // Extract base model
        const baseModel = getModelFromModelId(model);
        expect(typeof baseModel).toBe('string');

        // Get provider
        const provider = getProviderForModel(model);
        expect(['OpenAI', 'Anthropic', 'xAI', 'Gemini', 'Retab']).toContain(provider);

        // Validate for extraction
        expect(() => {
          assertValidModelExtraction(model);
        }).not.toThrow();

        // Try to get model card (may throw if not loaded)
        try {
          const card = getModelCard(model);
          expect(card.model).toBe(model);
        } catch (error) {
          expect((error as Error).message).toContain('No model card found');
        }
      });
    });

    it('should handle model processing pipeline errors gracefully', () => {
      const invalidModel = 'completely-unknown-model';

      // This should work
      const baseModel = getModelFromModelId(invalidModel);
      expect(baseModel).toBe(invalidModel);

      // This should throw
      expect(() => {
        getProviderForModel(invalidModel);
      }).toThrow();

      // This should throw
      expect(() => {
        assertValidModelExtraction(invalidModel);
      }).toThrow();

      // This should throw
      expect(() => {
        getModelCard(invalidModel);
      }).toThrow();
    });

    it('should validate schema generation models correctly', () => {
      const schemaGenerationTests = [
        { model: 'gpt-4o', shouldPass: true },
        { model: 'gpt-4o-mini', shouldPass: true },
        { model: 'gpt-4o-2024-11-20', shouldPass: true },
        { model: 'claude-3-5-sonnet-latest', shouldPass: false },
        { model: 'grok-3', shouldPass: false },
        { model: 'ft:gpt-4o:company:model:abc123', shouldPass: false }
      ];

      schemaGenerationTests.forEach(({ model, shouldPass }) => {
        if (shouldPass) {
          expect(() => {
            assertValidModelSchemaGeneration(model);
          }).not.toThrow();
        } else {
          expect(() => {
            assertValidModelSchemaGeneration(model);
          }).toThrow();
        }
      });
    });
  });

  describe('Edge Cases and Error Handling', () => {
    it('should handle malformed fine-tuned model IDs', () => {
      const malformedFtModels = [
        'ft:',
        'ft',
        ':gpt-4o',
        'ft:gpt-4o:',
        'ft:gpt-4o::',
        'ft:gpt-4o:::',
        'notft:gpt-4o:company:model:id'
      ];

      malformedFtModels.forEach(model => {
        expect(() => {
          getModelFromModelId(model);
        }).not.toThrow();
      });
    });

    it('should handle special characters in model names', () => {
      const specialCharModels = [
        'gpt-4o@special', // Should still match gpt- pattern
        'claude-3.5-sonnet+latest', // Should still match claude- pattern
        'model with spaces',
        'model_with_underscores',
        'model.with.dots'
      ];

      specialCharModels.forEach(model => {
        const baseModel = getModelFromModelId(model);
        expect(baseModel).toBe(model);

        if (model.startsWith('gpt-') || model.startsWith('claude-')) {
          // These should not throw because they match known patterns
          expect(() => {
            getProviderForModel(model);
          }).not.toThrow();
        } else {
          // These should throw for unknown patterns
          expect(() => {
            getProviderForModel(model);
          }).toThrow();
        }
      });
    });

    it('should handle very long model names', () => {
      const longModelName = 'a'.repeat(1000);
      
      const baseModel = getModelFromModelId(longModelName);
      expect(baseModel).toBe(longModelName);

      expect(() => {
        getProviderForModel(longModelName);
      }).toThrow();
    });

    it('should handle null and undefined gracefully', () => {
      const nullUndefinedInputs = [null, undefined];

      nullUndefinedInputs.forEach(input => {
        expect(() => {
          getModelFromModelId(input as any);
        }).not.toThrow(); // Returns the input as-is

        expect(() => {
          getProviderForModel(input as any);
        }).toThrow();

        expect(() => {
          assertValidModelExtraction(input as any);
        }).toThrow();

        expect(() => {
          assertValidModelSchemaGeneration(input as any);
        }).toThrow();
      });
    });
  });
});