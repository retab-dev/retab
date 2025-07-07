export interface CompletionUsage {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  prompt_tokens_details?: {
    cached_tokens?: number;
    audio_tokens?: number;
  };
  completion_tokens_details?: {
    audio_tokens?: number;
    reasoning_tokens?: number;
  };
}

export interface Amount {
  value: number;
  currency: string;
}

export interface TextPricing {
  prompt: number;
  completion: number;
  cached_discount: number;
}

export interface AudioPricing {
  prompt: number;
  completion: number;
}

export interface Pricing {
  text: TextPricing;
  audio?: AudioPricing;
  ft_price_hike?: number;
}

export interface CostBreakdown {
  text_cost: Amount;
  audio_cost: Amount;
  total_cost: Amount;
}

export function computeApiCallCost(pricing: Pricing, usage: CompletionUsage, isFt: boolean = false): Amount {
  // Process prompt tokens
  const promptCachedText = usage.prompt_tokens_details?.cached_tokens || 0;
  const promptAudio = usage.prompt_tokens_details?.audio_tokens || 0;
  const promptRegularText = usage.prompt_tokens - promptCachedText - promptAudio;

  // Process completion tokens
  const completionAudio = usage.completion_tokens_details?.audio_tokens || 0;
  const completionRegularText = usage.completion_tokens - completionAudio;

  // Calculate text token costs
  const costTextPrompt = promptRegularText * pricing.text.prompt;
  const textCachedPrice = pricing.text.prompt * pricing.text.cached_discount;
  const costTextCached = promptCachedText * textCachedPrice;
  const costTextCompletion = completionRegularText * pricing.text.completion;
  const totalTextCost = costTextPrompt + costTextCached + costTextCompletion;

  // Calculate audio token costs (if any)
  let totalAudioCost = 0.0;
  if (pricing.audio) {
    const costAudioPrompt = promptAudio * pricing.audio.prompt;
    const costAudioCompletion = completionAudio * pricing.audio.completion;
    totalAudioCost = costAudioPrompt + costAudioCompletion;
  }

  let totalCost = (totalTextCost + totalAudioCost) / 1e6;

  // Apply fine-tuning price hike if applicable
  if (isFt && pricing.ft_price_hike) {
    totalCost *= pricing.ft_price_hike;
  }

  return {
    value: totalCost,
    currency: "USD"
  };
}

export function computeCostBreakdown(pricing: Pricing, usage: CompletionUsage, isFt: boolean = false): CostBreakdown {
  // Process prompt tokens
  const promptCachedText = usage.prompt_tokens_details?.cached_tokens || 0;
  const promptAudio = usage.prompt_tokens_details?.audio_tokens || 0;
  const promptRegularText = usage.prompt_tokens - promptCachedText - promptAudio;

  // Process completion tokens
  const completionAudio = usage.completion_tokens_details?.audio_tokens || 0;
  const completionRegularText = usage.completion_tokens - completionAudio;

  // Calculate text token costs
  const costTextPrompt = promptRegularText * pricing.text.prompt;
  const textCachedPrice = pricing.text.prompt * pricing.text.cached_discount;
  const costTextCached = promptCachedText * textCachedPrice;
  const costTextCompletion = completionRegularText * pricing.text.completion;
  const totalTextCost = (costTextPrompt + costTextCached + costTextCompletion) / 1e6;

  // Calculate audio token costs (if any)
  let totalAudioCost = 0.0;
  if (pricing.audio) {
    const costAudioPrompt = promptAudio * pricing.audio.prompt;
    const costAudioCompletion = completionAudio * pricing.audio.completion;
    totalAudioCost = (costAudioPrompt + costAudioCompletion) / 1e6;
  }

  let totalCost = totalTextCost + totalAudioCost;

  // Apply fine-tuning price hike if applicable
  if (isFt && pricing.ft_price_hike) {
    totalCost *= pricing.ft_price_hike;
  }

  return {
    text_cost: { value: totalTextCost, currency: "USD" },
    audio_cost: { value: totalAudioCost, currency: "USD" },
    total_cost: { value: totalCost, currency: "USD" }
  };
}