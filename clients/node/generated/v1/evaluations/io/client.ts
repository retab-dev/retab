import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIEvaluationIdSub from "./evaluationId/client";
import APIGetPayloadForExportSub from "./getPayloadForExport/client";

export default class APIIo extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  evaluationId = new APIEvaluationIdSub(this._client);
  getPayloadForExport = new APIGetPayloadForExportSub(this._client);

}
