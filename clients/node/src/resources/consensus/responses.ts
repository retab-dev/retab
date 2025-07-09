import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';
import { PreparedRequest } from '../../types/standards.js';

export interface ResponseRequest {
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

export interface ResponseResponse {
  id: string;
  object: 'consensus.response';
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
    reconciliation_method: 'direct' | 'aligned';
  };
}

export class ResponsesMixin {
  prepareCreate(params: ResponseRequest): PreparedRequest {
    const { stream, idempotency_key, ...data } = params;
    
    return {
      method: 'POST',
      url: '/v1/consensus/responses',
      data: {
        ...data,
        stream: stream || false,
      },
      idempotencyKey: idempotency_key,
    };
  }

  prepareParse(params: ResponseRequest): PreparedRequest {
    const { stream, idempotency_key, ...data } = params;
    
    return {
      method: 'POST',
      url: '/v1/consensus/responses/parse',
      data: {
        ...data,
        stream: stream || false,
      },
      idempotencyKey: idempotency_key,
    };
  }

  prepareStream(params: ResponseRequest): PreparedRequest {
    const { idempotency_key, ...data } = params;
    
    return {
      method: 'POST',
      url: '/v1/consensus/responses',
      data: {
        ...data,
        stream: true,
      },
      idempotencyKey: idempotency_key,
    };
  }

  prepareStreamParse(params: ResponseRequest): PreparedRequest {
    const { idempotency_key, ...data } = params;
    
    return {
      method: 'POST',
      url: '/v1/consensus/responses/parse',
      data: {
        ...data,
        stream: true,
      },
      idempotencyKey: idempotency_key,
    };
  }
}

export class Responses extends SyncAPIResource {
  private mixin = new ResponsesMixin();

  create(params: ResponseRequest): Promise<ResponseResponse> {
    const preparedRequest = this.mixin.prepareCreate(params);
    return this._client._preparedRequest(preparedRequest);
  }

  parse(params: ResponseRequest): Promise<ResponseResponse> {
    const preparedRequest = this.mixin.prepareParse(params);
    return this._client._preparedRequest(preparedRequest);
  }

  async *stream(params: ResponseRequest): AsyncGenerator<ResponseResponse, void, unknown> {
    const preparedRequest = this.mixin.prepareStream(params);
    const streamGenerator = await this._client._preparedRequestStream(preparedRequest);
    yield* streamGenerator;
  }

  async *streamParse(params: ResponseRequest): AsyncGenerator<ResponseResponse, void, unknown> {
    const preparedRequest = this.mixin.prepareStreamParse(params);
    const streamGenerator = await this._client._preparedRequestStream(preparedRequest);
    yield* streamGenerator;
  }
}

export class AsyncResponses extends AsyncAPIResource {
  private mixin = new ResponsesMixin();

  async create(params: ResponseRequest): Promise<ResponseResponse> {
    const preparedRequest = this.mixin.prepareCreate(params);
    return await this._client._preparedRequest(preparedRequest);
  }

  async parse(params: ResponseRequest): Promise<ResponseResponse> {
    const preparedRequest = this.mixin.prepareParse(params);
    return await this._client._preparedRequest(preparedRequest);
  }

  async *stream(params: ResponseRequest): AsyncGenerator<ResponseResponse, void, unknown> {
    const preparedRequest = this.mixin.prepareStream(params);
    const streamGenerator = await this._client._preparedRequestStream(preparedRequest);
    yield* streamGenerator;
  }

  async *streamParse(params: ResponseRequest): AsyncGenerator<ResponseResponse, void, unknown> {
    const preparedRequest = this.mixin.prepareStreamParse(params);
    const streamGenerator = await this._client._preparedRequestStream(preparedRequest);
    yield* streamGenerator;
  }
}