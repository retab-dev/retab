import { AbstractClient, CompositionClient } from '@/client';
import APIDomainId from "./domainId/client";
import { ListDomainsResponse, AddDomainRequest, CustomDomain } from "@/types";

export default class APIDomains extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  domainId = new APIDomainId(this);

  async get(): Promise<ListDomainsResponse> {
    return this._fetch({
      url: `/v1/iam/domains`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
  async post({ ...body }: AddDomainRequest): Promise<CustomDomain> {
    return this._fetch({
      url: `/v1/iam/domains/`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
