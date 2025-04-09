import { AbstractClient, CompositionClient } from '@/client';
import APITestsSub from "./tests/client";
import APIVectorSearchSub from "./vectorSearch/client";
import APILogsSub from "./logs/client";
import APIOutlookPluginIdSub from "./outlookPluginId/client";
import APIOpenSub from "./open/client";
import APIConvertToEmailDataAndUploadFileSub from "./convertToEmailDataAndUploadFile/client";
import APIUpdateEmailDataSub from "./updateEmailData/client";
import APIConvertEmlBytesToEmailModelSub from "./convertEmlBytesToEmailModel/client";
import APIConvertMsgToEmailModelSub from "./convertMsgToEmailModel/client";
import APIManifestSub from "./manifest/client";
import { OutlookInput, OutlookOutput, PaginatedList } from "@/types";

export default class APIOutlook extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  tests = new APITestsSub(this._client);
  vectorSearch = new APIVectorSearchSub(this._client);
  logs = new APILogsSub(this._client);
  outlookPluginId = new APIOutlookPluginIdSub(this._client);
  open = new APIOpenSub(this._client);
  convertToEmailDataAndUploadFile = new APIConvertToEmailDataAndUploadFileSub(this._client);
  updateEmailData = new APIUpdateEmailDataSub(this._client);
  convertEmlBytesToEmailModel = new APIConvertEmlBytesToEmailModelSub(this._client);
  convertMsgToEmailModel = new APIConvertMsgToEmailModelSub(this._client);
  manifest = new APIManifestSub(this._client);

  async post({ ...body }: OutlookInput): Promise<OutlookOutput> {
    return this._fetch({
      url: `/v1/automations/outlook`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
  async get({ before, after, limit, order, id, name, webhookUrl, schemaId, schemaDataId }: { before?: string | null, after?: string | null, limit?: number | null, order?: "asc" | "desc" | null, id?: string | null, name?: string | null, webhookUrl?: string | null, schemaId?: string | null, schemaDataId?: string | null } = {}): Promise<PaginatedList> {
    return this._fetch({
      url: `/v1/automations/outlook`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "id": id, "name": name, "webhook_url": webhookUrl, "schema_id": schemaId, "schema_data_id": schemaDataId },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
