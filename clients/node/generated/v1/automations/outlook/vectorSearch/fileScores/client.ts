import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIFileIdSub from "./fileId/client";

export default class APIFileScores extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  fileId = new APIFileIdSub(this._client);

}
