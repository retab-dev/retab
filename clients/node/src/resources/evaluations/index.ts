import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';
import { Documents, AsyncDocuments } from './documents.js';
import { Iterations, AsyncIterations } from './iterations.js';

export class Evaluations extends SyncAPIResource {
  public documents: Documents;
  public iterations: Iterations;

  constructor(client: any) {
    super(client);
    this.documents = new Documents(client);
    this.iterations = new Iterations(client);
  }
}

export class AsyncEvaluations extends AsyncAPIResource {
  public documents: AsyncDocuments;
  public iterations: AsyncIterations;

  constructor(client: any) {
    super(client);
    this.documents = new AsyncDocuments(client);
    this.iterations = new AsyncIterations(client);
  }
}