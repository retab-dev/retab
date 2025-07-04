import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIParseSub from "./parse/client";

export default class APICompletions extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  parse = new APIParseSub(this._client);

}
