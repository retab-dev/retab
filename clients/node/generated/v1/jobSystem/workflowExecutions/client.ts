import { AbstractClient, CompositionClient } from '@/client';
import APIWorkflowExecutionId from "./workflowExecutionId/client";
import { CreateWorkflowExecutionRequest, WorkflowExecution, PaginatedList } from "@/types";

export default class APIWorkflowExecutions extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  workflowExecutionId = new APIWorkflowExecutionId(this);

  async post({ ...body }: CreateWorkflowExecutionRequest): Promise<WorkflowExecution> {
    return this._fetch({
      url: `/v1/job_system/workflow_executions/`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
  async get({ before, after, limit, order, workflowId, status }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", workflowId?: string | null, status?: string | null }): Promise<PaginatedList> {
    return this._fetch({
      url: `/v1/job_system/workflow_executions/`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "workflow_id": workflowId, "status": status },
      headers: {  },
    });
  }
  
}
