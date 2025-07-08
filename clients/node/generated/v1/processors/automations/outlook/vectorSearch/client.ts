import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIVectorSearchSub from "./vectorSearch/client";
import APIFileScoresSub from "./fileScores/client";
import APIAutomationDecisionSub from "./automationDecision/client";

export default class APIVectorSearch extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  vectorSearch = new APIVectorSearchSub(this._client);
  fileScores = new APIFileScoresSub(this._client);
  automationDecision = new APIAutomationDecisionSub(this._client);

}
