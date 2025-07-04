import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIDomainIdSub from "./domainId/client";
import { ListDomainsResponse, AddDomainRequest, CustomDomain } from "@/types";

export default class APIDomains extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  domainId = new APIDomainIdSub(this._client);

  async get(): Promise<ListDomainsResponse> {
    let res = await this._fetch({
      url: `/v1/iam/domains`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async post({ ...body }: AddDomainRequest): Promise<CustomDomain> {
    let res = await this._fetch({
      url: `/v1/iam/domains/`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
