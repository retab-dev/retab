import { z } from 'zod';

// Chat completion content part
export const ChatCompletionContentPartSchema = z.union([
  z.object({
    type: z.literal('text'),
    text: z.string(),
  }),
  z.object({
    type: z.literal('image_url'),
    image_url: z.object({
      url: z.string(),
      detail: z.enum(['low', 'high', 'auto']).optional(),
    }),
  }),
]);
export type ChatCompletionContentPart = z.infer<typeof ChatCompletionContentPartSchema>;

// Retab-specific message format
export const ChatCompletionRetabMessageSchema = z.object({
  role: z.enum(['user', 'system', 'assistant', 'developer']),
  content: z.union([z.string(), z.array(ChatCompletionContentPartSchema)]),
});
export type ChatCompletionRetabMessage = z.infer<typeof ChatCompletionRetabMessageSchema>;