import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';
import { z } from 'zod';
import { prepareMimeDocument } from '../../utils/mime.js';
import { loadJsonSchema } from '../../utils/json_schema.js';

// Processor configuration schema
const ProcessorConfigSchema = z.object({
  id: z.string(),
  name: z.string(),
  json_schema: z.record(z.any()),
  modality: z.string(),
  model: z.string(),
  temperature: z.number().optional(),
  reasoning_effort: z.string().optional(),
  image_resolution_dpi: z.number().optional(),
  browser_canvas: z.string().optional(),
  n_consensus: z.number().optional(),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});

const ListMetadataSchema = z.object({
  has_more: z.boolean().optional(),
  total_count: z.number().optional(),
  next_page_token: z.string().optional(),
  previous_page_token: z.string().optional(),
}).optional();

const ListProcessorsSchema = z.object({
  data: z.array(ProcessorConfigSchema),
  list_metadata: ListMetadataSchema.optional(),
});

export type ProcessorConfig = z.infer<typeof ProcessorConfigSchema>;
export type ListProcessors = z.infer<typeof ListProcessorsSchema>;

export class ProcessorsMixin {
  prepareCreate(params: {
    name: string;
    json_schema: Record<string, any> | string;
    modality?: string;
    model?: string;
    temperature?: number;
    reasoning_effort?: string;
    image_resolution_dpi?: number;
    browser_canvas?: string;
    n_consensus?: number;
  }): any {
    const {
      name,
      json_schema,
      modality = 'native',
      model = 'gpt-4o-mini',
      temperature,
      reasoning_effort,
      image_resolution_dpi,
      browser_canvas,
      n_consensus,
    } = params;

    // Load the JSON schema from file path, string, or dict
    const loadedSchema = loadJsonSchema(json_schema);

    const configDict: Record<string, any> = {
      name,
      json_schema: loadedSchema,
      modality,
      model,
    };

    if (temperature !== undefined) configDict.temperature = temperature;
    if (reasoning_effort !== undefined) configDict.reasoning_effort = reasoning_effort;
    if (image_resolution_dpi !== undefined) configDict.image_resolution_dpi = image_resolution_dpi;
    if (browser_canvas !== undefined) configDict.browser_canvas = browser_canvas;
    if (n_consensus !== undefined) configDict.n_consensus = n_consensus;

    return {
      method: 'POST' as const,
      url: '/v1/processors',
      data: configDict,
    };
  }

  prepareList(params: {
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
    name?: string;
    modality?: string;
    model?: string;
    schema_id?: string;
    schema_data_id?: string;
  } = {}): any {
    const {
      before,
      after,
      limit = 10,
      order = 'desc',
      name,
      modality,
      model,
      schema_id,
      schema_data_id,
    } = params;

    const queryParams: Record<string, any> = {};
    if (before !== undefined) queryParams.before = before;
    if (after !== undefined) queryParams.after = after;
    if (limit !== undefined) queryParams.limit = limit;
    if (order !== undefined) queryParams.order = order;
    if (name !== undefined) queryParams.name = name;
    if (modality !== undefined) queryParams.modality = modality;
    if (model !== undefined) queryParams.model = model;
    if (schema_id !== undefined) queryParams.schema_id = schema_id;
    if (schema_data_id !== undefined) queryParams.schema_data_id = schema_data_id;

    return {
      method: 'GET' as const,
      url: '/v1/processors',
      params: queryParams,
    };
  }

  prepareGet(processor_id: string): any {
    return {
      method: 'GET' as const,
      url: `/v1/processors/${processor_id}`,
    };
  }

  prepareUpdate(params: {
    processor_id: string;
    name?: string;
    modality?: string;
    image_resolution_dpi?: number;
    browser_canvas?: string;
    model?: string;
    json_schema?: Record<string, any> | string;
    temperature?: number;
    reasoning_effort?: string;
    n_consensus?: number;
  }): any {
    const {
      processor_id,
      name,
      modality,
      image_resolution_dpi,
      browser_canvas,
      model,
      json_schema,
      temperature,
      reasoning_effort,
      n_consensus,
    } = params;

    let loadedSchema = undefined;
    if (json_schema !== undefined) {
      loadedSchema = loadJsonSchema(json_schema);
    }

    const updateData: Record<string, any> = {};
    if (name !== undefined) updateData.name = name;
    if (modality !== undefined) updateData.modality = modality;
    if (image_resolution_dpi !== undefined) updateData.image_resolution_dpi = image_resolution_dpi;
    if (browser_canvas !== undefined) updateData.browser_canvas = browser_canvas;
    if (model !== undefined) updateData.model = model;
    if (loadedSchema !== undefined) updateData.json_schema = loadedSchema;
    if (temperature !== undefined) updateData.temperature = temperature;
    if (reasoning_effort !== undefined) updateData.reasoning_effort = reasoning_effort;
    if (n_consensus !== undefined) updateData.n_consensus = n_consensus;

    return {
      method: 'PUT' as const,
      url: `/v1/processors/${processor_id}`,
      data: updateData,
    };
  }

  prepareDelete(processor_id: string): any {
    return {
      method: 'DELETE' as const,
      url: `/v1/processors/${processor_id}`,
    };
  }

  prepareSubmit(params: {
    processor_id: string;
    document?: string;
    documents?: string[];
    temperature?: number;
    seed?: number;
    store?: boolean;
  }): any {
    const {
      processor_id,
      document,
      documents,
      temperature,
      seed,
      store = true,
    } = params;

    if (!document && !documents) {
      throw new Error("Either 'document' or 'documents' must be provided");
    }

    if (document && documents) {
      throw new Error("Provide either 'document' (single) or 'documents' (multiple), not both");
    }

    const formData: Record<string, any> = {};
    if (temperature !== undefined) formData.temperature = temperature;
    if (seed !== undefined) formData.seed = seed;
    if (store !== undefined) formData.store = store;

    const files: Record<string, any> = {};
    if (document) {
      const mimeDocument = prepareMimeDocument(document);
      files.document = {
        filename: mimeDocument.filename,
        content: Buffer.from(mimeDocument.content, 'base64'),
        mime_type: mimeDocument.mime_type,
      };
    } else if (documents) {
      files.documents = documents.map((doc) => {
        const mimeDoc = prepareMimeDocument(doc);
        return {
          filename: mimeDoc.filename,
          content: Buffer.from(mimeDoc.content, 'base64'),
          mime_type: mimeDoc.mime_type,
        };
      });
    }

    return {
      method: 'POST' as const,
      url: `/v1/processors/${processor_id}/submit`,
      form_data: formData,
      files,
    };
  }
}

export class Processors extends SyncAPIResource {
  mixin = new ProcessorsMixin();

  create(params: {
    name: string;
    json_schema: Record<string, any> | string;
    modality?: string;
    model?: string;
    temperature?: number;
    reasoning_effort?: string;
    image_resolution_dpi?: number;
    browser_canvas?: string;
    n_consensus?: number;
  }): ProcessorConfig {
    const preparedRequest = this.mixin.prepareCreate(params);
    const response = this._client._preparedRequest(preparedRequest);
    const parsedResponse = ProcessorConfigSchema.parse(response);
    console.log(`Processor ID: ${parsedResponse.id}. Processor available at https://www.retab.com/dashboard/processors/${parsedResponse.id}`);
    return parsedResponse;
  }

  list(params: {
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
    name?: string;
    modality?: string;
    model?: string;
    schema_id?: string;
    schema_data_id?: string;
  } = {}): ListProcessors {
    const preparedRequest = this.mixin.prepareList(params);
    const response = this._client._preparedRequest(preparedRequest);
    return ListProcessorsSchema.parse(response);
  }

  get(processor_id: string): ProcessorConfig {
    const preparedRequest = this.mixin.prepareGet(processor_id);
    const response = this._client._preparedRequest(preparedRequest);
    return ProcessorConfigSchema.parse(response);
  }

  update(params: {
    processor_id: string;
    name?: string;
    modality?: string;
    image_resolution_dpi?: number;
    browser_canvas?: string;
    model?: string;
    json_schema?: Record<string, any> | string;
    temperature?: number;
    reasoning_effort?: string;
    n_consensus?: number;
  }): ProcessorConfig {
    const preparedRequest = this.mixin.prepareUpdate(params);
    const response = this._client._preparedRequest(preparedRequest);
    return ProcessorConfigSchema.parse(response);
  }

  delete(processor_id: string): void {
    const preparedRequest = this.mixin.prepareDelete(processor_id);
    this._client._preparedRequest(preparedRequest);
    console.log(`Processor Deleted. ID: ${processor_id}`);
  }

  submit(params: {
    processor_id: string;
    document?: string;
    documents?: string[];
    temperature?: number;
    seed?: number;
    store?: boolean;
  }): any {
    const preparedRequest = this.mixin.prepareSubmit(params);
    const response = this._client._preparedRequest(preparedRequest);
    return response;
  }
}

export class AsyncProcessors extends AsyncAPIResource {
  mixin = new ProcessorsMixin();

  async create(params: {
    name: string;
    json_schema: Record<string, any> | string;
    modality?: string;
    model?: string;
    temperature?: number;
    reasoning_effort?: string;
    image_resolution_dpi?: number;
    browser_canvas?: string;
    n_consensus?: number;
  }): Promise<ProcessorConfig> {
    const preparedRequest = this.mixin.prepareCreate(params);
    const response = await this._client._preparedRequest(preparedRequest);
    const parsedResponse = ProcessorConfigSchema.parse(response);
    console.log(`Processor ID: ${parsedResponse.id}. Processor available at https://www.retab.com/dashboard/processors/${parsedResponse.id}`);
    return parsedResponse;
  }

  async list(params: {
    before?: string;
    after?: string;
    limit?: number;
    order?: 'asc' | 'desc';
    name?: string;
    modality?: string;
    model?: string;
    schema_id?: string;
    schema_data_id?: string;
  } = {}): Promise<ListProcessors> {
    const preparedRequest = this.mixin.prepareList(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return ListProcessorsSchema.parse(response);
  }

  async get(processor_id: string): Promise<ProcessorConfig> {
    const preparedRequest = this.mixin.prepareGet(processor_id);
    const response = await this._client._preparedRequest(preparedRequest);
    return ProcessorConfigSchema.parse(response);
  }

  async update(params: {
    processor_id: string;
    name?: string;
    modality?: string;
    image_resolution_dpi?: number;
    browser_canvas?: string;
    model?: string;
    json_schema?: Record<string, any> | string;
    temperature?: number;
    reasoning_effort?: string;
    n_consensus?: number;
  }): Promise<ProcessorConfig> {
    const preparedRequest = this.mixin.prepareUpdate(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return ProcessorConfigSchema.parse(response);
  }

  async delete(processor_id: string): Promise<void> {
    const preparedRequest = this.mixin.prepareDelete(processor_id);
    await this._client._preparedRequest(preparedRequest);
    console.log(`Processor Deleted. ID: ${processor_id}`);
  }

  async submit(params: {
    processor_id: string;
    document?: string;
    documents?: string[];
    temperature?: number;
    seed?: number;
    store?: boolean;
  }): Promise<any> {
    const preparedRequest = this.mixin.prepareSubmit(params);
    const response = await this._client._preparedRequest(preparedRequest);
    return response;
  }
}