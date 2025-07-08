import { z } from 'zod';
import { AIProviderSchema } from '../ai_models.js';

export const ExternalAPIKeyRequestSchema = z.object({
  provider: AIProviderSchema,
  api_key: z.string(),
});
export type ExternalAPIKeyRequest = z.infer<typeof ExternalAPIKeyRequestSchema>;

export const ExternalAPIKeySchema = z.object({
  provider: AIProviderSchema,
  is_configured: z.boolean(),
  last_updated: z.string().datetime().nullable().optional(),
});
export type ExternalAPIKey = z.infer<typeof ExternalAPIKeySchema>;