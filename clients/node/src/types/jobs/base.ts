/**
 * Base job system types for workflow orchestration
 * Equivalent to Python's types/jobs/ modules
 */

export interface BaseJobResponse {
  id: string;
  organization_id: string;
  job_type: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled' | 'paused';
  created_at: Date;
  updated_at: Date;
  started_at?: Date;
  completed_at?: Date;
  created_by?: string;
  priority: 'low' | 'normal' | 'high' | 'urgent';
  progress_percentage: number;
  error_message?: string;
  error_code?: string;
  retry_count: number;
  max_retries: number;
  timeout_seconds?: number;
  metadata?: Record<string, any>;
  dependencies?: string[]; // IDs of jobs this job depends on
  result_data?: Record<string, any>;
  resource_usage?: {
    cpu_seconds?: number;
    memory_mb_peak?: number;
    storage_mb_used?: number;
    network_mb_transferred?: number;
  };
  estimated_duration_seconds?: number;
  actual_duration_seconds?: number;
}

export interface JobCreate {
  organization_id: string;
  job_type: string;
  created_by?: string;
  priority?: 'low' | 'normal' | 'high' | 'urgent';
  timeout_seconds?: number;
  max_retries?: number;
  metadata?: Record<string, any>;
  dependencies?: string[];
  estimated_duration_seconds?: number;
}

export interface JobUpdate {
  status?: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled' | 'paused';
  progress_percentage?: number;
  error_message?: string;
  error_code?: string;
  result_data?: Record<string, any>;
  metadata?: Record<string, any>;
  resource_usage?: {
    cpu_seconds?: number;
    memory_mb_peak?: number;
    storage_mb_used?: number;
    network_mb_transferred?: number;
  };
  actual_duration_seconds?: number;
}

export interface JobQuery {
  organization_id?: string;
  job_type?: string;
  status?: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled' | 'paused';
  created_by?: string;
  priority?: 'low' | 'normal' | 'high' | 'urgent';
  created_after?: Date;
  created_before?: Date;
  completed_after?: Date;
  completed_before?: Date;
  has_dependencies?: boolean;
  dependency_of?: string; // Jobs that depend on this job ID
  limit?: number;
  offset?: number;
  order_by?: 'created_at' | 'updated_at' | 'priority' | 'progress_percentage';
  order_direction?: 'asc' | 'desc';
}

export interface JobStats {
  total_jobs: number;
  pending_jobs: number;
  running_jobs: number;
  completed_jobs: number;
  failed_jobs: number;
  cancelled_jobs: number;
  average_duration_seconds: number;
  total_cpu_seconds_used: number;
  total_memory_mb_peak: number;
  jobs_by_type: Record<string, number>;
  jobs_by_priority: Record<string, number>;
  success_rate_percentage: number;
  throughput_jobs_per_hour: number;
  queue_depth_by_priority: Record<string, number>;
}

export interface WorkflowDefinition {
  id: string;
  name: string;
  description?: string;
  version: string;
  organization_id: string;
  created_at: Date;
  updated_at: Date;
  created_by?: string;
  is_active: boolean;
  steps: WorkflowStep[];
  triggers: WorkflowTrigger[];
  variables?: Record<string, any>;
  timeout_seconds?: number;
  max_concurrent_executions?: number;
}

export interface WorkflowStep {
  id: string;
  name: string;
  job_type: string;
  depends_on?: string[]; // Step IDs this step depends on
  parameters?: Record<string, any>;
  timeout_seconds?: number;
  max_retries?: number;
  on_failure?: 'fail_workflow' | 'continue' | 'retry' | 'skip';
  conditions?: WorkflowCondition[];
}

export interface WorkflowTrigger {
  type: 'manual' | 'schedule' | 'webhook' | 'file_upload' | 'job_completion';
  configuration: Record<string, any>;
  is_active: boolean;
}

export interface WorkflowCondition {
  field: string;
  operator: 'equals' | 'not_equals' | 'greater_than' | 'less_than' | 'contains' | 'exists';
  value: any;
  logical_operator?: 'and' | 'or';
}

export interface WorkflowExecution {
  id: string;
  workflow_id: string;
  organization_id: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled' | 'paused';
  created_at: Date;
  updated_at: Date;
  started_at?: Date;
  completed_at?: Date;
  triggered_by: string;
  trigger_data?: Record<string, any>;
  input_variables?: Record<string, any>;
  output_variables?: Record<string, any>;
  current_step_id?: string;
  completed_steps: string[];
  failed_steps: string[];
  step_executions: Record<string, JobStepExecution>;
  error_message?: string;
  progress_percentage: number;
}

export interface JobStepExecution {
  step_id: string;
  job_id?: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled' | 'skipped';
  started_at?: Date;
  completed_at?: Date;
  input_data?: Record<string, any>;
  output_data?: Record<string, any>;
  error_message?: string;
  retry_count: number;
}

export default {
  BaseJobResponse,
  JobCreate,
  JobUpdate,
  JobQuery,
  JobStats,
  WorkflowDefinition,
  WorkflowStep,
  WorkflowTrigger,
  WorkflowCondition,
  WorkflowExecution,
  JobStepExecution,
};