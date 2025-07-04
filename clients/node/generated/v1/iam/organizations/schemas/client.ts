import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIDataStructureVersionsSub from "./dataStructureVersions/client";
import APISchemaIdSub from "./schemaId/client";
import APIVersionSub from "./version/client";
import { OrganizationSchemaEntry, CreateSchemaEntry, OrganizationSchemaEntry } from "@/types";

export default class APISchemas extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  dataStructureVersions = new APIDataStructureVersionsSub(this._client);
  schemaId = new APISchemaIdSub(this._client);
  version = new APIVersionSub(this._client);

  async get({ dataStructureVersion, isCurrent, isActive }: { dataStructureVersion?: string | null, isCurrent?: boolean | null, isActive?: boolean | null } = {}): Promise<OrganizationSchemaEntry[]> {
    let res = await this._fetch({
      url: `/v1/iam/organizations/schemas/`,
      method: "GET",
      params: { "data_structure_version": dataStructureVersion, "is_current": isCurrent, "is_active": isActive },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async post({ ...body }: CreateSchemaEntry): Promise<OrganizationSchemaEntry> {
    let res = await this._fetch({
      url: `/v1/iam/organizations/schemas/`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
