import { z } from 'zod';

// AI Provider types
export const AIProviderSchema = z.enum(['OpenAI', 'Anthropic', 'Gemini', 'xAI', 'Retab']);
export type AIProvider = z.infer<typeof AIProviderSchema>;

export const OpenAICompatibleProviderSchema = z.enum(['OpenAI', 'xAI']);
export type OpenAICompatibleProvider = z.infer<typeof OpenAICompatibleProviderSchema>;

// Model types
export const GeminiModelSchema = z.enum([
  'gemini-2.5-pro',
  'gemini-2.5-flash',
  'gemini-2.5-pro-preview-06-05',
  'gemini-2.5-pro-preview-05-06',
  'gemini-2.5-pro-preview-03-25',
  'gemini-2.5-flash-preview-05-20',
  'gemini-2.5-flash-preview-04-17',
  'gemini-2.5-flash-lite-preview-06-17',
  'gemini-2.5-pro-exp-03-25',
  'gemini-2.0-flash-lite',
  'gemini-2.0-flash',
]);
export type GeminiModel = z.infer<typeof GeminiModelSchema>;

export const AnthropicModelSchema = z.enum([
  'claude-3-5-sonnet-latest',
  'claude-3-5-sonnet-20241022',
  'claude-3-opus-20240229',
  'claude-3-sonnet-20240229',
  'claude-3-haiku-20240307',
]);
export type AnthropicModel = z.infer<typeof AnthropicModelSchema>;

export const OpenAIModelSchema = z.enum([
  'gpt-4o',
  'gpt-4o-mini',
  'chatgpt-4o-latest',
  'gpt-4.1',
  'gpt-4.1-mini',
  'gpt-4.1-mini-2025-04-14',
  'gpt-4.1-2025-04-14',
  'gpt-4.1-nano',
  'gpt-4.1-nano-2025-04-14',
  'gpt-4o-2024-11-20',
  'gpt-4o-2024-08-06',
  'gpt-4o-2024-05-13',
  'gpt-4o-mini-2024-07-18',
  'o1',
  'o1-2024-12-17',
  'o3',
  'o3-2025-04-16',
  'o4-mini',
  'o4-mini-2025-04-16',
  'gpt-4o-audio-preview-2024-12-17',
  'gpt-4o-audio-preview-2024-10-01',
  'gpt-4o-realtime-preview-2024-12-17',
  'gpt-4o-realtime-preview-2024-10-01',
  'gpt-4o-mini-audio-preview-2024-12-17',
  'gpt-4o-mini-realtime-preview-2024-12-17',
]);
export type OpenAIModel = z.infer<typeof OpenAIModelSchema>;

export const XAIModelSchema = z.enum(['grok-3', 'grok-3-mini']);
export type XAIModel = z.infer<typeof XAIModelSchema>;

export const RetabModelSchema = z.enum(['auto-large', 'auto-small', 'auto-micro']);
export type RetabModel = z.infer<typeof RetabModelSchema>;

export const PureLLMModelSchema = z.union([
  OpenAIModelSchema,
  AnthropicModelSchema,
  XAIModelSchema,
  GeminiModelSchema,
  RetabModelSchema,
]);
export type PureLLMModel = z.infer<typeof PureLLMModelSchema>;

export const LLMModelSchema = z.union([PureLLMModelSchema, z.literal('human')]);
export type LLMModel = z.infer<typeof LLMModelSchema>;

// Inference Settings
export const InferenceSettingsSchema = z.object({
  // Add inference settings fields as needed
  temperature: z.number().optional(),
  max_tokens: z.number().optional(),
  top_p: z.number().optional(),
});
export type InferenceSettings = z.infer<typeof InferenceSettingsSchema>;

// Finetuned Model
export const FinetunedModelSchema = z.object({
  object: z.literal('finetuned_model'),
  organization_id: z.string(),
  model: z.string(),
  schema_id: z.string(),
  schema_data_id: z.string(),
  finetuning_props: InferenceSettingsSchema,
  evaluation_id: z.string().nullable().optional(),
  created_at: z.string().datetime(),
});
export type FinetunedModel = z.infer<typeof FinetunedModelSchema>;

// Monthly Usage
export const MonthlyUsageResponseContentSchema = z.object({
  credits_count: z.number(),
});
export type MonthlyUsageResponseContent = z.infer<typeof MonthlyUsageResponseContentSchema>;
export type MonthlyUsageResponse = MonthlyUsageResponseContent;

// Amount and Pricing
export const AmountSchema = z.object({
  value: z.number(),
  currency: z.string(),
});
export type Amount = z.infer<typeof AmountSchema>;

export const TokenPriceSchema = z.object({
  prompt: z.number(),
  completion: z.number(),
  cached_discount: z.number().default(1.0),
});
export type TokenPrice = z.infer<typeof TokenPriceSchema>;

export const PricingSchema = z.object({
  text: TokenPriceSchema,
  audio: TokenPriceSchema.optional(),
  ft_price_hike: z.number().default(1.0),
});
export type Pricing = z.infer<typeof PricingSchema>;

// Model capabilities
export const ModelModalitySchema = z.enum(['text', 'audio', 'image']);
export type ModelModality = z.infer<typeof ModelModalitySchema>;

export const EndpointTypeSchema = z.enum([
  'chat_completions',
  'responses',
  'assistants',
  'batch',
  'fine_tuning',
  'embeddings',
  'speech_generation',
  'translation',
  'completions_legacy',
  'image_generation',
  'transcription',
  'moderation',
  'realtime',
]);
export type EndpointType = z.infer<typeof EndpointTypeSchema>;

export const FeatureTypeSchema = z.enum([
  'streaming',
  'function_calling',
  'structured_outputs',
  'distillation',
  'fine_tuning',
  'predicted_outputs',
  'schema_generation',
]);
export type FeatureType = z.infer<typeof FeatureTypeSchema>;

export const ModelCapabilitiesSchema = z.object({
  modalities: z.array(ModelModalitySchema),
  endpoints: z.array(EndpointTypeSchema),
  features: z.array(FeatureTypeSchema),
});
export type ModelCapabilities = z.infer<typeof ModelCapabilitiesSchema>;

export const ModelCardPermissionsSchema = z.object({
  show_in_free_picker: z.boolean().default(false),
  show_in_paid_picker: z.boolean().default(false),
});
export type ModelCardPermissions = z.infer<typeof ModelCardPermissionsSchema>;

export const ModelCardSchema = z.object({
  model: z.union([LLMModelSchema, z.string()]),
  pricing: PricingSchema,
  capabilities: ModelCapabilitiesSchema,
  temperature_support: z.boolean().default(true),
  reasoning_effort_support: z.boolean().default(false),
  permissions: ModelCardPermissionsSchema.default({}),
}).transform((data) => ({
  ...data,
  is_finetuned: typeof data.model === 'string' && data.model.includes('ft:'),
}));
export type ModelCard = z.infer<typeof ModelCardSchema>;