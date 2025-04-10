import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIJobExecutionIdSub from "./jobExecutionId/client";
import { CreateJobExecutionRequest, JobExecution, PaginatedList } from "@/types";

export default class APIJobExecutions extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  jobExecutionId = new APIJobExecutionIdSub(this._client);

  async post({ ...body }: CreateJobExecutionRequest): Promise<JobExecution> {
    let res = await this._fetch({
      url: `/v1/job_system/job_executions/`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async get({ before, after, limit, order, jobId, status }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", jobId?: string | null, status?: string | null } = {}): Promise<PaginatedList> {
    let res = await this._fetch({
      url: `/v1/job_system/job_executions/`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "job_id": jobId, "status": status },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
