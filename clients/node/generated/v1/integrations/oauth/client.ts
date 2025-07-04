import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIGoogleSub from "./google/client";

export default class APIOauth extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  google = new APIGoogleSub(this._client);

}
