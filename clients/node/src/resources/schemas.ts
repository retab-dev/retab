import fs from 'fs';
import path from 'path';
import { SyncAPIResource, AsyncAPIResource } from '../resource.js';
import { PreparedRequest } from '../types/standards.js';
import { Schema } from '../types/schemas/object.js';
import { loadJsonSchema } from '../utils/json_schema_utils.js';

// TypeScript equivalents of Python types
export type MIMEData = string | Buffer | { filename: string; content: Buffer; mimeType: string };
export type Modality = 'native' | 'text';
export type BrowserCanvas = 'A3' | 'A4' | 'A5';
export type ChatCompletionReasoningEffort = 'low' | 'medium' | 'high';

export interface GenerateSchemaRequest {
  documents: MIMEData[];
  instructions?: string | null;
  model: string;
  temperature: number;
  modality: Modality;
  reasoning_effort: ChatCompletionReasoningEffort;
}

export interface EvaluateSchemaRequest {
  documents: MIMEData[];
  json_schema: Record<string, any>;
  ground_truths?: Array<Record<string, any>> | null;
  model: string;
  reasoning_effort: ChatCompletionReasoningEffort;
  modality: Modality;
  image_resolution_dpi: number;
  browser_canvas: BrowserCanvas;
  n_consensus: number;
}

export interface EvaluateSchemaResponse {
  [key: string]: any; // Evaluation metrics
}

export interface EnhanceSchemaConfig {
  [key: string]: any;
}

export interface EnhanceSchemaRequest {
  json_schema: Record<string, any>;
  documents: MIMEData[];
  ground_truths?: Array<Record<string, any>> | null;
  instructions?: string | null;
  model: string;
  temperature: number;
  modality: Modality;
  flat_likelihoods?: Array<Record<string, number>> | Record<string, number> | null;
  tools_config: EnhanceSchemaConfig;
}

// Helper functions (simplified versions of Python utilities)
function assertValidModelSchemaGeneration(model: string): void {
  // Validate model for schema generation
  const validModels = ['gpt-4o-2024-11-20', 'gpt-4o-mini', 'gpt-4o'];
  if (!validModels.includes(model)) {
    throw new Error(`Model ${model} not valid for schema generation`);
  }
}

function prepareMimeDocumentList(documents: Array<string | Buffer | MIMEData>): MIMEData[] {
  return documents.map(doc => {
    if (typeof doc === 'string') {
      // If it's a file path, read the file
      if (fs.existsSync(doc)) {
        return {
          filename: path.basename(doc),
          content: fs.readFileSync(doc),
          mimeType: getMimeType(doc),
        };
      }
      // Otherwise treat as text content
      return doc;
    }
    return doc as MIMEData;
  });
}

function getMimeType(filePath: string): string {
  const ext = path.extname(filePath).toLowerCase();
  const mimeTypes: Record<string, string> = {
    '.pdf': 'application/pdf',
    '.jpg': 'image/jpeg',
    '.jpeg': 'image/jpeg',
    '.png': 'image/png',
    '.txt': 'text/plain',
    '.json': 'application/json',
  };
  return mimeTypes[ext] || 'application/octet-stream';
}


class SchemasMixin {
  public prepareGenerate(
    documents: Array<string | Buffer | MIMEData>,
    instructions?: string | null,
    model: string = 'gpt-4o-2024-11-20',
    temperature: number = 0,
    modality: Modality = 'native',
    reasoning_effort: ChatCompletionReasoningEffort = 'medium'
  ): PreparedRequest {
    assertValidModelSchemaGeneration(model);
    const mimeDocuments = prepareMimeDocumentList(documents);
    const request: GenerateSchemaRequest = {
      documents: mimeDocuments,
      instructions: instructions || undefined,
      model,
      temperature,
      modality,
      reasoning_effort,
    };
    return {
      method: 'POST',
      url: '/v1/schemas/generate',
      data: request,
    };
  }

  public prepareEvaluate(
    documents: Array<string | Buffer | MIMEData>,
    jsonSchema: Record<string, any>,
    groundTruths?: Array<Record<string, any>> | null,
    model: string = 'gpt-4o-mini',
    reasoning_effort: ChatCompletionReasoningEffort = 'medium',
    modality: Modality = 'native',
    imageResolutionDpi: number = 96,
    browserCanvas: BrowserCanvas = 'A4',
    nConsensus: number = 1
  ): PreparedRequest {
    const mimeDocuments = prepareMimeDocumentList(documents);
    const request: EvaluateSchemaRequest = {
      documents: mimeDocuments,
      json_schema: jsonSchema,
      ground_truths: groundTruths || undefined,
      model,
      reasoning_effort,
      modality,
      image_resolution_dpi: imageResolutionDpi,
      browser_canvas: browserCanvas,
      n_consensus: nConsensus,
    };
    return {
      method: 'POST',
      url: '/v1/schemas/evaluate',
      data: request,
    };
  }

  public prepareEnhance(
    jsonSchema: Record<string, any> | string,
    documents: Array<string | Buffer | MIMEData>,
    groundTruths?: Array<Record<string, any>> | null,
    instructions?: string | null,
    model: string = 'gpt-4o-mini',
    temperature: number = 0.0,
    modality: Modality = 'native',
    flatLikelihoods?: Array<Record<string, number>> | Record<string, number> | null,
    toolsConfig: EnhanceSchemaConfig = {}
  ): PreparedRequest {
    assertValidModelSchemaGeneration(model);
    const mimeDocuments = prepareMimeDocumentList(documents);
    const loadedJsonSchema = loadJsonSchema(jsonSchema);
    const request: EnhanceSchemaRequest = {
      json_schema: loadedJsonSchema,
      documents: mimeDocuments,
      ground_truths: groundTruths || undefined,
      instructions: instructions || undefined,
      model,
      temperature,
      modality,
      flat_likelihoods: flatLikelihoods || undefined,
      tools_config: toolsConfig,
    };
    return {
      method: 'POST',
      url: '/v1/schemas/enhance',
      data: request,
    };
  }

  public prepareGet(schemaId: string): PreparedRequest {
    return {
      method: 'GET',
      url: `/v1/schemas/${schemaId}`,
    };
  }
}

export class Schemas extends SyncAPIResource {
  private mixin = new SchemasMixin();

  load(jsonSchema?: Record<string, any> | string | null, pydanticModel?: any | null): Schema {
    if (jsonSchema) {
      return new Schema({ json_schema: loadJsonSchema(jsonSchema) });
    } else if (pydanticModel) {
      return new Schema({ pydanticModel });
    } else {
      throw new Error('Either json_schema or pydantic_model must be provided');
    }
  }

  async generate(
    documents: Array<string | Buffer | MIMEData>,
    instructions?: string | null,
    model: string = 'gpt-4o-2024-11-20',
    temperature: number = 0,
    modality: Modality = 'native'
  ): Promise<Schema> {
    const preparedRequest = this.mixin.prepareGenerate(documents, instructions, model, temperature, modality);
    const response = await this._client._preparedRequest(preparedRequest);
    return Schema.validate(response);
  }

  async evaluate(
    documents: Array<string | Buffer | MIMEData>,
    jsonSchema: Record<string, any>,
    groundTruths?: Array<Record<string, any>> | null,
    model: string = 'gpt-4o-mini',
    reasoning_effort: ChatCompletionReasoningEffort = 'medium',
    modality: Modality = 'native',
    imageResolutionDpi: number = 96,
    browserCanvas: BrowserCanvas = 'A4',
    nConsensus: number = 1
  ): Promise<EvaluateSchemaResponse> {
    const preparedRequest = this.mixin.prepareEvaluate(
      documents,
      jsonSchema,
      groundTruths,
      model,
      reasoning_effort,
      modality,
      imageResolutionDpi,
      browserCanvas,
      nConsensus
    );
    const response = await this._client._preparedRequest(preparedRequest);
    return response;
  }

  async enhance(
    jsonSchema: Record<string, any> | string,
    documents: Array<string | Buffer | MIMEData>,
    groundTruths?: Array<Record<string, any>> | null,
    instructions?: string | null,
    model: string = 'gpt-4o-2024-11-20',
    temperature: number = 0,
    modality: Modality = 'native',
    flatLikelihoods?: Array<Record<string, number>> | Record<string, number> | null,
    toolsConfig?: EnhanceSchemaConfig
  ): Promise<Schema> {
    const _toolsConfig = toolsConfig || {};
    const preparedRequest = this.mixin.prepareEnhance(
      jsonSchema,
      documents,
      groundTruths,
      instructions,
      model,
      temperature,
      modality,
      flatLikelihoods,
      _toolsConfig
    );
    const response = await this._client._preparedRequest(preparedRequest);
    return new Schema({ json_schema: response.json_schema });
  }
}

export class AsyncSchemas extends AsyncAPIResource {
  private mixin = new SchemasMixin();

  load(jsonSchema?: Record<string, any> | string | null, pydanticModel?: any | null): Schema {
    if (jsonSchema) {
      return new Schema({ json_schema: loadJsonSchema(jsonSchema) });
    } else if (pydanticModel) {
      return new Schema({ pydanticModel });
    } else {
      throw new Error('Either json_schema or pydantic_model must be provided');
    }
  }

  async get(schemaId: string): Promise<Schema> {
    const preparedRequest = this.mixin.prepareGet(schemaId);
    const response = await this._client._preparedRequest(preparedRequest);
    return Schema.validate(response);
  }

  async generate(
    documents: Array<string | Buffer | MIMEData>,
    instructions?: string | null,
    model: string = 'gpt-4o-2024-11-20',
    temperature: number = 0.0,
    modality: Modality = 'native'
  ): Promise<Schema> {
    const preparedRequest = this.mixin.prepareGenerate(
      documents,
      instructions,
      model,
      temperature,
      modality
    );
    const response = await this._client._preparedRequest(preparedRequest);
    return Schema.validate(response);
  }

  async evaluate(
    documents: Array<string | Buffer | MIMEData>,
    jsonSchema: Record<string, any>,
    groundTruths?: Array<Record<string, any>> | null,
    model: string = 'gpt-4o-mini',
    reasoning_effort: ChatCompletionReasoningEffort = 'medium',
    modality: Modality = 'native',
    imageResolutionDpi: number = 96,
    browserCanvas: BrowserCanvas = 'A4',
    nConsensus: number = 1
  ): Promise<EvaluateSchemaResponse> {
    const preparedRequest = this.mixin.prepareEvaluate(
      documents,
      jsonSchema,
      groundTruths,
      model,
      reasoning_effort,
      modality,
      imageResolutionDpi,
      browserCanvas,
      nConsensus
    );
    const response = await this._client._preparedRequest(preparedRequest);
    return response;
  }

  async enhance(
    jsonSchema: Record<string, any> | string,
    documents: Array<string | Buffer | MIMEData>,
    groundTruths?: Array<Record<string, any>> | null,
    instructions?: string | null,
    model: string = 'gpt-4o-2024-11-20',
    temperature: number = 0,
    modality: Modality = 'native',
    flatLikelihoods?: Array<Record<string, number>> | Record<string, number> | null,
    toolsConfig?: EnhanceSchemaConfig
  ): Promise<Schema> {
    const _toolsConfig = toolsConfig || {};
    const preparedRequest = this.mixin.prepareEnhance(
      jsonSchema,
      documents,
      groundTruths,
      instructions,
      model,
      temperature,
      modality,
      flatLikelihoods,
      _toolsConfig
    );
    const response = await this._client._preparedRequest(preparedRequest);
    return new Schema({ json_schema: response.json_schema });
  }
}