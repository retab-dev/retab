import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIModelsSub from "./models/client";
import APISchemasSub from "./schemas/client";
import APIEvalsSub from "./evals/client";
import APIEvaluationsSub from "./evaluations/client";
import APIIntegrationsSub from "./integrations/client";
import APIBenchmarkingSub from "./benchmarking/client";
import APIConsensusSub from "./consensus/client";
import APIEventsSub from "./events/client";
import APIExtractionsSub from "./extractions/client";
import APIProcessorsSub from "./processors/client";
import APIUsageSub from "./usage/client";
import APIEndpointsSub from "./endpoints/client";
import APIDocumentsSub from "./documents/client";
import APISecretsSub from "./secrets/client";
import APITestIngestCompletionSub from "./testIngestCompletion/client";

export default class APIV1 extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  models = new APIModelsSub(this._client);
  schemas = new APISchemasSub(this._client);
  evals = new APIEvalsSub(this._client);
  evaluations = new APIEvaluationsSub(this._client);
  integrations = new APIIntegrationsSub(this._client);
  benchmarking = new APIBenchmarkingSub(this._client);
  consensus = new APIConsensusSub(this._client);
  events = new APIEventsSub(this._client);
  extractions = new APIExtractionsSub(this._client);
  processors = new APIProcessorsSub(this._client);
  usage = new APIUsageSub(this._client);
  endpoints = new APIEndpointsSub(this._client);
  documents = new APIDocumentsSub(this._client);
  secrets = new APISecretsSub(this._client);
  testIngestCompletion = new APITestIngestCompletionSub(this._client);

}
