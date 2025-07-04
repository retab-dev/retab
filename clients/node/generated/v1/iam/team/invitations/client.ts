import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIInvitationIdSub from "./invitationId/client";
import { TeamInvitation, InvitationRequest, TeamInvitation } from "@/types";

export default class APIInvitations extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  invitationId = new APIInvitationIdSub(this._client);

  async get(): Promise<TeamInvitation[]> {
    let res = await this._fetch({
      url: `/v1/iam/team/invitations`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
  async post({ ...body }: InvitationRequest): Promise<TeamInvitation> {
    let res = await this._fetch({
      url: `/v1/iam/team/invitations`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
