import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIProcessorIdSub from "./processorId/client";

export default class APIProcessorId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  processorId = new APIProcessorIdSub(this._client);

}
