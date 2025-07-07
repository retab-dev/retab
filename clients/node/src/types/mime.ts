import { z } from 'zod';


// Core MIMEData schema
export const MIMEDataSchema = z.object({
  id: z.string(),
  extension: z.string(),
  content: z.string(),
  mime_type: z.string(),
  unique_filename: z.string(),
  size: z.number(),
  filename: z.string(),
  url: z.string(),
});

export type MIMEData = z.infer<typeof MIMEDataSchema>;

// Simplified file data that gets transformed to MIMEData
export const SimpleMIMEDataSchema = z.object({
  filename: z.string(),
  url: z.string(),
});

// OCR MIMEData schema
export const OCRMIMEDataSchema = MIMEDataSchema.extend({
  ocr_text: z.string().optional(),
});

export type OCRMIMEData = z.infer<typeof OCRMIMEDataSchema>;

// Generic document schema
export const GenericDocumentSchema = z.object({
  id: z.string(),
  tree_id: z.string(),
  subject: z.string().optional(),
  body_plain: z.string().optional(),
  body_html: z.string().optional(),
  from_email: z.string().optional(),
  from_name: z.string().optional(),
  to_email: z.string().optional(),
  to_name: z.string().optional(),
  cc_email: z.string().optional(),
  bcc_email: z.string().optional(),
  reply_to: z.string().optional(),
  in_reply_to: z.string().optional(),
  date: z.string().optional(),
  message_id: z.string().optional(),
  attachments: z.array(MIMEDataSchema).default([]),
});

export type GenericDocument = z.infer<typeof GenericDocumentSchema>;

// Email-specific MIME data
export const EmailMIMEDataSchema = GenericDocumentSchema;
export type EmailMIMEData = z.infer<typeof EmailMIMEDataSchema>;

// Union type for all MIME data types
export const AnyMIMEDataSchema = z.union([
  MIMEDataSchema,
  OCRMIMEDataSchema,
  EmailMIMEDataSchema,
]);

export type AnyMIMEData = z.infer<typeof AnyMIMEDataSchema>;