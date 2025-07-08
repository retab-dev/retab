import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APICreateSpreadsheetWithStoredTokenSub from "./createSpreadsheetWithStoredToken/client";
import APIListSpreadsheetsSub from "./listSpreadsheets/client";
import APIOrganizationIdSub from "./organizationId/client";
import APIVerifySpreadsheetSub from "./verifySpreadsheet/client";
import APIListWorksheetsSub from "./listWorksheets/client";
import APIGetSpreadsheetDetailsSub from "./getSpreadsheetDetails/client";
import APIWebhookSub from "./webhook/client";
import APIExportToCsvSub from "./exportToCsv/client";

export default class APIGoogleSheets extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  createSpreadsheetWithStoredToken = new APICreateSpreadsheetWithStoredTokenSub(this._client);
  listSpreadsheets = new APIListSpreadsheetsSub(this._client);
  organizationId = new APIOrganizationIdSub(this._client);
  verifySpreadsheet = new APIVerifySpreadsheetSub(this._client);
  listWorksheets = new APIListWorksheetsSub(this._client);
  getSpreadsheetDetails = new APIGetSpreadsheetDetailsSub(this._client);
  webhook = new APIWebhookSub(this._client);
  exportToCsv = new APIExportToCsvSub(this._client);

}
