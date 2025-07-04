import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { CreateAndLinkOrganizationRequest, CreateOrganizationResponse } from "@/types";

export default class APICreateAndLinkOrganization extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: CreateAndLinkOrganizationRequest): Promise<CreateOrganizationResponse> {
    let res = await this._fetch({
      url: `/v1/iam/team/create_and_link_organization`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
