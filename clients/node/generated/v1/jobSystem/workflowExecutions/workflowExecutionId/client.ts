import { AbstractClient, CompositionClient } from '@/client';
import { WorkflowExecution, UpdateWorkflowExecutionRequest, WorkflowExecution } from "@/types";

export default class APIWorkflowExecutionId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(workflowExecutionId: string): Promise<WorkflowExecution> {
    return this._fetch({
      url: `/v1/job_system/workflow_executions/${workflowExecutionId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async put(workflowExecutionId: string, { ...body }: UpdateWorkflowExecutionRequest): Promise<WorkflowExecution> {
    return this._fetch({
      url: `/v1/job_system/workflow_executions/${workflowExecutionId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
