import { AbstractClient, CompositionClient } from '@/client';
import APITestsSub from "./tests/client";
import APILogsSub from "./logs/client";
import APILinkIdSub from "./linkId/client";
import APIVerifyPasswordSub from "./verifyPassword/client";
import APIParseSub from "./parse/client";
import APIOpenSub from "./open/client";
import { LinkInput, LinkOutput } from "@/types";

export default class APILinks extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  tests = new APITestsSub(this._client);
  logs = new APILogsSub(this._client);
  linkId = new APILinkIdSub(this._client);
  verifyPassword = new APIVerifyPasswordSub(this._client);
  parse = new APIParseSub(this._client);
  open = new APIOpenSub(this._client);

  async post({ ...body }: LinkInput): Promise<LinkOutput> {
    return this._fetch({
      url: `/v1/automations/links`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
