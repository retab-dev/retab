import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APITemplateIdSub from "./templateId/client";

export default class APIHistoryTemplates extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  templateId = new APITemplateIdSub(this._client);

}
