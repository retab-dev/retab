import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIProjectIdSub from "./projectId/client";
import { ZPaginatedList, PaginatedList, ZProject, Project } from "@/types";

export default class APIProjects extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  projectId = new APIProjectIdSub(this._client);

  async get({ before, after, limit, order, name, fromDate, toDate }: { before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc", name?: string | null, fromDate?: string | null, toDate?: string | null } = {}): Promise<PaginatedList> {
    let res = await this._fetch({
      url: `/v1/evaluations/projects`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "name": name, "from_date": fromDate, "to_date": toDate },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZPaginatedList.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async post({ ...body }: Project): Promise<Project> {
    let res = await this._fetch({
      url: `/v1/evaluations/projects`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZProject.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
