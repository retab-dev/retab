import fs from 'fs';
import path from 'path';
import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';
import { PreparedRequest, FieldUnset } from '../../types/standards.js';
import { 
  DocumentExtractRequestSchema,
  RetabParsedChatCompletion,
  RetabParsedChatCompletionChunk,
  LogExtractionRequest,
  LogExtractionRequestSchema,
  LogExtractionResponse,
  ChatCompletionReasoningEffort
} from '../../types/documents/extractions.js';
import { MIMEData } from '../../types/mime.js';
import { BrowserCanvas } from '../../types/browser_canvas.js';
import { Modality } from '../../types/modalities.js';

// Helper function for MIME document preparation
function prepareMimeDocument(document: string | Buffer | MIMEData): MIMEData {
  if (typeof document === 'string') {
    // Resolve relative paths 
    const resolvedPath = path.resolve(document);
    // If it's a file path, read the file
    if (fs.existsSync(resolvedPath)) {
      const filename = path.basename(resolvedPath);
      const content = fs.readFileSync(resolvedPath);
      const base64Content = content.toString('base64');
      const mimeType = getMimeType(resolvedPath);
      return {
        id: 'doc_' + Date.now(),
        extension: path.extname(filename).slice(1) || 'unknown',
        content: base64Content,
        mime_type: mimeType,
        unique_filename: filename,
        size: content.length,
        filename,
        url: `data:${mimeType};base64,${base64Content}`,
      };
    }
    // Otherwise treat as text content
    const base64Content = Buffer.from(document).toString('base64');
    return {
      id: 'doc_' + Date.now(),
      extension: 'txt',
      content: base64Content,
      mime_type: 'text/plain',
      unique_filename: 'content.txt',
      size: Buffer.byteLength(document),
      filename: 'content.txt',
      url: `data:text/plain;base64,${base64Content}`,
    };
  } else if (Buffer.isBuffer(document)) {
    const base64Content = document.toString('base64');
    return {
      id: 'doc_' + Date.now(),
      extension: 'bin',
      content: base64Content,
      mime_type: 'application/octet-stream',
      unique_filename: 'content.bin',
      size: document.length,
      filename: 'content.bin',
      url: `data:application/octet-stream;base64,${base64Content}`,
    };
  }
  return document as MIMEData;
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
    '.html': 'text/html',
    '.xml': 'application/xml',
    '.csv': 'text/csv',
  };
  return mimeTypes[ext] || 'application/octet-stream';
}

function loadJsonSchema(jsonSchema: Record<string, any> | string): Record<string, any> {
  if (typeof jsonSchema === 'string') {
    if (fs.existsSync(jsonSchema)) {
      return JSON.parse(fs.readFileSync(jsonSchema, 'utf-8'));
    }
    return JSON.parse(jsonSchema);
  }
  return jsonSchema;
}

function assertValidModelExtraction(model: string): void {
  // Add model validation logic as needed
  if (!model || typeof model !== 'string') {
    throw new Error('Valid model must be provided for extraction');
  }
}



export class BaseExtractionsMixin {
  public prepareMimeDocument(document: string | Buffer | MIMEData): MIMEData {
    return prepareMimeDocument(document);
  }

  public prepareExtraction(
    jsonSchema: Record<string, any> | string,
    document?: string | Buffer | MIMEData | null,
    documents?: Array<string | Buffer | MIMEData> | null,
    imageResolutionDpi: number | typeof FieldUnset = FieldUnset,
    browserCanvas: BrowserCanvas | typeof FieldUnset = FieldUnset,
    model: string | typeof FieldUnset = FieldUnset,
    temperature: number | typeof FieldUnset = FieldUnset,
    modality: Modality | typeof FieldUnset = FieldUnset,
    reasoningEffort: ChatCompletionReasoningEffort | typeof FieldUnset = FieldUnset,
    stream: boolean = false,
    nConsensus: number | typeof FieldUnset = FieldUnset,
    store: boolean = false,
    idempotencyKey?: string | null
  ): PreparedRequest {
    if (model !== FieldUnset) {
      assertValidModelExtraction(model as string);
    }

    const loadedJsonSchema = loadJsonSchema(jsonSchema);

    // Handle both single document and multiple documents
    if (document !== null && document !== undefined && documents !== null && documents !== undefined) {
      throw new Error('Cannot provide both "document" and "documents" parameters. Use either one.');
    }

    // Convert single document to documents list for consistency
    let processedDocuments: MIMEData[];
    if (document !== null && document !== undefined) {
      processedDocuments = [prepareMimeDocument(document)];
    } else if (documents !== null && documents !== undefined) {
      processedDocuments = documents.map(doc => prepareMimeDocument(doc));
    } else {
      throw new Error('Must provide either "document" or "documents" parameter.');
    }

    // Build request dictionary with only provided fields
    const requestDict: any = {
      json_schema: loadedJsonSchema,
      documents: processedDocuments,
      stream,
      store,
    };

    if (model !== FieldUnset) requestDict.model = model;
    if (temperature !== FieldUnset) requestDict.temperature = temperature;
    if (modality !== FieldUnset) requestDict.modality = modality;
    if (reasoningEffort !== FieldUnset) requestDict.reasoning_effort = reasoningEffort;
    if (nConsensus !== FieldUnset) requestDict.n_consensus = nConsensus;
    if (imageResolutionDpi !== FieldUnset) requestDict.image_resolution_dpi = imageResolutionDpi;
    if (browserCanvas !== FieldUnset) requestDict.browser_canvas = browserCanvas;

    // Validate DocumentExtractRequest data
    const request = DocumentExtractRequestSchema.parse(requestDict);

    return {
      method: 'POST',
      url: '/v1/documents/extractions',
      data: request,
      idempotencyKey,
    };
  }

  public prepareLogExtraction(logRequest: LogExtractionRequest): PreparedRequest {
    const request = LogExtractionRequestSchema.parse(logRequest);
    return {
      method: 'POST',
      url: '/v1/documents/log_extraction',
      data: request,
    };
  }
}

export class Extractions extends SyncAPIResource {
  private mixin = new BaseExtractionsMixin();

  async extract(
    jsonSchema: Record<string, any> | string,
    options: {
      document?: string | Buffer | MIMEData;
      documents?: Array<string | Buffer | MIMEData>;
      imageResolutionDpi?: number;
      browserCanvas?: BrowserCanvas;
      model?: string;
      temperature?: number;
      modality?: Modality;
      reasoningEffort?: ChatCompletionReasoningEffort;
      stream?: boolean;
      nConsensus?: number;
      store?: boolean;
      idempotencyKey?: string;
    } = {}
  ): Promise<RetabParsedChatCompletion> {
    const preparedRequest = this.mixin.prepareExtraction(
      jsonSchema,
      options.document,
      options.documents,
      options.imageResolutionDpi || FieldUnset,
      options.browserCanvas || FieldUnset,
      options.model || FieldUnset,
      options.temperature || FieldUnset,
      options.modality || FieldUnset,
      options.reasoningEffort || FieldUnset,
      options.stream || false,
      options.nConsensus || FieldUnset,
      options.store || false,
      options.idempotencyKey
    );

    const response = await this._client._preparedRequest(preparedRequest);
    return response as RetabParsedChatCompletion;
  }

  async *extractStream(
    jsonSchema: Record<string, any> | string,
    options: {
      document?: string | Buffer | MIMEData;
      documents?: Array<string | Buffer | MIMEData>;
      imageResolutionDpi?: number;
      browserCanvas?: BrowserCanvas;
      model?: string;
      temperature?: number;
      modality?: Modality;
      reasoningEffort?: ChatCompletionReasoningEffort;
      nConsensus?: number;
      store?: boolean;
      idempotencyKey?: string;
    } = {}
  ): AsyncGenerator<RetabParsedChatCompletionChunk, void, unknown> {
    const preparedRequest = this.mixin.prepareExtraction(
      jsonSchema,
      options.document,
      options.documents,
      options.imageResolutionDpi || FieldUnset,
      options.browserCanvas || FieldUnset,
      options.model || FieldUnset,
      options.temperature || FieldUnset,
      options.modality || FieldUnset,
      options.reasoningEffort || FieldUnset,
      true, // stream = true
      options.nConsensus || FieldUnset,
      options.store || false,
      options.idempotencyKey
    );

    const streamGenerator = await this._client._preparedRequestStream(preparedRequest);
    for await (const chunk of streamGenerator) {
      yield chunk as RetabParsedChatCompletionChunk;
    }
  }

  async logExtraction(logRequest: LogExtractionRequest): Promise<LogExtractionResponse> {
    const preparedRequest = this.mixin.prepareLogExtraction(logRequest);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as LogExtractionResponse;
  }
}

export class AsyncExtractions extends AsyncAPIResource {
  private mixin = new BaseExtractionsMixin();

  async extract(
    jsonSchema: Record<string, any> | string,
    options: {
      document?: string | Buffer | MIMEData;
      documents?: Array<string | Buffer | MIMEData>;
      imageResolutionDpi?: number;
      browserCanvas?: BrowserCanvas;
      model?: string;
      temperature?: number;
      modality?: Modality;
      reasoningEffort?: ChatCompletionReasoningEffort;
      stream?: boolean;
      nConsensus?: number;
      store?: boolean;
      idempotencyKey?: string;
    } = {}
  ): Promise<RetabParsedChatCompletion> {
    const preparedRequest = this.mixin.prepareExtraction(
      jsonSchema,
      options.document,
      options.documents,
      options.imageResolutionDpi || FieldUnset,
      options.browserCanvas || FieldUnset,
      options.model || FieldUnset,
      options.temperature || FieldUnset,
      options.modality || FieldUnset,
      options.reasoningEffort || FieldUnset,
      options.stream || false,
      options.nConsensus || FieldUnset,
      options.store || false,
      options.idempotencyKey
    );

    const response = await this._client._preparedRequest(preparedRequest);
    return response as RetabParsedChatCompletion;
  }

  async *extractStream(
    jsonSchema: Record<string, any> | string,
    options: {
      document?: string | Buffer | MIMEData;
      documents?: Array<string | Buffer | MIMEData>;
      imageResolutionDpi?: number;
      browserCanvas?: BrowserCanvas;
      model?: string;
      temperature?: number;
      modality?: Modality;
      reasoningEffort?: ChatCompletionReasoningEffort;
      nConsensus?: number;
      store?: boolean;
      idempotencyKey?: string;
    } = {}
  ): AsyncGenerator<RetabParsedChatCompletionChunk, void, unknown> {
    const preparedRequest = this.mixin.prepareExtraction(
      jsonSchema,
      options.document,
      options.documents,
      options.imageResolutionDpi || FieldUnset,
      options.browserCanvas || FieldUnset,
      options.model || FieldUnset,
      options.temperature || FieldUnset,
      options.modality || FieldUnset,
      options.reasoningEffort || FieldUnset,
      true, // stream = true
      options.nConsensus || FieldUnset,
      options.store || false,
      options.idempotencyKey
    );

    const streamGenerator = this._client._preparedRequestStream(preparedRequest);
    for await (const chunk of streamGenerator) {
      yield chunk as RetabParsedChatCompletionChunk;
    }
  }

  async logExtraction(logRequest: LogExtractionRequest): Promise<LogExtractionResponse> {
    const preparedRequest = this.mixin.prepareLogExtraction(logRequest);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as LogExtractionResponse;
  }
}