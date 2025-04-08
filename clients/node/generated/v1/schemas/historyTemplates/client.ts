import { AbstractClient, CompositionClient } from '@/client';
import APITemplateId from "./templateId/client";

export default class APIHistoryTemplates extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  templateId = new APITemplateId(this);

}
