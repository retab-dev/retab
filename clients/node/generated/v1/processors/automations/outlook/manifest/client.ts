import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIIdSub from "./id/client";

export default class APIManifest extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  id = new APIIdSub(this._client);

}
