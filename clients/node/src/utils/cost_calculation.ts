import { Amount, Pricing } from '../types/ai_models.js';

// Basic pricing data for common models (this would typically come from a config or API)
const MODEL_PRICING: Record<string, Pricing> = {
  'gpt-4o': {
    text: { prompt: 2.5, completion: 10.0, cached_discount: 1.0 },
    ft_price_hike: 1.0,
  },
  'gpt-4o-mini': {
    text: { prompt: 0.15, completion: 0.6, cached_discount: 1.0 },
    ft_price_hike: 1.0,
  },
  'gpt-4o-2024-11-20': {
    text: { prompt: 2.5, completion: 10.0, cached_discount: 1.0 },
    ft_price_hike: 1.0,
  },
  'gpt-4o-2024-08-06': {
    text: { prompt: 2.5, completion: 10.0, cached_discount: 1.0 },
    ft_price_hike: 1.0,
  },
  'gpt-4o-mini-2024-07-18': {
    text: { prompt: 0.15, completion: 0.6, cached_discount: 1.0 },
    ft_price_hike: 1.0,
  },
  'claude-3-5-sonnet-latest': {
    text: { prompt: 3.0, completion: 15.0, cached_discount: 1.0 },
    ft_price_hike: 1.0,
  },
  'claude-3-5-sonnet-20241022': {
    text: { prompt: 3.0, completion: 15.0, cached_discount: 1.0 },
    ft_price_hike: 1.0,
  },
  'gemini-2.0-flash': {
    text: { prompt: 0.075, completion: 0.3, cached_discount: 1.0 },
    ft_price_hike: 1.0,
  },
  'gemini-2.5-pro': {
    text: { prompt: 1.25, completion: 5.0, cached_discount: 1.0 },
    ft_price_hike: 1.0,
  },
};

interface Usage {
  prompt_tokens?: number;
  completion_tokens?: number;
  total_tokens?: number;
  cached_tokens?: number;
}

/**
 * Compute the cost of a model usage.
 */
export function computeCostFromModel(
  model: string,
  usage: Usage,
  currency: string = 'USD'
): Amount {
  const pricing = MODEL_PRICING[model];
  if (!pricing) {
    // Return zero cost for unknown models
    return { value: 0, currency };
  }

  const promptTokens = usage.prompt_tokens || 0;
  const completionTokens = usage.completion_tokens || 0;
  const cachedTokens = usage.cached_tokens || 0;

  // Calculate costs per 1M tokens
  const promptCost = (promptTokens / 1_000_000) * pricing.text.prompt;
  const completionCost = (completionTokens / 1_000_000) * pricing.text.completion;
  
  // Apply cached discount if applicable
  const cachedCost = (cachedTokens / 1_000_000) * pricing.text.prompt * pricing.text.cached_discount;

  const totalCost = promptCost + completionCost + cachedCost;

  return {
    value: Math.round(totalCost * 100000) / 100000, // Round to 5 decimal places
    currency,
  };
}

/**
 * Compute cost breakdown for detailed analysis.
 */
export interface CostBreakdown {
  prompt_cost: Amount;
  completion_cost: Amount;
  cached_cost: Amount;
  total_cost: Amount;
  prompt_tokens: number;
  completion_tokens: number;
  cached_tokens: number;
}

export function computeCostFromModelWithBreakdown(
  model: string,
  usage: Usage,
  currency: string = 'USD'
): CostBreakdown {
  const pricing = MODEL_PRICING[model];
  if (!pricing) {
    const zeroCost = { value: 0, currency };
    return {
      prompt_cost: zeroCost,
      completion_cost: zeroCost,
      cached_cost: zeroCost,
      total_cost: zeroCost,
      prompt_tokens: usage.prompt_tokens || 0,
      completion_tokens: usage.completion_tokens || 0,
      cached_tokens: usage.cached_tokens || 0,
    };
  }

  const promptTokens = usage.prompt_tokens || 0;
  const completionTokens = usage.completion_tokens || 0;
  const cachedTokens = usage.cached_tokens || 0;

  const promptCostValue = (promptTokens / 1_000_000) * pricing.text.prompt;
  const completionCostValue = (completionTokens / 1_000_000) * pricing.text.completion;
  const cachedCostValue = (cachedTokens / 1_000_000) * pricing.text.prompt * pricing.text.cached_discount;

  const promptCost = { value: Math.round(promptCostValue * 100000) / 100000, currency };
  const completionCost = { value: Math.round(completionCostValue * 100000) / 100000, currency };
  const cachedCost = { value: Math.round(cachedCostValue * 100000) / 100000, currency };
  const totalCost = { 
    value: Math.round((promptCostValue + completionCostValue + cachedCostValue) * 100000) / 100000, 
    currency 
  };

  return {
    prompt_cost: promptCost,
    completion_cost: completionCost,
    cached_cost: cachedCost,
    total_cost: totalCost,
    prompt_tokens: promptTokens,
    completion_tokens: completionTokens,
    cached_tokens: cachedTokens,
  };
}