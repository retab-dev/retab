import { AbstractClient, CompositionClient } from '@/client';
import APITests from "./tests/client";
import APIVectorSearch from "./vectorSearch/client";
import APILogs from "./logs/client";
import APIOutlookPluginId from "./outlookPluginId/client";
import APIOpen from "./open/client";
import APIConvertToEmailDataAndUploadFile from "./convertToEmailDataAndUploadFile/client";
import APIUpdateEmailData from "./updateEmailData/client";
import APIConvertEmlBytesToEmailModel from "./convertEmlBytesToEmailModel/client";
import APIConvertMsgToEmailModel from "./convertMsgToEmailModel/client";
import APIManifest from "./manifest/client";
import { OutlookInput, OutlookOutput, PaginatedList } from "@/types";

export default class APIOutlook extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  tests = new APITests(this);
  vectorSearch = new APIVectorSearch(this);
  logs = new APILogs(this);
  outlookPluginId = new APIOutlookPluginId(this);
  open = new APIOpen(this);
  convertToEmailDataAndUploadFile = new APIConvertToEmailDataAndUploadFile(this);
  updateEmailData = new APIUpdateEmailData(this);
  convertEmlBytesToEmailModel = new APIConvertEmlBytesToEmailModel(this);
  convertMsgToEmailModel = new APIConvertMsgToEmailModel(this);
  manifest = new APIManifest(this);

  async post({ ...body }: OutlookInput): Promise<OutlookOutput> {
    return this._fetch({
      url: `/v1/automations/outlook`,
      method: "POST",
      params: {  },
      headers: {  },
      body: body,
      bodyMime: "application/json",
    });
  }
  
  async get({ before, after, limit, order, id, name, webhookUrl, schemaId, schemaDataId }: { before?: string | null, after?: string | null, limit?: number | null, order?: "asc" | "desc" | null, id?: string | null, name?: string | null, webhookUrl?: string | null, schemaId?: string | null, schemaDataId?: string | null }): Promise<PaginatedList> {
    return this._fetch({
      url: `/v1/automations/outlook`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "id": id, "name": name, "webhook_url": webhookUrl, "schema_id": schemaId, "schema_data_id": schemaDataId },
      headers: {  },
    });
  }
  
}
