import { AbstractClient, CompositionClient } from '@/client';
import APISampleDocumentSub from "./sampleDocument/client";

export default class APITemplateId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  sampleDocument = new APISampleDocumentSub(this._client);

}
