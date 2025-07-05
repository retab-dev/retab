export class AbstractClient {
  protected _fetch(params: {
    url: string,
    method: string,
    params?: Record<string, any>,
    headers?: Record<string, any>,
    bodyMime?: "application/json" | "multipart/form-data",
    body?: Record<string, any>,
    auth?: string[],
  }): Promise<Response> {
    throw new Error("Method not implemented");
  }
}

export class CompositionClient extends AbstractClient {
  protected _client: AbstractClient;
  constructor(client: AbstractClient) {
    super();
    this._client = client;
  }
  protected _fetch(params: {
    url: string,
    method: string,
    params?: Record<string, any>,
    headers?: Record<string, any>,
    bodyMime?: "application/json" | "multipart/form-data",
    body?: Record<string, any>,
    auth?: string[],
  }): Promise<Response> {
    return this._client["_fetch"](params);
  }
}

export class APIError extends Error {
  status: number;
  info: string;
  constructor(status: number, info: string) {
    super(`API Error ${status}: ${info}`);
    this.status = status;
    this.info = info;
  }
}

export async function* streamResponse<T>(response: Response): AsyncGenerator<T> {
  let body = "";
  let depth = 0;
  let inString = false;
  if (!response.body) {
    throw new APIError(response.status, "Response body is empty");
  }
  let reader = response.body.getReader();
  while (true) {
    let chunk = await reader.read();
    if (!chunk.value) break;
    let string = String.fromCharCode(...chunk.value);
    let prevBodyLength = body.length;
    body += string;
    for (let i = 0; i < string.length; i++) {
      let char = string[i];
      if (char === '"') {
        inString = !inString;
      }
      if (inString) {
        if (char === "\\") {
          i++;
        }
      } else {
        if (char === "{") {
          depth++;
        } else if (char === "}") {
          depth--;
          if (depth === 0) {
            yield JSON.parse(body.slice(0, prevBodyLength + i + 1)) as T;
            body = body.slice(prevBodyLength + i + 1);
            prevBodyLength = -i - 1;
          }
        }
      }
    }
    if (chunk.done) break;
  }
}

