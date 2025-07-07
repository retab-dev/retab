import { SyncAPIResource, AsyncAPIResource } from '../../../resource.js';

export class TestsMixin {
  prepareUpload(params: {
    automation_id: string;
    document: string;
    user_email?: string;
  }): any {
    const { automation_id, document, user_email } = params;

    const formData: Record<string, any> = {
      automation_id,
    };

    if (user_email !== undefined) {
      formData.user_email = user_email;
    }

    return {
      method: 'POST' as const,
      url: '/v1/processors/automations/tests/upload',
      form_data: formData,
      files: { document },
    };
  }

  prepareForward(params: {
    automation_id: string;
    email_file: string;
  }): any {
    const { automation_id, email_file } = params;

    const formData: Record<string, any> = {
      automation_id,
    };

    return {
      method: 'POST' as const,
      url: '/v1/processors/automations/tests/forward',
      form_data: formData,
      files: { email: email_file },
    };
  }

  prepareWebhook(params: {
    automation_id: string;
    completion: any;
    user?: string;
    file_payload: any;
    metadata?: Record<string, any>;
  }): any {
    const { automation_id, completion, user, file_payload, metadata } = params;

    const requestBody: Record<string, any> = {
      completion,
      file_payload,
    };

    if (user !== undefined) {
      requestBody.user = user;
    }
    if (metadata !== undefined) {
      requestBody.metadata = metadata;
    }

    return {
      method: 'POST' as const,
      url: `/v1/processors/automations/tests/webhook/${automation_id}`,
      data: requestBody,
    };
  }
}

export class Tests extends SyncAPIResource {
  mixin = new TestsMixin();

  upload(params: {
    automation_id: string;
    document: string;
    user_email?: string;
  }): any {
    const preparedRequest = this.mixin.prepareUpload(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response;
  }

  forward(params: {
    automation_id: string;
    email_file: string;
  }): any {
    const preparedRequest = this.mixin.prepareForward(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response;
  }

  webhook(params: {
    automation_id: string;
    completion: any;
    user?: string;
    file_payload: any;
    metadata?: Record<string, any>;
  }): any {
    const preparedRequest = this.mixin.prepareWebhook(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response;
  }
}

export class AsyncTests extends AsyncAPIResource {
  mixin = new TestsMixin();

  async upload(params: {
    automation_id: string;
    document: string;
    user_email?: string;
  }): Promise<any> {
    const preparedRequest = this.mixin.prepareUpload(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response;
  }

  async forward(params: {
    automation_id: string;
    email_file: string;
  }): Promise<any> {
    const preparedRequest = this.mixin.prepareForward(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response;
  }

  async webhook(params: {
    automation_id: string;
    completion: any;
    user?: string;
    file_payload: any;
    metadata?: Record<string, any>;
  }): Promise<any> {
    const preparedRequest = this.mixin.prepareWebhook(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response;
  }
}