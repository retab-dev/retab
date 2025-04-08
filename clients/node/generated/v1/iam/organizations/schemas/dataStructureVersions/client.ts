import { AbstractClient, CompositionClient } from '@/client';

export default class APIDataStructureVersions extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get({ isActive, isCurrent }: { isActive?: boolean | null, isCurrent?: boolean | null }): Promise<string[]> {
    return this._fetch({
      url: `/v1/iam/organizations/schemas/data_structure_versions`,
      method: "GET",
      params: { "is_active": isActive, "is_current": isCurrent },
      headers: {  },
    });
  }
  
}
