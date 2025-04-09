import { AbstractClient, CompositionClient } from '@/client';
import { CreateAndLinkOrganizationRequest, CreateOrganizationResponse } from "@/types";

export default class APICreateAndLinkOrganization extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: CreateAndLinkOrganizationRequest): Promise<CreateOrganizationResponse> {
    return this._fetch({
      url: `/v1/iam/team/create_and_link_organization`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
