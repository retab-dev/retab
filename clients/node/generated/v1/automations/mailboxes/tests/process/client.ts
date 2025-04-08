import { AbstractClient, CompositionClient } from '@/client';
import APIEmail from "./email/client";

export default class APIProcess extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  email = new APIEmail(this);

}
