import { AbstractClient, CompositionClient } from '@/client';
import APIDomainIdSub from "./domainId/client";
import { ListDomainsResponse, AddDomainRequest, CustomDomain } from "@/types";

export default class APIDomains extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  domainId = new APIDomainIdSub(this._client);

  async get(): Promise<ListDomainsResponse> {
    return this._fetch({
      url: `/v1/iam/domains`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async post({ ...body }: AddDomainRequest): Promise<CustomDomain> {
    return this._fetch({
      url: `/v1/iam/domains/`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
