/**
 * Database types for annotations and ORM integration
 * Equivalent to Python's types/db/ modules
 */

export interface AnnotationParameters {
  model: string;
  temperature: number;
  modality: 'native' | 'text';
  reasoning_effort: 'low' | 'medium' | 'high';
  provider: 'openai' | 'anthropic' | 'xai' | 'gemini';
  image_resolution_dpi?: number;
  browser_canvas?: 'A3' | 'A4' | 'A5';
  max_tokens?: number;
  response_format?: {
    type: 'json_schema';
    json_schema: {
      name: string;
      schema: Record<string, any>;
      strict?: boolean;
    };
  };
}

export interface Annotation {
  id: string;
  organization_id: string;
  document_id: string;
  schema_id: string;
  annotation_data: Record<string, any>;
  parameters: AnnotationParameters;
  created_at: Date;
  updated_at: Date;
  created_by?: string;
  updated_by?: string;
  version: number;
  status: 'pending' | 'completed' | 'failed' | 'cancelled';
  error_message?: string;
  execution_time_ms?: number;
  token_usage?: {
    input_tokens: number;
    output_tokens: number;
    total_tokens: number;
    cost_usd?: number;
  };
  confidence_score?: number;
  metadata?: Record<string, any>;
}

export interface AnnotationCreate {
  organization_id: string;
  document_id: string;
  schema_id: string;
  annotation_data: Record<string, any>;
  parameters: AnnotationParameters;
  created_by?: string;
  metadata?: Record<string, any>;
}

export interface AnnotationUpdate {
  annotation_data?: Record<string, any>;
  parameters?: Partial<AnnotationParameters>;
  updated_by?: string;
  status?: 'pending' | 'completed' | 'failed' | 'cancelled';
  error_message?: string;
  execution_time_ms?: number;
  token_usage?: {
    input_tokens: number;
    output_tokens: number;
    total_tokens: number;
    cost_usd?: number;
  };
  confidence_score?: number;
  metadata?: Record<string, any>;
}

export interface AnnotationQuery {
  organization_id?: string;
  document_id?: string;
  schema_id?: string;
  status?: 'pending' | 'completed' | 'failed' | 'cancelled';
  created_by?: string;
  created_after?: Date;
  created_before?: Date;
  updated_after?: Date;
  updated_before?: Date;
  min_confidence?: number;
  max_confidence?: number;
  provider?: 'openai' | 'anthropic' | 'xai' | 'gemini';
  model?: string;
  limit?: number;
  offset?: number;
  order_by?: 'created_at' | 'updated_at' | 'confidence_score';
  order_direction?: 'asc' | 'desc';
}

export interface AnnotationStats {
  total_annotations: number;
  completed_annotations: number;
  failed_annotations: number;
  pending_annotations: number;
  cancelled_annotations: number;
  average_execution_time_ms: number;
  total_tokens_used: number;
  total_cost_usd: number;
  average_confidence_score: number;
  annotations_by_provider: Record<string, number>;
  annotations_by_model: Record<string, number>;
  annotations_by_day: Array<{
    date: string;
    count: number;
  }>;
}

export default {
  AnnotationParameters,
  Annotation,
  AnnotationCreate,
  AnnotationUpdate,
  AnnotationQuery,
  AnnotationStats,
};