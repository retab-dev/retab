import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { OrganizationSchemaEntry, UpdateSchemaEntry, OrganizationSchemaEntry } from "@/types";

export default class APISchemaId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(schemaId: string): Promise<OrganizationSchemaEntry> {
    let res = await this._fetch({
      url: `/v1/iam/organizations/schemas/${schemaId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async post(schemaId: string, { ...body }: UpdateSchemaEntry): Promise<OrganizationSchemaEntry> {
    let res = await this._fetch({
      url: `/v1/iam/organizations/schemas/${schemaId}`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
