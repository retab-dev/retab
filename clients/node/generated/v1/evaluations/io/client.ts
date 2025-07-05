import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIEvaluationIdSub from "./evaluationId/client";
import APIGetPayloadForExportSub from "./getPayloadForExport/client";

export default class APIIo extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  evaluationId = new APIEvaluationIdSub(this._client);
  getPayloadForExport = new APIGetPayloadForExportSub(this._client);

}
