import { AbstractClient, CompositionClient } from '@/client';
import APIMemberId from "./memberId/client";
import { TeamMember } from "@/types";

export default class APIMembers extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  memberId = new APIMemberId(this);

  async get(): Promise<TeamMember[]> {
    return this._fetch({
      url: `/v1/iam/team/members`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
