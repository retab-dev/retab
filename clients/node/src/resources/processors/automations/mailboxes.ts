import { SyncAPIResource, AsyncAPIResource } from '../../../resource.js';
import { Mailbox, ListMailboxes } from '../../../types/automations/mailboxes.js';

export class MailboxesMixin {
  readonly mailboxesBaseUrl = '/v1/processors/automations/mailboxes';

  prepareCreate(params: {
    processor_id: string;
    name: string;
    email: string;
    webhook_url: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    authorized_domains?: string[];
    authorized_emails?: string[];
  }): any {
    const {
      processor_id,
      name,
      email,
      webhook_url,
      webhook_headers,
      need_validation,
      authorized_domains,
      authorized_emails,
    } = params;

    const mailboxDict: Record<string, any> = {
      processor_id,
      name,
      email: email.trim().toLowerCase(),
      webhook_url,
    };

    if (webhook_headers !== undefined) {
      mailboxDict.webhook_headers = webhook_headers;
    }
    if (need_validation !== undefined) {
      mailboxDict.need_validation = need_validation;
    }
    if (authorized_domains !== undefined) {
      mailboxDict.authorized_domains = authorized_domains;
    }
    if (authorized_emails !== undefined) {
      mailboxDict.authorized_emails = authorized_emails.map(email => email.trim().toLowerCase());
    }

    return {
      method: 'POST' as const,
      url: this.mailboxesBaseUrl,
      data: mailboxDict,
    };
  }

  prepareList(params: {
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
    processor_id?: string;
    name?: string;
    email?: string;
  } = {}): any {
    const {
      before,
      after,
      limit = 10,
      order = 'desc',
      processor_id,
      name,
      email,
    } = params;

    const queryParams: Record<string, any> = {};
    if (before !== undefined) queryParams.before = before;
    if (after !== undefined) queryParams.after = after;
    if (limit !== undefined) queryParams.limit = limit;
    if (order !== undefined) queryParams.order = order;
    if (processor_id !== undefined) queryParams.processor_id = processor_id;
    if (name !== undefined) queryParams.name = name;
    if (email !== undefined) queryParams.email = email;

    return {
      method: 'GET' as const,
      url: this.mailboxesBaseUrl,
      params: queryParams,
    };
  }

  prepareGet(mailbox_id: string): any {
    return {
      method: 'GET' as const,
      url: `${this.mailboxesBaseUrl}/${mailbox_id}`,
    };
  }

  prepareUpdate(params: {
    mailbox_id: string;
    name?: string;
    webhook_url?: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    authorized_domains?: string[];
    authorized_emails?: string[];
  }): any {
    const {
      mailbox_id,
      name,
      webhook_url,
      webhook_headers,
      need_validation,
      authorized_domains,
      authorized_emails,
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

    return {
      method: 'PUT' as const,
      url: `${this.mailboxesBaseUrl}/${mailbox_id}`,
      data: updateDict,
    };
  }

  prepareDelete(mailbox_id: string): any {
    return {
      method: 'DELETE' as const,
      url: `${this.mailboxesBaseUrl}/${mailbox_id}`,
    };
  }
}

export class Mailboxes extends SyncAPIResource {
  mixin = new MailboxesMixin();

  create(params: {
    processor_id: string;
    name: string;
    email: string;
    webhook_url: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    authorized_domains?: string[];
    authorized_emails?: string[];
  }): Promise<Mailbox> {
    const preparedRequest = this.mixin.prepareCreate(params);
    const response = this._client._preparedRequest(preparedRequest);
    // Note: response is a Promise, access id after awaiting
    response.then((r: any) => {
      console.log(`Mailbox Created. Url: https://www.retab.com/dashboard/processors/automations/${r?.id || 'unknown'}`);
    }).catch(() => {});
    return response as Promise<Mailbox>;
  }

  list(params: {
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
    processor_id?: string;
    name?: string;
    email?: string;
  } = {}): Promise<ListMailboxes> {
    const preparedRequest = this.mixin.prepareList(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Promise<ListMailboxes>;
  }

  get(mailbox_id: string): Promise<Mailbox> {
    const preparedRequest = this.mixin.prepareGet(mailbox_id);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Promise<Mailbox>;
  }

  update(params: {
    mailbox_id: string;
    name?: string;
    webhook_url?: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    authorized_domains?: string[];
    authorized_emails?: string[];
  }): Promise<Mailbox> {
    const preparedRequest = this.mixin.prepareUpdate(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Promise<Mailbox>;
  }

  delete(mailbox_id: string): Promise<void> {
    const preparedRequest = this.mixin.prepareDelete(mailbox_id);
    const response = this._client._preparedRequest(preparedRequest);
    console.log(`Mailbox Deleted. ID: ${mailbox_id}`);
    return response as Promise<void>;
  }
}

export class AsyncMailboxes extends AsyncAPIResource {
  mixin = new MailboxesMixin();

  async create(params: {
    processor_id: string;
    name: string;
    email: string;
    webhook_url: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    authorized_domains?: string[];
    authorized_emails?: string[];
  }): Promise<Mailbox> {
    const preparedRequest = this.mixin.prepareCreate(params);
    const response = await this._client._preparedRequest(preparedRequest);
    console.log(`Mailbox Created. Url: https://www.retab.com/dashboard/processors/automations/${response?.id || 'unknown'}`);
    return response as Mailbox;
  }

  async list(params: {
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
    processor_id?: string;
    name?: string;
    email?: string;
  } = {}): Promise<ListMailboxes> {
    const preparedRequest = this.mixin.prepareList(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as ListMailboxes;
  }

  async get(mailbox_id: string): Promise<Mailbox> {
    const preparedRequest = this.mixin.prepareGet(mailbox_id);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as Mailbox;
  }

  async update(params: {
    mailbox_id: string;
    name?: string;
    webhook_url?: string;
    webhook_headers?: Record<string, string>;
    need_validation?: boolean;
    authorized_domains?: string[];
    authorized_emails?: string[];
  }): Promise<Mailbox> {
    const preparedRequest = this.mixin.prepareUpdate(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as Mailbox;
  }

  async delete(mailbox_id: string): Promise<void> {
    const preparedRequest = this.mixin.prepareDelete(mailbox_id);
    await this._client._preparedRequest(preparedRequest);
    console.log(`Mailbox Deleted. ID: ${mailbox_id}`);
  }
}