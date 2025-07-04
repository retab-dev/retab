import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { Project } from "@/types";

export default class APIProjectId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(projectId: string): Promise<Project> {
    let res = await this._fetch({
      url: `/v1/evaluations/projects/${projectId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
  async put(projectId: string, { ...body }: Project): Promise<Project> {
    let res = await this._fetch({
      url: `/v1/evaluations/projects/${projectId}`,
      method: "PUT",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
  async delete(projectId: string): Promise<any> {
    let res = await this._fetch({
      url: `/v1/evaluations/projects/${projectId}`,
      method: "DELETE",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
