import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIEvaluationIdSub from "./evaluationId/client";

export default class APIList extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  evaluationId = new APIEvaluationIdSub(this._client);

}
