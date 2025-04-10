import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APICreateSheetWithStoredTokenSub from "./createSheetWithStoredToken/client";
import APIListSheetsSub from "./listSheets/client";
import APIOrganizationIdSub from "./organizationId/client";
import APIVerifySheetSub from "./verifySheet/client";
import APIWebhookVanillaSub from "./webhookVanilla/client";
import APIWebhookSub from "./webhook/client";
import APIWebhookCloudflareMichaelSub from "./webhookCloudflareMichael/client";

export default class APIGoogleSheets extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  createSheetWithStoredToken = new APICreateSheetWithStoredTokenSub(this._client);
  listSheets = new APIListSheetsSub(this._client);
  organizationId = new APIOrganizationIdSub(this._client);
  verifySheet = new APIVerifySheetSub(this._client);
  webhookVanilla = new APIWebhookVanillaSub(this._client);
  webhook = new APIWebhookSub(this._client);
  webhookCloudflareMichael = new APIWebhookCloudflareMichaelSub(this._client);

}
