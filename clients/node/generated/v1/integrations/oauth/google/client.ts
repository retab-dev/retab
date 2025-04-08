import { AbstractClient, CompositionClient } from '@/client';
import APIGetAuthUrl from "./getAuthUrl/client";
import APIGetToken from "./getToken/client";

export default class APIGoogle extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  getAuthUrl = new APIGetAuthUrl(this);
  getToken = new APIGetToken(this);

}
