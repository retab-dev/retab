import { AbstractClient, CompositionClient } from '@/client';
import APICheckOauthTokenSub from "./checkOauthToken/client";
import APIOauthSub from "./oauth/client";
import APIGoogleSheetsSub from "./googleSheets/client";

export default class APIIntegrations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  checkOauthToken = new APICheckOauthTokenSub(this._client);
  oauth = new APIOauthSub(this._client);
  googleSheets = new APIGoogleSheetsSub(this._client);

}
