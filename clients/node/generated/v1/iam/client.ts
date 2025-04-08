import { AbstractClient, CompositionClient } from '@/client';
import APIOrganizations from "./organizations/client";
import APIPayments from "./payments/client";
import APITeam from "./team/client";
import APIDomains from "./domains/client";

export default class APIIam extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  organizations = new APIOrganizations(this);
  payments = new APIPayments(this);
  team = new APITeam(this);
  domains = new APIDomains(this);

}
