/**
 * Schema enhancement and manipulation types
 * Equivalent to Python's types/schemas/ modules
 */

// Schema enhancement
export interface SchemaEnhancementConfig {
  enhancement_techniques: Array<'field_optimization' | 'constraint_improvement' | 'description_enhancement' | 'example_generation' | 'validation_strengthening'>;
  target_accuracy?: number; // 0-1
  max_iterations?: number;
  use_ground_truth: boolean;
  preserve_original_structure: boolean;
  add_field_descriptions: boolean;
  generate_examples: boolean;
  optimize_for_provider?: 'openai' | 'anthropic' | 'xai' | 'gemini';
  performance_weight: number; // 0-1, balance between accuracy and speed
}

export interface SchemaEnhancementRequest {
  base_schema: Record<string, any>;
  training_documents?: Array<{
    document: any;
    ground_truth: any;
  }>;
  performance_data?: Array<{
    model: string;
    accuracy: number;
    latency_ms: number;
    cost_per_request: number;
  }>;
  enhancement_config: SchemaEnhancementConfig;
}

export interface SchemaEnhancementResponse {
  enhanced_schema: Record<string, any>;
  enhancement_summary: {
    changes_made: Array<SchemaChange>;
    estimated_improvement: {
      accuracy_gain: number; // percentage points
      latency_change: number; // percentage change
      cost_change: number; // percentage change
    };
    compatibility: {
      backward_compatible: boolean;
      breaking_changes: string[];
    };
  };
  validation_results?: {
    test_accuracy: number;
    validation_errors: string[];
  };
}

export interface SchemaChange {
  type: 'field_added' | 'field_removed' | 'field_modified' | 'constraint_added' | 'description_updated' | 'example_added';
  field_path: string;
  old_value?: any;
  new_value?: any;
  reason: string;
  impact_score: number; // 0-1, estimated impact on accuracy
}

// Schema evaluation
export interface SchemaEvaluationConfig {
  evaluation_metrics: Array<'accuracy' | 'completeness' | 'consistency' | 'clarity' | 'efficiency'>;
  test_documents: Array<{
    document: any;
    expected_output: any;
  }>;
  models_to_test: Array<{
    model: string;
    provider: string;
    parameters?: Record<string, any>;
  }>;
  evaluation_parameters: {
    num_runs_per_test?: number;
    temperature?: number;
    max_tokens?: number;
    timeout_seconds?: number;
  };
}

export interface SchemaEvaluationRequest {
  schema: Record<string, any>;
  evaluation_config: SchemaEvaluationConfig;
}

export interface SchemaEvaluationResponse {
  overall_score: number; // 0-1
  metric_scores: {
    accuracy: number;
    completeness: number;
    consistency: number;
    clarity: number;
    efficiency: number;
  };
  model_performance: Array<ModelPerformance>;
  detailed_results: Array<TestResult>;
  recommendations: Array<SchemaRecommendation>;
  schema_analysis: SchemaAnalysis;
}

export interface ModelPerformance {
  model: string;
  provider: string;
  accuracy: number;
  average_latency_ms: number;
  cost_per_request_usd: number;
  success_rate: number;
  error_types: Record<string, number>;
}

export interface TestResult {
  test_id: string;
  document_summary: string;
  expected_output: any;
  model_outputs: Array<{
    model: string;
    output: any;
    latency_ms: number;
    success: boolean;
    error?: string;
    confidence_score?: number;
  }>;
  accuracy_scores: Record<string, number>; // model -> accuracy
  consensus_output?: any;
}

export interface SchemaRecommendation {
  type: 'performance' | 'accuracy' | 'cost' | 'compatibility';
  priority: 'low' | 'medium' | 'high' | 'critical';
  title: string;
  description: string;
  suggested_changes: Array<{
    field_path: string;
    current_value: any;
    suggested_value: any;
    rationale: string;
  }>;
  estimated_impact: {
    accuracy_change: number;
    cost_change: number;
    latency_change: number;
  };
}

// Schema layout optimization
export interface SchemaLayoutConfig {
  target_output_format: 'json' | 'xml' | 'yaml' | 'csv';
  optimize_for: 'readability' | 'processing_speed' | 'storage_efficiency' | 'api_compatibility';
  nesting_preference: 'flat' | 'nested' | 'balanced';
  field_naming_convention: 'camelCase' | 'snake_case' | 'PascalCase' | 'kebab-case';
  include_metadata_fields: boolean;
  add_version_field: boolean;
  add_timestamp_fields: boolean;
  max_nesting_depth?: number;
  required_fields_strategy: 'strict' | 'flexible' | 'minimal';
}

export interface SchemaLayoutRequest {
  base_schema: Record<string, any>;
  layout_config: SchemaLayoutConfig;
  sample_data?: any[];
}

export interface SchemaLayoutResponse {
  optimized_schema: Record<string, any>;
  layout_changes: Array<LayoutChange>;
  compatibility_report: {
    backward_compatible: boolean;
    migration_required: boolean;
    breaking_changes: string[];
    migration_guide?: string;
  };
  performance_estimate: {
    serialization_speed_change: number; // percentage
    storage_size_change: number; // percentage
    parsing_complexity_change: number; // percentage
  };
}

export interface LayoutChange {
  type: 'field_moved' | 'field_renamed' | 'structure_flattened' | 'structure_nested' | 'field_type_changed';
  field_path: string;
  old_location?: string;
  new_location?: string;
  old_name?: string;
  new_name?: string;
  reason: string;
}

// Schema templates
export interface SchemaTemplate {
  id: string;
  name: string;
  description: string;
  category: 'document_extraction' | 'data_analysis' | 'content_generation' | 'classification' | 'custom';
  version: string;
  author: string;
  created_at: Date;
  updated_at: Date;
  schema: Record<string, any>;
  example_inputs: any[];
  example_outputs: any[];
  use_cases: string[];
  tags: string[];
  complexity_level: 'beginner' | 'intermediate' | 'advanced' | 'expert';
  estimated_accuracy: number; // 0-1
  supported_providers: Array<'openai' | 'anthropic' | 'xai' | 'gemini'>;
  performance_benchmarks?: Array<{
    model: string;
    provider: string;
    accuracy: number;
    latency_ms: number;
    cost_per_request: number;
  }>;
}

export interface SchemaTemplateQuery {
  category?: string;
  complexity_level?: string;
  min_accuracy?: number;
  supported_provider?: string;
  tags?: string[];
  search_text?: string;
  limit?: number;
  offset?: number;
  order_by?: 'name' | 'created_at' | 'updated_at' | 'estimated_accuracy';
  order_direction?: 'asc' | 'desc';
}

// Schema analysis
export interface SchemaAnalysis {
  complexity_score: number; // 0-1
  field_count: number;
  nesting_depth: number;
  required_fields_count: number;
  optional_fields_count: number;
  field_types: Record<string, number>; // type -> count
  validation_rules_count: number;
  estimated_token_usage: {
    schema_tokens: number;
    response_tokens_avg: number;
    response_tokens_max: number;
  };
  potential_issues: Array<SchemaIssue>;
  optimization_opportunities: Array<OptimizationOpportunity>;
  provider_compatibility: Record<string, {
    compatible: boolean;
    issues?: string[];
    optimizations?: string[];
  }>;
}

export interface SchemaIssue {
  type: 'validation' | 'performance' | 'compatibility' | 'clarity';
  severity: 'low' | 'medium' | 'high' | 'critical';
  field_path?: string;
  message: string;
  suggestion?: string;
}

export interface OptimizationOpportunity {
  type: 'reduce_complexity' | 'improve_clarity' | 'enhance_validation' | 'optimize_performance';
  field_path?: string;
  description: string;
  estimated_benefit: {
    accuracy_improvement: number;
    speed_improvement: number;
    cost_reduction: number;
  };
  implementation_effort: 'low' | 'medium' | 'high';
}

// No default export for type-only modules