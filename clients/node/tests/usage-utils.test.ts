import {
  computeApiCallCost,
  computeCostBreakdown,
  CompletionUsage,
  Pricing,
  TextPricing,
  AudioPricing,
  Amount,
  CostBreakdown
} from '../src/utils/usage.js';

describe('Usage Utilities Tests', () => {
  // Test data fixtures
  const basicTextPricing: TextPricing = {
    prompt: 5.0,      // $5 per 1M tokens
    completion: 15.0, // $15 per 1M tokens
    cached_discount: 0.5 // 50% discount for cached tokens
  };

  const basicAudioPricing: AudioPricing = {
    prompt: 100.0,     // $100 per 1M tokens
    completion: 200.0  // $200 per 1M tokens
  };

  const textOnlyPricing: Pricing = {
    text: basicTextPricing
  };

  const textWithAudioPricing: Pricing = {
    text: basicTextPricing,
    audio: basicAudioPricing
  };

  const finetunedPricing: Pricing = {
    text: basicTextPricing,
    audio: basicAudioPricing,
    ft_price_hike: 3.0 // 3x price increase for fine-tuned models
  };

  const basicUsage: CompletionUsage = {
    prompt_tokens: 1000,
    completion_tokens: 500,
    total_tokens: 1500
  };

  const usageWithCached: CompletionUsage = {
    prompt_tokens: 1000,
    completion_tokens: 500,
    total_tokens: 1500,
    prompt_tokens_details: {
      cached_tokens: 200
    }
  };

  const usageWithAudio: CompletionUsage = {
    prompt_tokens: 1000,
    completion_tokens: 500,
    total_tokens: 1500,
    prompt_tokens_details: {
      audio_tokens: 100,
      cached_tokens: 50
    },
    completion_tokens_details: {
      audio_tokens: 75
    }
  };


  describe('computeApiCallCost', () => {
    describe('Basic Text-Only Calculations', () => {
      it('should calculate cost for basic text usage', () => {
        const result = computeApiCallCost(textOnlyPricing, basicUsage);

        // Expected: (1000 * 5 + 500 * 15) / 1e6 = 12500 / 1e6 = 0.0125
        expect(result.value).toBeCloseTo(0.0125, 6);
        expect(result.currency).toBe('USD');
      });

      it('should handle zero token usage', () => {
        const zeroUsage: CompletionUsage = {
          prompt_tokens: 0,
          completion_tokens: 0,
          total_tokens: 0
        };

        const result = computeApiCallCost(textOnlyPricing, zeroUsage);

        expect(result.value).toBe(0);
        expect(result.currency).toBe('USD');
      });

      it('should handle large token counts', () => {
        const largeUsage: CompletionUsage = {
          prompt_tokens: 1000000,  // 1M tokens
          completion_tokens: 500000, // 500K tokens
          total_tokens: 1500000
        };

        const result = computeApiCallCost(textOnlyPricing, largeUsage);

        // Expected: (1000000 * 5 + 500000 * 15) / 1e6 = 12500000 / 1e6 = 12.5
        expect(result.value).toBeCloseTo(12.5, 6);
        expect(result.currency).toBe('USD');
      });

      it('should handle partial token counts', () => {
        const partialUsage: CompletionUsage = {
          prompt_tokens: 123,
          completion_tokens: 456,
          total_tokens: 579
        };

        const result = computeApiCallCost(textOnlyPricing, partialUsage);

        // Expected: (123 * 5 + 456 * 15) / 1e6 = 7455 / 1e6 = 0.007455
        expect(result.value).toBeCloseTo(0.007455, 6);
        expect(result.currency).toBe('USD');
      });
    });

    describe('Cached Token Calculations', () => {
      it('should apply cached token discount correctly', () => {
        const result = computeApiCallCost(textOnlyPricing, usageWithCached);

        // Breakdown:
        // - Regular prompt tokens: (1000 - 200) * 5 = 800 * 5 = 4000
        // - Cached prompt tokens: 200 * (5 * 0.5) = 200 * 2.5 = 500
        // - Completion tokens: 500 * 15 = 7500
        // - Total: (4000 + 500 + 7500) / 1e6 = 12000 / 1e6 = 0.012
        expect(result.value).toBeCloseTo(0.012, 6);
        expect(result.currency).toBe('USD');
      });

      it('should handle all tokens being cached', () => {
        const allCachedUsage: CompletionUsage = {
          prompt_tokens: 1000,
          completion_tokens: 500,
          total_tokens: 1500,
          prompt_tokens_details: {
            cached_tokens: 1000 // All prompt tokens are cached
          }
        };

        const result = computeApiCallCost(textOnlyPricing, allCachedUsage);

        // Breakdown:
        // - Regular prompt tokens: 0 * 5 = 0
        // - Cached prompt tokens: 1000 * (5 * 0.5) = 1000 * 2.5 = 2500
        // - Completion tokens: 500 * 15 = 7500
        // - Total: (0 + 2500 + 7500) / 1e6 = 10000 / 1e6 = 0.01
        expect(result.value).toBeCloseTo(0.01, 6);
      });

      it('should handle zero cached tokens', () => {
        const noCachedUsage: CompletionUsage = {
          prompt_tokens: 1000,
          completion_tokens: 500,
          total_tokens: 1500,
          prompt_tokens_details: {
            cached_tokens: 0
          }
        };

        const result = computeApiCallCost(textOnlyPricing, noCachedUsage);

        // Should be same as basic usage
        expect(result.value).toBeCloseTo(0.0125, 6);
      });
    });

    describe('Audio Token Calculations', () => {
      it('should calculate cost including audio tokens', () => {
        const result = computeApiCallCost(textWithAudioPricing, usageWithAudio);

        // Breakdown:
        // Text tokens:
        // - Regular prompt: (1000 - 50 - 100) * 5 = 850 * 5 = 4250
        // - Cached prompt: 50 * (5 * 0.5) = 50 * 2.5 = 125
        // - Regular completion: (500 - 75) * 15 = 425 * 15 = 6375
        // Audio tokens:
        // - Audio prompt: 100 * 100 = 10000
        // - Audio completion: 75 * 200 = 15000
        // Total: (4250 + 125 + 6375 + 10000 + 15000) / 1e6 = 35750 / 1e6 = 0.035750
        expect(result.value).toBeCloseTo(0.035750, 6);
        expect(result.currency).toBe('USD');
      });

      it('should handle audio-only usage', () => {
        const audioOnlyUsage: CompletionUsage = {
          prompt_tokens: 100,
          completion_tokens: 75,
          total_tokens: 175,
          prompt_tokens_details: {
            audio_tokens: 100
          },
          completion_tokens_details: {
            audio_tokens: 75
          }
        };

        const result = computeApiCallCost(textWithAudioPricing, audioOnlyUsage);

        // Breakdown:
        // Text tokens: 0 (all tokens are audio)
        // Audio tokens: (100 * 100 + 75 * 200) / 1e6 = 25000 / 1e6 = 0.025
        expect(result.value).toBeCloseTo(0.025, 6);
      });

      it('should handle missing audio pricing gracefully', () => {
        // Using text-only pricing with audio tokens should ignore audio tokens
        const result = computeApiCallCost(textOnlyPricing, usageWithAudio);

        // Should only calculate text token costs, ignoring audio tokens
        // Text tokens: (850 * 5 + 50 * 2.5 + 425 * 15) / 1e6 = 10750 / 1e6 = 0.010750
        expect(result.value).toBeCloseTo(0.010750, 6);
      });
    });

    describe('Fine-Tuning Cost Calculations', () => {
      it('should apply fine-tuning price hike', () => {
        const result = computeApiCallCost(finetunedPricing, basicUsage, true);

        // Basic cost: 0.0125, with 3x multiplier: 0.0125 * 3 = 0.0375
        expect(result.value).toBeCloseTo(0.0375, 6);
        expect(result.currency).toBe('USD');
      });

      it('should not apply fine-tuning price hike when isFt is false', () => {
        const result = computeApiCallCost(finetunedPricing, basicUsage, false);

        // Should be same as basic cost without multiplier
        expect(result.value).toBeCloseTo(0.0125, 6);
      });

      it('should handle fine-tuning with audio tokens', () => {
        const result = computeApiCallCost(finetunedPricing, usageWithAudio, true);

        // Basic cost with audio: 0.035750, with 3x multiplier: 0.035750 * 3 = 0.107250
        expect(result.value).toBeCloseTo(0.107250, 6);
      });

      it('should handle missing ft_price_hike gracefully', () => {
        const pricingWithoutFt: Pricing = {
          text: basicTextPricing,
          audio: basicAudioPricing
          // No ft_price_hike
        };

        const result = computeApiCallCost(pricingWithoutFt, basicUsage, true);

        // Should not apply any multiplier if ft_price_hike is undefined
        expect(result.value).toBeCloseTo(0.0125, 6);
      });
    });

    describe('Edge Cases and Error Handling', () => {
      it('should handle negative token counts gracefully', () => {
        const negativeUsage: CompletionUsage = {
          prompt_tokens: -100,
          completion_tokens: -50,
          total_tokens: -150
        };

        const result = computeApiCallCost(textOnlyPricing, negativeUsage);

        // Negative costs should be mathematically correct
        expect(result.value).toBeCloseTo(-0.00125, 6);
      });

      it('should handle very small pricing values', () => {
        const microPricing: Pricing = {
          text: {
            prompt: 0.001,
            completion: 0.002,
            cached_discount: 1.0
          }
        };

        const result = computeApiCallCost(microPricing, basicUsage);

        expect(result.value).toBeCloseTo(0.000002, 9);
      });

      it('should handle very large pricing values', () => {
        const expensivePricing: Pricing = {
          text: {
            prompt: 1000000,
            completion: 2000000,
            cached_discount: 1.0
          }
        };

        const result = computeApiCallCost(expensivePricing, basicUsage);

        expect(result.value).toBeCloseTo(2000, 6);
      });

      it('should handle extreme cached discount values', () => {
        const extremeDiscountPricing: Pricing = {
          text: {
            prompt: 10.0,
            completion: 20.0,
            cached_discount: 0.0 // 100% discount
          }
        };

        const result = computeApiCallCost(extremeDiscountPricing, usageWithCached);

        // Cached tokens should cost 0
        expect(result.value).toBeCloseTo(0.018, 6);
      });
    });
  });

  describe('computeCostBreakdown', () => {
    describe('Basic Breakdown Calculations', () => {
      it('should provide detailed cost breakdown for text-only usage', () => {
        const result = computeCostBreakdown(textOnlyPricing, basicUsage);

        expect(result.text_cost.value).toBeCloseTo(0.0125, 6);
        expect(result.text_cost.currency).toBe('USD');
        expect(result.audio_cost.value).toBe(0);
        expect(result.audio_cost.currency).toBe('USD');
        expect(result.total_cost.value).toBeCloseTo(0.0125, 6);
        expect(result.total_cost.currency).toBe('USD');
      });

      it('should provide detailed cost breakdown with audio tokens', () => {
        const result = computeCostBreakdown(textWithAudioPricing, usageWithAudio);

        expect(result.text_cost.value).toBeCloseTo(0.010750, 6);
        expect(result.audio_cost.value).toBeCloseTo(0.025, 6);
        expect(result.total_cost.value).toBeCloseTo(0.035750, 6);
        
        // Verify that text + audio equals total
        expect(result.text_cost.value + result.audio_cost.value).toBeCloseTo(result.total_cost.value, 6);
      });

      it('should handle cached tokens in breakdown', () => {
        const result = computeCostBreakdown(textOnlyPricing, usageWithCached);

        expect(result.text_cost.value).toBeCloseTo(0.012, 6);
        expect(result.audio_cost.value).toBe(0);
        expect(result.total_cost.value).toBeCloseTo(0.012, 6);
      });
    });

    describe('Fine-Tuning in Breakdown', () => {
      it('should apply fine-tuning multiplier to total cost only', () => {
        const result = computeCostBreakdown(finetunedPricing, usageWithAudio, true);

        // Fine-tuning multiplier is only applied to the total, not individual components
        // Base costs: text ~0.010750, audio ~0.025
        // Total with 3x multiplier: (0.010750 + 0.025) * 3 = 0.107250
        expect(result.text_cost.value).toBeCloseTo(0.010750, 6);
        expect(result.audio_cost.value).toBeCloseTo(0.025, 6);
        expect(result.total_cost.value).toBeCloseTo(0.107250, 6);
      });

      it('should not apply fine-tuning when isFt is false', () => {
        const result = computeCostBreakdown(finetunedPricing, basicUsage, false);

        expect(result.total_cost.value).toBeCloseTo(0.0125, 6);
      });
    });

    describe('Breakdown Consistency', () => {
      it('should match computeApiCallCost results', () => {
        const costResult = computeApiCallCost(textWithAudioPricing, usageWithAudio);
        const breakdownResult = computeCostBreakdown(textWithAudioPricing, usageWithAudio);

        expect(breakdownResult.total_cost.value).toBeCloseTo(costResult.value, 6);
        expect(breakdownResult.total_cost.currency).toBe(costResult.currency);
      });

      it('should match computeApiCallCost results with fine-tuning', () => {
        const costResult = computeApiCallCost(finetunedPricing, usageWithAudio, true);
        const breakdownResult = computeCostBreakdown(finetunedPricing, usageWithAudio, true);

        expect(breakdownResult.total_cost.value).toBeCloseTo(costResult.value, 6);
        expect(breakdownResult.total_cost.currency).toBe(costResult.currency);
      });

      it('should have text + audio = total for cases without fine-tuning', () => {
        const testCases = [
          { pricing: textOnlyPricing, usage: basicUsage, isFt: false },
          { pricing: textWithAudioPricing, usage: usageWithAudio, isFt: false }
        ];

        testCases.forEach(({ pricing, usage, isFt }) => {
          const result = computeCostBreakdown(pricing, usage, isFt);
          const calculatedTotal = result.text_cost.value + result.audio_cost.value;
          
          expect(calculatedTotal).toBeCloseTo(result.total_cost.value, 6);
        });
      });

      it('should apply fine-tuning multiplier only to total', () => {
        const result = computeCostBreakdown(finetunedPricing, usageWithCached, true);
        const baseTotal = result.text_cost.value + result.audio_cost.value;
        
        // Fine-tuning multiplier should be applied to the combined total
        expect(result.total_cost.value).toBeCloseTo(baseTotal * 3.0, 6);
      });
    });
  });

  describe('Type Definitions and Interfaces', () => {
    it('should define CompletionUsage interface correctly', () => {
      const usage: CompletionUsage = {
        prompt_tokens: 100,
        completion_tokens: 50,
        total_tokens: 150,
        prompt_tokens_details: {
          cached_tokens: 10,
          audio_tokens: 5
        },
        completion_tokens_details: {
          audio_tokens: 3,
          reasoning_tokens: 20
        }
      };

      expect(usage.prompt_tokens).toBe(100);
      expect(usage.completion_tokens).toBe(50);
      expect(usage.total_tokens).toBe(150);
      expect(usage.prompt_tokens_details?.cached_tokens).toBe(10);
      expect(usage.prompt_tokens_details?.audio_tokens).toBe(5);
      expect(usage.completion_tokens_details?.audio_tokens).toBe(3);
      expect(usage.completion_tokens_details?.reasoning_tokens).toBe(20);
    });

    it('should define Amount interface correctly', () => {
      const amount: Amount = {
        value: 0.0125,
        currency: 'USD'
      };

      expect(amount.value).toBe(0.0125);
      expect(amount.currency).toBe('USD');
    });

    it('should define Pricing interfaces correctly', () => {
      const textPricing: TextPricing = {
        prompt: 5.0,
        completion: 15.0,
        cached_discount: 0.5
      };

      const audioPricing: AudioPricing = {
        prompt: 100.0,
        completion: 200.0
      };

      const fullPricing: Pricing = {
        text: textPricing,
        audio: audioPricing,
        ft_price_hike: 3.0
      };

      expect(textPricing.prompt).toBe(5.0);
      expect(textPricing.completion).toBe(15.0);
      expect(textPricing.cached_discount).toBe(0.5);
      expect(audioPricing.prompt).toBe(100.0);
      expect(audioPricing.completion).toBe(200.0);
      expect(fullPricing.ft_price_hike).toBe(3.0);
    });

    it('should define CostBreakdown interface correctly', () => {
      const breakdown: CostBreakdown = {
        text_cost: { value: 0.01, currency: 'USD' },
        audio_cost: { value: 0.005, currency: 'USD' },
        total_cost: { value: 0.015, currency: 'USD' }
      };

      expect(breakdown.text_cost.value).toBe(0.01);
      expect(breakdown.audio_cost.value).toBe(0.005);
      expect(breakdown.total_cost.value).toBe(0.015);
      expect(breakdown.text_cost.currency).toBe('USD');
    });
  });

  describe('Integration and Real-World Scenarios', () => {
    it('should handle GPT-4o pricing scenario', () => {
      const gpt4oPricing: Pricing = {
        text: {
          prompt: 2.5,      // $2.50 per 1M tokens
          completion: 10.0, // $10.00 per 1M tokens
          cached_discount: 0.5
        }
      };

      const usage: CompletionUsage = {
        prompt_tokens: 1500,
        completion_tokens: 800,
        total_tokens: 2300
      };

      const result = computeApiCallCost(gpt4oPricing, usage);
      
      // Expected: (1500 * 2.5 + 800 * 10) / 1e6 = 11750 / 1e6 = 0.011750
      expect(result.value).toBeCloseTo(0.011750, 6);
    });

    it('should handle Claude pricing scenario with cached tokens', () => {
      const claudePricing: Pricing = {
        text: {
          prompt: 3.0,      // $3.00 per 1M tokens
          completion: 15.0, // $15.00 per 1M tokens
          cached_discount: 0.1 // 90% discount
        }
      };

      const usage: CompletionUsage = {
        prompt_tokens: 2000,
        completion_tokens: 1000,
        total_tokens: 3000,
        prompt_tokens_details: {
          cached_tokens: 500
        }
      };

      const breakdown = computeCostBreakdown(claudePricing, usage);
      
      // Regular prompt: 1500 * 3.0 = 4500
      // Cached prompt: 500 * (3.0 * 0.1) = 150
      // Completion: 1000 * 15.0 = 15000
      // Total text: (4500 + 150 + 15000) / 1e6 = 0.01965
      expect(breakdown.text_cost.value).toBeCloseTo(0.01965, 6);
      expect(breakdown.total_cost.value).toBeCloseTo(0.01965, 6);
    });

    it('should handle audio transcription scenario', () => {
      const whisperPricing: Pricing = {
        text: {
          prompt: 0.0,
          completion: 0.0,
          cached_discount: 1.0
        },
        audio: {
          prompt: 6.0,      // $6.00 per 1M tokens (not 6000)
          completion: 0.0
        }
      };

      const usage: CompletionUsage = {
        prompt_tokens: 1000,
        completion_tokens: 100,
        total_tokens: 1100,
        prompt_tokens_details: {
          audio_tokens: 1000
        }
      };

      const result = computeApiCallCost(whisperPricing, usage);
      
      // Audio only: 1000 * 6.0 / 1e6 = 0.006
      expect(result.value).toBeCloseTo(0.006, 6);
    });

    it('should handle fine-tuned model scenario', () => {
      const ftPricing: Pricing = {
        text: {
          prompt: 8.0,      // Higher base price for FT models
          completion: 24.0,
          cached_discount: 1.0
        },
        ft_price_hike: 1.0 // No additional multiplier
      };

      const usage: CompletionUsage = {
        prompt_tokens: 1000,
        completion_tokens: 500,
        total_tokens: 1500
      };

      const result = computeApiCallCost(ftPricing, usage, true);
      
      // (1000 * 8 + 500 * 24) / 1e6 = 20000 / 1e6 = 0.02
      expect(result.value).toBeCloseTo(0.02, 6);
    });
  });
});