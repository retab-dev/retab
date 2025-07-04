import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIMemberIdSub from "./memberId/client";
import { TeamMember } from "@/types";

export default class APIMembers extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  memberId = new APIMemberIdSub(this._client);

  async get(): Promise<TeamMember[]> {
    let res = await this._fetch({
      url: `/v1/iam/team/members`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
