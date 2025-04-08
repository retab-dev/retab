import { AbstractClient, CompositionClient } from '@/client';
import APICheckOauthToken from "./checkOauthToken/client";
import APIOauth from "./oauth/client";
import APIGoogleSheets from "./googleSheets/client";

export default class APIIntegrations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  checkOauthToken = new APICheckOauthToken(this);
  oauth = new APIOauth(this);
  googleSheets = new APIGoogleSheets(this);

}
