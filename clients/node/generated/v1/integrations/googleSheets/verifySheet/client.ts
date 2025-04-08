import { AbstractClient, CompositionClient } from '@/client';
import APISheetId from "./sheetId/client";

export default class APIVerifySheet extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  sheetId = new APISheetId(this);

}
