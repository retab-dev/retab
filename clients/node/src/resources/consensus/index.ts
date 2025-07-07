import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';
import { ReconciliationRequest, ReconciliationResponse } from '../../types/consensus.js';

export class ConsensusMixin {
  prepareReconcile(params: {
    list_dicts: Array<Record<string, any>>;
    reference_schema?: Record<string, any>;
    mode?: 'direct' | 'aligned';
    idempotency_key?: string;
  }): any {
    const request: ReconciliationRequest = {
      list_dicts: params.list_dicts,
      reference_schema: params.reference_schema,
      mode: params.mode || 'direct',
    };

    return {
      method: 'POST' as const,
      url: '/v1/consensus/reconcile',
      data: request,
      idempotency_key: params.idempotency_key,
    };
  }

  prepareExtract(params: {
    document: string;
    model?: string;
    n_consensus?: number;
    json_schema?: Record<string, any> | string;
    modality?: string;
    image_resolution_dpi?: number;
    browser_canvas?: string;
    temperature?: number;
    reasoning_effort?: string;
    idempotency_key?: string;
  }): any {
    const {
      document,
      model = 'gpt-4o-mini',
      n_consensus = 3,
      json_schema,
      modality = 'native',
      image_resolution_dpi = 96,
      browser_canvas = 'A4',
      temperature = 0.0,
      reasoning_effort = 'medium',
      idempotency_key,
    } = params;

    const requestData: Record<string, any> = {
      model,
      n_consensus,
      modality,
      image_resolution_dpi,
      browser_canvas,
      temperature,
      reasoning_effort,
    };

    if (json_schema !== undefined) {
      requestData.json_schema = json_schema;
    }

    return {
      method: 'POST' as const,
      url: '/v1/consensus/extract',
      form_data: requestData,
      files: { document },
      idempotency_key,
    };
  }
}

export class Consensus extends SyncAPIResource {
  mixin = new ConsensusMixin();

  reconcile(params: {
    list_dicts: Array<Record<string, any>>;
    reference_schema?: Record<string, any>;
    mode?: 'direct' | 'aligned';
    idempotency_key?: string;
  }): any {
    const preparedRequest = this.mixin.prepareReconcile(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response;
  }

  extract(params: {
    document: string;
    model?: string;
    n_consensus?: number;
    json_schema?: Record<string, any> | string;
    modality?: string;
    image_resolution_dpi?: number;
    browser_canvas?: string;
    temperature?: number;
    reasoning_effort?: string;
    idempotency_key?: string;
  }): any {
    const preparedRequest = this.mixin.prepareExtract(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response;
  }
}

export class AsyncConsensus extends AsyncAPIResource {
  mixin = new ConsensusMixin();

  async reconcile(params: {
    list_dicts: Array<Record<string, any>>;
    reference_schema?: Record<string, any>;
    mode?: 'direct' | 'aligned';
    idempotency_key?: string;
  }): Promise<any> {
    const preparedRequest = this.mixin.prepareReconcile(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response;
  }

  async extract(params: {
    document: string;
    model?: string;
    n_consensus?: number;
    json_schema?: Record<string, any> | string;
    modality?: string;
    image_resolution_dpi?: number;
    browser_canvas?: string;
    temperature?: number;
    reasoning_effort?: string;
    idempotency_key?: string;
  }): Promise<any> {
    const preparedRequest = this.mixin.prepareExtract(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response;
  }
}