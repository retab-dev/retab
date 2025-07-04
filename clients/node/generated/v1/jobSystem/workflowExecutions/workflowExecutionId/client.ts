import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { WorkflowExecution, UpdateWorkflowExecutionRequest, WorkflowExecution } from "@/types";

export default class APIWorkflowExecutionId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(workflowExecutionId: string): Promise<WorkflowExecution> {
    let res = await this._fetch({
      url: `/v1/job_system/workflow_executions/${workflowExecutionId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async put(workflowExecutionId: string, { ...body }: UpdateWorkflowExecutionRequest): Promise<WorkflowExecution> {
    let res = await this._fetch({
      url: `/v1/job_system/workflow_executions/${workflowExecutionId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
