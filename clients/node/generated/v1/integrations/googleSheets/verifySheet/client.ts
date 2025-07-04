import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APISheetIdSub from "./sheetId/client";

export default class APIVerifySheet extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  sheetId = new APISheetIdSub(this._client);

}
