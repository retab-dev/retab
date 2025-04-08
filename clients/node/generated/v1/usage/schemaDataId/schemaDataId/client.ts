import { AbstractClient, CompositionClient } from '@/client';
import { Amount } from "@/types";

export default class APISchemaDataId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(schemaDataId: string, { startDate, endDate }: { startDate?: Date | null, endDate?: Date | null }): Promise<Amount> {
    return this._fetch({
      url: `/v1/usage/schema_data_id/${schemaDataId}`,
      method: "GET",
      params: { "start_date": startDate, "end_date": endDate },
      headers: {  },
    });
  }
  
}
