import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIEvaluationIdSub from "./evaluationId/client";

export default class APICreate extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  evaluationId = new APIEvaluationIdSub(this._client);

}
