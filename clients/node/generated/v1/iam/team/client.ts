import { AbstractClient, CompositionClient } from '@/client';
import APICreateAndLinkOrganization from "./createAndLinkOrganization/client";
import APIMembers from "./members/client";
import APIOrganizations from "./organizations/client";
import APIInvitations from "./invitations/client";

export default class APITeam extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  createAndLinkOrganization = new APICreateAndLinkOrganization(this);
  members = new APIMembers(this);
  organizations = new APIOrganizations(this);
  invitations = new APIInvitations(this);

}
