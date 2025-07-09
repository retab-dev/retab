/**
 * Specialized job types for specific workflows
 */

import { BaseJobResponse, JobCreate } from './base.js';

// Batch Annotation Job
export interface BatchAnnotationJobData {
  documents: Array<{
    id: string;
    path: string;
    mime_type: string;
  }>;
  schema_id: string;
  annotation_parameters: {
    model: string;
    temperature: number;
    modality: 'native' | 'text';
    reasoning_effort: 'low' | 'medium' | 'high';
    provider: 'openai' | 'anthropic' | 'xai' | 'gemini';
    batch_size?: number;
    max_concurrent?: number;
  };
  output_format: 'jsonl' | 'json' | 'csv';
  notification_config?: {
    webhook_url?: string;
    email_recipients?: string[];
  };
}

export interface BatchAnnotationJob extends BaseJobResponse {
  job_type: 'batch_annotation';
  job_data: BatchAnnotationJobData;
  results?: {
    total_documents: number;
    successful_annotations: number;
    failed_annotations: number;
    output_file_path?: string;
    error_file_path?: string;
    total_tokens_used: number;
    total_cost_usd: number;
    average_confidence_score: number;
  };
}

// Evaluation Job
export interface EvaluationJobData {
  dataset_path: string;
  ground_truth_path: string;
  models_to_evaluate: string[];
  evaluation_metrics: Array<'accuracy' | 'f1' | 'precision' | 'recall' | 'levenshtein' | 'jaccard'>;
  schema_id: string;
  test_parameters: {
    temperature: number;
    reasoning_effort: 'low' | 'medium' | 'high';
    max_samples?: number;
    random_seed?: number;
  };
  output_format: 'json' | 'csv' | 'html_report';
}

export interface EvaluationJob extends BaseJobResponse {
  job_type: 'evaluation';
  job_data: EvaluationJobData;
  results?: {
    models_evaluated: number;
    total_test_samples: number;
    benchmark_results: Array<{
      model: string;
      accuracy: number;
      f1_score: number;
      precision: number;
      recall: number;
      execution_time_ms: number;
      cost_usd: number;
    }>;
    best_performing_model: string;
    report_file_path?: string;
  };
}

// Fine-tuning Job
export interface FinetuneJobData {
  training_dataset_path: string;
  validation_dataset_path?: string;
  base_model: string;
  provider: 'openai' | 'anthropic';
  hyperparameters: {
    learning_rate?: number;
    batch_size?: number;
    epochs?: number;
    warmup_steps?: number;
    weight_decay?: number;
  };
  model_name: string;
  description?: string;
  tags?: string[];
}

export interface FinetuneJob extends BaseJobResponse {
  job_type: 'finetune';
  job_data: FinetuneJobData;
  results?: {
    fine_tuned_model_id: string;
    training_samples_processed: number;
    validation_samples_processed: number;
    final_training_loss: number;
    final_validation_loss: number;
    training_time_minutes: number;
    cost_usd: number;
    model_performance_metrics?: Record<string, number>;
  };
}

// Prompt Optimization Job
export interface PromptOptimizationJobData {
  base_prompt: string;
  test_dataset_path: string;
  optimization_objective: 'accuracy' | 'speed' | 'cost' | 'balanced';
  schema_id: string;
  model: string;
  optimization_parameters: {
    max_iterations: number;
    population_size: number;
    mutation_rate: number;
    convergence_threshold: number;
    techniques: Array<'clarity' | 'brevity' | 'structure' | 'examples' | 'constraints'>;
  };
  evaluation_metrics: string[];
}

export interface PromptOptimizationJob extends BaseJobResponse {
  job_type: 'prompt_optimization';
  job_data: PromptOptimizationJobData;
  results?: {
    optimized_prompt: string;
    improvement_percentage: number;
    iterations_completed: number;
    baseline_performance: Record<string, number>;
    optimized_performance: Record<string, number>;
    optimization_history: Array<{
      iteration: number;
      prompt: string;
      performance: Record<string, number>;
    }>;
  };
}

// Web Crawl Job
export interface WebcrawlJobData {
  start_urls: string[];
  crawl_parameters: {
    max_depth: number;
    max_pages: number;
    respect_robots_txt: boolean;
    delay_seconds: number;
    user_agent: string;
    include_patterns?: string[];
    exclude_patterns?: string[];
    javascript_enabled: boolean;
    screenshot_enabled: boolean;
  };
  extraction_config?: {
    schema_id: string;
    extract_text: boolean;
    extract_images: boolean;
    extract_links: boolean;
    extract_metadata: boolean;
  };
  storage_config: {
    store_html: boolean;
    store_screenshots: boolean;
    store_extracted_data: boolean;
    output_format: 'jsonl' | 'json' | 'csv';
  };
}

export interface WebcrawlJob extends BaseJobResponse {
  job_type: 'webcrawl';
  job_data: WebcrawlJobData;
  results?: {
    pages_crawled: number;
    pages_processed: number;
    pages_failed: number;
    total_size_mb: number;
    crawl_duration_minutes: number;
    output_file_paths: string[];
    discovered_urls: number;
    robots_txt_blocked: number;
    duplicate_pages_skipped: number;
  };
}

// Job type unions for type safety
export type SpecializedJob = 
  | BatchAnnotationJob
  | EvaluationJob 
  | FinetuneJob
  | PromptOptimizationJob
  | WebcrawlJob;

export type SpecializedJobData = 
  | BatchAnnotationJobData
  | EvaluationJobData
  | FinetuneJobData
  | PromptOptimizationJobData
  | WebcrawlJobData;

// Factory functions for creating specialized job requests
export const createBatchAnnotationJob = (
  organization_id: string,
  job_data: BatchAnnotationJobData,
  options?: Partial<JobCreate>
): JobCreate & { job_data: BatchAnnotationJobData } => ({
  organization_id,
  job_type: 'batch_annotation',
  priority: 'normal',
  max_retries: 3,
  timeout_seconds: 3600, // 1 hour
  ...options,
  job_data,
});

export const createEvaluationJob = (
  organization_id: string,
  job_data: EvaluationJobData,
  options?: Partial<JobCreate>
): JobCreate & { job_data: EvaluationJobData } => ({
  organization_id,
  job_type: 'evaluation',
  priority: 'normal',
  max_retries: 2,
  timeout_seconds: 7200, // 2 hours
  ...options,
  job_data,
});

export const createFinetuneJob = (
  organization_id: string,
  job_data: FinetuneJobData,
  options?: Partial<JobCreate>
): JobCreate & { job_data: FinetuneJobData } => ({
  organization_id,
  job_type: 'finetune',
  priority: 'high',
  max_retries: 1,
  timeout_seconds: 86400, // 24 hours
  ...options,
  job_data,
});

// Factory functions are fine to export by default as they are values
export default {
  createBatchAnnotationJob,
  createEvaluationJob,
  createFinetuneJob,
};