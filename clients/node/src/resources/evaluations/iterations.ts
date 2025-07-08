import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';
import { 
  Iteration,
  CreateIterationRequest,
  PatchIterationRequest,
  ProcessIterationRequest,
  IterationDocumentStatusResponse,
  AddIterationFromJsonlRequest
} from '../../types/evaluations/iterations.js';

export class IterationsMixin {
  prepareCreate(evaluation_id: string, params: CreateIterationRequest): any {
    return {
      method: 'POST' as const,
      url: `/v1/evaluations/${evaluation_id}/iterations`,
      data: params,
    };
  }

  prepareList(evaluation_id: string, params: {
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
  } = {}): any {
    const {
      before,
      after,
      limit = 10,
      order = 'desc',
    } = params;

    const queryParams: Record<string, any> = {};
    if (before !== undefined) queryParams.before = before;
    if (after !== undefined) queryParams.after = after;
    if (limit !== undefined) queryParams.limit = limit;
    if (order !== undefined) queryParams.order = order;

    return {
      method: 'GET' as const,
      url: `/v1/evaluations/${evaluation_id}/iterations`,
      params: queryParams,
    };
  }

  prepareGet(evaluation_id: string, iteration_id: string): any {
    return {
      method: 'GET' as const,
      url: `/v1/evaluations/${evaluation_id}/iterations/${iteration_id}`,
    };
  }

  preparePatch(evaluation_id: string, iteration_id: string, params: PatchIterationRequest): any {
    return {
      method: 'PATCH' as const,
      url: `/v1/evaluations/${evaluation_id}/iterations/${iteration_id}`,
      data: params,
    };
  }

  prepareDelete(evaluation_id: string, iteration_id: string): any {
    return {
      method: 'DELETE' as const,
      url: `/v1/evaluations/${evaluation_id}/iterations/${iteration_id}`,
    };
  }

  prepareProcess(evaluation_id: string, iteration_id: string, params: ProcessIterationRequest): any {
    return {
      method: 'POST' as const,
      url: `/v1/evaluations/${evaluation_id}/iterations/${iteration_id}/process`,
      data: params,
    };
  }

  prepareDocumentStatus(evaluation_id: string, iteration_id: string): any {
    return {
      method: 'GET' as const,
      url: `/v1/evaluations/${evaluation_id}/iterations/${iteration_id}/document-status`,
    };
  }

  prepareAddFromJsonl(evaluation_id: string, params: AddIterationFromJsonlRequest): any {
    return {
      method: 'POST' as const,
      url: `/v1/evaluations/${evaluation_id}/iterations/add-from-jsonl`,
      data: params,
    };
  }
}

export class Iterations extends SyncAPIResource {
  mixin = new IterationsMixin();

  create(evaluation_id: string, params: CreateIterationRequest): any {
    const preparedRequest = this.mixin.prepareCreate(evaluation_id, params);
    const response = this._client._preparedRequest(preparedRequest);
    return response;
  }

  list(evaluation_id: string, params: {
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
  } = {}): any {
    const preparedRequest = this.mixin.prepareList(evaluation_id, params);
    const response = this._client._preparedRequest(preparedRequest);
    return response;
  }

  get(evaluation_id: string, iteration_id: string): Promise<Iteration> {
    const preparedRequest = this.mixin.prepareGet(evaluation_id, iteration_id);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Promise<Iteration>;
  }

  patch(evaluation_id: string, iteration_id: string, params: PatchIterationRequest): Promise<Iteration> {
    const preparedRequest = this.mixin.preparePatch(evaluation_id, iteration_id, params);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Promise<Iteration>;
  }

  delete(evaluation_id: string, iteration_id: string): void {
    const preparedRequest = this.mixin.prepareDelete(evaluation_id, iteration_id);
    this._client._preparedRequest(preparedRequest);
  }

  process(evaluation_id: string, iteration_id: string, params: ProcessIterationRequest): any {
    const preparedRequest = this.mixin.prepareProcess(evaluation_id, iteration_id, params);
    const response = this._client._preparedRequest(preparedRequest);
    return response;
  }

  documentStatus(evaluation_id: string, iteration_id: string): Promise<IterationDocumentStatusResponse> {
    const preparedRequest = this.mixin.prepareDocumentStatus(evaluation_id, iteration_id);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Promise<IterationDocumentStatusResponse>;
  }

  addFromJsonl(evaluation_id: string, params: AddIterationFromJsonlRequest): Promise<Iteration> {
    const preparedRequest = this.mixin.prepareAddFromJsonl(evaluation_id, params);
    const response = this._client._preparedRequest(preparedRequest);
    return response as Promise<Iteration>;
  }
}

export class AsyncIterations extends AsyncAPIResource {
  mixin = new IterationsMixin();

  async create(evaluation_id: string, params: CreateIterationRequest): Promise<Iteration> {
    const preparedRequest = this.mixin.prepareCreate(evaluation_id, params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as Iteration;
  }

  async list(evaluation_id: string, params: {
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
  } = {}): Promise<any> {
    const preparedRequest = this.mixin.prepareList(evaluation_id, params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response;
  }

  async get(evaluation_id: string, iteration_id: string): Promise<Iteration> {
    const preparedRequest = this.mixin.prepareGet(evaluation_id, iteration_id);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as Iteration;
  }

  async patch(evaluation_id: string, iteration_id: string, params: PatchIterationRequest): Promise<Iteration> {
    const preparedRequest = this.mixin.preparePatch(evaluation_id, iteration_id, params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as Iteration;
  }

  async delete(evaluation_id: string, iteration_id: string): Promise<void> {
    const preparedRequest = this.mixin.prepareDelete(evaluation_id, iteration_id);
    await this._client._preparedRequest(preparedRequest);
  }

  async process(evaluation_id: string, iteration_id: string, params: ProcessIterationRequest): Promise<any> {
    const preparedRequest = this.mixin.prepareProcess(evaluation_id, iteration_id, params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response;
  }

  async documentStatus(evaluation_id: string, iteration_id: string): Promise<IterationDocumentStatusResponse> {
    const preparedRequest = this.mixin.prepareDocumentStatus(evaluation_id, iteration_id);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as IterationDocumentStatusResponse;
  }

  async addFromJsonl(evaluation_id: string, params: AddIterationFromJsonlRequest): Promise<Iteration> {
    const preparedRequest = this.mixin.prepareAddFromJsonl(evaluation_id, params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as Iteration;
  }
}