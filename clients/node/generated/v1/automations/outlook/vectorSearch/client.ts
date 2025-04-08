import { AbstractClient, CompositionClient } from '@/client';
import APIVectorSearch from "./vectorSearch/client";
import APIFileScores from "./fileScores/client";
import APIAutomationDecision from "./automationDecision/client";

export default class APIVectorSearch extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  vectorSearch = new APIVectorSearch(this);
  fileScores = new APIFileScores(this);
  automationDecision = new APIAutomationDecision(this);

}
