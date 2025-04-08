import { AbstractClient, CompositionClient } from '@/client';
import APIV1 from "./v1/client";

export default class APIGenerated extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  v1 = new APIV1(this);

}
