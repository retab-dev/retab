import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';
import { Extractions, AsyncExtractions, BaseExtractionsMixin } from './extractions.js';
import { FieldUnset } from '../../types/standards.js';

export class Documents extends SyncAPIResource {
  public extractions: Extractions;
  private mixin = new BaseExtractionsMixin();

  constructor(client: any) {
    super(client);
    this.extractions = new Extractions(client);
  }

  // Extract method directly on Documents class to match Python SDK
  extract(params: any): any {
    const preparedRequest = this.mixin.prepareExtraction(
      params.json_schema,
      params.document,
      undefined, // documents
      params.image_resolution_dpi || FieldUnset,
      params.browser_canvas || FieldUnset,
      params.model || FieldUnset,
      params.temperature || FieldUnset,
      params.modality || FieldUnset,
      params.reasoning_effort || FieldUnset,
      false, // stream
      params.n_consensus || FieldUnset,
      params.store || false,
      params.idempotency_key
    );

    return this._client._preparedRequest(preparedRequest);
  }

  create_messages(params: { document: string }): any {
    const mimeDocument = this.mixin.prepareMimeDocument(params.document);
    const preparedRequest = {
      method: 'POST' as const,
      url: '/v1/documents/create_messages',
      data: {
        document: mimeDocument,
        modality: 'native' // Default modality
      }
    };
    return this._client._preparedRequest(preparedRequest);
  }
}

export class AsyncDocuments extends AsyncAPIResource {
  public extractions: AsyncExtractions;
  private mixin = new BaseExtractionsMixin();

  constructor(client: any) {
    super(client);
    this.extractions = new AsyncExtractions(client);
  }

  // Async extract method
  async extract(params: any): Promise<any> {
    const preparedRequest = this.mixin.prepareExtraction(
      params.json_schema,
      params.document,
      undefined, // documents
      params.image_resolution_dpi || FieldUnset,
      params.browser_canvas || FieldUnset,
      params.model || FieldUnset,
      params.temperature || FieldUnset,
      params.modality || FieldUnset,
      params.reasoning_effort || FieldUnset,
      false, // stream
      params.n_consensus || FieldUnset,
      params.store || false,
      params.idempotency_key
    );

    return this._client._preparedRequest(preparedRequest);
  }

  async create_messages(params: { document: string }): Promise<any> {
    const mimeDocument = this.mixin.prepareMimeDocument(params.document);
    const preparedRequest = {
      method: 'POST' as const,
      url: '/v1/documents/create_messages',
      data: {
        document: mimeDocument,
        modality: 'native' // Default modality
      }
    };
    return this._client._preparedRequest(preparedRequest);
  }
}