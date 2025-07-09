import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';
import { PreparedRequest } from '../../types/standards.js';
import { StreamWrapper, streamUtils } from '../../utils/stream_context_managers.js';

export interface CompletionRequest {
  model: string;
  messages: Array<{
    role: 'system' | 'user' | 'assistant';
    content: string;
  }>;
  temperature?: number;
  n_consensus?: number;
  reasoning_effort?: 'low' | 'medium' | 'high';
  response_format?: {
    type: 'json_schema';
    json_schema: {
      name: string;
      schema: Record<string, any>;
      strict?: boolean;
    };
  };
  stream?: boolean;
  idempotency_key?: string;
}

export interface CompletionResponse {
  id: string;
  object: 'consensus.completion';
  created: number;
  model: string;
  choices: Array<{
    index: number;
    message: {
      role: 'assistant';
      content: string;
    };
    finish_reason: 'stop' | 'length' | 'content_filter';
    confidence?: number;
  }>;
  usage: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
  consensus_metadata: {
    n_consensus: number;
    agreement_score: number;
    individual_results: Array<Record<string, any>>;
  };
}

export class CompletionsMixin {
  prepareCreate(params: CompletionRequest): PreparedRequest {
    const { stream, idempotency_key, ...data } = params;
    
    return {
      method: 'POST',
      url: '/v1/consensus/completions',
      data: {
        ...data,
        stream: stream || false,
      },
      idempotencyKey: idempotency_key,
    };
  }

  prepareParse(params: CompletionRequest): PreparedRequest {
    const { stream, idempotency_key, ...data } = params;
    
    return {
      method: 'POST',
      url: '/v1/consensus/completions/parse',
      data: {
        ...data,
        stream: stream || false,
      },
      idempotencyKey: idempotency_key,
    };
  }

  prepareStream(params: CompletionRequest): PreparedRequest {
    const { idempotency_key, ...data } = params;
    
    return {
      method: 'POST',
      url: '/v1/consensus/completions',
      data: {
        ...data,
        stream: true,
      },
      idempotencyKey: idempotency_key,
    };
  }
}

export class Completions extends SyncAPIResource {
  private mixin = new CompletionsMixin();

  create(params: CompletionRequest): Promise<CompletionResponse> {
    const preparedRequest = this.mixin.prepareCreate(params);
    return this._client._preparedRequest(preparedRequest);
  }

  parse(params: CompletionRequest): Promise<CompletionResponse> {
    const preparedRequest = this.mixin.prepareParse(params);
    return this._client._preparedRequest(preparedRequest);
  }

  stream(params: CompletionRequest): StreamWrapper<CompletionResponse> {
    const preparedRequest = this.mixin.prepareStream(params);
    
    const generator = async function* (this: any): AsyncGenerator<CompletionResponse, void, unknown> {
      const streamGenerator = await this._client._preparedRequestStream(preparedRequest);
      yield* streamGenerator;
    }.bind(this);
    
    return streamUtils.wrap(generator());
  }
}

export class AsyncCompletions extends AsyncAPIResource {
  private mixin = new CompletionsMixin();

  async create(params: CompletionRequest): Promise<CompletionResponse> {
    const preparedRequest = this.mixin.prepareCreate(params);
    return await this._client._preparedRequest(preparedRequest);
  }

  async parse(params: CompletionRequest): Promise<CompletionResponse> {
    const preparedRequest = this.mixin.prepareParse(params);
    return await this._client._preparedRequest(preparedRequest);
  }

  stream(params: CompletionRequest): StreamWrapper<CompletionResponse> {
    const preparedRequest = this.mixin.prepareStream(params);
    
    const generator = async function* (this: any): AsyncGenerator<CompletionResponse, void, unknown> {
      const streamGenerator = await this._client._preparedRequestStream(preparedRequest);
      yield* streamGenerator;
    }.bind(this);
    
    return streamUtils.wrap(generator());
  }
}