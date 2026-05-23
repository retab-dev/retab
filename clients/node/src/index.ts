import APIV1 from './api/client.js';
import { FetcherClient } from './client.js';

export * from './types';
export * from './client';
// The new auto-paginating list class. `ListMetadata` is already re-exported
// via `./types` (same shape — re-exporting it here would be a duplicate).
export { PaginatedList } from './api/_pagination.js';
export type { PaginatedListInit } from './api/_pagination.js';

export interface ClientOptions {
  baseUrl?: string;
  apiKey?: string;
}

export class Retab extends APIV1 {
  constructor(options?: ClientOptions) {
    super(new FetcherClient(options));
  }
}
