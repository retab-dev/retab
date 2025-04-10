import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIDocumentsSub from "./documents/client";
import APISchemasSub from "./schemas/client";
import APIIamSub from "./iam/client";
import APIModelsSub from "./models/client";
import APIAutomationsSub from "./automations/client";
import APIDbSub from "./db/client";
import APISecretsSub from "./secrets/client";
import APIIntegrationsSub from "./integrations/client";
import APIAnalyticsSub from "./analytics/client";
import APIBrandingSub from "./branding/client";
import APIExtractionsLogsSub from "./extractionsLogs/client";
import APIEventsSub from "./events/client";
import APIUsageSub from "./usage/client";
import APIUsageschemaIdSub from "./usageschemaId/client";
import APIBenchmarkingSub from "./benchmarking/client";
import APIFinetunedModelsSub from "./finetunedModels/client";
import APIExperimentsSub from "./experiments/client";
import APICompletionsSub from "./completions/client";
import APIJobSystemSub from "./jobSystem/client";
import APICheckOutgoingIpSub from "./checkOutgoingIp/client";
import APIEndpointIdSub from "./endpointId/client";

export default class APIV1 extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  documents = new APIDocumentsSub(this._client);
  schemas = new APISchemasSub(this._client);
  iam = new APIIamSub(this._client);
  models = new APIModelsSub(this._client);
  automations = new APIAutomationsSub(this._client);
  db = new APIDbSub(this._client);
  secrets = new APISecretsSub(this._client);
  integrations = new APIIntegrationsSub(this._client);
  analytics = new APIAnalyticsSub(this._client);
  branding = new APIBrandingSub(this._client);
  extractionsLogs = new APIExtractionsLogsSub(this._client);
  events = new APIEventsSub(this._client);
  usage = new APIUsageSub(this._client);
  usageschemaId = new APIUsageschemaIdSub(this._client);
  benchmarking = new APIBenchmarkingSub(this._client);
  finetunedModels = new APIFinetunedModelsSub(this._client);
  experiments = new APIExperimentsSub(this._client);
  completions = new APICompletionsSub(this._client);
  jobSystem = new APIJobSystemSub(this._client);
  checkOutgoingIp = new APICheckOutgoingIpSub(this._client);
  endpointId = new APIEndpointIdSub(this._client);

}
