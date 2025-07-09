/**
 * Model cards and AI provider configuration management
 * Equivalent to Python's utils/_model_cards/ directory
 */

export interface ModelCard {
  id: string;
  name: string;
  provider: 'openai' | 'anthropic' | 'xai' | 'gemini' | 'custom';
  model_family: string;
  version: string;
  description: string;
  capabilities: ModelCapabilities;
  limitations: ModelLimitations;
  pricing: ModelPricing;
  performance_benchmarks: PerformanceBenchmarks;
  technical_specs: TechnicalSpecs;
  usage_guidelines: UsageGuidelines;
  supported_features: string[];
  deprecated: boolean;
  deprecation_date?: Date;
  replacement_model?: string;
  created_at: Date;
  updated_at: Date;
}

export interface ModelCapabilities {
  text_generation: boolean;
  text_analysis: boolean;
  document_extraction: boolean;
  image_understanding: boolean;
  function_calling: boolean;
  json_mode: boolean;
  streaming: boolean;
  batch_processing: boolean;
  fine_tuning: boolean;
  embedding_generation: boolean;
  reasoning: boolean;
  code_generation: boolean;
  multilingual: boolean;
  supported_languages?: string[];
  max_context_window: number;
  max_output_tokens: number;
}

export interface ModelLimitations {
  rate_limits: {
    requests_per_minute: number;
    tokens_per_minute: number;
    requests_per_day?: number;
  };
  content_restrictions: string[];
  geographical_restrictions?: string[];
  knowledge_cutoff_date?: Date;
  training_data_cutoff?: Date;
  known_biases?: string[];
  accuracy_limitations?: string[];
  performance_degradation_scenarios?: string[];
}

export interface ModelPricing {
  input_tokens_per_1k: number; // USD
  output_tokens_per_1k: number; // USD
  image_tokens_per_1k?: number; // USD for vision models
  fine_tuning_training_per_1k?: number; // USD
  fine_tuning_usage_per_1k?: number; // USD
  batch_discount_percentage?: number;
  currency: 'USD' | 'EUR' | 'GBP';
  billing_unit: 'token' | 'request' | 'minute';
  minimum_charge?: number;
  free_tier?: {
    requests_per_month: number;
    tokens_per_month: number;
  };
}

export interface PerformanceBenchmarks {
  accuracy_scores: {
    general_text: number; // 0-1
    document_extraction: number;
    reasoning: number;
    code_generation: number;
    multilingual: number;
  };
  speed_metrics: {
    average_latency_ms: number;
    tokens_per_second: number;
    throughput_requests_per_minute: number;
  };
  reliability_metrics: {
    uptime_percentage: number;
    error_rate_percentage: number;
    consistency_score: number; // 0-1
  };
  cost_efficiency: {
    cost_per_accurate_response: number; // USD
    cost_performance_ratio: number; // performance/cost
  };
}

export interface TechnicalSpecs {
  architecture: string;
  parameter_count?: string;
  training_data_size?: string;
  training_compute?: string;
  inference_hardware: string[];
  optimization_techniques: string[];
  supported_formats: {
    input: string[];
    output: string[];
  };
  api_compatibility: {
    openai_compatible: boolean;
    custom_endpoints: boolean;
    webhook_support: boolean;
  };
}

export interface UsageGuidelines {
  recommended_use_cases: string[];
  not_recommended_for: string[];
  best_practices: string[];
  prompt_engineering_tips: string[];
  parameter_recommendations: {
    temperature: {
      creative_tasks: number;
      analytical_tasks: number;
      extraction_tasks: number;
    };
    top_p?: number;
    frequency_penalty?: number;
    presence_penalty?: number;
  };
  context_window_optimization: string[];
}

// Provider configurations
export interface ProviderConfig {
  provider_id: 'openai' | 'anthropic' | 'xai' | 'gemini' | 'custom';
  provider_name: string;
  api_base_url: string;
  authentication: {
    type: 'api_key' | 'oauth' | 'bearer_token' | 'custom';
    header_name: string;
    header_prefix?: string;
  };
  rate_limiting: {
    default_rpm: number;
    default_tpm: number;
    burst_allowance: number;
    backoff_strategy: 'exponential' | 'linear' | 'fixed';
  };
  request_format: {
    content_type: string;
    encoding: string;
    supports_streaming: boolean;
    supports_batch: boolean;
  };
  response_format: {
    content_type: string;
    error_format: 'openai' | 'anthropic' | 'custom';
    streaming_format?: string;
  };
  supported_models: string[];
  model_aliases?: Record<string, string>;
  default_parameters: Record<string, any>;
  health_check: {
    endpoint: string;
    method: 'GET' | 'POST';
    expected_status: number;
    timeout_ms: number;
  };
}

// Model selection utilities
export interface ModelSelectionCriteria {
  task_type: 'extraction' | 'generation' | 'analysis' | 'classification' | 'reasoning';
  accuracy_priority: 'low' | 'medium' | 'high' | 'critical';
  speed_priority: 'low' | 'medium' | 'high' | 'critical';
  cost_priority: 'low' | 'medium' | 'high' | 'critical';
  context_length_required: number;
  output_length_required: number;
  supports_images: boolean;
  supports_function_calling: boolean;
  supports_json_mode: boolean;
  language_requirements?: string[];
  compliance_requirements?: string[];
  geographical_restrictions?: string[];
}

export interface ModelRecommendation {
  model_id: string;
  model_name: string;
  provider: string;
  confidence_score: number; // 0-1
  rationale: string[];
  expected_performance: {
    accuracy: number;
    latency_ms: number;
    cost_per_request: number;
  };
  trade_offs: string[];
  alternative_models: Array<{
    model_id: string;
    reason: string;
    trade_off: string;
  }>;
}

// Built-in model cards
export const modelCards: Record<string, ModelCard> = {
  'gpt-4o': {
    id: 'gpt-4o',
    name: 'GPT-4o',
    provider: 'openai',
    model_family: 'gpt-4',
    version: '2024-08-06',
    description: 'Most advanced multimodal model with vision capabilities',
    capabilities: {
      text_generation: true,
      text_analysis: true,
      document_extraction: true,
      image_understanding: true,
      function_calling: true,
      json_mode: true,
      streaming: true,
      batch_processing: true,
      fine_tuning: false,
      embedding_generation: false,
      reasoning: true,
      code_generation: true,
      multilingual: true,
      supported_languages: ['en', 'es', 'fr', 'de', 'it', 'pt', 'ru', 'ja', 'ko', 'zh'],
      max_context_window: 128000,
      max_output_tokens: 16384,
    },
    limitations: {
      rate_limits: {
        requests_per_minute: 10000,
        tokens_per_minute: 2000000,
      },
      content_restrictions: ['NSFW content', 'Illegal activities', 'Personal data'],
      knowledge_cutoff_date: new Date('2024-04-01'),
    },
    pricing: {
      input_tokens_per_1k: 0.0025,
      output_tokens_per_1k: 0.01,
      image_tokens_per_1k: 0.001275,
      currency: 'USD',
      billing_unit: 'token',
    },
    performance_benchmarks: {
      accuracy_scores: {
        general_text: 0.95,
        document_extraction: 0.92,
        reasoning: 0.94,
        code_generation: 0.91,
        multilingual: 0.89,
      },
      speed_metrics: {
        average_latency_ms: 1200,
        tokens_per_second: 45,
        throughput_requests_per_minute: 8500,
      },
      reliability_metrics: {
        uptime_percentage: 99.9,
        error_rate_percentage: 0.1,
        consistency_score: 0.93,
      },
      cost_efficiency: {
        cost_per_accurate_response: 0.015,
        cost_performance_ratio: 6.3,
      },
    },
    technical_specs: {
      architecture: 'Transformer',
      inference_hardware: ['GPU'],
      optimization_techniques: ['Mixed precision', 'Model parallelism'],
      supported_formats: {
        input: ['text', 'image'],
        output: ['text', 'json'],
      },
      api_compatibility: {
        openai_compatible: true,
        custom_endpoints: false,
        webhook_support: false,
      },
    },
    usage_guidelines: {
      recommended_use_cases: [
        'Complex document extraction',
        'Multimodal analysis',
        'Advanced reasoning tasks',
        'Code generation',
      ],
      not_recommended_for: [
        'Simple text classification',
        'High-frequency API calls',
        'Cost-sensitive applications',
      ],
      best_practices: [
        'Use clear, specific prompts',
        'Leverage function calling for structured output',
        'Optimize context window usage',
      ],
      prompt_engineering_tips: [
        'Be explicit about output format',
        'Provide examples for complex tasks',
        'Use system messages for role definition',
      ],
      parameter_recommendations: {
        temperature: {
          creative_tasks: 0.7,
          analytical_tasks: 0.1,
          extraction_tasks: 0.0,
        },
      },
      context_window_optimization: [
        'Place important information early',
        'Use concise prompts',
        'Chunk large documents',
      ],
    },
    supported_features: [
      'vision',
      'function_calling',
      'json_mode',
      'streaming',
      'batch',
    ],
    deprecated: false,
    created_at: new Date('2024-05-13'),
    updated_at: new Date('2024-08-06'),
  },
  'gpt-4o-mini': {
    id: 'gpt-4o-mini',
    name: 'GPT-4o Mini',
    provider: 'openai',
    model_family: 'gpt-4',
    version: '2024-07-18',
    description: 'Fast and cost-effective model for simple tasks',
    capabilities: {
      text_generation: true,
      text_analysis: true,
      document_extraction: true,
      image_understanding: true,
      function_calling: true,
      json_mode: true,
      streaming: true,
      batch_processing: true,
      fine_tuning: true,
      embedding_generation: false,
      reasoning: true,
      code_generation: true,
      multilingual: true,
      max_context_window: 128000,
      max_output_tokens: 16384,
    },
    limitations: {
      rate_limits: {
        requests_per_minute: 30000,
        tokens_per_minute: 5000000,
      },
      content_restrictions: ['NSFW content', 'Illegal activities'],
      knowledge_cutoff_date: new Date('2024-04-01'),
    },
    pricing: {
      input_tokens_per_1k: 0.00015,
      output_tokens_per_1k: 0.0006,
      image_tokens_per_1k: 0.001275,
      fine_tuning_training_per_1k: 0.003,
      fine_tuning_usage_per_1k: 0.0012,
      currency: 'USD',
      billing_unit: 'token',
    },
    performance_benchmarks: {
      accuracy_scores: {
        general_text: 0.87,
        document_extraction: 0.84,
        reasoning: 0.85,
        code_generation: 0.83,
        multilingual: 0.81,
      },
      speed_metrics: {
        average_latency_ms: 800,
        tokens_per_second: 65,
        throughput_requests_per_minute: 25000,
      },
      reliability_metrics: {
        uptime_percentage: 99.95,
        error_rate_percentage: 0.05,
        consistency_score: 0.89,
      },
      cost_efficiency: {
        cost_per_accurate_response: 0.003,
        cost_performance_ratio: 28.7,
      },
    },
    technical_specs: {
      architecture: 'Transformer',
      inference_hardware: ['GPU'],
      optimization_techniques: ['Quantization', 'Pruning'],
      supported_formats: {
        input: ['text', 'image'],
        output: ['text', 'json'],
      },
      api_compatibility: {
        openai_compatible: true,
        custom_endpoints: false,
        webhook_support: false,
      },
    },
    usage_guidelines: {
      recommended_use_cases: [
        'Document extraction',
        'Text classification',
        'Simple reasoning',
        'Batch processing',
      ],
      not_recommended_for: [
        'Complex creative writing',
        'Advanced mathematical reasoning',
        'Highly specialized domains',
      ],
      best_practices: [
        'Use for cost-sensitive applications',
        'Ideal for fine-tuning',
        'Good for high-volume processing',
      ],
      prompt_engineering_tips: [
        'Keep prompts concise',
        'Use examples for better performance',
        'Be specific about requirements',
      ],
      parameter_recommendations: {
        temperature: {
          creative_tasks: 0.5,
          analytical_tasks: 0.0,
          extraction_tasks: 0.0,
        },
      },
      context_window_optimization: [
        'Use efficient prompt design',
        'Minimize unnecessary context',
        'Leverage fine-tuning for domain-specific tasks',
      ],
    },
    supported_features: [
      'vision',
      'function_calling',
      'json_mode',
      'streaming',
      'batch',
      'fine_tuning',
    ],
    deprecated: false,
    created_at: new Date('2024-07-18'),
    updated_at: new Date('2024-07-18'),
  },
};

export const providerConfigs: Record<string, ProviderConfig> = {
  openai: {
    provider_id: 'openai',
    provider_name: 'OpenAI',
    api_base_url: 'https://api.openai.com/v1',
    authentication: {
      type: 'api_key',
      header_name: 'Authorization',
      header_prefix: 'Bearer',
    },
    rate_limiting: {
      default_rpm: 10000,
      default_tpm: 2000000,
      burst_allowance: 100,
      backoff_strategy: 'exponential',
    },
    request_format: {
      content_type: 'application/json',
      encoding: 'utf-8',
      supports_streaming: true,
      supports_batch: true,
    },
    response_format: {
      content_type: 'application/json',
      error_format: 'openai',
      streaming_format: 'text/event-stream',
    },
    supported_models: ['gpt-4o', 'gpt-4o-mini', 'gpt-4-turbo', 'gpt-3.5-turbo'],
    default_parameters: {
      temperature: 0.0,
      max_tokens: 4096,
    },
    health_check: {
      endpoint: '/models',
      method: 'GET',
      expected_status: 200,
      timeout_ms: 5000,
    },
  },
  anthropic: {
    provider_id: 'anthropic',
    provider_name: 'Anthropic',
    api_base_url: 'https://api.anthropic.com',
    authentication: {
      type: 'api_key',
      header_name: 'x-api-key',
    },
    rate_limiting: {
      default_rpm: 4000,
      default_tpm: 400000,
      burst_allowance: 50,
      backoff_strategy: 'exponential',
    },
    request_format: {
      content_type: 'application/json',
      encoding: 'utf-8',
      supports_streaming: true,
      supports_batch: false,
    },
    response_format: {
      content_type: 'application/json',
      error_format: 'anthropic',
      streaming_format: 'text/event-stream',
    },
    supported_models: ['claude-3-5-sonnet-20241022', 'claude-3-haiku-20240307'],
    default_parameters: {
      temperature: 0.0,
      max_tokens: 4096,
    },
    health_check: {
      endpoint: '/v1/messages',
      method: 'POST',
      expected_status: 400, // Expects validation error without proper request
      timeout_ms: 5000,
    },
  },
};

// Utility functions
export class ModelCardManager {
  private static instance: ModelCardManager;
  private cards: Map<string, ModelCard> = new Map();
  private configs: Map<string, ProviderConfig> = new Map();

  constructor() {
    // Load built-in model cards
    Object.entries(modelCards).forEach(([id, card]) => {
      this.cards.set(id, card);
    });

    // Load built-in provider configs
    Object.entries(providerConfigs).forEach(([id, config]) => {
      this.configs.set(id, config);
    });
  }

  static getInstance(): ModelCardManager {
    if (!ModelCardManager.instance) {
      ModelCardManager.instance = new ModelCardManager();
    }
    return ModelCardManager.instance;
  }

  getModelCard(modelId: string): ModelCard | undefined {
    return this.cards.get(modelId);
  }

  getProviderConfig(providerId: string): ProviderConfig | undefined {
    return this.configs.get(providerId);
  }

  getAllModels(): ModelCard[] {
    return Array.from(this.cards.values());
  }

  getModelsByProvider(providerId: string): ModelCard[] {
    return Array.from(this.cards.values()).filter(card => card.provider === providerId);
  }

  recommendModel(criteria: ModelSelectionCriteria): ModelRecommendation[] {
    const models = Array.from(this.cards.values());
    const recommendations: ModelRecommendation[] = [];

    for (const model of models) {
      let score = 0;
      const rationale: string[] = [];

      // Check basic requirements
      if (criteria.context_length_required > model.capabilities.max_context_window) continue;
      if (criteria.output_length_required > model.capabilities.max_output_tokens) continue;
      if (criteria.supports_images && !model.capabilities.image_understanding) continue;
      if (criteria.supports_function_calling && !model.capabilities.function_calling) continue;
      if (criteria.supports_json_mode && !model.capabilities.json_mode) continue;

      // Score based on priorities
      const accuracy = model.performance_benchmarks.accuracy_scores.general_text;
      const speed = 1 / (model.performance_benchmarks.speed_metrics.average_latency_ms / 1000);
      const cost = 1 / model.pricing.input_tokens_per_1k;

      switch (criteria.accuracy_priority) {
        case 'critical': score += accuracy * 0.5; break;
        case 'high': score += accuracy * 0.35; break;
        case 'medium': score += accuracy * 0.2; break;
        case 'low': score += accuracy * 0.1; break;
      }

      switch (criteria.speed_priority) {
        case 'critical': score += speed * 0.3; break;
        case 'high': score += speed * 0.2; break;
        case 'medium': score += speed * 0.15; break;
        case 'low': score += speed * 0.05; break;
      }

      switch (criteria.cost_priority) {
        case 'critical': score += cost * 0.3; break;
        case 'high': score += cost * 0.2; break;
        case 'medium': score += cost * 0.15; break;
        case 'low': score += cost * 0.05; break;
      }

      if (score > 0.3) { // Minimum threshold
        recommendations.push({
          model_id: model.id,
          model_name: model.name,
          provider: model.provider,
          confidence_score: Math.min(score, 1.0),
          rationale,
          expected_performance: {
            accuracy: accuracy,
            latency_ms: model.performance_benchmarks.speed_metrics.average_latency_ms,
            cost_per_request: model.pricing.input_tokens_per_1k * 0.5, // Rough estimate
          },
          trade_offs: [],
          alternative_models: [],
        });
      }
    }

    return recommendations.sort((a, b) => b.confidence_score - a.confidence_score);
  }

  addModelCard(card: ModelCard): void {
    this.cards.set(card.id, card);
  }

  updateModelCard(modelId: string, updates: Partial<ModelCard>): void {
    const existing = this.cards.get(modelId);
    if (existing) {
      this.cards.set(modelId, { ...existing, ...updates, updated_at: new Date() });
    }
  }
}

// Export values but not types in default export
export default {
  ModelCardManager,
  modelCards,
  providerConfigs,
};