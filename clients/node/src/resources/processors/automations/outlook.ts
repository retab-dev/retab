import { SyncAPIResource, AsyncAPIResource } from '../../../resource.js';
import { Outlook, ListOutlooks, MatchParams, FetchParams } from '../../../types/automations/outlook.js';

export class OutlookMixin {
  readonly outlookBaseUrl = '/v1/processors/automations/outlook';

  prepareCreate(params: {
    processor_id: string;
    name: string;
    webhook_url: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    authorized_domains?: string[];
    authorized_emails?: string[];
    layout_schema?: Record<string, any>;
    match_params?: MatchParams[];
    fetch_params?: FetchParams[];
  }): any {
    const {
      processor_id,
      name,
      webhook_url,
      webhook_headers,
      need_validation,
      authorized_domains,
      authorized_emails,
      layout_schema,
      match_params,
      fetch_params,
    } = params;

    const outlookDict: Record<string, any> = {
      processor_id,
      name,
      webhook_url,
    };

    if (webhook_headers !== undefined) {
      outlookDict.webhook_headers = webhook_headers;
    }
    if (need_validation !== undefined) {
      outlookDict.need_validation = need_validation;
    }
    if (authorized_domains !== undefined) {
      outlookDict.authorized_domains = authorized_domains;
    }
    if (authorized_emails !== undefined) {
      outlookDict.authorized_emails = authorized_emails.map(email => email.trim().toLowerCase());
    }
    if (layout_schema !== undefined) {
      outlookDict.layout_schema = layout_schema;
    }
    if (match_params !== undefined) {
      outlookDict.match_params = match_params;
    }
    if (fetch_params !== undefined) {
      outlookDict.fetch_params = fetch_params;
    }

    return {
      method: 'POST' as const,
      url: this.outlookBaseUrl,
      data: outlookDict,
    };
  }

  prepareList(params: {
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
    processor_id?: string;
    name?: string;
  } = {}): any {
    const {
      before,
      after,
      limit = 10,
      order = 'desc',
      processor_id,
      name,
    } = params;

    const queryParams: Record<string, any> = {};
    if (before !== undefined) queryParams.before = before;
    if (after !== undefined) queryParams.after = after;
    if (limit !== undefined) queryParams.limit = limit;
    if (order !== undefined) queryParams.order = order;
    if (processor_id !== undefined) queryParams.processor_id = processor_id;
    if (name !== undefined) queryParams.name = name;

    return {
      method: 'GET' as const,
      url: this.outlookBaseUrl,
      params: queryParams,
    };
  }

  prepareGet(outlook_id: string): any {
    return {
      method: 'GET' as const,
      url: `${this.outlookBaseUrl}/${outlook_id}`,
    };
  }

  prepareUpdate(params: {
    outlook_id: string;
    name?: string;
    webhook_url?: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    authorized_domains?: string[];
    authorized_emails?: string[];
    match_params?: MatchParams[];
    fetch_params?: FetchParams[];
    layout_schema?: Record<string, any>;
  }): any {
    const {
      outlook_id,
      name,
      webhook_url,
      webhook_headers,
      need_validation,
      authorized_domains,
      authorized_emails,
      match_params,
      fetch_params,
      layout_schema,
    } = params;

    const updateDict: Record<string, any> = {};
    if (name !== undefined) updateDict.name = name;
    if (webhook_url !== undefined) updateDict.webhook_url = webhook_url;
    if (webhook_headers !== undefined) updateDict.webhook_headers = webhook_headers;
    if (need_validation !== undefined) updateDict.need_validation = need_validation;
    if (authorized_domains !== undefined) updateDict.authorized_domains = authorized_domains;
    if (authorized_emails !== undefined) {
      updateDict.authorized_emails = authorized_emails.map(email => email.trim().toLowerCase());
    }
    if (match_params !== undefined) updateDict.match_params = match_params;
    if (fetch_params !== undefined) updateDict.fetch_params = fetch_params;
    if (layout_schema !== undefined) updateDict.layout_schema = layout_schema;

    return {
      method: 'PUT' as const,
      url: `${this.outlookBaseUrl}/${outlook_id}`,
      data: updateDict,
    };
  }

  prepareDelete(outlook_id: string): any {
    return {
      method: 'DELETE' as const,
      url: `${this.outlookBaseUrl}/${outlook_id}`,
    };
  }
}

export class OutlookAutomations extends SyncAPIResource {
  mixin = new OutlookMixin();

  create(params: {
    processor_id: string;
    name: string;
    webhook_url: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    authorized_domains?: string[];
    authorized_emails?: string[];
    layout_schema?: Record<string, any>;
    match_params?: MatchParams[];
    fetch_params?: FetchParams[];
  }): Promise<Outlook> {
    const preparedRequest = this.mixin.prepareCreate(params);
    const response = this._client._preparedRequest(preparedRequest);
    response.then((result: any) => {
      console.log(`Outlook automation Created. Url: https://www.retab.com/dashboard/processors/automations/${result.id}`);
    });
    return response as Promise<Outlook>;
  }

  list(params: {
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
    processor_id?: string;
    name?: string;
  } = {}): Promise<ListOutlooks> {
    const preparedRequest = this.mixin.prepareList(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Promise<ListOutlooks>;
  }

  get(outlook_id: string): Promise<Outlook> {
    const preparedRequest = this.mixin.prepareGet(outlook_id);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Promise<Outlook>;
  }

  update(params: {
    outlook_id: string;
    name?: string;
    webhook_url?: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    authorized_domains?: string[];
    authorized_emails?: string[];
    match_params?: MatchParams[];
    fetch_params?: FetchParams[];
    layout_schema?: Record<string, any>;
  }): Promise<Outlook> {
    const preparedRequest = this.mixin.prepareUpdate(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Promise<Outlook>;
  }

  delete(outlook_id: string): Promise<void> {
    const preparedRequest = this.mixin.prepareDelete(outlook_id);
    const response = this._client._preparedRequest(preparedRequest);
    response.then(() => {
      console.log(`Outlook automation Deleted. ID: ${outlook_id}`);
    });
    return response as Promise<void>;
  }
}

export class AsyncOutlookAutomations extends AsyncAPIResource {
  mixin = new OutlookMixin();

  async create(params: {
    processor_id: string;
    name: string;
    webhook_url: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    authorized_domains?: string[];
    authorized_emails?: string[];
    layout_schema?: Record<string, any>;
    match_params?: MatchParams[];
    fetch_params?: FetchParams[];
  }): Promise<Outlook> {
    const preparedRequest = this.mixin.prepareCreate(params);
    const response = await this._client._preparedRequest(preparedRequest);
    console.log(`Outlook automation Created. Url: https://www.retab.com/dashboard/processors/automations/${response.id}`);
    return response as Outlook;
  }

  async list(params: {
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
    processor_id?: string;
    name?: string;
  } = {}): Promise<ListOutlooks> {
    const preparedRequest = this.mixin.prepareList(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as ListOutlooks;
  }

  async get(outlook_id: string): Promise<Outlook> {
    const preparedRequest = this.mixin.prepareGet(outlook_id);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as Outlook;
  }

  async update(params: {
    outlook_id: string;
    name?: string;
    webhook_url?: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    authorized_domains?: string[];
    authorized_emails?: string[];
    match_params?: MatchParams[];
    fetch_params?: FetchParams[];
    layout_schema?: Record<string, any>;
  }): Promise<Outlook> {
    const preparedRequest = this.mixin.prepareUpdate(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as Outlook;
  }

  async delete(outlook_id: string): Promise<void> {
    const preparedRequest = this.mixin.prepareDelete(outlook_id);
    await this._client._preparedRequest(preparedRequest);
    console.log(`Outlook automation Deleted. ID: ${outlook_id}`);
  }
}