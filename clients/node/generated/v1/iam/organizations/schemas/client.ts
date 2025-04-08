import { AbstractClient, CompositionClient } from '@/client';
import APIDataStructureVersions from "./dataStructureVersions/client";
import APISchemaId from "./schemaId/client";
import APIVersion from "./version/client";
import { OrganizationSchemaEntry, CreateSchemaEntry, OrganizationSchemaEntry } from "@/types";

export default class APISchemas extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  dataStructureVersions = new APIDataStructureVersions(this);
  schemaId = new APISchemaId(this);
  version = new APIVersion(this);

  async get({ dataStructureVersion, isCurrent, isActive }: { dataStructureVersion?: string | null, isCurrent?: boolean | null, isActive?: boolean | null }): Promise<OrganizationSchemaEntry[]> {
    return this._fetch({
      url: `/v1/iam/organizations/schemas/`,
      method: "GET",
      params: { "data_structure_version": dataStructureVersion, "is_current": isCurrent, "is_active": isActive },
      headers: {  },
    });
  }
  
  async post({ ...body }: CreateSchemaEntry): Promise<OrganizationSchemaEntry> {
    return this._fetch({
      url: `/v1/iam/organizations/schemas/`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
