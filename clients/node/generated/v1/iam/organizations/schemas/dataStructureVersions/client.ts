import { AbstractClient, CompositionClient, streamResponse } from '@/client';

export default class APIDataStructureVersions extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get({ isActive, isCurrent }: { isActive?: boolean | null, isCurrent?: boolean | null } = {}): Promise<string[]> {
    let res = await this._fetch({
      url: `/v1/iam/organizations/schemas/data_structure_versions`,
      method: "GET",
      params: { "is_active": isActive, "is_current": isCurrent },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
