import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';

export class Consensus extends SyncAPIResource {
  reconcile(params: {
    list_dicts: Array<Record<string, any>>;
    reference_schema?: Record<string, any>;
    mode?: 'direct' | 'aligned';
    idempotency_key?: string;
  }): any {
    const preparedRequest = {
      method: 'POST' as const,
      url: '/v1/consensus/reconcile',
      data: {
        list_dicts: params.list_dicts,
        reference_schema: params.reference_schema,
        mode: params.mode || 'direct',
        idempotency_key: params.idempotency_key,
      }
    };
    return this._client._preparedRequest(preparedRequest);
  }
}

export class AsyncConsensus extends AsyncAPIResource {
  async reconcile(params: {
    list_dicts: Array<Record<string, any>>;
    reference_schema?: Record<string, any>;
    mode?: 'direct' | 'aligned';
    idempotency_key?: string;
  }): Promise<any> {
    const preparedRequest = {
      method: 'POST' as const,
      url: '/v1/consensus/reconcile',
      data: {
        list_dicts: params.list_dicts,
        reference_schema: params.reference_schema,
        mode: params.mode || 'direct',
        idempotency_key: params.idempotency_key,
      }
    };
    return this._client._preparedRequest(preparedRequest);
  }
}