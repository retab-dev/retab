import { AbstractClient, CompositionClient } from '@/client';
import APIJobExecutions from "./jobExecutions/client";
import APIWorkflowExecutions from "./workflowExecutions/client";

export default class APIJobSystem extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  jobExecutions = new APIJobExecutions(this);
  workflowExecutions = new APIWorkflowExecutions(this);

}
