import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIWorkflowExecutionIdSub from "./workflowExecutionId/client";
import { CreateWorkflowExecutionRequest, WorkflowExecution, PaginatedList } from "@/types";

export default class APIWorkflowExecutions extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  workflowExecutionId = new APIWorkflowExecutionIdSub(this._client);

  async post({ ...body }: CreateWorkflowExecutionRequest): Promise<WorkflowExecution> {
    let res = await this._fetch({
      url: `/v1/job_system/workflow_executions/`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async get({ before, after, limit, order, workflowId, status }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", workflowId?: string | null, status?: string | null } = {}): Promise<PaginatedList> {
    let res = await this._fetch({
      url: `/v1/job_system/workflow_executions/`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "workflow_id": workflowId, "status": status },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
