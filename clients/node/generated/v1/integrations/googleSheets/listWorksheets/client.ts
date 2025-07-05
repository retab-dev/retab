import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APISpreadsheetIdSub from "./spreadsheetId/client";

export default class APIListWorksheets extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  spreadsheetId = new APISpreadsheetIdSub(this._client);

}
