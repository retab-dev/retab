import { AbstractClient, CompositionClient } from '@/client';
import APIOrganizationsSub from "./organizations/client";
import APIPaymentsSub from "./payments/client";
import APITeamSub from "./team/client";
import APIDomainsSub from "./domains/client";

export default class APIIam extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  organizations = new APIOrganizationsSub(this._client);
  payments = new APIPaymentsSub(this._client);
  team = new APITeamSub(this._client);
  domains = new APIDomainsSub(this._client);

}
