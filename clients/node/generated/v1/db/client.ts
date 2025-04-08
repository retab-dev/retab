import { AbstractClient, CompositionClient } from '@/client';
import APIDatasets from "./datasets/client";
import APIFiles from "./files/client";
import APIEvals from "./evals/client";

export default class APIDb extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  datasets = new APIDatasets(this);
  files = new APIFiles(this);
  evals = new APIEvals(this);

}
