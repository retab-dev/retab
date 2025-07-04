import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIEvaluationIdSub from "./evaluationId/client";

export default class APICreate extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  evaluationId = new APIEvaluationIdSub(this._client);

}
