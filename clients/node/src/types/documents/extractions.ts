import { z } from 'zod';
import { MIMEDataSchema } from '../mime.js';
import { ModalitySchema } from '../modalities.js';
import { BrowserCanvasSchema } from '../browser_canvas.js';
import { ChatCompletionRetabMessageSchema } from '../chat.js';
import { ErrorDetail } from '../standards.js';

// Reasoning effort type
export const ChatCompletionReasoningEffortSchema = z.enum(['low', 'medium', 'high']);
export type ChatCompletionReasoningEffort = z.infer<typeof ChatCompletionReasoningEffortSchema>;

// Document Extract Request
export const DocumentExtractRequestSchema = z.object({
  document: MIMEDataSchema.optional(),
  documents: z.array(MIMEDataSchema),
  modality: ModalitySchema,
  image_resolution_dpi: z.number().int().default(96),
  browser_canvas: BrowserCanvasSchema.default('A4'),
  model: z.string(),
  json_schema: z.record(z.any()),
  temperature: z.number().default(0.0),
  reasoning_effort: ChatCompletionReasoningEffortSchema.default('medium'),
  n_consensus: z.number().int().default(1),
  stream: z.boolean().default(false),
  seed: z.number().int().nullable().optional(),
  store: z.boolean().default(true),
  need_validation: z.boolean().default(false),
}).refine((data) => {
  if (data.n_consensus > 1 && data.temperature === 0) {
    throw new Error('n_consensus greater than 1 but temperature is 0');
  }
  return true;
}).transform((data) => {
  // Handle document/documents compatibility
  if (data.documents && data.documents.length > 0) {
    return { ...data, document: data.documents[0] };
  } else if (data.document) {
    return { ...data, documents: [data.document] };
  }
  throw new Error('document or documents must be provided');
});
export type DocumentExtractRequest = z.infer<typeof DocumentExtractRequestSchema>;

// Consensus Model
export const ConsensusModelSchema = z.object({
  model: z.string(),
  temperature: z.number().default(0.0),
  reasoning_effort: ChatCompletionReasoningEffortSchema.default('medium'),
});
export type ConsensusModel = z.infer<typeof ConsensusModelSchema>;

// Field Location
export const FieldLocationSchema = z.object({
  label: z.string(),
  value: z.string(),
  quote: z.string(),
  file_id: z.string().nullable().optional(),
  page: z.number().int().nullable().optional(),
  bbox_normalized: z.tuple([z.number(), z.number(), z.number(), z.number()]).nullable().optional(),
  score: z.number().nullable().optional(),
  match_level: z.enum(['token', 'line', 'block']).nullable().optional(),
});
export type FieldLocation = z.infer<typeof FieldLocationSchema>;

// Parsed Choice
export const RetabParsedChoiceSchema = z.object({
  finish_reason: z.enum(['stop', 'length', 'tool_calls', 'content_filter', 'function_call']).nullable().optional(),
  field_locations: z.record(FieldLocationSchema).nullable().optional(),
  key_mapping: z.record(z.string().nullable()).nullable().optional(),
  message: z.object({
    content: z.string().nullable().optional(),
    role: z.string(),
    parsed: z.any().nullable().optional(),
  }),
  index: z.number().int(),
});
export type RetabParsedChoice = z.infer<typeof RetabParsedChoiceSchema>;

// Likelihoods source
export const LikelihoodsSourceSchema = z.enum(['consensus', 'log_probs']);
export type LikelihoodsSource = z.infer<typeof LikelihoodsSourceSchema>;

// Amount (for cost calculation)
export const AmountSchema = z.object({
  value: z.number(),
  currency: z.string(),
});
export type Amount = z.infer<typeof AmountSchema>;

// Usage stats
export const UsageSchema = z.object({
  prompt_tokens: z.number().int().optional(),
  completion_tokens: z.number().int().optional(),
  total_tokens: z.number().int().optional(),
});
export type Usage = z.infer<typeof UsageSchema>;

// Parsed Chat Completion
export const RetabParsedChatCompletionSchema = z.object({
  id: z.string(),
  object: z.literal('chat.completion'),
  created: z.number().int(),
  model: z.string(),
  choices: z.array(RetabParsedChoiceSchema),
  usage: UsageSchema.optional(),
  extraction_id: z.string().nullable().optional(),
  likelihoods: z.record(z.any()).nullable().optional(),
  schema_validation_error: z.custom<ErrorDetail>().nullable().optional(),
  request_at: z.string().datetime().nullable().optional(),
  first_token_at: z.string().datetime().nullable().optional(),
  last_token_at: z.string().datetime().nullable().optional(),
}).transform((data) => ({
  ...data,
  get api_cost(): Amount | null {
    // Implementation would compute cost from model and usage
    return null;
  },
}));
export type RetabParsedChatCompletion = z.infer<typeof RetabParsedChatCompletionSchema>;

// UI Response
export const UiResponseSchema = z.object({
  id: z.string(),
  object: z.string(),
  choices: z.array(z.any()),
  model: z.string(),
  extraction_id: z.string().nullable().optional(),
  likelihoods: z.record(z.any()).nullable().optional(),
  schema_validation_error: z.custom<ErrorDetail>().nullable().optional(),
  request_at: z.string().datetime().nullable().optional(),
  first_token_at: z.string().datetime().nullable().optional(),
  last_token_at: z.string().datetime().nullable().optional(),
});
export type UiResponse = z.infer<typeof UiResponseSchema>;

// Log Extraction Request
export const LogExtractionRequestSchema = z.object({
  messages: z.array(ChatCompletionRetabMessageSchema).nullable().optional(),
  openai_messages: z.array(z.any()).nullable().optional(),
  openai_responses_input: z.array(z.any()).nullable().optional(),
  anthropic_messages: z.array(z.any()).nullable().optional(),
  anthropic_system_prompt: z.string().nullable().optional(),
  document: MIMEDataSchema.default({
    id: 'dummy_doc',
    extension: 'txt',
    content: Buffer.from('No document provided').toString('base64'),
    mime_type: 'text/plain',
    unique_filename: 'dummy.txt',
    size: 20,
    filename: 'dummy.txt',
    url: 'data:text/plain;base64,' + Buffer.from('No document provided').toString('base64'),
  }),
  completion: z.union([z.record(z.any()), RetabParsedChatCompletionSchema]).nullable().optional(),
  openai_responses_output: z.any().nullable().optional(),
  json_schema: z.record(z.any()),
  model: z.string(),
  temperature: z.number(),
}).refine((data) => {
  const messagesCandidates = [
    data.messages,
    data.openai_messages,
    data.anthropic_messages,
    data.openai_responses_input,
  ].filter(candidate => candidate !== null && candidate !== undefined);
  
  if (messagesCandidates.length !== 1) {
    throw new Error('Exactly one of messages, openai_messages, anthropic_messages, openai_responses_input must be provided');
  }

  if (data.anthropic_messages && !data.anthropic_system_prompt) {
    throw new Error('anthropic_system_prompt must be provided if anthropic_messages is provided');
  }

  const completionCandidates = [
    data.completion,
    data.openai_responses_output,
  ].filter(candidate => candidate !== null && candidate !== undefined);
  
  if (completionCandidates.length !== 1) {
    throw new Error('Exactly one of completion, openai_responses_output must be provided');
  }

  return true;
});
export type LogExtractionRequest = z.infer<typeof LogExtractionRequestSchema>;

// Log Extraction Response
export const LogExtractionResponseSchema = z.object({
  extraction_id: z.string().nullable().optional(),
  status: z.enum(['success', 'error']),
  error_message: z.string().nullable().optional(),
});
export type LogExtractionResponse = z.infer<typeof LogExtractionResponseSchema>;

// Streaming types
export const RetabParsedChoiceDeltaChunkSchema = z.object({
  content: z.string().nullable().optional(),
  flat_likelihoods: z.record(z.number()).default({}),
  flat_parsed: z.record(z.any()).default({}),
  flat_deleted_keys: z.array(z.string()).default([]),
  field_locations: z.record(z.array(FieldLocationSchema)).nullable().optional(),
  is_valid_json: z.boolean().default(false),
  key_mapping: z.record(z.string().nullable()).nullable().optional(),
});
export type RetabParsedChoiceDeltaChunk = z.infer<typeof RetabParsedChoiceDeltaChunkSchema>;

export const RetabParsedChoiceChunkSchema = z.object({
  delta: RetabParsedChoiceDeltaChunkSchema,
  index: z.number().int(),
  finish_reason: z.enum(['stop', 'length', 'tool_calls', 'content_filter', 'function_call']).nullable().optional(),
});
export type RetabParsedChoiceChunk = z.infer<typeof RetabParsedChoiceChunkSchema>;

export const RetabParsedChatCompletionChunkSchema = z.object({
  id: z.string(),
  object: z.literal('chat.completion.chunk'),
  created: z.number().int(),
  model: z.string(),
  choices: z.array(RetabParsedChoiceChunkSchema),
  usage: UsageSchema.optional(),
  extraction_id: z.string().nullable().optional(),
  schema_validation_error: z.custom<ErrorDetail>().nullable().optional(),
  request_at: z.string().datetime().nullable().optional(),
  first_token_at: z.string().datetime().nullable().optional(),
  last_token_at: z.string().datetime().nullable().optional(),
  streaming_error: z.custom<ErrorDetail>().nullable().optional(),
}).transform((data) => ({
  ...data,
  get api_cost(): Amount | null {
    // Implementation would compute cost from model and usage
    return null;
  },
  chunk_accumulator(_previous?: RetabParsedChatCompletionChunk): RetabParsedChatCompletionChunk {
    // Implementation would accumulate chunks
    return data as any;
  },
}));
export type RetabParsedChatCompletionChunk = z.infer<typeof RetabParsedChatCompletionChunkSchema>;