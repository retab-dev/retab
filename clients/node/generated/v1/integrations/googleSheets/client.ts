import { AbstractClient, CompositionClient } from '@/client';
import APICreateSheetWithStoredToken from "./createSheetWithStoredToken/client";
import APIListSheets from "./listSheets/client";
import APIOrganizationId from "./organizationId/client";
import APIVerifySheet from "./verifySheet/client";
import APIWebhookVanilla from "./webhookVanilla/client";
import APIWebhook from "./webhook/client";
import APIWebhookCloudflareMichael from "./webhookCloudflareMichael/client";

export default class APIGoogleSheets extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  createSheetWithStoredToken = new APICreateSheetWithStoredToken(this);
  listSheets = new APIListSheets(this);
  organizationId = new APIOrganizationId(this);
  verifySheet = new APIVerifySheet(this);
  webhookVanilla = new APIWebhookVanilla(this);
  webhook = new APIWebhook(this);
  webhookCloudflareMichael = new APIWebhookCloudflareMichael(this);

}
