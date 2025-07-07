export interface Iteration {
  id: string;
  updated_at: string;
  inference_settings: any; // InferenceSettings
  json_schema: Record<string, any>;
  predictions: Record<string, any>; // PredictionData
  metric_results?: any; // MetricResult
  schema_data_id: string;
  schema_id: string;
}

export interface CreateIterationRequest {
  inference_settings: any; // InferenceSettings
  json_schema?: Record<string, any>;
  from_iteration_id?: string;
}

export interface PatchIterationRequest {
  inference_settings?: any; // InferenceSettings
  json_schema?: Record<string, any>;
  version?: number;
}

export interface ProcessIterationRequest {
  document_ids?: string[];
  only_outdated?: boolean;
}

export interface DocumentStatus {
  document_id: string;
  filename: string;
  needs_update: boolean;
  has_prediction: boolean;
  prediction_updated_at?: string;
  iteration_updated_at: string;
}

export interface IterationDocumentStatusResponse {
  iteration_id: string;
  documents: DocumentStatus[];
  total_documents: number;
  documents_needing_update: number;
  documents_up_to_date: number;
}

export interface AddIterationFromJsonlRequest {
  jsonl_gcs_path: string;
}