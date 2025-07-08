import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APILinkIdSub from "./linkId/client";
import APIParseSub from "./parse/client";
import APIVerifyPasswordSub from "./verifyPassword/client";
import { ZLinkInput, LinkInput, ZLinkOutput, LinkOutput, ZListLinks, ListLinks } from "@/types";

export default class APILinks extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  linkId = new APILinkIdSub(this._client);
  parse = new APIParseSub(this._client);
  verifyPassword = new APIVerifyPasswordSub(this._client);

  async post({ ...body }: LinkInput): Promise<LinkOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/links`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZLinkOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async get({ processorId, before, after, limit, order }: { processorId: string, before?: string | null, after?: string | null, limit?: number, order?: "asc" | "desc" }): Promise<ListLinks> {
    let res = await this._fetch({
      url: `/v1/processors/automations/links`,
      method: "GET",
      params: { "processor_id": processorId, "before": before, "after": after, "limit": limit, "order": order },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZListLinks.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
