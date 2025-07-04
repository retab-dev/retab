import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APICreateAndLinkOrganizationSub from "./createAndLinkOrganization/client";
import APIMembersSub from "./members/client";
import APIOrganizationsSub from "./organizations/client";
import APIInvitationsSub from "./invitations/client";

export default class APITeam extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  createAndLinkOrganization = new APICreateAndLinkOrganizationSub(this._client);
  members = new APIMembersSub(this._client);
  organizations = new APIOrganizationsSub(this._client);
  invitations = new APIInvitationsSub(this._client);

}
