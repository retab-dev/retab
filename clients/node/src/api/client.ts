import * as z from 'zod';

import { AbstractClient, CompositionClient } from '../client.js';
import APIDocuments from './documents/client';
import APISchemas from './schemas/client';
import APIExtractions from './extractions/client';
import APIWorkflows from './workflows/client';
import APIFiles from './files/client';
import APIJobs from './jobs/client';
import APIModels from './models/client';
import APIParses from './parses/client';
import APIClassifications from './classifications/client';
import APISplits from './splits/client';
import APIPartitions from './partitions/client';
import APIEdits from './edits/client';
import {
  MIMEDataInput,
  RetabParsedChatCompletion,
  RetabParsedChatCompletionChunk,
  ClassifyResponse,
  ZJSONSchema,
  ZMIMEData,
  ZRetabParsedChatCompletion,
  ZRetabParsedChatCompletionChunk,
  ZClassifyResponse,
  dataArray,
} from '../types.js';
import { RequestOptions } from '../client.js';

type EvalExtractProcessParams = {
  eval_id: string;
  iteration_id?: string;
  document: MIMEDataInput;
  documents?: MIMEDataInput[];
  metadata?: Record<string, unknown>;
};

type EvalSplitProcessParams = {
  eval_id: string;
  iteration_id?: string;
  document: MIMEDataInput;
  model?: string;
};

type EvalClassifyProcessParams = {
  eval_id: string;
  iteration_id?: string;
  document: MIMEDataInput;
};

type ProjectCreateParams = {
  name?: string;
  json_schema: string | Record<string, unknown>;
};

type ProjectUpdateParams = {
  name?: string;
  json_schema?: string | Record<string, unknown>;
};

type DatasetCreateParams = {
  name: string;
  base_json_schema: string | Record<string, unknown>;
};

type DatasetUpdateParams = {
  name?: string;
};

type DatasetDocumentCreateParams = {
  mime_data: MIMEDataInput;
  prediction_data?: Record<string, unknown>;
};

type DatasetDocumentUpdateParams = {
  validation_flags?: Record<string, unknown>;
  extraction_id?: string;
};

type IterationUpdateDraftParams = {
  schema_overrides?: Record<string, unknown>;
  inference_settings?: Record<string, unknown>;
};

type IterationDocumentUpdateParams = {
  extraction_id?: string;
  prediction_data?: Record<string, unknown>;
};

type ProjectsExtractParams = {
  project_id: string;
  document?: MIMEDataInput;
  documents?: MIMEDataInput[];
};

function buildEvalProcessPath(
  prefix: string,
  evalId: string,
  iterationId?: string,
  stream = false
): string {
  const suffix = iterationId ? `/${evalId}/${iterationId}` : `/${evalId}`;
  return stream ? `${prefix}${suffix}/stream` : `${prefix}${suffix}`;
}

async function buildEvalMultipartBody(
  document: MIMEDataInput,
  extraBody: Record<string, unknown> = {}
): Promise<Record<string, unknown>> {
  const parsedDocument = await ZMIMEData.parseAsync(document);
  return {
    document: JSON.stringify(parsedDocument),
    ...extraBody,
  };
}

async function normalizeMimeDocument(document: MIMEDataInput): Promise<string> {
  return JSON.stringify(await ZMIMEData.parseAsync(document));
}

async function normalizeMimeDocuments(documents: MIMEDataInput[]): Promise<string[]> {
  return Promise.all(documents.map(normalizeMimeDocument));
}

class APIProjectIterations extends CompositionClient {
  async create(projectId: string, datasetId: string, options?: RequestOptions): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/iterations`,
      method: 'POST',
      body: { ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  async get(
    projectId: string,
    datasetId: string,
    iterationId: string,
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async list(projectId: string, datasetId: string, options?: RequestOptions): Promise<any[]> {
    return this._fetchJson(dataArray(z.any()), {
      url: `/projects/${projectId}/datasets/${datasetId}/iterations`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async updateDraft(
    projectId: string,
    datasetId: string,
    iterationId: string,
    draft: IterationUpdateDraftParams,
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}`,
      method: 'PATCH',
      body: { draft, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  async getSchema(
    projectId: string,
    datasetId: string,
    iterationId: string,
    { useDraft }: { useDraft?: boolean } = {},
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/schema`,
      method: 'GET',
      params: {
        ...(useDraft !== undefined ? { use_draft: useDraft } : {}),
        ...(options?.params || {}),
      },
      headers: options?.headers,
    });
  }

  async processDocuments(
    projectId: string,
    datasetId: string,
    iterationId: string,
    datasetDocumentId: string,
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/processDocumentsFromDatasetId`,
      method: 'POST',
      body: { dataset_document_id: datasetDocumentId, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  async getDocument(
    projectId: string,
    datasetId: string,
    iterationId: string,
    iterationDocumentId: string,
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${iterationDocumentId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async listDocuments(
    projectId: string,
    datasetId: string,
    iterationId: string,
    options?: RequestOptions
  ): Promise<any[]> {
    return this._fetchJson(dataArray(z.any()), {
      url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async updateDocument(
    projectId: string,
    datasetId: string,
    iterationId: string,
    iterationDocumentId: string,
    body: IterationDocumentUpdateParams,
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${iterationDocumentId}`,
      method: 'PATCH',
      body: { ...body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  async deleteDocument(
    projectId: string,
    datasetId: string,
    iterationId: string,
    iterationDocumentId: string,
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${iterationDocumentId}`,
      method: 'DELETE',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async getMetrics(
    projectId: string,
    datasetId: string,
    iterationId: string,
    { forceRefresh }: { forceRefresh?: boolean } = {},
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/metrics`,
      method: 'GET',
      params: {
        ...(forceRefresh !== undefined ? { force_refresh: forceRefresh } : {}),
        ...(options?.params || {}),
      },
      headers: options?.headers,
    });
  }

  async finalize(
    projectId: string,
    datasetId: string,
    iterationId: string,
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/finalize`,
      method: 'POST',
      body: { ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  async delete(
    projectId: string,
    datasetId: string,
    iterationId: string,
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}`,
      method: 'DELETE',
      params: options?.params,
      headers: options?.headers,
    });
  }
}

class APIProjectDatasets extends CompositionClient {
  iterations: APIProjectIterations;

  constructor(client: CompositionClient) {
    super(client);
    this.iterations = new APIProjectIterations(this);
  }

  async create(
    projectId: string,
    params: DatasetCreateParams,
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets`,
      method: 'POST',
      body: {
        name: params.name,
        base_json_schema: await ZJSONSchema.parseAsync(params.base_json_schema),
        ...(options?.body || {}),
      },
      params: options?.params,
      headers: options?.headers,
    });
  }

  async get(projectId: string, datasetId: string, options?: RequestOptions): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async list(projectId: string, options?: RequestOptions): Promise<any[]> {
    return this._fetchJson(dataArray(z.any()), {
      url: `/projects/${projectId}/datasets`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async update(
    projectId: string,
    datasetId: string,
    params: DatasetUpdateParams,
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}`,
      method: 'PATCH',
      body: { ...params, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  async duplicate(
    projectId: string,
    datasetId: string,
    params: { name: string },
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/duplicate`,
      method: 'POST',
      body: { ...params, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  async addDocument(
    projectId: string,
    datasetId: string,
    params: DatasetDocumentCreateParams,
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/dataset-documents`,
      method: 'POST',
      body: {
        project_id: projectId,
        dataset_id: datasetId,
        mime_data: await ZMIMEData.parseAsync(params.mime_data),
        ...(params.prediction_data ? { prediction_data: params.prediction_data } : {}),
        ...(options?.body || {}),
      },
      params: options?.params,
      headers: options?.headers,
    });
  }

  async getDocument(
    projectId: string,
    datasetId: string,
    documentId: string,
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/dataset-documents/${documentId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async listDocuments(
    projectId: string,
    datasetId: string,
    options?: RequestOptions
  ): Promise<any[]> {
    return this._fetchJson(z.array(z.any()), {
      url: `/projects/${projectId}/datasets/${datasetId}/dataset-documents`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async updateDocument(
    projectId: string,
    datasetId: string,
    documentId: string,
    body: DatasetDocumentUpdateParams,
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/dataset-documents/${documentId}`,
      method: 'PATCH',
      body: { ...body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  async deleteDocument(
    projectId: string,
    datasetId: string,
    documentId: string,
    options?: RequestOptions
  ): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}/dataset-documents/${documentId}`,
      method: 'DELETE',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async delete(projectId: string, datasetId: string, options?: RequestOptions): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/datasets/${datasetId}`,
      method: 'DELETE',
      params: options?.params,
      headers: options?.headers,
    });
  }
}

class APIProjects extends CompositionClient {
  datasets: APIProjectDatasets;

  constructor(client: CompositionClient) {
    super(client);
    this.datasets = new APIProjectDatasets(this);
  }

  async create(params: ProjectCreateParams, options?: RequestOptions): Promise<any> {
    return this._fetchJson(z.any(), {
      url: '/projects',
      method: 'POST',
      body: {
        ...(params.name !== undefined ? { name: params.name } : {}),
        json_schema: await ZJSONSchema.parseAsync(params.json_schema),
        ...(options?.body || {}),
      },
      params: options?.params,
      headers: options?.headers,
    });
  }

  async get(projectId: string, options?: RequestOptions): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async list(options?: RequestOptions): Promise<any[]> {
    return this._fetchJson(dataArray(z.any()), {
      url: '/projects',
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async prepare_update(
    projectId: string,
    params: ProjectUpdateParams,
    options?: RequestOptions
  ): Promise<Record<string, unknown>> {
    const body: Record<string, unknown> = { ...(options?.body || {}) };
    if (params.name !== undefined) {
      body['name'] = params.name;
    }
    if (params.json_schema !== undefined) {
      body['json_schema'] = await ZJSONSchema.parseAsync(params.json_schema);
    }
    return {
      url: `/projects/${projectId}`,
      method: 'PATCH',
      body,
      params: options?.params,
      headers: options?.headers,
    };
  }

  async publish(projectId: string, options?: RequestOptions): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}/publish`,
      method: 'POST',
      body: { ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  async delete(projectId: string, options?: RequestOptions): Promise<any> {
    return this._fetchJson(z.any(), {
      url: `/projects/${projectId}`,
      method: 'DELETE',
      params: options?.params,
      headers: options?.headers,
    });
  }

  async extract(
    params: ProjectsExtractParams,
    options?: RequestOptions
  ): Promise<RetabParsedChatCompletion> {
    const documents = params.documents ?? (params.document ? [params.document] : []);
    return this._fetchJson(ZRetabParsedChatCompletion, {
      url: `/projects/extract/${params.project_id}`,
      method: 'POST',
      bodyMime: 'multipart/form-data',
      body: {
        documents: await normalizeMimeDocuments(documents),
        ...(options?.body || {}),
      },
      params: options?.params,
      headers: options?.headers,
    });
  }

  async split(
    params: ProjectsExtractParams,
    options?: RequestOptions
  ): Promise<RetabParsedChatCompletion> {
    const documents = params.documents ?? (params.document ? [params.document] : []);
    return this._fetchJson(ZRetabParsedChatCompletion, {
      url: `/projects/split/${params.project_id}`,
      method: 'POST',
      bodyMime: 'multipart/form-data',
      body: {
        documents: await normalizeMimeDocuments(documents),
        ...(options?.body || {}),
      },
      params: options?.params,
      headers: options?.headers,
    });
  }
}

class APIEvalsExtract extends CompositionClient {
  async process(
    params: EvalExtractProcessParams,
    options?: RequestOptions
  ): Promise<RetabParsedChatCompletion> {
    if (params.documents && params.documents.length > 0) {
      throw new Error("evals.extract.process accepts only 'document'");
    }
    const body = await buildEvalMultipartBody(params.document, {
      ...(params.metadata ? { metadata: JSON.stringify(params.metadata) } : {}),
      ...(options?.body || {}),
    });
    return this._fetchJson(ZRetabParsedChatCompletion, {
      url: buildEvalProcessPath('/evals/extract/extract', params.eval_id, params.iteration_id),
      method: 'POST',
      bodyMime: 'multipart/form-data',
      body,
      params: options?.params,
      headers: options?.headers,
    });
  }

  async processStream(
    params: EvalExtractProcessParams,
    options?: RequestOptions
  ): Promise<AsyncGenerator<RetabParsedChatCompletionChunk>> {
    if (params.documents && params.documents.length > 0) {
      throw new Error("evals.extract.process accepts only 'document'");
    }
    const body = await buildEvalMultipartBody(params.document, {
      ...(params.metadata ? { metadata: JSON.stringify(params.metadata) } : {}),
      ...(options?.body || {}),
    });
    return this._fetchStream(ZRetabParsedChatCompletionChunk, {
      url: buildEvalProcessPath(
        '/evals/extract/extract',
        params.eval_id,
        params.iteration_id,
        true
      ),
      method: 'POST',
      bodyMime: 'multipart/form-data',
      body,
      params: options?.params,
      headers: options?.headers,
    });
  }
}

class APIEvalsSplit extends CompositionClient {
  async process(
    params: EvalSplitProcessParams,
    options?: RequestOptions
  ): Promise<RetabParsedChatCompletion> {
    const body = await buildEvalMultipartBody(params.document, {
      ...(params.model ? { model: params.model } : {}),
      ...(options?.body || {}),
    });
    return this._fetchJson(ZRetabParsedChatCompletion, {
      url: buildEvalProcessPath('/evals/split/extract', params.eval_id, params.iteration_id),
      method: 'POST',
      bodyMime: 'multipart/form-data',
      body,
      params: options?.params,
      headers: options?.headers,
    });
  }
}

class APIEvalsClassify extends CompositionClient {
  async process(
    params: EvalClassifyProcessParams,
    options?: RequestOptions
  ): Promise<ClassifyResponse> {
    const body = await buildEvalMultipartBody(params.document, options?.body || {});
    return this._fetchJson(ZClassifyResponse, {
      url: buildEvalProcessPath('/evals/classify/extract', params.eval_id, params.iteration_id),
      method: 'POST',
      bodyMime: 'multipart/form-data',
      body,
      params: options?.params,
      headers: options?.headers,
    });
  }
}

export default class APIV1 extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
    Object.defineProperty(this, 'evals', {
      configurable: true,
      enumerable: false,
      writable: false,
      value: {
        extract: new APIEvalsExtract(this),
        split: new APIEvalsSplit(this),
        classify: new APIEvalsClassify(this),
      },
    });
    Object.defineProperty(this, 'projects', {
      configurable: true,
      enumerable: false,
      writable: false,
      value: new APIProjects(this),
    });
  }
  files = new APIFiles(this);
  documents = new APIDocuments(this);
  schemas = new APISchemas(this);
  extractions = new APIExtractions(this);
  workflows = new APIWorkflows(this);
  jobs = new APIJobs(this);
  models = new APIModels(this);
  parses = new APIParses(this);
  classifications = new APIClassifications(this);
  splits = new APISplits(this);
  partitions = new APIPartitions(this);
  edits = new APIEdits(this);
  declare evals: {
    extract: APIEvalsExtract;
    split: APIEvalsSplit;
    classify: APIEvalsClassify;
  };
  declare projects: APIProjects;
}
