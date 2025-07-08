import { SyncAPIResource, AsyncAPIResource } from '../../../resource.js';
import { Endpoint, ListEndpoints } from '../../../types/automations/endpoints.js';

export class EndpointsMixin {
  prepareCreate(params: {
    processor_id: string;
    name: string;
    webhook_url: string;
    model?: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
  }): any {
    const {
      processor_id,
      name,
      webhook_url,
      model = 'gpt-4o-mini',
      webhook_headers,
      need_validation,
    } = params;

    const endpointDict: Record<string, any> = {
      processor_id,
      name,
      webhook_url,
      model,
    };

    if (webhook_headers !== undefined) {
      endpointDict.webhook_headers = webhook_headers;
    }
    if (need_validation !== undefined) {
      endpointDict.need_validation = need_validation;
    }

    return {
      method: 'POST' as const,
      url: '/v1/processors/automations/endpoints',
      data: endpointDict,
    };
  }

  prepareList(params: {
    processor_id: string;
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
    name?: string;
    webhook_url?: string;
  }): any {
    const {
      processor_id,
      before,
      after,
      limit = 10,
      order = 'desc',
      name,
      webhook_url,
    } = params;

    const queryParams: Record<string, any> = {
      processor_id,
    };

    if (before !== undefined) queryParams.before = before;
    if (after !== undefined) queryParams.after = after;
    if (limit !== undefined) queryParams.limit = limit;
    if (order !== undefined) queryParams.order = order;
    if (name !== undefined) queryParams.name = name;
    if (webhook_url !== undefined) queryParams.webhook_url = webhook_url;

    return {
      method: 'GET' as const,
      url: '/v1/processors/automations/endpoints',
      params: queryParams,
    };
  }

  prepareGet(endpoint_id: string): any {
    return {
      method: 'GET' as const,
      url: `/v1/processors/automations/endpoints/${endpoint_id}`,
    };
  }

  prepareUpdate(params: {
    endpoint_id: string;
    name?: string;
    default_language?: string;
    webhook_url?: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
  }): any {
    const {
      endpoint_id,
      name,
      default_language,
      webhook_url,
      webhook_headers,
      need_validation,
    } = params;

    const updateDict: Record<string, any> = {};
    if (name !== undefined) updateDict.name = name;
    if (default_language !== undefined) updateDict.default_language = default_language;
    if (webhook_url !== undefined) updateDict.webhook_url = webhook_url;
    if (webhook_headers !== undefined) updateDict.webhook_headers = webhook_headers;
    if (need_validation !== undefined) updateDict.need_validation = need_validation;

    return {
      method: 'PUT' as const,
      url: `/v1/processors/automations/endpoints/${endpoint_id}`,
      data: updateDict,
    };
  }

  prepareDelete(endpoint_id: string): any {
    return {
      method: 'DELETE' as const,
      url: `/v1/processors/automations/endpoints/${endpoint_id}`,
    };
  }
}

export class Endpoints extends SyncAPIResource {
  mixin = new EndpointsMixin();

  create(params: {
    processor_id: string;
    name: string;
    webhook_url: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
  }): Promise<Endpoint> {
    const preparedRequest = this.mixin.prepareCreate(params);
    const response = this._client._preparedRequest(preparedRequest);
    // Note: response is a Promise, access id after awaiting
    response.then((r: any) => {
      console.log(`Endpoint Created. Url: https://www.retab.com/dashboard/processors/automations/${r?.id || 'unknown'}`);
    }).catch(() => {});
    return response as Promise<Endpoint>;
  }

  list(params: {
    processor_id: string;
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
    name?: string;
    webhook_url?: string;
  }): Promise<ListEndpoints> {
    const preparedRequest = this.mixin.prepareList(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Promise<ListEndpoints>;
  }

  get(endpoint_id: string): Promise<Endpoint> {
    const preparedRequest = this.mixin.prepareGet(endpoint_id);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Promise<Endpoint>;
  }

  update(params: {
    endpoint_id: string;
    name?: string;
    default_language?: string;
    webhook_url?: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
  }): Promise<Endpoint> {
    const preparedRequest = this.mixin.prepareUpdate(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Promise<Endpoint>;
  }

  delete(endpoint_id: string): Promise<void> {
    const preparedRequest = this.mixin.prepareDelete(endpoint_id);
    const response = this._client._preparedRequest(preparedRequest);
    console.log(`Endpoint Deleted. ID: ${endpoint_id}`);
    return response as Promise<void>;
  }
}

export class AsyncEndpoints extends AsyncAPIResource {
  mixin = new EndpointsMixin();

  async create(params: {
    processor_id: string;
    name: string;
    webhook_url: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
  }): Promise<Endpoint> {
    const preparedRequest = this.mixin.prepareCreate(params);
    const response = await this._client._preparedRequest(preparedRequest);
    console.log(`Endpoint Created. Url: https://www.retab.com/dashboard/processors/automations/${response?.id || 'unknown'}`);
    return response as Endpoint;
  }

  async list(params: {
    processor_id: string;
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
    name?: string;
    webhook_url?: string;
  }): Promise<ListEndpoints> {
    const preparedRequest = this.mixin.prepareList(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as ListEndpoints;
  }

  async get(endpoint_id: string): Promise<Endpoint> {
    const preparedRequest = this.mixin.prepareGet(endpoint_id);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as Endpoint;
  }

  async update(params: {
    endpoint_id: string;
    name?: string;
    default_language?: string;
    webhook_url?: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
  }): Promise<Endpoint> {
    const preparedRequest = this.mixin.prepareUpdate(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as Endpoint;
  }

  async delete(endpoint_id: string): Promise<void> {
    const preparedRequest = this.mixin.prepareDelete(endpoint_id);
    await this._client._preparedRequest(preparedRequest);
    console.log(`Endpoint Deleted. ID: ${endpoint_id}`);
  }
}