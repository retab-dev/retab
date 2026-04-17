import z from 'zod';
import { CompositionClient, RequestOptions } from '../../client.js';
import {
  ZPaginatedList,
  PaginatedList,
  ZMIMEData,
  MIMEDataInput,
  ZJSONSchema,
  JSONSchemaInput,
  ZExtractionV2,
  ExtractionV2,
} from '../../types.js';

export type ExtractionCreateParams = {
  json_schema: JSONSchemaInput;
  model: string;
  document: MIMEDataInput;
  image_resolution_dpi?: number;
  n_consensus?: number;
  metadata?: Record<string, string>;
  additional_messages?: Record<string, unknown>[];
  bust_cache?: boolean;
};

export default class APIExtractions extends CompositionClient {
  constructor(client: CompositionClient) {
    super(client);
  }

  /**
   * Create an extraction. Posts to `/extractions` (the v2 resource route).
   *
   * Mirrors the Python `retab.extractions.create(...)` call.
   *
   * @returns Stored `ExtractionV2` resource.
   */
  async create(params: ExtractionCreateParams, options?: RequestOptions): Promise<ExtractionV2> {
    const document = await ZMIMEData.parseAsync(params.document);
    const json_schema = await ZJSONSchema.parseAsync(params.json_schema);
    const body: Record<string, unknown> = {
      document,
      json_schema,
      model: params.model,
    };
    if (params.image_resolution_dpi !== undefined) {
      body['image_resolution_dpi'] = params.image_resolution_dpi;
    }
    if (params.n_consensus !== undefined) {
      body['n_consensus'] = params.n_consensus;
    }
    if (params.metadata !== undefined) {
      body['metadata'] = params.metadata;
    }
    if (params.additional_messages !== undefined) {
      body['additional_messages'] = params.additional_messages;
    }
    if (params.bust_cache) {
      body['bust_cache'] = true;
    }
    return this._fetchJson(ZExtractionV2, {
      url: '/extractions',
      method: 'POST',
      body: { ...body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * List extractions with pagination and filtering.
   */
  async list(
    {
      before,
      after,
      limit = 10,
      order = 'desc',
      origin_type,
      origin_id,
      from_date,
      to_date,
      metadata,
      filename,
    }: {
      before?: string;
      after?: string;
      limit?: number;
      order?: 'asc' | 'desc';
      origin_type?: string;
      origin_id?: string;
      from_date?: Date;
      to_date?: Date;
      metadata?: Record<string, string>;
      filename?: string;
    } = {},
    options?: RequestOptions
  ): Promise<PaginatedList> {
    const params: Record<string, any> = {
      before,
      after,
      limit,
      order,
      origin_type,
      origin_id,
      from_date: from_date?.toISOString(),
      to_date: to_date?.toISOString(),
      filename,
      // Note: metadata must be JSON-serialized as the backend expects a JSON string
      metadata: metadata ? JSON.stringify(metadata) : undefined,
    };

    const cleanParams = Object.fromEntries(
      Object.entries(params).filter(([_, v]) => v !== undefined)
    );

    return this._fetchJson(ZPaginatedList, {
      url: '/extractions',
      method: 'GET',
      params: { ...cleanParams, ...(options?.params || {}) },
      headers: options?.headers,
    });
  }

  /**
   * Get an extraction by ID.
   */
  async get(extraction_id: string, options?: RequestOptions): Promise<ExtractionV2> {
    return this._fetchJson(ZExtractionV2, {
      url: `/extractions/${extraction_id}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Get extraction result enriched with per-leaf source provenance.
   *
   * Each extracted leaf value is wrapped as {value, source} where source
   * contains citation content, surrounding context, and a format-specific anchor.
   *
   * @param extraction_id - ID of the extraction to source
   */
  async sources(extraction_id: string, options?: RequestOptions): Promise<Record<string, any>> {
    return this._fetchJson(z.record(z.any()), {
      url: `/extractions/${extraction_id}/sources`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Delete an extraction by ID.
   */
  async delete(extraction_id: string, options?: RequestOptions): Promise<void> {
    return this._fetchJson({
      url: `/extractions/${extraction_id}`,
      method: 'DELETE',
      params: options?.params,
      headers: options?.headers,
    });
  }
}
