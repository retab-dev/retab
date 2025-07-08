import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import axiosRetry from 'axios-retry';
import { PreparedRequest, FieldUnset } from './types/standards.js';
import { MaxRetriesExceeded, APIError, ValidationError } from './errors.js';
import { 
  Consensus, AsyncConsensus,
  Documents, AsyncDocuments,
  Files, AsyncFiles,
  FineTuning, AsyncFineTuning,
  Models, AsyncModels,
  Processors, AsyncProcessors,
  Schemas, AsyncSchemas,
  Secrets, AsyncSecrets,
  Usage, AsyncUsage,
  Evaluations, AsyncEvaluations
} from './resources/index.js';

export interface RetabConfig {
  apiKey?: string;
  baseUrl?: string;
  timeout?: number;
  maxRetries?: number;
  openaiApiKey?: string | typeof FieldUnset;
  geminiApiKey?: string | typeof FieldUnset;
  xaiApiKey?: string | typeof FieldUnset;
}

export abstract class BaseRetab {
  protected apiKey: string;
  protected baseUrl: string;
  protected timeout: number;
  protected maxRetries: number;
  protected headers: Record<string, string>;

  constructor(config: RetabConfig = {}) {
    const apiKey = config.apiKey || process.env.RETAB_API_KEY;

    if (!apiKey) {
      throw new Error(
        'No API key provided. You can create an API key at https://retab.com\n' +
        'Then either pass it to the client (apiKey: "your-key") or set the RETAB_API_KEY environment variable'
      );
    }

    this.apiKey = apiKey;
    this.baseUrl = (config.baseUrl || process.env.RETAB_API_BASE_URL || 'https://api.retab.com').replace(/\/$/, '');
    this.timeout = config.timeout || 240000; // 240 seconds in milliseconds
    this.maxRetries = config.maxRetries || 3;

    this.headers = {
      'Api-Key': this.apiKey,
      'Content-Type': 'application/json',
    };

    // Handle API keys
    const openaiApiKey = config.openaiApiKey === FieldUnset 
      ? process.env.OPENAI_API_KEY 
      : config.openaiApiKey;
    
    const geminiApiKey = config.geminiApiKey === FieldUnset 
      ? process.env.GEMINI_API_KEY 
      : config.geminiApiKey;
    
    const xaiApiKey = config.xaiApiKey === FieldUnset
      ? process.env.XAI_API_KEY
      : config.xaiApiKey;

    if (openaiApiKey && typeof openaiApiKey === 'string') {
      this.headers['OpenAI-Api-Key'] = openaiApiKey;
    }

    if (xaiApiKey && typeof xaiApiKey === 'string') {
      this.headers['XAI-Api-Key'] = xaiApiKey;
    }

    if (geminiApiKey && typeof geminiApiKey === 'string') {
      this.headers['Gemini-Api-Key'] = geminiApiKey;
    }
  }

  protected _prepareUrl(endpoint: string): string {
    return `${this.baseUrl}/${endpoint.replace(/^\//, '')}`;
  }

  protected _validateResponse(response: AxiosResponse): void {
    if (response.status >= 500) {
      throw new APIError(response.status, response.statusText, response.data);
    } else if (response.status === 422) {
      throw new ValidationError(`Validation error (422): ${JSON.stringify(response.data)}`, response.data);
    } else if (response.status >= 400) {
      throw new APIError(response.status, `Request failed (${response.status}): ${JSON.stringify(response.data)}`, response.data);
    }
  }

  public getHeaders(): Record<string, string> {
    return { ...this.headers };
  }

  protected _getHeaders(idempotencyKey?: string | null): Record<string, string> {
    const headers = { ...this.headers };
    if (idempotencyKey) {
      headers['Idempotency-Key'] = idempotencyKey;
    }
    return headers;
  }

  protected _parseResponse(response: AxiosResponse): any {
    const contentType = response.headers['content-type'] || '';

    if (contentType.includes('application/json') || contentType.includes('application/stream+json')) {
      return response.data;
    } else if (contentType.includes('text/plain') || contentType.includes('text/')) {
      return response.data;
    } else {
      // Default to returning data as-is (axios already parses JSON)
      return response.data;
    }
  }
}

export class Retab extends BaseRetab {
  private client: AxiosInstance;

  public evaluations: Evaluations;
  public files: Files;
  public fineTuning: FineTuning;
  public documents: Documents;
  public models: Models;
  public schemas: Schemas;
  public processors: Processors;
  public secrets: Secrets;
  public usage: Usage;
  public consensus: Consensus;

  constructor(config: RetabConfig = {}) {
    super(config);

    this.client = axios.create({
      timeout: this.timeout,
    });

    // Configure retry logic
    axiosRetry(this.client, {
      retries: this.maxRetries,
      retryDelay: axiosRetry.exponentialDelay,
      retryCondition: (error) => {
        return axiosRetry.isNetworkOrIdempotentRequestError(error) || 
               (error.response?.status ? error.response.status >= 500 : false);
      },
      onRetry: (retryCount, error) => {
        if (retryCount >= this.maxRetries) {
          throw new MaxRetriesExceeded(retryCount + 1, error as Error);
        }
      }
    });

    // Initialize resources
    this.evaluations = new Evaluations(this);
    this.files = new Files(this);
    this.fineTuning = new FineTuning(this);
    this.documents = new Documents(this);
    this.models = new Models(this);
    this.schemas = new Schemas(this);
    this.processors = new Processors(this);
    this.secrets = new Secrets(this);
    this.usage = new Usage(this);
    this.consensus = new Consensus(this);
  }

  async _request(
    method: string,
    endpoint: string,
    data?: Record<string, any> | null,
    params?: Record<string, any> | null,
    formData?: Record<string, any> | null,
    files?: Record<string, any> | Array<[string, [string, Buffer, string]]> | null,
    idempotencyKey?: string | null,
    _raiseForStatus?: boolean
  ): Promise<any> {
    const config: AxiosRequestConfig = {
      method: method as any,
      url: this._prepareUrl(endpoint),
      params,
      headers: this._getHeaders(idempotencyKey),
    };

    // Handle different content types
    if (files || formData) {
      // For multipart/form-data requests
      const FormData = (await import('form-data')).default;
      const form = new FormData();
      
      if (formData) {
        Object.entries(formData).forEach(([key, value]) => {
          form.append(key, value);
        });
      }
      
      if (files) {
        if (Array.isArray(files)) {
          files.forEach(([fieldName, [filename, buffer, contentType]]) => {
            form.append(fieldName, buffer, { filename, contentType });
          });
        } else {
          Object.entries(files).forEach(([key, value]) => {
            form.append(key, value);
          });
        }
      }
      
      config.data = form;
      config.headers = {
        ...config.headers,
        ...form.getHeaders(),
      };
      delete config.headers['Content-Type'];
    } else if (data) {
      config.data = data;
    }

    try {
      const response = await this.client.request(config);
      this._validateResponse(response);
      return this._parseResponse(response);
    } catch (error) {
      if (_raiseForStatus && axios.isAxiosError(error)) {
        throw error;
      }
      throw error;
    }
  }

  async _requestStream(
    method: string,
    endpoint: string,
    data?: Record<string, any> | null,
    params?: Record<string, any> | null,
    formData?: Record<string, any> | null,
    files?: Record<string, any> | Array<[string, [string, Buffer, string]]> | null,
    idempotencyKey?: string | null,
    _raiseForStatus?: boolean
  ): Promise<AsyncGenerator<any, void, unknown>> {
    const config: AxiosRequestConfig = {
      method: method as any,
      url: this._prepareUrl(endpoint),
      params,
      headers: this._getHeaders(idempotencyKey),
      responseType: 'stream',
    };

    // Handle different content types
    if (files || formData) {
      const FormData = (await import('form-data')).default;
      const form = new FormData();
      
      if (formData) {
        Object.entries(formData).forEach(([key, value]) => {
          form.append(key, value);
        });
      }
      
      if (files) {
        if (Array.isArray(files)) {
          files.forEach(([fieldName, [filename, buffer, contentType]]) => {
            form.append(fieldName, buffer, { filename, contentType });
          });
        } else {
          Object.entries(files).forEach(([key, value]) => {
            form.append(key, value);
          });
        }
      }
      
      config.data = form;
      config.headers = {
        ...config.headers,
        ...form.getHeaders(),
      };
      delete config.headers['Content-Type'];
    } else if (data) {
      config.data = data;
    }

    const response = await this.client.request(config);
    this._validateResponse(response);

    const contentType = response.headers['content-type'] || '';
    const isJsonStream = contentType.includes('application/json') || contentType.includes('application/stream+json');
    const isTextStream = contentType.includes('text/plain') || (contentType.includes('text/') && !isJsonStream);

    return (async function* () {
      const stream = response.data;
      let buffer = '';

      for await (const chunk of stream) {
        buffer += chunk.toString();
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (!line) continue;

          if (isJsonStream) {
            try {
              yield JSON.parse(line);
            } catch {
              // Skip invalid JSON
            }
          } else if (isTextStream) {
            yield line;
          } else {
            try {
              yield JSON.parse(line);
            } catch {
              yield line;
            }
          }
        }
      }

      // Process any remaining buffer
      if (buffer) {
        if (isJsonStream) {
          try {
            yield JSON.parse(buffer);
          } catch {
            // Skip invalid JSON
          }
        } else {
          yield buffer;
        }
      }
    })();
  }

  async _preparedRequest(request: PreparedRequest): Promise<any> {
    return this._request(
      request.method,
      request.url,
      request.data,
      request.params,
      request.formData,
      request.files,
      request.idempotencyKey,
      request.raiseForStatus
    );
  }

  async _preparedRequestStream(request: PreparedRequest): Promise<AsyncGenerator<any, void, unknown>> {
    return this._requestStream(
      request.method,
      request.url,
      request.data,
      request.params,
      request.formData,
      request.files,
      request.idempotencyKey,
      request.raiseForStatus
    );
  }

  close(): void {
    // axios doesn't need explicit cleanup
  }
}

export class AsyncRetab extends BaseRetab {
  private client: AxiosInstance;

  public evaluations: AsyncEvaluations;
  public files: AsyncFiles;
  public fineTuning: AsyncFineTuning;
  public documents: AsyncDocuments;
  public models: AsyncModels;
  public schemas: AsyncSchemas;
  public processors: AsyncProcessors;
  public secrets: AsyncSecrets;
  public usage: AsyncUsage;
  public consensus: AsyncConsensus;

  constructor(config: RetabConfig = {}) {
    super(config);

    this.client = axios.create({
      timeout: this.timeout,
    });

    // Configure retry logic
    axiosRetry(this.client, {
      retries: this.maxRetries,
      retryDelay: axiosRetry.exponentialDelay,
      retryCondition: (error) => {
        return axiosRetry.isNetworkOrIdempotentRequestError(error) || 
               (error.response?.status ? error.response.status >= 500 : false);
      },
      onRetry: (retryCount, error) => {
        if (retryCount >= this.maxRetries) {
          throw new MaxRetriesExceeded(retryCount + 1, error as Error);
        }
      }
    });

    // Initialize resources
    this.evaluations = new AsyncEvaluations(this);
    this.files = new AsyncFiles(this);
    this.fineTuning = new AsyncFineTuning(this);
    this.documents = new AsyncDocuments(this);
    this.models = new AsyncModels(this);
    this.schemas = new AsyncSchemas(this);
    this.processors = new AsyncProcessors(this);
    this.secrets = new AsyncSecrets(this);
    this.usage = new AsyncUsage(this);
    this.consensus = new AsyncConsensus(this);
  }

  async _request(
    method: string,
    endpoint: string,
    data?: Record<string, any> | null,
    params?: Record<string, any> | null,
    formData?: Record<string, any> | null,
    files?: Record<string, any> | Array<[string, [string, Buffer, string]]> | null,
    idempotencyKey?: string | null,
    _raiseForStatus?: boolean
  ): Promise<any> {
    const config: AxiosRequestConfig = {
      method: method as any,
      url: this._prepareUrl(endpoint),
      params,
      headers: this._getHeaders(idempotencyKey),
    };

    // Handle different content types
    if (files || formData) {
      const FormData = (await import('form-data')).default;
      const form = new FormData();
      
      if (formData) {
        Object.entries(formData).forEach(([key, value]) => {
          form.append(key, value);
        });
      }
      
      if (files) {
        if (Array.isArray(files)) {
          files.forEach(([fieldName, [filename, buffer, contentType]]) => {
            form.append(fieldName, buffer, { filename, contentType });
          });
        } else {
          Object.entries(files).forEach(([key, value]) => {
            form.append(key, value);
          });
        }
      }
      
      config.data = form;
      config.headers = {
        ...config.headers,
        ...form.getHeaders(),
      };
      delete config.headers['Content-Type'];
    } else if (data) {
      config.data = data;
    }

    try {
      const response = await this.client.request(config);
      this._validateResponse(response);
      return this._parseResponse(response);
    } catch (error) {
      if (_raiseForStatus && axios.isAxiosError(error)) {
        throw error;
      }
      throw error;
    }
  }

  async *_requestStream(
    method: string,
    endpoint: string,
    data?: Record<string, any> | null,
    params?: Record<string, any> | null,
    formData?: Record<string, any> | null,
    files?: Record<string, any> | Array<[string, [string, Buffer, string]]> | null,
    idempotencyKey?: string | null,
    _raiseForStatus?: boolean
  ): AsyncGenerator<any, void, unknown> {
    const config: AxiosRequestConfig = {
      method: method as any,
      url: this._prepareUrl(endpoint),
      params,
      headers: this._getHeaders(idempotencyKey),
      responseType: 'stream',
    };

    // Handle different content types
    if (files || formData) {
      const FormData = (await import('form-data')).default;
      const form = new FormData();
      
      if (formData) {
        Object.entries(formData).forEach(([key, value]) => {
          form.append(key, value);
        });
      }
      
      if (files) {
        if (Array.isArray(files)) {
          files.forEach(([fieldName, [filename, buffer, contentType]]) => {
            form.append(fieldName, buffer, { filename, contentType });
          });
        } else {
          Object.entries(files).forEach(([key, value]) => {
            form.append(key, value);
          });
        }
      }
      
      config.data = form;
      config.headers = {
        ...config.headers,
        ...form.getHeaders(),
      };
      delete config.headers['Content-Type'];
    } else if (data) {
      config.data = data;
    }

    const response = await this.client.request(config);
    this._validateResponse(response);

    const contentType = response.headers['content-type'] || '';
    const isJsonStream = contentType.includes('application/json') || contentType.includes('application/stream+json');
    const isTextStream = contentType.includes('text/plain') || (contentType.includes('text/') && !isJsonStream);

    const stream = response.data;
    let buffer = '';

    for await (const chunk of stream) {
      buffer += chunk.toString();
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';

      for (const line of lines) {
        if (!line) continue;

        if (isJsonStream) {
          try {
            yield JSON.parse(line);
          } catch {
            // Skip invalid JSON
          }
        } else if (isTextStream) {
          yield line;
        } else {
          try {
            yield JSON.parse(line);
          } catch {
            yield line;
          }
        }
      }
    }

    // Process any remaining buffer
    if (buffer) {
      if (isJsonStream) {
        try {
          yield JSON.parse(buffer);
        } catch {
          // Skip invalid JSON
        }
      } else {
        yield buffer;
      }
    }
  }

  async _preparedRequest(request: PreparedRequest): Promise<any> {
    return this._request(
      request.method,
      request.url,
      request.data,
      request.params,
      request.formData,
      request.files,
      request.idempotencyKey,
      request.raiseForStatus
    );
  }

  async *_preparedRequestStream(request: PreparedRequest): AsyncGenerator<any, void, unknown> {
    yield* this._requestStream(
      request.method,
      request.url,
      request.data,
      request.params,
      request.formData,
      request.files,
      request.idempotencyKey,
      request.raiseForStatus
    );
  }

  async close(): Promise<void> {
    // axios doesn't need explicit cleanup
  }
}