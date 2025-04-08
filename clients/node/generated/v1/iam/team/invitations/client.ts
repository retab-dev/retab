import { AbstractClient, CompositionClient } from '@/client';
import APIInvitationId from "./invitationId/client";
import { TeamInvitation, InvitationRequest, TeamInvitation } from "@/types";

export default class APIInvitations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  invitationId = new APIInvitationId(this);

  async get(): Promise<TeamInvitation[]> {
    return this._fetch({
      url: `/v1/iam/team/invitations`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
  async post({ ...body }: InvitationRequest): Promise<TeamInvitation> {
    return this._fetch({
      url: `/v1/iam/team/invitations`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
