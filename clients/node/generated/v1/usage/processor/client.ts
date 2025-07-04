import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIProcessorIdSub from "./processorId/client";

export default class APIProcessor extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  processorId = new APIProcessorIdSub(this._client);

}
