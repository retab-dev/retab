import { AbstractClient, CompositionClient } from '@/client';
import APISampleDocument from "./sampleDocument/client";

export default class APITemplateId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  sampleDocument = new APISampleDocument(this);

}
