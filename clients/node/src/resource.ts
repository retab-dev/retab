import type { Retab, AsyncRetab } from './client.js';

export class SyncAPIResource {
  protected _client: Retab;

  constructor(client: Retab) {
    this._client = client;
  }

  protected _sleep(seconds: number): void {
    const start = Date.now();
    while (Date.now() - start < seconds * 1000) {
      // Busy wait
    }
  }
}

export class AsyncAPIResource {
  protected _client: AsyncRetab;

  constructor(client: AsyncRetab) {
    this._client = client;
  }

  protected async _sleep(seconds: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, seconds * 1000));
  }
}