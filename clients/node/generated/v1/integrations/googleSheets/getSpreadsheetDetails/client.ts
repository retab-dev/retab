import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APISpreadsheetIdSub from "./spreadsheetId/client";

export default class APIGetSpreadsheetDetails extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  spreadsheetId = new APISpreadsheetIdSub(this._client);

}
