import { SyncAPIResource, AsyncAPIResource } from '../../../resource.js';
import { ListLogs } from '../../../types/logs.js';

export class LogsMixin {
  prepareList(params: {
    automation_id: string;
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
  }): any {
    const {
      automation_id,
      before,
      after,
      limit = 10,
      order = 'desc',
    } = params;

    const queryParams: Record<string, any> = {
      automation_id,
    };

    if (before !== undefined) queryParams.before = before;
    if (after !== undefined) queryParams.after = after;
    if (limit !== undefined) queryParams.limit = limit;
    if (order !== undefined) queryParams.order = order;

    return {
      method: 'GET' as const,
      url: '/v1/processors/automations/logs',
      params: queryParams,
    };
  }

  prepareGet(log_id: string): any {
    return {
      method: 'GET' as const,
      url: `/v1/processors/automations/logs/${log_id}`,
    };
  }
}

export class Logs extends SyncAPIResource {
  mixin = new LogsMixin();

  list(params: {
    automation_id: string;
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
  }): ListLogs {
    const preparedRequest = this.mixin.prepareList(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response as ListLogs;
  }

  get(log_id: string): any {
    const preparedRequest = this.mixin.prepareGet(log_id);
    const response = this._client._preparedRequest(preparedRequest);
    return response;
  }
}

export class AsyncLogs extends AsyncAPIResource {
  mixin = new LogsMixin();

  async list(params: {
    automation_id: string;
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
  }): Promise<ListLogs> {
    const preparedRequest = this.mixin.prepareList(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as ListLogs;
  }

  async get(log_id: string): Promise<any> {
    const preparedRequest = this.mixin.prepareGet(log_id);
    const response = await this._client._preparedRequest(preparedRequest);
    return response;
  }
}