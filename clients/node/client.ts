export interface AbstractClient {
  _fetch<T>(params: {
    url: string,
    method: string,
    params?: Record<string, any>,
    headers?: Record<string, any>,
    bodyMime?: "application/json" | "multipart/form-data",
    body?: Record<string, any>,
    auth?: string[],
  }): Promise<T>
}

export class CompositionClient implements AbstractClient {
  protected _client: AbstractClient;
  constructor(client: AbstractClient) {
    this._client = client;
  }
  protected _fetch<T>(params: {
    url: string,
    method: string,
    params?: Record<string, any>,
    headers?: Record<string, any>,
    bodyMime?: "application/json" | "multipart/form-data",
    body?: Record<string, any>,
    auth?: string[],
  }): Promise<T> {
    return this._client._fetch(params);
  }
}

export class APIError extends Error {
  status: number;
  info: object;
  constructor(status: number, info: object) {
    super(`API Error ${status}: ${JSON.stringify(info)}`);
    this.status = status;
    this.info = info;
  }
}

