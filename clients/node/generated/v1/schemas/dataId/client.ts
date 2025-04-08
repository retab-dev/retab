import { AbstractClient, CompositionClient } from '@/client';
import APISchemaId from "./schemaId/client";
import { StoredSchema } from "@/types";

export default class APIDataId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  schemaId = new APISchemaId(this);

  async get(dataId: string): Promise<StoredSchema> {
    return this._fetch({
      url: `/v1/schemas/${dataId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
