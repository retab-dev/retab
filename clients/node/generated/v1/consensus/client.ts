import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APICompletionsSub from "./completions/client";
import APIResponsesSub from "./responses/client";
import APIReconcileSub from "./reconcile/client";
import APIAlignDictsSub from "./alignDicts/client";
import APISimilaritySub from "./similarity/client";

export default class APIConsensus extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  completions = new APICompletionsSub(this._client);
  responses = new APIResponsesSub(this._client);
  reconcile = new APIReconcileSub(this._client);
  alignDicts = new APIAlignDictsSub(this._client);
  similarity = new APISimilaritySub(this._client);

}
