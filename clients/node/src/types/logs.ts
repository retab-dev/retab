import { BrowserCanvas } from './browser_canvas.js';
import { Modality } from './modalities.js';
import { ListMetadata } from './pagination.js';

export interface ProcessorConfig {
  object: string;
  id: string;
  updated_at: string;
  name: string;
  modality: Modality;
  image_resolution_dpi: number;
  browser_canvas: BrowserCanvas;
  model: string;
  json_schema: Record<string, any>;
  temperature: number;
  reasoning_effort: 'low' | 'medium' | 'high';
  n_consensus: number;
  schema_data_id: string;
  schema_id: string;
}

export interface AutomationConfig {
  object: string;
  id: string;
  name: string;
  processor_id: string;
  updated_at: string;
  default_language: string;
  webhook_url: string;
  webhook_headers: Record<string, string>;
  need_validation: boolean;
}

export interface UpdateProcessorRequest {
  name?: string;
  modality?: Modality;
  image_resolution_dpi?: number;
  browser_canvas?: BrowserCanvas;
  model?: string;
  json_schema?: Record<string, any>;
  temperature?: number;
  reasoning_effort?: 'low' | 'medium' | 'high';
  n_consensus?: number;
  schema_data_id?: string;
  schema_id?: string;
}

export interface UpdateAutomationRequest {
  name?: string;
  default_language?: string;
  webhook_url?: string;
  webhook_headers?: Record<string, string>;
  need_validation?: boolean;
}

export interface OpenAIRequestConfig {
  object: 'openai_request';
  id: string;
  model: string;
  json_schema: Record<string, any>;
  reasoning_effort?: 'low' | 'medium' | 'high';
}

export interface ExternalRequestLog {
  webhook_url?: string;
  request_body: Record<string, any>;
  request_headers: Record<string, string>;
  request_at: string;
  response_body: Record<string, any>;
  response_headers: Record<string, string>;
  response_at: string;
  status_code: number;
  error?: string;
  duration_ms: number;
}

export interface LogCompletionRequest {
  json_schema: Record<string, any>;
  completion: any; // ChatCompletion type
}

export interface AutomationLog {
  object: 'automation_log';
  id: string;
  user_email?: string;
  organization_id: string;
  created_at: string;
  automation_snapshot: AutomationConfig;
  completion: any; // RetabParsedChatCompletion | ChatCompletion
  file_metadata?: any; // BaseMIMEData
  external_request_log?: ExternalRequestLog;
  extraction_id?: string;
  api_cost?: any; // Amount
  cost_breakdown?: any; // CostBreakdown
}

export interface ListLogs {
  data: AutomationLog[];
  list_metadata: ListMetadata;
}