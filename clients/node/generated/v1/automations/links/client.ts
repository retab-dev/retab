import { AbstractClient, CompositionClient } from '@/client';
import APITests from "./tests/client";
import APILogs from "./logs/client";
import APILinkId from "./linkId/client";
import APIVerifyPassword from "./verifyPassword/client";
import APIParse from "./parse/client";
import APIOpen from "./open/client";
import { LinkInput, LinkOutput } from "@/types";

export default class APILinks extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  tests = new APITests(this);
  logs = new APILogs(this);
  linkId = new APILinkId(this);
  verifyPassword = new APIVerifyPassword(this);
  parse = new APIParse(this);
  open = new APIOpen(this);

  async post({ ...body }: LinkInput): Promise<LinkOutput> {
    return this._fetch({
      url: `/v1/automations/links`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
