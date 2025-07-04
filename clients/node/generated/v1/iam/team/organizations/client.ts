import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { Organization } from "@/types";

export default class APIOrganizations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<Organization[]> {
    let res = await this._fetch({
      url: `/v1/iam/team/organizations`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
