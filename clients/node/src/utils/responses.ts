/**
 * Response processing utilities
 * Equivalent to Python's utils/responses.py
 */

export interface ResponseMetadata {
  request_id: string;
  timestamp: Date;
  processing_time_ms: number;
  model_used: string;
  provider: string;
  tokens_used: {
    input: number;
    output: number;
    total: number;
  };
  cost_usd?: number;
}

export interface ProcessedResponse<T = any> {
  data: T;
  metadata: ResponseMetadata;
  raw_response?: any;
  validation_errors?: string[];
  confidence_score?: number;
}

/**
 * Process raw API response and extract metadata
 */
export function processResponse<T>(rawResponse: any): ProcessedResponse<T> {
  return {
    data: extractResponseData(rawResponse),
    metadata: extractResponseMetadata(rawResponse),
  };
}

function extractResponseData(rawResponse: any): any {
  if (rawResponse.choices && rawResponse.choices.length > 0) {
    return rawResponse.choices[0].message?.content || rawResponse.choices[0].text;
  }
  return rawResponse.content || rawResponse.data || rawResponse;
}

function extractResponseMetadata(rawResponse: any): ResponseMetadata {
  return {
    request_id: rawResponse.id || 'unknown',
    timestamp: new Date(),
    processing_time_ms: rawResponse.processing_time_ms || 0,
    model_used: rawResponse.model || 'unknown',
    provider: 'unknown',
    tokens_used: {
      input: rawResponse.usage?.prompt_tokens || 0,
      output: rawResponse.usage?.completion_tokens || 0,
      total: rawResponse.usage?.total_tokens || 0,
    },
  };
}

export default {
  processResponse,
  extractResponseData,
};