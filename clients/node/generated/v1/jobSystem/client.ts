import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIJobExecutionsSub from "./jobExecutions/client";
import APIWorkflowExecutionsSub from "./workflowExecutions/client";

export default class APIJobSystem extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  jobExecutions = new APIJobExecutionsSub(this._client);
  workflowExecutions = new APIWorkflowExecutionsSub(this._client);

}
