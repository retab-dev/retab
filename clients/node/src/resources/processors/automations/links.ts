import { SyncAPIResource, AsyncAPIResource } from '../../../resource.js';
import { Link, ListLinks, UpdateLinkRequest } from '../../../types/automations/links.js';

export class LinksMixin {
  readonly linksBaseUrl = '/v1/processors/automations/links';

  prepareCreate(params: {
    processor_id: string;
    name: string;
    webhook_url: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    password?: string;
  }): any {
    const {
      processor_id,
      name,
      webhook_url,
      webhook_headers,
      need_validation,
      password,
    } = params;

    const linkDict: Record<string, any> = {
      processor_id,
      name,
      webhook_url,
    };

    if (webhook_headers !== undefined) {
      linkDict.webhook_headers = webhook_headers;
    }
    if (need_validation !== undefined) {
      linkDict.need_validation = need_validation;
    }
    if (password !== undefined) {
      linkDict.password = password;
    }

    return {
      method: 'POST' as const,
      url: this.linksBaseUrl,
      data: linkDict,
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
      url: this.linksBaseUrl,
      params: queryParams,
    };
  }

  prepareGet(link_id: string): any {
    return {
      method: 'GET' as const,
      url: `${this.linksBaseUrl}/${link_id}`,
    };
  }

  prepareUpdate(params: {
    link_id: string;
    name?: string;
    webhook_url?: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    password?: string;
  }): any {
    const {
      link_id,
      name,
      webhook_url,
      webhook_headers,
      need_validation,
      password,
    } = params;

    const updateDict: Record<string, any> = {};
    if (name !== undefined) updateDict.name = name;
    if (webhook_url !== undefined) updateDict.webhook_url = webhook_url;
    if (webhook_headers !== undefined) updateDict.webhook_headers = webhook_headers;
    if (need_validation !== undefined) updateDict.need_validation = need_validation;
    if (password !== undefined) updateDict.password = password;

    return {
      method: 'PUT' as const,
      url: `${this.linksBaseUrl}/${link_id}`,
      data: updateDict,
    };
  }

  prepareDelete(link_id: string): any {
    return {
      method: 'DELETE' as const,
      url: `${this.linksBaseUrl}/${link_id}`,
      raiseForStatus: true,
    };
  }
}

export class Links extends SyncAPIResource {
  mixin = new LinksMixin();

  create(params: {
    processor_id: string;
    name: string;
    webhook_url: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    password?: string;
  }): Link {
    const preparedRequest = this.mixin.prepareCreate(params);
    const response = this._client._preparedRequest(preparedRequest);
    console.log(`Link Created. Url: https://www.retab.com/dashboard/processors/automations/${response.id}`);
    return response as Link;
  }

  list(params: {
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
    processor_id?: string;
    name?: string;
  } = {}): ListLinks {
    const preparedRequest = this.mixin.prepareList(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response as ListLinks;
  }

  get(link_id: string): Link {
    const preparedRequest = this.mixin.prepareGet(link_id);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Link;
  }

  update(params: {
    link_id: string;
    name?: string;
    webhook_url?: string;
    webhook_headers?: Record<string, string>;
    password?: string;
    need_validation?: boolean;
  }): Link {
    const preparedRequest = this.mixin.prepareUpdate(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Link;
  }

  delete(link_id: string): void {
    const preparedRequest = this.mixin.prepareDelete(link_id);
    this._client._preparedRequest(preparedRequest);
  }
}

export class AsyncLinks extends AsyncAPIResource {
  mixin = new LinksMixin();

  async create(params: {
    processor_id: string;
    name: string;
    webhook_url: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    password?: string;
  }): Promise<Link> {
    const preparedRequest = this.mixin.prepareCreate(params);
    const response = await this._client._preparedRequest(preparedRequest);
    console.log(`Link Created. Url: https://www.retab.com/dashboard/processors/automations/${response.id}`);
    return response as Link;
  }

  async list(params: {
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
    processor_id?: string;
    name?: string;
  } = {}): Promise<ListLinks> {
    const preparedRequest = this.mixin.prepareList(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as ListLinks;
  }

  async get(link_id: string): Promise<Link> {
    const preparedRequest = this.mixin.prepareGet(link_id);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as Link;
  }

  async update(params: {
    link_id: string;
    name?: string;
    webhook_url?: string;
    webhook_headers?: Record<string, string>;
    password?: string;
    need_validation?: boolean;
  }): Promise<Link> {
    const preparedRequest = this.mixin.prepareUpdate(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as Link;
  }

  async delete(link_id: string): Promise<void> {
    const preparedRequest = this.mixin.prepareDelete(link_id);
    await this._client._preparedRequest(preparedRequest);
  }
}