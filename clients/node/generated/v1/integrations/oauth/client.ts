import { AbstractClient, CompositionClient } from '@/client';
import APIGoogle from "./google/client";

export default class APIOauth extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  google = new APIGoogle(this);

}
