import { AbstractClient, CompositionClient } from '@/client';
import APIJobExecutionId from "./jobExecutionId/client";
import { CreateJobExecutionRequest, JobExecution, PaginatedList } from "@/types";

export default class APIJobExecutions extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  jobExecutionId = new APIJobExecutionId(this);

  async post({ ...body }: CreateJobExecutionRequest): Promise<JobExecution> {
    return this._fetch({
      url: `/v1/job_system/job_executions/`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
  async get({ before, after, limit, order, jobId, status }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", jobId?: string | null, status?: string | null }): Promise<PaginatedList> {
    return this._fetch({
      url: `/v1/job_system/job_executions/`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "job_id": jobId, "status": status },
      headers: {  },
    });
  }
  
}
