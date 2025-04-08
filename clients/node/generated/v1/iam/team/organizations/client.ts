import { AbstractClient, CompositionClient } from '@/client';
import { Organization } from "@/types";

export default class APIOrganizations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(): Promise<Organization[]> {
    return this._fetch({
      url: `/v1/iam/team/organizations`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
