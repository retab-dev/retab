import { AbstractClient, CompositionClient } from '@/client';
import { StoredSchema } from "@/types";

export default class APISchemaId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(schemaId: string): Promise<StoredSchema> {
    return this._fetch({
      url: `/v1/schemas/schema/${schemaId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
