import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIFileIdSub from "./fileId/client";

export default class APIFileScores extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  fileId = new APIFileIdSub(this._client);

}
