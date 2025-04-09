import { AbstractClient, CompositionClient } from '@/client';
import APISchemaIdSub from "./schemaId/client";
import { StoredSchema } from "@/types";

export default class APIDataId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaId = new APISchemaIdSub(this._client);

  async get(dataId: string): Promise<StoredSchema> {
    return this._fetch({
      url: `/v1/schemas/${dataId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
