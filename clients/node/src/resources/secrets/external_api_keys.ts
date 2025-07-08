import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';
import { PreparedRequest } from '../../types/standards.js';
import { AIProvider } from '../../types/ai_models.js';
import { ExternalAPIKey, ExternalAPIKeyRequest } from '../../types/secrets/external_api_keys.js';

export class ExternalAPIKeysMixin {
  prepareCreate(provider: AIProvider, api_key: string): PreparedRequest {
    const request: ExternalAPIKeyRequest = {
      provider,
      api_key,
    };
    
    return {
      method: 'POST',
      url: '/v1/secrets/external_api_keys',
      data: request,
    };
  }

  prepareGet(provider: AIProvider): PreparedRequest {
    return {
      method: 'GET',
      url: `/v1/secrets/external_api_keys/${provider}`,
    };
  }

  prepareList(): PreparedRequest {
    return {
      method: 'GET',
      url: '/v1/secrets/external_api_keys',
    };
  }

  prepareDelete(provider: AIProvider): PreparedRequest {
    return {
      method: 'DELETE',
      url: `/v1/secrets/external_api_keys/${provider}`,
    };
  }
}

export class ExternalAPIKeys extends SyncAPIResource {
  private mixin = new ExternalAPIKeysMixin();

  /**
   * Add or update an external API key.
   *
   * @param provider - The API provider (OpenAI, Anthropic, Gemini, xAI)
   * @param api_key - The API key to store
   * @returns Response indicating success
   */
  create(provider: AIProvider, api_key: string): Promise<Record<string, any>> {
    const request = this.mixin.prepareCreate(provider, api_key);
    const response = this._client._preparedRequest(request);
    return response as Promise<Record<string, any>>;
  }

  /**
   * Get an external API key configuration.
   *
   * @param provider - The API provider to get the key for
   * @returns The API key configuration
   */
  get(provider: AIProvider): Promise<ExternalAPIKey> {
    const request = this.mixin.prepareGet(provider);
    const response = this._client._preparedRequest(request);
    return response as Promise<ExternalAPIKey>;
  }

  /**
   * List all configured external API keys.
   *
   * @returns List of API key configurations
   */
  list(): Promise<ExternalAPIKey[]> {
    const request = this.mixin.prepareList();
    const response = this._client._preparedRequest(request);
    return response as Promise<ExternalAPIKey[]>;
  }

  /**
   * Delete an external API key configuration.
   *
   * @param provider - The API provider to delete the key for
   * @returns Response indicating success
   */
  delete(provider: AIProvider): Promise<Record<string, any>> {
    const request = this.mixin.prepareDelete(provider);
    const response = this._client._preparedRequest(request);
    return response as Promise<Record<string, any>>;
  }
}

export class AsyncExternalAPIKeys extends AsyncAPIResource {
  private mixin = new ExternalAPIKeysMixin();

  /**
   * Add or update an external API key asynchronously.
   */
  async create(provider: AIProvider, api_key: string): Promise<Record<string, any>> {
    const request = this.mixin.prepareCreate(provider, api_key);
    const response = await this._client._preparedRequest(request);
    return response as Record<string, any>;
  }

  /**
   * Get an external API key configuration asynchronously.
   */
  async get(provider: AIProvider): Promise<ExternalAPIKey> {
    const request = this.mixin.prepareGet(provider);
    const response = await this._client._preparedRequest(request);
    return response as ExternalAPIKey;
  }

  /**
   * List all configured external API keys asynchronously.
   */
  async list(): Promise<ExternalAPIKey[]> {
    const request = this.mixin.prepareList();
    const response = await this._client._preparedRequest(request);
    return response as ExternalAPIKey[];
  }

  /**
   * Delete an external API key configuration asynchronously.
   */
  async delete(provider: AIProvider): Promise<Record<string, any>> {
    const request = this.mixin.prepareDelete(provider);
    const response = await this._client._preparedRequest(request);
    return response as Record<string, any>;
  }
}