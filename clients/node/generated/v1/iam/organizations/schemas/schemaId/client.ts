import { AbstractClient, CompositionClient } from '@/client';
import { OrganizationSchemaEntry, UpdateSchemaEntry, OrganizationSchemaEntry } from "@/types";

export default class APISchemaId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(schemaId: string): Promise<OrganizationSchemaEntry> {
    return this._fetch({
      url: `/v1/iam/organizations/schemas/${schemaId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async post(schemaId: string, { ...body }: UpdateSchemaEntry): Promise<OrganizationSchemaEntry> {
    return this._fetch({
      url: `/v1/iam/organizations/schemas/${schemaId}`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
