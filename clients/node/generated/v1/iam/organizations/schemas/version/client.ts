import { AbstractClient, CompositionClient } from '@/client';

export default class APIVersion extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async delete({ schemaId }: { schemaId: string }): Promise<any> {
    return this._fetch({
      url: `/v1/iam/organizations/schemas/${version}`,
      method: "DELETE",
      params: { "schema_id": schemaId },
      headers: {  },
    });
  }
  
}
