import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIVersion extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async delete({ schemaId }: { schemaId: string }): Promise<any> {
    let res = await this._fetch({
      url: `/v1/iam/organizations/schemas/${version}`,
      method: "DELETE",
      params: { "schema_id": schemaId },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
