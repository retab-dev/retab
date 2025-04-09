import { AbstractClient, CompositionClient } from '@/client';
import { JobExecution, UpdateJobExecutionRequest, JobExecution } from "@/types";

export default class APIJobExecutionId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(jobExecutionId: string): Promise<JobExecution> {
    return this._fetch({
      url: `/v1/job_system/job_executions/${jobExecutionId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async put(jobExecutionId: string, { ...body }: UpdateJobExecutionRequest): Promise<JobExecution> {
    return this._fetch({
      url: `/v1/job_system/job_executions/${jobExecutionId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
