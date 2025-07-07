import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';
import { Documents, AsyncDocuments } from './documents.js';

export class Evaluations extends SyncAPIResource {
  public documents: Documents;

  constructor(client: any) {
    super(client);
    this.documents = new Documents(client);
  }
}

export class AsyncEvaluations extends AsyncAPIResource {
  public documents: AsyncDocuments;

  constructor(client: any) {
    super(client);
    this.documents = new AsyncDocuments(client);
  }
}