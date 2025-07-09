import fs from 'fs';
import axios from 'axios';
import { readJSONL, writeJSONL } from './jsonl.js';

/**
 * Batch processing utilities for OpenAI Batch API and other providers
 * Equivalent to Python's batch processing functionality
 */

export interface BatchRequest {
  custom_id: string;
  method: 'POST' | 'GET' | 'PUT' | 'DELETE';
  url: string;
  body?: Record<string, any>;
  headers?: Record<string, string>;
}

export interface BatchResponse {
  id: string;
  custom_id: string;
  response: {
    status_code: number;
    request_id: string;
    body: Record<string, any>;
  };
  error?: {
    code: string;
    message: string;
  };
}

export interface BatchJob {
  id: string;
  object: 'batch';
  endpoint: string;
  errors?: {
    object: 'list';
    data: Array<{
      code: string;
      message: string;
      param?: string;
      line?: number;
    }>;
  };
  input_file_id: string;
  completion_window: '24h';
  status: 'validating' | 'failed' | 'in_progress' | 'finalizing' | 'completed' | 'expired' | 'cancelling' | 'cancelled';
  output_file_id?: string;
  error_file_id?: string;
  created_at: number;
  in_progress_at?: number;
  expires_at?: number;
  finalizing_at?: number;
  completed_at?: number;
  failed_at?: number;
  expired_at?: number;
  cancelling_at?: number;
  cancelled_at?: number;
  request_counts: {
    total: number;
    completed: number;
    failed: number;
  };
  metadata?: Record<string, string>;
}

export interface BatchProcessingOptions {
  apiKey: string;
  baseUrl?: string;
  timeout?: number;
  maxRetries?: number;
  completionWindow?: '24h';
  metadata?: Record<string, string>;
}

export interface BatchProgressInfo {
  jobId: string;
  status: string;
  progress: {
    total: number;
    completed: number;
    failed: number;
    percentage: number;
  };
  timeElapsed: number;
  estimatedTimeRemaining?: number;
}

/**
 * OpenAI Batch API client
 */
export class OpenAIBatchProcessor {
  private apiKey: string;
  private baseUrl: string;
  private timeout: number;
  constructor(options: BatchProcessingOptions) {
    this.apiKey = options.apiKey;
    this.baseUrl = options.baseUrl || 'https://api.openai.com/v1';
    this.timeout = options.timeout || 300000; // 5 minutes
  }

  /**
   * Upload file for batch processing
   */
  async uploadFile(filePath: string, purpose: 'batch' = 'batch'): Promise<{ id: string; filename: string }> {
    if (!fs.existsSync(filePath)) {
      throw new Error(`File not found: ${filePath}`);
    }

    const formData = new FormData();
    const fileBuffer = fs.readFileSync(filePath);
    const blob = new Blob([fileBuffer]);
    
    formData.append('file', blob, filePath.split('/').pop());
    formData.append('purpose', purpose);

    const response = await axios.post(`${this.baseUrl}/files`, formData, {
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'multipart/form-data',
      },
      timeout: this.timeout,
    });

    return {
      id: response.data.id,
      filename: response.data.filename,
    };
  }

  /**
   * Create batch job
   */
  async createBatch(
    inputFileId: string,
    endpoint: string,
    completionWindow: '24h' = '24h',
    metadata?: Record<string, string>
  ): Promise<BatchJob> {
    const response = await axios.post(
      `${this.baseUrl}/batches`,
      {
        input_file_id: inputFileId,
        endpoint,
        completion_window: completionWindow,
        metadata,
      },
      {
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json',
        },
        timeout: this.timeout,
      }
    );

    return response.data;
  }

  /**
   * Get batch job status
   */
  async getBatch(batchId: string): Promise<BatchJob> {
    const response = await axios.get(`${this.baseUrl}/batches/${batchId}`, {
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
      },
      timeout: this.timeout,
    });

    return response.data;
  }

  /**
   * Cancel batch job
   */
  async cancelBatch(batchId: string): Promise<BatchJob> {
    const response = await axios.post(
      `${this.baseUrl}/batches/${batchId}/cancel`,
      {},
      {
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json',
        },
        timeout: this.timeout,
      }
    );

    return response.data;
  }

  /**
   * List batch jobs
   */
  async listBatches(after?: string, limit: number = 20): Promise<{ data: BatchJob[] }> {
    const params = new URLSearchParams();
    if (after) params.append('after', after);
    params.append('limit', limit.toString());

    const response = await axios.get(`${this.baseUrl}/batches?${params}`, {
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
      },
      timeout: this.timeout,
    });

    return response.data;
  }

  /**
   * Download file content
   */
  async downloadFile(fileId: string): Promise<string> {
    const response = await axios.get(`${this.baseUrl}/files/${fileId}/content`, {
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
      },
      timeout: this.timeout,
    });

    return response.data;
  }

  /**
   * Monitor batch job progress
   */
  async *monitorBatch(batchId: string, pollInterval: number = 30000): AsyncGenerator<BatchProgressInfo, void, unknown> {
    const startTime = Date.now();
    let lastProgress = 0;
    
    while (true) {
      const batch = await this.getBatch(batchId);
      const timeElapsed = Date.now() - startTime;
      
      const progress = {
        total: batch.request_counts.total,
        completed: batch.request_counts.completed,
        failed: batch.request_counts.failed,
        percentage: batch.request_counts.total > 0 ? 
          (batch.request_counts.completed / batch.request_counts.total) * 100 : 0,
      };

      // Estimate time remaining
      let estimatedTimeRemaining: number | undefined;
      if (progress.completed > lastProgress && progress.completed > 0) {
        const completionRate = progress.completed / timeElapsed;
        const remaining = progress.total - progress.completed;
        estimatedTimeRemaining = remaining / completionRate;
      }
      
      yield {
        jobId: batchId,
        status: batch.status,
        progress,
        timeElapsed,
        estimatedTimeRemaining,
      };

      // Break if job is complete
      if (['completed', 'failed', 'cancelled', 'expired'].includes(batch.status)) {
        break;
      }

      lastProgress = progress.completed;
      await new Promise(resolve => setTimeout(resolve, pollInterval));
    }
  }

  /**
   * Process batch end-to-end
   */
  async processBatch(
    requestsFilePath: string,
    outputFilePath: string,
    endpoint: string = '/v1/chat/completions',
    options: {
      completionWindow?: '24h';
      metadata?: Record<string, string>;
      pollInterval?: number;
      showProgress?: boolean;
    } = {}
  ): Promise<BatchJob> {
    const {
      completionWindow = '24h',
      metadata,
      pollInterval = 30000,
      showProgress = true,
    } = options;

    console.log('🚀 Starting batch processing...');
    
    // Upload input file
    console.log('📤 Uploading input file...');
    const uploadResult = await this.uploadFile(requestsFilePath);
    console.log(`✅ File uploaded: ${uploadResult.id}`);

    // Create batch job
    console.log('⚙️  Creating batch job...');
    const batch = await this.createBatch(uploadResult.id, endpoint, completionWindow, metadata);
    console.log(`✅ Batch created: ${batch.id}`);

    // Monitor progress
    if (showProgress) {
      console.log('📊 Monitoring batch progress...');
      for await (const progress of this.monitorBatch(batch.id, pollInterval)) {
        const { percentage, completed, total } = progress.progress;
        const timeStr = `${Math.round(progress.timeElapsed / 1000)}s`;
        const etaStr = progress.estimatedTimeRemaining ? 
          ` (ETA: ${Math.round(progress.estimatedTimeRemaining / 1000)}s)` : '';
        
        console.log(`   Status: ${progress.status} - ${percentage.toFixed(1)}% (${completed}/${total}) - ${timeStr}${etaStr}`);
        
        if (['completed', 'failed', 'cancelled', 'expired'].includes(progress.status)) {
          break;
        }
      }
    }

    // Get final batch status
    const finalBatch = await this.getBatch(batch.id);

    if (finalBatch.status === 'completed' && finalBatch.output_file_id) {
      console.log('📥 Downloading results...');
      const results = await this.downloadFile(finalBatch.output_file_id);
      fs.writeFileSync(outputFilePath, results);
      console.log(`✅ Results saved to: ${outputFilePath}`);
    } else {
      console.error(`❌ Batch failed with status: ${finalBatch.status}`);
      if (finalBatch.error_file_id) {
        const errors = await this.downloadFile(finalBatch.error_file_id);
        console.error('Error details:', errors);
      }
    }

    return finalBatch;
  }
}

/**
 * Utility functions for batch processing
 */
export const batchUtils = {
  /**
   * Create batch requests for chat completions
   */
  createChatCompletionRequests: (
    messages: Array<{ messages: any[]; customId?: string }>,
    model: string = 'gpt-4o-mini',
    options: {
      temperature?: number;
      maxTokens?: number;
      responseFormat?: { type: string };
    } = {}
  ): BatchRequest[] => {
    return messages.map((msg, index) => ({
      custom_id: msg.customId || `request_${index}`,
      method: 'POST' as const,
      url: '/v1/chat/completions',
      body: {
        model,
        messages: msg.messages,
        temperature: options.temperature || 0.0,
        max_tokens: options.maxTokens,
        response_format: options.responseFormat,
      },
    }));
  },

  /**
   * Create batch requests for document extraction
   */
  createExtractionRequests: (
    documents: Array<{ document: any; schema: any; customId?: string }>,
    model: string = 'gpt-4o-mini'
  ): BatchRequest[] => {
    return documents.map((doc, index) => ({
      custom_id: doc.customId || `extraction_${index}`,
      method: 'POST' as const,
      url: '/v1/documents/extractions',
      body: {
        json_schema: doc.schema,
        documents: [doc.document],
        model,
      },
    }));
  },

  /**
   * Save batch requests to JSONL file
   */
  saveBatchRequests: async (requests: BatchRequest[], filePath: string): Promise<void> => {
    await writeJSONL(filePath, requests);
    console.log(`📄 Saved ${requests.length} batch requests to ${filePath}`);
  },

  /**
   * Parse batch results from JSONL file
   */
  parseBatchResults: async (filePath: string): Promise<BatchResponse[]> => {
    const results = await readJSONL(filePath);
    return results as BatchResponse[];
  },

  /**
   * Extract successful results from batch responses
   */
  extractSuccessfulResults: (responses: BatchResponse[]): Array<{ customId: string; result: any }> => {
    return responses
      .filter(response => !response.error && response.response.status_code === 200)
      .map(response => ({
        customId: response.custom_id,
        result: response.response.body,
      }));
  },

  /**
   * Extract failed results from batch responses
   */
  extractFailedResults: (responses: BatchResponse[]): Array<{ customId: string; error: any }> => {
    return responses
      .filter(response => response.error || response.response.status_code !== 200)
      .map(response => ({
        customId: response.custom_id,
        error: response.error || { code: 'http_error', message: `HTTP ${response.response.status_code}` },
      }));
  },
};

export default {
  OpenAIBatchProcessor,
  batchUtils,
};