import { z } from 'zod';

export const BaseModalitySchema = z.enum(['text', 'image']);
export type BaseModality = z.infer<typeof BaseModalitySchema>;

export const ModalitySchema = z.union([
  BaseModalitySchema,
  z.literal('native'),
  z.literal('image+text'),
]);
export type Modality = z.infer<typeof ModalitySchema>;

export const TypeFamiliesSchema = z.enum([
  'excel',
  'word',
  'powerpoint',
  'pdf',
  'image',
  'text',
  'email',
  'audio',
  'html',
  'web',
]);
export type TypeFamilies = z.infer<typeof TypeFamiliesSchema>;

export const NativeModalities: Record<TypeFamilies, Modality> = {
  excel: 'image',
  word: 'image',
  html: 'text',
  powerpoint: 'image',
  pdf: 'image',
  image: 'image',
  web: 'image',
  text: 'text',
  email: 'native',
  audio: 'text',
};

// File type literals
export const ExcelTypesSchema = z.enum(['.xls', '.xlsx', '.ods']);
export type ExcelTypes = z.infer<typeof ExcelTypesSchema>;

export const WordTypesSchema = z.enum(['.doc', '.docx', '.odt']);
export type WordTypes = z.infer<typeof WordTypesSchema>;

export const PPTTypesSchema = z.enum(['.ppt', '.pptx', '.odp']);
export type PPTTypes = z.infer<typeof PPTTypesSchema>;

export const PDFTypesSchema = z.enum(['.pdf']);
export type PDFTypes = z.infer<typeof PDFTypesSchema>;

export const ImageTypesSchema = z.enum([
  '.jpg',
  '.jpeg',
  '.png',
  '.gif',
  '.bmp',
  '.tiff',
  '.webp',
]);
export type ImageTypes = z.infer<typeof ImageTypesSchema>;

export const TextTypesSchema = z.enum([
  '.txt',
  '.csv',
  '.tsv',
  '.md',
  '.log',
  '.xml',
  '.json',
  '.yaml',
  '.yml',
  '.rtf',
  '.ini',
  '.conf',
  '.cfg',
  '.nfo',
  '.srt',
  '.sql',
  '.sh',
  '.bat',
  '.ps1',
  '.js',
  '.jsx',
  '.ts',
  '.tsx',
  '.py',
  '.java',
  '.c',
  '.cpp',
  '.cs',
  '.rb',
  '.php',
  '.swift',
  '.kt',
  '.go',
  '.rs',
  '.pl',
  '.r',
  '.m',
  '.scala',
]);
export type TextTypes = z.infer<typeof TextTypesSchema>;

export const HTMLTypesSchema = z.enum(['.html', '.htm']);
export type HTMLTypes = z.infer<typeof HTMLTypesSchema>;

export const WebTypesSchema = z.enum(['.mhtml']);
export type WebTypes = z.infer<typeof WebTypesSchema>;

export const EmailTypesSchema = z.enum(['.eml', '.msg']);
export type EmailTypes = z.infer<typeof EmailTypesSchema>;

export const AudioTypesSchema = z.enum([
  '.mp3',
  '.mp4',
  '.mpeg',
  '.mpga',
  '.m4a',
  '.wav',
  '.webm',
]);
export type AudioTypes = z.infer<typeof AudioTypesSchema>;

export const SupportedTypesSchema = z.union([
  ExcelTypesSchema,
  WordTypesSchema,
  PPTTypesSchema,
  PDFTypesSchema,
  ImageTypesSchema,
  TextTypesSchema,
  HTMLTypesSchema,
  WebTypesSchema,
  EmailTypesSchema,
  AudioTypesSchema,
]);
export type SupportedTypes = z.infer<typeof SupportedTypesSchema>;