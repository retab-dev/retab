import { AbstractClient, CompositionClient } from '@/client';
import APIDocuments from "./documents/client";
import APISchemas from "./schemas/client";
import APIIam from "./iam/client";
import APIModels from "./models/client";
import APIAutomations from "./automations/client";
import APIDb from "./db/client";
import APISecrets from "./secrets/client";
import APIIntegrations from "./integrations/client";
import APIAnalytics from "./analytics/client";
import APIBranding from "./branding/client";
import APIExtractionsLogs from "./extractionsLogs/client";
import APIEvents from "./events/client";
import APIUsage from "./usage/client";
import APIUsageschemaId from "./usageschemaId/client";
import APIBenchmarking from "./benchmarking/client";
import APIFinetunedModels from "./finetunedModels/client";
import APIExperiments from "./experiments/client";
import APIJobSystem from "./jobSystem/client";
import APICheckOutgoingIp from "./checkOutgoingIp/client";
import APIEndpointId from "./endpointId/client";

export default class APIV1 extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  documents = new APIDocuments(this);
  schemas = new APISchemas(this);
  iam = new APIIam(this);
  models = new APIModels(this);
  automations = new APIAutomations(this);
  db = new APIDb(this);
  secrets = new APISecrets(this);
  integrations = new APIIntegrations(this);
  analytics = new APIAnalytics(this);
  branding = new APIBranding(this);
  extractionsLogs = new APIExtractionsLogs(this);
  events = new APIEvents(this);
  usage = new APIUsage(this);
  usageschemaId = new APIUsageschemaId(this);
  benchmarking = new APIBenchmarking(this);
  finetunedModels = new APIFinetunedModels(this);
  experiments = new APIExperiments(this);
  jobSystem = new APIJobSystem(this);
  checkOutgoingIp = new APICheckOutgoingIp(this);
  endpointId = new APIEndpointId(this);

}
