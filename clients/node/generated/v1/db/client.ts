import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIDatasetsSub from "./datasets/client";
import APIFilesSub from "./files/client";
import APIEvalsSub from "./evals/client";

export default class APIDb extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  datasets = new APIDatasetsSub(this._client);
  files = new APIFilesSub(this._client);
  evals = new APIEvalsSub(this._client);

}
