import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APITestsSub from "./tests/client";
import APIVectorSearchSub from "./vectorSearch/client";
import APIOutlookPluginIdSub from "./outlookPluginId/client";
import APIConvertToEmailDataAndUploadFileSub from "./convertToEmailDataAndUploadFile/client";
import APIUpdateEmailDataSub from "./updateEmailData/client";
import APIConvertEmlBytesToEmailModelSub from "./convertEmlBytesToEmailModel/client";
import APIConvertMsgToEmailModelSub from "./convertMsgToEmailModel/client";
import APIManifestSub from "./manifest/client";
import { ZOutlookInput, OutlookInput, ZOutlookOutput, OutlookOutput, ZPaginatedList, PaginatedList } from "@/types";

export default class APIOutlook extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  tests = new APITestsSub(this._client);
  vectorSearch = new APIVectorSearchSub(this._client);
  outlookPluginId = new APIOutlookPluginIdSub(this._client);
  convertToEmailDataAndUploadFile = new APIConvertToEmailDataAndUploadFileSub(this._client);
  updateEmailData = new APIUpdateEmailDataSub(this._client);
  convertEmlBytesToEmailModel = new APIConvertEmlBytesToEmailModelSub(this._client);
  convertMsgToEmailModel = new APIConvertMsgToEmailModelSub(this._client);
  manifest = new APIManifestSub(this._client);

  async post({ ...body }: OutlookInput): Promise<OutlookOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/outlook`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZOutlookOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async get({ before, after, limit, order, id, name, webhookUrl, processorId }: { before?: string | null, after?: string | null, limit?: number | null, order?: "asc" | "desc" | null, id?: string | null, name?: string | null, webhookUrl?: string | null, processorId?: string | null } = {}): Promise<PaginatedList> {
    let res = await this._fetch({
      url: `/v1/processors/automations/outlook`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "id": id, "name": name, "webhook_url": webhookUrl, "processor_id": processorId },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZPaginatedList.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
