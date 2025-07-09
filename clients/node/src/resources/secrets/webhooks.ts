import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';
import { PreparedRequest } from '../../types/standards.js';

export interface WebhookSecret {
  id: string;
  name: string;
  secret: string;
  algorithm: 'sha256' | 'sha1' | 'md5';
  created_at: string;
  updated_at: string;
}

export interface CreateWebhookSecretRequest {
  name: string;
  secret?: string; // Auto-generated if not provided
  algorithm?: 'sha256' | 'sha1' | 'md5';
}

export interface UpdateWebhookSecretRequest {
  name?: string;
  secret?: string;
  algorithm?: 'sha256' | 'sha1' | 'md5';
}

export interface WebhookSignatureConfig {
  secret: string;
  algorithm: 'sha256' | 'sha1' | 'md5';
  headerName?: string;
  prefix?: string;
}

export class WebhooksMixin {
  prepareCreate(params: CreateWebhookSecretRequest): PreparedRequest {
    return {
      method: 'POST',
      url: '/v1/secrets/webhooks',
      data: {
        name: params.name,
        secret: params.secret,
        algorithm: params.algorithm || 'sha256',
      },
    };
  }

  prepareList(): PreparedRequest {
    return {
      method: 'GET',
      url: '/v1/secrets/webhooks',
    };
  }

  prepareGet(webhookId: string): PreparedRequest {
    return {
      method: 'GET',
      url: `/v1/secrets/webhooks/${webhookId}`,
    };
  }

  prepareUpdate(webhookId: string, params: UpdateWebhookSecretRequest): PreparedRequest {
    return {
      method: 'PATCH',
      url: `/v1/secrets/webhooks/${webhookId}`,
      data: params,
    };
  }

  prepareDelete(webhookId: string): PreparedRequest {
    return {
      method: 'DELETE',
      url: `/v1/secrets/webhooks/${webhookId}`,
    };
  }

  prepareRotate(webhookId: string): PreparedRequest {
    return {
      method: 'POST',
      url: `/v1/secrets/webhooks/${webhookId}/rotate`,
    };
  }
}

export class Webhooks extends SyncAPIResource {
  private mixin = new WebhooksMixin();

  create(params: CreateWebhookSecretRequest): Promise<WebhookSecret> {
    const preparedRequest = this.mixin.prepareCreate(params);
    return this._client._preparedRequest(preparedRequest);
  }

  list(): Promise<{ data: WebhookSecret[] }> {
    const preparedRequest = this.mixin.prepareList();
    return this._client._preparedRequest(preparedRequest);
  }

  get(webhookId: string): Promise<WebhookSecret> {
    const preparedRequest = this.mixin.prepareGet(webhookId);
    return this._client._preparedRequest(preparedRequest);
  }

  update(webhookId: string, params: UpdateWebhookSecretRequest): Promise<WebhookSecret> {
    const preparedRequest = this.mixin.prepareUpdate(webhookId, params);
    return this._client._preparedRequest(preparedRequest);
  }

  delete(webhookId: string): Promise<void> {
    const preparedRequest = this.mixin.prepareDelete(webhookId);
    return this._client._preparedRequest(preparedRequest);
  }

  rotate(webhookId: string): Promise<WebhookSecret> {
    const preparedRequest = this.mixin.prepareRotate(webhookId);
    return this._client._preparedRequest(preparedRequest);
  }

  /**
   * Generate webhook signature for verification
   */
  generateSignature(payload: string, config: WebhookSignatureConfig): string {
    const crypto = require('crypto');
    const { secret, algorithm, prefix = '' } = config;
    
    const hmac = crypto.createHmac(algorithm, secret);
    hmac.update(payload);
    const signature = hmac.digest('hex');
    
    return prefix ? `${prefix}${signature}` : signature;
  }

  /**
   * Verify webhook signature
   */
  verifySignature(
    payload: string,
    receivedSignature: string,
    config: WebhookSignatureConfig
  ): boolean {
    const expectedSignature = this.generateSignature(payload, config);
    
    // Use timingSafeEqual to prevent timing attacks
    const crypto = require('crypto');
    return crypto.timingSafeEqual(
      Buffer.from(receivedSignature),
      Buffer.from(expectedSignature)
    );
  }
}

export class AsyncWebhooks extends AsyncAPIResource {
  private mixin = new WebhooksMixin();

  async create(params: CreateWebhookSecretRequest): Promise<WebhookSecret> {
    const preparedRequest = this.mixin.prepareCreate(params);
    return await this._client._preparedRequest(preparedRequest);
  }

  async list(): Promise<{ data: WebhookSecret[] }> {
    const preparedRequest = this.mixin.prepareList();
    return await this._client._preparedRequest(preparedRequest);
  }

  async get(webhookId: string): Promise<WebhookSecret> {
    const preparedRequest = this.mixin.prepareGet(webhookId);
    return await this._client._preparedRequest(preparedRequest);
  }

  async update(webhookId: string, params: UpdateWebhookSecretRequest): Promise<WebhookSecret> {
    const preparedRequest = this.mixin.prepareUpdate(webhookId, params);
    return await this._client._preparedRequest(preparedRequest);
  }

  async delete(webhookId: string): Promise<void> {
    const preparedRequest = this.mixin.prepareDelete(webhookId);
    return await this._client._preparedRequest(preparedRequest);
  }

  async rotate(webhookId: string): Promise<WebhookSecret> {
    const preparedRequest = this.mixin.prepareRotate(webhookId);
    return await this._client._preparedRequest(preparedRequest);
  }

  /**
   * Generate webhook signature for verification
   */
  generateSignature(payload: string, config: WebhookSignatureConfig): string {
    const crypto = require('crypto');
    const { secret, algorithm, prefix = '' } = config;
    
    const hmac = crypto.createHmac(algorithm, secret);
    hmac.update(payload);
    const signature = hmac.digest('hex');
    
    return prefix ? `${prefix}${signature}` : signature;
  }

  /**
   * Verify webhook signature
   */
  verifySignature(
    payload: string,
    receivedSignature: string,
    config: WebhookSignatureConfig
  ): boolean {
    const expectedSignature = this.generateSignature(payload, config);
    
    // Use timingSafeEqual to prevent timing attacks
    const crypto = require('crypto');
    return crypto.timingSafeEqual(
      Buffer.from(receivedSignature),
      Buffer.from(expectedSignature)
    );
  }
}