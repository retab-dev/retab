import { AbstractClient, CompositionClient } from '@/client';
import APIDefault from "./default/client";

export default class APIDomainId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  default = new APIDefault(this);

  async delete(domainId: string): Promise<any> {
    return this._fetch({
      url: `/v1/iam/domains/${domainId}`,
      method: "DELETE",
      params: {  },
      headers: {  },
    });
  }
  
}
