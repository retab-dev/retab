import { z } from 'zod';
import { MIMEDataSchema } from '../mime.js';
import { ModalitySchema } from '../modalities.js';
import { BrowserCanvasSchema } from '../browser_canvas.js';
import { ChatCompletionReasoningEffortSchema } from '../documents/extractions.js';

export const GenerateSchemaRequestSchema = z.object({
  documents: z.array(MIMEDataSchema),
  model: z.string().default('gpt-4o-mini'),
  temperature: z.number().default(0.0),
  reasoning_effort: ChatCompletionReasoningEffortSchema.default('medium'),
  modality: ModalitySchema,
  instructions: z.string().nullable().optional(),
  image_resolution_dpi: z.number().int().default(96),
  browser_canvas: BrowserCanvasSchema.default('A4'),
  stream: z.boolean().default(false),
});
export type GenerateSchemaRequest = z.infer<typeof GenerateSchemaRequestSchema>;

export const GenerateSystemPromptRequestSchema = GenerateSchemaRequestSchema.extend({
  json_schema: z.record(z.any()),
});
export type GenerateSystemPromptRequest = z.infer<typeof GenerateSystemPromptRequestSchema>;