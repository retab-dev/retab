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
      params: {  },
      headers: {  },
    });
  }
  
  async put(workflowExecutionId: string, { ...body }: UpdateWorkflowExecutionRequest): Promise<WorkflowExecution> {
    return this._fetch({
      url: `/v1/job_system/workflow_executions/${workflowExecutionId}`,
      method: "PUT",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
