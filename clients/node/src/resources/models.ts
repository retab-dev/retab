import { SyncAPIResource, AsyncAPIResource } from '../resource.js';
import { PreparedRequest } from '../types/standards.js';

// Model interface to match OpenAI's Model type structure
export interface Model {
  id: string;
  object: 'model';
  created: number;
  owned_by: string;
  permission?: any[];
  root?: string;
  parent?: string;
}

export interface ModelListResponse {
  data: Model[];
  object: 'list';
}

export class ModelsMixin {
  prepareList(params: {
    supports_finetuning?: boolean;
    supports_image?: boolean;
    include_finetuned_models?: boolean;
  } = {}): PreparedRequest {
    const {
      supports_finetuning = false,
      supports_image = false,
      include_finetuned_models = true,
    } = params;

    const queryParams = {
      supports_finetuning,
      supports_image,
      include_finetuned_models,
    };

    return {
      method: 'GET',
      url: '/v1/models',
      params: queryParams,
    };
  }
}

export class Models extends SyncAPIResource {
  private mixin = new ModelsMixin();

  /**
   * List all available models.
   *
   * @param supports_finetuning - Filter for models that support fine-tuning
   * @param supports_image - Filter for models that support image inputs
   * @param include_finetuned_models - Include fine-tuned models in results
   * @returns Promise<Model[]> - List of available models
   * @throws HTTPException if the request fails
   */
  list(params: {
    supports_finetuning?: boolean;
    supports_image?: boolean;
    include_finetuned_models?: boolean;
  } = {}): Promise<Model[]> {
    const request = this.mixin.prepareList(params);
    const response = this._client._preparedRequest(request) as Promise<ModelListResponse>;
    return response.then(output => output.data.map(model => ({
      id: model.id,
      object: 'model' as const,
      created: model.created,
      owned_by: model.owned_by,
      permission: model.permission,
      root: model.root,
      parent: model.parent,
    })));
  }
}

export class AsyncModels extends AsyncAPIResource {
  private mixin = new ModelsMixin();

  /**
   * List all available models asynchronously.
   *
   * @param supports_finetuning - Filter for models that support fine-tuning
   * @param supports_image - Filter for models that support image inputs
   * @param include_finetuned_models - Include fine-tuned models in results
   * @returns Promise<Model[]> - List of available models
   * @throws HTTPException if the request fails
   */
  async list(params: {
    supports_finetuning?: boolean;
    supports_image?: boolean;
    include_finetuned_models?: boolean;
  } = {}): Promise<Model[]> {
    const request = this.mixin.prepareList(params);
    const output = await this._client._preparedRequest(request) as ModelListResponse;
    return output.data.map(model => ({
      id: model.id,
      object: 'model' as const,
      created: model.created,
      owned_by: model.owned_by,
      permission: model.permission,
      root: model.root,
      parent: model.parent,
    }));
  }
}